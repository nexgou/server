package app

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/core"
	"github.com/nexgou/server/src/router"
	nexgougrpc "github.com/nexgou/server/src/grpc"
	nexgouws "github.com/nexgou/server/src/websocket"
)

// App is the Nexgou application instance. It holds the router, exposes
// the top-level configuration API (Use, SetFilter, Listen), and can
// optionally start a gRPC server alongside the HTTP server.
type App struct {
	router     *router.Router
	grpcServer *nexgougrpc.GRPCServer
}

// CreateApp initializes a Nexgou application from a root module.
// It walks the full module tree, registers all providers in the IoC container,
// instantiates controllers, and binds their routes to the router.
//
// A default GET / route is registered automatically and returns a 200 JSON
// health-check response. Any controller that registers GET / will override it.
func CreateApp(root common.IModule) *App {
	a := &App{router: router.New()}

	// Register the default root route first so module routes take precedence.
	a.router.Add(common.Route{Method: "GET", Path: "/", Handler: defaultRootHandler})

	container := core.NewContainer()
	a.walkModule(root, container)
	return a
}

// defaultRootHandler is the built-in GET / handler.
// It is automatically replaced if a controller registers the same route.
func defaultRootHandler(ctx *common.Context) error {
	return ctx.JSON(200, common.H{
		"framework": "Nexgou",
		"status":    "ok",
	})
}

// Use appends a global middleware to the request pipeline.
// Middlewares are executed in registration order (first registered = outermost).
func (a *App) Use(mw common.MiddlewareFunc) {
	a.router.Use(mw)
}

// SetFilter sets the global exception filter that handles all unhandled errors.
func (a *App) SetFilter(f common.ExceptionFilter) {
	a.router.SetFilter(f)
}

// Listen starts the HTTP server on the given port with an optional host IP.
//
//	app.Listen(3000)                 // listens on 0.0.0.0:3000
//	app.Listen(3000, "127.0.0.1")   // listens on 127.0.0.1:3000
func (a *App) Listen(port int, ip ...string) error {
	host := ""
	if len(ip) > 0 && ip[0] != "" {
		host = ip[0]
	}
	addr := fmt.Sprintf("%s:%d", host, port)

	displayHost := host
	if displayHost == "" {
		displayHost = "localhost"
	}
	common.PrintBanner(common.BannerConfig{
		AppName:     "Nexgou",
		Description: "A high-performance Go framework for building scalable server-side applications.",
		Version:     "0.1.0",
		Environment: "development",
		Port:        fmt.Sprintf("%d", port),
		URL:         fmt.Sprintf("http://%s:%d", displayHost, port),
	})
	a.printRoutes()
	return http.ListenAndServe(addr, a.router)
}

// ListenGRPC starts a gRPC server on the given port (default host 0.0.0.0).
// It must be called from a separate goroutine when combined with Listen:
//
//	go app.ListenGRPC(50051)
//	app.Listen(3000)
//
// Returns an error if the port cannot be bound or if no gRPC controllers were
// registered in the module tree (the gRPC server would be empty).
func (a *App) ListenGRPC(port int, ip ...string) error {
	if a.grpcServer == nil {
		return fmt.Errorf("nexgou: ListenGRPC called but no GRPCController was registered in the module tree")
	}
	return a.grpcServer.Listen(port, ip...)
}

// Handler returns the underlying http.Handler so the app can be mounted on
// a custom server (e.g. httptest.Server for integration tests).
func (a *App) Handler() http.Handler {
	return a.router
}

// printRoutes logs all registered routes to stdout after the banner.
// Columns are dynamically padded to align without visible borders.
func (a *App) printRoutes() {
	methodColor := map[string]string{
		"GET":    "\033[32m",
		"POST":   "\033[34m",
		"PUT":    "\033[33m",
		"PATCH":  "\033[35m",
		"DELETE": "\033[31m",
	}
	reset := "\033[0m"
	dim   := "\033[2m"
	gray  := "\033[90m"

	routes := a.router.Routes()

	// Calculate column widths from actual data.
	maxMethod := 6
	maxPath   := 15
	for _, r := range routes {
		if len(r.Method) > maxMethod {
			maxMethod = len(r.Method)
		}
		if len(r.Path) > maxPath {
			maxPath = len(r.Path)
		}
	}

	fmt.Println("Mapped routes:")
	fmt.Printf("%s────────────────────────────────%s\n", gray, reset)
	for _, r := range routes {
		color, ok := methodColor[r.Method]
		if !ok {
			color = "\033[37m"
		}

		// Badge column
		badge := ""
		if len(r.Guards) == 0 {
			badge = fmt.Sprintf("%s🌐 public%s", gray, reset)
		} else {
			badge = fmt.Sprintf("🔒 %s%s%s", dim, strings.Join(r.Guards, ", "), reset)
		}
		if len(r.Interceptors) > 0 {
			badge += fmt.Sprintf("   ⚡ %s%s%s", dim, strings.Join(r.Interceptors, ", "), reset)
		}

		method := fmt.Sprintf("%-*s", maxMethod, r.Method)
		path   := fmt.Sprintf("%-*s", maxPath, r.Path)

		fmt.Printf("%s%s%s   %s%s%s   %s\n", color, method, reset, dim, path, reset, badge)
	}

	// Print WebSocket routes.
	wsRoutes := a.router.WSRoutes()
	if len(wsRoutes) > 0 {
		cyan := "\033[36m"
		for _, r := range wsRoutes {
			badge := ""
			if len(r.Guards) == 0 {
				badge = fmt.Sprintf("%s🌐 public%s", gray, reset)
			} else {
				badge = fmt.Sprintf("🔒 %s%s%s", dim, strings.Join(r.Guards, ", "), reset)
			}
			wspath := fmt.Sprintf("%-*s", maxPath, r.Path)
			fmt.Printf("%sWS    %s   %s%s%s   %s\n", cyan, reset, dim, wspath, reset, badge)
		}
	}

	fmt.Println()
}

// walkModule recursively processes the module tree:
// registers providers, resolves and wires controllers.
func (a *App) walkModule(m common.IModule, c *core.Container) {
	opts := m.Options()

	// Recurse into imported modules first so their providers are available.
	for _, imp := range opts.Imports {
		a.walkModule(imp, c)
	}

	// Register all providers in the container.
	for _, p := range opts.Providers {
		c.Register(p)
	}

	// Register controller factories so their return types can be resolved.
	for _, cf := range opts.Controllers {
		c.Register(cf)
	}

	// Resolve each controller and register its routes.
	for _, cf := range opts.Controllers {
		cfType := reflect.TypeOf(cf)
		if cfType == nil || cfType.Kind() != reflect.Func || cfType.NumOut() == 0 {
			panic("nexgou: controller factory must be a function returning a Controller")
		}

		returnType := cfType.Out(0)
		val, err := c.Resolve(returnType)
		if err != nil {
			panic(fmt.Sprintf("nexgou: failed to resolve controller %s: %v", returnType, err))
		}

		ctrl, ok := val.Interface().(common.Controller)
		if !ok {
			panic(fmt.Sprintf("nexgou: %s does not implement the Controller interface", returnType))
		}

		for _, route := range ctrl.Register() {
			a.router.Add(route)
		}

		// If the controller also implements WSController, register its WS routes.
		if wsCtrl, ok := val.Interface().(nexgouws.WSController); ok {
			for _, route := range wsCtrl.RegisterWS() {
				a.router.AddWS(route)
			}
		}

		// If the controller also implements GRPCController, register its gRPC services.
		if grpcCtrl, ok := val.Interface().(nexgougrpc.GRPCController); ok {
			if a.grpcServer == nil {
				a.grpcServer = nexgougrpc.NewGRPCServer()
			}
			for _, route := range grpcCtrl.RegisterGRPC() {
				a.grpcServer.RegisterRoute(route)
			}
		}
	}
}
