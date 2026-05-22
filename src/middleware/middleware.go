package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/nexgou/server/src/common"
)

// Logger returns a middleware that logs each request's method, path, and duration.
// Duration is displayed as µs, ms, or s depending on magnitude.
func Logger() common.MiddlewareFunc {
	return func(next common.HandlerFunc) common.HandlerFunc {
		return func(ctx *common.Context) error {
			start := time.Now()
			err := next(ctx)
			log.Printf("[Nexgou] %s %s — %s", ctx.Method(), ctx.Path(), formatDuration(time.Since(start)))
			return err
		}
	}
}

// formatDuration returns a human-readable duration with the most appropriate unit.
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Microsecond:
		return fmt.Sprintf("%dns", d.Nanoseconds())
	case d < time.Millisecond:
		return fmt.Sprintf("%.2fµs", float64(d.Nanoseconds())/1e3)
	case d < time.Second:
		return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1e6)
	default:
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}

// Recovery returns a middleware that recovers from panics and returns a 500 error.
func Recovery() common.MiddlewareFunc {
	return func(next common.HandlerFunc) common.HandlerFunc {
		return func(ctx *common.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Nexgou] panic recovered: %v", r)
					err = common.NewInternalServerErrorException("Internal Server Error")
				}
			}()
			return next(ctx)
		}
	}
}

// Cors returns a middleware that sets permissive CORS headers.
// For production, replace with a configured policy.
func Cors() common.MiddlewareFunc {
	return func(next common.HandlerFunc) common.HandlerFunc {
		return func(ctx *common.Context) error {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			return next(ctx)
		}
	}
}
