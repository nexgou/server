# NexGou Benchmark

Implementacion NexGou del contrato HTTP comun definido en `docs/BENCHMARK.md`.

## Contrato

```txt
GET    /health
POST   /users
GET    /users/:id
GET    /users?limit=20&offset=0
PUT    /users/:id
DELETE /users/:id
```

## Ejecutar local

```bash
PORT=3001 DB_PATH=benchmark/nexgou/data/db.sqlite go run ./benchmark/nexgou/cmd/server
```

En PowerShell:

```powershell
$env:PORT="3001"; $env:DB_PATH="benchmark/nexgou/data/db.sqlite"; go run ./benchmark/nexgou/cmd/server
```

## Smoke

```bash
go test ./test/benchmark
```

## k6

```bash
k6 run -e BASE_URL=http://localhost:3001 benchmark/_shared/k6/scenarios/smoke.js
k6 run -e BASE_URL=http://localhost:3001 --summary-export benchmark/_shared/k6/results/nexgou-crud-mixed.json benchmark/_shared/k6/scenarios/crud-mixed.js
```
