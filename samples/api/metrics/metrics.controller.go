package metrics

import (
	"math/rand"
	"time"

	nexgou "github.com/nexgou/server"
	nexgousse "github.com/nexgou/server/src/sse"
)

// MetricsController streams simulated system metrics as Server-Sent Events.
// It exposes two routes:
//
//	GET /metrics/live          — continuous stream, one event per second
//	GET /metrics/live/:topic   — stream filtered to a named topic
type MetricsController struct{}

// NewMetricsController creates a MetricsController (used by the IoC container).
func NewMetricsController() *MetricsController {
	return &MetricsController{}
}

// Register declares the HTTP routes for this controller.
func (c *MetricsController) Register() []nexgou.Route {
	return []nexgou.Route{
		// Full metrics stream — all topics.
		nexgou.SSE("/metrics/live", c.StreamAll),

		// Filtered stream — only the requested topic (cpu | mem | goroutines).
		nexgou.SSE("/metrics/live/:topic", c.StreamTopic),
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// StreamAll emits a "metrics" event every second with cpu, mem and goroutines.
// The stream runs until the client disconnects.
func (c *MetricsController) StreamAll(ctx *nexgou.SSEContext) error {
	// Instruct the browser to reconnect after 5 s on disconnect.
	if err := ctx.SetRetry(5000); err != nil {
		return err
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Client disconnected — stop streaming cleanly.
			return nil

		case <-ticker.C:
			payload := sampleMetrics()
			if err := ctx.SendNamedJSON("metrics", payload); err != nil {
				return err
			}
		}
	}
}

// StreamTopic emits events every second for a single topic: cpu, mem or goroutines.
// Unknown topics send an error event and close the stream.
func (c *MetricsController) StreamTopic(ctx *nexgou.SSEContext) error {
	topic := ctx.Param("topic")

	if topic != "cpu" && topic != "mem" && topic != "goroutines" {
		return nexgou.BadRequestException("unknown topic: " + topic + "; valid values: cpu, mem, goroutines")
	}

	if err := ctx.SetRetry(5000); err != nil {
		return err
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			m := sampleMetrics()
			var value any
			switch topic {
			case "cpu":
				value = m["cpu"]
			case "mem":
				value = m["mem"]
			case "goroutines":
				value = m["goroutines"]
			}
			payload := nexgou.H{"topic": topic, "value": value, "ts": m["ts"]}
			if err := ctx.SendNamedJSON(topic, payload); err != nil {
				return err
			}
		}
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// sampleMetrics returns simulated system metrics.
// In a real application you would read from runtime.MemStats, gopsutil, etc.
func sampleMetrics() nexgou.H {
	return nexgou.H{
		"cpu":        round(20 + rand.Float64()*60),   // 20–80 %
		"mem":        round(30 + rand.Float64()*50),   // 30–80 %
		"goroutines": 10 + rand.Intn(90),              // 10–99
		"ts":         time.Now().UTC().Format(time.RFC3339),
	}
}

// round rounds a float64 to two decimal places.
func round(v float64) float64 {
	return float64(int(v*100)) / 100
}

// Ensure MetricsController is wired via the SSE helper (compile-time check).
var _ nexgousse.HandlerFunc = (*MetricsController)(nil).StreamAll
