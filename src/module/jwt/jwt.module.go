package jwt

import (
	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/config"
	"github.com/nexgou/server/src/core"
	"github.com/nexgou/server/src/logger"
)

// Module is a ready-to-use Nexgou module that registers and exports JwtService.
// It requires ConfigModule and LogModule to be imported (directly or transitively).
//
// Usage:
//
//	var AppModule = nexgou.Module(nexgou.ModuleOptions{
//	    Imports: []nexgou.IModule{
//	        nexgou.ConfigModule,
//	        nexgou.LogModule,
//	        jwt.Module,
//	    },
//	})
var Module common.IModule = core.NewModule(common.ModuleOptions{
	Imports: []common.IModule{
		config.ConfigModule,
		logger.LogModule,
	},
	Providers: []any{NewJwtService},
	Exports:   []any{NewJwtService},
})
