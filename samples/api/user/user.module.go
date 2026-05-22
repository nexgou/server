package user

import nexgou "github.com/nexgou/server"

// Module groups all user-related controllers and providers.
var Module = nexgou.Module(nexgou.ModuleOptions{
	Controllers: []any{NewUserController},
	Providers:   []any{NewUserService},
})
