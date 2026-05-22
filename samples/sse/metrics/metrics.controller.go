package metrics

import (
	"math/rand"
	"time"

	nexgou "github.com/nexgou/server"
	"github.com/nexgou/server/src/logger"
)

// MetricsController streams simulated system metrics as Server-Sent Events.
//
// Routes:
//
//	GET /metrics/live             — all metrics, one event per second
//	GET /metrics/live/:topic      — single metric (cpu | mem | goroutines)
//	GET /metrics/snapshot         — one-shot JSON response (not SSE)
type MetricsController struct {
	log *logger.ScopedLogger
}

// NewMetricsController creates a MetricsController (used by the IoC container).
func NewMetricsController(log *logger.LoggerService) *MetricsController {
	return &MetricsController{
		log: log.WithContext("MetricsController"),
	}
}

// Register declares all routes for this controller.
func (c *MetricsController) Register() []nexgou.Route {
	return []nexgou.Route{
		// SSE: continuous stream of all metrics.
		nexgou.SSE("/metrics/live", c.StreamAll),

		// SSE: stream a single metric topic.
		nexgou.SSE("/metrics/live/:topic", c.StreamTopic),

		// HTTP: one-shot snapshot (regular JSON response, not SSE).
		nexgou.Get("/metrics/snapshot", c.Snapshot),
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// StreamAll emits a "metrics" event every second with cpu, mem and goroutines.
func (c *MetricsController) StreamAll(ctx *nexgou.SSEContext) error {
	c.log.Info("stream started", "remote", ctx.RemoteAddr(), "topic", "all")
	defer c.log.Info("stream ended", "remote", ctx.RemoteAddr(), "topic", "all")

	// Tell the browser to reconnect after 5 s on disconnect.
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
			if err := ctx.SendNamedJSON("metrics", sampleMetrics()); err != nil {
				return err
			}
		}
	}
}

// StreamTopic emits events every second for a single topic.
func (c *MetricsController) StreamTopic(ctx *nexgou.SSEContext) error {
	topic := ctx.Param("topic")

	validTopics := map[string]bool{"cpu": true, "mem": true, "goroutines": true}
	if !validTopics[topic] {
		return nexgou.BadRequestException(
			"unknown topic '" + topic + "'; valid values: cpu, mem, goroutines",
		)
	}

	c.log.Info("stream started", "remote", ctx.RemoteAddr(), "topic", topic)
	defer c.log.Info("stream ended", "remote", ctx.RemoteAddr(), "topic", topic)

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
			payload := nexgou.H{
				"topic": topic,
				"value": m[topic],
				"ts":    m["ts"],
			}
			if err := ctx.SendNamedJSON(topic, payload); err != nil {
				return err
			}
		}
	}
}

// Snapshot returns a one-shot JSON response with the current metrics.
// This is a regular HTTP handler — no SSE involved.
func (c *MetricsController) Snapshot(ctx *nexgou.Context) error {
	return ctx.JSON(200, sampleMetrics())
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func sampleMetrics() nexgou.H {
	return nexgou.H{
		"cpu":        round(20 + rand.Float64()*60),
		"mem":        round(30 + rand.Float64()*50),
		"goroutines": 10 + rand.Intn(90),
		"ts":         time.Now().UTC().Format(time.RFC3339),
	}
}

func round(v float64) float64 {
	return float64(int(v*100)) / 100
}
