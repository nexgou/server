# Getting Started

> **[← Back to README](../../README.md)**

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Project Structure](#project-structure)
- [Bootstrap Lifecycle](#bootstrap-lifecycle)
- [First Application](#first-application)
- [Running the Server](#running-the-server)
- [Next Steps](#next-steps)

---

## Prerequisites

- **Go 1.21** or higher

```bash
go version  # go version go1.21.x ...
```

---

## Installation

```bash
go get github.com/nexgou/server
```

This installs the core framework. Additional sub-packages (`src/middleware`, `src/filter`, `test/nexgoutest`) are part of the same module and are imported as needed.

---

## Project Structure

Nexgou imposes no mandatory directory structure, but a typical feature-based layout looks like:

```
myapp/
├── main.go              # Entry point — wires middleware and starts the server
├── app.module.go        # Root module — imports all feature modules
├── user/
│   ├── user.module.go   # Feature module
│   ├── user.controller.go
│   └── user.service.go
└── order/
    ├── order.module.go
    ├── order.controller.go
    └── order.service.go
```

Each feature lives in its own package (directory). The root module (`AppModule`) imports all feature modules.

---

## Bootstrap Lifecycle

When you call `nexgou.CreateApp(root)`, the framework:

1. Walks the module tree recursively (depth-first, `Imports` first)
2. Builds a per-module IoC container — registers all `Providers` constructors
3. Detects provider exports and makes them available to importing modules
4. Instantiates all `Controllers` via the container (injects dependencies automatically)
5. Calls `Register()` on HTTP controllers → registers routes
6. Returns a ready `*App` — no network ports are open yet

`app.Listen(port)` opens the actual network listener.

---

## First Application

### 1. Initialize a Go module

```bash
mkdir myapp && cd myapp
go mod init myapp
go get github.com/nexgou/server
```

### 2. Create the service

```go
// user/user.service.go
package user

type UserService struct{}

func NewUserService() *UserService {
    return &UserService{}
}

type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

func (s *UserService) FindAll() []User {
    return []User{
        {ID: "1", Name: "Alice"},
        {ID: "2", Name: "Bob"},
    }
}

func (s *UserService) FindOne(id string) *User {
    return &User{ID: id, Name: "Alice"}
}
```

### 3. Create the controller

```go
// user/user.controller.go
package user

import nexgou "github.com/nexgou/server"

type UserController struct {
    svc *UserService
}

func NewUserController(svc *UserService) *UserController {
    return &UserController{svc: svc}
}

func (c *UserController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.Get("/users",     c.FindAll),
        nexgou.Post("/users",    c.Create),
        nexgou.Get("/users/:id", c.FindOne),
    }
}

func (c *UserController) FindAll(ctx *nexgou.Context) error {
    return ctx.JSON(200, c.svc.FindAll())
}

func (c *UserController) Create(ctx *nexgou.Context) error {
    var body struct {
        Name string `json:"name"`
    }
    if err := ctx.Body(&body); err != nil {
        return nexgou.BadRequestException("invalid body")
    }
    return ctx.JSON(201, nexgou.H{"message": "created", "name": body.Name})
}

func (c *UserController) FindOne(ctx *nexgou.Context) error {
    id := ctx.Param("id")
    return ctx.JSON(200, c.svc.FindOne(id))
}
```

### 4. Create the feature module

```go
// user/user.module.go
package user

import nexgou "github.com/nexgou/server"

var UserModule = nexgou.Module(nexgou.ModuleOptions{
    Controllers: []any{NewUserController},
    Providers:   []any{NewUserService},
})
```

### 5. Create the root module

```go
// app.module.go
package main

import nexgou "github.com/nexgou/server"
import "myapp/user"

var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{
        nexgou.LogModule,
        nexgou.ConfigModule,
        user.UserModule,
    },
})
```

### 6. Create the entry point

```go
// main.go
package main

import (
    "log"
    "time"

    nexgou "github.com/nexgou/server"
    "github.com/nexgou/server/src/filter"
    "github.com/nexgou/server/src/middleware"
)

func main() {
    app := nexgou.CreateApp(AppModule)

    // Middleware (order matters — executed top to bottom per request)
    app.Use(middleware.Recovery())          // recover from panics → 500
    app.Use(middleware.SecurityHeaders())   // secure HTTP headers
    app.Use(middleware.CorsWithOptions(middleware.CorsOptions{
        AllowedOrigins: []string{"*"},
    }))
    app.Use(middleware.RateLimit(100, time.Minute))  // 100 req/min per IP
    app.Use(middleware.Timeout(30 * time.Second))    // 30s deadline
    app.Use(middleware.BodyLimit(1 << 20))           // 1 MB body cap
    app.Use(middleware.Logger())                     // request logging

    // Global exception filter — structured JSON errors
    app.SetFilter(&filter.HttpExceptionFilter{})

    if err := app.Listen(3000); err != nil {
        log.Fatal(err)
    }
}
```

---

## Running the Server

```bash
go run .
```

The startup banner will print all registered routes. Test with curl:

```bash
curl http://localhost:3000/users
curl http://localhost:3000/users/42
curl -X POST http://localhost:3000/users -H 'Content-Type: application/json' -d '{"name":"Charlie"}'
```

---

## Next Steps

| Topic                                         | Guide                           |
| :-------------------------------------------- | :------------------------------ |
| Module system, DI, exports                    | [Modules](modules.md)           |
| Route versioning, guards, interceptors, pipes | [Controllers](controllers.md)   |
| Full middleware reference                     | [Middleware](middleware.md)     |
| Security headers, rate limiting, timeout      | [Security](security.md)         |
| Env vars & config                             | [Config](config.md)             |
| Structured logging                            | [Logger](logger.md)             |
| Unit & integration tests                      | [Testing](testing.md)           |
| Complete working sample                       | [`samples/api`](../samples/api) |
