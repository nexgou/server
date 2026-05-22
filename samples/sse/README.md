# Sample: Server-Sent Events (SSE)

A complete example of a Nexgou SSE application demonstrating real-time one-way streaming from server to client. The sample streams simulated system metrics (CPU, memory, goroutines) as Server-Sent Events.

## What this sample covers

| Feature | Where |
|---------|-------|
| `nexgou.SSE()` route helper | `metrics/metrics.controller.go` |
| `SSEContext` — named events, typed JSON | `metrics/metrics.controller.go` |
| Client disconnect detection (`ctx.Done()`) | `metrics/metrics.controller.go` |
| URL parameter in SSE route (`:topic`) | `metrics/metrics.controller.go` — `StreamTopic` |
| `SetRetry` — browser reconnect delay | `metrics/metrics.controller.go` |
| Mixed SSE + HTTP routes in one controller | `metrics/metrics.controller.go` — `Snapshot` |
| `LoggerService` injection | `metrics/metrics.controller.go` |
| CORS middleware | `main.go` |

## Project structure

```
samples/sse/
├── main.go               # Bootstrap — middleware pipeline + server start (port 3002)
├── app.module.go         # Root module — imports LogModule, MetricsModule
├── README.md
└── metrics/
    ├── metrics.controller.go  # Routes: SSE streams + JSON snapshot
    └── metrics.module.go      # MetricsModule definition
```

## Routes

| Method | Path | Type | Description |
|--------|------|------|-------------|
| `GET` | `/metrics/live` | SSE | All metrics every second |
| `GET` | `/metrics/live/:topic` | SSE | Single metric: `cpu`, `mem`, or `goroutines` |
| `GET` | `/metrics/snapshot` | HTTP JSON | One-shot current metrics snapshot |

## Run

```bash
cd samples/sse
go run .
# Server starts on http://localhost:3002
```

## Try it with curl

```bash
# Full stream — all metrics, one event per second
curl -N http://localhost:3002/metrics/live

# Single topic stream
curl -N http://localhost:3002/metrics/live/cpu
curl -N http://localhost:3002/metrics/live/mem
curl -N http://localhost:3002/metrics/live/goroutines

# One-shot snapshot (regular HTTP, not SSE)
curl http://localhost:3002/metrics/snapshot
```

The `-N` flag disables curl's output buffering so events appear as they arrive.

### Expected SSE output

```
event: metrics
data: {"cpu":47.23,"goroutines":34,"mem":61.05,"ts":"2026-05-22T10:00:01Z"}

event: metrics
data: {"cpu":52.11,"goroutines":41,"mem":58.99,"ts":"2026-05-22T10:00:02Z"}
```

```
event: cpu
data: {"topic":"cpu","ts":"2026-05-22T10:00:01Z","value":47.23}

event: cpu
data: {"topic":"cpu","ts":"2026-05-22T10:00:02Z","value":52.11}
```

## Try it with Postman

1. **New → HTTP**
2. Method: `GET`, URL: `http://localhost:3002/metrics/live`
3. Click **Send**
4. Events appear in the response body as they arrive (requires Postman v10+)

## Try it from a browser

Open the browser console and run:

```js
// Full metrics stream
const es = new EventSource('http://localhost:3002/metrics/live')

es.addEventListener('metrics', (e) => {
    const m = JSON.parse(e.data)
    console.log(`CPU: ${m.cpu}% | MEM: ${m.mem}% | Goroutines: ${m.goroutines}`)
})

es.onerror = () => console.warn('Disconnected — browser will reconnect automatically')

// Stop
// es.close()
```

```js
// Single topic stream
const cpu = new EventSource('http://localhost:3002/metrics/live/cpu')

cpu.addEventListener('cpu', (e) => {
    const { value, ts } = JSON.parse(e.data)
    console.log(`[${ts}] CPU: ${value}%`)
})
```

## How the disconnect detection works

Every SSE handler listens on `ctx.Done()`, which is backed by `http.Request.Context().Done()`. Go's HTTP server closes this channel automatically when the TCP connection drops:

```go
func (c *MetricsController) StreamAll(ctx *nexgou.SSEContext) error {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return nil   // ← clean exit, no goroutine leak
        case <-ticker.C:
            ctx.SendNamedJSON("metrics", sampleMetrics())
        }
    }
}
```

## Mixing SSE and HTTP routes

A controller can expose both SSE streaming routes and regular HTTP routes at the same time. `GET /metrics/snapshot` is a normal JSON handler on the same controller:

```go
func (c *MetricsController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.SSE("/metrics/live",         c.StreamAll),    // SSE
        nexgou.SSE("/metrics/live/:topic",  c.StreamTopic),  // SSE
        nexgou.Get("/metrics/snapshot",     c.Snapshot),     // HTTP JSON
    }
}
```

## Adding authentication

To protect the stream with a guard:

```go
nexgou.SSE("/metrics/live", c.StreamAll).Guard(&AuthGuard{})
```

Guards run as normal HTTP middleware before the streaming starts. See [`docs/sse.md`](../../docs/sse.md) for the full reference.

## SSE vs WebSocket

| | SSE (this sample) | WebSocket (`samples/chat`) |
|---|---|---|
| Direction | Server → client | Bidirectional |
| Reconnect | Automatic (`EventSource`) | Manual |
| Best for | Metrics, notifications, feeds | Chat, collaboration, gaming |

See [`docs/sse.md`](../../docs/sse.md) for a detailed comparison.
