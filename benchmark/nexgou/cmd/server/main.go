package main

import (
	"log"
	"os"

	nexgou "github.com/nexgou/server"
	benchapp "github.com/nexgou/server/benchmark/nexgou/app"
)

func main() {
	config := benchapp.Config{
		ServiceName: env("SERVICE_NAME", "nexgou"),
		Version:     env("SERVICE_VERSION", "2.0.0"),
		DBPath:      env("DB_PATH", "benchmark/nexgou/data/db.sqlite"),
	}
	store, err := benchapp.NewStore(config.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	app := benchapp.NewNexGouApp(config, store)
	if err := nexgou.ListenAndServe(":"+env("PORT", "3001"), app); err != nil {
		log.Fatal(err)
	}
}

func env(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
