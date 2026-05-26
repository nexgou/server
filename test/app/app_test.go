package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nexgou/server/src/app"
	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/core"
)

// ── minimal module & controller helpers ──────────────────────────────────────

type helloController struct{}

func (c *helloController) Register() []common.Route {
	return []common.Route{
		{Method: http.MethodGet, Path: "/hello", Handler: func(ctx *common.Context) error {
			return ctx.JSON(http.StatusOK, common.H{"msg": "hello"})
		}},
	}
}

func newHelloController() *helloController { return &helloController{} }

func helloModule() common.IModule {
	return core.NewModule(common.ModuleOptions{
		Controllers: []any{newHelloController},
	})
}

// ── CreateApp & Handler ───────────────────────────────────────────────────────

func TestCreateApp_DefaultRoot(t *testing.T) {
	a := app.CreateApp(helloModule())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	// Default root is overrideable but the hello module doesn't override GET /
	// so the built-in handler should respond 200.
	if w.Code != http.StatusOK {
		t.Errorf("default root: got %d, want 200", w.Code)
	}
}

func TestCreateApp_ControllerRoute(t *testing.T) {
	a := app.CreateApp(helloModule())
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("controller route: got %d, want 200", w.Code)
	}
}

func TestApp_Use(t *testing.T) {
	var ran bool
	mw := func(next common.HandlerFunc) common.HandlerFunc {
		return func(ctx *common.Context) error {
			ran = true
			return next(ctx)
		}
	}
	a := app.CreateApp(helloModule())
	a.Use(mw)
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	a.Handler().ServeHTTP(httptest.NewRecorder(), req)
	if !ran {
		t.Error("middleware was not executed")
	}
}

func TestApp_SetFilter(t *testing.T) {
	a := app.CreateApp(core.NewModule(common.ModuleOptions{
		Controllers: []any{func() *errController { return &errController{} }},
	}))
	a.SetFilter(&captureFilter{})

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	// Filter writes 418 on any error
	if w.Code != http.StatusTeapot {
		t.Errorf("filter: got %d, want 418", w.Code)
	}
}

// ── module tree ───────────────────────────────────────────────────────────────

func TestCreateApp_NestedModules(t *testing.T) {
	inner := core.NewModule(common.ModuleOptions{
		Controllers: []any{newHelloController},
	})
	outer := core.NewModule(common.ModuleOptions{
		Imports: []common.IModule{inner},
	})
	a := app.CreateApp(outer)
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("nested modules: got %d, want 200", w.Code)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

type errController struct{}

func (c *errController) Register() []common.Route {
	return []common.Route{
		{Method: http.MethodGet, Path: "/boom", Handler: func(ctx *common.Context) error {
			return common.NewNotFoundException("not found")
		}},
	}
}

type captureFilter struct{}

func (f *captureFilter) Catch(_ error, ctx *common.Context) error {
	ctx.Writer.WriteHeader(http.StatusTeapot)
	return nil
}
