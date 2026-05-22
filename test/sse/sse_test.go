package nexgousse_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	nexgousse "github.com/nexgou/server/src/sse"
	"github.com/nexgou/server/src/common"
)

// flushRecorder is an httptest.ResponseRecorder that also implements http.Flusher.
type flushRecorder struct {
	*httptest.ResponseRecorder
	flushed int
}

func (f *flushRecorder) Flush() { f.flushed++ }

func newFlushRecorder() *flushRecorder {
	return &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
}

// ── ToHTTPHandler ──────────────────────────────────────────────────────────────

func TestToHTTPHandler_NoFlusher(t *testing.T) {
	// httptest.ResponseRecorder does NOT implement http.Flusher in standard library
	// but actually it does since Go 1.20. Use a plain struct to ensure the non-flusher path.
	handler := nexgousse.ToHTTPHandler(func(ctx *nexgousse.SSEContext) error {
		return ctx.Send("hello")
	})

	// Use a plain ResponseWriter that doesn't support Flush.
	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	w := &noFlushWriter{header: make(http.Header)}
	ctx := common.NewContext(w, req, nil)

	err := handler(ctx)
	if err == nil {
		t.Fatal("expected error when ResponseWriter does not implement http.Flusher")
	}
}

func TestToHTTPHandler_WithFlusher(t *testing.T) {
	var called bool
	handler := nexgousse.ToHTTPHandler(func(ctx *nexgousse.SSEContext) error {
		called = true
		return ctx.Send("hello")
	})

	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	w := newFlushRecorder()
	ctx := common.NewContext(w, req, nil)

	if err := handler(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("SSE handler was not called")
	}
}

// ── SSEContext methods ─────────────────────────────────────────────────────────

func runSSEHandler(t *testing.T, fn nexgousse.HandlerFunc, params map[string]string) *flushRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	w := newFlushRecorder()
	ctx := common.NewContext(w, req, params)
	handler := nexgousse.ToHTTPHandler(fn)
	if err := handler(ctx); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	return w
}

func TestSSEContext_Send(t *testing.T) {
	w := runSSEHandler(t, func(ctx *nexgousse.SSEContext) error {
		return ctx.Send("hello world")
	}, nil)
	body := w.Body.String()
	if !strings.Contains(body, "data: hello world") {
		t.Errorf("Send: body %q does not contain 'data: hello world'", body)
	}
}

func TestSSEContext_SendNamed(t *testing.T) {
	w := runSSEHandler(t, func(ctx *nexgousse.SSEContext) error {
		return ctx.SendNamed("user.created", `{"id":1}`)
	}, nil)
	body := w.Body.String()
	if !strings.Contains(body, "event: user.created") {
		t.Errorf("SendNamed: missing event line, body=%q", body)
	}
	if !strings.Contains(body, `data: {"id":1}`) {
		t.Errorf("SendNamed: missing data line, body=%q", body)
	}
}

func TestSSEContext_SendJSON(t *testing.T) {
	w := runSSEHandler(t, func(ctx *nexgousse.SSEContext) error {
		return ctx.SendJSON(map[string]any{"cpu": 42})
	}, nil)
	body := w.Body.String()
	if !strings.Contains(body, "data:") {
		t.Errorf("SendJSON: missing data prefix, body=%q", body)
	}
	if !strings.Contains(body, `"cpu"`) {
		t.Errorf("SendJSON: missing cpu key, body=%q", body)
	}
}

func TestSSEContext_SendNamedJSON(t *testing.T) {
	w := runSSEHandler(t, func(ctx *nexgousse.SSEContext) error {
		return ctx.SendNamedJSON("metrics", map[string]any{"mem": 80})
	}, nil)
	body := w.Body.String()
	if !strings.Contains(body, "event: metrics") {
		t.Errorf("SendNamedJSON: missing event line, body=%q", body)
	}
}

func TestSSEContext_SendComment(t *testing.T) {
	w := runSSEHandler(t, func(ctx *nexgousse.SSEContext) error {
		return ctx.SendComment("keepalive")
	}, nil)
	body := w.Body.String()
	if !strings.Contains(body, ": keepalive") {
		t.Errorf("SendComment: body=%q", body)
	}
}

func TestSSEContext_SetRetry(t *testing.T) {
	w := runSSEHandler(t, func(ctx *nexgousse.SSEContext) error {
		return ctx.SetRetry(3000)
	}, nil)
	if !strings.Contains(w.Body.String(), "retry: 3000") {
		t.Errorf("SetRetry: body=%q", w.Body.String())
	}
}

func TestSSEContext_SetID(t *testing.T) {
	w := runSSEHandler(t, func(ctx *nexgousse.SSEContext) error {
		return ctx.SetID("42")
	}, nil)
	if !strings.Contains(w.Body.String(), "id: 42") {
		t.Errorf("SetID: body=%q", w.Body.String())
	}
}

func TestSSEContext_Param(t *testing.T) {
	runSSEHandler(t, func(ctx *nexgousse.SSEContext) error {
		if ctx.Param("topic") != "golang" {
			return common.NewBadRequestException("wrong param")
		}
		return nil
	}, map[string]string{"topic": "golang"})
}

func TestSSEContext_Header(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	req.Header.Set("Last-Event-ID", "99")
	w := newFlushRecorder()
	ctx := common.NewContext(w, req, nil)
	handler := nexgousse.ToHTTPHandler(func(sseCtx *nexgousse.SSEContext) error {
		if sseCtx.Header("Last-Event-ID") != "99" {
			return common.NewBadRequestException("wrong header")
		}
		if sseCtx.LastEventID() != "99" {
			return common.NewBadRequestException("wrong last event id")
		}
		return nil
	})
	if err := handler(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSEContext_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	w := newFlushRecorder()
	ctx := common.NewContext(w, req, nil)
	handler := nexgousse.ToHTTPHandler(func(sseCtx *nexgousse.SSEContext) error {
		// RemoteAddr is set by httptest; just ensure it doesn't panic.
		_ = sseCtx.RemoteAddr()
		return nil
	})
	if err := handler(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSEContext_Headers(t *testing.T) {
	w := runSSEHandler(t, func(ctx *nexgousse.SSEContext) error {
		return nil
	}, nil)
	resp := w.Result()
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Errorf("Content-Type: got %q", resp.Header.Get("Content-Type"))
	}
	if resp.Header.Get("Cache-Control") != "no-cache" {
		t.Errorf("Cache-Control: got %q", resp.Header.Get("Cache-Control"))
	}
}

func TestSSEContext_Done(t *testing.T) {
	// context.Background().Done() returns a nil channel — that is valid Go stdlib
	// behaviour for a non-cancelable context. We just verify Done() doesn't panic.
	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	w := newFlushRecorder()
	ctx := common.NewContext(w, req, nil)
	handler := nexgousse.ToHTTPHandler(func(sseCtx *nexgousse.SSEContext) error {
		_ = sseCtx.Done() // must not panic
		return nil
	})
	if err := handler(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSEContext_SendJSON_MarshalError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	w := newFlushRecorder()
	ctx := common.NewContext(w, req, nil)
	handler := nexgousse.ToHTTPHandler(func(sseCtx *nexgousse.SSEContext) error {
		// channels cannot be marshalled to JSON
		return sseCtx.SendJSON(make(chan int))
	})
	err := handler(ctx)
	if err == nil {
		t.Fatal("expected marshal error for chan type")
	}
}

func TestSSEContext_SendNamedJSON_MarshalError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	w := newFlushRecorder()
	ctx := common.NewContext(w, req, nil)
	handler := nexgousse.ToHTTPHandler(func(sseCtx *nexgousse.SSEContext) error {
		return sseCtx.SendNamedJSON("ev", make(chan int))
	})
	err := handler(ctx)
	if err == nil {
		t.Fatal("expected marshal error for chan type")
	}
}

// ── noFlushWriter ─────────────────────────────────────────────────────────────

type noFlushWriter struct {
	header http.Header
	code   int
	body   strings.Builder
}

func (w *noFlushWriter) Header() http.Header         { return w.header }
func (w *noFlushWriter) WriteHeader(code int)        { w.code = code }
func (w *noFlushWriter) Write(b []byte) (int, error) { return w.body.Write(b) }
