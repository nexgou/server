package auth

import (
	nexgou "github.com/nexgou/server"
	"github.com/nexgou/server/src/module/jwt"
	"github.com/nexgou/server/src/module/validation"
)

// Module groups all auth-related providers and controllers.
var Module = nexgou.Module(nexgou.ModuleOptions{
	Imports: []nexgou.IModule{
		jwt.Module,
		validation.ValidationModule,
	},
	Controllers: []any{NewAuthController},
	Providers:   []any{NewAuthService},
})
