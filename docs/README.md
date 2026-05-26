# NexGou

NexGou es un framework HTTP para Go inspirado en la experiencia de NestJS: una forma de construir APIs con modulos, inyeccion de dependencias, guards, interceptors, pipes, filtros de excepciones y una estructura pensada para aplicaciones mantenibles.

El objetivo no es crear solo otro router rapido. La meta es combinar una experiencia de desarrollo ordenada con un servidor competitivo en rendimiento, latencia y consumo de recursos.

## Estado actual

Este repositorio contiene la linea `2.0.0` del framework. El alcance HTTP inicial esta implementado y validado localmente: core, app lifecycle, DI, router, pipeline HTTP, logger, API publica raiz, samples HTTP y benchmark NexGou.

Ya existe una primera version funcional en:

```txt
F:\personal\server\nexgou\server - copia
```

Esa version valida la idea principal del producto: una API inspirada en NestJS para Go, con modulos, DI, ciclo de vida de aplicacion y componentes reutilizables. Sin embargo, todavia necesita mucho trabajo de performance antes de tratarla como base final.

La decision para esta nueva etapa es clara: continuar con la misma idea, pero apoyarse en `fasthttp` para construir un servidor de alto rendimiento sin rehacer desde cero la logica que ya funciona. WebSocket, SSE y gRPC quedan fuera del alcance inicial de `2.0.0` y podran evaluarse despues de estabilizar el core HTTP.

## Vision

NexGou debe sentirse como un framework de aplicacion, no como una libreria minima de rutas.

Queremos conservar estas piezas conceptuales:

- modulos para organizar la aplicacion por dominio;
- inyeccion de dependencias para evitar globals y wiring manual repetitivo;
- guards para autenticacion y autorizacion;
- interceptors para logica antes y despues del handler;
- pipes para validacion y transformacion de entrada;
- filtros de excepciones para normalizar errores;
- middlewares globales y por ruta;
- contexto HTTP ergonomico para leer params, body, query y responder JSON;
- soporte futuro para WebSocket, SSE y gRPC, tomando como referencia la primera version, fuera del alcance inicial de `2.0.0`.

La inspiracion de NestJS esta en la arquitectura y en la DX. La implementacion debe ser Go idiomatic y medible con benchmarks reales.

## Direccion tecnica

La nueva arquitectura debe separar el core del transporte HTTP.

```txt
core de aplicacion
	app lifecycle
	module system
	dependency injection
	handlers
	guards
	interceptors
	pipes
	filters
	middleware pipeline

adapters HTTP
	fasthttp   camino principal para performance
	net/http   compatibilidad futura o adaptador secundario
```

`fasthttp` debe ser el camino principal para exprimir throughput, reducir allocations y mejorar latencias p95/p99. Aun asi, la logica del framework no debe quedar acoplada de forma innecesaria al transporte. El core deberia poder evolucionar sin que cada feature dependa directamente de detalles internos de `fasthttp`.

## Reutilizacion de la version funcional

No queremos rehacer logica por costumbre. La carpeta `F:\personal\server\nexgou\server - copia` debe usarse como baseline funcional para migrar o adaptar las piezas que ya prueban la idea.

Piezas importantes a revisar y aprovechar:

- `nexgou.go`: API publica y reexports principales;
- `src/app/app.go`: orquestacion de aplicacion, `CreateApp`, `Use`, `Listen`;
- `src/core/container.go`: contenedor IoC e inyeccion por constructores;
- `src/common/context.go`: contexto HTTP, helpers de respuesta y lifecycle;
- `src/common/types.go`: contratos base del framework;
- `src/router/router.go`: router y pipeline de ejecucion;
- `src/common/exceptions.go`: excepciones HTTP y errores tipados;
- `src/filter`, `src/guard`, `src/interceptor`, `src/middleware`, `src/pipe`: componentes de arquitectura tipo NestJS;
- `samples`: ejemplos funcionales que deben servir como referencia de DX.

La regla de trabajo es: primero entender lo que ya existe, luego migrarlo o adaptarlo con mejor rendimiento.

## Benchmark como validacion

La mejora de performance no se debe declarar por intuicion. Debe validarse con un benchmark reproducible y un informe generico que permita comparar resultados entre servidores.

La estructura del laboratorio vive en:

```txt
benchmark/[type_server]
```

Implementaciones esperadas:

- `benchmark/nexgou`: implementacion del CRUD usando NexGou;
- `benchmark/fastify`: comparacion con Fastify;
- `benchmark/asp-kestrel`: comparacion con ASP.NET Core/Kestrel;
- `benchmark/actix-web`: comparacion con Actix Web;
- `benchmark/hyper`: comparacion Rust de bajo nivel;
- `benchmark/vert-x`: comparacion JVM;
- `benchmark/ajax-php`: comparacion PHP.

La documentacion completa del laboratorio esta en [docs/BENCHMARK.md](BENCHMARK.md).

El benchmark debe cubrir:

- `GET /health`;
- `POST /users`;
- `GET /users/:id`;
- `GET /users`;
- `PUT /users/:id`;
- `DELETE /users/:id`.

Todos los competidores deben usar el mismo contrato HTTP, el mismo payload JSON, SQLite, Docker, k6 y la misma politica de errores. El resultado final debe producir un resumen comparable en Markdown y JSON.

## Criterios de exito

NexGou sera competitivo cuando pueda demostrar, con datos reproducibles, que mantiene una buena DX sin sacrificar rendimiento.

Criterios minimos:

- error rate menor a 1% en CRUD mixto;
- p95 menor a 300 ms en CRUD mixto;
- p99 menor a 800 ms en CRUD mixto;
- sin crashes bajo carga;
- sin perdida de datos;
- sin memory leak evidente en soak test;
- rendimiento cercano o superior a Gin/Fastify en escenarios equivalentes.

Criterios deseables:

- acercarse a Fiber/fasthttp en escenarios read-heavy;
- superar a Gin en throughput manteniendo mejor arquitectura de aplicacion;
- consumir menos memoria que Fastify en cargas comparables;
- mantener p95/p99 estables bajo spike;
- generar reportes reproducibles para cada cambio importante del core o del adapter.

## Roadmap inicial

1. Revisar la primera version funcional en `F:\personal\server\nexgou\server - copia`.
2. Migrar el core minimo: app, context, types, container, router y pipeline.
3. Crear el adapter principal sobre `fasthttp`.
4. Mantener la API publica lo mas cercana posible a la version funcional cuando tenga sentido.
5. Implementar `benchmark/nexgou` con el CRUD comun sobre SQLite.
6. Ejecutar smoke test y CRUD mixto contra NexGou.
7. Comparar contra Fastify, ASP.NET Core/Kestrel y los demas servidores del laboratorio.
8. Generar un informe generico con throughput, p50, p95, p99, errores, CPU y memoria.
9. Optimizar solo cuando el benchmark muestre una mejora real en latencia, throughput o allocations.

## Documentacion relacionada

- [docs/BENCHMARK.md](BENCHMARK.md): laboratorio reproducible con k6, Docker, SQLite, Prometheus y Grafana.
- [docs/COMPETETION_REFERENCE.md](COMPETETION_REFERENCE.md): referencias competitivas y criterios de arquitectura para el framework.
- [docs/API.md](API.md): API publica HTTP de `2.0.0`.
- [docs/RELEASE_2.0.0.md](RELEASE_2.0.0.md): checklist de release readiness.

## Principio de trabajo

NexGou debe avanzar con una regla sencilla: reutilizar lo que ya valida la idea, medir cada cambio importante y dejar que el benchmark decida si una optimizacion merece quedarse.
