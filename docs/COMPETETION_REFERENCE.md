# Referencia de competencia: servidores y frameworks HTTP

Documento de referencia para analizar servidores/frameworks HTTP de distintos lenguajes antes de diseñar un framework propio en Go.

> **Importante:** los valores de `req/s` son orientativos. No todos los frameworks se han medido en el mismo hardware, con el mismo test, runtime, payload o configuración. Úsalos como referencia inicial, no como ranking absoluto.

---

## Objetivo

Comparar competidores relevantes para extraer ideas útiles en:

- rendimiento bruto;
- diseño de routing;
- middlewares;
- arquitectura modular;
- developer experience;
- validación;
- observabilidad;
- extensibilidad;
- compatibilidad con ecosistemas existentes.

El objetivo no es copiar uno, sino combinar lo mejor de varios mundos:

```txt
Go performance + ASP.NET Core architecture + Fastify/NestJS developer experience
```

---

## Ranking general: mejor referencia para crear un framework

Este ranking prioriza valor arquitectónico + rendimiento + ecosistema + ideas aprovechables.

| Rank | Servidor / framework    | Lenguaje                 |                             Req/s orientativo | Tipo                          | Valor principal                                             |
| ---: | ----------------------- | ------------------------ | --------------------------------------------: | ----------------------------- | ----------------------------------------------------------- |
|    1 | ASP.NET Core / Kestrel  | C#                       |    1.000.000+ req/s en benchmarks optimizados | Runtime/framework             | Arquitectura, DI, middleware pipeline, hosting, lifecycle   |
|    2 | fasthttp                | Go                       |                                 334.184 req/s | HTTP server/lib               | Rendimiento extremo en Go, bajo overhead, pooling           |
|    3 | actix-web               | Rust                     |                                 294.315 req/s | Framework                     | Rendimiento alto con estructura completa                    |
|    4 | hyper                   | Rust                     |                                 299.736 req/s | HTTP lib                      | Diseño HTTP low-level, async, separación protocolo/app      |
|    5 | Vert.x                  | Java                     |      500.000+ req/s en escenarios optimizados | Toolkit reactivo              | Event loop, concurrencia, arquitectura modular              |
|    6 | uWebSockets.js          | C++ / JavaScript binding |               137.000+ req/s en tests simples | HTTP/WebSocket server         | Realtime extremo, muchas conexiones, bajo consumo           |
|    7 | Fastify                 | TypeScript / JavaScript  | 46.664 req/s en benchmark oficial de overhead | Framework                     | Plugins, schemas, validación, DX                            |
|    8 | Phoenix / Cowboy        | Elixir / Erlang VM       |                100.000+ req/s según escenario | Framework/server              | Concurrencia, tolerancia a fallos, supervisión              |
|    9 | FastAPI / Uvicorn       | Python                   |             20.000 – 80.000 req/s según stack | Framework/API server          | Typing, OpenAPI, validación y productividad                 |
|   10 | Laravel Octane / Swoole | PHP                      |            50.000 – 100.000+ req/s según caso | Framework/runtime persistente | Workers persistentes, productividad, modelo app persistente |

---

## Ranking solo por rendimiento bruto aproximado

Este ranking es menos útil para diseñar arquitectura, pero sirve para tener una referencia visual de rendimiento.

| Rank | Servidor / framework    | Lenguaje                |       Req/s orientativo |
| ---: | ----------------------- | ----------------------- | ----------------------: |
|    1 | ASP.NET Core / Kestrel  | C#                      |        1.000.000+ req/s |
|    2 | Vert.x                  | Java                    |          500.000+ req/s |
|    3 | fasthttp                | Go                      |           334.184 req/s |
|    4 | hyper                   | Rust                    |           299.736 req/s |
|    5 | actix-web               | Rust                    |           294.315 req/s |
|    6 | uWebSockets.js          | C++ / JavaScript        |          137.000+ req/s |
|    7 | Phoenix / Cowboy        | Elixir / Erlang         |          100.000+ req/s |
|    8 | Laravel Octane / Swoole | PHP                     | 50.000 – 100.000+ req/s |
|    9 | Fastify                 | TypeScript / JavaScript |           46.664+ req/s |
|   10 | FastAPI / Uvicorn       | Python                  |   20.000 – 80.000 req/s |

---

## Tabla comparativa por lenguaje

| Lenguaje   | Referencia principal            | Qué estudiar                                               |
| ---------- | ------------------------------- | ---------------------------------------------------------- |
| Go         | net/http, fasthttp, Chi, Gin    | Compatibilidad, router, middleware chain, bajo overhead    |
| Rust       | actix-web, hyper, axum          | Tipado fuerte, extractors, async, separación por capas     |
| C#         | ASP.NET Core / Kestrel          | DI, hosting, lifecycle, configuration, middleware pipeline |
| Java       | Vert.x, Quarkus, Spring WebFlux | Event loop, arquitectura enterprise, reactive programming  |
| TypeScript | Fastify, NestJS, Hono           | Plugins, decorators opcionales, schemas, DX                |
| C++        | uWebSockets                     | WebSocket, HTTP extremo, memoria, realtime                 |
| Elixir     | Phoenix / Cowboy                | Supervisión, resiliencia, concurrencia masiva              |
| Python     | FastAPI / Uvicorn               | OpenAPI-first, validación, typing                          |
| PHP        | Laravel Octane / Swoole         | Workers persistentes, productividad                        |
| Ruby       | Falcon / Puma                   | Simplicidad, concurrencia moderna                          |

---

## Qué copiar de cada uno

### ASP.NET Core / Kestrel

Copiar:

- pipeline de middlewares;
- dependency injection integrada;
- lifecycle hooks;
- configuración unificada;
- hosting model;
- separación entre framework, runtime y aplicación;
- observabilidad integrada.

Idea aplicable en Go:

```go
type Module interface {
    Configure(app *Application) error
    OnStart(ctx context.Context) error
    OnStop(ctx context.Context) error
}
```

---

### fasthttp

Copiar:

- pooling agresivo;
- bajo número de allocations;
- separación entre request/response;
- foco en throughput;
- API orientada a performance.

Cuidado:

- no es compatible directamente con `net/http`;
- puede complicar integraciones estándar;
- mejor usarlo como adaptador opcional, no como única base.

---

### actix-web

Copiar:

- routing potente;
- extractors;
- app state;
- middleware bien integrado;
- enfoque de framework completo sin perder rendimiento.

Idea aplicable:

```go
type Extractor[T any] interface {
    Extract(ctx *Context) (T, error)
}
```

---

### hyper

Copiar:

- separación clara entre protocolo HTTP y capa de aplicación;
- streams;
- backpressure;
- HTTP/1 y HTTP/2;
- diseño low-level robusto.

Idea aplicable:

```txt
core/http-protocol != core/application-framework
```

---

### Vert.x

Copiar:

- event loop;
- modelo reactivo;
- separación por verticles/módulos;
- facilidad para sistemas distribuidos;
- buen enfoque para apps concurrentes.

---

### uWebSockets.js

Copiar:

- WebSocket como ciudadano de primera clase;
- manejo eficiente de muchas conexiones;
- bajo consumo de memoria;
- buen modelo para realtime.

Aplicación directa:

```txt
/framework/realtime
/framework/sse
/framework/websocket
```

---

### Fastify

Copiar:

- sistema de plugins;
- schemas por ruta;
- serialización rápida;
- validación integrada;
- buen developer experience.

Idea aplicable:

```go
type Plugin interface {
    Name() string
    Register(app *Application) error
}
```

---

### Phoenix / Cowboy

Copiar:

- resiliencia;
- supervisión;
- procesos aislados;
- canales realtime;
- tolerancia a fallos.

Aunque Go no tenga la BEAM, puedes aplicar ideas como:

```txt
- worker supervision
- graceful restart
- isolated background processes
- health checks por módulo
```

---

### FastAPI

Copiar:

- OpenAPI-first;
- validación clara;
- contratos tipados;
- documentación automática;
- request/response schemas.

Idea aplicable:

```go
type RouteDefinition struct {
    Method      string
    Path        string
    InputSchema  any
    OutputSchema any
}
```

---

### Laravel Octane / Swoole

Copiar:

- workers persistentes;
- arranque de aplicación una sola vez;
- reutilización de recursos;
- productividad de framework full-stack.

Cuidado:

- en Go esto ya es natural porque el proceso suele ser persistente;
- la lección importante es la gestión del estado compartido.

---

## Recomendación para tu framework en Go

No basaría el framework exclusivamente en `fasthttp`. La mejor estrategia es crear un **core propio** con adaptadores.

### Core recomendado

```txt
/framework
  /core
    application.go
    context.go
    handler.go
    middleware.go
    module.go
    lifecycle.go

  /router
    tree.go
    route.go
    params.go

  /adapters
    /nethttp
    /fasthttp

  /validation
  /errors
  /logger
  /config
  /openapi
  /realtime
```

---

## Contratos base

### Handler

```go
type HandlerFunc func(ctx *Context) error
```

### Middleware

```go
type Middleware func(next HandlerFunc) HandlerFunc
```

### Adapter

```go
type ServerAdapter interface {
    Use(middleware Middleware)
    Handle(method string, path string, handler HandlerFunc)
    Listen(addr string) error
    Shutdown(ctx context.Context) error
}
```

### Context

```go
type Context struct {
    RequestID string
    Params    map[string]string
    Query     map[string][]string
    Locals    map[string]any
}
```

---

## Estrategia recomendada por versiones

### v0.1

Objetivo: framework usable y limpio.

```txt
- net/http adapter
- router propio
- middleware chain
- context propio
- error handler centralizado
- logger
- config/env
- graceful shutdown
```

### v0.2

Objetivo: productividad.

```txt
- modules
- plugin system
- validation
- typed request body
- response helpers
- OpenAPI básico
```

### v0.3

Objetivo: rendimiento y observabilidad.

```txt
- benchmark suite
- metrics
- tracing
- profiling
- fasthttp adapter experimental
```

### v1.0

Objetivo: framework competitivo.

```txt
- API estable
- documentación completa
- testing helpers
- CLI
- templates
- security defaults
- production examples
```

---

## Benchmark propio recomendado

No te fíes solo de benchmarks externos. Crea uno propio.

### Escenarios mínimos

```txt
1. Plain text response
2. JSON response
3. Params dinámicos
4. Query parsing
5. Body JSON parsing
6. Middleware simple
7. Middleware con auth fake
8. Error handler
9. 404 route
10. Concurrent connections
```

### Herramientas

```txt
- wrk
- autocannon
- hey
- vegeta
- k6
```

### Métricas

```txt
- req/s
- latency average
- latency p50
- latency p95
- latency p99
- memory usage
- allocations/op
- CPU usage
- error rate
- startup time
- binary size
```

Ejemplo:

```bash
wrk -t12 -c1000 -d15s http://localhost:3000/hello
```

---

## Decisión final

Para construir un framework serio en Go:

```txt
Base:
Go + net/http + router propio + middleware chain + context propio

Arquitectura:
ASP.NET Core + Fastify + NestJS

Performance:
fasthttp como adapter opcional

Realtime:
uWebSockets/Phoenix como inspiración conceptual

OpenAPI/validación:
FastAPI/Fastify como inspiración
```

La ventaja competitiva no debería ser únicamente:

```txt
"soy el más rápido"
```

Debería ser:

```txt
"soy rápido, limpio, modular, testeable, observable y productivo"
```

---

## Fuentes de referencia

- TechEmpower Framework Benchmarks: https://www.techempower.com/benchmarks/
- TechEmpower FrameworkBenchmarks GitHub: https://github.com/TechEmpower/FrameworkBenchmarks
- Issue de sunsetting / archivado de TechEmpower: https://github.com/TechEmpower/FrameworkBenchmarks/issues/10932
- Fastify benchmarks oficiales: https://fastify.io/benchmarks/
- Fastify documentación principal: https://fastify.io/
- Benchmark wrk con fasthttp, hyper, actix y httprouter: https://gist.github.com/r0mdau/ac0f416d2305e33a14d3c754b7bde27a
- Microsoft .NET performance blog: https://devblogs.microsoft.com/dotnet/performance-improvements-in-aspnet-core-8/
