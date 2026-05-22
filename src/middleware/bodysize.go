package middleware

import (
	"errors"
	"net/http"

	"github.com/nexgou/server/src/common"
)

// ── Global body limit middleware ───────────────────────────────────────────────

// BodyLimit returns a global middleware that rejects request bodies larger than
// maxBytes with 413 Payload Too Large.
//
// Common size constants for convenience:
//
//	1 << 10  =   1 KB
//	1 << 20  =   1 MB
//	10 << 20 =  10 MB
//
// Usage:
//
//	app.Use(middleware.BodyLimit(1 << 20)) // 1 MB global limit
func BodyLimit(maxBytes int64) common.MiddlewareFunc {
	return func(next common.HandlerFunc) common.HandlerFunc {
		return func(ctx *common.Context) error {
			return runWithBodyLimit(ctx, maxBytes, next)
		}
	}
}

// ── Per-route body limit interceptor ──────────────────────────────────────────

// BodyLimitInterceptor is an Interceptor that enforces a per-route body size limit.
// Attach it to individual routes using .Intercept(...).
//
// Usage:
//
//	nexgou.Post("/upload", c.Upload).
//	    Intercept(&middleware.BodyLimitInterceptor{MaxBytes: 50 << 20}) // 50 MB
type BodyLimitInterceptor struct {
	// MaxBytes is the maximum number of bytes allowed in the request body.
	MaxBytes int64
}

func (b *BodyLimitInterceptor) Intercept(ctx *common.Context, next common.HandlerFunc) error {
	return runWithBodyLimit(ctx, b.MaxBytes, next)
}

// ── shared implementation ──────────────────────────────────────────────────────

// runWithBodyLimit wraps the request body with http.MaxBytesReader so that
// any read beyond maxBytes returns an error. If the body is oversized, the
// handler will get an error when reading, which is caught here and translated
// into a 413 response before it can reach user code.
func runWithBodyLimit(ctx *common.Context, maxBytes int64, next common.HandlerFunc) error {
	if ctx.Request.Body != nil && ctx.Request.Body != http.NoBody {
		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxBytes)
	}

	err := next(ctx)
	if err != nil && isBodyTooLarge(err) {
		return common.NewHttpException(http.StatusRequestEntityTooLarge, "request body too large")
	}
	return err
}

// isBodyTooLarge reports whether the error was produced by http.MaxBytesReader.
func isBodyTooLarge(err error) bool {
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		return true
	}
	// Fallback for Go versions < 1.19 where MaxBytesError was not exported.
	return err != nil && err.Error() == "http: request body too large"
}
