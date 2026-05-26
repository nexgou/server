param(
    [Parameter(Mandatory = $true)][string]$Service,
    [string]$BaseUrl,
    [int]$Vus = 20,
    [string]$Duration = "30s",
    [switch]$Prometheus,
    [string]$Root = (Resolve-Path (Join-Path $PSScriptRoot "../../..")).Path
)

$ErrorActionPreference = "Stop"
$urls = @{
    "nexgou" = "http://localhost:3001"
    "fastify" = "http://localhost:3002"
    "asp-kestrel" = "http://localhost:3003"
    "actix-web" = "http://localhost:3004"
    "hyper" = "http://localhost:3005"
    "vert-x" = "http://localhost:3006"
    "ajax-php" = "http://localhost:3007"
}

if (-not $BaseUrl) {
    $BaseUrl = $urls[$Service]
}
if (-not $BaseUrl) {
    throw "Unknown service '$Service'. Provide -BaseUrl."
}

$result = Join-Path $Root "benchmark/_shared/k6/results/$Service-crud-mixed.json"
$scenario = Join-Path $Root "benchmark/_shared/k6/scenarios/crud-mixed.js"

if ($Prometheus) {
    if (-not $env:K6_PROMETHEUS_RW_SERVER_URL) {
        $env:K6_PROMETHEUS_RW_SERVER_URL = "http://localhost:9090/api/v1/write"
    }
    k6 run -o experimental-prometheus-rw -e BASE_URL=$BaseUrl -e VUS=$Vus -e DURATION=$Duration --summary-export $result $scenario
} else {
    k6 run -e BASE_URL=$BaseUrl -e VUS=$Vus -e DURATION=$Duration --summary-export $result $scenario
}
