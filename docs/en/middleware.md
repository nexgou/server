# Middleware

> **[← Back to README](../../README.md)**

---

## Table of Contents

- [How Middleware Works](#how-middleware-works)
- [Recommended Pipeline Order](#recommended-pipeline-order)
- [Recovery](#recovery)
- [Logger](#logger)
- [CORS](#cors)
- [Security Headers](#security-headers)
- [Rate Limiting](#rate-limiting)
- [Timeout](#timeout)
- [Body Size Limit](#body-size-limit)
- [Custom Middleware](#custom-middleware)
- [Global vs Per-Route](#global-vs-per-route)

> For security-focused configuration details (header values, CORS options, rate limit headers, etc.) see [Security](security.md).

---

## How Middleware Works

Middleware is a function that wraps a handler. It receives the next handler in the chain and can execute code before, after, or instead of it.

```go
type MiddlewareFunc func(HandlerFunc) HandlerFunc
type HandlerFunc    func(*Context) error
```

Middleware is registered on the `App` and executed **in registration order** on every request:

```go
app.Use(middleware.Recovery())   // runs first
app.Use(middleware.Logger())     // runs second
// ...handler runs last
```

The execution flow for a request with two middlewares:

```
Request
  → Recovery.before
    → Logger.before
      → Handler
    → Logger.after
  → Recovery.after
Response
```

If any middleware or handler returns an error, it propagates up the chain. The global exception filter (if set) catches it at the top level.

---

## Recommended Pipeline Order

```go
app.Use(middleware.Recovery())          // 1. Catch panics first — nothing above this
app.Use(middleware.SecurityHeaders())   // 2. Security headers on every response
app.Use(middleware.CorsWithOptions(...))// 3. CORS — must run before actual handlers
app.Use(middleware.RateLimit(...))      // 4. Reject abuse early
app.Use(middleware.Timeout(...))        // 5. Set a deadline
app.Use(middleware.BodyLimit(...))      // 6. Cap request body before reading it
app.Use(middleware.Logger())           // 7. Log after the above filters (cleaner metrics)
```

---

## Recovery

**Package:** `github.com/nexgou/server/src/middleware`

Recovers from Go panics anywhere in the handler chain and converts them to a `500 Internal Server Error` response. **Always register this first** so it covers all subsequent middleware.

```go
app.Use(middleware.Recovery())
```

Behavior:
- Calls `recover()` in a deferred function
- Returns `500` with the panic message (plain text, not JSON — pair with `HttpExceptionFilter` for JSON)
- Logs the panic to stderr with a stack trace

---

## Logger

**Package:** `github.com/nexgou/server/src/middleware`

Logs every HTTP request with method, path, status code, and duration.

```go
app.Use(middleware.Logger())
```

Sample output (colored in terminals):

```
[Nexgou] GET /users 200 1.23ms
[Nexgou] POST /users 201 456µs
[Nexgou] GET /users/999 404 89µs
```

The built-in logger uses the framework's internal colorized writer. If you need structured JSON logging, use [LogModule](logger.md) together with a custom middleware.

---

## CORS

**Package:** `github.com/nexgou/server/src/middleware`

### Simple CORS (allow all)

```go
app.Use(middleware.Cors())
// Sets: Access-Control-Allow-Origin: *
```

### Configurable CORS

```go
app.Use(middleware.CorsWithOptions(middleware.CorsOptions{
    AllowedOrigins:   []string{"https://app.example.com", "https://admin.example.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
    AllowedHeaders:   []string{"Authorization", "Content-Type"},
    ExposedHeaders:   []string{"X-Request-ID"},
    AllowCredentials: true,
    MaxAge:           86400, // preflight cache: 24h
}))
```

| Option | Type | Default | Description |
|:---|:---|:---|:---|
| `AllowedOrigins` | `[]string` | `["*"]` | Allowed `Origin` values |
| `AllowedMethods` | `[]string` | common verbs | Allowed HTTP methods |
| `AllowedHeaders` | `[]string` | `["*"]` | Allowed request headers |
| `ExposedHeaders` | `[]string` | `[]` | Response headers exposed to the browser |
| `AllowCredentials` | `bool` | `false` | Allow `credentials: include` in requests |
| `MaxAge` | `int` | `0` | Preflight result cache in seconds |

Preflight `OPTIONS` requests are handled automatically with `204 No Content`.

---

## Security Headers

Sets 7 security-related HTTP response headers on every response.

```go
app.Use(middleware.SecurityHeaders())
```

Defaults:

| Header | Default Value |
|:---|:---|
| `Content-Security-Policy` | `default-src 'self'` |
| `X-Frame-Options` | `DENY` |
| `X-Content-Type-Options` | `nosniff` |
| `X-XSS-Protection` | `1; mode=block` |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` |
| `Referrer-Policy` | `strict-origin-when-cross-origin` |
| `Permissions-Policy` | `geolocation=(), microphone=(), camera=()` |

Override or disable individual headers:

```go
app.Use(middleware.SecurityHeaders(middleware.SecurityOptions{
    ContentSecurityPolicy: "default-src 'self'; script-src 'self' cdn.example.com",
    XFrameOptions:         "-",  // "-" disables the header entirely
}))
```

See [Security](security.md) for full details.

---

## Rate Limiting

### Global rate limit

```go
app.Use(middleware.RateLimit(100, time.Minute))
// 100 requests per minute per IP address, globally
```

Returns `429 Too Many Requests` with `Retry-After` header when the limit is exceeded.

### Per-route rate limit guard

```go
nexgou.Post("/auth/login", c.Login).
    Guard(&middleware.RateLimitGuard{Max: 5, Window: time.Minute})
```

Sets `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset` headers on every response.

Both global and per-route limits apply independently — the most restrictive wins.

See [Security](security.md) for response headers reference.

---

## Timeout

### Global timeout

```go
app.Use(middleware.Timeout(30 * time.Second))
```

Returns `408 Request Timeout` if the handler does not complete within the deadline.

### Per-route timeout interceptor

```go
nexgou.Post("/slow-operation", c.Process).
    Intercept(&middleware.TimeoutInterceptor{Duration: 120 * time.Second})
```

The per-route interceptor overrides the effective deadline for that specific route.

---

## Body Size Limit

### Global body limit

```go
app.Use(middleware.BodyLimit(1 << 20)) // 1 MB
```

Returns `413 Content Too Large` if the request body exceeds the limit.

### Per-route body limit interceptor

```go
nexgou.Post("/uploads", c.Upload).
    Intercept(&middleware.BodyLimitInterceptor{MaxBytes: 50 << 20}) // 50 MB
```

Useful to allow large uploads on specific routes while keeping a strict global cap.

---

## Custom Middleware

Any function matching `func(nexgou.HandlerFunc) nexgou.HandlerFunc` is valid middleware.

```go
func RequestIDMiddleware(next nexgou.HandlerFunc) nexgou.HandlerFunc {
    return func(ctx *nexgou.Context) error {
        id := uuid.New().String()
        ctx.Writer.Header().Set("X-Request-ID", id)
        ctx.Request.Header.Set("X-Request-ID", id)
        return next(ctx)
    }
}

app.Use(RequestIDMiddleware)
```

### Accessing the underlying http objects

```go
func MyMiddleware(next nexgou.HandlerFunc) nexgou.HandlerFunc {
    return func(ctx *nexgou.Context) error {
        // Access the raw request
        r := ctx.Request
        w := ctx.Writer

        // Set response headers before the handler runs
        w.Header().Set("X-Powered-By", "Nexgou")

        return next(ctx)
    }
}
```

---

## Global vs Per-Route

| Mechanism | Registration | Scope |
|:---|:---|:---|
| `app.Use(mw)` | On the `App` | Every request |
| `.Guard(g)` on route | On a `Route` | That route only, runs before handler |
| `.Intercept(i)` on route | On a `Route` | That route only, wraps the handler |

Global middleware and per-route guards/interceptors are **additive**. If a global rate limit of 100 req/min is set AND a per-route limit of 5 req/min is set on `/login`, both apply independently — hitting 5 on `/login` triggers the route-level guard even if the global counter is still at 4.
