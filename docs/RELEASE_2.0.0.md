# Release readiness 2.0.0

Estado: release candidate local.

## Alcance

- HTTP core sobre API publica raiz.
- Adapter `fasthttp` encapsulado.
- Samples HTTP: `samples/api`, `samples/taskboard`.
- Benchmark NexGou con SQLite y k6.

Fuera de alcance inicial:

- WebSocket;
- SSE;
- gRPC.

## Validacion ejecutada

```txt
go test ./test/... OK
go test ./... OK
go test ./test/router -run ^$ -bench FullPipeline -benchmem OK
go test ./test/logger -run ^$ -bench . -benchmem OK
golangci-lint run ./... OK
k6 smoke OK
k6 crud-mixed OK
```

## Resultados k6 locales

| Escenario        |      Checks | Error rate |
| ---------------- | ----------: | ---------: |
| smoke            |       12/12 |      0.00% |
| crud-mixed corto | 23368/23368 |      0.00% |

## Checklist

- [x] API publica raiz disponible en `nexgou.go`.
- [x] Tests bajo `test`.
- [x] Samples compilan.
- [x] Taskboard valida SQLite real.
- [x] Benchmark NexGou produce JSON k6.
- [x] Informe Markdown inicial generado.
- [x] Documentacion actualizada para alcance HTTP-only.
- [x] `go.mod` ordenado.
- [x] Lint ejecutado.

## Pendientes antes de publicar tag remoto

- Ejecutar benchmark competitivo largo con Docker/observabilidad.
- Comparar contra competidores cuando sus implementaciones esten completas.
