package module

import (
	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/core"
)

// New creates a NexGou module. It mirrors core.NewModule for public module imports.
func New(options common.ModuleOptions) common.IModule {
	return core.NewModule(options)
}
