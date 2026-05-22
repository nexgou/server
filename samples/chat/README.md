# Sample: WebSocket Chat

A complete example of a Nexgou WebSocket application demonstrating real-time bidirectional communication. The sample implements two endpoints: a simple echo server and a multi-client broadcast room.

## What this sample covers

| Feature | Where |
|---------|-------|
| `WSController` interface | `chat/chat.controller.go` |
| `RegisterWS()` — WS route registration | `chat/chat.controller.go` |
| `WSContext` — send, receive messages | `chat/chat.controller.go` |
| Echo handler (one client ↔ server) | `chat/chat.controller.go` — `HandleEcho` |
| Broadcast room (N clients) | `chat/chat.controller.go` — `HandleRoom` |
| Goroutine-safe client registry | `chat/chat.controller.go` — `room` struct |
| `LoggerService` injection | `chat/chat.controller.go` |
| Guards on WS routes | Ready to add — see docs |
| CORS middleware | `main.go` |

## Project structure

```
samples/chat/
├── main.go               # Bootstrap — middleware pipeline + server start (port 3001)
├── app.module.go         # Root module — imports LogModule, ChatModule
├── README.md
└── chat/
    ├── chat.controller.go  # WSController: /chat/echo + /chat/room
    └── chat.module.go      # ChatModule definition
```

## WebSocket routes

| Path | Description |
|------|-------------|
| `ws://localhost:3001/chat/echo` | Echo server — every message is returned to the sender with `echo: ` prefix |
| `ws://localhost:3001/chat/room` | Broadcast room — every message is forwarded to all other connected clients |

## Run

```bash
cd samples/chat
go run .
# Server starts on ws://localhost:3001
```

## Try it with Postman

### Echo endpoint

1. **New → WebSocket**
2. URL: `ws://localhost:3001/chat/echo`
3. **Connect**
4. Send: `Hello Nexgou!`
5. Receive: `echo: Hello Nexgou!`

### Broadcast room

Open two Postman WebSocket tabs simultaneously:

1. Both connect to `ws://localhost:3001/chat/room`
2. Send a message from tab 1
3. Tab 2 receives it (and vice versa)
4. The sender does **not** receive their own message back

## Try it with wscat

```bash
# Install once
npm install -g wscat

# Echo
wscat -c ws://localhost:3001/chat/echo
> Hello!
< echo: Hello!

# Room — open two terminals
# Terminal 1:
wscat -c ws://localhost:3001/chat/room

# Terminal 2:
wscat -c ws://localhost:3001/chat/room
> Hi from terminal 2!
# Terminal 1 receives: Hi from terminal 2!
```

## Try it from a browser

```js
// Echo
const echo = new WebSocket('ws://localhost:3001/chat/echo')
echo.onopen    = () => echo.send('Hello from browser!')
echo.onmessage = (e) => console.log(e.data) // echo: Hello from browser!

// Room
const room = new WebSocket('ws://localhost:3001/chat/room')
room.onopen    = () => room.send('Hi everyone!')
room.onmessage = (e) => console.log('Received:', e.data)
```

## How the broadcast room works

```
Client A ──send("hi")──▶ HandleRoom ──broadcast──▶ Client B
                                              └──▶ Client C
                          (A does NOT receive its own message)
```

The `room` struct maintains a goroutine-safe map of active connections using `sync.RWMutex`. On each message, it iterates all registered clients and sends to everyone except the sender:

```go
func (r *room) broadcast(sender *nexgou.WSContext, msg string) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    for c := range r.clients {
        if c != sender {
            _ = c.Send(msg)
        }
    }
}
```

## Adding authentication

To protect routes with a guard, update `RegisterWS()` in `chat.controller.go`:

```go
func (c *ChatController) RegisterWS() []nexgouws.WSRoute {
    return []nexgouws.WSRoute{
        nexgouws.NewRoute("/chat/echo", c.HandleEcho),
        nexgouws.NewRoute("/chat/room", c.HandleRoom).Guard(&AuthGuard{}),
    }
}
```

Guards run during the HTTP upgrade handshake — before the WebSocket connection is opened. See [`docs/websocket.md`](../../docs/websocket.md) for the full reference.
