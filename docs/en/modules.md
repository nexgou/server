# Modules

> **[← Back to README](../../README.md)**

---

## Table of Contents

- [What is a Module?](#what-is-a-module)
- [Module Options](#module-options)
- [Root Module](#root-module)
- [Feature Modules](#feature-modules)
- [Providers & Dependency Injection](#providers--dependency-injection)
- [Importing & Exporting Providers](#importing--exporting-providers)
- [Built-in Modules](#built-in-modules)
- [Module Tree](#module-tree)
- [IoC Container Resolution](#ioc-container-resolution)

---

## What is a Module?

A **module** is the fundamental organizational unit of a Nexgou application. Every module encapsulates a cohesive slice of your application — its controllers, services (providers), and the sub-modules it depends on.

```
AppModule
├── ConfigModule   (built-in)
├── LogModule      (built-in)
├── UserModule
│   ├── UserController
│   └── UserService
└── OrderModule
    ├── OrderController
    ├── OrderService
    └── (imports UserModule.UserService via Exports)
```

Every application has exactly **one root module** passed to `nexgou.CreateApp()`. Feature modules form a tree rooted there.

---

## Module Options

```go
var MyModule = nexgou.Module(nexgou.ModuleOptions{
    Imports:     []nexgou.IModule{...},   // modules whose exports are available here
    Controllers: []any{NewMyController},  // constructor functions for controllers
    Providers:   []any{NewMyService},     // constructor functions for services/repos/etc.
    Exports:     []any{NewMyService},     // subset of Providers to expose to importing modules
})
```

| Field | Type | Description |
|:---|:---|:---|
| `Imports` | `[]nexgou.IModule` | Other modules to import. Their exported providers become available for injection here. |
| `Controllers` | `[]any` | **Constructor functions** (not instances) for HTTP/WS/gRPC controllers. |
| `Providers` | `[]any` | **Constructor functions** for services, repositories, factories, etc. |
| `Exports` | `[]any` | Subset of `Providers` constructors to make available to modules that import this one. |

> **Important:** `Controllers` and `Providers` take **constructor functions**, not struct instances. The IoC container calls them, resolving their parameters automatically.

---

## Root Module

The root module is the entry point of the application. It is the only module passed to `nexgou.CreateApp()`.

```go
// app.module.go
package main

import (
    nexgou "github.com/nexgou/server"
    "myapp/catalog"
    "myapp/order"
    "myapp/user"
)

var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{
        nexgou.ConfigModule,
        nexgou.LogModule,
        user.UserModule,
        catalog.CatalogModule,
        order.OrderModule,
    },
})
```

The root module typically has no `Controllers` or `Providers` of its own — it only imports feature modules.

---

## Feature Modules

A **feature module** groups everything related to a specific domain (users, orders, products, etc.).

```go
// user/user.module.go
package user

import nexgou "github.com/nexgou/server"

var UserModule = nexgou.Module(nexgou.ModuleOptions{
    Controllers: []any{NewUserController},
    Providers:   []any{NewUserService, NewUserRepository},
    Exports:     []any{NewUserService}, // UserService is usable by modules that import UserModule
})
```

---

## Providers & Dependency Injection

**Providers** are any value produced by a constructor function. They are registered in the module's IoC container and injected automatically wherever their type appears as a parameter.

### Constructor function rules

1. Must be a plain Go function (any name).
2. Its parameters are the dependencies to inject.
3. Its **first return value** is the provided type (used as the registry key).
4. An optional second return value of `error` is allowed (panics if non-nil at boot).

```go
// Two providers, one depending on the other
func NewUserRepository(cfg *nexgou.ConfigService) *UserRepository {
    dsn := cfg.MustGet("DATABASE_URL")
    return &UserRepository{dsn: dsn}
}

func NewUserService(
    repo *UserRepository,
    log  *nexgou.LoggerService,
) *UserService {
    return &UserService{repo: repo, log: log.WithContext("UserService")}
}
```

### Injection into controllers

```go
func NewUserController(
    svc *UserService,
    cfg *nexgou.ConfigService,
) *UserController {
    return &UserController{svc: svc, cfg: cfg}
}
```

The container resolves the entire dependency graph — you never call constructors manually.

---

## Importing & Exporting Providers

For a provider from module **A** to be available in module **B**, module A must **export** it and module B must **import** module A.

```
AppModule
└── OrderModule   (imports UserModule)
    └── UserModule (exports UserService)
```

```go
// user/user.module.go
var UserModule = nexgou.Module(nexgou.ModuleOptions{
    Providers: []any{NewUserService},
    Exports:   []any{NewUserService}, // <-- makes UserService available to importers
})

// order/order.module.go
var OrderModule = nexgou.Module(nexgou.ModuleOptions{
    Imports:     []nexgou.IModule{UserModule}, // <-- gets UserService
    Controllers: []any{NewOrderController},
    Providers:   []any{NewOrderService},
})

// Now OrderService can declare *UserService as a parameter:
func NewOrderService(userSvc *UserService) *OrderService {
    return &OrderService{userSvc: userSvc}
}
```

Providers that are **not** in `Exports` are private to their module.

---

## Built-in Modules

Nexgou ships two ready-to-use modules. Import them into your root module to make their providers available throughout the app.

### `nexgou.ConfigModule`

Provides: `*nexgou.ConfigService` (typed, safe environment variable access).

```go
var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{nexgou.ConfigModule, ...},
})

// Inject in any provider:
func NewMyService(cfg *nexgou.ConfigService) *MyService {
    port := cfg.GetInt("PORT", 3000)
    ...
}
```

See [Config](config.md) for full reference.

### `nexgou.LogModule`

Provides: `*nexgou.LoggerService` (structured logger with JSON/text output).

```go
var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{nexgou.LogModule, ...},
})

// Inject and scope to a context:
func NewMyService(log *nexgou.LoggerService) *MyService {
    return &MyService{log: log.WithContext("MyService")}
}
```

See [Logger](logger.md) for full reference.

---

## Module Tree

The framework walks the module tree **depth-first** before resolving any dependencies. This means a module's `Imports` are fully processed before the module itself is initialized — imported providers are always ready when a module's constructors run.

```
AppModule
├── ConfigModule   → registers *ConfigService
├── LogModule      → registers *LoggerService
└── UserModule
    ├── (has access to ConfigService, LogService via Imports if declared)
    ├── UserRepository  ← constructor called with *ConfigService injected
    └── UserController  ← constructor called with *UserService injected
```

---

## IoC Container Resolution

Each module has its own container. Resolution follows these steps:

1. Look up the required type in the module's own providers.
2. If not found, look in exported providers of all imported modules.
3. Call the constructor function, recursively resolving its parameters.
4. Cache the result as a **singleton** — each type is instantiated once per module.

If a dependency cannot be resolved, `CreateApp` panics with a descriptive error.

```go
// This panics at startup if *DatabaseService is not registered
// anywhere in the module tree visible to OrderModule
func NewOrderService(db *DatabaseService) *OrderService { ... }
```

Always ensure every dependency type is either provided locally or exported by an imported module.
