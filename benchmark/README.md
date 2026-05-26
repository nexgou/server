# Benchmark HTTP

Laboratorio reproducible para comparar NexGou contra otros servidores HTTP usando el mismo contrato CRUD, SQLite, Docker Compose, k6, Prometheus, Grafana y reportes Markdown.

## Servidores

| Servicio      | Stack                    | Puerto |
| ------------- | ------------------------ | ------ |
| `nexgou`      | Go + NexGou              | 3001   |
| `fastify`     | Node.js + Fastify        | 3002   |
| `asp-kestrel` | ASP.NET Core Minimal API | 3003   |
| `actix-web`   | Rust + Actix Web         | 3004   |
| `hyper`       | Rust + Hyper             | 3005   |
| `vert-x`      | Java + Vert.x Web        | 3006   |
| `ajax-php`    | PHP + PDO SQLite         | 3007   |

## Contrato

Todos los servicios implementan:

```txt
GET    /health
POST   /users
GET    /users/:id
GET    /users?limit=20&offset=0
PUT    /users/:id
DELETE /users/:id
```

Cada contenedor usa su propio archivo SQLite en `/app/data/db.sqlite`, con volumen Docker separado.

## Levantar servicios

```powershell
docker compose -f benchmark/_shared/docker-compose.yml up --build
```

Para levantar un servicio puntual:

```powershell
docker compose -f benchmark/_shared/docker-compose.yml up --build fastify
```

## Smoke por servicio

```powershell
benchmark/_shared/scripts/run-smoke.ps1 -Service nexgou
benchmark/_shared/scripts/run-smoke.ps1 -Service fastify
benchmark/_shared/scripts/run-smoke.ps1 -Service asp-kestrel
benchmark/_shared/scripts/run-smoke.ps1 -Service actix-web
benchmark/_shared/scripts/run-smoke.ps1 -Service hyper
benchmark/_shared/scripts/run-smoke.ps1 -Service vert-x
benchmark/_shared/scripts/run-smoke.ps1 -Service ajax-php
```

## Carga mixta corta

```powershell
benchmark/_shared/scripts/run-crud-mixed.ps1 -Service nexgou -Vus 2 -Duration 3s
```

## Ejecutar todo

```powershell
benchmark/_shared/scripts/run-all.ps1 -Build -Vus 2 -Duration 3s
```

## Observabilidad

```powershell
Push-Location benchmark/_shared
docker compose -f docker-compose.observability.yml up -d
Pop-Location
benchmark/_shared/scripts/run-crud-mixed.ps1 -Service nexgou -Vus 20 -Duration 30s -Prometheus
```

Grafana queda disponible en `http://localhost:3000` con usuario `admin` y password `admin`. Prometheus queda en `http://localhost:9090`.

## Reportes

Los JSON de k6 se guardan en `benchmark/_shared/k6/results/`. Para generar un reporte Markdown:

```powershell
benchmark/_shared/scripts/generate-report.ps1 -Service nexgou -Scenario crud-mixed
```

La salida queda en `benchmark/_shared/reports/generated/`.

## Limpieza

```powershell
benchmark/_shared/scripts/clean.ps1 -Results
benchmark/_shared/scripts/clean.ps1 -Volumes
```
