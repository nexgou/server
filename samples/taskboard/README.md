# Taskboard Sample

A demo application that showcases five Nexgou ecosystem modules working together:

| Module | Package | Responsibility |
|---|---|---|
| `sqlite` | `src/module/sqlite` | Persistent SQLite storage (pure Go, no CGO) |
| `jwt` | `src/module/jwt` | HS256 token signing & `JwtGuard` |
| `validation` | `src/module/validation` | Struct validation via `validate` tags |
| `events` | `src/module/events` | In-process pub/sub (`task.created`, `task.completed`) |
| `cron` | `src/module/cron` | Scheduled jobs (stats every minute, cleanup at midnight) |

## Run

```bash
# from repo root
JWT_SECRET=supersecret go run ./samples/taskboard
# Server listens on :3001
```

## Routes

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/v1/auth/login` | — | Get a JWT token |
| `GET` | `/v1/tasks` | Bearer JWT | List your tasks |
| `POST` | `/v1/tasks` | Bearer JWT | Create a task |
| `PATCH` | `/v1/tasks/:id/complete` | Bearer JWT | Mark a task done |
| `DELETE` | `/v1/tasks/:id` | Bearer JWT | Delete a task |

## Quick Start (curl)

```bash
# 1. Login — password is always "password" in this demo
TOKEN=$(curl -s -X POST http://localhost:3001/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"password"}' | jq -r .access_token)

# 2. Create a task
curl -s -X POST http://localhost:3001/v1/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Buy groceries"}'

# 3. List tasks
curl -s http://localhost:3001/v1/tasks \
  -H "Authorization: Bearer $TOKEN"

# 4. Complete task with id=1
curl -s -X PATCH http://localhost:3001/v1/tasks/1/complete \
  -H "Authorization: Bearer $TOKEN"

# 5. Delete task with id=1
curl -s -X DELETE http://localhost:3001/v1/tasks/1 \
  -H "Authorization: Bearer $TOKEN"
```

## Configuration

| Variable | Default | Description |
|---|---|---|
| `JWT_SECRET` | `nexgou-secret` | HMAC signing key |
| `JWT_EXPIRATION` | `24h` | Token TTL (Go duration string) |
| `JWT_ISSUER` | `nexgou` | `iss` claim value |
| `SQLITE_PATH` | `nexgou.db` | SQLite file path (use `:memory:` for ephemeral) |

## Cron Jobs

| Name | Expression | Action |
|---|---|---|
| `task.stats` | `0 * * * * *` (every minute) | Logs total/done/pending count |
| `task.cleanup` | `0 0 0 * * *` (midnight daily) | Deletes all completed tasks |

## Module Graph

```
AppModule
├── ConfigModule
├── LogModule
├── auth.Module
│   ├── jwt.Module
│   └── validation.ValidationModule
└── task.Module
    ├── sqlite.Module  →  database.DatabaseService
    ├── events.Module  →  EventEmitter
    ├── cron.Module    →  CronService
    ├── jwt.Module     →  JwtService + JwtGuard
    └── validation.ValidationModule
```
