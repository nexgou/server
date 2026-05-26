package fasthttpadapter

import (
	"net/http"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// NewHandler adapts a net/http handler to a fasthttp request handler.
func NewHandler(handler http.Handler) fasthttp.RequestHandler {
	return fasthttpadaptor.NewFastHTTPHandler(handler)
}

// ListenAndServe starts a fasthttp server for the given net/http handler.
func ListenAndServe(address string, handler http.Handler) error {
	return fasthttp.ListenAndServe(address, NewHandler(handler))
}
