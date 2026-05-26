param(
    [switch]$Results,
    [switch]$Volumes,
    [string]$Root = (Resolve-Path (Join-Path $PSScriptRoot "../../..")).Path
)

$ErrorActionPreference = "Stop"

if ($Results) {
    Get-ChildItem (Join-Path $Root "benchmark/_shared/k6/results") -Filter "*.json" | Remove-Item -Force
    Get-ChildItem (Join-Path $Root "benchmark/_shared/reports/generated") -Filter "*.md" | Remove-Item -Force
}

if ($Volumes) {
    docker compose -f (Join-Path $Root "benchmark/_shared/docker-compose.yml") down -v
} else {
    docker compose -f (Join-Path $Root "benchmark/_shared/docker-compose.yml") down
}
