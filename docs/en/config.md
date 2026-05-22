# Config

> **[← Back to README](../../README.md)**

---

## Table of Contents

- [Overview](#overview)
- [Enabling ConfigModule](#enabling-configmodule)
- [Injecting ConfigService](#injecting-configservice)
- [Reading Values](#reading-values)
- [MustGet (required variables)](#mustget-required-variables)
- [Environment-Specific Configuration](#environment-specific-configuration)
- [API Reference](#api-reference)

---

## Overview

`ConfigService` provides typed, safe access to environment variables. It is provided by `nexgou.ConfigModule` — a ready-to-use module you import into your root module. Once imported, `*nexgou.ConfigService` becomes injectable in any provider or controller.

---

## Enabling ConfigModule

Import `nexgou.ConfigModule` in your root module:

```go
var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{
        nexgou.ConfigModule, // <-- enables ConfigService injection everywhere
        nexgou.LogModule,
        UserModule,
    },
})
```

That's all. No configuration file needed — values are read from the OS environment.

---

## Injecting ConfigService

Declare `*nexgou.ConfigService` as a parameter in any constructor:

```go
type DatabaseService struct {
    dsn  string
    pool int
}

func NewDatabaseService(cfg *nexgou.ConfigService) *DatabaseService {
    return &DatabaseService{
        dsn:  cfg.MustGet("DATABASE_URL"),
        pool: cfg.GetInt("DB_POOL_SIZE", 10),
    }
}
```

The IoC container injects the singleton `ConfigService` instance automatically.

---

## Reading Values

### `Get(key string) string`

Returns the value of the environment variable `key`. Returns an empty string if the variable is not set.

```go
region := cfg.Get("AWS_REGION") // "" if not set
```

### `GetOrDefault(key, fallback string) string`

Returns the value if set, or `fallback` if the variable is not set or empty.

```go
host := cfg.GetOrDefault("HOST", "localhost")
```

### `GetInt(key string, fallback int) int`

Parses the value as an integer. Returns `fallback` if the variable is not set or cannot be parsed.

```go
port    := cfg.GetInt("PORT", 3000)
workers := cfg.GetInt("WORKER_COUNT", 4)
```

### `GetBool(key string, fallback bool) bool`

Parses the value as a boolean. Truthy values: `"1"`, `"true"`, `"yes"` (case-insensitive). Falsy: `"0"`, `"false"`, `"no"`. Returns `fallback` if unset or unrecognized.

```go
debug   := cfg.GetBool("DEBUG", false)
verbose := cfg.GetBool("VERBOSE", false)
```

### `MustGet(key string) string`

Returns the value. **Panics at startup** if the variable is not set. Use this for required configuration that makes no sense to have a default for.

```go
jwtSecret  := cfg.MustGet("JWT_SECRET")
databaseURL := cfg.MustGet("DATABASE_URL")
```

---

## Environment-Specific Configuration

Use a `.env` loader (e.g. [`godotenv`](https://github.com/joho/godotenv)) before starting the app to load a `.env` file:

```go
// main.go
import "github.com/joho/godotenv"

func main() {
    _ = godotenv.Load() // loads .env if present, ignores error in production

    app := nexgou.CreateApp(AppModule)
    // ...
}
```

`.env` file:

```dotenv
PORT=3000
DATABASE_URL=postgres://user:pass@localhost/mydb
JWT_SECRET=super-secret
DEBUG=true
LOG_LEVEL=info
LOG_FORMAT=json
```

---

## API Reference

```go
// Package: github.com/nexgou/server (accessed as nexgou.ConfigService)

func (c *ConfigService) Get(key string) string
func (c *ConfigService) GetOrDefault(key, fallback string) string
func (c *ConfigService) GetInt(key string, fallback int) int
func (c *ConfigService) GetBool(key string, fallback bool) bool
func (c *ConfigService) MustGet(key string) string
```

| Method | Panics? | Returns on missing |
|:---|:---:|:---|
| `Get` | No | `""` |
| `GetOrDefault` | No | `fallback` |
| `GetInt` | No | `fallback` |
| `GetBool` | No | `fallback` |
| `MustGet` | **Yes** | — |
