package main

import (
	nexgou "github.com/nexgou/server"
	"github.com/nexgou/server/samples/api/user"
)

// AppModule is the root module of the REST API sample application.
// It imports ConfigModule and LogModule for environment config and structured
// logging, plus the UserModule which owns the /v1/users resource.
var AppModule = nexgou.Module(nexgou.ModuleOptions{
	Imports: []nexgou.IModule{
		nexgou.ConfigModule,
		nexgou.LogModule,
		user.Module,
	},
})
