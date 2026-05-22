package main

import (
	nexgou "github.com/nexgou/server"
	"github.com/nexgou/server/samples/sse/metrics"
)

// AppModule is the root module of the SSE metrics streaming sample.
var AppModule = nexgou.Module(nexgou.ModuleOptions{
	Imports: []nexgou.IModule{
		nexgou.LogModule,
		metrics.Module,
	},
})
