package nexgougrpc

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/nexgou/server/src/common"
)

// GRPCServer wraps a *grpc.Server and integrates Nexgou's Guard system
// as a grpc.UnaryServerInterceptor.
//
// Obtain a GRPCServer via NewGRPCServer and register services via RegisterRoute.
// Then call Listen to start accepting connections.
type GRPCServer struct {
	server *grpc.Server
	routes []GRPCRoute
}

// NewGRPCServer creates a new GRPCServer.
// Nexgou automatically prepends its own guard interceptor. Additional
// grpc.ServerOption values (e.g. grpc.MaxRecvMsgSize) are appended after it.
//
//	srv := nexgougrpc.NewGRPCServer()
//	srv := nexgougrpc.NewGRPCServer(grpc.MaxRecvMsgSize(4 << 20))
func NewGRPCServer(opts ...grpc.ServerOption) *GRPCServer {
	g := &GRPCServer{}

	guardOpt := grpc.ChainUnaryInterceptor(g.guardInterceptor)
	allOpts := append([]grpc.ServerOption{guardOpt}, opts...)
	g.server = grpc.NewServer(allOpts...)
	return g
}

// RegisterRoute registers a GRPCRoute (service + guards) on the underlying grpc.Server.
// Call this before Listen.
//
//	srv.RegisterRoute(nexgougrpc.NewRoute(greeter.Greeter_ServiceDesc, impl).Guard(&AuthGuard{}))
func (g *GRPCServer) RegisterRoute(route GRPCRoute) {
	g.routes = append(g.routes, route)
	g.server.RegisterService(&route.Desc, route.Impl)
}

// Server returns the underlying *grpc.Server for advanced configuration
// (e.g. server reflection, health checking).
func (g *GRPCServer) Server() *grpc.Server {
	return g.server
}

// Listen starts the gRPC server on the given TCP port.
// It blocks until the server is stopped or returns a non-nil error.
//
//	srv.Listen(50051)
//	srv.Listen(50051, "127.0.0.1")
func (g *GRPCServer) Listen(port int, ip ...string) error {
	host := ""
	if len(ip) > 0 {
		host = ip[0]
	}
	addr := fmt.Sprintf("%s:%d", host, port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("nexgou/grpc: failed to listen on %s: %w", addr, err)
	}

	g.printRoutes(port)
	return g.server.Serve(lis)
}

// Stop performs a graceful shutdown of the gRPC server.
func (g *GRPCServer) Stop() {
	g.server.GracefulStop()
}

// guardInterceptor is the Nexgou-generated grpc.UnaryServerInterceptor.
// It evaluates every guard registered on the matching GRPCRoute before
// delegating to the actual RPC handler.
//
// Guard adapter contract:
//   - guard returns (false, nil)  → codes.PermissionDenied
//   - guard returns (_, err)      → gRPC status derived from HttpException or codes.Internal
//   - guard returns (true, nil)   → proceed
func (g *GRPCServer) guardInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	guards := g.guardsForMethod(info.FullMethod)

	for _, guard := range guards {
		httpCtx := grpcContextToHTTPContext(ctx, info.FullMethod)
		ok, err := guard.CanActivate(httpCtx)
		if err != nil {
			return nil, grpcStatusFromError(err)
		}
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "Forbidden")
		}
	}

	return handler(ctx, req)
}

// guardsForMethod returns the guards for the service that owns fullMethod.
// "/greeter.Greeter/SayHello" → looks up service "greeter.Greeter".
func (g *GRPCServer) guardsForMethod(fullMethod string) []common.Guard {
	svcName := serviceNameFromMethod(fullMethod)
	for i := range g.routes {
		if g.routes[i].Desc.ServiceName == svcName {
			return g.routes[i].Guards
		}
	}
	return nil
}

// printRoutes logs the registered gRPC services to stdout at startup.
func (g *GRPCServer) printRoutes(port int) {
	cyan := "\033[36m"
	reset := "\033[0m"
	dim := "\033[2m"
	gray := "\033[90m"

	fmt.Printf("\n%s[Nexgou gRPC]%s listening on :%d\n", cyan, reset, port)
	fmt.Printf("%s────────────────────────────────%s\n", gray, reset)
	for i := range g.routes {
		r := &g.routes[i]
		badge := fmt.Sprintf("%s🌐 public%s", gray, reset)
		if len(r.Guards) > 0 {
			badge = fmt.Sprintf("🔒 %s%d guard(s)%s", dim, len(r.Guards), reset)
		}
		ver := ""
		if r.ver != "" {
			ver = fmt.Sprintf(" %s[%s]%s", dim, r.ver, reset)
		}
		fmt.Printf("%sRPC%s   /%s%s   %s\n", cyan, reset, r.Desc.ServiceName, ver, badge)

		for _, m := range r.Desc.Methods {
			fmt.Printf("       %s↳ %s (unary)%s\n", dim, m.MethodName, reset)
		}
		for _, s := range r.Desc.Streams {
			fmt.Printf("       %s↳ %s %s%s\n", dim, s.StreamName, streamLabel(s), reset)
		}
	}
	fmt.Println()
}

// ── helpers ───────────────────────────────────────────────────────────────────

// grpcContextToHTTPContext creates a *common.Context (HTTP adapter) from a gRPC
// server context so that existing Guard implementations can read metadata via
// ctx.Header(key) without modification.
//
// Each incoming gRPC metadata key is written as a canonical HTTP header.
// The synthetic request carries the gRPC full method as the URL path and
// no body — guards that call ctx.Body() will receive an EOF error.
func grpcContextToHTTPContext(ctx context.Context, fullMethod string) *common.Context {
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, fullMethod, http.NoBody)

	// Copy all gRPC incoming metadata into HTTP request headers.
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for k, vals := range md {
			for _, v := range vals {
				r.Header.Add(k, v)
			}
		}
	}

	return common.NewContext(noopResponseWriter{}, r, nil)
}

// grpcStatusFromError converts a Nexgou HttpException (or any error) into a
// gRPC status error so the client receives a meaningful status code.
func grpcStatusFromError(err error) error {
	if ex, ok := err.(*common.HttpException); ok {
		code := httpStatusToGRPCCode(ex.Status)
		return status.Error(code, ex.Message)
	}
	return status.Error(codes.Internal, err.Error())
}

// httpStatusToGRPCCode maps common HTTP status codes to gRPC equivalents.
func httpStatusToGRPCCode(httpStatus int) codes.Code {
	switch httpStatus {
	case 400:
		return codes.InvalidArgument
	case 401:
		return codes.Unauthenticated
	case 403:
		return codes.PermissionDenied
	case 404:
		return codes.NotFound
	case 409:
		return codes.AlreadyExists
	case 429:
		return codes.ResourceExhausted
	case 501:
		return codes.Unimplemented
	case 503:
		return codes.Unavailable
	default:
		return codes.Internal
	}
}

// serviceNameFromMethod extracts the service name from a gRPC full method path.
// "/greeter.Greeter/SayHello" → "greeter.Greeter"
func serviceNameFromMethod(fullMethod string) string {
	if fullMethod == "" || fullMethod[0] != '/' {
		return fullMethod
	}
	trimmed := fullMethod[1:]
	for i, ch := range trimmed {
		if ch == '/' {
			return trimmed[:i]
		}
	}
	return trimmed
}

// streamLabel returns a human-readable stream direction label.
func streamLabel(s grpc.StreamDesc) string {
	switch {
	case s.ClientStreams && s.ServerStreams:
		return "(bidi-stream)"
	case s.ClientStreams:
		return "(client-stream)"
	case s.ServerStreams:
		return "(server-stream)"
	default:
		return "(stream)"
	}
}

// noopResponseWriter satisfies http.ResponseWriter with no-op implementations.
// Used only to satisfy the common.Context constructor for gRPC-adapted contexts.
// Guards that attempt to write a response (e.g. ctx.JSON) will silently discard output.
type noopResponseWriter struct{}

func (noopResponseWriter) Header() http.Header         { return http.Header{} }
func (noopResponseWriter) Write(b []byte) (int, error) { return len(b), nil }
func (noopResponseWriter) WriteHeader(_ int)           {}
