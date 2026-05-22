# Logger

> **[← Back to README](../../README.md)**

---

## Table of Contents

- [Overview](#overview)
- [Enabling LogModule](#enabling-logmodule)
- [Injecting LoggerService](#injecting-loggerservice)
- [Log Levels](#log-levels)
- [Scoped Loggers](#scoped-loggers)
- [Structured Fields](#structured-fields)
- [Output Formats](#output-formats)
- [Environment Variables](#environment-variables)
- [API Reference](#api-reference)

---

## Overview

`LoggerService` is a structured logger with two output modes: colored text (development) and JSON (production). It is provided by `nexgou.LogModule` and injectable anywhere in your application via the IoC container.

---

## Enabling LogModule

Import `nexgou.LogModule` in your root module:

```go
var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{
        nexgou.ConfigModule,
        nexgou.LogModule, // <-- enables LoggerService injection
        UserModule,
    },
})
```

---

## Injecting LoggerService

Declare `*nexgou.LoggerService` as a constructor parameter:

```go
type UserService struct {
    log *nexgou.ScopedLogger
}

func NewUserService(logger *nexgou.LoggerService) *UserService {
    return &UserService{
        log: logger.WithContext("UserService"),
    }
}
```

Always call `.WithContext(name)` immediately in the constructor to bind the logger to the service name.

---

## Log Levels

The four standard levels, in order of severity:

| Level | Constant | When to use |
|:---|:---|:---|
| DEBUG | `nexgou.LevelDebug` | Verbose development diagnostics |
| INFO | `nexgou.LevelInfo` | Normal operational events |
| WARN | `nexgou.LevelWarn` | Recoverable issues, degraded behavior |
| ERROR | `nexgou.LevelError` | Errors that require attention |

```go
log.Debug("cache miss", "key", "user:42")
log.Info("user created", "id", "42", "name", "Alice")
log.Warn("rate limit approaching", "ip", "1.2.3.4", "remaining", 5)
log.Error("database connection failed", "err", err)
```

---

## Scoped Loggers

`logger.WithContext(name)` returns a `*nexgou.ScopedLogger` that prepends `[name]` to every log line. This makes it immediately obvious which component produced each log entry.

```go
// In each service, scope the logger at construction time
func NewOrderService(logger *nexgou.LoggerService) *OrderService {
    return &OrderService{log: logger.WithContext("OrderService")}
}

func NewPaymentService(logger *nexgou.LoggerService) *PaymentService {
    return &PaymentService{log: logger.WithContext("PaymentService")}
}
```

Output (text format):

```
[INFO]  [OrderService]   order created  id=ORD-001 total=99.99
[INFO]  [PaymentService] payment queued id=PAY-001 amount=99.99
[WARN]  [PaymentService] retry 1/3      id=PAY-001
[ERROR] [PaymentService] payment failed id=PAY-001 err=timeout
```

---

## Structured Fields

Log methods accept key-value pairs after the message:

```go
func (s *UserService) CreateUser(name, email string) (*User, error) {
    s.log.Info("creating user", "name", name, "email", email)

    user, err := s.repo.Insert(name, email)
    if err != nil {
        s.log.Error("failed to create user", "name", name, "err", err)
        return nil, err
    }

    s.log.Info("user created", "id", user.ID, "name", user.Name)
    return user, nil
}
```

Keys and values are pairs: each key must be followed by its value. The number of arguments after the message must be even.

```go
// Correct
log.Info("event", "key1", val1, "key2", val2)

// Incorrect — odd number of extra args (last val has no key)
log.Info("event", "key1", val1, val2)
```

---

## Output Formats

### Text format (default / development)

Colorized, human-readable output to stdout:

```
[INFO]  [UserService] user created  id=42 name=Alice
[WARN]  [AuthService] slow token validation  duration=250ms
[ERROR] [DBService]   connection failed  err=dial tcp refused
```

### JSON format (production)

Machine-readable JSON, one object per line, suitable for log aggregators (Datadog, Loki, CloudWatch):

```json
{"level":"INFO","context":"UserService","msg":"user created","id":"42","name":"Alice","ts":"2026-05-22T10:30:00Z"}
{"level":"WARN","context":"AuthService","msg":"slow token validation","duration":"250ms","ts":"2026-05-22T10:30:01Z"}
```

---

## Environment Variables

| Variable | Values | Default | Description |
|:---|:---|:---|:---|
| `LOG_LEVEL` | `debug`, `info`, `warn`, `error` | `info` | Minimum level to output |
| `LOG_FORMAT` | `json`, `text` | `text` | Output format |

```bash
LOG_LEVEL=debug LOG_FORMAT=json go run .
```

The logger reads these variables once at construction time (`NewLoggerService()` is called during app bootstrap).

---

## API Reference

### `LoggerService`

```go
func (l *LoggerService) WithContext(name string) *ScopedLogger
func (l *LoggerService) Debug(msg string, args ...any)
func (l *LoggerService) Info(msg string, args ...any)
func (l *LoggerService) Warn(msg string, args ...any)
func (l *LoggerService) Error(msg string, args ...any)
```

### `ScopedLogger`

```go
func (s *ScopedLogger) Debug(msg string, args ...any)
func (s *ScopedLogger) Info(msg string, args ...any)
func (s *ScopedLogger) Warn(msg string, args ...any)
func (s *ScopedLogger) Error(msg string, args ...any)
```

### Log level constants

```go
nexgou.LevelDebug  // 0
nexgou.LevelInfo   // 1
nexgou.LevelWarn   // 2
nexgou.LevelError  // 3
```
