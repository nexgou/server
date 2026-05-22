# Testing

> **[← Volver al README](../../README.es.md)**

---

## Tabla de Contenidos

- [Descripción general](#descripción-general)
- [Instalación](#instalación)
- [Tests unitarios con NewContext](#tests-unitarios-con-newcontext)
- [Tests de integración con TestSuite](#tests-de-integración-con-testsuite)
- [RequestBuilder](#requestbuilder)
- [Assertions](#assertions)
- [Probar Guards e Interceptores](#probar-guards-e-interceptores)
- [Probar con módulos reales](#probar-con-módulos-reales)
- [Cobertura](#cobertura)
- [Referencia de la API](#referencia-de-la-api)

---

## Descripción general

Nexgou incluye un paquete de testing dedicado en `github.com/nexgou/server/test/nexgoutest` con dos capas de helpers:

| Capa | Cuándo usar |
|:---|:---|
| **Unitaria** (`NewContext`) | Probar una sola función handler de forma aislada — sin red, sin servidor |
| **Integración** (`TestSuite`) | Probar un módulo/aplicación completo sobre un servidor HTTP real (en proceso) |

Ambas capas usan constructores de assertions fluidas para que los tests sean concisos y legibles.

---

## Instalación

El paquete forma parte del mismo módulo Go — no se necesita `go get` separado:

```go
import "github.com/nexgou/server/test/nexgoutest"
```

---

## Tests unitarios con NewContext

Usar `nexgoutest.NewContext` para crear un `*nexgou.Context` sintético que está conectado a un `httptest.ResponseRecorder`. Pasar el contexto directamente a la función handler y hacer assertions sobre la respuesta grabada.

### Ejemplo básico

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
        t.Fatalf("error inesperado: %v", err)
    }
    tc.Assert(t).
        Status(200).
        BodyContains(`"Alice"`)
}
```

### Con parámetros de ruta

```go
tc := nexgoutest.NewContext(t,
    nexgoutest.WithMethod("GET"),
    nexgoutest.WithPath("/users/42"),
    nexgoutest.WithParam("id", "42"),
)

err := ctrl.FindOne(tc.Context)
tc.Assert(t).Status(200).BodyContains(`"42"`)
```

### Con cuerpo JSON

```go
tc := nexgoutest.NewContext(t,
    nexgoutest.WithMethod("POST"),
    nexgoutest.WithPath("/users"),
    nexgoutest.WithJSONBody(`{"name":"Charlie"}`),
)

err := ctrl.Create(tc.Context)
tc.Assert(t).Status(201).BodyContains("creado")
```

### Con cabeceras

```go
tc := nexgoutest.NewContext(t,
    nexgoutest.WithHeader("Authorization", "Bearer test-token"),
    nexgoutest.WithHeader("X-Request-ID",  "abc-123"),
)
```

### Referencia de opciones de contexto

| Opción | Descripción |
|:---|:---|
| `WithMethod(method string)` | Método HTTP (por defecto: `"GET"`) |
| `WithPath(path string)` | Path de la solicitud (por defecto: `"/"`) |
| `WithBody(body []byte)` | Bytes del cuerpo de la solicitud raw |
| `WithJSONBody(json string)` | Cuerpo JSON + establece `Content-Type: application/json` |
| `WithHeader(key, value string)` | Añadir una cabecera de solicitud |
| `WithParam(key, value string)` | Añadir un parámetro de ruta URL |

---

## Tests de integración con TestSuite

`nexgoutest.NewSuite` arranca un `httptest.Server` real desde un módulo de Nexgou y devuelve un `*TestSuite` con un cliente HTTP preconfigurado.

### Ejemplo básico

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

### Con cabecera de autenticación

```go
suite.POST("/users").
    Header("Authorization", "Bearer valid-token").
    JSONBody(`{"name":"Dave"}`).
    Do(t).
    Status(201).
    BodyContains("creado")
```

### Probar rutas 404 / de error

```go
suite.GET("/users/id-no-existente").
    Do(t).
    Status(404).
    BodyContains("not found")
```

### Ejemplo completo de archivo de test

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

    t.Run("GET /users devuelve lista", func(t *testing.T) {
        suite.GET("/users").
            Do(t).
            Status(200).
            BodyContains("Alice").
            BodyContains("Bob")
    })

    t.Run("GET /users/:id devuelve usuario", func(t *testing.T) {
        suite.GET("/users/1").
            Do(t).
            Status(200).
            BodyContains(`"id":"1"`)
    })

    t.Run("POST /users crea usuario", func(t *testing.T) {
        suite.POST("/users").
            JSONBody(`{"name":"Eve"}`).
            Do(t).
            Status(201)
    })

    t.Run("DELETE /users/:id desconocido devuelve 404", func(t *testing.T) {
        suite.DELETE("/users/999").
            Do(t).
            Status(404)
    })
}
```

---

## RequestBuilder

Los métodos de `TestSuite` (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`) devuelven un `*RequestBuilder` para construcción fluida de solicitudes:

```go
builder := suite.POST("/endpoint")
builder.Header("Authorization", "Bearer token")
builder.Header("X-Trace-ID", "trace-123")
builder.JSONBody(`{"key":"value"}`)
// o cuerpo raw:
builder.Body("raw body string")

assertion := builder.Do(t)
```

| Método | Descripción |
|:---|:---|
| `.Header(key, value string)` | Añadir una cabecera de solicitud |
| `.Body(body string)` | Establecer cuerpo string raw |
| `.JSONBody(json string)` | Establecer cuerpo JSON + `Content-Type: application/json` |
| `.Do(t *testing.T)` | Ejecutar la solicitud; devuelve `*ResponseAssertion` |

---

## Assertions

Tanto `TestContext.Assert()` como `RequestBuilder.Do()` devuelven un constructor de assertions fluido.

### `ContextAssertion` (tests unitarios)

```go
tc.Assert(t).
    Status(200).
    BodyContains(`"name":"Alice"`).
    BodyEquals(`{"name":"Alice"}`).
    Header("Content-Type", "application/json")
```

### `ResponseAssertion` (tests de integración)

```go
suite.GET("/users").Do(t).
    Status(200).
    BodyContains("Alice").
    Header("Content-Type", "application/json")
```

### Leer valores raw

```go
assertion := suite.GET("/users").Do(t)

body       := assertion.Body()           // string
statusCode := assertion.StatusCode()     // int
headers    := assertion.ResponseHeader() // http.Header
```

### Todos los métodos de assertion

| Método | Descripción |
|:---|:---|
| `.Status(code int)` | Verificar código de estado HTTP (llama a `t.Errorf` si no coincide) |
| `.BodyContains(sub string)` | Verificar que el cuerpo contiene la subcadena |
| `.BodyEquals(expected string)` | Verificar que el cuerpo es una coincidencia exacta |
| `.Header(key, value string)` | Verificar el valor de la cabecera de respuesta |
| `.Body() string` | Devolver el cuerpo de respuesta raw |
| `.StatusCode() int` | Devolver el código de estado de la respuesta |
| `.ResponseHeader() http.Header` | Devolver todas las cabeceras de respuesta |

Todos los métodos de assertion devuelven la misma assertion para que puedan encadenarse.

---

## Probar Guards e Interceptores

### Probar una ruta protegida

```go
t.Run("rechaza solicitud sin token", func(t *testing.T) {
    suite.GET("/admin/users").
        Do(t).
        Status(403)
})

t.Run("permite solicitud con token válido", func(t *testing.T) {
    suite.GET("/admin/users").
        Header("Authorization", "Bearer valid-token").
        Do(t).
        Status(200)
})
```

### Probar un guard de forma unitaria

```go
func TestAuthGuard(t *testing.T) {
    guard := &AuthGuard{}

    t.Run("deniega cuando no hay token", func(t *testing.T) {
        tc := nexgoutest.NewContext(t)
        ok, err := guard.CanActivate(tc.Context)
        if ok || err != nil {
            t.Errorf("se esperaba denegado sin error, got ok=%v err=%v", ok, err)
        }
    })

    t.Run("permite token válido", func(t *testing.T) {
        tc := nexgoutest.NewContext(t, nexgoutest.WithHeader("Authorization", "Bearer valid"))
        ok, err := guard.CanActivate(tc.Context)
        if !ok || err != nil {
            t.Errorf("se esperaba permitido, got ok=%v err=%v", ok, err)
        }
    })
}
```

---

## Probar con módulos reales

`NewSuite` acepta cualquier `nexgou.IModule`, incluyendo tu `AppModule` completo. Esto permite ejecutar tests de integración contra el grafo de dependencias real:

```go
suite := nexgoutest.NewSuite(t, AppModule)
defer suite.Close()

// Usa servicios reales, DI real, middleware real (ninguno por defecto en el suite de test)
suite.GET("/users/1").Do(t).Status(200)
```

> Nota: `NewSuite` **no** registra ningún middleware automáticamente. Si tus tests requieren comportamiento de middleware (limitación de tasa, cabeceras de autenticación, etc.), usa un módulo específico de test que envuelva tu módulo de funcionalidad.

---

## Cobertura

Ejecutar tests con el detector de carreras y cobertura:

```bash
go test -race ./test/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

La pipeline de CI impone un mínimo de **65% de cobertura** en el árbol `./test/...`.

---

## Referencia de la API

### `nexgoutest.NewContext`

```go
func NewContext(t *testing.T, opts ...ContextOption) *TestContext
```

### `nexgoutest.NewSuite`

```go
func NewSuite(t *testing.T, root nexgou.IModule) *TestSuite
```

### Métodos de `TestSuite`

```go
func (s *TestSuite) Close()
func (s *TestSuite) URL() string
func (s *TestSuite) GET(path string) *RequestBuilder
func (s *TestSuite) POST(path string) *RequestBuilder
func (s *TestSuite) PUT(path string) *RequestBuilder
func (s *TestSuite) PATCH(path string) *RequestBuilder
func (s *TestSuite) DELETE(path string) *RequestBuilder
```
