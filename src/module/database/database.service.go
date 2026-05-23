// Package database provides a generic SQL database service for Nexgou applications.
//
// It wraps Go's standard database/sql package and provides a uniform API over
// any registered SQL driver (MySQL, PostgreSQL, SQLite, etc.).
//
// Use the driver-specific sub-modules to register a driver and configure the DSN:
//
//	import "github.com/nexgou/server/src/module/mysql"
//	import "github.com/nexgou/server/src/module/postgres"
//	import "github.com/nexgou/server/src/module/sqlite"
//
// Usage:
//
//	func NewUserRepo(db *database.DatabaseService) *UserRepo {
//	    rows, err := db.Query(ctx, "SELECT id, name FROM users WHERE active = ?", true)
//	    defer rows.Close()
//	    for rows.Next() { ... }
//	}
package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/nexgou/server/src/logger"
)

// Config holds the connection pool settings for DatabaseService.
// The DSN and driver are provided by the driver-specific sub-modules.
type Config struct {
	// Driver is the SQL driver name (e.g. "mysql", "pgx", "sqlite").
	Driver string
	// DSN is the data source name / connection string.
	DSN string
	// MaxOpenConns limits the number of open connections (default: 25).
	MaxOpenConns int
	// MaxIdleConns limits the number of idle connections (default: 5).
	MaxIdleConns int
	// ConnMaxLifetime is the maximum connection lifetime (default: 5m).
	ConnMaxLifetime time.Duration
}

// DatabaseService wraps *sql.DB and exposes common database operations.
// It is safe for concurrent use.
type DatabaseService struct {
	db  *sql.DB
	log *logger.ScopedLogger
}

// NewDatabaseServiceFromConfig opens a connection pool using the provided Config.
// Called by driver-specific modules (mysql, postgres, sqlite).
func NewDatabaseServiceFromConfig(cfg Config, log *logger.LoggerService) *DatabaseService {
	svcLog := log.WithContext("DatabaseService[" + cfg.Driver + "]")

	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		svcLog.Error("failed to open database", "driver", cfg.Driver, "err", err)
		panic("nexgou/database: " + err.Error())
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	} else {
		db.SetMaxOpenConns(25)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	} else {
		db.SetMaxIdleConns(5)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	} else {
		db.SetConnMaxLifetime(5 * time.Minute)
	}

	svc := &DatabaseService{db: db, log: svcLog}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		svcLog.Warn("database ping failed — service may be unavailable", "driver", cfg.Driver, "err", err)
	} else {
		svcLog.Info("connected", "driver", cfg.Driver)
	}

	return svc
}

// DB returns the underlying *sql.DB for advanced usage.
func (s *DatabaseService) DB() *sql.DB {
	return s.db
}

// ── Query ─────────────────────────────────────────────────────────────────────

// Query executes a SELECT and returns the result rows.
// Callers must call rows.Close() when done.
func (s *DatabaseService) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a SELECT that is expected to return at most one row.
func (s *DatabaseService) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return s.db.QueryRowContext(ctx, query, args...)
}

// ── Exec ──────────────────────────────────────────────────────────────────────

// Exec executes a statement (INSERT, UPDATE, DELETE, DDL) and returns the result.
func (s *DatabaseService) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

// ExecWithLastID executes an INSERT and returns the last inserted row ID.
func (s *DatabaseService) ExecWithLastID(ctx context.Context, query string, args ...any) (int64, error) {
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ExecWithRowsAffected executes a statement and returns the number of rows affected.
func (s *DatabaseService) ExecWithRowsAffected(ctx context.Context, query string, args ...any) (int64, error) {
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ── Transaction ───────────────────────────────────────────────────────────────

// Transaction runs fn inside a database transaction.
// If fn returns an error the transaction is rolled back; otherwise it is committed.
func (s *DatabaseService) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// ── Lifecycle ─────────────────────────────────────────────────────────────────

// Ping verifies the database connection is still alive.
func (s *DatabaseService) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Close closes the underlying connection pool.
func (s *DatabaseService) Close() error {
	return s.db.Close()
}

// Stats returns connection pool statistics.
func (s *DatabaseService) Stats() sql.DBStats {
	return s.db.Stats()
}
