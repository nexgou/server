package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	nexgou "github.com/nexgou/server"
	_ "modernc.org/sqlite"
)

type Config struct {
	ServiceName string
	Version     string
	DBPath      string
}

type User struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Age       int    `json:"age"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type UserPayload struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type UserList struct {
	Items  []User `json:"items"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Total  int    `json:"total"`
}

type Store struct {
	db *sql.DB
}

func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	store := &Store{db: db}
	if err := store.Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (store *Store) Close() error {
	return store.db.Close()
}

func (store *Store) Migrate(ctx context.Context) error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA temp_store = MEMORY",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA foreign_keys = ON",
	}
	for _, statement := range pragmas {
		if _, err := store.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	_, err := store.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			age INTEGER NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
	`)
	return err
}

func (store *Store) Create(ctx context.Context, payload UserPayload) (*User, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := store.db.ExecContext(ctx,
		"INSERT INTO users (name, email, age, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		payload.Name, payload.Email, payload.Age, now, now,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return store.Find(ctx, strconv.FormatInt(id, 10))
}

func (store *Store) Find(ctx context.Context, id string) (*User, error) {
	user := &User{}
	err := store.db.QueryRowContext(ctx,
		"SELECT id, name, email, age, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Age, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (store *Store) List(ctx context.Context, limit int, offset int) (*UserList, error) {
	rows, err := store.db.QueryContext(ctx,
		"SELECT id, name, email, age, created_at, updated_at FROM users ORDER BY id DESC LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]User, 0, limit)
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Age, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	total := 0
	if err := store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&total); err != nil {
		return nil, err
	}
	return &UserList{Items: items, Limit: limit, Offset: offset, Total: total}, nil
}

func (store *Store) Update(ctx context.Context, id string, payload UserPayload) (*User, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := store.db.ExecContext(ctx,
		"UPDATE users SET name = ?, email = ?, age = ?, updated_at = ? WHERE id = ?",
		payload.Name, payload.Email, payload.Age, now, id,
	)
	if err != nil {
		return nil, err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if changed == 0 {
		return nil, sql.ErrNoRows
	}
	return store.Find(ctx, id)
}

func (store *Store) Delete(ctx context.Context, id string) error {
	result, err := store.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed == 0 {
		return sql.ErrNoRows
	}
	return nil
}

type Controller struct {
	store   *Store
	service string
	version string
}

func NewController(store *Store, config Config) *Controller {
	return &Controller{store: store, service: config.ServiceName, version: config.Version}
}

func (controller *Controller) Register() []nexgou.Route {
	return []nexgou.Route{
		nexgou.Get("/health", controller.Health),
		nexgou.Post("/users", controller.Create),
		nexgou.Get("/users", controller.List),
		nexgou.Get("/users/:id", controller.Find),
		nexgou.Put("/users/:id", controller.Update),
		nexgou.Delete("/users/:id", controller.Delete),
	}
}

func (controller *Controller) Health(ctx *nexgou.Context) error {
	return ctx.JSON(http.StatusOK, nexgou.H{"status": "ok", "service": controller.service, "version": controller.version})
}

func (controller *Controller) Create(ctx *nexgou.Context) error {
	payload, err := readPayload(ctx)
	if err != nil {
		return err
	}
	user, err := controller.store.Create(ctx.Request.Context(), payload)
	if err != nil {
		return nexgou.BadRequestException("user could not be created")
	}
	return ctx.JSON(http.StatusCreated, user)
}

func (controller *Controller) Find(ctx *nexgou.Context) error {
	user, err := controller.store.Find(ctx.Request.Context(), ctx.Param("id"))
	if errors.Is(err, sql.ErrNoRows) {
		return nexgou.NotFoundException("user not found")
	}
	if err != nil {
		return nexgou.InternalServerErrorException("user could not be fetched")
	}
	return ctx.JSON(http.StatusOK, user)
}

func (controller *Controller) List(ctx *nexgou.Context) error {
	query := ctx.Request.URL.Query()
	limit := parsePositiveInt(query.Get("limit"), 20)
	offset := parsePositiveInt(query.Get("offset"), 0)
	users, err := controller.store.List(ctx.Request.Context(), limit, offset)
	if err != nil {
		return nexgou.InternalServerErrorException("users could not be listed")
	}
	return ctx.JSON(http.StatusOK, users)
}

func (controller *Controller) Update(ctx *nexgou.Context) error {
	payload, err := readPayload(ctx)
	if err != nil {
		return err
	}
	user, err := controller.store.Update(ctx.Request.Context(), ctx.Param("id"), payload)
	if errors.Is(err, sql.ErrNoRows) {
		return nexgou.NotFoundException("user not found")
	}
	if err != nil {
		return nexgou.BadRequestException("user could not be updated")
	}
	return ctx.JSON(http.StatusOK, user)
}

func (controller *Controller) Delete(ctx *nexgou.Context) error {
	err := controller.store.Delete(ctx.Request.Context(), ctx.Param("id"))
	if errors.Is(err, sql.ErrNoRows) {
		return nexgou.NotFoundException("user not found")
	}
	if err != nil {
		return nexgou.InternalServerErrorException("user could not be deleted")
	}
	return ctx.JSON(http.StatusOK, nexgou.H{"deleted": true})
}

func NewNexGouApp(config Config, store *Store) *nexgou.App {
	module := nexgou.Module(nexgou.ModuleOptions{
		Controllers: []any{func() *Controller { return NewController(store, config) }},
	})
	app := nexgou.CreateApp(module)
	app.Use(nexgou.Recovery())
	app.SetFilter(&nexgou.HttpExceptionFilter{})
	return app
}

func readPayload(ctx *nexgou.Context) (UserPayload, error) {
	var payload UserPayload
	if err := ctx.Body(&payload); err != nil {
		return payload, nexgou.BadRequestException("invalid request body")
	}
	if payload.Name == "" || payload.Email == "" || payload.Age <= 0 {
		return payload, nexgou.BadRequestException("name, email and age are required")
	}
	return payload, nil
}

func parsePositiveInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}
