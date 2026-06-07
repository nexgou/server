package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/nexgou/server/src/common"
)

// ── Global timeout middleware ──────────────────────────────────────────────────

// Timeout returns a global middleware that cancels the request context after d.
// If the handler exceeds the deadline, the client receives 408 Request Timeout.
//
// Usage:
//
//	app.Use(middleware.Timeout(30 * time.Second))
// Timeout returns a global middleware that cancels the request context after d.
// WebSocket upgrade requests are exempt — they are long-lived connections and
// must not be subject to a short HTTP request deadline.
func Timeout(d time.Duration) common.MiddlewareFunc {
	return func(next common.HandlerFunc) common.HandlerFunc {
		return func(ctx *common.Context) error {
			if strings.EqualFold(ctx.Request.Header.Get("Upgrade"), "websocket") {
				return next(ctx)
			}
			return runWithTimeout(ctx, d, next)
		}
	}
}

// ── Per-route timeout interceptor ─────────────────────────────────────────────

// TimeoutInterceptor is an Interceptor that enforces a per-route request timeout.
// Attach it to individual routes using .Intercept(...).
//
// Usage:
//
//	nexgou.Get("/report", c.HeavyReport).
//	    Intercept(&middleware.TimeoutInterceptor{Duration: 60 * time.Second})
type TimeoutInterceptor struct {
	// Duration is the maximum time allowed for the handler to complete.
	Duration time.Duration
}

func (t *TimeoutInterceptor) Intercept(ctx *common.Context, next common.HandlerFunc) error {
	return runWithTimeout(ctx, t.Duration, next)
}

// ── shared implementation ──────────────────────────────────────────────────────

// runWithTimeout executes next within a bounded context.
// It replaces ctx.Request with a request carrying the derived context,
// then races the handler against a timer. On timeout it writes 408 and
// returns early — if the handler already wrote headers the response is
// left as-is and only the error is returned.
func runWithTimeout(ctx *common.Context, d time.Duration, next common.HandlerFunc) error {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), d)
	defer cancel()

	// Swap the request for one that carries the timeout context.
	ctx.Request = ctx.Request.WithContext(reqCtx)

	done := make(chan error, 1)
	go func() {
		done <- next(ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-reqCtx.Done():
		if reqCtx.Err() == context.DeadlineExceeded {
			return common.NewHttpException(http.StatusRequestTimeout, "request timeout")
		}
		return common.NewHttpException(http.StatusServiceUnavailable, "request canceled")
	}
}
