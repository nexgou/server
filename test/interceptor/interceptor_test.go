package interceptor_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/interceptor"
)

type tagInterceptor struct{ tag string }

func (i *tagInterceptor) Intercept(ctx *common.Context, next common.HandlerFunc) error {
	ctx.Writer.Header().Add("X-Order", i.tag)
	return next(ctx)
}

func newCtx() *common.Context {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	return common.NewContext(httptest.NewRecorder(), r, nil)
}

func handler(ctx *common.Context) error {
	ctx.Writer.Header().Add("X-Order", "handler")
	return nil
}

func TestExecute_NoInterceptors(t *testing.T) {
	called := false
	ctx := newCtx()
	err := interceptor.Execute(ctx, func(c *common.Context) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("no interceptors: unexpected error: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestExecute_Order(t *testing.T) {
	ctx := newCtx()
	err := interceptor.Execute(ctx, handler,
		&tagInterceptor{"a"},
		&tagInterceptor{"b"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	values := ctx.Writer.(interface{ Header() http.Header }).Header()["X-Order"]
	// Expected: a, b, handler (outermost first)
	if len(values) != 3 || values[0] != "a" || values[1] != "b" || values[2] != "handler" {
		t.Errorf("order: got %v, want [a b handler]", values)
	}
}

func TestExecute_Single(t *testing.T) {
	ctx := newCtx()
	err := interceptor.Execute(ctx, handler, &tagInterceptor{"only"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
