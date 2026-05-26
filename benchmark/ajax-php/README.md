# Ajax PHP Benchmark

PHP implementation of the shared HTTP benchmark contract using PDO SQLite.

## Docker

```powershell
docker compose -f benchmark/_shared/docker-compose.yml up --build ajax-php
```

## k6

```powershell
k6 run -e BASE_URL=http://localhost:3007 benchmark/_shared/k6/scenarios/smoke.js
k6 run -e BASE_URL=http://localhost:3007 -e VUS=2 -e DURATION=3s --summary-export benchmark/_shared/k6/results/ajax-php-crud-mixed.json benchmark/_shared/k6/scenarios/crud-mixed.js
```
