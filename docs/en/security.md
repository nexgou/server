# Security Middleware

Nexgou ships five security-focused middleware and per-route primitives out of the box, all implemented using the Go standard library — no external dependencies required.

> All middleware and guards live in the `github.com/nexgou/server/src/middleware` package.

---

## Table of Contents

- [SecurityHeaders](#securityheaders)
- [CorsWithOptions](#corswithoptions)
- [RateLimit / RateLimitGuard](#ratelimit--ratelimitguard)
- [Timeout / TimeoutInterceptor](#timeout--timeoutinterceptor)
- [BodyLimit / BodyLimitInterceptor](#bodylimit--bodylimitinterceptor)
- [Recommended pipeline order](#recommended-pipeline-order)
- [Combining global and per-route limits](#combining-global-and-per-route-limits)

---

## SecurityHeaders

Sets secure HTTP response headers on every request.

### Usage

```go
// Defaults — safe for most applications
app.Use(middleware.SecurityHeaders())

// Custom — override individual headers
app.Use(middleware.SecurityHeaders(middleware.SecurityOptions{
    ContentSecurityPolicy:   "default-src 'self'; img-src *; script-src 'self'",
    XFrameOptions:           "SAMEORIGIN",
    StrictTransportSecurity: "-", // "-" disables the header (e.g. on local dev)
}))
```

### SecurityOptions fields

| Field                     | Default                                    | Description                                    |
| ------------------------- | ------------------------------------------ | ---------------------------------------------- |
| `ContentSecurityPolicy`   | `default-src 'self'`                       | Controls which resources the browser may load  |
| `XFrameOptions`           | `DENY`                                     | Prevents clickjacking via `<iframe>` embedding |
| `XContentTypeOptions`     | `nosniff`                                  | Prevents MIME-type sniffing                    |
| `XXSSProtection`          | `1; mode=block`                            | Enables browser XSS filter (legacy browsers)   |
| `StrictTransportSecurity` | `max-age=31536000; includeSubDomains`      | Forces HTTPS for 1 year                        |
| `ReferrerPolicy`          | `strict-origin-when-cross-origin`          | Controls the `Referer` header                  |
| `PermissionsPolicy`       | `geolocation=(), microphone=(), camera=()` | Restricts browser feature access               |

Set any field to `"-"` to omit that header entirely.

---

## CorsWithOptions

Configurable CORS policy with automatic preflight (`OPTIONS`) handling.

The original `Cors()` helper remains available for simple `Access-Control-Allow-Origin: *` use cases.

### Usage

```go
// Open API — allow any origin
app.Use(middleware.CorsWithOptions(middleware.CorsOptions{
    AllowedOrigins: []string{"*"},
}))

// Restricted API — specific origins, with credentials
app.Use(middleware.CorsWithOptions(middleware.CorsOptions{
    AllowedOrigins:   []string{"https://app.example.com", "https://admin.example.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Content-Type", "Authorization"},
    ExposedHeaders:   []string{"X-Request-Id"},
    AllowCredentials: true,
    MaxAge:           3600,
}))
```

### CorsOptions fields

| Field              | Default                                        | Description                                               |
| ------------------ | ---------------------------------------------- | --------------------------------------------------------- |
| `AllowedOrigins`   | `["*"]`                                        | Origins allowed to make cross-origin requests             |
| `AllowedMethods`   | `GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS` | Allowed HTTP methods                                      |
| `AllowedHeaders`   | `Content-Type, Authorization`                  | Allowed request headers                                   |
| `ExposedHeaders`   | `[]`                                           | Headers the browser may access in the response            |
| `AllowCredentials` | `false`                                        | Allow cookies / HTTP auth. Incompatible with `"*"` origin |
| `MaxAge`           | `600`                                          | Seconds to cache preflight response. Set to `-1` to omit  |

Preflight (`OPTIONS`) requests are handled automatically: the middleware responds `204 No Content` and stops the chain.

Preflight hardening behavior:

- If `Access-Control-Request-Method` is present and not in `AllowedMethods`, the middleware returns `204` without CORS preflight allow headers.
- If `Access-Control-Request-Headers` contains headers outside `AllowedHeaders` (case-insensitive), the middleware returns `204` without CORS preflight allow headers.
- If `AllowCredentials=true` and `AllowedOrigins=["*"]`, the middleware reflects the incoming `Origin` (instead of `*`) to keep browser behavior compliant.

---

## RateLimit / RateLimitGuard

Fixed-window rate limiter per client IP. Excess requests receive `429 Too Many Requests`.

Response headers set on every request:

| Header                  | Description                                   |
| ----------------------- | --------------------------------------------- |
| `X-RateLimit-Limit`     | Maximum requests allowed in the window        |
| `X-RateLimit-Remaining` | Requests remaining in the current window      |
| `Retry-After`           | Seconds until the window resets (only on 429) |

Client IP is resolved in order: `X-Forwarded-For` → `X-Real-IP` → `RemoteAddr`.

### Global — middleware

Applies to all routes.

```go
// 100 requests per IP per minute
app.Use(middleware.RateLimit(100, time.Minute))
```

### Per-route — RateLimitGuard

Implements the `Guard` interface. Attach with `.Guard(...)`.  
Per-route limits are **independent** from the global limit — the client must satisfy both.

```go
nexgou.Post("/login", c.Login).
    Guard(&middleware.RateLimitGuard{Max: 5, Window: time.Minute})
```

### RateLimitGuard fields

| Field    | Type            | Description                              |
| -------- | --------------- | ---------------------------------------- |
| `Max`    | `int`           | Maximum requests allowed within `Window` |
| `Window` | `time.Duration` | The time window for the counter          |

---

## Timeout / TimeoutInterceptor

Cancels the request context after the configured duration.  
Times out with `408 Request Timeout`.

### Global — middleware

```go
app.Use(middleware.Timeout(30 * time.Second))
```

### Per-route — TimeoutInterceptor

Implements the `Interceptor` interface. Attach with `.Intercept(...)`.

```go
nexgou.Get("/report", c.HeavyReport).
    Intercept(&middleware.TimeoutInterceptor{Duration: 60 * time.Second})
```

### TimeoutInterceptor fields

| Field      | Type            | Description                                  |
| ---------- | --------------- | -------------------------------------------- |
| `Duration` | `time.Duration` | Maximum time the handler may take to respond |

> The per-route timeout overrides the global timeout for that specific route — the most restrictive deadline wins (whichever fires first).

---

## BodyLimit / BodyLimitInterceptor

Caps the maximum size of the request body using `http.MaxBytesReader`.  
Oversized requests receive `413 Payload Too Large`.

### Global — middleware

```go
app.Use(middleware.BodyLimit(1 << 20)) // 1 MB
```

### Per-route — BodyLimitInterceptor

Implements the `Interceptor` interface. Attach with `.Intercept(...)`.

```go
nexgou.Post("/upload", c.Upload).
    Intercept(&middleware.BodyLimitInterceptor{MaxBytes: 50 << 20}) // 50 MB
```

### BodyLimitInterceptor fields

| Field      | Type    | Description                        |
| ---------- | ------- | ---------------------------------- |
| `MaxBytes` | `int64` | Maximum allowed body size in bytes |

### Common size constants

```go
1 << 10  =    1 KB
1 << 20  =    1 MB
10 << 20 =   10 MB
50 << 20 =   50 MB
```

---

## Recommended pipeline order

```go
app.Use(middleware.Recovery())        // 1. catch panics from everything below
app.Use(middleware.SecurityHeaders()) // 2. set headers before any write
app.Use(middleware.CorsWithOptions(middleware.CorsOptions{
    AllowedOrigins: []string{"*"},
}))                                   // 3. handle OPTIONS preflight early
app.Use(middleware.RateLimit(100, time.Minute)) // 4. reject abusive IPs early
app.Use(middleware.Timeout(30 * time.Second))   // 5. bound all downstream work
app.Use(middleware.BodyLimit(1 << 20))          // 6. cap payload before handlers read
app.Use(middleware.Logger())          // 7. log final status (incl. 429, 408, 413)
```

---

## Combining global and per-route limits

Global and per-route limits are **independent and additive**. A request must pass both.

```go
// Global: 100 req/min, 30s timeout, 1 MB body
app.Use(middleware.RateLimit(100, time.Minute))
app.Use(middleware.Timeout(30 * time.Second))
app.Use(middleware.BodyLimit(1 << 20))

// Route: additionally 10 req/min, 10s timeout, 64 KB body
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

In the example above, a request to `POST /users` is subject to:

- Rate limit: whichever counter is exhausted first (global or route-level)
- Timeout: `min(30s, 10s)` = 10 seconds (the per-route deadline fires first)
- Body: `min(1 MB, 64 KB)` = 64 KB (the per-route limit fires first)
