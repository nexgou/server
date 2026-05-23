package cron

import (
	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/core"
	"github.com/nexgou/server/src/logger"
)

// Module is a ready-to-use Nexgou module that registers and exports CronService.
//
// Usage:
//
//	var AppModule = nexgou.Module(nexgou.ModuleOptions{
//	    Imports: []nexgou.IModule{nexgou.LogModule, cron.Module},
//	})
//
//	func NewReportService(cron *cron.CronService) *ReportService { ... }
var Module common.IModule = core.NewModule(common.ModuleOptions{
	Imports: []common.IModule{
		logger.LogModule,
	},
	Providers: []any{NewCronService},
	Exports:   []any{NewCronService},
})
