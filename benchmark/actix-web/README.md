# Actix Web Benchmark

Rust Actix Web implementation of the shared HTTP benchmark contract.

## Docker

```powershell
docker compose -f benchmark/_shared/docker-compose.yml up --build actix-web
```

## k6

```powershell
k6 run -e BASE_URL=http://localhost:3004 benchmark/_shared/k6/scenarios/smoke.js
k6 run -e BASE_URL=http://localhost:3004 -e VUS=2 -e DURATION=3s --summary-export benchmark/_shared/k6/results/actix-web-crud-mixed.json benchmark/_shared/k6/scenarios/crud-mixed.js
```
