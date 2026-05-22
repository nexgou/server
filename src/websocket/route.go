package nexgouws

import (
	"net/http"
	"strings"

	"golang.org/x/net/websocket"

	"github.com/nexgou/server/src/common"
)

// Handler is the function signature for WebSocket connection handlers.
// The handler is called once per accepted connection and runs until it returns.
// Return nil on clean close, or an error to signal an abnormal disconnect.
//
//	func (c *ChatController) HandleChat(ctx *nexgouws.WSContext) error {
//	    for {
//	        msg, err := ctx.Receive()
//	        if err != nil { return err }
//	        return ctx.Send("echo: " + msg)
//	    }
//	}
type Handler func(*WSContext) error

// NewRoute creates a WebSocket route binding for the given path and handler.
// Use .Guard(...) and .Version(...) to configure access control and versioning.
//
//	nexgouws.NewRoute("/chat", c.HandleChat).Guard(&AuthGuard{}).Version("v1")
func NewRoute(path string, handler Handler) WSRoute {
	return WSRoute{Path: path, Handler: handler}
}

// WSRoute defines a WebSocket route binding with optional Guards and a version prefix.
// Construct one using the nexgou.WS helper.
type WSRoute struct {
	Path    string
	Handler Handler
	Guards  []common.Guard
	ver     string
}

// Guard attaches one or more guards to the WebSocket route.
// Guards run during the HTTP upgrade handshake — before the connection opens.
//
//	nexgou.WS("/chat", c.HandleChat).Guard(&AuthGuard{})
func (r WSRoute) Guard(guards ...common.Guard) WSRoute {
	r.Guards = append(r.Guards, guards...)
	return r
}

// Version sets a version prefix for the route (e.g. "v1" → /v1/chat).
func (r WSRoute) Version(v string) WSRoute {
	r.ver = v
	return r
}

// Ver returns the version label.
func (r WSRoute) Ver() string { return r.ver }

// FullPath returns the effective URL path including the version prefix.
func (r WSRoute) FullPath() string {
	if r.ver != "" {
		return "/" + r.ver + r.Path
	}
	return r.Path
}

// Upgrade performs the WebSocket upgrade and runs the handler.
// Guards are evaluated by the router BEFORE this method is called,
// using the original http.ResponseWriter and *http.Request.
// On a successful upgrade the handler runs for the lifetime of the connection.
//
// Origin check is disabled so tools like Postman and curl can connect without
// a browser Origin header. Guards are the intended access-control mechanism.
func (r WSRoute) Upgrade(w http.ResponseWriter, req *http.Request, params map[string]string) {
	srv := &websocket.Server{
		// Disable the default same-origin check so non-browser clients
		// (Postman, wscat, integration tests) can connect freely.
		// Use Guards on the route for access control instead.
		Handshake: func(cfg *websocket.Config, req *http.Request) error { return nil },
		Handler: func(conn *websocket.Conn) {
			ctx := newWSContext(conn, req, params)
			_ = r.Handler(ctx)
		},
	}
	srv.ServeHTTP(w, req)
}

// ── Internal router helpers ───────────────────────────────────────────────────

// WSEntry is the internal representation of a registered WebSocket route,
// with pre-split path segments for efficient matching.
type WSEntry struct {
	Segments []string
	Route    WSRoute
}

// NewEntry creates a WSEntry from a WSRoute.
func NewEntry(r WSRoute) WSEntry {
	return WSEntry{
		Segments: SplitPath(r.FullPath()),
		Route:    r,
	}
}

// Match reports whether pathSegments match this entry,
// populating params with any captured URL parameters.
func (e *WSEntry) Match(pathSegments []string, params map[string]string) bool {
	return matchSegments(e.Segments, pathSegments, params)
}

// SplitPath splits a URL path into clean segments, stripping slashes.
func SplitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return []string{}
	}
	return strings.Split(trimmed, "/")
}

// matchSegments checks whether pattern matches path, capturing :param values.
func matchSegments(pattern, path []string, params map[string]string) bool {
	if len(pattern) != len(path) {
		return false
	}
	for i, seg := range pattern {
		if strings.HasPrefix(seg, ":") {
			params[seg[1:]] = path[i]
		} else if seg != path[i] {
			return false
		}
	}
	return true
}

// ── WSController interface ─────────────────────────────────────────────────────

// WSController is implemented by controllers that expose WebSocket routes.
// It is a companion to the standard Controller interface — a controller can
// implement both to serve HTTP and WebSocket routes from the same struct.
//
// Usage:
//
//	func (c *ChatController) Register() []nexgou.Route {
//	    return []nexgou.Route{ /* HTTP routes if any */ }
//	}
//
//	func (c *ChatController) RegisterWS() []nexgouws.WSRoute {
//	    return []nexgouws.WSRoute{
//	        nexgouws.NewRoute("/chat", c.HandleChat).Guard(&AuthGuard{}),
//	    }
//	}
type WSController interface {
	RegisterWS() []WSRoute
}
