package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/middleware"
)

// ── helpers ────────────────────────────────────────────────────────────────────

func okHandler(ctx *common.Context) error {
	ctx.Writer.WriteHeader(http.StatusOK)
	return nil
}

func applyMW(mw common.MiddlewareFunc, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	ctx := common.NewContext(w, req, nil)
	_ = mw(okHandler)(ctx)
	return w
}

// ── CORS ──────────────────────────────────────────────────────────────────────

func TestCors_Wildcard(t *testing.T) {
	mw := middleware.CorsWithOptions(middleware.CorsOptions{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := applyMW(mw, req)
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("wildcard CORS: got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCors_SpecificOrigin_Allowed(t *testing.T) {
	mw := middleware.CorsWithOptions(middleware.CorsOptions{
		AllowedOrigins: []string{"https://example.com"},
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	w := applyMW(mw, req)
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("allowed origin: got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCors_SpecificOrigin_Blocked(t *testing.T) {
	mw := middleware.CorsWithOptions(middleware.CorsOptions{
		AllowedOrigins: []string{"https://example.com"},
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	w := applyMW(mw, req)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("blocked origin: got %q, want empty", got)
	}
}

func TestCors_Preflight(t *testing.T) {
	mw := middleware.CorsWithOptions(middleware.CorsOptions{})
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	w := applyMW(mw, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("preflight: got %d, want 204", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("preflight: missing Access-Control-Allow-Methods")
	}
}

func TestCors_Credentials(t *testing.T) {
	mw := middleware.CorsWithOptions(middleware.CorsOptions{
		AllowCredentials: true,
		AllowedOrigins:   []string{"https://app.example.com"},
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://app.example.com")
	w := applyMW(mw, req)
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("credentials: expected true")
	}
}

func TestCors_ExposedHeaders(t *testing.T) {
	mw := middleware.CorsWithOptions(middleware.CorsOptions{
		ExposedHeaders: []string{"X-Custom-Header"},
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := applyMW(mw, req)
	if !strings.Contains(w.Header().Get("Access-Control-Expose-Headers"), "X-Custom-Header") {
		t.Error("exposed headers: missing X-Custom-Header")
	}
}

// ── Security Headers ──────────────────────────────────────────────────────────

func TestSecurityHeaders_Defaults(t *testing.T) {
	mw := middleware.SecurityHeaders()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := applyMW(mw, req)

	checks := map[string]string{
		"X-Frame-Options":           "DENY",
		"X-Content-Type-Options":    "nosniff",
		"X-XSS-Protection":          "1; mode=block",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Content-Security-Policy":   "default-src 'self'",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
	}
	for header, want := range checks {
		if got := w.Header().Get(header); got != want {
			t.Errorf("%s: got %q, want %q", header, got, want)
		}
	}
}

func TestSecurityHeaders_Override(t *testing.T) {
	mw := middleware.SecurityHeaders(middleware.SecurityOptions{
		XFrameOptions:           "SAMEORIGIN",
		StrictTransportSecurity: "-", // omit
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := applyMW(mw, req)

	if w.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Error("override: X-Frame-Options should be SAMEORIGIN")
	}
	if w.Header().Get("Strict-Transport-Security") != "" {
		t.Error("omit: Strict-Transport-Security should be absent")
	}
}

// ── Body Limit ─────────────────────────────────────────────────────────────────

func TestBodyLimit_Under(t *testing.T) {
	mw := middleware.BodyLimit(1 << 10) // 1 KB
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("small body"))
	req.Header.Set("Content-Type", "application/json")
	w := applyMW(mw, req)
	if w.Code != http.StatusOK {
		t.Errorf("body under limit: got %d, want 200", w.Code)
	}
}

func TestBodyLimitInterceptor(t *testing.T) {
	li := &middleware.BodyLimitInterceptor{MaxBytes: 5}
	handler := func(ctx *common.Context) error {
		return li.Intercept(ctx, okHandler)
	}
	mw := func(next common.HandlerFunc) common.HandlerFunc { return next }

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("small"))
	w := httptest.NewRecorder()
	ctx := common.NewContext(w, req, nil)
	_ = mw(handler)(ctx)
}

// ── Rate Limit ────────────────────────────────────────────────────────────────

func TestRateLimit_Allows(t *testing.T) {
	mw := middleware.RateLimit(10, time.Minute)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:9999"
	w := applyMW(mw, req)
	if w.Code != http.StatusOK {
		t.Errorf("rate limit allows: got %d, want 200", w.Code)
	}
	if w.Header().Get("X-RateLimit-Limit") != "10" {
		t.Errorf("X-RateLimit-Limit: got %q", w.Header().Get("X-RateLimit-Limit"))
	}
}

func TestRateLimit_Exceeds(t *testing.T) {
	mw := middleware.RateLimit(2, time.Minute)
	ip := "10.0.0.2:5678"

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip
		w := httptest.NewRecorder()
		ctx := common.NewContext(w, req, nil)
		err := mw(okHandler)(ctx)

		if i < 2 {
			if err != nil {
				t.Errorf("request %d: unexpected error: %v", i, err)
			}
		} else {
			if err == nil {
				t.Errorf("request %d: expected rate limit error, got nil", i)
			}
			if ex, ok := err.(*common.HttpException); !ok || ex.Status != http.StatusTooManyRequests {
				t.Errorf("request %d: expected 429, got %v", i, err)
			}
		}
	}
}

func TestRateLimitGuard_Allows(t *testing.T) {
	g := &middleware.RateLimitGuard{Max: 5, Window: time.Minute}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:80"
	ctx := common.NewContext(httptest.NewRecorder(), req, nil)
	ok, err := g.CanActivate(ctx)
	if err != nil {
		t.Fatalf("guard error: %v", err)
	}
	if !ok {
		t.Error("guard should allow")
	}
}

func TestRateLimitGuard_Exceeds(t *testing.T) {
	g := &middleware.RateLimitGuard{Max: 1, Window: time.Minute}
	ip := "172.16.0.1:80"
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip
		ctx := common.NewContext(httptest.NewRecorder(), req, nil)
		ok, _ := g.CanActivate(ctx)
		if i == 0 && !ok {
			t.Error("first request: should be allowed")
		}
		if i == 1 && ok {
			t.Error("second request: should be denied")
		}
	}
}

func TestRateLimit_XForwardedFor(t *testing.T) {
	mw := middleware.RateLimit(10, time.Minute)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	w := applyMW(mw, req)
	if w.Code != http.StatusOK {
		t.Errorf("XFF: got %d", w.Code)
	}
}

func TestRateLimit_XRealIP(t *testing.T) {
	mw := middleware.RateLimit(10, time.Minute)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "198.51.100.1")
	w := applyMW(mw, req)
	if w.Code != http.StatusOK {
		t.Errorf("XRealIP: got %d", w.Code)
	}
}

// ── Logger ────────────────────────────────────────────────────────────────────

func TestLogger_Runs(t *testing.T) {
	mw := middleware.Logger()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := applyMW(mw, req)
	if w.Code != http.StatusOK {
		t.Errorf("Logger: got %d, want 200", w.Code)
	}
}

// ── Recovery ──────────────────────────────────────────────────────────────────

func TestRecovery_NoPanic(t *testing.T) {
	mw := middleware.Recovery()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := applyMW(mw, req)
	if w.Code != http.StatusOK {
		t.Errorf("Recovery (no panic): got %d, want 200", w.Code)
	}
}

func TestRecovery_PanicReturns500(t *testing.T) {
	mw := middleware.Recovery()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ctx := common.NewContext(w, req, nil)
	err := mw(func(ctx *common.Context) error {
		panic("test panic")
	})(ctx)
	if err == nil {
		t.Fatal("expected error from recovered panic")
	}
	if ex, ok := err.(*common.HttpException); !ok || ex.Status != http.StatusInternalServerError {
		t.Errorf("recovery error: got %v, want 500", err)
	}
}

// ── Cors (simple) ─────────────────────────────────────────────────────────────

func TestCors_Simple(t *testing.T) {
	mw := middleware.Cors()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := applyMW(mw, req)
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Cors: got %q, want *", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Cors: missing Allow-Methods")
	}
}

func TestTimeout_WithinDeadline(t *testing.T) {
	mw := middleware.Timeout(5 * time.Second)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := applyMW(mw, req)
	if w.Code != http.StatusOK {
		t.Errorf("timeout within deadline: got %d, want 200", w.Code)
	}
}

func TestTimeout_Exceeds(t *testing.T) {
	mw := middleware.Timeout(10 * time.Millisecond)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ctx := common.NewContext(w, req, nil)

	slowHandler := func(ctx *common.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}
	err := mw(slowHandler)(ctx)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if ex, ok := err.(*common.HttpException); !ok || ex.Status != http.StatusRequestTimeout {
		t.Errorf("timeout error: got %v", err)
	}
}

func TestTimeoutInterceptor(t *testing.T) {
	ti := &middleware.TimeoutInterceptor{Duration: 5 * time.Second}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := common.NewContext(httptest.NewRecorder(), req, nil)
	err := ti.Intercept(ctx, okHandler)
	if err != nil {
		t.Errorf("interceptor: unexpected error: %v", err)
	}
}
