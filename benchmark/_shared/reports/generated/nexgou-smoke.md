# NexGou Smoke Benchmark

Estado: ejecutado correctamente con k6 local.

Comando esperado:

```bash
k6 run -e BASE_URL=http://localhost:3001 --summary-export benchmark/_shared/k6/results/nexgou-smoke.json benchmark/_shared/k6/scenarios/smoke.js
```

Metricas smoke:

| Metrica     |                              Valor |
| ----------- | ---------------------------------: |
| throughput  |                       214.64 req/s |
| p50         |                          814.95 us |
| p95         |                            1.68 ms |
| p99         |             no reportado por smoke |
| error rate  |                              0.00% |
| checks      |                        12/12, 100% |
| CPU         | pendiente de Docker/observabilidad |
| memoria     | pendiente de Docker/observabilidad |
| estabilidad |                           smoke OK |

Metricas CRUD mixto corto:

| Metrica    |             Valor |
| ---------- | ----------------: |
| VUs        |                 2 |
| duracion   |                3s |
| throughput |     3884.95 req/s |
| p50        |          526.5 us |
| p95        |         735.02 us |
| error rate |             0.00% |
| checks     | 23368/23368, 100% |
