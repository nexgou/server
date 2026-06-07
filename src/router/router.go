package router

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/nexgou/server/src/common"
)

type entry struct {
	method          string
	segments        []string
	handler         common.HandlerFunc
	compiledHandler common.HandlerFunc
	guards          []common.Guard
	interceptors    []common.Interceptor
	version         string
	hasParams       bool
}

type responseWriteTracker struct {
	http.ResponseWriter
	written bool
}

func (t *responseWriteTracker) Write(b []byte) (int, error) {
	t.written = true
	return t.ResponseWriter.Write(b)
}

func (t *responseWriteTracker) WriteHeader(statusCode int) {
	t.written = true
	t.ResponseWriter.WriteHeader(statusCode)
}

// Router handles HTTP routing with support for parameterized path segments (:param).
type Router struct {
	entries        []entry
	staticRoutes   map[string]map[string]int
	dynamicEntries []int
	middlewares    []common.MiddlewareFunc
	filter         common.ExceptionFilter
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
	newEntry := entry{
		method:       route.Method,
		segments:     segments,
		handler:      route.Handler,
		guards:       route.Guards,
		interceptors: route.Interceptors,
		version:      route.Ver(),
		hasParams:    hasParamSegment(segments),
	}
	newEntry.compiledHandler = r.compileHandler(newEntry)
	for i, e := range r.entries {
		if e.method == route.Method && segmentsEqual(e.segments, segments) {
			r.entries[i] = newEntry
			r.rebuildIndexes()
			return
		}
	}
	r.entries = append(r.entries, newEntry)
	r.rebuildIndexes()
}

// RouteInfo holds the public metadata of a registered HTTP route.
type RouteInfo struct {
	Method       string
	Path         string
	Guards       []string
	Interceptors []string
	Version      string
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
	for i := range r.entries {
		r.entries[i].compiledHandler = r.compileHandler(r.entries[i])
	}
}

// SetFilter sets the global exception filter used to handle handler errors.
func (r *Router) SetFilter(f common.ExceptionFilter) {
	r.filter = f
}

func (r *Router) applyMiddlewares(handler common.HandlerFunc) common.HandlerFunc {
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		handler = r.middlewares[i](handler)
	}
	return handler
}

func (r *Router) tryOptionsWithMiddleware(w http.ResponseWriter, req *http.Request) bool {
	if req.Method != http.MethodOptions {
		return false
	}

	tracker := &responseWriteTracker{ResponseWriter: w}
	ctx := common.NewContext(tracker, req, nil)
	handler := r.applyMiddlewares(func(ctx *common.Context) error {
		return nil
	})

	if err := handler(ctx); err != nil {
		r.handleError(err, ctx, tracker)
		return true
	}

	return tracker.written
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var matched *entry
	var params map[string]string

	if r.staticRoutes != nil {
		methodRoutes := r.staticRoutes[req.Method]
		if index, ok := methodRoutes[req.URL.Path]; ok {
			matched = &r.entries[index]
		}
	} else {
		reqSegments := splitPath(req.URL.Path)
		for index := range r.entries {
			e := &r.entries[index]
			if e.method != req.Method {
				continue
			}
			matchedParams, ok := match(e.segments, reqSegments, e.hasParams)
			if ok {
				matched = e
				params = matchedParams
				break
			}
		}
	}

	if matched == nil && len(r.dynamicEntries) == 0 && !strings.HasSuffix(req.URL.Path, "/") {
		if !r.tryOptionsWithMiddleware(w, req) {
			r.writeNotFound(w, req)
		}
		return
	}

	if matched == nil && r.staticRoutes != nil {
		reqSegments := splitPath(req.URL.Path)

		// ── HTTP routes ────────────────────────────────────────────────────────────
		for _, index := range r.dynamicEntries {
			e := &r.entries[index]
			if e.method != req.Method {
				continue
			}
			matchedParams, ok := match(e.segments, reqSegments, e.hasParams)
			if ok {
				matched = e
				params = matchedParams
				break
			}
		}

		if matched == nil && strings.HasSuffix(req.URL.Path, "/") {
			for index := range r.entries {
				e := &r.entries[index]
				if e.hasParams || e.method != req.Method {
					continue
				}
				if _, ok := match(e.segments, reqSegments, false); ok {
					matched = e
					break
				}
			}
		}
	}

	if matched == nil {
		if !r.tryOptionsWithMiddleware(w, req) {
			r.writeNotFound(w, req)
		}
		return
	}

	ctx := common.NewContext(w, req, params)

	// Run guards — any denial short-circuits the request.
	for _, g := range matched.guards {
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

	if err := matched.compiledHandler(ctx); err != nil {
		r.handleError(err, ctx, w)
	}
}

func (r *Router) writeNotFound(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`{"statusCode":404,"message":"Cannot ` + req.Method + ` ` + req.URL.Path + `"}`))
}

func (r *Router) rebuildIndexes() {
	r.staticRoutes = nil
	r.dynamicEntries = r.dynamicEntries[:0]
	for index, e := range r.entries {
		if e.hasParams {
			r.dynamicEntries = append(r.dynamicEntries, index)
			continue
		}
		if r.staticRoutes == nil {
			r.staticRoutes = make(map[string]map[string]int)
		}
		methodRoutes := r.staticRoutes[e.method]
		if methodRoutes == nil {
			methodRoutes = make(map[string]int)
			r.staticRoutes[e.method] = methodRoutes
		}
		methodRoutes[entryPath(e.segments)] = index
	}
}

func (r *Router) compileHandler(e entry) common.HandlerFunc {
	handler := e.handler
	for i := len(e.interceptors) - 1; i >= 0; i-- {
		ic := e.interceptors[i]
		next := handler
		handler = func(c *common.Context) error {
			return ic.Intercept(c, next)
		}
	}

	return r.applyMiddlewares(handler)
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

// match checks whether a path matches a pattern and populates params.
func match(pattern, path []string, captureParams bool) (map[string]string, bool) {
	if len(pattern) != len(path) {
		return nil, false
	}
	var params map[string]string
	for i, seg := range pattern {
		if strings.HasPrefix(seg, ":") {
			if captureParams && params == nil {
				params = make(map[string]string)
			}
			if params != nil {
				params[seg[1:]] = path[i]
			}
		} else if seg != path[i] {
			return nil, false
		}
	}
	return params, true
}

func hasParamSegment(segments []string) bool {
	for _, seg := range segments {
		if strings.HasPrefix(seg, ":") {
			return true
		}
	}
	return false
}

func entryPath(segments []string) string {
	if len(segments) == 0 {
		return "/"
	}
	return "/" + strings.Join(segments, "/")
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
