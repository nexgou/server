package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/core"
)

// Level represents a logging severity level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelSilent
)

// Format represents the logger output format.
type Format int

const (
	FormatText Format = iota
	FormatJSON
)

// LoggerOptions configures a logger instance without reading environment variables.
type LoggerOptions struct {
	Level  Level
	Format Format
}

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelSilent:
		return "SILENT"
	default:
		return "INFO"
	}
}

func ParseLoggerLevel(value string) Level {
	switch value {
	case "debug":
		return LevelDebug
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	case "silent":
		return LevelSilent
	default:
		return LevelInfo
	}
}

func ParseLoggerFormat(value string) Format {
	if value == "json" {
		return FormatJSON
	}
	return FormatText
}

// levelColor returns the ANSI color for a log level.
func levelColor(l Level) string {
	switch l {
	case LevelDebug:
		return "\033[90m" // gray
	case LevelInfo:
		return "\033[36m" // cyan
	case LevelWarn:
		return "\033[33m" // yellow
	case LevelError:
		return "\033[31m" // red
	default:
		return "\033[0m"
	}
}

// LoggerService is a structured logger injectable via the IoC container.
// Output format is controlled by the LOG_FORMAT environment variable:
//
//	LOG_FORMAT=json  → machine-readable JSON (for production / log aggregators)
//	(unset)          → human-readable colored text (for development)
//
// Log level is controlled by the LOG_LEVEL environment variable:
//
//	LOG_LEVEL=debug  → show all levels
//	LOG_LEVEL=info   → show info, warn, error  (default)
//	LOG_LEVEL=warn   → show warn, error
//	LOG_LEVEL=error  → show error only
type LoggerService struct {
	minLevel Level
	json     bool
}

// NewLoggerService creates a new LoggerService.
// Format and level are resolved from environment variables at creation time.
func NewLoggerService() *LoggerService {
	l := &LoggerService{}
	l.json = ParseLoggerFormat(os.Getenv("LOG_FORMAT")) == FormatJSON
	l.minLevel = ParseLoggerLevel(os.Getenv("LOG_LEVEL"))
	return l
}

// NewLogger creates a LoggerService from explicit options.
func NewLogger(options ...LoggerOptions) *LoggerService {
	opt := LoggerOptions{Level: LevelInfo, Format: FormatText}
	if len(options) > 0 {
		opt = options[0]
	}
	return &LoggerService{minLevel: opt.Level, json: opt.Format == FormatJSON}
}

// Enabled reports whether the given level would be written.
func (l *LoggerService) Enabled(level Level) bool {
	return l.minLevel != LevelSilent && level >= l.minLevel
}

// WithContext returns a ScopedLogger that prefixes every message with name.
// Use this inside services and controllers to identify the log source.
//
//	log := logger.WithContext("UserService")
//	log.Info("User created", "id", 42)
func (l *LoggerService) WithContext(name string) *ScopedLogger {
	return &ScopedLogger{parent: l, context: name}
}

// Debug logs a message at DEBUG level with optional key-value pairs.
func (l *LoggerService) Debug(msg string, args ...any) { l.log(LevelDebug, "", msg, args) }

// Info logs a message at INFO level with optional key-value pairs.
func (l *LoggerService) Info(msg string, args ...any) { l.log(LevelInfo, "", msg, args) }

// Warn logs a message at WARN level with optional key-value pairs.
func (l *LoggerService) Warn(msg string, args ...any) { l.log(LevelWarn, "", msg, args) }

// Error logs a message at ERROR level with optional key-value pairs.
func (l *LoggerService) Error(msg string, args ...any) { l.log(LevelError, "", msg, args) }

// log is the central write method.
func (l *LoggerService) log(level Level, context, msg string, args []any) {
	if !l.Enabled(level) {
		return
	}
	if l.json {
		l.writeJSON(level, context, msg, args)
	} else {
		l.writeText(level, context, msg, args)
	}
}

// writeText writes a human-readable colored log line.
//
//	[Nexgou] INFO  UserService — User created id=42
func (l *LoggerService) writeText(level Level, context, msg string, args []any) {
	reset := "\033[0m"
	gray := "\033[90m"
	color := levelColor(level)

	ctx := ""
	if context != "" {
		ctx = fmt.Sprintf(" %s%s%s —", "\033[2m", context, reset)
	}

	fields := formatTextFields(args)
	suffix := ""
	if fields != "" {
		suffix = " " + gray + fields + reset
	}

	fmt.Fprintf(os.Stdout, "%s[Nexgou]%s %s%-5s%s%s %s%s\n",
		gray, reset,
		color, level.String(), reset,
		ctx, msg, suffix)
}

// writeJSON writes a machine-readable JSON log entry.
//
//	{"ts":"2026-05-22T10:00:00Z","level":"info","context":"UserService","msg":"User created","id":42}
func (l *LoggerService) writeJSON(level Level, context, msg string, args []any) {
	entry := map[string]any{
		"ts":    time.Now().UTC().Format(time.RFC3339),
		"level": level.String(),
		"msg":   msg,
	}
	if context != "" {
		entry["context"] = context
	}
	// Merge key-value pairs into the entry map.
	for i := 0; i+1 < len(args); i += 2 {
		key := fmt.Sprintf("%v", args[i])
		entry[key] = args[i+1]
	}
	b, _ := json.Marshal(entry)
	fmt.Fprintf(os.Stdout, "%s\n", b)
}

// formatTextFields formats key-value pairs as "key=value key=value".
func formatTextFields(args []any) string {
	if len(args) == 0 {
		return ""
	}
	out := ""
	for i := 0; i+1 < len(args); i += 2 {
		if out != "" {
			out += " "
		}
		out += fmt.Sprintf("%v=%v", args[i], args[i+1])
	}
	// Odd trailing arg — print as-is.
	if len(args)%2 != 0 {
		if out != "" {
			out += " "
		}
		out += fmt.Sprintf("%v", args[len(args)-1])
	}
	return out
}

// ── ScopedLogger ──────────────────────────────────────────────────────────────

// ScopedLogger is a LoggerService bound to a named context (e.g. a service name).
// Obtained via LoggerService.WithContext.
type ScopedLogger struct {
	parent  *LoggerService
	context string
}

// Debug logs at DEBUG level.
func (s *ScopedLogger) Debug(msg string, args ...any) {
	s.parent.log(LevelDebug, s.context, msg, args)
}

// Info logs at INFO level.
func (s *ScopedLogger) Info(msg string, args ...any) {
	s.parent.log(LevelInfo, s.context, msg, args)
}

// Warn logs at WARN level.
func (s *ScopedLogger) Warn(msg string, args ...any) {
	s.parent.log(LevelWarn, s.context, msg, args)
}

// Error logs at ERROR level.
func (s *ScopedLogger) Error(msg string, args ...any) {
	s.parent.log(LevelError, s.context, msg, args)
}

// ── LogModule ─────────────────────────────────────────────────────────────────

// LogModule is a ready-to-use Nexgou module that registers and exports
// LoggerService. Import it in any module that needs structured logging.
//
// Usage:
//
//	var AppModule = nexgou.Module(nexgou.ModuleOptions{
//	    Imports: []nexgou.IModule{nexgou.LogModule, UserModule},
//	})
//
//	func NewUserService(log *logger.LoggerService) *UserService {
//	    l := log.WithContext("UserService")
//	    l.Info("initialized")
//	}
var LogModule common.IModule = core.NewModule(common.ModuleOptions{
	Providers: []any{NewLoggerService},
	Exports:   []any{NewLoggerService},
})
