param(
    [Parameter(Mandatory = $true)][string]$Service,
    [string]$BaseUrl,
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

$result = Join-Path $Root "benchmark/_shared/k6/results/$Service-smoke.json"
k6 run -e BASE_URL=$BaseUrl --summary-export $result (Join-Path $Root "benchmark/_shared/k6/scenarios/smoke.js")
