param(
    [Parameter(Mandatory = $true)][string]$Service,
    [Parameter(Mandatory = $true)][string]$Scenario,
    [string]$Root = (Resolve-Path (Join-Path $PSScriptRoot "../../..")).Path
)

$ErrorActionPreference = "Stop"
$resultPath = Join-Path $Root "benchmark/_shared/k6/results/$Service-$Scenario.json"
$templatePath = Join-Path $Root "benchmark/_shared/reports/template.md"
$outputPath = Join-Path $Root "benchmark/_shared/reports/generated/$Service-$Scenario.md"

if (-not (Test-Path $resultPath)) {
    throw "Result file not found: $resultPath"
}

$summary = Get-Content $resultPath -Raw | ConvertFrom-Json
$template = Get-Content $templatePath -Raw
$metrics = $summary.metrics
$duration = $metrics.http_req_duration
$requests = $metrics.http_reqs
$failed = $metrics.http_req_failed
$checks = $metrics.checks

function MetricValue($metric, $name, $fallback = "n/a") {
    if ($null -eq $metric) { return $fallback }
    $property = $metric.PSObject.Properties[$name]
    if ($null -eq $property) { return $fallback }
    return $property.Value
}

$throughput = MetricValue $requests "rate"
$p50 = MetricValue $duration "p(50)"
$p95 = MetricValue $duration "p(95)"
$p99 = MetricValue $duration "p(99)"
$errorRate = MetricValue $failed "rate"
$checksRate = MetricValue $checks "rate"

$urls = @{
    "nexgou" = "http://localhost:3001"
    "fastify" = "http://localhost:3002"
    "asp-kestrel" = "http://localhost:3003"
    "actix-web" = "http://localhost:3004"
    "hyper" = "http://localhost:3005"
    "vert-x" = "http://localhost:3006"
    "ajax-php" = "http://localhost:3007"
}

$report = $template.Replace("{{SERVICE}}", $Service)
$report = $report.Replace("{{DATE}}", (Get-Date -Format "yyyy-MM-ddTHH:mm:ssK"))
$report = $report.Replace("{{SCENARIO}}", $Scenario)
$baseUrl = $urls[$Service]
if (-not $baseUrl) {
    $baseUrl = "n/a"
}
$report = $report.Replace("{{BASE_URL}}", $baseUrl)
$report = $report.Replace("{{THROUGHPUT}}", "$throughput req/s")
$report = $report.Replace("{{P50}}", "$p50 ms")
$report = $report.Replace("{{P95}}", "$p95 ms")
$report = $report.Replace("{{P99}}", "$p99 ms")
$report = $report.Replace("{{ERROR_RATE}}", "$errorRate")
$report = $report.Replace("{{CHECKS}}", "$checksRate")
$report = $report.Replace("{{CPU}}", "n/a")
$report = $report.Replace("{{MEMORY}}", "n/a")
$report = $report.Replace("{{STABILITY}}", "pending review")

New-Item -ItemType Directory -Force -Path (Split-Path $outputPath) | Out-Null
Set-Content -Path $outputPath -Value $report -NoNewline
Write-Host $outputPath
