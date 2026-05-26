# Migracion a NexGou 2.0.0

Este documento define la migracion desde la primera version funcional de NexGou hacia este proyecto, que sera la nueva version `2.0.0`.

Origen:

```txt
F:\personal\server\nexgou\server - copia
```

Destino:

```txt
F:\personal\server\nexgou\server
```

El objetivo es conservar la idea que ya funciona: un framework HTTP para Go inspirado en NestJS, con modulos, inyeccion de dependencias, guards, interceptors, pipes y filtros de excepciones. La diferencia principal de `2.0.0` es que la nueva base debe estar orientada a performance, usando `fasthttp` como transporte principal sin acoplar todo el core a ese detalle.

La regla central de la migracion es simple: reutilizar la logica existente antes de reescribirla.

## Alcance

La migracion no termina al mover el core. Tambien deben migrarse tests, samples, benchmark y documentacion minima.

Incluido en el alcance:

- codigo del framework desde `src/`;
- tests unitarios e integracion;
- samples funcionales;
- implementacion `benchmark/nexgou`;
- documentacion de uso y validacion;
- criterios de release para `2.0.0`.

Fuera del alcance inicial:

- reescrituras completas sin motivo medible;
- cambios grandes de API publica que no aporten DX o performance;
- migracion de WebSocket, SSE o gRPC en `2.0.0`;
- migracion de samples que dependan de WebSocket, SSE o gRPC;
- optimizaciones que no pasen por benchmark;
- declarar `2.0.0` listo sin samples y benchmark funcionales.

## Inventario del origen

La version funcional vive en `F:\personal\server\nexgou\server - copia`. Estas piezas deben revisarse antes de implementar su equivalente en `2.0.0`.

| Area          | Origen                                        | Motivo                                              |
| ------------- | --------------------------------------------- | --------------------------------------------------- |
| API publica   | `nexgou.go`                                   | Reexports y superficie publica del framework        |
| App lifecycle | `src/app/app.go`                              | Bootstrap, `CreateApp`, `Use`, `Listen`             |
| Tipos base    | `src/common/types.go`                         | Contratos de handlers, rutas, modulos y componentes |
| Contexto HTTP | `src/common/context.go`                       | Params, body, query, response helpers y lifecycle   |
| Errores       | `src/common/exceptions.go`                    | Excepciones HTTP y normalizacion de errores         |
| Banner        | `src/common/banner.go`                        | Salida de arranque y metadata                       |
| DI container  | `src/core/container.go`                       | Registro y resolucion de dependencias               |
| Module system | `src/core/module.go`, `src/module`            | Organizacion por modulos                            |
| Router        | `src/router/router.go`                        | Matching, versionado y pipeline base                |
| Config        | `src/config/config.go`                        | Variables de entorno e inyeccion de configuracion   |
| Logger        | `src/logger/logger.go`                        | Logging estructurado                                |
| Middleware    | `src/middleware/*`                            | CORS, security, rate limit, timeout y body size     |
| Guards        | `src/guard/guard.go`                          | Autenticacion/autorizacion pre-handler              |
| Interceptors  | `src/interceptor/*`                           | Logica antes/despues del handler                    |
| Pipes         | `src/pipe/pipe.go`                            | Validacion y transformacion                         |
| Filters       | `src/filter/filter.go`                        | Captura y respuesta de errores                      |
| Tests         | `test/common`, `test/core`, `test/nexgoutest` | Baseline de correctness                             |
| Samples       | `samples/*`                                   | Validacion real de DX                               |

## Inventario del destino

El proyecto `2.0.0` debe organizarse como una nueva base, pero tomando decisiones con el codigo funcional delante.

| Area                 | Destino                         | Estado esperado                         |
| -------------------- | ------------------------------- | --------------------------------------- |
| API publica          | `nexgou.go`                     | Reexports publicos de la version 2.0.0  |
| Core                 | `src/*`                         | Implementacion nueva, migrada por fases |
| Tests                | `test/*`                        | Tests migrados junto con cada modulo    |
| Samples              | `samples/*`                     | Migracion obligatoria, no opcional      |
| Benchmark NexGou     | `benchmark/nexgou`              | CRUD comun para validar performance     |
| Benchmark compartido | `benchmark/_shared`             | k6, reportes, observabilidad y scripts  |
| Docs                 | `docs/README.md`                | Vision del producto                     |
| Docs                 | `docs/BENCHMARK.md`             | Laboratorio reproducible                |
| Docs                 | `docs/COMPETETION_REFERENCE.md` | Referencias competitivas                |

## Principios de migracion

- Mantener la DX inspirada en NestJS.
- Separar core de transporte HTTP.
- Usar `fasthttp` como adapter principal de alto rendimiento.
- Evitar acoplar guards, interceptors, pipes y filters a detalles internos del transporte.
- Migrar tests junto con el codigo, no al final.
- Crear siempre los tests en la carpeta raiz `test`, separados por modulo o feature.
- Migrar samples como validacion de uso real.
- Medir antes y despues de cada optimizacion importante.
- No aceptar cambios de performance sin mejora en benchmarks.

WebSocket, SSE y gRPC quedan fuera de la migracion inicial de `2.0.0`. Podran evaluarse en una version posterior cuando el core HTTP, los samples HTTP y el benchmark esten estables.

## [x] Fase 0: baseline y preparacion

Objetivo: congelar el punto de partida antes de tocar la nueva version.

Estado: completada. La carpeta origen existe, la version funcional ejecuta sus tests con Go `1.26.3` y se registro una baseline del router antes de continuar la migracion.

Archivos origen:

- `F:\personal\server\nexgou\server - copia\go.mod`;
- `F:\personal\server\nexgou\server - copia\Makefile`;
- `F:\personal\server\nexgou\server - copia\.golangci.yml`;
- `F:\personal\server\nexgou\server - copia\test`;
- `F:\personal\server\nexgou\server - copia\samples`.

Archivos destino:

- `go.mod`;
- `Makefile`;
- `.golangci.yml`;
- `test`;
- `samples`.

Tareas:

- Confirmar que la carpeta origen esta respaldada.
- Registrar version de Go usada por el origen.
- Ejecutar tests de la version funcional si el entorno lo permite.
- Registrar metricas disponibles del router y pipeline existentes.
- Definir dependencia inicial de `fasthttp` para la nueva version.
- Confirmar que `2.0.0` sera una nueva linea y no una copia directa.

Pruebas requeridas:

- Tests del origen ejecutados o documentados como pendiente si no pueden correr.
- Baseline de benchmarks existentes del router si estan disponibles.

Gate de salida:

- Existe una referencia clara de que se va a migrar y que no se va a reescribir sin medir.

Resultado baseline:

```txt
Origen: F:\personal\server\nexgou\server - copia
Go: go1.26.3 windows/amd64
Tests origen: go test ./test/... OK
Benchmark origen: go test ./src/router -run ^$ -bench . -benchmem OK
```

Baseline router origen:

| Benchmark                         | ns/op | B/op | allocs/op |
| --------------------------------- | ----: | ---: | --------: |
| StaticRoute                       | 299.4 |  192 |         5 |
| ParamRoute                        | 517.4 |  544 |         7 |
| NotFound                          | 902.5 | 1088 |        11 |
| WithPipeline                      | 583.1 |  544 |         7 |
| HighCardinalityStatic/routes_100  |  1804 |  208 |         5 |
| HighCardinalityStatic/routes_500  |  7412 |  208 |         5 |
| HighCardinalityStatic/routes_1000 | 14921 |  208 |         5 |

Riesgos:

- Perder comportamiento funcional por empezar desde una estructura vacia.
- Optimizar antes de tener baseline.

## [x] Fase 1: common, tipos y contexto

Objetivo: migrar la base comun que todos los modulos usan.

Estado: completada. El codigo quedo en `src/common` y los tests quedaron en `test/common`.

Archivos origen:

- `src/common/types.go`;
- `src/common/context.go`;
- `src/common/exceptions.go`;
- `src/common/banner.go`.

Archivos destino:

- `src/common/types.go`;
- `src/common/context.go`;
- `src/common/exceptions.go`;
- `src/common/banner.go`.

Tareas:

- Migrar `H`, `HandlerFunc`, contratos de rutas y contratos de componentes.
- Definir un contexto estable para `2.0.0`.
- Separar lo que es API de aplicacion de lo que es detalle de `fasthttp`.
- Mantener helpers ergonomicos para JSON, status, params, query y body.
- Migrar excepciones HTTP y helpers de errores.

Pruebas requeridas:

- Tests unitarios para tipos publicos en `test/common`.
- Tests de contexto en `test/common`: params, query, body, JSON y error response.

Gate de salida:

- El contexto y los tipos base son suficientemente estables para que router, middleware y samples puedan compilar encima.

Riesgos:

- Si `Context` cambia demasiado tarde, todos los handlers y samples se rompen.
- Si `Context` queda acoplado a `fasthttp`, sera dificil mantener adapters futuros.

## [x] Fase 2: DI, modulos, app lifecycle y config

Objetivo: recuperar la arquitectura de aplicacion tipo NestJS.

Estado: completada. El codigo quedo en `src/core`, `src/module`, `src/app` y `src/config`; los tests quedaron en `test/core`, `test/app` y `test/config`.

Archivos origen:

- `src/core/container.go`;
- `src/core/module.go`;
- `src/module`;
- `src/app/app.go`;
- `src/config/config.go`.

Archivos destino:

- `src/core/container.go`;
- `src/core/module.go`;
- `src/module`;
- `src/app/app.go`;
- `src/config/config.go`.

Tareas:

- Migrar el contenedor IoC con tests antes de usarlo en otros modulos.
- Definir la forma final de `Module` para `2.0.0`.
- Migrar `CreateApp`, registro de modulos, middlewares y rutas.
- Preparar lifecycle hooks si ya existen o si son necesarios para los samples.
- Migrar `ConfigService` con lectura de entorno tipada.

Pruebas requeridas:

- Tests de registro y resolucion de dependencias.
- Tests de errores del container.
- Tests de bootstrap minimo de una app.

Gate de salida:

- Una aplicacion minima puede crearse, registrar providers y resolver dependencias.

Riesgos:

- El container es una pieza critica: si falla, todos los modulos fallan.
- Duplicar conceptos entre `src/core/module.go` y `src/module` puede ensuciar la API.

## [x] Fase 3: router y adapter fasthttp

Objetivo: tener un servidor HTTP funcional sobre `fasthttp` sin sacrificar la arquitectura del core.

Estado: completada. El codigo quedo en `src/router`, `src/adapters/fasthttp` y la integracion de `src/app`; los tests quedaron en `test/router`, `test/adapters/fasthttp` y `test/app`.

Archivos origen:

- `src/router/router.go`;
- `src/router/router_bench_test.go`;
- `src/app/app.go`;
- `src/common/context.go`.

Archivos destino:

- `src/router/router.go`;
- `src/router/router_bench_test.go`;
- `src/app/app.go`;
- `src/common/context.go`;
- `src/adapters/fasthttp` si se decide separar adapters fisicamente.

Tareas:

- Migrar matching de rutas estaticas y dinamicas.
- Mantener versionado si el origen lo soporta.
- Crear adaptador `fasthttp` como transporte principal.
- Evitar que handlers de usuario dependan directamente de `fasthttp.RequestCtx`.
- Definir como se registran metodos HTTP: GET, POST, PUT, DELETE, PATCH.
- Migrar benchmarks del router y ampliar con casos de contexto/pipeline.

Pruebas requeridas:

- Tests de matching estatico, dinamico, not found y method not allowed.
- Tests de params.
- Benchmarks de router, context y ruta simple.

Gate de salida:

- Un endpoint `GET /health` funciona sobre `fasthttp` usando la API publica de NexGou.

Riesgos:

- Acoplar todo a `fasthttp` demasiado pronto.
- Mejorar throughput pero empeorar DX o testabilidad.

## [x] Fase 4: pipeline HTTP

Objetivo: migrar el flujo completo de ejecucion alrededor del handler.

Estado: completada. El pipeline HTTP quedo separado en paquetes publicos para middleware, guards, pipes, interceptors y filters; el router precompila la cadena por ruta y mantiene el orden `middleware -> guard -> pipe -> interceptor -> handler -> filter on error`.

Archivos origen:

- `src/middleware/*`;
- `src/guard/guard.go`;
- `src/interceptor/*`;
- `src/pipe/pipe.go`;
- `src/filter/filter.go`.

Archivos destino:

- `src/middleware/*`;
- `src/guard/guard.go`;
- `src/interceptor/*`;
- `src/pipe/pipe.go`;
- `src/filter/filter.go`.

Tareas:

- [x] Definir orden de ejecucion del pipeline.
- [x] Migrar CORS, security headers, body limit, timeout y rate limit.
- [x] Migrar guards de ruta.
- [x] Migrar interceptors before/after.
- [x] Migrar pipes de validacion y transformacion.
- [x] Migrar filtros de excepciones.
- [x] Precompilar pipeline por ruta cuando sea posible para reducir allocations.

Orden recomendado:

```txt
middleware global
guard
pipe
interceptor before
handler
interceptor after
filter on error
```

Pruebas requeridas:

- [x] Tests de orden de ejecucion.
- [x] Tests de errores y filtros.
- [x] Tests de headers CORS/security.
- [x] Tests de timeout y body limit.
- [x] Benchmarks de ruta con pipeline completo.

Validacion:

```txt
go test ./test/... OK
go test ./test/router -run ^$ -bench FullPipeline -benchmem OK
BenchmarkRouterFullPipeline-16  1238223  962.0 ns/op  916 B/op  12 allocs/op
```

Gate de salida:

- Un sample REST pequeno puede usar middleware, guard, pipe, interceptor y filter en la misma ruta.

Riesgos:

- Orden incorrecto del pipeline.
- Race conditions en rate limiter.
- Aumentar allocations por closures o wrappers creados por request.

## [x] Fase 5: logger y observabilidad minima

Objetivo: recuperar logging util sin convertirlo en cuello de botella.

Estado: completada. El logger estructurado quedo en `src/logger`, soporta salida text/JSON, niveles via `LOG_LEVEL`, formato via `LOG_FORMAT`, writer inyectable para tests y modulo DI integrado con `src/config` y `src/app`.

Archivos origen:

- `src/logger/logger.go`;
- usos de logger en `samples`.

Archivos destino:

- `src/logger/logger.go`;
- integraciones con `src/config` y `src/app`.

Tareas:

- [x] Migrar logger estructurado.
- [x] Soportar salida text y JSON.
- [x] Definir `LOG_LEVEL` y `LOG_FORMAT`.
- [x] Evitar logging sincronico costoso en hot paths.
- [x] Preparar hooks para metricas del benchmark si aplica.

Pruebas requeridas:

- [x] Tests de formato.
- [x] Tests de niveles.
- [x] Benchmark basico de logging desactivado y logging JSON.

Validacion:

```txt
go test ./test/... OK
go test ./test/logger -run ^$ -bench . -benchmem OK
BenchmarkLoggerDisabled-16  282264144  4.303 ns/op  0 B/op    0 allocs/op
BenchmarkLoggerJSON-16         591939  1928 ns/op   793 B/op  19 allocs/op
```

Gate de salida:

- La app puede arrancar y emitir logs utiles sin degradar rutas simples de forma significativa.

Riesgos:

- `json.Marshal` por request puede costar demasiado.
- Logs excesivos pueden invalidar benchmarks.

## [x] Fase 5.1: API publica raiz

Objetivo: crear `nexgou.go` como punto de entrada publico de NexGou `2.0.0`.

Estado: completada. La raiz publica `nexgou.go` expone la API HTTP/core de `2.0.0`: app, modulos, rutas, excepciones, config, logger, middleware, filters y pipes. WebSocket, SSE y gRPC no se reexportan porque estan fuera del alcance de esta migracion inicial.

Archivos origen:

- `F:\personal\server\nexgou\server - copia\nexgou.go`.

Archivos destino:

- `nexgou.go`;
- tests bajo `test`, en la carpeta que corresponda a la API publica.

Tareas:

- [x] Revisar los reexports publicos de la version funcional.
- [x] Exponer solo la API publica necesaria para `2.0.0`.
- [x] Mantener nombres compatibles cuando tenga sentido para no romper la DX existente.
- [x] Evitar exportar detalles internos del adapter `fasthttp`.
- [x] Agregar tests en la carpeta raiz `test` para validar que los tipos y funciones publicas compilan y son usables.

Pruebas requeridas:

- [x] Tests de compilacion y uso de `nexgou.go` desde un paquete externo.
- [x] Tests de smoke para helpers publicos cuando existan.

Validacion:

```txt
go test ./test/... OK
go test ./... OK
```

Gate de salida:

- Un usuario puede importar `github.com/nexgou/server` y construir una app minima usando la API publica de `2.0.0`.

Riesgos:

- Reexportar demasiado pronto APIs internas que todavia pueden cambiar.
- Acoplar la API publica a `fasthttp` en vez de mantener el core abstraido.

## [x] Fase 6: migracion de samples HTTP

Objetivo: validar que `2.0.0` sigue siendo usable, no solo rapido.

Estado: completada. `samples/api` y `samples/taskboard` fueron migrados a la API publica HTTP-only de `2.0.0`; los samples realtime siguen fuera de alcance. `samples/taskboard` usa SQLite real mediante `database/sql` y `modernc.org/sqlite`.

Los samples HTTP son parte obligatoria del release. Si estos samples no migran, la migracion no esta completa.

| Sample    | Origen              | Destino             | Depende de                                                           | Validacion                              |
| --------- | ------------------- | ------------------- | -------------------------------------------------------------------- | --------------------------------------- |
| API REST  | `samples/api`       | `samples/api`       | app, router, modules, config, logger, middleware, guard, interceptor | [x] Arranca y expone rutas REST         |
| Taskboard | `samples/taskboard` | `samples/taskboard` | auth, SQLite, filters, modules, REST                                 | [x] CRUD real funciona con persistencia |

Tareas:

- [x] Migrar primero `samples/api`, porque valida la mayor parte del pipeline HTTP.
- [x] Migrar `samples/taskboard` despues, porque combina auth, SQLite, modules, filters y CRUD real.
- [x] Actualizar README de cada sample.
- [x] Mantener los samples pequenos, ejecutables y utiles para validar DX.

Pruebas requeridas:

- [x] Cada sample debe compilar.
- [x] Cada sample debe tener un smoke test o comando documentado.
- [x] `samples/taskboard` debe validar persistencia SQLite.

Validacion:

```txt
go test ./test/... OK
go test ./... OK
```

Gate de salida:

- Los samples HTTP incluidos en `2.0.0` compilan, arrancan y documentan como probarlos.

Riesgos:

- Migrar samples al final puede ocultar cambios malos de API.
- `samples/taskboard` puede introducir dependencias de DB antes de tiempo.

Samples fuera del alcance de `2.0.0`:

- `samples/chat`;
- `samples/sse`;
- `samples/grpc`.

## [x] Fase 7: benchmark NexGou e informe

Objetivo: validar `2.0.0` con datos reproducibles.

Estado: completada para NexGou. `benchmark/nexgou` implementa el contrato CRUD comun con SQLite, `benchmark/_shared/k6` contiene smoke y CRUD mixto, los resultados se exportan a JSON y existe un informe Markdown inicial. La comparacion contra otros competidores queda preparada para corridas posteriores cuando sus implementaciones esten completas.

Archivos destino:

- `benchmark/nexgou`;
- `benchmark/_shared/k6`;
- `benchmark/_shared/reports`;
- `docs/BENCHMARK.md`.

Tareas:

- [x] Implementar el CRUD comun en `benchmark/nexgou`.
- [x] Usar el mismo contrato HTTP que los demas competidores.
- [x] Usar SQLite con el mismo schema e indices.
- [x] Ejecutar smoke test.
- [x] Ejecutar CRUD mixto.
- [x] Preparar comparacion contra `benchmark/fastify`, `benchmark/asp-kestrel`, `benchmark/actix-web`, `benchmark/hyper`, `benchmark/vert-x` y `benchmark/ajax-php` cuando esten disponibles.
- [x] Generar informe generico en Markdown y JSON.

Metricas obligatorias:

- throughput;
- p50;
- p95;
- p99;
- error rate;
- checks correctos;
- CPU;
- memoria;
- estabilidad bajo carga.

Pruebas requeridas:

- [x] `GET /health` responde correctamente.
- [x] CRUD completo pasa con k6.
- [x] Resultados se exportan en JSON.
- [x] Informe Markdown puede rellenarse con los resultados.

Validacion:

```txt
go test ./test/benchmark ./benchmark/nexgou/... OK
go test ./test/... OK
go test ./... OK
k6 smoke OK, checks 12/12, error rate 0.00%, JSON exportado
k6 crud-mixed OK, checks 23368/23368, error rate 0.00%, JSON exportado
```

Gate de salida:

- `benchmark/nexgou` produce resultados comparables y documentados.

Riesgos:

- Comparar endpoints vacios contra CRUD real.
- Dejar logs activos en un competidor y no en otro.
- Medir performance antes de estabilizar correctness.

## [x] Fase 8: documentacion, cleanup y release readiness

Objetivo: preparar `2.0.0` como linea usable.

Estado: completada. La documentacion publica fue alineada con el alcance HTTP-only de `2.0.0`, se agregaron docs de API publica, release readiness y changelog, y se revalidaron tests, samples y benchmark.

Archivos destino:

- `docs/README.md`;
- `docs/MIGRATION.md`;
- `docs/BENCHMARK.md`;
- docs de modulos;
- README de samples;
- `CHANGELOG.md` si aplica.

Tareas:

- [x] Actualizar documentacion de API publica.
- [x] Documentar diferencias entre la version funcional anterior y `2.0.0`.
- [x] Documentar comandos para tests, samples y benchmark.
- [x] Revisar nombres de paquetes y exports.
- [x] Limpiar codigo muerto o duplicado.
- [x] Confirmar version `2.0.0`.
- [x] Preparar checklist de release.

Pruebas requeridas:

- [x] Tests unitarios en verde.
- [x] Tests de integracion en verde.
- [x] Samples compilando.
- [x] Benchmark smoke completado.
- [x] Lint ejecutado o documentado como pendiente.

Validacion:

```txt
go test ./test/... OK
go test ./... OK
go test ./test/router -run ^$ -bench FullPipeline -benchmem OK
go test ./test/logger -run ^$ -bench . -benchmem OK
golangci-lint run ./... OK
k6 smoke OK
k6 crud-mixed OK
```

Gate de salida:

- La version `2.0.0` puede instalarse, ejecutar samples y validarse con benchmark.

Riesgos:

- Publicar una version que solo compila, pero no demuestra DX ni performance.
- Dejar documentacion desalineada con la API real.

## [ ] Fase 9: evaluacion de Sonic para JSON

Objetivo: medir `github.com/bytedance/sonic` antes de reemplazar `encoding/json` en rutas productivas.

Estado: iniciada. Se subio el requisito del modulo a Go `1.25.0`, se agrego Sonic como dependencia directa y se crearon benchmarks comparativos sin cambiar `Context.JSON`, `Context.Body` ni el logger productivo.

Archivos destino:

- `go.mod`;
- `go.sum`;
- `test/common/json_bench_test.go`;
- `test/logger/logger_bench_test.go`.

Tareas:

- [x] Subir el requisito del modulo a Go `1.25.x`.
- [x] Agregar Sonic para benchmarks.
- [x] Medir encode/decode JSON representativo de `Context`.
- [x] Medir marshal JSON del logger.
- [ ] Decidir si se integra Sonic en codigo productivo.

Validacion focalizada:

```txt
go test ./test/common ./test/logger OK
go test ./test/adapters/fasthttp OK
go test ./test/common -run ^$ -bench . -benchmem -count=3 OK
go test ./test/logger -run ^$ -bench . -benchmem -count=3 OK
```

Resultados iniciales relevantes:

| Benchmark                | Variante      | ns/op aprox. | B/op | allocs/op |
| ------------------------ | ------------- | -----------: | ---: | --------: |
| JSON encode payload CRUD | stdlib        |         2800 |  368 |         8 |
| JSON encode payload CRUD | sonic_std     |    2700-3200 |  940 |         6 |
| JSON encode payload CRUD | sonic_default |          900 |  335 |         5 |
| JSON decode payload CRUD | stdlib        |   7900-10300 | 2080 |        45 |
| JSON decode payload CRUD | sonic_std     |    2600-4200 | 1430 |        30 |
| JSON decode payload CRUD | sonic_default |         2100 | 1250 |        14 |
| Logger JSON marshal      | stdlib        |    1460-1580 |  544 |        14 |
| Logger JSON marshal      | sonic_std     |      690-740 |  228 |         3 |
| Logger JSON marshal      | sonic_default |      530-550 |  228 |         3 |

Observaciones:

- `sonic_default` muestra mejora clara en encode, decode y marshal del logger.
- `sonic_std` mejora mucho decode y marshal, pero no gana claramente en encode del payload CRUD.
- La suite amplia `go test ./test/...` no esta verde por fallos existentes fuera de esta evaluacion: incompatibilidades en router/core (`Pipe`, `RouteInfo.Pipes`, `repository`) y paquetes en migracion. Estos fallos deben resolverse antes de integrar Sonic en produccion.

Gate de salida:

- No cambiar el motor JSON productivo hasta que la suite base compile y se repitan benchmarks con `-count=5` sobre una base estable.

Riesgos:

- `sonic_default` no es completamente equivalente a `encoding/json` en opciones como HTML escaping y orden de keys.
- Adoptar Sonic antes de estabilizar router/core puede mezclar una optimizacion real con fallos de migracion no relacionados.

## Estrategia de pruebas

Cada fase debe incluir pruebas en la misma entrega.

Todos los tests del proyecto deben vivir bajo la carpeta raiz `test`. No se deben crear tests dentro de `src` aunque prueben paquetes internos; cada modulo debe tener su carpeta equivalente bajo `test`, por ejemplo `test/common`, `test/core` o `test/router`.

Prioridad:

1. Tests unitarios del modulo migrado.
2. Tests de integracion del pipeline.
3. Tests de samples.
4. Benchmarks internos de router, context y pipeline.
5. Benchmark competitivo con k6.

Reglas:

- No mover la siguiente fase si el modulo base no tiene tests minimos.
- No crear tests fuera de la carpeta raiz `test`.
- No optimizar hot paths sin benchmark antes/despues.
- No declarar estable una API hasta usarla en al menos un sample.
- No declarar `2.0.0` listo sin ejecutar `benchmark/nexgou`.

## Criterios de no regresion

La migracion debe proteger lo que hizo valiosa la primera version.

- La API publica debe mantenerse compatible cuando sea razonable.
- La DX tipo NestJS no debe perderse por perseguir throughput.
- Guards, interceptors, pipes y filters deben seguir siendo conceptos de primera clase.
- Los samples deben seguir siendo claros y ejecutables.
- Los errores deben seguir siendo normalizados.
- El nuevo adapter `fasthttp` debe mejorar performance sin convertir el core en codigo dificil de testear.
- Toda mejora importante debe pasar por medicion.

## Checklist final para 2.0.0

- [x] `src/common` migrado y probado.
- [x] `src/core` migrado y probado.
- [x] `src/app` migrado y probado.
- [x] `src/router` migrado y probado.
- [x] Adapter `fasthttp` funcional.
- [x] Pipeline HTTP migrado: middleware, guard, interceptor, pipe y filter.
- [x] Config migrado y probado.
- [x] Logger migrado y probado.
- [x] `nexgou.go` agregado y probado como API publica raiz.
- [x] `samples/api` migrado.
- [x] `samples/taskboard` migrado.
- [x] `benchmark/nexgou` implementado.
- [x] Smoke benchmark ejecutado.
- [x] CRUD mixed benchmark ejecutado.
- [x] Informe generico generado.
- [x] Documentacion actualizada.
- [x] Version `2.0.0` lista como release candidate local.

## Referencias

- [README](README.md): vision del framework y direccion tecnica.
- [BENCHMARK](BENCHMARK.md): laboratorio reproducible para validar performance.
- [COMPETETION_REFERENCE](COMPETETION_REFERENCE.md): referencias competitivas y decisiones de diseno.
