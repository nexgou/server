// Package sqlite provides a SQLite driver module for Nexgou.
//
// It uses modernc.org/sqlite — a pure-Go SQLite implementation that requires
// no CGO and no external libraries.
//
// Configuration (environment variables):
//
//	SQLITE_PATH — path to the SQLite database file (default: "nexgou.db")
//	              Use ":memory:" for an in-memory database.
//
// Usage:
//
//	var AppModule = nexgou.Module(nexgou.ModuleOptions{
//	    Imports: []nexgou.IModule{
//	        nexgou.ConfigModule,
//	        nexgou.LogModule,
//	        sqlite.Module,
//	    },
//	})
//
//	func NewUserRepo(db *database.DatabaseService) *UserRepo { ... }
package sqlite

import (
	_ "modernc.org/sqlite" // register "sqlite" driver

	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/config"
	"github.com/nexgou/server/src/core"
	"github.com/nexgou/server/src/logger"
	"github.com/nexgou/server/src/module/database"
)

// NewSQLiteDatabaseService constructs a DatabaseService backed by SQLite.
func NewSQLiteDatabaseService(cfg *config.ConfigService, log *logger.LoggerService) *database.DatabaseService {
	path := cfg.GetOrDefault("SQLITE_PATH", "nexgou.db")
	return database.NewDatabaseServiceFromConfig(database.Config{
		Driver:       "sqlite",
		DSN:          path,
		MaxOpenConns: 1, // SQLite supports only one writer at a time
		MaxIdleConns: 1,
	}, log)
}

// Module registers and exports a *database.DatabaseService backed by SQLite.
var Module common.IModule = core.NewModule(common.ModuleOptions{
	Imports: []common.IModule{
		config.ConfigModule,
		logger.LogModule,
	},
	Providers: []any{NewSQLiteDatabaseService},
	Exports:   []any{NewSQLiteDatabaseService},
})
