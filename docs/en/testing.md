# Testing

> **[← Back to README](../../README.md)**

---

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Unit Testing with NewContext](#unit-testing-with-newcontext)
- [Integration Testing with TestSuite](#integration-testing-with-testsuite)
- [RequestBuilder](#requestbuilder)
- [Assertions](#assertions)
- [Testing Guards & Interceptors](#testing-guards--interceptors)
- [Testing with Real Modules](#testing-with-real-modules)
- [Coverage](#coverage)
- [API Reference](#api-reference)

---

## Overview

Nexgou ships a dedicated testing package at `github.com/nexgou/server/test/nexgoutest` with two layers of helpers:

| Layer | When to use |
|:---|:---|
| **Unit** (`NewContext`) | Test a single handler function in isolation — no network, no server |
| **Integration** (`TestSuite`) | Test a complete module/app over a real (in-process) HTTP server |

Both layers use fluent assertion builders so tests are concise and readable.

---

## Installation

The package is part of the same Go module — no separate `go get` needed:

```go
import "github.com/nexgou/server/test/nexgoutest"
```

---

## Unit Testing with NewContext

Use `nexgoutest.NewContext` to create a synthetic `*nexgou.Context` that is wired to an `httptest.ResponseRecorder`. Pass the context directly to your handler function and assert on the recorded response.

### Basic example

```go
func TestUserController_FindAll(t *testing.T) {
    // Arrange
    svc := &UserService{}
    ctrl := NewUserController(svc)

    tc := nexgoutest.NewContext(t,
        nexgoutest.WithMethod("GET"),
        nexgoutest.WithPath("/users"),
    )

    // Act
    err := ctrl.FindAll(tc.Context)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    tc.Assert(t).
        Status(200).
        BodyContains(`"Alice"`)
}
```

### With route params

```go
tc := nexgoutest.NewContext(t,
    nexgoutest.WithMethod("GET"),
    nexgoutest.WithPath("/users/42"),
    nexgoutest.WithParam("id", "42"),
)

err := ctrl.FindOne(tc.Context)
tc.Assert(t).Status(200).BodyContains(`"42"`)
```

### With JSON body

```go
tc := nexgoutest.NewContext(t,
    nexgoutest.WithMethod("POST"),
    nexgoutest.WithPath("/users"),
    nexgoutest.WithJSONBody(`{"name":"Charlie"}`),
)

err := ctrl.Create(tc.Context)
tc.Assert(t).Status(201).BodyContains("created")
```

### With headers

```go
tc := nexgoutest.NewContext(t,
    nexgoutest.WithHeader("Authorization", "Bearer test-token"),
    nexgoutest.WithHeader("X-Request-ID",  "abc-123"),
)
```

### Context options reference

| Option | Description |
|:---|:---|
| `WithMethod(method string)` | HTTP method (default: `"GET"`) |
| `WithPath(path string)` | Request path (default: `"/"`) |
| `WithBody(body []byte)` | Raw request body bytes |
| `WithJSONBody(json string)` | JSON body + sets `Content-Type: application/json` |
| `WithHeader(key, value string)` | Add a request header |
| `WithParam(key, value string)` | Add a URL path parameter |

---

## Integration Testing with TestSuite

`nexgoutest.NewSuite` boots a real `httptest.Server` from a Nexgou module and returns a `*TestSuite` with a pre-configured HTTP client.

### Basic example

```go
func TestUserModule_Integration(t *testing.T) {
    suite := nexgoutest.NewSuite(t, UserModule)
    defer suite.Close()

    suite.GET("/users").
        Do(t).
        Status(200).
        BodyContains("Alice")
}
```

### With authentication header

```go
suite.POST("/users").
    Header("Authorization", "Bearer valid-token").
    JSONBody(`{"name":"Dave"}`).
    Do(t).
    Status(201).
    BodyContains("created")
```

### Testing 404 / error paths

```go
suite.GET("/users/non-existent-id").
    Do(t).
    Status(404).
    BodyContains("not found")
```

### Full test file example

```go
package user_test

import (
    "testing"

    "github.com/nexgou/server/test/nexgoutest"
    "myapp/user"
)

func TestUserModule(t *testing.T) {
    suite := nexgoutest.NewSuite(t, user.UserModule)
    defer suite.Close()

    t.Run("GET /users returns list", func(t *testing.T) {
        suite.GET("/users").
            Do(t).
            Status(200).
            BodyContains("Alice").
            BodyContains("Bob")
    })

    t.Run("GET /users/:id returns user", func(t *testing.T) {
        suite.GET("/users/1").
            Do(t).
            Status(200).
            BodyContains(`"id":"1"`)
    })

    t.Run("POST /users creates user", func(t *testing.T) {
        suite.POST("/users").
            JSONBody(`{"name":"Eve"}`).
            Do(t).
            Status(201)
    })

    t.Run("DELETE /users/:id unknown returns 404", func(t *testing.T) {
        suite.DELETE("/users/999").
            Do(t).
            Status(404)
    })
}
```

---

## RequestBuilder

`TestSuite` methods (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`) return a `*RequestBuilder` for fluent request construction:

```go
builder := suite.POST("/endpoint")
builder.Header("Authorization", "Bearer token")
builder.Header("X-Trace-ID", "trace-123")
builder.JSONBody(`{"key":"value"}`)
// or raw body:
builder.Body("raw body string")

assertion := builder.Do(t)
```

| Method | Description |
|:---|:---|
| `.Header(key, value string)` | Add a request header |
| `.Body(body string)` | Set raw string body |
| `.JSONBody(json string)` | Set JSON body + `Content-Type: application/json` |
| `.Do(t *testing.T)` | Execute the request; returns `*ResponseAssertion` |

---

## Assertions

Both `TestContext.Assert()` and `RequestBuilder.Do()` return a fluent assertion builder.

### `ContextAssertion` (unit tests)

```go
tc.Assert(t).
    Status(200).
    BodyContains(`"name":"Alice"`).
    BodyEquals(`{"name":"Alice"}`).
    Header("Content-Type", "application/json")
```

### `ResponseAssertion` (integration tests)

```go
suite.GET("/users").Do(t).
    Status(200).
    BodyContains("Alice").
    Header("Content-Type", "application/json")
```

### Reading raw values

```go
assertion := suite.GET("/users").Do(t)

body       := assertion.Body()          // string
statusCode := assertion.StatusCode()    // int
headers    := assertion.ResponseHeader() // http.Header
```

### All assertion methods

| Method | Description |
|:---|:---|
| `.Status(code int)` | Assert HTTP status code (calls `t.Errorf` on mismatch) |
| `.BodyContains(sub string)` | Assert body contains substring |
| `.BodyEquals(expected string)` | Assert body is exact match |
| `.Header(key, value string)` | Assert response header value |
| `.Body() string` | Return raw response body |
| `.StatusCode() int` | Return response status code |
| `.ResponseHeader() http.Header` | Return all response headers |

All assertion methods return the same assertion so they can be chained.

---

## Testing Guards & Interceptors

### Testing a guarded route

```go
t.Run("rejects request without token", func(t *testing.T) {
    suite.GET("/admin/users").
        Do(t).
        Status(403)
})

t.Run("allows request with valid token", func(t *testing.T) {
    suite.GET("/admin/users").
        Header("Authorization", "Bearer valid-token").
        Do(t).
        Status(200)
})
```

### Unit-testing a guard directly

```go
func TestAuthGuard(t *testing.T) {
    guard := &AuthGuard{}

    t.Run("denies when no token", func(t *testing.T) {
        tc := nexgoutest.NewContext(t)
        ok, err := guard.CanActivate(tc.Context)
        if ok || err != nil {
            t.Errorf("expected denied with no error, got ok=%v err=%v", ok, err)
        }
    })

    t.Run("allows valid token", func(t *testing.T) {
        tc := nexgoutest.NewContext(t, nexgoutest.WithHeader("Authorization", "Bearer valid"))
        ok, err := guard.CanActivate(tc.Context)
        if !ok || err != nil {
            t.Errorf("expected allowed, got ok=%v err=%v", ok, err)
        }
    })
}
```

---

## Testing with Real Modules

`NewSuite` accepts any `nexgou.IModule`, including your full `AppModule`. This lets you run integration tests against the real dependency graph:

```go
suite := nexgoutest.NewSuite(t, AppModule)
defer suite.Close()

// Uses real services, real DI, real middleware (none by default in test suite)
suite.GET("/users/1").Do(t).Status(200)
```

> Note: `NewSuite` does **not** register any middleware automatically. If your tests require middleware behavior (rate limiting, authentication headers, etc.), use a test-specific module that wraps your feature module.

---

## Coverage

Run tests with the race detector and coverage:

```bash
go test -race ./test/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

The CI pipeline enforces a minimum of **80% coverage** on the `./test/...` tree.

---

## API Reference

### `nexgoutest.NewContext`

```go
func NewContext(t *testing.T, opts ...ContextOption) *TestContext
```

### `nexgoutest.NewSuite`

```go
func NewSuite(t *testing.T, root nexgou.IModule) *TestSuite
```

### `TestSuite` methods

```go
func (s *TestSuite) Close()
func (s *TestSuite) URL() string
func (s *TestSuite) GET(path string) *RequestBuilder
func (s *TestSuite) POST(path string) *RequestBuilder
func (s *TestSuite) PUT(path string) *RequestBuilder
func (s *TestSuite) PATCH(path string) *RequestBuilder
func (s *TestSuite) DELETE(path string) *RequestBuilder
```
