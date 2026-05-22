package chat

import (
	"sync"

	nexgou   "github.com/nexgou/server"
	nexgouws "github.com/nexgou/server/src/websocket"
	"github.com/nexgou/server/src/logger"
)

// room is a simple in-memory broadcast hub.
// Every connected client receives every message sent by any other client.
type room struct {
	mu      sync.RWMutex
	clients map[*nexgou.WSContext]struct{}
}

func newRoom() *room {
	return &room{clients: make(map[*nexgou.WSContext]struct{})}
}

func (r *room) join(ctx *nexgou.WSContext) {
	r.mu.Lock()
	r.clients[ctx] = struct{}{}
	r.mu.Unlock()
}

func (r *room) leave(ctx *nexgou.WSContext) {
	r.mu.Lock()
	delete(r.clients, ctx)
	r.mu.Unlock()
}

func (r *room) broadcast(sender *nexgou.WSContext, msg string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for c := range r.clients {
		if c != sender {
			_ = c.Send(msg)
		}
	}
}

// ── Controller ────────────────────────────────────────────────────────────────

// ChatController manages a single broadcast chat room over WebSocket.
// Every client that connects to /chat/room receives all messages sent by others.
type ChatController struct {
	log  *logger.ScopedLogger
	room *room
}

// NewChatController creates a ChatController (used by the IoC container).
func NewChatController(log *logger.LoggerService) *ChatController {
	return &ChatController{
		log:  log.WithContext("ChatController"),
		room: newRoom(),
	}
}

// Register returns no HTTP routes — this controller is WebSocket-only.
func (c *ChatController) Register() []nexgou.Route { return nil }

// RegisterWS registers the WebSocket routes.
func (c *ChatController) RegisterWS() []nexgouws.WSRoute {
	return []nexgouws.WSRoute{
		// Public echo endpoint — one client, messages echoed back.
		nexgouws.NewRoute("/chat/echo", c.HandleEcho),
		// Broadcast room — all connected clients receive every message.
		nexgouws.NewRoute("/chat/room", c.HandleRoom),
	}
}

// HandleEcho echoes every received message back to the same client.
func (c *ChatController) HandleEcho(ctx *nexgou.WSContext) error {
	c.log.Info("echo client connected", "remote", ctx.RemoteAddr())
	defer c.log.Info("echo client disconnected", "remote", ctx.RemoteAddr())

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

// HandleRoom joins the client to the broadcast room.
// Every message the client sends is forwarded to all other connected clients.
func (c *ChatController) HandleRoom(ctx *nexgou.WSContext) error {
	c.log.Info("room client joined", "remote", ctx.RemoteAddr())
	c.room.join(ctx)
	defer func() {
		c.room.leave(ctx)
		c.log.Info("room client left", "remote", ctx.RemoteAddr())
	}()

	for {
		msg, err := ctx.Receive()
		if err != nil {
			return nil // client disconnected
		}
		c.room.broadcast(ctx, msg)
	}
}
