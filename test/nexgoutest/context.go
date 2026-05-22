package nexgoutest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nexgou/server/src/common"
)

// ContextOption configures a synthetic *common.Context for unit testing.
type ContextOption func(*contextConfig)

type contextConfig struct {
	method  string
	path    string
	body    []byte
	headers map[string]string
	params  map[string]string
}

// WithMethod sets the HTTP method (default: "GET").
func WithMethod(method string) ContextOption {
	return func(c *contextConfig) { c.method = method }
}

// WithPath sets the request URL path (default: "/").
func WithPath(path string) ContextOption {
	return func(c *contextConfig) { c.path = path }
}

// WithBody sets the raw request body bytes.
func WithBody(body []byte) ContextOption {
	return func(c *contextConfig) { c.body = body }
}

// WithJSONBody sets the raw JSON request body.
func WithJSONBody(json string) ContextOption {
	return func(c *contextConfig) { c.body = []byte(json) }
}

// WithHeader sets a single request header.
func WithHeader(key, value string) ContextOption {
	return func(c *contextConfig) { c.headers[key] = value }
}

// WithParam sets a single URL path parameter (e.g. "id" → "42").
func WithParam(key, value string) ContextOption {
	return func(c *contextConfig) { c.params[key] = value }
}

// testResponseWriter is a minimal http.ResponseWriter that records the response.
// It wraps httptest.ResponseRecorder so we can retrieve status/body after the handler.
type testContext struct {
	*common.Context
	recorder *httptest.ResponseRecorder
}

// NewContext creates a *common.Context backed by a synthetic HTTP request,
// suitable for unit-testing individual handlers without an HTTP server.
//
//	ctx := nexgoutest.NewContext(t,
//	    nexgoutest.WithMethod("POST"),
//	    nexgoutest.WithPath("/users"),
//	    nexgoutest.WithJSONBody(`{"name":"Alice"}`),
//	    nexgoutest.WithHeader("Authorization", "Bearer token"),
//	)
func NewContext(t *testing.T, opts ...ContextOption) *TestContext {
	t.Helper()

	cfg := &contextConfig{
		method:  "GET",
		path:    "/",
		headers: make(map[string]string),
		params:  make(map[string]string),
	}
	for _, o := range opts {
		o(cfg)
	}

	req := httptest.NewRequest(cfg.method, cfg.path, bytes.NewReader(cfg.body))
	for k, v := range cfg.headers {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	ctx := common.NewContext(rec, req, cfg.params)

	return &TestContext{Context: ctx, recorder: rec}
}

// TestContext wraps *common.Context and exposes assertion helpers.
type TestContext struct {
	*common.Context
	recorder *httptest.ResponseRecorder
}

// Assert returns a ContextAssertion for fluent post-handler assertions.
func (tc *TestContext) Assert(t *testing.T) *ContextAssertion {
	t.Helper()
	return &ContextAssertion{t: t, rec: tc.recorder}
}

// ContextAssertion provides fluent assertions on a unit-test handler response.
type ContextAssertion struct {
	t   *testing.T
	rec *httptest.ResponseRecorder
}

// Status asserts the HTTP response status code and returns itself for chaining.
func (a *ContextAssertion) Status(code int) *ContextAssertion {
	a.t.Helper()
	if a.rec.Code != code {
		a.t.Errorf("expected status %d, got %d", code, a.rec.Code)
	}
	return a
}

// BodyContains asserts that the response body contains the given substring.
func (a *ContextAssertion) BodyContains(sub string) *ContextAssertion {
	a.t.Helper()
	body := a.rec.Body.String()
	if !contains(body, sub) {
		a.t.Errorf("expected body to contain %q, got: %s", sub, body)
	}
	return a
}

// BodyEquals asserts that the response body exactly matches the given string.
func (a *ContextAssertion) BodyEquals(expected string) *ContextAssertion {
	a.t.Helper()
	body := a.rec.Body.String()
	if body != expected {
		a.t.Errorf("expected body %q, got %q", expected, body)
	}
	return a
}

// Header asserts that the response header key equals value.
func (a *ContextAssertion) Header(key, value string) *ContextAssertion {
	a.t.Helper()
	got := a.rec.Header().Get(key)
	if got != value {
		a.t.Errorf("expected header %s=%q, got %q", key, value, got)
	}
	return a
}

// Body returns the raw response body string for custom assertions.
func (a *ContextAssertion) Body() string {
	return a.rec.Body.String()
}

// StatusCode returns the recorded status code for custom assertions.
func (a *ContextAssertion) StatusCode() int {
	return a.rec.Code
}

// ResponseHeader returns the recorded response header for custom assertions.
func (a *ContextAssertion) ResponseHeader() http.Header {
	return a.rec.Header()
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && searchString(s, sub))
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
