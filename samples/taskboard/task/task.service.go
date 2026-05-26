package task

import (
	"context"
	"fmt"
	"strconv"

	nexgou "github.com/nexgou/server"
	"github.com/nexgou/server/src/logger"
	"github.com/nexgou/server/src/module/database"
	"github.com/nexgou/server/src/module/events"
)

// Task is the domain model.
type Task struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Done   bool   `json:"done"`
	UserID string `json:"user_id"`
}

// TaskService manages tasks in SQLite and publishes domain events.
type TaskService struct {
	db     *database.DatabaseService
	events *events.EventEmitter
	log    *logger.ScopedLogger
}

// Service is a lightweight task service backed by Store for sample tests.
type Service struct {
	store *Store
	log   *nexgou.ScopedLogger
}

// NewService creates a Store-backed task service.
func NewService(store *Store, log *nexgou.LoggerService) *Service {
	return &Service{store: store, log: log.WithContext("TaskService")}
}

// NewTaskService creates a new TaskService and runs migrations.
func NewTaskService(db *database.DatabaseService, emitter *events.EventEmitter, log *logger.LoggerService) *TaskService {
	svc := &TaskService{
		db:     db,
		events: emitter,
		log:    log.WithContext("TaskService"),
	}
	svc.migrate()
	return svc
}

// migrate creates the tasks table if it does not exist.
func (s *TaskService) migrate() {
	_, err := s.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS tasks (
			id      INTEGER PRIMARY KEY AUTOINCREMENT,
			title   TEXT    NOT NULL,
			done    INTEGER NOT NULL DEFAULT 0,
			user_id TEXT    NOT NULL
		)
	`)
	if err != nil {
		s.log.Error("migration failed", "err", err)
		panic("taskboard: migrate: " + err.Error())
	}
	s.log.Info("tasks table ready")
}

// FindAll returns all tasks belonging to a user.
func (s *TaskService) FindAll(userID string) ([]Task, error) {
	rows, err := s.db.Query(context.Background(),
		"SELECT id, title, done, user_id FROM tasks WHERE user_id = ? ORDER BY id DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var done int
		if err := rows.Scan(&t.ID, &t.Title, &done, &t.UserID); err != nil {
			return nil, err
		}
		t.Done = done == 1
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []Task{}
	}
	return tasks, nil
}

// FindOne returns a single task by ID.
func (s *TaskService) FindOne(id string) (*Task, error) {
	var t Task
	var done int
	err := s.db.QueryRow(context.Background(),
		"SELECT id, title, done, user_id FROM tasks WHERE id = ?", id).
		Scan(&t.ID, &t.Title, &done, &t.UserID)
	if err != nil {
		return nil, err
	}
	t.Done = done == 1
	return &t, nil
}

// Create inserts a new task and emits "task.created".
func (s *TaskService) Create(title, userID string) (*Task, error) {
	id, err := s.db.ExecWithLastID(context.Background(),
		"INSERT INTO tasks (title, done, user_id) VALUES (?, 0, ?)", title, userID)
	if err != nil {
		return nil, err
	}
	t := &Task{ID: id, Title: title, Done: false, UserID: userID}
	s.events.Emit("task.created", t)
	s.log.Info("task created", "id", id, "title", title)
	return t, nil
}

// Complete marks a task as done and emits "task.completed".
func (s *TaskService) Complete(id string) (*Task, error) {
	_, err := s.db.Exec(context.Background(),
		"UPDATE tasks SET done = 1 WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	t, err := s.FindOne(id)
	if err != nil {
		return nil, err
	}
	s.events.Emit("task.completed", t)
	s.log.Info("task completed", "id", id)
	return t, nil
}

// Delete removes a task by ID.
func (s *TaskService) Delete(id string) error {
	n, err := s.db.ExecWithRowsAffected(context.Background(),
		"DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("task %s not found", id)
	}
	s.log.Info("task deleted", "id", id)
	return nil
}

// Count returns the total number of tasks.
func (s *TaskService) Count() int {
	var n int
	_ = s.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM tasks").Scan(&n)
	return n
}

// CountDone returns the number of completed tasks.
func (s *TaskService) CountDone() int {
	var n int
	_ = s.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM tasks WHERE done = 1").Scan(&n)
	return n
}

// DeleteCompleted removes all tasks marked as done. Called by the cron job.
func (s *TaskService) DeleteCompleted() (int64, error) {
	return s.db.ExecWithRowsAffected(context.Background(), "DELETE FROM tasks WHERE done = 1")
}

// idToInt64 converts a string task ID to int64 (utility used by delete).
func idToInt64(id string) int64 {
	n, _ := strconv.ParseInt(id, 10, 64)
	return n
}

func (s *Service) FindAll(userID string) ([]Task, error) {
	return s.store.List(context.Background(), userID)
}

func (s *Service) Create(title, userID string) (*Task, error) {
	task, err := s.store.Insert(context.Background(), title, userID)
	if err == nil {
		s.log.Info("task created", "id", task.ID, "title", title)
	}
	return task, err
}

func (s *Service) Complete(id string) (*Task, error) {
	task, err := s.store.Complete(context.Background(), id)
	if err == nil {
		s.log.Info("task completed", "id", id)
	}
	return task, err
}

func (s *Service) Delete(id string) error {
	err := s.store.Delete(context.Background(), id)
	if err == nil {
		s.log.Info("task deleted", "id", id)
	}
	return err
}
