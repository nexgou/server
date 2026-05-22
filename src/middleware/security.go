package middleware

import "github.com/nexgou/server/src/common"

// SecurityOptions allows overriding individual security header values.
// Leave a field empty to use the secure default. Set it to "-" to omit the header entirely.
type SecurityOptions struct {
	// ContentSecurityPolicy sets the Content-Security-Policy header.
	// Default: "default-src 'self'"
	ContentSecurityPolicy string

	// XFrameOptions sets the X-Frame-Options header.
	// Default: "DENY"
	XFrameOptions string

	// XContentTypeOptions sets the X-Content-Type-Options header.
	// Default: "nosniff"
	XContentTypeOptions string

	// XXSSProtection sets the X-XSS-Protection header.
	// Default: "1; mode=block"
	XXSSProtection string

	// StrictTransportSecurity sets the Strict-Transport-Security header.
	// Default: "max-age=31536000; includeSubDomains"
	StrictTransportSecurity string

	// ReferrerPolicy sets the Referrer-Policy header.
	// Default: "strict-origin-when-cross-origin"
	ReferrerPolicy string

	// PermissionsPolicy sets the Permissions-Policy header.
	// Default: "geolocation=(), microphone=(), camera=()"
	PermissionsPolicy string
}

// SecurityHeaders returns a middleware that sets secure HTTP response headers on every request.
// It accepts an optional SecurityOptions to override or disable individual headers.
//
// Usage (defaults):
//
//	app.Use(middleware.SecurityHeaders())
//
// Usage (custom):
//
//	app.Use(middleware.SecurityHeaders(middleware.SecurityOptions{
//	    ContentSecurityPolicy: "default-src 'self'; img-src *",
//	    XFrameOptions:         "SAMEORIGIN",
//	    StrictTransportSecurity: "-", // omit HSTS (e.g. during local dev)
//	}))
func SecurityHeaders(opts ...SecurityOptions) common.MiddlewareFunc {
	var o SecurityOptions
	if len(opts) > 0 {
		o = opts[0]
	}

	resolve := func(override, defaultVal string) string {
		switch override {
		case "-":
			return "" // caller explicitly disabled this header
		case "":
			return defaultVal
		default:
			return override
		}
	}

	csp := resolve(o.ContentSecurityPolicy, "default-src 'self'")
	xfo := resolve(o.XFrameOptions, "DENY")
	xcto := resolve(o.XContentTypeOptions, "nosniff")
	xxss := resolve(o.XXSSProtection, "1; mode=block")
	hsts := resolve(o.StrictTransportSecurity, "max-age=31536000; includeSubDomains")
	rp := resolve(o.ReferrerPolicy, "strict-origin-when-cross-origin")
	pp := resolve(o.PermissionsPolicy, "geolocation=(), microphone=(), camera=()")

	return func(next common.HandlerFunc) common.HandlerFunc {
		return func(ctx *common.Context) error {
			h := ctx.Writer.Header()

			if csp != "" {
				h.Set("Content-Security-Policy", csp)
			}
			if xfo != "" {
				h.Set("X-Frame-Options", xfo)
			}
			if xcto != "" {
				h.Set("X-Content-Type-Options", xcto)
			}
			if xxss != "" {
				h.Set("X-XSS-Protection", xxss)
			}
			if hsts != "" {
				h.Set("Strict-Transport-Security", hsts)
			}
			if rp != "" {
				h.Set("Referrer-Policy", rp)
			}
			if pp != "" {
				h.Set("Permissions-Policy", pp)
			}

			return next(ctx)
		}
	}
}
