package nexgouws

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/websocket"
)

// WSContext holds the state of an active WebSocket connection.
// It wraps the underlying websocket.Conn with a clean, idiomatic API
// consistent with Nexgou's HTTP Context.
type WSContext struct {
	// Request is the original HTTP upgrade request.
	// Use it to access headers, URL params, etc.
	Request *http.Request

	conn   *websocket.Conn
	params map[string]string
}

// newWSContext creates a WSContext from a websocket.Conn and the upgrade request.
func newWSContext(conn *websocket.Conn, r *http.Request, params map[string]string) *WSContext {
	return &WSContext{Request: r, conn: conn, params: params}
}

// Send writes a UTF-8 text message to the client.
func (c *WSContext) Send(msg string) error {
	return websocket.Message.Send(c.conn, msg)
}

// SendBytes writes a binary message to the client.
func (c *WSContext) SendBytes(data []byte) error {
	return websocket.Message.Send(c.conn, data)
}

// SendJSON encodes v as JSON and sends it as a text message.
func (c *WSContext) SendJSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.Send(string(b))
}

// Receive reads the next text message from the client.
// Returns io.EOF when the connection is closed by the client.
func (c *WSContext) Receive() (string, error) {
	var msg string
	err := websocket.Message.Receive(c.conn, &msg)
	return msg, err
}

// ReceiveBytes reads the next binary message from the client.
func (c *WSContext) ReceiveBytes() ([]byte, error) {
	var data []byte
	err := websocket.Message.Receive(c.conn, &data)
	return data, err
}

// ReceiveJSON reads the next text message and decodes it as JSON into target.
func (c *WSContext) ReceiveJSON(target any) error {
	msg, err := c.Receive()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(msg), target)
}

// Param returns a URL route parameter by name (e.g. :id → "id").
func (c *WSContext) Param(key string) string {
	return c.params[key]
}

// Header returns the value of a request header by name.
func (c *WSContext) Header(key string) string {
	return c.Request.Header.Get(key)
}

// RemoteAddr returns the client's network address.
func (c *WSContext) RemoteAddr() string {
	return c.conn.Request().RemoteAddr
}

// Close closes the WebSocket connection.
func (c *WSContext) Close() error {
	return c.conn.Close()
}
