package chat

import nexgou "github.com/nexgou/server"

// Module groups all chat-related controllers and providers.
var Module = nexgou.Module(nexgou.ModuleOptions{
	Controllers: []any{NewChatController},
})
