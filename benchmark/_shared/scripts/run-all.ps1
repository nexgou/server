param(
    [switch]$Build,
    [int]$Vus = 2,
    [string]$Duration = "3s",
    [string]$Root = (Resolve-Path (Join-Path $PSScriptRoot "../../..")).Path
)

$ErrorActionPreference = "Stop"
$compose = Join-Path $Root "benchmark/_shared/docker-compose.yml"
$services = @(
    @{ Name = "nexgou"; Url = "http://localhost:3001" },
    @{ Name = "fastify"; Url = "http://localhost:3002" },
    @{ Name = "asp-kestrel"; Url = "http://localhost:3003" },
    @{ Name = "actix-web"; Url = "http://localhost:3004" },
    @{ Name = "hyper"; Url = "http://localhost:3005" },
    @{ Name = "vert-x"; Url = "http://localhost:3006" },
    @{ Name = "ajax-php"; Url = "http://localhost:3007" }
)

Push-Location (Join-Path $Root "benchmark/_shared")
try {
    if ($Build) {
        docker compose -f $compose build
    }
    docker compose -f $compose up -d
} finally {
    Pop-Location
}

foreach ($service in $services) {
    Write-Host "Waiting for $($service.Name) at $($service.Url)"
    $ready = $false
    for ($attempt = 1; $attempt -le 60; $attempt++) {
        try {
            $response = Invoke-WebRequest -Uri "$($service.Url)/health" -UseBasicParsing -TimeoutSec 2
            if ($response.StatusCode -eq 200) {
                $ready = $true
                break
            }
        } catch {
        }
        Start-Sleep -Seconds 1
    }
    if (-not $ready) {
        throw "Service $($service.Name) did not become healthy."
    }
    & (Join-Path $PSScriptRoot "run-smoke.ps1") -Service $service.Name -BaseUrl $service.Url -Root $Root
    & (Join-Path $PSScriptRoot "run-crud-mixed.ps1") -Service $service.Name -BaseUrl $service.Url -Vus $Vus -Duration $Duration -Root $Root
}
