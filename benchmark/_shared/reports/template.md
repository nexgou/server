# HTTP Benchmark Report

## Metadata

| Campo     | Valor          |
| --------- | -------------- |
| Servicio  | `{{SERVICE}}`  |
| Fecha     | `{{DATE}}`     |
| Escenario | `{{SCENARIO}}` |
| Base URL  | `{{BASE_URL}}` |

## Metricas obligatorias

| Metrica     |            Valor |
| ----------- | ---------------: |
| throughput  | `{{THROUGHPUT}}` |
| p50         |        `{{P50}}` |
| p95         |        `{{P95}}` |
| p99         |        `{{P99}}` |
| error rate  | `{{ERROR_RATE}}` |
| checks      |     `{{CHECKS}}` |
| CPU         |        `{{CPU}}` |
| memoria     |     `{{MEMORY}}` |
| estabilidad |  `{{STABILITY}}` |

## Notas

- El JSON de k6 debe guardarse en `benchmark/_shared/k6/results/`.
- El informe generado debe guardarse en `benchmark/_shared/reports/generated/`.
