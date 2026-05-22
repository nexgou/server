package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/nexgou/server/src/common"
)

// CorsOptions configures the CORS policy applied by CorsWithOptions.
type CorsOptions struct {
	// AllowedOrigins is the list of origins allowed to make cross-origin requests.
	// Use []string{"*"} to allow any origin (default).
	// Credentials cannot be used with wildcard origins.
	AllowedOrigins []string

	// AllowedMethods is the list of HTTP methods allowed.
	// Default: GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS.
	AllowedMethods []string

	// AllowedHeaders is the list of request headers the client is allowed to send.
	// Default: Content-Type, Authorization.
	AllowedHeaders []string

	// ExposedHeaders lists headers that are safe to expose to the browser.
	// Default: none.
	ExposedHeaders []string

	// AllowCredentials indicates whether cookies or HTTP authentication may be included.
	// Cannot be combined with AllowedOrigins: ["*"].
	AllowCredentials bool

	// MaxAge is the number of seconds browsers may cache the preflight response.
	// Default: 600 (10 minutes). Set to -1 to omit the header.
	MaxAge int
}

// CorsWithOptions returns a middleware with a fully configurable CORS policy.
// It handles OPTIONS preflight requests automatically (responds 204 and stops the chain).
//
// Usage:
//
//	app.Use(middleware.CorsWithOptions(middleware.CorsOptions{
//	    AllowedOrigins:   []string{"https://example.com", "https://app.example.com"},
//	    AllowCredentials: true,
//	    MaxAge:           3600,
//	}))
func CorsWithOptions(opts CorsOptions) common.MiddlewareFunc {
	// Apply defaults
	if len(opts.AllowedOrigins) == 0 {
		opts.AllowedOrigins = []string{"*"}
	}
	if len(opts.AllowedMethods) == 0 {
		opts.AllowedMethods = []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if len(opts.AllowedHeaders) == 0 {
		opts.AllowedHeaders = []string{"Content-Type", "Authorization"}
	}
	if opts.MaxAge == 0 {
		opts.MaxAge = 600
	}

	isWildcard := len(opts.AllowedOrigins) == 1 && opts.AllowedOrigins[0] == "*"
	methods := strings.Join(opts.AllowedMethods, ", ")
	headers := strings.Join(opts.AllowedHeaders, ", ")
	exposed := strings.Join(opts.ExposedHeaders, ", ")
	maxAge := strconv.Itoa(opts.MaxAge)

	return func(next common.HandlerFunc) common.HandlerFunc {
		return func(ctx *common.Context) error {
			h := ctx.Writer.Header()
			origin := ctx.Request.Header.Get("Origin")

			// Determine the effective allowed origin for this request
			if isWildcard {
				h.Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" && originAllowed(origin, opts.AllowedOrigins) {
				h.Set("Access-Control-Allow-Origin", origin)
				h.Add("Vary", "Origin")
			}

			if opts.AllowCredentials {
				h.Set("Access-Control-Allow-Credentials", "true")
			}
			if exposed != "" {
				h.Set("Access-Control-Expose-Headers", exposed)
			}

			// Preflight
			if ctx.Request.Method == http.MethodOptions {
				h.Set("Access-Control-Allow-Methods", methods)
				h.Set("Access-Control-Allow-Headers", headers)
				if opts.MaxAge > 0 {
					h.Set("Access-Control-Max-Age", maxAge)
				}
				ctx.Writer.WriteHeader(http.StatusNoContent)
				return nil
			}

			return next(ctx)
		}
	}
}

// originAllowed reports whether the given origin is in the allowed list.
func originAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == origin {
			return true
		}
	}
	return false
}
