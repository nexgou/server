# WebSocket

Nexgou has first-class WebSocket support built on [`golang.org/x/net/websocket`](https://pkg.go.dev/golang.org/x/net/websocket). WebSocket routes integrate seamlessly with the module system, the IoC container, and Guards — no separate server or extra setup required.

---

## Table of Contents

- [How it works](#how-it-works)
- [Implementing a WebSocket controller](#implementing-a-websocket-controller)
  - [Handler signature](#handler-signature)
  - [Registering routes](#registering-routes)
  - [WSContext API](#wscontext-api)
- [Combining HTTP and WebSocket routes](#combining-http-and-websocket-routes)
- [URL parameters](#url-parameters)
- [Guards on WebSocket routes](#guards-on-websocket-routes)
- [Versioning](#versioning)
- [Registering the module](#registering-the-module)
- [Full example — chat echo server](#full-example--chat-echo-server)
- [Testing with Postman](#testing-with-postman)
- [Testing with wscat](#testing-with-wscat)

---

## How it works

The Nexgou router inspects every incoming request for the `Upgrade: websocket` header **before** matching HTTP routes. When found:

1. The request path is matched against registered WS routes.
2. **Guards** run against the original HTTP upgrade request — a denied guard responds with HTTP `403` before the connection is opened.
3. If all guards pass, `Upgrade()` performs the WebSocket handshake and calls the handler for the lifetime of the connection.

Origin checking is disabled by default so non-browser clients (Postman, wscat, integration tests) can connect. Use Guards for access control.

---

## Implementing a WebSocket controller

A WebSocket controller implements the `nexgouws.WSController` interface:

```go
type WSController interface {
    RegisterWS() []nexgouws.WSRoute
}
```

Import the `nexgouws` sub-package alongside the main `nexgou` package:

```go
import (
    nexgou   "github.com/nexgou/server"
    nexgouws "github.com/nexgou/server/src/websocket"
)
```

### Handler signature

Every WebSocket handler receives a `*nexgou.WSContext` and returns an error:

```go
func (c *MyController) HandleConn(ctx *nexgou.WSContext) error {
    // read / write loop
    return nil // nil = clean close
}
```

Return `nil` on a normal client disconnect; return an error only for unexpected failures.

### Registering routes

```go
func (c *MyController) RegisterWS() []nexgouws.WSRoute {
    return []nexgouws.WSRoute{
        nexgouws.NewRoute("/my-path", c.HandleConn),
    }
}
```

Or use the top-level helper exposed in the `nexgou` package:

```go
func (c *MyController) RegisterWS() []nexgouws.WSRoute {
    return []nexgouws.WSRoute{
        nexgou.WS("/my-path", c.HandleConn),
    }
}
```

### WSContext API

`WSContext` wraps the WebSocket connection and exposes a clean, idiomatic API consistent with Nexgou's HTTP `Context`.

| Method | Description |
|--------|-------------|
| `Send(msg string) error` | Send a UTF-8 text message |
| `SendBytes(data []byte) error` | Send a binary message |
| `SendJSON(v any) error` | Marshal `v` as JSON and send it as text |
| `Receive() (string, error)` | Read the next text message |
| `ReceiveBytes() ([]byte, error)` | Read the next binary message |
| `ReceiveJSON(target any) error` | Read and unmarshal the next text message into `target` |
| `Param(key string) string` | Read a URL path parameter (e.g. `:room` → `"room"`) |
| `Header(key string) string` | Read a header from the original upgrade request |
| `RemoteAddr() string` | Client's network address |
| `Close() error` | Close the connection |
| `Request *http.Request` | The original HTTP upgrade request (read-only) |

---

## Combining HTTP and WebSocket routes

A controller can implement **both** `nexgou.Controller` (HTTP) and `nexgouws.WSController` (WS) at the same time:

```go
type RoomController struct{}

// HTTP routes
func (c *RoomController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.Get("/rooms", c.ListRooms),
    }
}

// WebSocket routes
func (c *RoomController) RegisterWS() []nexgouws.WSRoute {
    return []nexgouws.WSRoute{
        nexgou.WS("/rooms/:id/ws", c.JoinRoom),
    }
}
```

---

## URL parameters

Define parameters with a colon prefix in the path, then read them with `ctx.Param`:

```go
nexgou.WS("/rooms/:id/ws", c.JoinRoom)

func (c *RoomController) JoinRoom(ctx *nexgou.WSContext) error {
    roomID := ctx.Param("id")
    // ...
}
```

---

## Guards on WebSocket routes

Guards run during the HTTP upgrade handshake — **before** the WebSocket connection is established. A failed guard responds with HTTP `403` and the connection is never opened.

```go
nexgouws.NewRoute("/chat", c.HandleChat).Guard(&AuthGuard{})
```

A Guard implements the standard `nexgou.Guard` interface:

```go
type AuthGuard struct{}

func (g *AuthGuard) CanActivate(ctx *nexgou.Context) (bool, error) {
    token := ctx.Header("Authorization")
    return token != "", nil
}
```

Multiple guards are evaluated in order; the first failure short-circuits:

```go
nexgou.WS("/admin/ws", c.HandleAdmin).
    Guard(&AuthGuard{}, &AdminRoleGuard{})
```

---

## Versioning

Use `.Version("v1")` to prefix the route path:

```go
nexgou.WS("/chat", c.HandleChat).Version("v1")
// effective path: /v1/chat
```

---

## Registering the module

Register the controller factory in a module's `Controllers` list as usual. No special WS registration is needed — the framework detects `WSController` automatically during `walkModule`.

```go
var ChatModule = nexgou.Module(nexgou.ModuleOptions{
    Controllers: []any{NewChatController},
})
```

Import it in your root module:

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

## Full example — chat echo server

### `chat/chat.controller.go`

```go
package chat

import (
    nexgou   "github.com/nexgou/server"
    nexgouws "github.com/nexgou/server/src/websocket"
)

type ChatController struct{}

func NewChatController() *ChatController { return &ChatController{} }

// Register — no HTTP routes for this controller.
func (c *ChatController) Register() []nexgou.Route { return nil }

// RegisterWS — one WS route, protected by AuthGuard.
func (c *ChatController) RegisterWS() []nexgouws.WSRoute {
    return []nexgouws.WSRoute{
        nexgou.WS("/chat", c.HandleChat).Guard(&AuthGuard{}),
    }
}

// HandleChat echoes every message back with the prefix "echo: ".
func (c *ChatController) HandleChat(ctx *nexgou.WSContext) error {
    for {
        msg, err := ctx.Receive()
        if err != nil {
            return nil // client disconnected
        }
        if err := ctx.Send("echo: " + msg); err != nil {
            return err
        }
    }
}

// AuthGuard requires a non-empty Authorization header on the upgrade request.
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
// ... other middleware
app.Listen(3000)
```

Server output on startup:
```
WS      /chat   🔒 AuthGuard
```

---

## Testing with Postman

1. Open Postman → **New → WebSocket**
2. URL: `ws://localhost:3000/chat`
3. If the route has a Guard that checks `Authorization`, add it in the **Headers** tab before connecting:
   ```
   Authorization: Bearer my-token
   ```
4. Click **Connect**
5. In the **Message** panel, type any text and click **Send**
6. You will receive `echo: <your message>` back

---

## Testing with wscat

```bash
# Public route
npx wscat -c ws://localhost:3000/chat

# Route protected by Authorization header
npx wscat -c ws://localhost:3000/chat \
  -H "Authorization: Bearer my-token"
```

Once connected:
```
> Hello Nexgou!
< echo: Hello Nexgou!
```
