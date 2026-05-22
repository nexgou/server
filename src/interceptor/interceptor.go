package interceptor

import "github.com/nexgou/server/src/common"

// Execute runs a chain of interceptors around the given handler.
// Interceptors are applied in order: first registered = outermost wrapper.
func Execute(ctx *common.Context, handler common.HandlerFunc, interceptors ...common.Interceptor) error {
	if len(interceptors) == 0 {
		return handler(ctx)
	}
	first := interceptors[0]
	rest := interceptors[1:]
	return first.Intercept(ctx, func(c *common.Context) error {
		return Execute(c, handler, rest...)
	})
}
