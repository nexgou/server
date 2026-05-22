package metrics

import nexgou "github.com/nexgou/server"

// Module groups all metrics-related controllers.
var Module = nexgou.Module(nexgou.ModuleOptions{
	Controllers: []any{NewMetricsController},
})
