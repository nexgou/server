package database

import (
	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/core"
	"github.com/nexgou/server/src/logger"
)

// Module is NOT meant to be imported directly.
// Use one of the driver-specific modules instead:
//
//	sqlite.Module   — SQLite (embedded, no server required)
//	postgres.Module — PostgreSQL
//	mysql.Module    — MySQL / MariaDB
//
// Those modules register a DatabaseService with the correct driver and DSN
// and re-export it so your services can inject *database.DatabaseService.
var Module common.IModule = core.NewModule(common.ModuleOptions{
	Imports: []common.IModule{
		logger.LogModule,
	},
	// No providers — driver modules provide DatabaseService via their own constructors.
})
