package common

import (
	"encoding/json"
	"net/http"
)

// Context holds the state of a single HTTP request/response cycle.
type Context struct {
	Request *http.Request
	Writer  http.ResponseWriter
	params  map[string]string
}

// NewContext creates a new Context from an http.ResponseWriter, *http.Request, and route params.
func NewContext(w http.ResponseWriter, r *http.Request, params map[string]string) *Context {
	return &Context{Request: r, Writer: w, params: params}
}

// JSON writes a JSON-encoded response with the given HTTP status code.
func (c *Context) JSON(status int, data any) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	return json.NewEncoder(c.Writer).Encode(data)
}

// Param returns a URL route parameter by name (e.g. :id → "id").
func (c *Context) Param(key string) string {
	return c.params[key]
}

// Params returns a copy of all URL route parameters.
// Useful when passing params to sub-contexts (e.g. SSE, WebSocket adapters).
func (c *Context) Params() map[string]string {
	out := make(map[string]string, len(c.params))
	for k, v := range c.params {
		out[k] = v
	}
	return out
}

// Header returns the value of a request header by name.
func (c *Context) Header(key string) string {
	return c.Request.Header.Get(key)
}

// Method returns the HTTP method of the current request (GET, POST, etc.).
func (c *Context) Method() string {
	return c.Request.Method
}

// Path returns the URL path of the current request.
func (c *Context) Path() string {
	return c.Request.URL.Path
}

// Body decodes the JSON request body into the given target struct.
func (c *Context) Body(target any) error {
	defer c.Request.Body.Close()
	return json.NewDecoder(c.Request.Body).Decode(target)
}
