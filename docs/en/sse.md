# Server-Sent Events (SSE)

Nexgou has first-class SSE support via the `nexgousse` sub-package. SSE handlers integrate seamlessly with the module system, the IoC container, Guards, and the full middleware pipeline — no separate server or protocol upgrade required.

---

## Table of Contents

- [How it works](#how-it-works)
- [SSEContext API](#ssecontext-api)
- [Registering an SSE route](#registering-an-sse-route)
  - [Using the nexgou.SSE helper (recommended)](#using-the-nexgousse-helper-recommended)
  - [Using ToHTTPHandler directly](#using-tohttphandler-directly)
- [Detecting client disconnect](#detecting-client-disconnect)
- [Named events](#named-events)
- [Event IDs and reconnection](#event-ids-and-reconnection)
- [URL parameters](#url-parameters)
- [Guards on SSE routes](#guards-on-sse-routes)
- [Versioning](#versioning)
- [Full example — metrics streaming](#full-example--metrics-streaming)
- [Testing with curl](#testing-with-curl)
- [Testing with Postman](#testing-with-postman)
- [Testing from a browser](#testing-from-a-browser)
- [SSE vs WebSocket — when to use each](#sse-vs-websocket--when-to-use-each)

---

## How it works

An SSE route is a standard HTTP `GET` handler that keeps the connection open and writes events in the `text/event-stream` format. The framework:

1. Matches the request against the normal HTTP route table.
2. Runs the global middleware pipeline and any route Guards.
3. Calls `ToHTTPHandler` which writes the required SSE response headers and creates an `SSEContext`.
4. Invokes the SSE handler function for the lifetime of the connection.
5. Returns when the handler returns (client disconnect or explicit return).

No protocol upgrade takes place — SSE works over plain HTTP/1.1 and passes through every proxy, load balancer, and CDN without special configuration.

### Response headers written automatically

| Header | Value |
|--------|-------|
| `Content-Type` | `text/event-stream` |
| `Cache-Control` | `no-cache` |
| `Connection` | `keep-alive` |
| `X-Accel-Buffering` | `no` *(disables nginx/proxy buffering)* |

---

## SSEContext API

`SSEContext` wraps the `http.ResponseWriter` and exposes a clean API for writing SSE events.

### Write methods

| Method | SSE output | Description |
|--------|-----------|-------------|
| `Send(data string) error` | `data: <data>\n\n` | Send a plain-text unnamed event |
| `SendNamed(event, data string) error` | `event: <name>\ndata: <data>\n\n` | Send a named event |
| `SendJSON(v any) error` | `data: <json>\n\n` | Marshal `v` as JSON and send |
| `SendNamedJSON(event string, v any) error` | `event: <name>\ndata: <json>\n\n` | Marshal `v` as JSON and send as named event |
| `SendComment(comment string) error` | `: <comment>\n\n` | Send a keepalive comment (invisible to `EventSource`) |
| `SetRetry(ms int) error` | `retry: <ms>\n\n` | Set browser reconnect delay in milliseconds |
| `SetID(id string) error` | `id: <id>\n\n` | Set the last event ID |

> Multi-line data values are split automatically: each `\n` in the data string produces a separate `data:` line, as required by the SSE specification.

### Request helpers

| Method | Description |
|--------|-------------|
| `Param(key string) string` | Read a URL path parameter (e.g. `:topic` → `"topic"`) |
| `Header(key string) string` | Read a header from the original request |
| `LastEventID() string` | Value of `Last-Event-ID` header (sent on browser reconnect) |
| `RemoteAddr() string` | Client's network address |
| `Done() <-chan struct{}` | Closed when the client disconnects |
| `Request *http.Request` | The original HTTP request (read-only) |

---

## Registering an SSE route

### Using the nexgou.SSE helper (recommended)

```go
func (c *NotificationsController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.SSE("/notifications", c.Stream),
    }
}
```

`nexgou.SSE` returns a standard `nexgou.Route` so you can chain `.Guard()`, `.Version()`, and `.Intercept()` exactly like any HTTP route:

```go
nexgou.SSE("/notifications", c.Stream).
    Guard(&AuthGuard{}).
    Version("v1")
// effective path: /v1/notifications
```

### Using ToHTTPHandler directly

If you prefer explicit control or need to mix SSE and HTTP in the same `Register()` list without the helper:

```go
import nexgousse "github.com/nexgou/server/src/sse"

func (c *NotificationsController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.Get("/notifications", nexgousse.ToHTTPHandler(c.Stream)),
    }
}
```

---

## Detecting client disconnect

Use `ctx.Done()` to stop the streaming loop cleanly when the client closes the connection. This prevents goroutine leaks.

```go
func (c *NotificationsController) Stream(ctx *nexgou.SSEContext) error {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            // Client disconnected — return nil for a clean exit.
            return nil
        case <-ticker.C:
            if err := ctx.SendJSON(nexgou.H{"ts": time.Now().Unix()}); err != nil {
                return err
            }
        }
    }
}
```

`ctx.Done()` is backed by `http.Request.Context().Done()`, which is closed by Go's HTTP server as soon as the client TCP connection is detected as closed.

---

## Named events

Named events let the browser register specific listeners with `EventSource.addEventListener`:

```go
// Server
ctx.SendNamedJSON("user.created", nexgou.H{"id": 42, "name": "Alice"})
ctx.SendNamedJSON("order.shipped", nexgou.H{"order": "ORD-001"})
```

```js
// Browser
const es = new EventSource('/notifications')

es.addEventListener('user.created', (e) => {
    const user = JSON.parse(e.data)
    console.log('New user:', user.name)
})

es.addEventListener('order.shipped', (e) => {
    const order = JSON.parse(e.data)
    console.log('Shipped:', order.order)
})
```

Unnamed events (`Send`, `SendJSON`) are received by `es.onmessage`.

---

## Event IDs and reconnection

Set an ID on each event so the browser can resume from where it left off after a reconnect:

```go
func (c *FeedController) Stream(ctx *nexgou.SSEContext) error {
    ctx.SetRetry(3000)

    // Resume from the last seen event if the client is reconnecting.
    lastID := ctx.LastEventID()
    events := c.service.EventsSince(lastID)

    for _, ev := range events {
        ctx.SetID(ev.ID)
        ctx.SendNamedJSON("feed", ev)
    }

    // Continue with live events...
}
```

On reconnect, the browser sends `Last-Event-ID: <last-id>` automatically.

---

## URL parameters

Define parameters with a colon prefix in the path, then read them with `ctx.Param`:

```go
nexgou.SSE("/metrics/live/:topic", c.StreamTopic)

func (c *MetricsController) StreamTopic(ctx *nexgou.SSEContext) error {
    topic := ctx.Param("topic") // "cpu", "mem", etc.
    // ...
}
```

---

## Guards on SSE routes

Guards run as normal HTTP middleware before the connection is handed to the SSE handler. A failed guard responds with HTTP `403` and the stream never starts.

```go
nexgou.SSE("/notifications", c.Stream).Guard(&AuthGuard{})
```

```go
type AuthGuard struct{}

func (g *AuthGuard) CanActivate(ctx *nexgou.Context) (bool, error) {
    return ctx.Header("Authorization") != "", nil
}
```

Multiple guards are evaluated in order; the first failure short-circuits:

```go
nexgou.SSE("/admin/events", c.Stream).Guard(&AuthGuard{}, &AdminRoleGuard{})
```

---

## Versioning

```go
nexgou.SSE("/notifications", c.Stream).Version("v1")
// effective path: /v1/notifications
```

---

## Full example — metrics streaming

### `metrics/metrics.controller.go`

```go
package metrics

import (
    "math/rand"
    "time"

    nexgou "github.com/nexgou/server"
)

type MetricsController struct{}

func NewMetricsController() *MetricsController { return &MetricsController{} }

func (c *MetricsController) Register() []nexgou.Route {
    return []nexgou.Route{
        // Full stream: all metrics every second.
        nexgou.SSE("/metrics/live", c.StreamAll),
        // Filtered stream: single topic (cpu | mem | goroutines).
        nexgou.SSE("/metrics/live/:topic", c.StreamTopic),
    }
}

func (c *MetricsController) StreamAll(ctx *nexgou.SSEContext) error {
    ctx.SetRetry(5000)
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return nil
        case <-ticker.C:
            ctx.SendNamedJSON("metrics", nexgou.H{
                "cpu":        20 + rand.Float64()*60,
                "mem":        30 + rand.Float64()*50,
                "goroutines": 10 + rand.Intn(90),
                "ts":         time.Now().UTC().Format(time.RFC3339),
            })
        }
    }
}

func (c *MetricsController) StreamTopic(ctx *nexgou.SSEContext) error {
    topic := ctx.Param("topic")
    if topic != "cpu" && topic != "mem" && topic != "goroutines" {
        return nexgou.BadRequestException("unknown topic: " + topic)
    }
    ctx.SetRetry(5000)
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return nil
        case <-ticker.C:
            ctx.SendNamedJSON(topic, nexgou.H{
                "value": 20 + rand.Float64()*60,
                "ts":    time.Now().UTC().Format(time.RFC3339),
            })
        }
    }
}
```

### `metrics/metrics.module.go`

```go
package metrics

import nexgou "github.com/nexgou/server"

var Module = nexgou.Module(nexgou.ModuleOptions{
    Controllers: []any{NewMetricsController},
})
```

### Register in `app.module.go`

```go
var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{
        nexgou.ConfigModule,
        nexgou.LogModule,
        metrics.Module,
        // ...
    },
})
```

---

## Testing with curl

```bash
# Full metrics stream
curl -N http://localhost:3000/metrics/live

# Single topic
curl -N http://localhost:3000/metrics/live/cpu
```

The `-N` flag (`--no-buffer`) disables curl's output buffering so events appear immediately. Expected output:

```
event: metrics
data: {"cpu":47.23,"goroutines":34,"mem":61.05,"ts":"2026-05-22T10:00:01Z"}

event: metrics
data: {"cpu":52.11,"goroutines":41,"mem":58.99,"ts":"2026-05-22T10:00:02Z"}
```

---

## Testing with Postman

1. Open Postman → **New → HTTP**
2. Method: `GET`, URL: `http://localhost:3000/metrics/live`
3. If the route has a Guard, add the required header in the **Headers** tab
4. Click **Send**
5. Postman displays each event as it arrives in the response body panel

> Postman renders SSE responses in real time as of v10+. If you see the full response only after closing the stream, upgrade Postman.

---

## Testing from a browser

Open the browser console and run:

```js
const es = new EventSource('http://localhost:3000/metrics/live')

// Named events
es.addEventListener('metrics', (e) => {
    console.log(JSON.parse(e.data))
})

// Error / disconnect
es.onerror = () => console.warn('SSE disconnected — browser will reconnect automatically')
```

To stop the stream:

```js
es.close()
```

---

## SSE vs WebSocket — when to use each

| | SSE | WebSocket |
|---|---|---|
| Direction | Server → client only | Bidirectional |
| Protocol | Plain HTTP | HTTP upgrade to WS |
| Automatic reconnect | Yes — built into `EventSource` | No — must implement manually |
| Proxy / CDN support | Works out of the box | Requires WS-aware proxy |
| Complexity | Simple — standard HTTP handler | Higher — persistent connection lifecycle |
| **Best for** | Notifications, live feeds, progress, dashboards | Chat, collaborative editing, gaming |

**Rule of thumb:** use SSE when the server talks and the client listens. Use WebSocket when both sides need to send messages at any time.
