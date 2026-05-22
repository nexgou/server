package greeter

import nexgou "github.com/nexgou/server"

// Module registers the GreeterController which exposes both the gRPC Greeter
// service (port 50051) and a companion HTTP /health route (port 3003).
var Module = nexgou.Module(nexgou.ModuleOptions{
	Controllers: []any{NewGreeterController},
})
