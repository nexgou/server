# Sample: REST API

A complete example of a Nexgou REST API application demonstrating the core framework features: modules, controllers, services, dependency injection, guards, interceptors, and the full security middleware pipeline.

## What this sample covers

| Feature | Where |
|---------|-------|
| Module system & DI | `app.module.go`, `user/user.module.go` |
| Controller & declarative routing | `user/user.controller.go` |
| Service with injected dependencies | `user/user.service.go` |
| `ConfigService` — environment variables | `user/user.service.go` |
| `LoggerService` — structured logging | `user/user.service.go` |
| Guards per route (`AuthGuard`, `RoleGuard`) | `user/user.controller.go` |
| Interceptors per route (`LogInterceptor`) | `user/user.controller.go` |
| Per-route `RateLimitGuard` | `user/user.controller.go` |
| Per-route `TimeoutInterceptor` | `user/user.controller.go` |
| Per-route `BodyLimitInterceptor` | `user/user.controller.go` |
| Global security pipeline | `main.go` |
| Exception filter | `main.go` |
| Route versioning (`/v1/...`) | `user/user.controller.go` |

## Project structure

```
samples/api/
├── main.go               # Bootstrap — middleware pipeline + server start
├── app.module.go         # Root module — imports ConfigModule, LogModule, UserModule
├── README.md
└── user/
    ├── user.controller.go  # Routes: GET/POST /v1/users, GET /v1/users/:id
    ├── user.service.go     # Business logic, injects ConfigService + LoggerService
    └── user.module.go      # UserModule definition
```

## Routes

| Method | Path | Guards | Interceptors |
|--------|------|--------|--------------|
| `GET` | `/v1/users` | `AuthGuard`, `RoleGuard` | `LogInterceptor` |
| `POST` | `/v1/users` | `AuthGuard`, `RateLimitGuard` | `TimeoutInterceptor`, `BodyLimitInterceptor` |
| `GET` | `/v1/users/:id` | — | `TimeoutInterceptor` |

## Run

```bash
cd samples/api
go run .
# Server starts on http://localhost:3000
```

## Try it

```bash
# GET /v1/users — requires Authorization header
curl http://localhost:3000/v1/users \
  -H "Authorization: Bearer token"

# GET /v1/users/:id
curl http://localhost:3000/v1/users/42 \
  -H "Authorization: Bearer token"

# POST /v1/users
curl -X POST http://localhost:3000/v1/users \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice"}'

# Missing Authorization → 403 Forbidden
curl http://localhost:3000/v1/users

# Health check
curl http://localhost:3000/
```

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3000` | Overrides the default port in `main.go` if used |
| `LOG_LEVEL` | `info` | `debug` \| `info` \| `warn` \| `error` |
| `LOG_FORMAT` | text | Set to `json` for machine-readable output |

## Global middleware order

```
Recovery → SecurityHeaders → CORS → RateLimit(100/min) → Timeout(30s) → BodyLimit(1MB) → Logger
```

## Key concepts demonstrated

### Dependency Injection

`UserService` declares its dependencies as constructor parameters. The IoC container resolves them automatically:

```go
func NewUserService(cfg *config.ConfigService, log *logger.LoggerService) *UserService {
    return &UserService{cfg: cfg, log: log.WithContext("UserService")}
}
```

`ConfigModule` and `LogModule` are imported in `AppModule` and export their services, making them available to all modules in the tree.

### Per-route security

Routes can layer additional guards and interceptors on top of the global pipeline:

```go
nexgou.Post("/users", c.Create).
    Guard(
        &AuthGuard{},
        &middleware.RateLimitGuard{Max: 10, Window: time.Minute},
    ).
    Intercept(
        &middleware.TimeoutInterceptor{Duration: 10 * time.Second},
        &middleware.BodyLimitInterceptor{MaxBytes: 64 << 10},
    ).
    Version("v1")
```

Both the global limit (100 req/min) and the per-route limit (10 req/min) apply to `POST /v1/users`.
