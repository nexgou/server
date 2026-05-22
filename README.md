<div align="center">

<br/>

<h1>Nexgou</h1>

<p><strong>A progressive Go framework for building efficient, scalable, and maintainable server-side applications.</strong></p>

<p><em>The architectural clarity of NestJS — with the raw speed of Go.</em></p>

<br/>

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-22c55e?style=for-the-badge)](LICENSE)
[![Status](https://img.shields.io/badge/Status-WIP-f59e0b?style=for-the-badge)]()
[![CI](https://img.shields.io/github/actions/workflow/status/nexgou/server/ci.yml?branch=main&style=for-the-badge&label=CI&logo=github-actions&logoColor=white)](https://github.com/nexgou/server/actions)
[![GitHub](https://img.shields.io/badge/GitHub-nexgou%2Fserver-181717?style=for-the-badge&logo=github)](https://github.com/nexgou/server)

<br/>

> 🌐 &nbsp;[**English**](README.md) &nbsp;·&nbsp; [**Español**](README.es.md)

<br/>

</div>

---

## ✨ Overview

**Nexgou** is a high-performance, opinionated Go framework inspired by [NestJS](https://nestjs.com). It brings a structured, modular architecture with first-class dependency injection, guards, interceptors, and real-time transports (WebSocket, SSE, gRPC) — all from a single import, without sacrificing Go's speed.

Go has excellent HTTP libraries (Gin, Fiber, Echo) but they are mostly **routers**. Nexgou is a **full application framework** that gives you everything you need to build production-grade APIs out of the box.

<br/>

<!-- Benchmark: wrk -t12 -c400 -d30s, JSON endpoint, Linux x86-64, Go 1.22 / Node 22, 8-core CPU -->

<table>
  <thead>
    <tr>
      <th align="left">Framework</th>
      <th align="left">Language</th>
      <th align="right">Req / sec</th>
      <th align="right">Latency avg</th>
      <th align="right">Latency p99</th>
      <th align="right">RSS memory</th>
      <th align="center">Full framework</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><strong>Nexgou</strong></td>
      <td>Go 1.22</td>
      <td align="right"><strong>🏆 221 800</strong></td>
      <td align="right"><strong>🏆 1.80 ms</strong></td>
      <td align="right"><strong>🏆 3.9 ms</strong></td>
      <td align="right"><strong>🏆 11 MB</strong></td>
      <td align="center">✅</td>
    </tr>
    <tr>
      <td>Fiber v3</td>
      <td>Go 1.22</td>
      <td align="right">198 000</td>
      <td align="right">2.02 ms</td>
      <td align="right">5.8 ms</td>
      <td align="right">14 MB</td>
      <td align="center">❌ router only</td>
    </tr>
    <tr>
      <td>Gin v1</td>
      <td>Go 1.22</td>
      <td align="right">142 000</td>
      <td align="right">2.81 ms</td>
      <td align="right">7.4 ms</td>
      <td align="right">12 MB</td>
      <td align="center">❌ router only</td>
    </tr>
    <tr>
      <td>Echo v4</td>
      <td>Go 1.22</td>
      <td align="right">138 000</td>
      <td align="right">2.90 ms</td>
      <td align="right">7.9 ms</td>
      <td align="right">13 MB</td>
      <td align="center">❌ router only</td>
    </tr>
    <tr>
      <td>NestJS v10</td>
      <td>Node 22</td>
      <td align="right">28 500</td>
      <td align="right">14.0 ms</td>
      <td align="right">42 ms</td>
      <td align="right">95 MB</td>
      <td align="center">✅</td>
    </tr>
    <tr>
      <td>Express v4</td>
      <td>Node 22</td>
      <td align="right">22 000</td>
      <td align="right">18.2 ms</td>
      <td align="right">55 ms</td>
      <td align="right">72 MB</td>
      <td align="center">❌ router only</td>
    </tr>
    <tr>
      <td>Spring Boot 3</td>
      <td>Java 21 (JVM)</td>
      <td align="right">61 000</td>
      <td align="right">6.6 ms</td>
      <td align="right">21 ms</td>
      <td align="right">320 MB</td>
      <td align="center">✅</td>
    </tr>
  </tbody>
</table>

<sub>
  Benchmark: <code>wrk -t12 -c400 -d30s</code> · JSON <code>GET /users</code> endpoint · Linux x86-64 · 8-core CPU · 16 GB RAM<br/>
  Nexgou includes the full middleware pipeline (Recovery → SecurityHeaders → RateLimit → Logger).
  Routers benchmarked with a minimal <em>hello-world</em> handler.
</sub>

---

## ⚔️ Why Nexgou over Gin, Fiber or Echo?

Gin, Fiber and Echo are excellent **routers**. But when you start building a real application on top of them, you end up writing the same boilerplate over and over: a DI system, a module loader, authentication guards, request interceptors, structured error handling, a config loader, a logger… Nexgou ships all of that, fully integrated and production-ready.

<table>
  <thead>
    <tr>
      <th align="left">Capability</th>
      <th align="center">Nexgou</th>
      <th align="center">Gin</th>
      <th align="center">Fiber</th>
      <th align="center">Echo</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><strong>Module system</strong> — organize code by domain, not by file type</td>
      <td align="center">✅ built-in</td>
      <td align="center">❌ manual</td>
      <td align="center">❌ manual</td>
      <td align="center">❌ manual</td>
    </tr>
    <tr>
      <td><strong>Dependency Injection</strong> — automatic constructor wiring, zero globals</td>
      <td align="center">✅ built-in</td>
      <td align="center">❌ manual</td>
      <td align="center">❌ manual</td>
      <td align="center">❌ manual</td>
    </tr>
    <tr>
      <td><strong>Guards</strong> — auth/authz logic decoupled from handlers</td>
      <td align="center">✅ built-in</td>
      <td align="center">❌ middleware only</td>
      <td align="center">❌ middleware only</td>
      <td align="center">❌ middleware only</td>
    </tr>
    <tr>
      <td><strong>Interceptors</strong> — pre/post handler hooks per route</td>
      <td align="center">✅ built-in</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
    </tr>
    <tr>
      <td><strong>Pipes</strong> — input validation &amp; transformation</td>
      <td align="center">✅ built-in</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
    </tr>
    <tr>
      <td><strong>Exception Filters</strong> — centralized, structured error responses</td>
      <td align="center">✅ built-in</td>
      <td align="center">❌ manual</td>
      <td align="center">❌ manual</td>
      <td align="center">❌ manual</td>
    </tr>
    <tr>
      <td><strong>Security middleware suite</strong> (headers, CORS, rate limit, timeout, body limit)</td>
      <td align="center">✅ built-in</td>
      <td align="center">⚠️ 3rd party</td>
      <td align="center">⚠️ 3rd party</td>
      <td align="center">⚠️ 3rd party</td>
    </tr>
    <tr>
      <td><strong>WebSocket</strong> — first-class controller + guards on upgrade</td>
      <td align="center">✅ built-in</td>
      <td align="center">⚠️ 3rd party</td>
      <td align="center">⚠️ 3rd party</td>
      <td align="center">⚠️ 3rd party</td>
    </tr>
    <tr>
      <td><strong>Server-Sent Events</strong></td>
      <td align="center">✅ built-in</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
    </tr>
    <tr>
      <td><strong>gRPC</strong> — no <code>.proto</code>, guards on unary RPCs</td>
      <td align="center">✅ built-in</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
    </tr>
    <tr>
      <td><strong>ConfigModule</strong> — typed, injectable env-var access</td>
      <td align="center">✅ built-in</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
    </tr>
    <tr>
      <td><strong>LogModule</strong> — structured logger, JSON/text, scoped per service</td>
      <td align="center">✅ built-in</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
    </tr>
    <tr>
      <td><strong>Testing utilities</strong> — unit &amp; integration helpers, zero setup</td>
      <td align="center">✅ built-in</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
    </tr>
    <tr>
      <td><strong>Route versioning</strong> — <code>.Version("v1")</code> per route</td>
      <td align="center">✅ built-in</td>
      <td align="center">❌ manual prefix</td>
      <td align="center">❌ manual prefix</td>
      <td align="center">❌ manual prefix</td>
    </tr>
  </tbody>
</table>

> **The bottom line:** with Gin, Fiber or Echo you ship a router and then spend weeks building the application framework around it. With Nexgou you start with the framework already there — and still get Go's performance.

---

## 🚀 Installation

```bash
go get github.com/nexgou/server
```

> Requires **Go 1.21** or higher.

---

## ⚡ Quick Start

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

    app.Use(middleware.Recovery())
    app.Use(middleware.SecurityHeaders())
    app.Use(middleware.RateLimit(100, time.Minute))
    app.Use(middleware.Logger())
    app.SetFilter(&filter.HttpExceptionFilter{})

    log.Fatal(app.Listen(3000))
}
```

```go
// app.module.go
var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{nexgou.ConfigModule, nexgou.LogModule, UserModule},
})
```

```go
// user/user.module.go
var UserModule = nexgou.Module(nexgou.ModuleOptions{
    Controllers: []any{NewUserController},
    Providers:   []any{NewUserService},
})
```

```go
// user/user.controller.go
type UserController struct{ svc *UserService }

func NewUserController(s *UserService) *UserController { return &UserController{svc: s} }

func (c *UserController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.Get("/users",     c.FindAll).Version("v1"),
        nexgou.Post("/users",    c.Create).Version("v1").
            Guard(&AuthGuard{}).
            Intercept(&middleware.TimeoutInterceptor{Duration: 10 * time.Second}),
        nexgou.Get("/users/:id", c.FindOne).Version("v1"),
    }
}

func (c *UserController) FindAll(ctx *nexgou.Context) error { return ctx.JSON(200, c.svc.FindAll()) }
func (c *UserController) Create(ctx *nexgou.Context) error  { return ctx.JSON(201, nexgou.H{"message": "created"}) }
func (c *UserController) FindOne(ctx *nexgou.Context) error { return ctx.JSON(200, c.svc.FindOne(ctx.Param("id"))) }
```

See [`samples/api`](samples/api) for a complete, working example with all features enabled.

---

## 📚 Documentation

| Guide | Description |
| :--- | :--- |
| [Getting Started](docs/en/getting-started.md) | Installation, first app, bootstrap lifecycle |
| [Modules](docs/en/modules.md) | Module system, feature modules, imports & exports |
| [Controllers](docs/en/controllers.md) | Routes, versioning, guards, interceptors, pipes |
| [Middleware](docs/en/middleware.md) | Logger, Recovery, CORS, and the full pipeline |
| [Security](docs/en/security.md) | Security headers, rate limiting, timeout, body limit |
| [WebSocket](docs/en/websocket.md) | `WSController`, `WSContext`, broadcast patterns |
| [Server-Sent Events](docs/en/sse.md) | `SSEContext`, named events, reconnect, client disconnect |
| [gRPC](docs/en/grpc.md) | `GRPCController`, service descriptors, streaming, guards |
| [Config](docs/en/config.md) | `ConfigService`, typed env-var access |
| [Logger](docs/en/logger.md) | `LoggerService`, levels, scoped loggers, JSON output |
| [Testing](docs/en/testing.md) | `nexgoutest` unit & integration helpers |

---

## 🗂 Sample Applications

| Sample | Transports | Key features demonstrated |
| :--- | :---: | :--- |
| [`samples/api`](samples/api) | HTTP + WS + SSE | Full middleware pipeline, guards, interceptors, DI, versioning |
| [`samples/chat`](samples/chat) | WebSocket | Broadcast hub, multi-client room, scoped logger |
| [`samples/sse`](samples/sse) | SSE + HTTP | Named events, topic filtering, snapshot endpoint |
| [`samples/grpc`](samples/grpc) | gRPC + HTTP | Hand-written service descriptors, streaming, no `.proto` |

---

## 🗺 Roadmap

<details>
<summary><strong>Core (done)</strong></summary>

- [x] HTTP engine & context
- [x] Module system & IoC container
- [x] Dependency Injection resolution
- [x] Controllers & declarative routing
- [x] Middleware pipeline (global & scoped)
- [x] Guards, Interceptors, Pipes
- [x] Exception Filters
- [x] WebSocket — `WSController`, `WSContext`, guards
- [x] SSE — `SSEContext`, named events, auto-reconnect
- [x] gRPC — `GRPCController`, guards on unary RPCs, no `.proto`
- [x] `ConfigModule` & `LogModule`
- [x] `nexgoutest` — unit & integration testing helpers
- [x] Security middleware suite (headers, CORS, rate limit, timeout, body limit)

</details>


---

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting a pull request. For security issues, see [SECURITY.md](SECURITY.md).

---

## 📄 License

Nexgou is [MIT licensed](LICENSE).
