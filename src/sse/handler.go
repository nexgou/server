package nexgousse

import (
	"github.com/nexgou/server/src/common"
)

// HandlerFunc is the function signature for SSE stream handlers.
// The handler is called once per accepted connection and runs until it returns.
//
// Return nil on a normal client disconnect (ctx.Done() closed).
// Return an error only for unexpected failures — it is forwarded to the
// router's exception filter just like any other handler error.
//
//	func (c *MetricsController) Stream(ctx *nexgousse.SSEContext) error {
//	    ctx.SetRetry(3000)
//	    for {
//	        select {
//	        case <-ctx.Done():
//	            return nil
//	        case m := <-c.metrics:
//	            ctx.SendNamedJSON("metrics", m)
//	        }
//	    }
//	}
type HandlerFunc func(*SSEContext) error

// ToHTTPHandler wraps an SSE HandlerFunc into a standard common.HandlerFunc
// so it can be registered on any HTTP route using the normal nexgou helpers.
//
// The wrapper initializes the SSEContext (writes SSE headers, verifies Flusher
// support) and delegates to the SSE handler. If the ResponseWriter does not
// support flushing it responds with HTTP 500 immediately.
//
// Usage:
//
//	nexgou.Get("/events", nexgousse.ToHTTPHandler(c.Stream))
//	nexgou.Get("/events", nexgousse.ToHTTPHandler(c.Stream)).Guard(&AuthGuard{})
func ToHTTPHandler(fn HandlerFunc) common.HandlerFunc {
	return func(ctx *common.Context) error {
		sseCtx, err := newSSEContext(ctx.Writer, ctx.Request, ctx.Params())
		if err != nil {
			return common.NewInternalServerErrorException(err.Error())
		}
		return fn(sseCtx)
	}
}
