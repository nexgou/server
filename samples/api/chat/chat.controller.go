package chat

import (
	nexgou "github.com/nexgou/server"
	nexgouws "github.com/nexgou/server/src/websocket"
)

// ChatController handles WebSocket connections for the /chat endpoint.
// It implements both nexgou.Controller (HTTP) and nexgouws.WSController (WS).
type ChatController struct{}

// NewChatController creates a ChatController (used by the IoC container).
func NewChatController() *ChatController {
	return &ChatController{}
}

// Register returns an empty HTTP route list — this controller is WS-only.
func (c *ChatController) Register() []nexgou.Route {
	return nil
}

// RegisterWS returns the WebSocket routes for this controller.
func (c *ChatController) RegisterWS() []nexgouws.WSRoute {
	return []nexgouws.WSRoute{
		nexgouws.NewRoute("/chat", c.HandleChat),
	}
}

// HandleChat is an echo handler: it reads each message and writes it back
// with an "echo: " prefix until the client closes the connection.
func (c *ChatController) HandleChat(ctx *nexgou.WSContext) error {
	for {
		msg, err := ctx.Receive()
		if err != nil {
			// Client disconnected — normal close.
			return nil
		}
		if err := ctx.Send("echo: " + msg); err != nil {
			return err
		}
	}
}
