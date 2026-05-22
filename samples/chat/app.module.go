package main

import (
	nexgou "github.com/nexgou/server"
	"github.com/nexgou/server/samples/chat/chat"
)

// AppModule is the root module of the WebSocket chat sample.
var AppModule = nexgou.Module(nexgou.ModuleOptions{
	Imports: []nexgou.IModule{
		nexgou.LogModule,
		chat.Module,
	},
})
