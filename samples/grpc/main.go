package main

import (
	"log"

	nexgou "github.com/nexgou/server"
	"github.com/nexgou/server/samples/grpc/greeter"
)

// AppModule is the root module of the gRPC sample application.
// It imports the GreeterModule which registers both HTTP health routes
// and gRPC Greeter service routes.
var AppModule = nexgou.Module(nexgou.ModuleOptions{
	Imports: []nexgou.IModule{
		nexgou.ConfigModule,
		nexgou.LogModule,
		greeter.Module,
	},
})

func main() {
	app := nexgou.CreateApp(AppModule)

	// Start gRPC server on port 50051 in a goroutine.
	go func() {
		if err := app.ListenGRPC(50051); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// Start HTTP server on port 3003 (health + REST companion routes).
	if err := app.Listen(3003); err != nil {
		log.Fatal(err)
	}
}
