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
	app.Use(middleware.Recovery())
	app.Use(middleware.SecurityHeaders())
	app.Use(middleware.CorsWithOptions(middleware.CorsOptions{
		AllowedOrigins: []string{"*"},
		MaxAge:         600,
	}))
	app.Use(middleware.RateLimit(200, time.Minute))
	app.Use(middleware.Timeout(30 * time.Second))
	app.Use(middleware.BodyLimit(1 << 20)) // 1 MB
	app.Use(middleware.Logger())

	app.SetFilter(&filter.HttpExceptionFilter{})

	if err := app.Listen(3001); err != nil {
		log.Fatal(err)
	}
}
