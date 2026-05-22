package nexgougrpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// GRPCContext holds the state of a single gRPC RPC call.
// It is passed to Guards and is available for inspection during the
// UnaryServerInterceptor phase (before the handler executes).
//
// For streaming RPCs the raw stream carries its own context — use
// stream.Context() directly inside streaming handlers.
type GRPCContext struct {
	// ctx is the Go context for this RPC (carries deadlines, cancellation, etc.)
	ctx context.Context

	// fullMethod is the full gRPC method path, e.g. "/greeter.Greeter/SayHello".
	fullMethod string
}

// Context returns the underlying Go context.Context for this RPC.
// It carries the deadline, cancellation signal, and any values attached by
// middleware or the client.
func (c *GRPCContext) Context() context.Context {
	return c.ctx
}

// Method returns the full gRPC method name, including the package and service.
//
//	"/greeter.Greeter/SayHello"
func (c *GRPCContext) Method() string {
	return c.fullMethod
}

// Metadata returns the value of the first entry for the given incoming
// metadata key (case-insensitive).
// Returns an empty string if the key is not present.
//
// Keys follow the gRPC metadata naming convention (lowercase, hyphens allowed).
//
//	token := ctx.Metadata("authorization")
func (c *GRPCContext) Metadata(key string) string {
	md, ok := metadata.FromIncomingContext(c.ctx)
	if !ok {
		return ""
	}
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// MetadataAll returns all values for the given incoming metadata key.
// Returns nil if the key is not present.
func (c *GRPCContext) MetadataAll(key string) []string {
	md, ok := metadata.FromIncomingContext(c.ctx)
	if !ok {
		return nil
	}
	values := md.Get(key)
	if len(values) == 0 {
		return nil
	}
	return values
}
