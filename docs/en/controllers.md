# Controllers

> **[← Back to README](../../README.md)**

---

## Table of Contents

- [HTTP Controllers](#http-controllers)
- [Route Helpers](#route-helpers)
- [Context API](#context-api)
- [Route Versioning](#route-versioning)
- [Guards](#guards)
- [Interceptors](#interceptors)
- [Pipes](#pipes)
- [Exception Filters](#exception-filters)

---

## HTTP Controllers

A controller is a struct with a `Register()` method that returns a slice of routes.

```go
type ProductController struct {
    svc *ProductService
}

func NewProductController(svc *ProductService) *ProductController {
    return &ProductController{svc: svc}
}

func (c *ProductController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.Get("/products",      c.List),
        nexgou.Get("/products/:id",  c.Get),
        nexgou.Post("/products",     c.Create),
        nexgou.Patch("/products/:id",c.Update),
        nexgou.Delete("/products/:id",c.Delete),
    }
}
```

Register the controller in a module:

```go
var ProductModule = nexgou.Module(nexgou.ModuleOptions{
    Controllers: []any{NewProductController},
    Providers:   []any{NewProductService},
})
```

---

## Route Helpers

All route helpers return a `nexgou.Route` that supports fluent chaining.

| Function                       | HTTP Method | Example                                 |
| :----------------------------- | :---------- | :-------------------------------------- |
| `nexgou.Get(path, handler)`    | GET         | `nexgou.Get("/users", c.List)`          |
| `nexgou.Post(path, handler)`   | POST        | `nexgou.Post("/users", c.Create)`       |
| `nexgou.Put(path, handler)`    | PUT         | `nexgou.Put("/users/:id", c.Replace)`   |
| `nexgou.Patch(path, handler)`  | PATCH       | `nexgou.Patch("/users/:id", c.Update)`  |
| `nexgou.Delete(path, handler)` | DELETE      | `nexgou.Delete("/users/:id", c.Remove)` |

### URL parameters

Use `:param` syntax in the path:

```go
nexgou.Get("/users/:id/posts/:postId", c.GetPost)

func (c *Controller) GetPost(ctx *nexgou.Context) error {
    userID  := ctx.Param("id")
    postID  := ctx.Param("postId")
    return ctx.JSON(200, nexgou.H{"userId": userID, "postId": postID})
}
```

---

## Context API

`*nexgou.Context` is passed to every handler. It wraps the underlying `*http.Request` and `http.ResponseWriter`.

### Reading the request

```go
func (c *Controller) Handler(ctx *nexgou.Context) error {
    method  := ctx.Method()          // "GET", "POST", ...
    path    := ctx.Path()            // "/users/42"
    id      := ctx.Param("id")       // URL path parameter
    all     := ctx.Params()          // map[string]string of all path params
    token   := ctx.Header("Authorization")  // request header

    // JSON body decoding
    var payload struct {
        Name string `json:"name"`
    }
    if err := ctx.Body(&payload); err != nil {
        return nexgou.BadRequestException("invalid JSON body")
    }

    return nil
}
```

### Writing the response

```go
// JSON response with status
return ctx.JSON(200, nexgou.H{"message": "ok"})

// Any struct
return ctx.JSON(201, &User{ID: "1", Name: "Alice"})

// Errors
return nexgou.NotFoundException("user not found")
return nexgou.BadRequestException("invalid id format")
return nexgou.UnauthorizedException("missing token")
return nexgou.ForbiddenException("insufficient permissions")
return nexgou.InternalServerErrorException("something went wrong")
return nexgou.Exception(422, "Unprocessable Entity")
```

### Full Context API

| Method   | Signature                      | Description                            |
| :------- | :----------------------------- | :------------------------------------- |
| `Method` | `() string`                    | HTTP verb                              |
| `Path`   | `() string`                    | URL path                               |
| `Param`  | `(key string) string`          | Named URL parameter                    |
| `Params` | `() map[string]string`         | All URL parameters (copy)              |
| `Header` | `(key string) string`          | Request header value                   |
| `Body`   | `(target any) error`           | JSON-decode request body into `target` |
| `JSON`   | `(status int, data any) error` | Write JSON response                    |

---

## Route Versioning

Add a version prefix to a route with `.Version()`:

```go
nexgou.Get("/users", c.List).Version("v1")
// Registers as: GET /v1/users

nexgou.Get("/users", c.ListV2).Version("v2")
// Registers as: GET /v2/users
```

The version prefix is prepended to the route path automatically.

---

## Guards

Guards run **before** the handler and decide whether the request should proceed. Return `false` to deny access (the framework returns `403 Forbidden`).

### Implementing a guard

```go
type AuthGuard struct{}

func (g *AuthGuard) CanActivate(ctx *nexgou.Context) (bool, error) {
    token := ctx.Header("Authorization")
    if token == "" {
        return false, nil  // denied → 403
    }
    if !isValidJWT(token) {
        return false, nexgou.UnauthorizedException("invalid token")  // custom error
    }
    return true, nil  // allowed
}
```

### Attaching guards to routes

```go
nexgou.Get("/admin/users", c.ListAll).Guard(&AuthGuard{}, &AdminRoleGuard{})
```

Multiple guards run in order. The request is denied on the first `false`.

### Per-route rate limit guard (built-in)

```go
import "github.com/nexgou/server/src/middleware"

nexgou.Post("/auth/login", c.Login).
    Guard(&middleware.RateLimitGuard{Max: 5, Window: time.Minute})
```

---

## Interceptors

Interceptors wrap the handler — they run code **before and after** the handler executes. This is useful for logging, timing, caching, and response transformation.

### Implementing an interceptor

```go
type TimingInterceptor struct{}

func (i *TimingInterceptor) Intercept(ctx *nexgou.Context, next nexgou.HandlerFunc) error {
    start := time.Now()
    err := next(ctx)  // call the actual handler
    log.Printf("[%s %s] %s", ctx.Method(), ctx.Path(), time.Since(start))
    return err
}
```

### Attaching interceptors to routes

```go
nexgou.Post("/uploads", c.Upload).
    Intercept(
        &middleware.TimeoutInterceptor{Duration: 60 * time.Second},
        &middleware.BodyLimitInterceptor{MaxBytes: 50 << 20}, // 50 MB
    )
```

Multiple interceptors form a nested chain:

```
interceptor1.before → interceptor2.before → handler → interceptor2.after → interceptor1.after
```

### Built-in interceptors

| Interceptor            | Package          | Description                |
| :--------------------- | :--------------- | :------------------------- |
| `TimeoutInterceptor`   | `src/middleware` | Per-route request deadline |
| `BodyLimitInterceptor` | `src/middleware` | Per-route body size cap    |

See [Security](security.md) for details.

---

## Pipes

**Pipes** validate and transform URL parameters or body values before they reach the handler. They are used manually inside handler functions.

### Using built-in pipes

```go
import "github.com/nexgou/server/src/pipe"

func (c *Controller) GetUser(ctx *nexgou.Context) error {
    rawID := ctx.Param("id")

    // ParseIntPipe: validates the param is a valid integer
    idAny, err := (&pipe.ParseIntPipe{}).Transform(rawID)
    if err != nil {
        return err  // returns 400 BadRequest automatically
    }
    id := idAny.(int)

    return ctx.JSON(200, c.svc.FindOne(id))
}
```

### Built-in pipes

| Pipe                               | Description                             | Returns               |
| :--------------------------------- | :-------------------------------------- | :-------------------- |
| `ParseIntPipe`                     | Validates and parses string as `int`    | `int` or 400 error    |
| `ParseUUIDPipe`                    | Validates string is 36-char UUID format | `string` or 400 error |
| `DefaultValuePipe{Default: "..."}` | Returns fallback when input is empty    | `string`              |

### Custom pipe

```go
type ParsePositiveIntPipe struct{}

func (p *ParsePositiveIntPipe) Transform(value string) (any, error) {
    n, err := strconv.Atoi(value)
    if err != nil || n <= 0 {
        return nil, nexgou.BadRequestException("must be a positive integer")
    }
    return n, nil
}
```

---

## Exception Filters

An **exception filter** is a global error handler that intercepts any error returned from a handler (including errors from guards and interceptors).

### Using the built-in filter

```go
import "github.com/nexgou/server/src/filter"

app.SetFilter(&filter.HttpExceptionFilter{})
```

This returns structured JSON for all errors:

```json
// nexgou.NotFoundException("user not found")
{ "statusCode": 404, "message": "user not found" }

// unexpected error (not HttpException)
{ "statusCode": 500, "message": "Internal Server Error" }
```

### Custom exception filter

```go
type AppExceptionFilter struct{}

func (f *AppExceptionFilter) Catch(err error, ctx *nexgou.Context) error {
    if ex, ok := err.(*nexgou.HttpException); ok {
        return ctx.JSON(ex.Status, nexgou.H{
            "error":     ex.Message,
            "timestamp": time.Now().UTC(),
            "path":      ctx.Path(),
        })
    }
    // log the unexpected error internally
    log.Printf("unhandled error: %v", err)
    return ctx.JSON(500, nexgou.H{"error": "Internal Server Error"})
}

app.SetFilter(&AppExceptionFilter{})
```
