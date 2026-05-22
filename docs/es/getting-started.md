# Primeros Pasos

> **[← Volver al README](../../README.es.md)**

---

## Tabla de Contenidos

- [Requisitos previos](#requisitos-previos)
- [Instalación](#instalación)
- [Estructura del proyecto](#estructura-del-proyecto)
- [Ciclo de arranque](#ciclo-de-arranque)
- [Primera aplicación](#primera-aplicación)
- [Ejecutar el servidor](#ejecutar-el-servidor)
- [Próximos pasos](#próximos-pasos)

---

## Requisitos previos

- **Go 1.21** o superior

```bash
go version  # go version go1.21.x ...
```

---

## Instalación

```bash
go get github.com/nexgou/server
```

Esto instala el núcleo del framework. Los sub-paquetes adicionales (`src/middleware`, `src/filter`, `test/nexgoutest`) forman parte del mismo módulo y se importan según sea necesario.

---

## Estructura del proyecto

Nexgou no impone una estructura de directorios obligatoria, pero un layout típico basado en funcionalidades es:

```
myapp/
├── main.go              # Punto de entrada — conecta middleware e inicia el servidor
├── app.module.go        # Módulo raíz — importa todos los módulos de funcionalidades
├── user/
│   ├── user.module.go   # Módulo de funcionalidad
│   ├── user.controller.go
│   └── user.service.go
└── order/
    ├── order.module.go
    ├── order.controller.go
    └── order.service.go
```

Cada funcionalidad vive en su propio paquete (directorio). El módulo raíz (`AppModule`) importa todos los módulos de funcionalidades.

---

## Ciclo de arranque

Cuando se llama a `nexgou.CreateApp(root)`, el framework:

1. Recorre el árbol de módulos de forma recursiva (primero en profundidad, `Imports` primero)
2. Construye un contenedor IoC por módulo — registra todos los constructores de `Providers`
3. Detecta las exportaciones de proveedores y las pone a disposición de los módulos importadores
4. Instancia todos los `Controllers` a través del contenedor (inyecta dependencias automáticamente)
5. Llama a `Register()` en los controladores HTTP → registra rutas
6. Llama a `RegisterWS()` en los controladores WebSocket → registra rutas de actualización WS
7. Llama a `RegisterGRPC()` en los controladores gRPC → registra descriptores de servicio gRPC
8. Devuelve un `*App` listo — aún no hay puertos de red abiertos

`app.Listen(port)` / `app.ListenGRPC(port)` abren los listeners de red reales.

---

## Primera aplicación

### 1. Inicializar un módulo Go

```bash
mkdir myapp && cd myapp
go mod init myapp
go get github.com/nexgou/server
```

### 2. Crear el servicio

```go
// user/user.service.go
package user

type UserService struct{}

func NewUserService() *UserService {
    return &UserService{}
}

type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

func (s *UserService) FindAll() []User {
    return []User{
        {ID: "1", Name: "Alice"},
        {ID: "2", Name: "Bob"},
    }
}

func (s *UserService) FindOne(id string) *User {
    return &User{ID: id, Name: "Alice"}
}
```

### 3. Crear el controlador

```go
// user/user.controller.go
package user

import nexgou "github.com/nexgou/server"

type UserController struct {
    svc *UserService
}

func NewUserController(svc *UserService) *UserController {
    return &UserController{svc: svc}
}

func (c *UserController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.Get("/users",     c.FindAll),
        nexgou.Post("/users",    c.Create),
        nexgou.Get("/users/:id", c.FindOne),
    }
}

func (c *UserController) FindAll(ctx *nexgou.Context) error {
    return ctx.JSON(200, c.svc.FindAll())
}

func (c *UserController) Create(ctx *nexgou.Context) error {
    var body struct {
        Name string `json:"name"`
    }
    if err := ctx.Body(&body); err != nil {
        return nexgou.BadRequestException("cuerpo inválido")
    }
    return ctx.JSON(201, nexgou.H{"message": "creado", "name": body.Name})
}

func (c *UserController) FindOne(ctx *nexgou.Context) error {
    id := ctx.Param("id")
    return ctx.JSON(200, c.svc.FindOne(id))
}
```

### 4. Crear el módulo de funcionalidad

```go
// user/user.module.go
package user

import nexgou "github.com/nexgou/server"

var UserModule = nexgou.Module(nexgou.ModuleOptions{
    Controllers: []any{NewUserController},
    Providers:   []any{NewUserService},
})
```

### 5. Crear el módulo raíz

```go
// app.module.go
package main

import nexgou "github.com/nexgou/server"
import "myapp/user"

var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{
        nexgou.LogModule,
        nexgou.ConfigModule,
        user.UserModule,
    },
})
```

### 6. Crear el punto de entrada

```go
// main.go
package main

import (
    "log"
    "time"

    nexgou "github.com/nexgou/server"
    "github.com/nexgou/server/src/filter"
    "github.com/nexgou/server/src/middleware"
)

func main() {
    app := nexgou.CreateApp(AppModule)

    // Middleware (el orden importa — se ejecuta de arriba a abajo por cada solicitud)
    app.Use(middleware.Recovery())          // recuperar de panics → 500
    app.Use(middleware.SecurityHeaders())   // cabeceras HTTP seguras
    app.Use(middleware.CorsWithOptions(middleware.CorsOptions{
        AllowedOrigins: []string{"*"},
    }))
    app.Use(middleware.RateLimit(100, time.Minute))  // 100 req/min por IP
    app.Use(middleware.Timeout(30 * time.Second))    // deadline de 30s
    app.Use(middleware.BodyLimit(1 << 20))           // límite de 1 MB en el cuerpo
    app.Use(middleware.Logger())                     // registro de solicitudes

    // Filtro global de excepciones — errores JSON estructurados
    app.SetFilter(&filter.HttpExceptionFilter{})

    if err := app.Listen(3000); err != nil {
        log.Fatal(err)
    }
}
```

---

## Ejecutar el servidor

```bash
go run .
```

El banner de inicio imprimirá todas las rutas registradas. Probar con curl:

```bash
curl http://localhost:3000/users
curl http://localhost:3000/users/42
curl -X POST http://localhost:3000/users -H 'Content-Type: application/json' -d '{"name":"Charlie"}'
```

---

## Próximos pasos

| Tema | Guía |
|:---|:---|
| Sistema de módulos, DI, exportaciones | [Módulos](modules.md) |
| Versionado de rutas, guards, interceptores, pipes | [Controladores](controllers.md) |
| Referencia completa de middleware | [Middleware](middleware.md) |
| Cabeceras de seguridad, limitación de tasa, timeout | [Seguridad](security.md) |
| WebSocket en tiempo real | [WebSocket](websocket.md) |
| Eventos enviados por el servidor | [Server-Sent Events](sse.md) |
| gRPC (sin `.proto`) | [gRPC](grpc.md) |
| Variables de entorno y configuración | [Config](config.md) |
| Logging estructurado | [Logger](logger.md) |
| Tests unitarios e integración | [Testing](testing.md) |
| Ejemplo completo funcional | [`samples/api`](../samples/api) |
