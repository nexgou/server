package logger_test

import (
	"testing"

	"github.com/nexgou/server/src/logger"
)

// Tests just verify no panic occurs — the logger writes to os.Stdout.

func TestLoggerService_TextFormat(t *testing.T) {
	// Default: text format
	t.Setenv("LOG_FORMAT", "")
	t.Setenv("LOG_LEVEL", "debug")
	l := logger.NewLoggerService()
	l.Debug("debug message", "key", "value")
	l.Info("info message")
	l.Warn("warn message", "k", 1)
	l.Error("error message")
}

func TestLoggerService_JSONFormat(t *testing.T) {
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("LOG_LEVEL", "debug")
	l := logger.NewLoggerService()
	l.Debug("debug json")
	l.Info("info json", "user", "alice", "id", 42)
	l.Warn("warn json")
	l.Error("error json", "code", 500)
}

func TestLoggerService_LevelFiltering(t *testing.T) {
	t.Setenv("LOG_FORMAT", "")
	t.Setenv("LOG_LEVEL", "error")
	l := logger.NewLoggerService()
	// These should be silently filtered (no panic).
	l.Debug("should not appear")
	l.Info("should not appear")
	l.Warn("should not appear")
	l.Error("should appear")
}

func TestLoggerService_WarnLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "warn")
	l := logger.NewLoggerService()
	l.Debug("filtered")
	l.Info("filtered")
	l.Warn("visible")
	l.Error("visible")
}

func TestLoggerService_UnknownLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "verbose") // unknown → defaults to INFO
	l := logger.NewLoggerService()
	l.Info("should appear")
}

func TestLoggerService_WithContext(t *testing.T) {
	t.Setenv("LOG_FORMAT", "")
	t.Setenv("LOG_LEVEL", "debug")
	l := logger.NewLoggerService()
	scoped := l.WithContext("UserService")
	scoped.Debug("debug scoped")
	scoped.Info("info scoped", "id", 1)
	scoped.Warn("warn scoped")
	scoped.Error("error scoped")
}

func TestLoggerService_WithContext_JSON(t *testing.T) {
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("LOG_LEVEL", "debug")
	l := logger.NewLoggerService()
	scoped := l.WithContext("OrderService")
	scoped.Info("order created", "orderId", "abc123")
}

func TestLoggerService_OddArgs(t *testing.T) {
	// Odd number of args — trailing arg printed as-is.
	t.Setenv("LOG_FORMAT", "")
	t.Setenv("LOG_LEVEL", "debug")
	l := logger.NewLoggerService()
	l.Info("odd args", "key1", "val1", "orphan")
}

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level logger.Level
		want  string
	}{
		{logger.LevelDebug, "DEBUG"},
		{logger.LevelInfo, "INFO"},
		{logger.LevelWarn, "WARN"},
		{logger.LevelError, "ERROR"},
		{logger.Level(99), "INFO"}, // unknown → INFO
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}
