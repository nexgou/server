package core

import "github.com/nexgou/server/src/common"

type module struct {
	opts *common.ModuleOptions
}

// Options returns the module's configuration options.
func (m *module) Options() *common.ModuleOptions {
	return m.opts
}

// NewModule creates a new Nexgou module with the given options.
func NewModule(opts common.ModuleOptions) common.IModule {
	return &module{opts: &opts}
}
