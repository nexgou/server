package task

import (
	nexgou "github.com/nexgou/server"
	"github.com/nexgou/server/src/module/cron"
	"github.com/nexgou/server/src/module/events"
	jwtmod "github.com/nexgou/server/src/module/jwt"
	"github.com/nexgou/server/src/module/sqlite"
	"github.com/nexgou/server/src/module/validation"
)

// Module groups all task-related controllers, providers, and cron jobs.
var Module = nexgou.Module(nexgou.ModuleOptions{
	Imports: []nexgou.IModule{
		sqlite.Module,
		events.Module,
		cron.Module,
		jwtmod.Module,
		validation.ValidationModule,
	},
	Controllers: []any{NewTaskController},
	Providers:   []any{NewTaskService, NewTaskCron},
})
