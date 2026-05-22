package main

import (
	"log"

	nexgou "github.com/nexgou/server"
	"github.com/nexgou/server/src/filter"
	"github.com/nexgou/server/src/middleware"
)

func main() {
	app := nexgou.CreateApp(AppModule)

	app.Use(middleware.Recovery())
	app.Use(middleware.CorsWithOptions(middleware.CorsOptions{
		AllowedOrigins: []string{"*"},
	}))
	app.Use(middleware.Logger())

	app.SetFilter(&filter.HttpExceptionFilter{})

	if err := app.Listen(3000); err != nil {
		log.Fatal(err)
	}
}
