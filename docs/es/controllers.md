# Controladores

> **[← Volver al README](../../README.es.md)**

---

## Tabla de Contenidos

- [Controladores HTTP](#controladores-http)
- [Helpers de rutas](#helpers-de-rutas)
- [API de Context](#api-de-context)
- [Versionado de rutas](#versionado-de-rutas)
- [Guards](#guards)
- [Interceptores](#interceptores)
- [Pipes](#pipes)
- [Filtros de excepción](#filtros-de-excepción)

---

## Controladores HTTP

Un controlador es un struct con un método `Register()` que devuelve una lista de rutas.

```go
type ProductController struct {
    svc *ProductService
}

func NewProductController(svc *ProductService) *ProductController {
    return &ProductController{svc: svc}
}

func (c *ProductController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.Get("/products",       c.List),
        nexgou.Get("/products/:id",   c.Get),
        nexgou.Post("/products",      c.Create),
        nexgou.Patch("/products/:id", c.Update),
        nexgou.Delete("/products/:id",c.Delete),
    }
}
```

Registrar el controlador en un módulo:

```go
var ProductModule = nexgou.Module(nexgou.ModuleOptions{
    Controllers: []any{NewProductController},
    Providers:   []any{NewProductService},
})
```

---

## Helpers de rutas

Todos los helpers de rutas devuelven un `nexgou.Route` que admite encadenamiento fluido.

| Función                        | Método HTTP | Ejemplo                                 |
| :----------------------------- | :---------- | :-------------------------------------- |
| `nexgou.Get(path, handler)`    | GET         | `nexgou.Get("/users", c.List)`          |
| `nexgou.Post(path, handler)`   | POST        | `nexgou.Post("/users", c.Create)`       |
| `nexgou.Put(path, handler)`    | PUT         | `nexgou.Put("/users/:id", c.Replace)`   |
| `nexgou.Patch(path, handler)`  | PATCH       | `nexgou.Patch("/users/:id", c.Update)`  |
| `nexgou.Delete(path, handler)` | DELETE      | `nexgou.Delete("/users/:id", c.Remove)` |

### Parámetros de URL

Usa la sintaxis `:param` en el path:

```go
nexgou.Get("/users/:id/posts/:postId", c.GetPost)

func (c *Controller) GetPost(ctx *nexgou.Context) error {
    userID  := ctx.Param("id")
    postID  := ctx.Param("postId")
    return ctx.JSON(200, nexgou.H{"userId": userID, "postId": postID})
}
```

---

## API de Context

`*nexgou.Context` se pasa a cada handler. Envuelve el `*http.Request` subyacente y el `http.ResponseWriter`.

### Leer la solicitud

```go
func (c *Controller) Handler(ctx *nexgou.Context) error {
    method  := ctx.Method()          // "GET", "POST", ...
    path    := ctx.Path()            // "/users/42"
    id      := ctx.Param("id")       // parámetro de ruta URL
    all     := ctx.Params()          // map[string]string de todos los parámetros de ruta
    token   := ctx.Header("Authorization")  // cabecera de la solicitud

    // Decodificación del cuerpo JSON
    var payload struct {
        Name string `json:"name"`
    }
    if err := ctx.Body(&payload); err != nil {
        return nexgou.BadRequestException("cuerpo JSON inválido")
    }

    return nil
}
```

### Escribir la respuesta

```go
// Respuesta JSON con estado
return ctx.JSON(200, nexgou.H{"message": "ok"})

// Cualquier struct
return ctx.JSON(201, &User{ID: "1", Name: "Alice"})

// Errores
return nexgou.NotFoundException("usuario no encontrado")
return nexgou.BadRequestException("formato de id inválido")
return nexgou.UnauthorizedException("token faltante")
return nexgou.ForbiddenException("permisos insuficientes")
return nexgou.InternalServerErrorException("algo salió mal")
return nexgou.Exception(422, "Entidad no procesable")
```

### API completa de Context

| Método   | Firma                          | Descripción                                             |
| :------- | :----------------------------- | :------------------------------------------------------ |
| `Method` | `() string`                    | Verbo HTTP                                              |
| `Path`   | `() string`                    | Ruta URL                                                |
| `Param`  | `(key string) string`          | Parámetro URL con nombre                                |
| `Params` | `() map[string]string`         | Todos los parámetros URL (copia)                        |
| `Header` | `(key string) string`          | Valor de cabecera de la solicitud                       |
| `Body`   | `(target any) error`           | Decodificar JSON del cuerpo de la solicitud en `target` |
| `JSON`   | `(status int, data any) error` | Escribir respuesta JSON                                 |

---

## Versionado de rutas

Agrega un prefijo de versión a una ruta con `.Version()`:

```go
nexgou.Get("/users", c.List).Version("v1")
// Se registra como: GET /v1/users

nexgou.Get("/users", c.ListV2).Version("v2")
// Se registra como: GET /v2/users
```

El prefijo de versión se antepone automáticamente al path de la ruta.

---

## Guards

Los guards se ejecutan **antes** del handler y deciden si la solicitud debe continuar. Devolver `false` deniega el acceso (el framework devuelve `403 Forbidden`).

### Implementar un guard

```go
type AuthGuard struct{}

func (g *AuthGuard) CanActivate(ctx *nexgou.Context) (bool, error) {
    token := ctx.Header("Authorization")
    if token == "" {
        return false, nil  // denegado → 403
    }
    if !isValidJWT(token) {
        return false, nexgou.UnauthorizedException("token inválido")  // error personalizado
    }
    return true, nil  // permitido
}
```

### Asociar guards a rutas

```go
nexgou.Get("/admin/users", c.ListAll).Guard(&AuthGuard{}, &AdminRoleGuard{})
```

Los múltiples guards se ejecutan en orden. La solicitud se deniega en el primer `false`.

### Guard de límite de tasa por ruta (incorporado)

```go
import "github.com/nexgou/server/src/middleware"

nexgou.Post("/auth/login", c.Login).
    Guard(&middleware.RateLimitGuard{Max: 5, Window: time.Minute})
```

---

## Interceptores

Los interceptores envuelven el handler — ejecutan código **antes y después** de que el handler se ejecute. Útiles para logging, temporización, caché y transformación de respuestas.

### Implementar un interceptor

```go
type TimingInterceptor struct{}

func (i *TimingInterceptor) Intercept(ctx *nexgou.Context, next nexgou.HandlerFunc) error {
    start := time.Now()
    err := next(ctx)  // llamar al handler real
    log.Printf("[%s %s] %s", ctx.Method(), ctx.Path(), time.Since(start))
    return err
}
```

### Asociar interceptores a rutas

```go
nexgou.Post("/uploads", c.Upload).
    Intercept(
        &middleware.TimeoutInterceptor{Duration: 60 * time.Second},
        &middleware.BodyLimitInterceptor{MaxBytes: 50 << 20}, // 50 MB
    )
```

Los múltiples interceptores forman una cadena anidada:

```
interceptor1.antes → interceptor2.antes → handler → interceptor2.después → interceptor1.después
```

### Interceptores incorporados

| Interceptor            | Paquete          | Descripción                         |
| :--------------------- | :--------------- | :---------------------------------- |
| `TimeoutInterceptor`   | `src/middleware` | Deadline de solicitud por ruta      |
| `BodyLimitInterceptor` | `src/middleware` | Límite de tamaño de cuerpo por ruta |

Ver [Seguridad](security.md) para más detalles.

---

## Pipes

Los **pipes** validan y transforman parámetros URL o valores del cuerpo antes de que lleguen al handler. Se usan manualmente dentro de las funciones handler.

### Usar pipes incorporados

```go
import "github.com/nexgou/server/src/pipe"

func (c *Controller) GetUser(ctx *nexgou.Context) error {
    rawID := ctx.Param("id")

    // ParseIntPipe: valida que el parámetro sea un entero válido
    idAny, err := (&pipe.ParseIntPipe{}).Transform(rawID)
    if err != nil {
        return err  // devuelve 400 BadRequest automáticamente
    }
    id := idAny.(int)

    return ctx.JSON(200, c.svc.FindOne(id))
}
```

### Pipes incorporados

| Pipe                               | Descripción                                              | Devuelve             |
| :--------------------------------- | :------------------------------------------------------- | :------------------- |
| `ParseIntPipe`                     | Valida y parsea string como `int`                        | `int` o error 400    |
| `ParseUUIDPipe`                    | Valida que el string tenga formato UUID de 36 caracteres | `string` o error 400 |
| `DefaultValuePipe{Default: "..."}` | Devuelve el fallback cuando la entrada está vacía        | `string`             |

### Pipe personalizado

```go
type ParsePositiveIntPipe struct{}

func (p *ParsePositiveIntPipe) Transform(value string) (any, error) {
    n, err := strconv.Atoi(value)
    if err != nil || n <= 0 {
        return nil, nexgou.BadRequestException("debe ser un entero positivo")
    }
    return n, nil
}
```

---

## Filtros de excepción

Un **filtro de excepción** es un manejador de errores global que intercepta cualquier error devuelto por un handler (incluidos errores de guards e interceptores).

### Usar el filtro incorporado

```go
import "github.com/nexgou/server/src/filter"

app.SetFilter(&filter.HttpExceptionFilter{})
```

Esto devuelve JSON estructurado para todos los errores:

```json
// nexgou.NotFoundException("usuario no encontrado")
{ "statusCode": 404, "message": "usuario no encontrado" }

// error inesperado (no HttpException)
{ "statusCode": 500, "message": "Internal Server Error" }
```

### Filtro de excepción personalizado

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
    // registrar el error inesperado internamente
    log.Printf("error no manejado: %v", err)
    return ctx.JSON(500, nexgou.H{"error": "Error Interno del Servidor"})
}

app.SetFilter(&AppExceptionFilter{})
```
