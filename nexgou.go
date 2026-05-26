// Package nexgou is a high-performance Go framework for building efficient,
// scalable server-side applications, inspired by NestJS.
//
// Users import this single package to access the full public API:
//
//	import nexgou "github.com/nexgou/server"
//
//	func main() {
//	    app := nexgou.CreateApp(AppModule)
//	    app.Listen(":3000")
//	}
package nexgou

import (
	"os"
	"strings"

	fasthttpadapter "github.com/nexgou/server/src/adapters/fasthttp"
	appPkg "github.com/nexgou/server/src/app"
	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/config"
	"github.com/nexgou/server/src/core"
	"github.com/nexgou/server/src/filter"
	"github.com/nexgou/server/src/logger"
	"github.com/nexgou/server/src/middleware"
	"github.com/nexgou/server/src/pipe"
)

// ── Type Aliases ──────────────────────────────────────────────────────────────
// All core types are aliased here so callers only import this package.

type (
	// H is a shorthand for map[string]any, used for JSON responses.
	H = common.H

	// Context holds the state of a single HTTP request/response cycle.
	Context = common.Context

	// HandlerFunc is the function signature for all route handlers.
	HandlerFunc = common.HandlerFunc

	// MiddlewareFunc wraps a HandlerFunc to implement middleware.
	MiddlewareFunc = common.MiddlewareFunc

	// Route defines an HTTP route binding (method + path + handler).
	Route = common.Route

	// Controller is implemented by all controllers to register their routes.
	Controller = common.Controller

	// Guard determines whether a request should proceed to the handler.
	Guard = common.Guard

	// Interceptor intercepts execution before and after the handler.
	Interceptor = common.Interceptor

	// Pipe validates and transforms input values before the handler.
	Pipe = common.Pipe

	// ExceptionFilter catches errors and returns structured HTTP responses.
	ExceptionFilter = common.ExceptionFilter

	// IModule is implemented by all Nexgou modules.
	IModule = common.IModule

	// ModuleOptions defines the composition of a module.
	ModuleOptions = common.ModuleOptions

	// HttpException represents a structured HTTP error.
	HttpException = common.HttpException

	// BannerConfig configures the startup banner.
	BannerConfig = common.BannerConfig

	// App is the Nexgou application instance.
	App = appPkg.App

	// HttpExceptionFilter is the built-in exception filter.
	HttpExceptionFilter = filter.HttpExceptionFilter

	// SecurityOptions allows overriding security header values.
	SecurityOptions = middleware.SecurityOptions

	// ParseIntPipe validates and parses strings as integers.
	ParseIntPipe = pipe.ParseIntPipe

	// ParseUUIDPipe validates UUID-shaped strings.
	ParseUUIDPipe = pipe.ParseUUIDPipe

	// DefaultValuePipe returns a fallback value for empty input.
	DefaultValuePipe = pipe.DefaultValuePipe

	// ── Config ────────────────────────────────────────────────────────────────

	// ConfigService provides typed access to environment variables.
	ConfigService = config.ConfigService

	// ── Logger ────────────────────────────────────────────────────────────────

	// LoggerService is a structured logger injectable via the IoC container.
	LoggerService = logger.LoggerService

	// ScopedLogger is a LoggerService bound to a named context.
	ScopedLogger = logger.ScopedLogger

	// LoggerOptions configures a logger instance.
	LoggerOptions = logger.LoggerOptions

	// LoggerLevel represents a logging severity level.
	LoggerLevel = logger.Level

	// LoggerFormat represents the logger output format.
	LoggerFormat = logger.Format
)

const (
	LevelDebug  = logger.LevelDebug
	LevelInfo   = logger.LevelInfo
	LevelWarn   = logger.LevelWarn
	LevelError  = logger.LevelError
	LevelSilent = logger.LevelSilent

	FormatText = logger.FormatText
	FormatJSON = logger.FormatJSON
)

// ── Route Helpers ─────────────────────────────────────────────────────────────

// Get creates a GET route binding.
func Get(path string, handler HandlerFunc) Route {
	return Route{Method: "GET", Path: path, Handler: handler}
}

// Post creates a POST route binding.
func Post(path string, handler HandlerFunc) Route {
	return Route{Method: "POST", Path: path, Handler: handler}
}

// Put creates a PUT route binding.
func Put(path string, handler HandlerFunc) Route {
	return Route{Method: "PUT", Path: path, Handler: handler}
}

// Patch creates a PATCH route binding.
func Patch(path string, handler HandlerFunc) Route {
	return Route{Method: "PATCH", Path: path, Handler: handler}
}

// Delete creates a DELETE route binding.
func Delete(path string, handler HandlerFunc) Route {
	return Route{Method: "DELETE", Path: path, Handler: handler}
}

// ── Module ────────────────────────────────────────────────────────────────────

// Module creates a new Nexgou module from the given options.
func Module(opts ModuleOptions) IModule {
	return core.NewModule(opts)
}

// ── Application ───────────────────────────────────────────────────────────────

// CreateApp initializes a new Nexgou application from the root module,
// resolves all dependencies, and registers all controller routes.
func CreateApp(root IModule) *App {
	return appPkg.CreateApp(root)
}

// ListenAndServe prints startup information and serves app through the fasthttp adapter.
func ListenAndServe(address string, app *App) error {
	PrintBanner(startupBannerConfig(address))
	PrintRoutes(app)
	return fasthttpadapter.ListenAndServe(address, app.Handler())
}

// PrintBanner writes the Nexgou startup banner to stdout.
func PrintBanner(config BannerConfig) {
	common.PrintBanner(config)
}

// PrintRoutes writes the registered HTTP routes to stdout.
func PrintRoutes(app *App) {
	app.PrintRoutes()
}

func startupBannerConfig(address string) BannerConfig {
	return BannerConfig{
		AppName:     firstEnv("NEXGOU_APP_NAME", "SERVICE_NAME", "Nexgou"),
		Description: "A high-performance Go framework for building scalable server-side applications.",
		Version:     firstEnv("NEXGOU_APP_VERSION", "SERVICE_VERSION", "0.1.0"),
		Environment: firstEnv("NEXGOU_ENV", "APP_ENV", "development"),
		Port:        strings.TrimPrefix(address, ":"),
		URL:         "http://localhost" + address,
	}
}

func firstEnv(primary string, secondary string, fallback string) string {
	if value := os.Getenv(primary); value != "" {
		return value
	}
	if value := os.Getenv(secondary); value != "" {
		return value
	}
	return fallback
}

// SecurityHeaders returns middleware that sets secure HTTP response headers.
func SecurityHeaders(opts ...SecurityOptions) MiddlewareFunc {
	return middleware.SecurityHeaders(opts...)
}

// NewLogger creates a LoggerService from explicit options.
func NewLogger(options ...LoggerOptions) *LoggerService {
	return logger.NewLogger(options...)
}

// ParseLoggerLevel parses LOG_LEVEL-compatible values.
func ParseLoggerLevel(value string) LoggerLevel {
	return logger.ParseLoggerLevel(value)
}

// ParseLoggerFormat parses LOG_FORMAT-compatible values.
func ParseLoggerFormat(value string) LoggerFormat {
	return logger.ParseLoggerFormat(value)
}

// ── Exception Helpers ─────────────────────────────────────────────────────────

// Exception creates an HTTP exception with a custom status code and message.
func Exception(status int, message string) *HttpException {
	return common.NewHttpException(status, message)
}

// BadRequestException creates a 400 Bad Request exception.
func BadRequestException(message string) *HttpException {
	return common.NewBadRequestException(message)
}

// UnauthorizedException creates a 401 Unauthorized exception.
func UnauthorizedException(message string) *HttpException {
	return common.NewUnauthorizedException(message)
}

// ForbiddenException creates a 403 Forbidden exception.
func ForbiddenException(message string) *HttpException {
	return common.NewForbiddenException(message)
}

// NotFoundException creates a 404 Not Found exception.
func NotFoundException(message string) *HttpException {
	return common.NewNotFoundException(message)
}

// InternalServerErrorException creates a 500 Internal Server Error exception.
func InternalServerErrorException(message string) *HttpException {
	return common.NewInternalServerErrorException(message)
}

// ── Config & Logger modules ───────────────────────────────────────────────────

// ConfigModule is a ready-to-use module that registers and exports ConfigService.
// Import it in any module that needs environment variable access.
var ConfigModule = config.ConfigModule

// LogModule is a ready-to-use module that registers and exports LoggerService.
// Import it in any module that needs structured logging.
var LogModule = logger.LogModule
