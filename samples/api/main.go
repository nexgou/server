package main

import (
	"log"
	"time"

	nexgou "github.com/nexgou/server"
	"github.com/nexgou/server/src/filter"
	"github.com/nexgou/server/src/middleware"
)

func main() {
	app := nexgou.CreateApp(AppModule)

	// ── Security middleware pipeline ───────────────────────────────────────────
	//
	// Order matters:
	//  1. Recovery      — always first, catches panics from any later middleware
	//  2. SecurityHeaders — set before any response can be written
	//  3. Cors          — must run before the handler; handles OPTIONS preflight
	//  4. RateLimit     — reject abusive clients early, before heavier work
	//  5. Timeout       — bound all subsequent work to a deadline
	//  6. BodyLimit     — cap incoming payload size before handlers read the body
	//  7. Logger        — last so it captures the final status (incl. 429, 408…)

	app.Use(middleware.Recovery())

	app.Use(middleware.SecurityHeaders())

	// CorsWithOptions gives full control over the CORS policy.
	// For open APIs use AllowedOrigins: []string{"*"} (the default).
	// For restricted APIs list the exact allowed origins and set AllowCredentials: true.
	app.Use(middleware.CorsWithOptions(middleware.CorsOptions{
		AllowedOrigins: []string{"*"},
		MaxAge:         600,
	}))

	// Global rate limit: 100 requests per IP per minute.
	app.Use(middleware.RateLimit(100, time.Minute))

	// Global request timeout: 30 seconds.
	app.Use(middleware.Timeout(30 * time.Second))

	// Global body size limit: 1 MB.
	app.Use(middleware.BodyLimit(1 << 20))

	app.Use(middleware.Logger())

	// ── Exception filter ───────────────────────────────────────────────────────
	app.SetFilter(&filter.HttpExceptionFilter{})

	if err := app.Listen(3000); err != nil {
		log.Fatal(err)
	}
}
