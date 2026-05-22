# gRPC

> **[← Back to README](../../README.md)**

---

## Table of Contents

- [Overview](#overview)
- [How It Works (No .proto Required)](#how-it-works-no-proto-required)
- [Message Types](#message-types)
- [Service Interface & Descriptor](#service-interface--descriptor)
- [Controller Implementation](#controller-implementation)
- [Server-Streaming RPCs](#server-streaming-rpcs)
- [Module Registration](#module-registration)
- [Starting Both Servers](#starting-both-servers)
- [Guards on gRPC Routes](#guards-on-grpc-routes)
- [GRPCContext](#grpccontext)
- [Combining HTTP + gRPC](#combining-http--grpc)
- [Testing with grpcurl](#testing-with-grpcurl)
- [API Reference](#api-reference)

---

## Overview

Nexgou provides first-class gRPC support without requiring `.proto` files or `protoc` code generation. Service descriptors are written in pure Go using `grpc.ServiceDesc`, making them **wire-compatible** with any standard gRPC client. Guards work on unary RPCs transparently, receiving gRPC metadata mapped to HTTP headers.

---

## How It Works (No `.proto` Required)

Normally, `protoc` generates:
1. Message struct types from your `.proto` file
2. A `ServiceDesc` struct describing your service

Nexgou lets you write both by hand. The resulting wire format is identical to protoc-generated code because it uses the same `google.golang.org/grpc` and `google.golang.org/protobuf` packages under the hood.

**Structure per gRPC service:**

```
myservice/
├── service.go      ← message types + ServiceDesc + handler interface
├── controller.go   ← implements the handler interface + Register()/RegisterGRPC()
└── module.go       ← nexgou.Module(...)
```

---

## Message Types

Each protobuf message needs four elements to be wire-compatible:

```go
import "google.golang.org/protobuf/runtime/protoimpl"

type HelloRequest struct {
    state         protoimpl.MessageState `protogen:"open.v1"`
    Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
    unknownFields protoimpl.UnknownFields
    sizeCache     protoimpl.SizeCache
}

// Required by proto.Message interface
func (x *HelloRequest) ProtoMessage() {}
func (x *HelloRequest) Reset()        { *x = HelloRequest{} }
func (x *HelloRequest) String() string { return x.Name }
```

The `protobuf:"..."` struct tag mirrors what protoc would generate for a proto3 field:

| Tag component | Meaning |
|:---|:---|
| `bytes` | Wire type (bytes for strings, messages) |
| `1` | Field number (must match the `.proto` definition if interoperating) |
| `opt` | Optional field |
| `name=name` | Field name as it appears in the proto definition |
| `proto3` | Uses proto3 semantics (zero-value omission) |

---

## Service Interface & Descriptor

Define an interface that your controller will implement, then wire it into a `grpc.ServiceDesc`:

### Unary RPC

```go
// Unary handler low-level signature
func _Greeter_SayHello_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
    in := new(HelloRequest)
    if err := dec(in); err != nil {
        return nil, err
    }
    if interceptor == nil {
        return srv.(GreeterHandler).SayHello(ctx, in)
    }
    info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/greeter.Greeter/SayHello"}
    handler := func(ctx context.Context, req any) (any, error) {
        return srv.(GreeterHandler).SayHello(ctx, req.(*HelloRequest))
    }
    return interceptor(ctx, in, info, handler)
}

type GreeterHandler interface {
    SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error)
}

var Greeter_ServiceDesc = grpc.ServiceDesc{
    ServiceName: "greeter.Greeter",     // must be "package.ServiceName"
    HandlerType: (*GreeterHandler)(nil),
    Methods: []grpc.MethodDesc{
        {
            MethodName: "SayHello",
            Handler:    _Greeter_SayHello_Handler,
        },
    },
    Streams: []grpc.StreamDesc{},
}
```

---

## Server-Streaming RPCs

For server-side streaming, use `grpc.StreamDesc` and `grpc.ServerStream`:

```go
// Low-level stream handler
func _Greeter_SayHelloStream_Handler(srv any, stream grpc.ServerStream) error {
    in := new(HelloRequest)
    if err := stream.RecvMsg(in); err != nil {
        return err
    }
    return srv.(StreamingGreeterHandler).SayHelloStream(in, stream)
}

type StreamingGreeterHandler interface {
    SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error)
    SayHelloStream(req *HelloRequest, stream grpc.ServerStream) error
}

var Greeter_ServiceDesc = grpc.ServiceDesc{
    ServiceName: "greeter.Greeter",
    HandlerType: (*StreamingGreeterHandler)(nil),
    Methods: []grpc.MethodDesc{
        {MethodName: "SayHello", Handler: _Greeter_SayHello_Handler},
    },
    Streams: []grpc.StreamDesc{
        {
            StreamName:    "SayHelloStream",
            Handler:       _Greeter_SayHelloStream_Handler,
            ServerStreams: true,
        },
    },
}
```

Implement the streaming handler in your controller:

```go
func (c *GreeterController) SayHelloStream(req *HelloRequest, stream grpc.ServerStream) error {
    for i := 0; i < 5; i++ {
        reply := &HelloReply{Message: fmt.Sprintf("Hello %s (%d)", req.Name, i+1)}
        if err := stream.SendMsg(reply); err != nil {
            return err
        }
        time.Sleep(500 * time.Millisecond)
    }
    return nil
}
```

---

## Controller Implementation

```go
// greeter/controller.go
package greeter

import (
    "context"
    "fmt"

    nexgou "github.com/nexgou/server"
)

type GreeterController struct {
    log *nexgou.ScopedLogger
}

func NewGreeterController(log *nexgou.LoggerService) *GreeterController {
    return &GreeterController{log: log.WithContext("GreeterController")}
}

// HTTP health check (optional companion route)
func (c *GreeterController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.Get("/health", c.Health),
    }
}

func (c *GreeterController) Health(ctx *nexgou.Context) error {
    return ctx.JSON(200, nexgou.H{"status": "ok"})
}

// gRPC routes
func (c *GreeterController) RegisterGRPC() []nexgou.GRPCRoute {
    return []nexgou.GRPCRoute{
        nexgou.GRPC(Greeter_ServiceDesc, c).
            Guard(&AuthGuard{}).
            Version("v1"),
    }
}

// Unary RPC implementation
func (c *GreeterController) SayHello(_ context.Context, req *HelloRequest) (*HelloReply, error) {
    c.log.Info("SayHello called", "name", req.Name)
    return &HelloReply{Message: fmt.Sprintf("Hello, %s!", req.Name)}, nil
}
```

---

## Module Registration

```go
// greeter/module.go
package greeter

import nexgou "github.com/nexgou/server"

var Module = nexgou.Module(nexgou.ModuleOptions{
    Controllers: []any{NewGreeterController},
})
```

The framework auto-detects `GRPCController` by checking for the `RegisterGRPC()` method.

---

## Starting Both Servers

Run gRPC and HTTP on separate ports concurrently:

```go
// main.go
func main() {
    app := nexgou.CreateApp(AppModule)

    app.Use(middleware.Recovery())
    app.Use(middleware.Logger())
    app.SetFilter(&filter.HttpExceptionFilter{})

    // gRPC on :50051 (non-blocking)
    go func() {
        if err := app.ListenGRPC(50051); err != nil {
            log.Fatalf("gRPC server error: %v", err)
        }
    }()

    // HTTP on :3000 (blocking)
    log.Fatal(app.Listen(3000))
}
```

Both servers share the same module tree and DI container — services are instantiated once and shared.

---

## Guards on gRPC Routes

Guards attached to a `GRPCRoute` run on every **unary** RPC in that service. They receive a `*nexgou.Context` with:

- `ctx.Header("key")` → reads gRPC incoming metadata value for `key`
- `ctx.Method()` → returns the full gRPC method path (e.g. `/greeter.Greeter/SayHello`)

This means the **same guard** works for both HTTP and gRPC without modification:

```go
type AuthGuard struct{}

func (g *AuthGuard) CanActivate(ctx *nexgou.Context) (bool, error) {
    token := ctx.Header("authorization") // works for both HTTP and gRPC metadata
    return validateToken(token), nil
}

// Works on HTTP:
nexgou.Get("/users", c.List).Guard(&AuthGuard{})

// Works on gRPC:
nexgou.GRPC(MyService_ServiceDesc, c).Guard(&AuthGuard{})
```

> Note: Guards on `GRPCRoute` only apply to **unary** RPCs. Streaming RPCs are not currently intercepted by guards.

---

## GRPCContext

For advanced use cases (e.g. middleware that needs to inspect gRPC metadata), use `nexgou.GRPCContext`:

```go
func (c *MyController) MyRPC(ctx context.Context, req *MyRequest) (*MyResponse, error) {
    // The standard context.Context carries gRPC metadata
    md, ok := metadata.FromIncomingContext(ctx)
    if ok {
        tokens := md.Get("authorization")
    }
    return &MyResponse{}, nil
}
```

When guards run, Nexgou internally wraps the gRPC context into a `*nexgou.Context`. You don't need to use `GRPCContext` directly in most cases.

---

## Combining HTTP + gRPC

A single controller can implement both:

```go
func (c *GreeterController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.Get("/health", c.Health),
        nexgou.Get("/greet/:name", c.HTTPGreet),
    }
}

func (c *GreeterController) RegisterGRPC() []nexgou.GRPCRoute {
    return []nexgou.GRPCRoute{
        nexgou.GRPC(Greeter_ServiceDesc, c),
    }
}
```

Both are registered from the same `Controllers` entry in the module.

---

## Testing with grpcurl

[grpcurl](https://github.com/fullstorydev/grpcurl) is the easiest way to test gRPC services from the command line.

```bash
# List services (requires server reflection — add grpc/reflection to your server for this)
grpcurl -plaintext localhost:50051 list

# Call a unary RPC
grpcurl -plaintext -d '{"name": "World"}' localhost:50051 greeter.Greeter/SayHello

# With metadata (maps to gRPC metadata, read by guards via ctx.Header)
grpcurl -plaintext \
  -H 'authorization: Bearer my-token' \
  -d '{"name": "Alice"}' \
  localhost:50051 greeter.Greeter/SayHello
```

---

## API Reference

### `nexgou.GRPC(desc, impl)`

```go
func GRPC(desc GRPCServiceDesc, impl any) GRPCRoute
```

Creates a `GRPCRoute` binding a service descriptor to its implementation.

### `GRPCRoute` methods

| Method | Returns | Description |
|:---|:---|:---|
| `.Guard(guards ...Guard)` | `GRPCRoute` | Attach guards to all unary RPCs (fluent) |
| `.Version(v string)` | `GRPCRoute` | Set a version tag (informational, fluent) |
| `.Ver()` | `string` | Return the version label |

### `App` methods

| Method | Description |
|:---|:---|
| `app.ListenGRPC(port int, ip ...string) error` | Start the gRPC server (blocking) |

### Type aliases

| Alias | Actual type | Description |
|:---|:---|:---|
| `nexgou.GRPCServiceDesc` | `grpc.ServiceDesc` | Service descriptor |
| `nexgou.GRPCServerStream` | `grpc.ServerStream` | Server-side streaming interface |
| `nexgou.GRPCServer` | `*nexgougrpc.GRPCServer` | The underlying gRPC server wrapper |
