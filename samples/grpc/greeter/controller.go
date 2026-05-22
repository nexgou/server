package greeter

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	nexgou "github.com/nexgou/server"
)

// GreeterController implements GreeterHandler (gRPC) and also exposes an HTTP
// health endpoint so the companion HTTP server has at least one route.
type GreeterController struct{}

// NewGreeterController is the IoC constructor used by the Nexgou container.
func NewGreeterController() *GreeterController {
	return &GreeterController{}
}

// ── HTTP routes ───────────────────────────────────────────────────────────────

// Register returns the HTTP routes exposed by this controller.
func (c *GreeterController) Register() []nexgou.Route {
	return []nexgou.Route{
		nexgou.Get("/health", c.Health),
	}
}

// Health handles GET /health — lightweight liveness check.
func (c *GreeterController) Health(ctx *nexgou.Context) error {
	return ctx.JSON(200, map[string]string{"status": "ok"})
}

// ── gRPC routes ───────────────────────────────────────────────────────────────

// RegisterGRPC returns the gRPC services exposed by this controller.
func (c *GreeterController) RegisterGRPC() []nexgou.GRPCRoute {
	return []nexgou.GRPCRoute{
		nexgou.GRPC(Greeter_ServiceDesc, c),
	}
}

// ── GreeterHandler implementation ─────────────────────────────────────────────

// SayHello handles the unary SayHello RPC.
func (c *GreeterController) SayHello(_ context.Context, req *HelloRequest) (*HelloReply, error) {
	name := req.Name
	if name == "" {
		name = "World"
	}
	return &HelloReply{Message: fmt.Sprintf("Hello, %s!", name)}, nil
}

// SayHelloStream handles the server-streaming SayHelloStream RPC.
// It sends three replies with a short delay between them.
func (c *GreeterController) SayHelloStream(req *HelloRequest, stream grpc.ServerStream) error {
	name := req.Name
	if name == "" {
		name = "World"
	}
	for i := 1; i <= 3; i++ {
		reply := &HelloReply{Message: fmt.Sprintf("Hello #%d, %s!", i, name)}
		if err := stream.SendMsg(reply); err != nil {
			return err
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}
