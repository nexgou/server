# ASP.NET Kestrel Benchmark

ASP.NET Core Minimal API implementation of the shared HTTP benchmark contract.

## Local

```powershell
$env:PORT="3003"; $env:DB_PATH="benchmark/asp-kestrel/data/db.sqlite"; dotnet run --project benchmark/asp-kestrel/asp-kestrel.csproj
```

## Docker

```powershell
docker compose -f benchmark/_shared/docker-compose.yml up --build asp-kestrel
```

## k6

```powershell
k6 run -e BASE_URL=http://localhost:3003 benchmark/_shared/k6/scenarios/smoke.js
k6 run -e BASE_URL=http://localhost:3003 -e VUS=2 -e DURATION=3s --summary-export benchmark/_shared/k6/results/asp-kestrel-crud-mixed.json benchmark/_shared/k6/scenarios/crud-mixed.js
```
