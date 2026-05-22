package nexgoutest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appPkg "github.com/nexgou/server/src/app"
	"github.com/nexgou/server/src/common"
)

// TestSuite wraps a real httptest.Server backed by a Nexgou application.
// Use NewSuite to create one and always defer suite.Close() in your tests.
type TestSuite struct {
	server *httptest.Server
	client *http.Client
}

// NewSuite creates a TestSuite from a Nexgou root module.
// It starts an httptest.Server and configures a matching HTTP client.
//
//	suite := nexgoutest.NewSuite(t, AppModule)
//	defer suite.Close()
func NewSuite(t *testing.T, root common.IModule) *TestSuite {
	t.Helper()
	nexgouApp := appPkg.CreateApp(root)
	server := httptest.NewServer(nexgouApp.Handler())
	return &TestSuite{
		server: server,
		client: server.Client(),
	}
}

// Close shuts down the underlying httptest.Server.
func (s *TestSuite) Close() {
	s.server.Close()
}

// URL returns the base URL of the test server (e.g. "http://127.0.0.1:PORT").
func (s *TestSuite) URL() string {
	return s.server.URL
}

// GET starts a GET RequestBuilder for the given path.
func (s *TestSuite) GET(path string) *RequestBuilder {
	return s.newBuilder("GET", path)
}

// POST starts a POST RequestBuilder for the given path.
func (s *TestSuite) POST(path string) *RequestBuilder {
	return s.newBuilder("POST", path)
}

// PUT starts a PUT RequestBuilder for the given path.
func (s *TestSuite) PUT(path string) *RequestBuilder {
	return s.newBuilder("PUT", path)
}

// PATCH starts a PATCH RequestBuilder for the given path.
func (s *TestSuite) PATCH(path string) *RequestBuilder {
	return s.newBuilder("PATCH", path)
}

// DELETE starts a DELETE RequestBuilder for the given path.
func (s *TestSuite) DELETE(path string) *RequestBuilder {
	return s.newBuilder("DELETE", path)
}

func (s *TestSuite) newBuilder(method, path string) *RequestBuilder {
	return &RequestBuilder{
		suite:   s,
		method:  method,
		path:    path,
		headers: make(map[string]string),
	}
}

// RequestBuilder constructs and executes an HTTP request against the test server.
type RequestBuilder struct {
	suite   *TestSuite
	method  string
	path    string
	headers map[string]string
	body    string
}

// Header adds a request header.
func (b *RequestBuilder) Header(key, value string) *RequestBuilder {
	b.headers[key] = value
	return b
}

// Body sets the request body (for POST/PUT/PATCH).
func (b *RequestBuilder) Body(body string) *RequestBuilder {
	b.body = body
	return b
}

// JSONBody sets the request body and Content-Type: application/json.
func (b *RequestBuilder) JSONBody(json string) *RequestBuilder {
	b.body = json
	b.headers["Content-Type"] = "application/json"
	return b
}

// Do executes the HTTP request and returns a ResponseAssertion.
func (b *RequestBuilder) Do(t *testing.T) *ResponseAssertion {
	t.Helper()
	url := b.suite.server.URL + b.path

	var bodyReader io.Reader
	if b.body != "" {
		bodyReader = strings.NewReader(b.body)
	}

	req, err := http.NewRequest(b.method, url, bodyReader)
	if err != nil {
		t.Fatalf("nexgoutest: failed to create request: %v", err)
	}
	for k, v := range b.headers {
		req.Header.Set(k, v)
	}

	resp, err := b.suite.client.Do(req)
	if err != nil {
		t.Fatalf("nexgoutest: request failed: %v", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("nexgoutest: failed to read response body: %v", err)
	}

	return &ResponseAssertion{
		t:      t,
		status: resp.StatusCode,
		body:   string(rawBody),
		header: resp.Header,
	}
}

// ResponseAssertion provides fluent assertions on a live HTTP response.
type ResponseAssertion struct {
	t      *testing.T
	status int
	body   string
	header http.Header
}

// Status asserts the HTTP response status code.
func (a *ResponseAssertion) Status(code int) *ResponseAssertion {
	a.t.Helper()
	if a.status != code {
		a.t.Errorf("expected status %d, got %d\nbody: %s", code, a.status, a.body)
	}
	return a
}

// BodyContains asserts that the response body contains the given substring.
func (a *ResponseAssertion) BodyContains(sub string) *ResponseAssertion {
	a.t.Helper()
	if !contains(a.body, sub) {
		a.t.Errorf("expected body to contain %q\ngot: %s", sub, a.body)
	}
	return a
}

// BodyEquals asserts that the response body exactly matches the given string.
func (a *ResponseAssertion) BodyEquals(expected string) *ResponseAssertion {
	a.t.Helper()
	if a.body != expected {
		a.t.Errorf("expected body %q\ngot %q", expected, a.body)
	}
	return a
}

// Header asserts that the response header key equals value.
func (a *ResponseAssertion) Header(key, value string) *ResponseAssertion {
	a.t.Helper()
	got := a.header.Get(key)
	if got != value {
		a.t.Errorf("expected header %s=%q, got %q", key, value, got)
	}
	return a
}

// Body returns the raw response body string for custom assertions.
func (a *ResponseAssertion) Body() string { return a.body }

// StatusCode returns the response status code for custom assertions.
func (a *ResponseAssertion) StatusCode() int { return a.status }

// ResponseHeader returns the response headers for custom assertions.
func (a *ResponseAssertion) ResponseHeader() http.Header { return a.header }
