package common

// H is a shorthand for map[string]any, used for JSON responses.
type H map[string]any

// HandlerFunc is the function signature for all route handlers.
type HandlerFunc func(*Context) error

// MiddlewareFunc wraps a HandlerFunc to implement middleware.
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Route defines an HTTP route binding with optional metadata.
type Route struct {
	Method       string
	Path         string
	Handler      HandlerFunc
	Guards       []Guard
	Interceptors []Interceptor
	ver          string // use the Version() builder method to set this
}

// Ver returns the version label of the route.
func (r Route) Ver() string { return r.ver }

// Guard attaches one or more guards to the route.
// Guards are evaluated before the handler executes.
//
//	nexgou.Get("/users", handler).Guard(&AuthGuard{}, &RoleGuard{})
func (r Route) Guard(guards ...Guard) Route {
	r.Guards = append(r.Guards, guards...)
	return r
}

// Intercept attaches one or more interceptors to the route.
// Interceptors wrap the handler execution (before + after).
//
//	nexgou.Get("/users", handler).Intercept(&LogInterceptor{})
func (r Route) Intercept(interceptors ...Interceptor) Route {
	r.Interceptors = append(r.Interceptors, interceptors...)
	return r
}

// Version sets a version label for the route (e.g. "v1", "v2").
// The version is displayed in the startup route log.
//
//	nexgou.Get("/users", handler).Version("v1")
func (r Route) Version(v string) Route {
	r.ver = v
	return r
}

// Controller is implemented by all controllers to register their routes.
type Controller interface {
	Register() []Route
}

// Guard determines whether a given request should be handled by the route handler.
type Guard interface {
	CanActivate(*Context) (bool, error)
}

// Interceptor intercepts request/response before and after the handler executes.
type Interceptor interface {
	Intercept(*Context, HandlerFunc) error
}

// Pipe validates and transforms input values before they reach the handler.
type Pipe interface {
	Transform(string) (any, error)
}

// ExceptionFilter catches errors thrown by handlers and returns structured responses.
type ExceptionFilter interface {
	Catch(error, *Context) error
}

// IModule is implemented by all Nexgou modules.
type IModule interface {
	Options() *ModuleOptions
}

// ModuleOptions defines the composition of a module.
type ModuleOptions struct {
	// Imports lists modules whose exported providers are available in this module.
	Imports []IModule
	// Controllers lists constructor functions that return a Controller.
	Controllers []any
	// Providers lists constructor functions for services, repositories, etc.
	Providers []any
	// Exports lists a subset of Providers to expose to other modules that import this one.
	Exports []any
}
