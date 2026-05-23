package validation

import (
	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/core"
	"github.com/nexgou/server/src/logger"
)

// ValidationModule registers *ValidationService in the Nexgou DI container.
//
// Import it in any module that needs struct validation:
//
//	core.NewModule(common.ModuleOptions{
//	    Imports:  []common.IModule{validation.ValidationModule},
//	    Providers: []any{NewUserService},
//	})
var ValidationModule common.IModule = core.NewModule(common.ModuleOptions{
	Imports: []common.IModule{
		logger.LogModule,
	},
	Providers: []any{
		NewValidationService,
	},
	Exports: []any{
		NewValidationService,
	},
})
