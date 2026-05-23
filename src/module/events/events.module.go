package events

import (
	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/core"
	"github.com/nexgou/server/src/logger"
)

// Module is a ready-to-use Nexgou module that registers and exports EventEmitter.
//
// Usage:
//
//	var AppModule = nexgou.Module(nexgou.ModuleOptions{
//	    Imports: []nexgou.IModule{nexgou.LogModule, events.Module},
//	})
//
//	func NewOrderService(emitter *events.EventEmitter) *OrderService { ... }
var Module common.IModule = core.NewModule(common.ModuleOptions{
	Imports: []common.IModule{
		logger.LogModule,
	},
	Providers: []any{NewEventEmitter},
	Exports:   []any{NewEventEmitter},
})
