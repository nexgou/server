package main

import (
	nexgou "github.com/nexgou/server"
	"github.com/nexgou/server/samples/taskboard/auth"
	"github.com/nexgou/server/samples/taskboard/task"
)

// AppModule is the root module of the Taskboard demo application.
// It wires together ConfigModule, LogModule, and the feature modules.
var AppModule = nexgou.Module(nexgou.ModuleOptions{
	Imports: []nexgou.IModule{
		nexgou.ConfigModule,
		nexgou.LogModule,
		auth.Module,
		task.Module,
	},
})
