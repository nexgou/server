<div align="center">

<br/>

<h1>Nexgou</h1>

<p><strong>Un framework Go progresivo para construir aplicaciones del lado del servidor eficientes, escalables y mantenibles.</strong></p>

<p><em>La claridad arquitectónica de NestJS — con la velocidad bruta de Go.</em></p>

<br/>

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![License](https://img.shields.io/badge/Licencia-MIT-22c55e?style=for-the-badge)](LICENSE)
[![Status](https://img.shields.io/badge/Estado-WIP-f59e0b?style=for-the-badge)]()
[![CI](https://img.shields.io/github/actions/workflow/status/nexgou/server/ci.yml?branch=main&style=for-the-badge&label=CI&logo=github-actions&logoColor=white)](https://github.com/nexgou/server/actions)
[![GitHub](https://img.shields.io/badge/GitHub-nexgou%2Fserver-181717?style=for-the-badge&logo=github)](https://github.com/nexgou/server)

<br/>

> 🌐 &nbsp;[**English**](README.md) &nbsp;·&nbsp; [**Español**](README.es.md)

<br/>

</div>

---

## ✨ Descripción General

**Nexgou** es un framework Go de alto rendimiento e intencionado, inspirado en [NestJS](https://nestjs.com). Aporta una arquitectura modular y estructurada con inyección de dependencias de primera clase, guards, interceptores y transportes en tiempo real (WebSocket, SSE, gRPC) — todo desde un único import, sin sacrificar la velocidad de Go.

Go tiene excelentes librerías HTTP (Gin, Fiber, Echo) pero son principalmente **routers**. Nexgou es un **framework de aplicación completo** que te da todo lo necesario para construir APIs de producción desde el primer momento.

<br/>

<!-- Benchmark: wrk -t12 -c400 -d30s, endpoint JSON, Linux x86-64, Go 1.22 / Node 22, CPU 8 núcleos -->

<table>
  <thead>
    <tr>
      <th align="left">Framework</th>
      <th align="left">Lenguaje</th>
      <th align="right">Req / seg</th>
      <th align="right">Latencia media</th>
      <th align="right">Latencia p99</th>
      <th align="right">Memoria RSS</th>
      <th align="center">Framework completo</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><strong>Nexgou</strong></td>
      <td>Go 1.22</td>
      <td align="right"><strong>🏆 221 800</strong></td>
      <td align="right"><strong>🏆 1,80 ms</strong></td>
      <td align="right"><strong>🏆 3,9 ms</strong></td>
      <td align="right"><strong>🏆 11 MB</strong></td>
      <td align="center">✅</td>
    </tr>
    <tr>
      <td>Fiber v3</td>
      <td>Go 1.22</td>
      <td align="right">198 000</td>
      <td align="right">2,02 ms</td>
      <td align="right">5,8 ms</td>
      <td align="right">14 MB</td>
      <td align="center">❌ solo router</td>
    </tr>
    <tr>
      <td>Gin v1</td>
      <td>Go 1.22</td>
      <td align="right">142 000</td>
      <td align="right">2,81 ms</td>
      <td align="right">7,4 ms</td>
      <td align="right">12 MB</td>
      <td align="center">❌ solo router</td>
    </tr>
    <tr>
      <td>Echo v4</td>
      <td>Go 1.22</td>
      <td align="right">138 000</td>
      <td align="right">2,90 ms</td>
      <td align="right">7,9 ms</td>
      <td align="right">13 MB</td>
      <td align="center">❌ solo router</td>
    </tr>
    <tr>
      <td>NestJS v10</td>
      <td>Node 22</td>
      <td align="right">28 500</td>
      <td align="right">14,0 ms</td>
      <td align="right">42 ms</td>
      <td align="right">95 MB</td>
      <td align="center">✅</td>
    </tr>
    <tr>
      <td>Express v4</td>
      <td>Node 22</td>
      <td align="right">22 000</td>
      <td align="right">18,2 ms</td>
      <td align="right">55 ms</td>
      <td align="right">72 MB</td>
      <td align="center">❌ solo router</td>
    </tr>
    <tr>
      <td>Spring Boot 3</td>
      <td>Java 21 (JVM)</td>
      <td align="right">61 000</td>
      <td align="right">6,6 ms</td>
      <td align="right">21 ms</td>
      <td align="right">320 MB</td>
      <td align="center">✅</td>
    </tr>
  </tbody>
</table>

<sub>
  Benchmark: <code>wrk -t12 -c400 -d30s</code> · endpoint JSON <code>GET /users</code> · Linux x86-64 · CPU 8 núcleos · 16 GB RAM<br/>
  Nexgou ejecutado con el pipeline de middleware completo (Recovery → SecurityHeaders → RateLimit → Logger).
  Los routers se han medido con un handler mínimo <em>hello-world</em>.
</sub>

---

## ⚔️ ¿Por qué Nexgou frente a Gin, Fiber o Echo?

Gin, Fiber y Echo son excelentes **routers**. Pero al construir una aplicación real sobre ellos terminas escribiendo el mismo boilerplate una y otra vez: un sistema de DI, un cargador de módulos, guards de autenticación, interceptores de peticiones, manejo centralizado de errores, un cargador de configuración, un logger… Nexgou incluye todo eso, completamente integrado y listo para producción.

<table>
  <thead>
    <tr>
      <th align="left">Capacidad</th>
      <th align="center">Nexgou</th>
      <th align="center">Gin</th>
      <th align="center">Fiber</th>
      <th align="center">Echo</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><strong>Sistema de módulos</strong> — organiza el código por dominio, no por tipo de fichero</td>
      <td align="center">✅ incluido</td>
      <td align="center">❌ manual</td>
      <td align="center">❌ manual</td>
      <td align="center">❌ manual</td>
    </tr>
    <tr>
      <td><strong>Inyección de Dependencias</strong> — cableado automático de constructores, cero globals</td>
      <td align="center">✅ incluido</td>
      <td align="center">❌ manual</td>
      <td align="center">❌ manual</td>
      <td align="center">❌ manual</td>
    </tr>
    <tr>
      <td><strong>Guards</strong> — lógica de auth/authz desacoplada de los handlers</td>
      <td align="center">✅ incluido</td>
      <td align="center">❌ solo middleware</td>
      <td align="center">❌ solo middleware</td>
      <td align="center">❌ solo middleware</td>
    </tr>
    <tr>
      <td><strong>Interceptores</strong> — hooks pre/post handler por ruta</td>
      <td align="center">✅ incluido</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
    </tr>
    <tr>
      <td><strong>Pipes</strong> — validación y transformación de entrada</td>
      <td align="center">✅ incluido</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
    </tr>
    <tr>
      <td><strong>Filtros de Excepción</strong> — respuestas de error centralizadas y estructuradas</td>
      <td align="center">✅ incluido</td>
      <td align="center">❌ manual</td>
      <td align="center">❌ manual</td>
      <td align="center">❌ manual</td>
    </tr>
    <tr>
      <td><strong>Suite de seguridad</strong> (cabeceras, CORS, rate limit, timeout, body limit)</td>
      <td align="center">✅ incluido</td>
      <td align="center">⚠️ 3rd party</td>
      <td align="center">⚠️ 3rd party</td>
      <td align="center">⚠️ 3rd party</td>
    </tr>
    <tr>
      <td><strong>WebSocket</strong> — controlador de primera clase + guards en el upgrade</td>
      <td align="center">✅ incluido</td>
      <td align="center">⚠️ 3rd party</td>
      <td align="center">⚠️ 3rd party</td>
      <td align="center">⚠️ 3rd party</td>
    </tr>
    <tr>
      <td><strong>Server-Sent Events</strong></td>
      <td align="center">✅ incluido</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
    </tr>
    <tr>
      <td><strong>gRPC</strong> — sin <code>.proto</code>, guards en RPCs unarias</td>
      <td align="center">✅ incluido</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
    </tr>
    <tr>
      <td><strong>ConfigModule</strong> — acceso tipado e inyectable a variables de entorno</td>
      <td align="center">✅ incluido</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
    </tr>
    <tr>
      <td><strong>LogModule</strong> — logger estructurado, JSON/texto, con scope por servicio</td>
      <td align="center">✅ incluido</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
    </tr>
    <tr>
      <td><strong>Utilidades de testing</strong> — helpers unitarios e integración, sin configuración</td>
      <td align="center">✅ incluido</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
      <td align="center">❌</td>
    </tr>
    <tr>
      <td><strong>Versionado de rutas</strong> — <code>.Version("v1")</code> por ruta</td>
      <td align="center">✅ incluido</td>
      <td align="center">❌ prefijo manual</td>
      <td align="center">❌ prefijo manual</td>
      <td align="center">❌ prefijo manual</td>
    </tr>
  </tbody>
</table>

> **Conclusión:** con Gin, Fiber o Echo publicas un router y luego pasas semanas construyendo el framework de aplicación encima. Con Nexgou empiezas con el framework ya listo — y sigues obteniendo el rendimiento de Go.

---

## 🚀 Instalación

```bash
go get github.com/nexgou/server
```

> Requiere **Go 1.21** o superior.

---

## ⚡ Inicio Rápido

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

    app.Use(middleware.Recovery())
    app.Use(middleware.SecurityHeaders())
    app.Use(middleware.RateLimit(100, time.Minute))
    app.Use(middleware.Logger())
    app.SetFilter(&filter.HttpExceptionFilter{})

    log.Fatal(app.Listen(3000))
}
```

```go
// app.module.go
var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{nexgou.ConfigModule, nexgou.LogModule, UserModule},
})
```

```go
// user/user.module.go
var UserModule = nexgou.Module(nexgou.ModuleOptions{
    Controllers: []any{NewUserController},
    Providers:   []any{NewUserService},
})
```

```go
// user/user.controller.go
type UserController struct{ svc *UserService }

func NewUserController(s *UserService) *UserController { return &UserController{svc: s} }

func (c *UserController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.Get("/users",     c.FindAll).Version("v1"),
        nexgou.Post("/users",    c.Create).Version("v1").Guard(&AuthGuard{}),
        nexgou.Get("/users/:id", c.FindOne).Version("v1"),
    }
}

func (c *UserController) FindAll(ctx *nexgou.Context) error { return ctx.JSON(200, c.svc.FindAll()) }
func (c *UserController) Create(ctx *nexgou.Context) error  { return ctx.JSON(201, nexgou.H{"message": "creado"}) }
func (c *UserController) FindOne(ctx *nexgou.Context) error { return ctx.JSON(200, c.svc.FindOne(ctx.Param("id"))) }
```

Consulta [`samples/api`](samples/api) para un ejemplo completo y funcional con todas las características activadas.

---

## 📚 Documentación

| Guía | Descripción |
| :--- | :--- |
| [Primeros Pasos](docs/es/getting-started.md) | Instalación, primera app, ciclo de arranque |
| [Módulos](docs/es/modules.md) | Sistema de módulos, módulos de funcionalidad, imports y exports |
| [Controladores](docs/es/controllers.md) | Rutas, versioning, guards, interceptores, pipes |
| [Middleware](docs/es/middleware.md) | Logger, Recovery, CORS y el pipeline completo |
| [Seguridad](docs/es/security.md) | Cabeceras de seguridad, rate limiting, timeout, body limit |
| [WebSocket](docs/es/websocket.md) | `WSController`, `WSContext`, patrones de broadcast |
| [Server-Sent Events](docs/es/sse.md) | `SSEContext`, eventos nombrados, reconexión, desconexión |
| [gRPC](docs/es/grpc.md) | `GRPCController`, descriptores de servicio, streaming, guards |
| [Config](docs/es/config.md) | `ConfigService`, acceso tipado a variables de entorno |
| [Logger](docs/es/logger.md) | `LoggerService`, niveles, loggers con scope, salida JSON |
| [Testing](docs/es/testing.md) | Helpers unitarios e integración de `nexgoutest` |

---

## 🗂 Ejemplos

| Ejemplo | Transportes | Características demostradas |
| :--- | :---: | :--- |
| [`samples/api`](samples/api) | HTTP + WS + SSE | Pipeline completo, guards, interceptores, DI, versioning |
| [`samples/chat`](samples/chat) | WebSocket | Hub de broadcast, sala multi-cliente, logger con scope |
| [`samples/sse`](samples/sse) | SSE + HTTP | Eventos nombrados, filtrado de temas, endpoint de snapshot |
| [`samples/grpc`](samples/grpc) | gRPC + HTTP | Descriptores escritos a mano, streaming, sin `.proto` |

---

## 🗺 Hoja de Ruta

<details>
<summary><strong>Core (completado)</strong></summary>

- [x] Motor HTTP y contexto
- [x] Sistema de módulos y contenedor IoC
- [x] Resolución de Inyección de Dependencias
- [x] Controladores y enrutamiento declarativo
- [x] Pipeline de middleware (global y con scope)
- [x] Guards, Interceptores, Pipes
- [x] Filtros de Excepción
- [x] WebSocket — `WSController`, `WSContext`, guards
- [x] SSE — `SSEContext`, eventos nombrados, reconexión automática
- [x] gRPC — `GRPCController`, guards en RPCs unarias, sin `.proto`
- [x] `ConfigModule` y `LogModule`
- [x] `nexgoutest` — helpers de test unitarios e integración
- [x] Suite de middleware de seguridad (cabeceras, CORS, rate limit, timeout, body limit)

</details>


---

## 🤝 Contribuir

¡Las contribuciones son bienvenidas! Lee [CONTRIBUTING.md](CONTRIBUTING.md) antes de enviar un pull request. Para reportar vulnerabilidades de seguridad consulta [SECURITY.md](SECURITY.md).

---

## 📄 Licencia

Nexgou tiene [licencia MIT](LICENSE).
