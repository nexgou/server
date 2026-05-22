# WebSocket

Nexgou tiene soporte de primera clase para WebSocket construido sobre [`golang.org/x/net/websocket`](https://pkg.go.dev/golang.org/x/net/websocket). Las rutas WebSocket se integran perfectamente con el sistema de módulos, el contenedor IoC y los Guards — sin servidor separado ni configuración adicional.

---

## Tabla de Contenidos

- [Cómo funciona](#cómo-funciona)
- [Implementar un controlador WebSocket](#implementar-un-controlador-websocket)
  - [Firma del handler](#firma-del-handler)
  - [Registrar rutas](#registrar-rutas)
  - [API de WSContext](#api-de-wscontext)
- [Combinar rutas HTTP y WebSocket](#combinar-rutas-http-y-websocket)
- [Parámetros de URL](#parámetros-de-url)
- [Guards en rutas WebSocket](#guards-en-rutas-websocket)
- [Versionado](#versionado)
- [Registrar el módulo](#registrar-el-módulo)
- [Ejemplo completo — servidor echo de chat](#ejemplo-completo--servidor-echo-de-chat)
- [Probar con Postman](#probar-con-postman)
- [Probar con wscat](#probar-con-wscat)

---

## Cómo funciona

El router de Nexgou inspecciona cada solicitud entrante por la cabecera `Upgrade: websocket` **antes** de hacer coincidir las rutas HTTP. Cuando se encuentra:

1. El path de la solicitud se compara con las rutas WS registradas.
2. Los **Guards** se ejecutan contra la solicitud HTTP de actualización original — un guard denegado responde con HTTP `403` antes de que se abra la conexión.
3. Si todos los guards pasan, `Upgrade()` realiza el handshake WebSocket y llama al handler durante el tiempo de vida de la conexión.

La verificación de origen está deshabilitada por defecto para que los clientes no-navegador (Postman, wscat, tests de integración) puedan conectarse. Usar Guards para control de acceso.

---

## Implementar un controlador WebSocket

Un controlador WebSocket implementa la interfaz `nexgouws.WSController`:

```go
type WSController interface {
    RegisterWS() []nexgouws.WSRoute
}
```

Importar el sub-paquete `nexgouws` junto con el paquete principal `nexgou`:

```go
import (
    nexgou   "github.com/nexgou/server"
    nexgouws "github.com/nexgou/server/src/websocket"
)
```

### Firma del handler

Cada handler WebSocket recibe un `*nexgou.WSContext` y devuelve un error:

```go
func (c *MyController) HandleConn(ctx *nexgou.WSContext) error {
    // bucle de lectura / escritura
    return nil // nil = cierre limpio
}
```

Devolver `nil` en una desconexión normal del cliente; devolver un error solo para fallos inesperados.

### Registrar rutas

```go
func (c *MyController) RegisterWS() []nexgouws.WSRoute {
    return []nexgouws.WSRoute{
        nexgouws.NewRoute("/my-path", c.HandleConn),
    }
}
```

O usar el helper de nivel superior expuesto en el paquete `nexgou`:

```go
func (c *MyController) RegisterWS() []nexgouws.WSRoute {
    return []nexgouws.WSRoute{
        nexgou.WS("/my-path", c.HandleConn),
    }
}
```

### API de WSContext

`WSContext` envuelve la conexión WebSocket y expone una API limpia e idiomática consistente con el `Context` HTTP de Nexgou.

| Método | Descripción |
|--------|-------------|
| `Send(msg string) error` | Enviar un mensaje de texto UTF-8 |
| `SendBytes(data []byte) error` | Enviar un mensaje binario |
| `SendJSON(v any) error` | Serializar `v` como JSON y enviarlo como texto |
| `Receive() (string, error)` | Leer el siguiente mensaje de texto |
| `ReceiveBytes() ([]byte, error)` | Leer el siguiente mensaje binario |
| `ReceiveJSON(target any) error` | Leer y deserializar el siguiente mensaje de texto en `target` |
| `Param(key string) string` | Leer un parámetro de ruta URL (p. ej. `:room` → `"room"`) |
| `Header(key string) string` | Leer una cabecera de la solicitud de actualización original |
| `RemoteAddr() string` | Dirección de red del cliente |
| `Close() error` | Cerrar la conexión |
| `Request *http.Request` | La solicitud HTTP de actualización original (solo lectura) |

---

## Combinar rutas HTTP y WebSocket

Un controlador puede implementar **ambas** interfaces `nexgou.Controller` (HTTP) y `nexgouws.WSController` (WS) al mismo tiempo:

```go
type RoomController struct{}

// Rutas HTTP
func (c *RoomController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.Get("/rooms", c.ListRooms),
    }
}

// Rutas WebSocket
func (c *RoomController) RegisterWS() []nexgouws.WSRoute {
    return []nexgouws.WSRoute{
        nexgou.WS("/rooms/:id/ws", c.JoinRoom),
    }
}
```

---

## Parámetros de URL

Define parámetros con un prefijo de dos puntos en el path, luego léelos con `ctx.Param`:

```go
nexgou.WS("/rooms/:id/ws", c.JoinRoom)

func (c *RoomController) JoinRoom(ctx *nexgou.WSContext) error {
    roomID := ctx.Param("id")
    // ...
}
```

---

## Guards en rutas WebSocket

Los guards se ejecutan durante el handshake de actualización HTTP — **antes** de que se establezca la conexión WebSocket. Un guard fallido responde con HTTP `403` y la conexión nunca se abre.

```go
nexgouws.NewRoute("/chat", c.HandleChat).Guard(&AuthGuard{})
```

Un Guard implementa la interfaz estándar `nexgou.Guard`:

```go
type AuthGuard struct{}

func (g *AuthGuard) CanActivate(ctx *nexgou.Context) (bool, error) {
    token := ctx.Header("Authorization")
    return token != "", nil
}
```

Los múltiples guards se evalúan en orden; el primer fallo hace un cortocircuito:

```go
nexgou.WS("/admin/ws", c.HandleAdmin).
    Guard(&AuthGuard{}, &AdminRoleGuard{})
```

---

## Versionado

Usar `.Version("v1")` para prefijar el path de la ruta:

```go
nexgou.WS("/chat", c.HandleChat).Version("v1")
// path efectivo: /v1/chat
```

---

## Registrar el módulo

Registrar la factory del controlador en la lista `Controllers` de un módulo como de costumbre. No se necesita registro WS especial — el framework detecta `WSController` automáticamente durante `walkModule`.

```go
var ChatModule = nexgou.Module(nexgou.ModuleOptions{
    Controllers: []any{NewChatController},
})
```

Importarlo en el módulo raíz:

```go
var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{
        nexgou.ConfigModule,
        nexgou.LogModule,
        ChatModule,
    },
})
```

---

## Ejemplo completo — servidor echo de chat

### `chat/chat.controller.go`

```go
package chat

import (
    nexgou   "github.com/nexgou/server"
    nexgouws "github.com/nexgou/server/src/websocket"
)

type ChatController struct{}

func NewChatController() *ChatController { return &ChatController{} }

// Register — sin rutas HTTP para este controlador.
func (c *ChatController) Register() []nexgou.Route { return nil }

// RegisterWS — una ruta WS, protegida por AuthGuard.
func (c *ChatController) RegisterWS() []nexgouws.WSRoute {
    return []nexgouws.WSRoute{
        nexgou.WS("/chat", c.HandleChat).Guard(&AuthGuard{}),
    }
}

// HandleChat devuelve cada mensaje con el prefijo "echo: ".
func (c *ChatController) HandleChat(ctx *nexgou.WSContext) error {
    for {
        msg, err := ctx.Receive()
        if err != nil {
            return nil // cliente desconectado
        }
        if err := ctx.Send("echo: " + msg); err != nil {
            return err
        }
    }
}

// AuthGuard requiere una cabecera Authorization no vacía en la solicitud de actualización.
type AuthGuard struct{}

func (g *AuthGuard) CanActivate(ctx *nexgou.Context) (bool, error) {
    return ctx.Header("Authorization") != "", nil
}
```

### `chat/chat.module.go`

```go
package chat

import nexgou "github.com/nexgou/server"

var Module = nexgou.Module(nexgou.ModuleOptions{
    Controllers: []any{NewChatController},
})
```

### `main.go`

```go
app := nexgou.CreateApp(AppModule)
app.Use(middleware.Recovery())
// ... otro middleware
app.Listen(3000)
```

Salida del servidor en el arranque:
```
WS      /chat   🔒 AuthGuard
```

---

## Probar con Postman

1. Abrir Postman → **New → WebSocket**
2. URL: `ws://localhost:3000/chat`
3. Si la ruta tiene un Guard que verifica `Authorization`, añadirlo en la pestaña **Headers** antes de conectar:
   ```
   Authorization: Bearer my-token
   ```
4. Hacer clic en **Connect**
5. En el panel **Message**, escribir cualquier texto y hacer clic en **Send**
6. Recibirás `echo: <tu mensaje>` de vuelta

---

## Probar con wscat

```bash
# Ruta pública
npx wscat -c ws://localhost:3000/chat

# Ruta protegida por cabecera Authorization
npx wscat -c ws://localhost:3000/chat \
  -H "Authorization: Bearer my-token"
```

Una vez conectado:
```
> Hola Nexgou!
< echo: Hola Nexgou!
```
