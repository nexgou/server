# Nexgou Ecosystem Modules

Nexgou is organized as a **modular ecosystem**. The core framework (`github.com/nexgou/server`) ships built-in modules for the most common concerns, and additional first-party modules are published as independent packages under the [`nexgou`](https://github.com/nexgou) GitHub organization.

Each module follows the same convention: import it, add it to `Imports` in any `nexgou.Module(...)`, and its services become available for Dependency Injection.

---

## Table of Contents

| Module | Import path | Category |
|---|---|---|
| [server](#server) | `github.com/nexgou/server` | Core |
| [validation](#validation) | `github.com/nexgou/validation` | Core |
| [caching](#caching) | `github.com/nexgou/caching` | Data |
| [database](#database) | `github.com/nexgou/database` | Data |
| [mongo](#mongo) | `github.com/nexgou/mongo` | Data |
| [postgres](#postgres) | `github.com/nexgou/postgres` | Data |
| [redis](#redis) | `github.com/nexgou/redis` | Data |
| [sqlite](#sqlite) | `github.com/nexgou/sqlite` | Data |
| [compression](#compression) | `github.com/nexgou/compression` | HTTP |
| [cookie](#cookie) | `github.com/nexgou/cookie` | HTTP |
| [fileupload](#fileupload) | `github.com/nexgou/fileupload` | HTTP |
| [graphql](#graphql) | `github.com/nexgou/graphql` | HTTP |
| [streaming](#streaming) | `github.com/nexgou/streaming` | HTTP |
| [jwt](#jwt) | `github.com/nexgou/jwt` | Security |
| [cron](#cron) | `github.com/nexgou/cron` | Scheduling |
| [scheduler](#scheduler) | `github.com/nexgou/scheduler` | Scheduling |
| [events](#events) | `github.com/nexgou/events` | Messaging |
| [mqtt](#mqtt) | `github.com/nexgou/mqtt` | Messaging |
| [nats](#nats) | `github.com/nexgou/nats` | Messaging |
| [queues](#queues) | `github.com/nexgou/queues` | Messaging |
| [rabbitmq](#rabbitmq) | `github.com/nexgou/rabbitmq` | Messaging |
| [serialization](#serialization) | `github.com/nexgou/serialization` | Messaging |
| [sqs](#sqs) | `github.com/nexgou/sqs` | Messaging |

---

## Core

### server

**`github.com/nexgou/server`**

The main framework package. Provides the application bootstrap, HTTP engine, module system, IoC container, Dependency Injection, Controllers, Guards, Interceptors, Pipes, Exception Filters, WebSocket, SSE, gRPC, `ConfigModule`, `LogModule`, and the `nexgoutest` testing utilities.

This is the only required dependency — all other modules in this list are optional add-ons.

```go
go get github.com/nexgou/server
```

---

### validation

**`github.com/nexgou/validation`**

Struct validation via field tags. Wraps `go-playground/validator/v10` and returns structured per-field errors that are API-friendly.

Use it to:
- Validate incoming request bodies (DTOs) before processing them.
- Enforce business rules on structs with tags (`required`, `email`, `min`, `max`, `oneof`, `uuid`, `url`, …).
- Register custom named validation functions.

```go
go get github.com/nexgou/validation
```

```go
var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{nexgou.ValidationModule},
})
```

---

## Data

### caching

**`github.com/nexgou/caching`**

Generic, backend-agnostic caching layer. Provides a `CacheService` with `Get`, `Set`, `Delete`, and `Clear` operations.

Use it to:
- Cache expensive computations or database query results.
- Store session data with TTL.
- Switch cache backends (in-memory, Redis, …) without changing application code.

```go
go get github.com/nexgou/caching
```

---

### database

**`github.com/nexgou/database`**

Database-agnostic abstraction layer built on top of `database/sql`. Provides connection pooling, transaction helpers, and query builders compatible with any `database/sql` driver.

Use it to:
- Manage SQL connections with a single injectable `DatabaseService`.
- Run queries and transactions in a consistent way across different SQL engines.
- Switch between SQLite, PostgreSQL, MySQL, and others by changing the driver.

```go
go get github.com/nexgou/database
```

---

### mongo

**`github.com/nexgou/mongo`**

MongoDB integration module. Wraps the official `mongo-driver` and exposes a `MongoService` ready for Dependency Injection.

Use it to:
- Connect to MongoDB clusters with a typed configuration.
- Access collections through an injectable service.
- Use change streams and aggregation pipelines within the Nexgou lifecycle.

```go
go get github.com/nexgou/mongo
```

---

### postgres

**`github.com/nexgou/postgres`**

PostgreSQL integration module. Wraps `pgx` and exposes a `PostgresService` with connection pooling, prepared statements, and typed query helpers.

Use it to:
- Connect to PostgreSQL with full `pgx` performance.
- Run parameterized queries, batch operations, and COPY FROM.
- Use PostgreSQL-specific features (LISTEN/NOTIFY, advisory locks, JSONB).

```go
go get github.com/nexgou/postgres
```

---

### redis

**`github.com/nexgou/redis`**

Redis integration module. Wraps `go-redis` and exposes a `RedisService` for key-value operations, pub/sub, streams, and distributed locking.

Use it to:
- Store and retrieve cached values with TTL.
- Implement pub/sub communication between services.
- Use Redis Streams for event sourcing or task queues.
- Distributed locks and rate limiting backed by Redis.

```go
go get github.com/nexgou/redis
```

---

### sqlite

**`github.com/nexgou/sqlite`**

SQLite integration module. Wraps `modernc.org/sqlite` (pure-Go, no CGO) and exposes a `SQLiteService` for embedded local databases.

Use it to:
- Embed a fully-featured SQL database with zero external dependencies.
- Run development environments or edge deployments without a database server.
- Store configuration, state, or local event logs on disk.

```go
go get github.com/nexgou/sqlite
```

---

## HTTP

### compression

**`github.com/nexgou/compression`**

Response compression middleware and service. Supports gzip, deflate, and brotli encoding with configurable thresholds and MIME-type filters.

Use it to:
- Reduce response payload size automatically.
- Configure compression per-route or globally.
- Control which content types and minimum sizes trigger compression.

```go
go get github.com/nexgou/compression
```

---

### cookie

**`github.com/nexgou/cookie`**

Typed cookie management. Provides a `CookieService` for reading, writing, and deleting HTTP cookies with support for signed and encrypted values.

Use it to:
- Set and read HTTP cookies in a type-safe way.
- Sign cookies to prevent client-side tampering.
- Encrypt cookie values for sensitive data.
- Manage cookie attributes (SameSite, HttpOnly, Secure, MaxAge).

```go
go get github.com/nexgou/cookie
```

---

### fileupload

**`github.com/nexgou/fileupload`**

Multipart file upload handling. Provides a `FileUploadService` with size limits, MIME-type validation, local and remote storage backends.

Use it to:
- Accept single or multiple file uploads in HTTP handlers.
- Validate file types and sizes before processing.
- Stream files directly to object storage (S3, GCS, local disk).
- Generate upload URLs and metadata.

```go
go get github.com/nexgou/fileupload
```

---

### graphql

**`github.com/nexgou/graphql`**

GraphQL server integration. Registers a `GraphQLController` that serves a schema defined with Go code (schema-first or code-first) and optionally exposes a GraphiQL playground.

Use it to:
- Build a GraphQL API alongside (or instead of) REST endpoints.
- Define queries, mutations, and subscriptions.
- Reuse existing Nexgou services and guards in resolvers.
- Enable real-time GraphQL subscriptions over WebSocket.

```go
go get github.com/nexgou/graphql
```

---

### streaming

**`github.com/nexgou/streaming`**

Chunked HTTP response streaming utilities. Provides helpers for streaming large datasets, NDJSON, CSV, or arbitrary byte streams without loading everything into memory.

Use it to:
- Stream large query results to the client incrementally.
- Export large reports or CSV files on-the-fly.
- Send NDJSON (newline-delimited JSON) for log or event feeds.
- Avoid memory spikes when serving large payloads.

```go
go get github.com/nexgou/streaming
```

---

## Security

### jwt

**`github.com/nexgou/jwt`**

JWT (JSON Web Token) authentication module. Provides a `JWTService` for signing and verifying tokens, plus a ready-to-use `JWTGuard` for protecting routes.

Use it to:
- Issue access and refresh tokens with configurable expiry.
- Verify and decode JWTs in guards or interceptors.
- Support RS256, ES256, and HS256 signing algorithms.
- Attach decoded claims to `nexgou.Context` for downstream handlers.

```go
go get github.com/nexgou/jwt
```

```go
// Protect a route with JWTGuard
nexgou.Get("/profile", c.Profile).Guard(&jwt.JWTGuard{})
```

---

## Scheduling

### cron

**`github.com/nexgou/cron`**

Cron-expression based task scheduler. Provides a `CronService` for registering functions that run on a schedule defined by a standard cron expression.

Use it to:
- Run recurring background jobs (e.g. nightly reports, cache warming).
- Define schedules with standard 5 or 6-field cron expressions.
- Start and stop individual jobs at runtime.

```go
go get github.com/nexgou/cron
```

```go
cron.Schedule("0 0 * * *", func() { /* nightly job */ })
```

---

### scheduler

**`github.com/nexgou/scheduler`**

Flexible task scheduler with support for interval-based, delayed, and one-shot jobs. Complements `cron` when you need dynamic runtime scheduling rather than static cron expressions.

Use it to:
- Schedule a task to run after a delay (`RunAfter(5*time.Minute, fn)`).
- Run a task at a fixed interval (`Every(10*time.Second, fn)`).
- Cancel or reschedule jobs at runtime.
- Manage task queues with concurrency limits.

```go
go get github.com/nexgou/scheduler
```

---

## Messaging

### events

**`github.com/nexgou/events`**

In-process event bus. Provides an `EventEmitter` for publishing and subscribing to typed events within a single application instance.

Use it to:
- Decouple services with a publish/subscribe pattern.
- Emit domain events from one module and handle them in another.
- Replace direct service calls with async, fire-and-forget notifications.

```go
go get github.com/nexgou/events
```

```go
emitter.Emit("user.created", UserCreatedEvent{ID: "123"})
emitter.On("user.created", func(e UserCreatedEvent) { /* handler */ })
```

---

### mqtt

**`github.com/nexgou/mqtt`**

MQTT messaging module. Connects to any MQTT broker (Mosquitto, HiveMQ, AWS IoT, …) and exposes a `MQTTService` for publishing and subscribing to topics.

Use it to:
- Connect IoT devices or edge nodes via MQTT.
- Publish sensor data or commands to topics.
- Subscribe to wildcard topics and process incoming messages in handlers.
- Handle QoS levels (0, 1, 2) and persistent sessions.

```go
go get github.com/nexgou/mqtt
```

---

### nats

**`github.com/nexgou/nats`**

NATS messaging module. Wraps the official `nats.go` client and provides a `NATSService` for pub/sub, request/reply, and JetStream persistence.

Use it to:
- Broadcast events across multiple service instances.
- Implement request/reply microservice communication over NATS.
- Use JetStream for durable, at-least-once message delivery.
- Replace HTTP calls between internal services with fast NATS messages.

```go
go get github.com/nexgou/nats
```

---

### queues

**`github.com/nexgou/queues`**

Backend-agnostic job queue abstraction. Provides a `QueueService` interface that works with multiple backends (Redis, RabbitMQ, SQS, in-memory) through a common API.

Use it to:
- Offload slow or heavy tasks to background workers.
- Produce and consume jobs with retries and dead-letter handling.
- Switch queue backends without changing producer or consumer code.
- Implement fan-out, delayed jobs, and priority queues.

```go
go get github.com/nexgou/queues
```

---

### rabbitmq

**`github.com/nexgou/rabbitmq`**

RabbitMQ integration module. Wraps `amqp091-go` and exposes a `RabbitMQService` for publishing and consuming AMQP messages with exchange/queue topology management.

Use it to:
- Publish messages to exchanges (direct, fanout, topic, headers).
- Consume messages with manual or automatic acknowledgment.
- Declare exchanges, queues, and bindings through code.
- Implement RPC patterns over RabbitMQ.

```go
go get github.com/nexgou/rabbitmq
```

---

### serialization

**`github.com/nexgou/serialization`**

Multi-format serialization and deserialization. Provides a `SerializationService` supporting JSON, MessagePack, Protocol Buffers, and CBOR through a unified interface.

Use it to:
- Encode/decode data between formats without changing business logic.
- Serialize messages before publishing to a queue or broker.
- Choose the most efficient wire format per use case (JSON for APIs, MessagePack for queues, Protobuf for gRPC).

```go
go get github.com/nexgou/serialization
```

---

### sqs

**`github.com/nexgou/sqs`**

AWS SQS integration module. Wraps the AWS SDK v2 SQS client and exposes a `SQSService` for sending, receiving, and deleting messages from SQS queues.

Use it to:
- Send messages to SQS standard or FIFO queues.
- Poll and process messages in a worker loop.
- Handle visibility timeouts, retries, and dead-letter queues.
- Integrate with AWS Lambda triggers or EC2-based workers in the same codebase.

```go
go get github.com/nexgou/sqs
```

---

## Using multiple modules together

```go
var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{
        nexgou.ConfigModule,
        nexgou.LogModule,
        nexgou.ValidationModule,   // github.com/nexgou/validation
        postgres.PostgresModule,   // github.com/nexgou/postgres
        redis.RedisModule,         // github.com/nexgou/redis
        jwt.JWTModule,             // github.com/nexgou/jwt
        cron.CronModule,           // github.com/nexgou/cron
        events.EventsModule,       // github.com/nexgou/events
        UserModule,
        OrderModule,
    },
})
```

Each module exposes its services automatically through DI — inject them in any controller or service constructor without manual wiring.
