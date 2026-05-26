# Fastify Benchmark

Fastify implementation of the shared HTTP benchmark contract.

## Local

```powershell
$env:PORT="3002"; $env:DB_PATH="benchmark/fastify/data/db.sqlite"; npm --prefix benchmark/fastify install; npm --prefix benchmark/fastify start
```

## Docker

```powershell
docker compose -f benchmark/_shared/docker-compose.yml up --build fastify
```

## k6

```powershell
k6 run -e BASE_URL=http://localhost:3002 benchmark/_shared/k6/scenarios/smoke.js
k6 run -e BASE_URL=http://localhost:3002 -e VUS=2 -e DURATION=3s --summary-export benchmark/_shared/k6/results/fastify-crud-mixed.json benchmark/_shared/k6/scenarios/crud-mixed.js
```
