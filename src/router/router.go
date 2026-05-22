package router

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/nexgou/server/src/common"
	nexgouws "github.com/nexgou/server/src/websocket"
)

type entry struct {
	method       string
	segments     []string
	handler      common.HandlerFunc
	guards       []common.Guard
	interceptors []common.Interceptor
	version      string
}

// Router handles HTTP routing with support for parameterized path segments (:param),
// and WebSocket routing via AddWS.
type Router struct {
	entries     []entry
	wsEntries   []nexgouws.WSEntry
	middlewares []common.MiddlewareFunc
	filter      common.ExceptionFilter
}

// New creates a new Router.
func New() *Router {
	return &Router{}
}

// Add registers an HTTP route. If a route with the same method and path already
// exists it is replaced, allowing module routes to override framework defaults.
// If the route has a version set, it is prepended to the path (e.g. /v1/users).
func (r *Router) Add(route common.Route) {
	path := route.Path
	if v := route.Ver(); v != "" {
		path = "/" + v + path
	}
	segments := splitPath(path)
	for i, e := range r.entries {
		if e.method == route.Method && segmentsEqual(e.segments, segments) {
			r.entries[i] = entry{
				method:       route.Method,
				segments:     segments,
				handler:      route.Handler,
				guards:       route.Guards,
				interceptors: route.Interceptors,
				version:      route.Ver(),
			}
			return
		}
	}
	r.entries = append(r.entries, entry{
		method:       route.Method,
		segments:     segments,
		handler:      route.Handler,
		guards:       route.Guards,
		interceptors: route.Interceptors,
		version:      route.Ver(),
	})
}

// AddWS registers a WebSocket route.
func (r *Router) AddWS(route nexgouws.WSRoute) {
	r.wsEntries = append(r.wsEntries, nexgouws.NewEntry(route))
}

// RouteInfo holds the public metadata of a registered HTTP route.
type RouteInfo struct {
	Method       string
	Path         string
	Guards       []string
	Interceptors []string
	Version      string
}

// WSRouteInfo holds the public metadata of a registered WebSocket route.
type WSRouteInfo struct {
	Path    string
	Guards  []string
	Version string
}

// Routes returns a snapshot of all registered HTTP routes in registration order.
func (r *Router) Routes() []RouteInfo {
	out := make([]RouteInfo, len(r.entries))
	for i, e := range r.entries {
		out[i] = RouteInfo{
			Method:       e.method,
			Path:         "/" + strings.Join(e.segments, "/"),
			Guards:       typeNames(e.guards),
			Interceptors: typeNames(e.interceptors),
			Version:      e.version,
		}
	}
	return out
}

// WSRoutes returns a snapshot of all registered WebSocket routes.
func (r *Router) WSRoutes() []WSRouteInfo {
	out := make([]WSRouteInfo, len(r.wsEntries))
	for i, e := range r.wsEntries {
		out[i] = WSRouteInfo{
			Path:    e.Route.FullPath(),
			Guards:  typeNames(e.Route.Guards),
			Version: e.Route.Ver(),
		}
	}
	return out
}

// typeNames extracts the struct type name of each element in a slice of interfaces.
func typeNames[T any](items []T) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		t := reflect.TypeOf(item)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		names = append(names, t.Name())
	}
	return names
}

// Use appends a middleware to the global middleware chain.
// First registered middleware is the outermost wrapper.
func (r *Router) Use(mw common.MiddlewareFunc) {
	r.middlewares = append(r.middlewares, mw)
}

// SetFilter sets the global exception filter used to handle handler errors.
func (r *Router) SetFilter(f common.ExceptionFilter) {
	r.filter = f
}

// ServeHTTP implements http.Handler. It first attempts to match WebSocket
// upgrade requests, then falls through to the regular HTTP route table.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	reqSegments := splitPath(req.URL.Path)
	params := make(map[string]string)

	// ── WebSocket upgrade ──────────────────────────────────────────────────────
	if isWebSocketUpgrade(req) {
		for _, e := range r.wsEntries {
			if e.Match(reqSegments, params) {
				ctx := common.NewContext(w, req, params)

				// Run guards before the upgrade — denial responds with HTTP 403.
				for _, g := range e.Route.Guards {
					ok, err := g.CanActivate(ctx)
					if err != nil {
						r.handleError(err, ctx, w)
						return
					}
					if !ok {
						r.handleError(common.NewForbiddenException("Forbidden"), ctx, w)
						return
					}
				}

				// Guards passed — perform the upgrade.
				e.Route.Upgrade(w, req, params)
				return
			}
			// Clear params before trying the next WS entry.
			for k := range params {
				delete(params, k)
			}
		}
	}

	// ── HTTP routes ────────────────────────────────────────────────────────────
	for _, e := range r.entries {
		if e.method != req.Method {
			continue
		}
		if match(e.segments, reqSegments, params) {
			ctx := common.NewContext(w, req, params)

			// Run guards — any denial short-circuits the request.
			for _, g := range e.guards {
				ok, err := g.CanActivate(ctx)
				if err != nil {
					r.handleError(err, ctx, w)
					return
				}
				if !ok {
					r.handleError(common.NewForbiddenException("Forbidden"), ctx, w)
					return
				}
			}

			// Build the final handler wrapped by interceptors (innermost first).
			handler := e.handler
			for i := len(e.interceptors) - 1; i >= 0; i-- {
				ic := e.interceptors[i]
				next := handler
				handler = func(c *common.Context) error {
					return ic.Intercept(c, next)
				}
			}

			// Wrap with global middleware chain (last registered = innermost).
			for i := len(r.middlewares) - 1; i >= 0; i-- {
				handler = r.middlewares[i](handler)
			}

			if err := handler(ctx); err != nil {
				r.handleError(err, ctx, w)
			}
			return
		}
		// Clear params before trying the next route.
		for k := range params {
			delete(params, k)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`{"statusCode":404,"message":"Cannot ` + req.Method + ` ` + req.URL.Path + `"}`))
}

// handleError dispatches an error to the exception filter or falls back to plain HTTP errors.
func (r *Router) handleError(err error, ctx *common.Context, w http.ResponseWriter) {
	if r.filter != nil {
		_ = r.filter.Catch(err, ctx)
	} else if ex, ok := err.(*common.HttpException); ok {
		http.Error(w, ex.Message, ex.Status)
	} else {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// isWebSocketUpgrade reports whether the request is a WebSocket upgrade.
func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

// match checks whether a path matches a pattern and populates params.
func match(pattern, path []string, params map[string]string) bool {
	if len(pattern) != len(path) {
		return false
	}
	for i, seg := range pattern {
		if strings.HasPrefix(seg, ":") {
			params[seg[1:]] = path[i]
		} else if seg != path[i] {
			return false
		}
	}
	return true
}

// splitPath splits a URL path into clean segments, stripping leading/trailing slashes.
func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return []string{}
	}
	return strings.Split(trimmed, "/")
}

// segmentsEqual reports whether two segment slices are identical.
func segmentsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
