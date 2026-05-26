<div align="center">

<br/>

<img src="docs/banner.png" alt="Nexgou Banner" width="100%" />

# Nexgou

**A progressive Go framework for building fast, modular, production-ready HTTP APIs.**

Architectural clarity inspired by NestJS, with Go performance and simple deployment.

<br/>

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-22c55e?style=for-the-badge)](LICENSE)
[![Status](https://img.shields.io/badge/Status-WIP-f59e0b?style=for-the-badge)]()
[![CI](https://img.shields.io/github/actions/workflow/status/nexgou/server/ci.yml?branch=main&style=for-the-badge&label=CI&logo=github-actions&logoColor=white)](https://github.com/nexgou/server/actions)

[English](README.md) · [Español](README.es.md)

</div>

---

## What Is It?

Nexgou is an application framework for Go HTTP services. It gives you modules, dependency injection, controllers, middleware, guards, interceptors, pipes, filters, config, logging, testing helpers, and benchmark-ready samples in one project structure.

It is for teams that want more than a router, but still want Go's simplicity and speed.

## What Is It For?

- Building REST APIs with clear module boundaries.
- Keeping business logic injectable and testable.
- Applying auth, validation, logging, security, and error handling consistently.
- Shipping services that are easy to benchmark, observe, and evolve.

## Why Nexgou?

| Need                      | Nexgou   | Typical router-first stack |
| ------------------------- | -------- | -------------------------- |
| Modules by domain         | Built in | Manual conventions         |
| Dependency injection      | Built in | Usually custom             |
| Guards and interceptors   | Built in | Middleware workarounds     |
| Pipes and validation flow | Built in | Handler-level code         |
| Exception filters         | Built in | Manual error mapping       |
| Config and logger modules | Built in | Extra packages/globals     |
| Testing helpers           | Built in | Project-specific setup     |

Nexgou is not trying to replace small routers for tiny handlers. It is for applications where structure, testing, and repeated workflows matter.

## Benchmark Snapshot

Current benchmark data lives in [benchmark](benchmark) and [benchmark/RESULT_HTTP_PERFORMANCE.md](benchmark/RESULT_HTTP_PERFORMANCE.md).

| Rank | Service       | Stack                    | Avg req/s |  Avg p95 | Errors |
| ---: | ------------- | ------------------------ | --------: | -------: | -----: |
|    1 | `actix-web`   | Rust + Actix Web         | 21,788.89 | 15.37 ms |     0% |
|    2 | `hyper`       | Rust + Hyper             | 21,453.22 | 15.10 ms |     0% |
|    3 | `nexgou`      | Go + NexGou              | 20,102.40 | 16.84 ms |     0% |
|    4 | `vert-x`      | Java + Vert.x Web        | 17,285.19 | 19.40 ms |     0% |
|    5 | `asp-kestrel` | ASP.NET Core Minimal API | 16,124.23 | 26.59 ms |     0% |
|    6 | `fastify`     | Node.js + Fastify        |  9,182.02 | 32.95 ms |     0% |
|    7 | `ajax-php`    | PHP + PDO SQLite         |  2,384.51 | 84.85 ms |     0% |

Environment: Docker Compose on Windows, k6, 200 VUs, 30s per scenario, 4 CPU and 2 GB RAM per service.

## Install

```bash
go get github.com/nexgou/server
```

Requires Go 1.25 or newer.

## Ecosystem Modules

Public modules are available from the [nexgou GitHub organization](https://github.com/orgs/nexgou/repositories). Import paths follow `github.com/nexgou/<module>`.

| Module | Purpose | Import |
| --- | --- | --- |
| [server](https://github.com/nexgou/server) | Core framework | `github.com/nexgou/server` |
| [caching](https://github.com/nexgou/caching) | Cache abstraction | `github.com/nexgou/caching` |
| [compression](https://github.com/nexgou/compression) | HTTP compression | `github.com/nexgou/compression` |
| [cookie](https://github.com/nexgou/cookie) | Cookie helpers | `github.com/nexgou/cookie` |
| [cron](https://github.com/nexgou/cron) | Scheduled jobs | `github.com/nexgou/cron` |
| [database](https://github.com/nexgou/database) | Database base module | `github.com/nexgou/database` |
| [events](https://github.com/nexgou/events) | Event emitter | `github.com/nexgou/events` |
| [fileupload](https://github.com/nexgou/fileupload) | File uploads | `github.com/nexgou/fileupload` |
| [graphql](https://github.com/nexgou/graphql) | GraphQL integration | `github.com/nexgou/graphql` |
| [grpc](https://github.com/nexgou/grpc) | gRPC transport | `github.com/nexgou/grpc` |
| [jwt](https://github.com/nexgou/jwt) | JWT auth | `github.com/nexgou/jwt` |
| [mongo](https://github.com/nexgou/mongo) | MongoDB integration | `github.com/nexgou/mongo` |
| [mqtt](https://github.com/nexgou/mqtt) | MQTT integration | `github.com/nexgou/mqtt` |
| [nats](https://github.com/nexgou/nats) | NATS integration | `github.com/nexgou/nats` |
| [postgres](https://github.com/nexgou/postgres) | PostgreSQL integration | `github.com/nexgou/postgres` |
| [queues](https://github.com/nexgou/queues) | Queue abstraction | `github.com/nexgou/queues` |
| [rabbitmq](https://github.com/nexgou/rabbitmq) | RabbitMQ integration | `github.com/nexgou/rabbitmq` |
| [redis](https://github.com/nexgou/redis) | Redis integration | `github.com/nexgou/redis` |
| [scheduler](https://github.com/nexgou/scheduler) | Task scheduling | `github.com/nexgou/scheduler` |
| [serialization](https://github.com/nexgou/serialization) | Serialization helpers | `github.com/nexgou/serialization` |
| [sqlite](https://github.com/nexgou/sqlite) | SQLite integration | `github.com/nexgou/sqlite` |
| [sqs](https://github.com/nexgou/sqs) | AWS SQS integration | `github.com/nexgou/sqs` |
| [streaming](https://github.com/nexgou/streaming) | Streaming helpers | `github.com/nexgou/streaming` |
| [validation](https://github.com/nexgou/validation) | Validation module | `github.com/nexgou/validation` |
| [websocket](https://github.com/nexgou/websocket) | WebSocket transport | `github.com/nexgou/websocket` |

## Quick Start

```go
package main

import (
    "log"
    "time"

    nexgou "github.com/nexgou/server"
    "github.com/nexgou/server/src/filter"
    "github.com/nexgou/server/src/middleware"
)

var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Controllers: []any{NewUserController},
})

type UserController struct{}

func NewUserController() *UserController { return &UserController{} }

func (c *UserController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.Get("/users", c.FindAll).Version("v1"),
    }
}

func (c *UserController) FindAll(ctx *nexgou.Context) error {
    return ctx.JSON(200, nexgou.H{"users": []string{"alice", "bob"}})
}

func main() {
    app := nexgou.CreateApp(AppModule)
    app.Use(middleware.Recovery())
    app.Use(middleware.SecurityHeaders())
    app.Use(middleware.RateLimit(100, time.Minute))
    app.SetFilter(&filter.HttpExceptionFilter{})

    log.Fatal(app.Listen(3000))
}
```

Try the complete samples in [samples/api](samples/api) and [samples/taskboard](samples/taskboard).

## Documentation

- [Getting Started](docs/en/getting-started.md)
- [Modules](docs/en/modules.md)
- [Controllers](docs/en/controllers.md)
- [Middleware](docs/en/middleware.md)
- [Security](docs/en/security.md)
- [Config](docs/en/config.md)
- [Logger](docs/en/logger.md)
- [Testing](docs/en/testing.md)
- [Benchmark](docs/BENCHMARK.md)

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

For security issues, use [SECURITY.md](SECURITY.md).

## License

Nexgou is released under the [MIT License](LICENSE).
