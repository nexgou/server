package nexgousse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// SSEContext holds the state of an active Server-Sent Events stream.
// It wraps the underlying http.ResponseWriter with a clean, idiomatic API
// consistent with Nexgou's HTTP Context.
//
// SSE headers (Content-Type, Cache-Control, Connection, X-Accel-Buffering)
// are written automatically when the context is created. The connection stays
// open until the handler returns or the client disconnects.
type SSEContext struct {
	// Request is the original HTTP request that initiated the SSE stream.
	Request *http.Request

	writer  http.ResponseWriter
	flusher http.Flusher
	params  map[string]string
}

// newSSEContext creates and initializes an SSEContext.
// It writes the required SSE response headers and returns an error if the
// ResponseWriter does not support flushing (required for streaming).
func newSSEContext(w http.ResponseWriter, r *http.Request, params map[string]string) (*SSEContext, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("nexgou/sse: ResponseWriter does not implement http.Flusher; streaming is not supported")
	}

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	// Disable proxy/nginx response buffering so events reach the client immediately.
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	return &SSEContext{
		Request: r,
		writer:  w,
		flusher: flusher,
		params:  params,
	}, nil
}

// ── Write helpers ─────────────────────────────────────────────────────────────

// Send writes a plain-text data event and flushes it to the client.
//
//	ctx.Send("hello world")
//	// →  data: hello world\n\n
func (c *SSEContext) Send(data string) error {
	return c.writeAndFlush(formatData(data))
}

// SendNamed writes a named event with a plain-text data field.
// The browser EventSource API fires a custom event listener for named events.
//
//	ctx.SendNamed("user.created", `{"id":42}`)
//	// →  event: user.created\n
//	//    data: {"id":42}\n\n
func (c *SSEContext) SendNamed(event, data string) error {
	return c.writeAndFlush("event: " + event + "\n" + formatData(data))
}

// SendJSON serializes v as JSON and sends it as an unnamed data event.
//
//	ctx.SendJSON(map[string]any{"cpu": 42})
//	// →  data: {"cpu":42}\n\n
func (c *SSEContext) SendJSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("nexgou/sse: SendJSON marshal: %w", err)
	}
	return c.Send(string(b))
}

// SendNamedJSON serializes v as JSON and sends it as a named event.
//
//	ctx.SendNamedJSON("metrics", payload)
//	// →  event: metrics\n
//	//    data: {...}\n\n
func (c *SSEContext) SendNamedJSON(event string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("nexgou/sse: SendNamedJSON marshal: %w", err)
	}
	return c.SendNamed(event, string(b))
}

// SendComment writes an SSE comment line.
// Comments are invisible to the EventSource API but keep the TCP connection
// alive through proxies and load balancers that close idle connections.
//
//	ctx.SendComment("keepalive")
//	// →  : keepalive\n\n
func (c *SSEContext) SendComment(comment string) error {
	return c.writeAndFlush(": " + comment + "\n\n")
}

// SetRetry instructs the browser to wait ms milliseconds before attempting
// to reconnect after a lost connection. Call it once at the start of the stream.
//
//	ctx.SetRetry(3000) // reconnect after 3 s
//	// →  retry: 3000\n\n
func (c *SSEContext) SetRetry(ms int) error {
	return c.writeAndFlush(fmt.Sprintf("retry: %d\n\n", ms))
}

// SetID sets the last event ID. The browser sends this value back in the
// Last-Event-ID header on reconnect, allowing the server to resume the stream.
//
//	ctx.SetID("42")
//	// →  id: 42\n\n
func (c *SSEContext) SetID(id string) error {
	return c.writeAndFlush("id: " + id + "\n\n")
}

// ── Request helpers ───────────────────────────────────────────────────────────

// Param returns a URL route parameter by name (e.g. :id → "id").
func (c *SSEContext) Param(key string) string {
	return c.params[key]
}

// Header returns the value of a request header by name.
func (c *SSEContext) Header(key string) string {
	return c.Request.Header.Get(key)
}

// LastEventID returns the value of the Last-Event-ID header sent by the
// browser on reconnect. Returns an empty string on the first connection.
func (c *SSEContext) LastEventID() string {
	return c.Request.Header.Get("Last-Event-ID")
}

// RemoteAddr returns the client's network address.
func (c *SSEContext) RemoteAddr() string {
	return c.Request.RemoteAddr
}

// Done returns a channel that is closed when the client disconnects.
// Use it to stop the streaming loop cleanly:
//
//	for {
//	    select {
//	    case <-ctx.Done():
//	        return nil
//	    case event := <-eventCh:
//	        ctx.SendNamedJSON("update", event)
//	    }
//	}
func (c *SSEContext) Done() <-chan struct{} {
	return c.Request.Context().Done()
}

// ── Internal ──────────────────────────────────────────────────────────────────

// writeAndFlush writes raw SSE bytes to the ResponseWriter and immediately
// flushes them to the client so they are not held in a buffer.
func (c *SSEContext) writeAndFlush(raw string) error {
	if _, err := fmt.Fprint(c.writer, raw); err != nil {
		return err
	}
	c.flusher.Flush()
	return nil
}

// formatData formats a (possibly multi-line) data value according to the SSE spec.
// Each line of the value is prefixed with "data: " and the block ends with "\n\n".
//
//	"hello\nworld" → "data: hello\ndata: world\n\n"
func formatData(data string) string {
	lines := strings.Split(data, "\n")
	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString("data: ")
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
	sb.WriteByte('\n')
	return sb.String()
}
