package core_test

import (
	"testing"

	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/core"
	modulepkg "github.com/nexgou/server/src/module"
)

func TestCoreNewModuleReturnsOptions(t *testing.T) {
	provider := func() *repository { return &repository{} }
	module := core.NewModule(common.ModuleOptions{Providers: []any{provider}})

	options := module.Options()
	if options == nil {
		t.Fatal("Options should not be nil")
	}

	if len(options.Providers) != 1 {
		t.Fatalf("len(Providers) = %d, want 1", len(options.Providers))
	}
}

func TestModulePackageNewReturnsOptions(t *testing.T) {
	controller := func() *testController { return &testController{} }
	module := modulepkg.New(common.ModuleOptions{Controllers: []any{controller}})

	options := module.Options()
	if options == nil {
		t.Fatal("Options should not be nil")
	}

	if len(options.Controllers) != 1 {
		t.Fatalf("len(Controllers) = %d, want 1", len(options.Controllers))
	}
}

type testController struct{}

type repository struct{}

func (controller *testController) Register() []common.Route {
	return nil
}
