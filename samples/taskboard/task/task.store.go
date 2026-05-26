package task

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"

	nexgou "github.com/nexgou/server"
	_ "modernc.org/sqlite"
)

type Store struct {
	db  *sql.DB
	log *nexgou.ScopedLogger
}

func NewStore(config *nexgou.ConfigService, log *nexgou.LoggerService) *Store {
	path := config.GetOrDefault("SQLITE_PATH", defaultSQLitePath())
	if err := prepareSQLitePath(path); err != nil {
		panic("taskboard: prepare sqlite path: " + err.Error())
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		panic("taskboard: open sqlite: " + err.Error())
	}

	store := &Store{db: db, log: log.WithContext("TaskStore")}
	if err := store.Migrate(context.Background()); err != nil {
		panic("taskboard: migrate sqlite: " + err.Error())
	}
	return store
}

func defaultSQLitePath() string {
	if _, err := os.Stat(filepath.Join("samples", "taskboard")); err == nil {
		return filepath.Join("samples", "taskboard", "nexgou.db")
	}
	return "nexgou.db"
}

func prepareSQLitePath(path string) error {
	if path == ":memory:" || strings.HasPrefix(path, "file:") {
		return nil
	}
	directory := filepath.Dir(path)
	if directory == "." || directory == "" {
		return nil
	}
	return os.MkdirAll(directory, 0o755)
}

func (store *Store) Migrate(ctx context.Context) error {
	_, err := store.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			done INTEGER NOT NULL DEFAULT 0,
			user_id TEXT NOT NULL
		)
	`)
	if err == nil {
		store.log.Info("tasks table ready")
	}
	return err
}

func (store *Store) Close() error {
	return store.db.Close()
}

func (store *Store) Insert(ctx context.Context, title string, userID string) (*Task, error) {
	result, err := store.db.ExecContext(ctx, "INSERT INTO tasks (title, done, user_id) VALUES (?, 0, ?)", title, userID)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &Task{ID: id, Title: title, Done: false, UserID: userID}, nil
}

func (store *Store) List(ctx context.Context, userID string) ([]Task, error) {
	rows, err := store.db.QueryContext(ctx, "SELECT id, title, done, user_id FROM tasks WHERE user_id = ? ORDER BY id DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]Task, 0)
	for rows.Next() {
		var task Task
		var done int
		if err := rows.Scan(&task.ID, &task.Title, &done, &task.UserID); err != nil {
			return nil, err
		}
		task.Done = done == 1
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (store *Store) Find(ctx context.Context, id string) (*Task, error) {
	var task Task
	var done int
	err := store.db.QueryRowContext(ctx, "SELECT id, title, done, user_id FROM tasks WHERE id = ?", id).Scan(&task.ID, &task.Title, &done, &task.UserID)
	if err != nil {
		return nil, err
	}
	task.Done = done == 1
	return &task, nil
}

func (store *Store) Complete(ctx context.Context, id string) (*Task, error) {
	result, err := store.db.ExecContext(ctx, "UPDATE tasks SET done = 1 WHERE id = ?", id)
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
	result, err := store.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
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
