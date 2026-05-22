// Package greeter implements the Greeter gRPC service without protoc / .proto
// files. Service descriptors and message types are defined in pure Go using
// protobuf struct tags so they are wire-compatible with generated code.
//
// Wire format: google.golang.org/protobuf/proto (binary protobuf).
// Client can use any language/tool that speaks gRPC (grpcurl, Postman, generated stubs).
package greeter

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/runtime/protoimpl"
)

// ── Message types ─────────────────────────────────────────────────────────────

// HelloRequest is the request message for the SayHello and SayHelloStream RPCs.
//
// Proto equivalent:
//
//	message HelloRequest { string name = 1; }
type HelloRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HelloRequest) ProtoMessage()             {}
func (x *HelloRequest) Reset()                    { *x = HelloRequest{} }
func (x *HelloRequest) String() string            { return fmt.Sprintf("name:%q", x.Name) }
func (x *HelloRequest) ProtoReflect() protoreflect { return nil } // simplified — not used by grpc transport

// HelloReply is the response message for the SayHello and SayHelloStream RPCs.
//
// Proto equivalent:
//
//	message HelloReply { string message = 1; }
type HelloReply struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Message       string                 `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HelloReply) ProtoMessage()             {}
func (x *HelloReply) Reset()                    { *x = HelloReply{} }
func (x *HelloReply) String() string            { return fmt.Sprintf("message:%q", x.Message) }
func (x *HelloReply) ProtoReflect() protoreflect { return nil }

// ── Service descriptor ────────────────────────────────────────────────────────

// GreeterServiceName is the fully-qualified gRPC service name.
// This must match the package + service name in the .proto file if you later
// generate a client stub with protoc.
const GreeterServiceName = "greeter.Greeter"

// GreeterHandler is the interface that the server implementation must satisfy.
// grpc.ServiceDesc references these method names at registration time.
type GreeterHandler interface {
	// SayHello handles a single-shot unary "Hello" RPC.
	SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error)

	// SayHelloStream streams a sequence of HelloReply messages back to the client.
	SayHelloStream(req *HelloRequest, stream grpc.ServerStream) error
}

// Greeter_ServiceDesc is the grpc.ServiceDesc for the Greeter service.
// Register it with nexgougrpc.NewRoute(Greeter_ServiceDesc, impl).
//
// Hand-written equivalent of what protoc-gen-go-grpc generates.
var Greeter_ServiceDesc = grpc.ServiceDesc{
	ServiceName: GreeterServiceName,
	HandlerType: (*GreeterHandler)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SayHello",
			Handler:    _Greeter_SayHello_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SayHelloStream",
			Handler:       _Greeter_SayHelloStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "greeter/greeter.proto",
}

// _Greeter_SayHello_Handler is the low-level handler called by the gRPC runtime
// for the unary SayHello RPC.
func _Greeter_SayHello_Handler(
	srv any,
	ctx context.Context,
	dec func(any) error,
	interceptor grpc.UnaryServerInterceptor,
) (any, error) {
	in := new(HelloRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GreeterHandler).SayHello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: fmt.Sprintf("/%s/SayHello", GreeterServiceName),
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(GreeterHandler).SayHello(ctx, req.(*HelloRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// _Greeter_SayHelloStream_Handler is the low-level handler for the server-streaming
// SayHelloStream RPC.
func _Greeter_SayHelloStream_Handler(
	srv any,
	stream grpc.ServerStream,
) error {
	in := new(HelloRequest)
	if err := stream.RecvMsg(in); err != nil {
		return err
	}
	return srv.(GreeterHandler).SayHelloStream(in, stream)
}

// protoreflect is a minimal stand-in to satisfy the proto.Message interface
// without importing google.golang.org/protobuf/reflect/protoreflect.
// The real reflection API is not needed for transport-only usage.
type protoreflect = interface{}
