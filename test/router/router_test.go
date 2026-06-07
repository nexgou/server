package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/middleware"
	"github.com/nexgou/server/src/router"
)

// ── helpers ────────────────────────────────────────────────────────────────────

func okHandler(ctx *common.Context) error {
	return ctx.JSON(http.StatusOK, common.H{"ok": true})
}

func newRouter(routes ...common.Route) *router.Router {
	r := router.New()
	for _, route := range routes {
		r.Add(route)
	}
	return r
}

// ── basic routing ─────────────────────────────────────────────────────────────

func TestRouter_BasicGET(t *testing.T) {
	r := newRouter(common.Route{Method: http.MethodGet, Path: "/hello", Handler: okHandler})
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GET /hello: got %d, want 200", w.Code)
	}
}

func TestRouter_NotFound(t *testing.T) {
	r := router.New()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("not found: got %d, want 404", w.Code)
	}
}

func TestRouter_MethodMismatch(t *testing.T) {
	r := newRouter(common.Route{Method: http.MethodPost, Path: "/data", Handler: okHandler})
	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("method mismatch: got %d, want 404", w.Code)
	}
}

func TestRouter_ParamExtraction(t *testing.T) {
	r := newRouter(common.Route{
		Method: http.MethodGet,
		Path:   "/users/:id",
		Handler: func(ctx *common.Context) error {
			return ctx.JSON(http.StatusOK, common.H{"id": ctx.Param("id")})
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("param route: got %d, want 200", w.Code)
	}
	if w.Body.String() == "" {
		t.Error("empty body")
	}
}

func TestRouter_VersionedRoute(t *testing.T) {
	r := newRouter(common.Route{Method: http.MethodGet, Path: "/users", Handler: okHandler}.Version("v1"))

	// Must match /v1/users
	req := httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("versioned route: got %d, want 200", w.Code)
	}

	// Must not match /users
	req2 := httptest.NewRequest(http.MethodGet, "/users", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusNotFound {
		t.Errorf("unversioned path should 404, got %d", w2.Code)
	}
}

func TestRouter_RouteOverride(t *testing.T) {
	var called string
	first := common.Route{Method: http.MethodGet, Path: "/ep", Handler: func(ctx *common.Context) error {
		called = "first"
		return nil
	}}
	second := common.Route{Method: http.MethodGet, Path: "/ep", Handler: func(ctx *common.Context) error {
		called = "second"
		return nil
	}}

	r := router.New()
	r.Add(first)
	r.Add(second) // override

	req := httptest.NewRequest(http.MethodGet, "/ep", nil)
	r.ServeHTTP(httptest.NewRecorder(), req)
	if called != "second" {
		t.Errorf("override: got %q, want %q", called, "second")
	}
}

// ── guards ────────────────────────────────────────────────────────────────────

type denyGuard struct{}

func (g *denyGuard) CanActivate(_ *common.Context) (bool, error) { return false, nil }

type errorGuard struct{}

func (g *errorGuard) CanActivate(_ *common.Context) (bool, error) {
	return false, common.NewUnauthorizedException("no token")
}

func TestRouter_GuardDenies(t *testing.T) {
	r := newRouter(common.Route{Method: http.MethodGet, Path: "/secret", Handler: okHandler}.Guard(&denyGuard{}))
	req := httptest.NewRequest(http.MethodGet, "/secret", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("guard deny: got %d, want 403", w.Code)
	}
}

func TestRouter_GuardError(t *testing.T) {
	r := newRouter(common.Route{Method: http.MethodGet, Path: "/auth", Handler: okHandler}.Guard(&errorGuard{}))
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("guard error: got %d, want 401", w.Code)
	}
}

// ── interceptors ──────────────────────────────────────────────────────────────

type addHeaderInterceptor struct{ key, val string }

func (i *addHeaderInterceptor) Intercept(ctx *common.Context, next common.HandlerFunc) error {
	ctx.Writer.Header().Set(i.key, i.val)
	return next(ctx)
}

func TestRouter_Interceptor(t *testing.T) {
	r := newRouter(common.Route{
		Method:  http.MethodGet,
		Path:    "/intercepted",
		Handler: okHandler,
	}.Intercept(&addHeaderInterceptor{"X-Test", "nexgou"}))

	req := httptest.NewRequest(http.MethodGet, "/intercepted", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("X-Test") != "nexgou" {
		t.Errorf("interceptor header: got %q, want %q", w.Header().Get("X-Test"), "nexgou")
	}
}

// ── middleware ────────────────────────────────────────────────────────────────

func TestRouter_Middleware(t *testing.T) {
	var order []string
	mw := func(next common.HandlerFunc) common.HandlerFunc {
		return func(ctx *common.Context) error {
			order = append(order, "before")
			err := next(ctx)
			order = append(order, "after")
			return err
		}
	}

	r := newRouter(common.Route{Method: http.MethodGet, Path: "/mw", Handler: func(ctx *common.Context) error {
		order = append(order, "handler")
		return nil
	}})
	r.Use(mw)

	req := httptest.NewRequest(http.MethodGet, "/mw", nil)
	r.ServeHTTP(httptest.NewRecorder(), req)

	if len(order) != 3 || order[0] != "before" || order[1] != "handler" || order[2] != "after" {
		t.Errorf("middleware order: %v", order)
	}
}

func TestRouter_OptionsNotFound_WithCorsMiddleware(t *testing.T) {
	r := router.New()
	r.Use(middleware.CorsWithOptions(middleware.CorsOptions{}))

	req := httptest.NewRequest(http.MethodOptions, "/skills", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "content-type")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS with CORS middleware: got %d, want 204", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("allow origin: got %q, want *", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("expected Access-Control-Allow-Methods header")
	}
}

func TestRouter_OptionsNotFound_WithoutMiddleware(t *testing.T) {
	r := router.New()
	req := httptest.NewRequest(http.MethodOptions, "/skills", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("OPTIONS without middleware: got %d, want 404", w.Code)
	}
}

// ── exception filter ──────────────────────────────────────────────────────────

type jsonFilter struct{}

func (f *jsonFilter) Catch(err error, ctx *common.Context) error {
	if ex, ok := err.(*common.HttpException); ok {
		_ = ctx.JSON(ex.Status, common.H{"error": ex.Message})
	}
	return nil
}

func TestRouter_ExceptionFilter(t *testing.T) {
	r := router.New()
	r.SetFilter(&jsonFilter{})
	r.Add(common.Route{Method: http.MethodGet, Path: "/err", Handler: func(ctx *common.Context) error {
		return common.NewNotFoundException("user not found")
	}})

	req := httptest.NewRequest(http.MethodGet, "/err", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("filter: got %d, want 404", w.Code)
	}
}

// ── Routes metadata ───────────────────────────────────────────────────────────

func TestRouter_RoutesMetadata(t *testing.T) {
	r := router.New()
	r.Add(common.Route{Method: http.MethodGet, Path: "/a", Handler: okHandler})
	r.Add(common.Route{Method: http.MethodPost, Path: "/b", Handler: okHandler})

	routes := r.Routes()
	if len(routes) != 2 {
		t.Fatalf("Routes: got %d, want 2", len(routes))
	}
	if routes[0].Method != http.MethodGet {
		t.Errorf("Routes[0].Method: got %q, want GET", routes[0].Method)
	}
}

// ── handler error without filter ─────────────────────────────────────────────

func TestRouter_HandlerError_NoFilter(t *testing.T) {
	r := newRouter(common.Route{Method: http.MethodGet, Path: "/fail", Handler: func(ctx *common.Context) error {
		return common.NewInternalServerErrorException("boom")
	}})

	req := httptest.NewRequest(http.MethodGet, "/fail", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("handler error no filter: got %d, want 500", w.Code)
	}
}

func TestRouter_HandlerGenericError_NoFilter(t *testing.T) {
	r := newRouter(common.Route{Method: http.MethodGet, Path: "/generic", Handler: func(ctx *common.Context) error {
		return &customErr{}
	}})

	req := httptest.NewRequest(http.MethodGet, "/generic", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("generic error no filter: got %d, want 500", w.Code)
	}
}

type customErr struct{}

func (e *customErr) Error() string { return "custom" }
