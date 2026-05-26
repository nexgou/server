# Integraciones y librerias externas

Este documento recoge librerias que pueden aportar valor real a Nexgou. La regla es deliberadamente estricta: una dependencia entra si mejora rendimiento medible, reduce complejidad propia, abre una capacidad de produccion dificil de mantener internamente o mejora la validacion del framework.

No se recomienda agregar dependencias solo por popularidad. Cada integracion debe tener un punto de entrada claro, tests, documentacion y un criterio de salida medible.

## Criterios de adopcion

Una libreria candidata debe cumplir al menos una de estas condiciones:

- Reduce `ns/op`, `B/op` o `allocs/op` en hot paths relevantes.
- Evita implementar seguridad, criptografia, protocolos o observabilidad a mano.
- Encaja como modulo opcional sin contaminar el core ni la API publica principal.
- Mejora tests de integracion o reproducibilidad de benchmarks.
- Tiene mantenimiento activo, versionado claro y soporte para Go `1.25+`.

Antes de integrarla en codigo productivo:

1. Crear un benchmark o test de contrato.
2. Comparar contra la implementacion actual.
3. Documentar tradeoffs.
4. Mantener la API publica de Nexgou estable.
5. Evitar dependencias globales en el core cuando puedan vivir en modulos opcionales.

## Resumen ejecutivo

| Prioridad  | Libreria                                      | Area                    | Decision recomendada                                                                       |
| ---------- | --------------------------------------------- | ----------------------- | ------------------------------------------------------------------------------------------ |
| Alta       | `github.com/bytedance/sonic`                  | JSON                    | Seguir evaluacion; integrar solo si la suite base compila y `-count=5` confirma la mejora. |
| Alta       | `go.opentelemetry.io/otel`                    | Observabilidad          | Crear modulo opcional de tracing/metrics. Es la base mas util para produccion.             |
| Alta       | `github.com/go-playground/validator/v10`      | Validacion              | Mantener/adoptar como motor de pipes de validacion.                                        |
| Alta       | `github.com/golang-jwt/jwt/v5`                | Auth                    | Mantener/adoptar para JWT con reglas estrictas de algoritmo y claims.                      |
| Media-alta | `github.com/rs/zerolog`                       | Logging                 | Evaluar como backend del logger si supera al logger actual + Sonic.                        |
| Media      | `go.uber.org/zap`                             | Logging                 | Alternativa madura a zerolog; no usar ambas.                                               |
| Media      | `github.com/prometheus/client_golang`         | Metricas                | Usar si se quiere `/metrics` Prometheus directo; si no, preferir OTel.                     |
| Media      | `github.com/pressly/goose/v3`                 | Migraciones DB          | Adoptar para samples, benchmarks y modulo SQL opcional.                                    |
| Media      | `github.com/sqlc-dev/sqlc`                    | SQL tipado              | Usar en samples/benchmark si las consultas crecen. No es dependencia runtime obligatoria.  |
| Media      | `github.com/jackc/pgx/v5`                     | PostgreSQL              | Modulo opcional de Postgres; no meter en core HTTP.                                        |
| Media      | `github.com/redis/go-redis/v9`                | Cache/pubsub/rate-limit | Modulo opcional para cache y rate limit distribuido.                                       |
| Media      | `github.com/testcontainers/testcontainers-go` | Tests                   | Usar en tests de integracion para Postgres, Redis y servicios externos.                    |
| Baja       | ORMs completos tipo GORM                      | Persistencia            | Evitar como core; demasiada opinion y overhead para el objetivo del framework.             |
| Baja       | Routers como Gin/Chi/Echo/Fiber               | HTTP                    | No aportan al core porque Nexgou ya tiene router y adapter propio.                         |

## Candidatas de alto valor

### Sonic

Paquete: `github.com/bytedance/sonic`

Encaje en Nexgou:

- `Context.JSON` para respuestas JSON.
- `Context.Body` para decode de request body.
- respuestas JSON de error del router.
- logging JSON si el logger sigue usando mapas o structs intermedios.

Estado actual:

- Ya existe evaluacion inicial en benchmarks.
- `sonic_default` mostro mejora clara en encode/decode y marshal de logger.
- `sonic_std` mejora decode y marshal, pero encode puede quedar similar o peor que stdlib para payloads pequenos.

Gate de adopcion:

- `go test ./test/...` debe compilar antes de reemplazar codigo productivo.
- Repetir benchmarks con `-count=5`.
- Agregar tests de compatibilidad para HTML escaping, numeros, errores de decode, `json.Marshaler`, `json.RawMessage` y newline final de `Encode`.

Decision recomendada:

Mantener en evaluacion. Si se adopta, hacerlo detras de un wrapper interno, no como API publica nueva.

### OpenTelemetry

Paquetes base: `go.opentelemetry.io/otel`, `go.opentelemetry.io/otel/sdk`, exportadores OTLP/Prometheus segun necesidad.

Encaje en Nexgou:

- Middleware/interceptor para spans por request HTTP.
- Propagacion de trace context.
- Metricas de latencia, estado HTTP, errores y throughput.
- Integracion futura con gRPC y Redis/Postgres.

Por que aporta:

- Traces y metrics son estables en OpenTelemetry Go.
- Evita inventar APIs propias de observabilidad.
- Conecta con Grafana, Prometheus, Jaeger, Tempo, Datadog, New Relic y backends OTLP.

Riesgos:

- Puede aumentar complejidad si entra directo al core.
- Los exporters deben ser opcionales.

Decision recomendada:

Crear `src/module/otel` o `src/observability` como modulo opcional. El core solo deberia exponer hooks o interceptors neutros.

### go-playground/validator

Paquete: `github.com/go-playground/validator/v10`

Encaje en Nexgou:

- Motor por defecto de `ValidationPipe`.
- Validacion de DTOs con tags.
- Traduccion de errores a respuestas `400` consistentes.

Por que aporta:

- Valida structs, campos cruzados, slices, maps y tipos custom.
- Tiene validadores comunes: `required`, `email`, `uuid`, `url`, `oneof`, `min`, `max`, `datetime`, `jwt`, `cron`.
- Permite extraer nombres JSON para errores legibles.

Gate de adopcion:

- Tests de errores de validacion con payloads reales.
- No filtrar nombres internos de structs en respuestas publicas.
- Inicializar con `validator.WithRequiredStructEnabled()` para comportamiento futuro-compatible.

Decision recomendada:

Adoptar como motor del pipe de validacion. Es una dependencia que ahorra mucho codigo propio y encaja con la DX tipo NestJS.

### golang-jwt/jwt

Paquete: `github.com/golang-jwt/jwt/v5`

Encaje en Nexgou:

- `JwtModule` opcional.
- `JwtGuard` para Bearer auth.
- Helpers de firma/verificacion con claims tipados.

Por que aporta:

- Implementa parsing, verificacion y firma JWT con soporte HMAC, RSA, RSA-PSS, ECDSA y Ed25519.
- Es mantenida y estable.
- Evita implementar detalles criptograficos a mano.

Gate de adopcion:

- Exigir algoritmo esperado, no confiar solo en el header del token.
- Usar `RegisteredClaims` y validar `exp`, `nbf`, `iss`, `aud` cuando aplique.
- Tests para token expirado, algoritmo incorrecto, firma invalida y claims incompletos.

Decision recomendada:

Adoptar en un modulo auth opcional, no en el core HTTP.

## Observabilidad y logging

### Prometheus client_golang

Paquete: `github.com/prometheus/client_golang`

Encaje en Nexgou:

- Endpoint `/metrics` opcional.
- Contadores por metodo/ruta/status.
- Histogramas de latencia.

Por que aporta:

- Es la libreria oficial de instrumentacion Prometheus para Go.
- Encaja con el laboratorio existente de k6, Prometheus y Grafana.

Riesgos:

- Si tambien se adopta OpenTelemetry, puede duplicar metricas.

Decision recomendada:

Si el objetivo es Prometheus directo, usarla. Si el objetivo es vendor-neutral, preferir OpenTelemetry y exportador Prometheus.

### zerolog

Paquete: `github.com/rs/zerolog`

Encaje en Nexgou:

- Backend opcional para `LoggerService`.
- Logs JSON de baja asignacion.
- Sampling y context fields.

Por que aporta:

- API pensada para logging JSON con muy pocas asignaciones.
- Tiene integracion con `context.Context` y `log/slog`.

Gate de adopcion:

- Comparar contra el logger actual y contra logger actual + Sonic.
- Mantener la API publica de `LoggerService` si ya esta documentada.

Decision recomendada:

Evaluar. Si gana claramente, usar como backend interno. No exponer tipos de zerolog salvo en un adaptador avanzado.

### zap

Paquete: `go.uber.org/zap`

Encaje en Nexgou:

- Backend alternativo de logger estructurado.
- Buen soporte para produccion y API estable.

Por que aporta:

- Logger estructurado, estable, muy usado y con encoder JSON sin reflexion para campos tipados.

Decision recomendada:

Comparar con zerolog y elegir uno. No conviene mantener dos backends de logging en el framework base.

## Datos y persistencia

### goose

Paquete: `github.com/pressly/goose/v3`

Encaje en Nexgou:

- Migraciones SQL para `samples/taskboard`.
- Migraciones del benchmark CRUD.
- Futuro modulo `DatabaseModule`.

Por que aporta:

- CLI y libreria.
- Soporta SQLite, Postgres, MySQL, MSSQL, ClickHouse y otros.
- Soporta migraciones SQL embebidas con `embed`.

Decision recomendada:

Adoptar para samples y benchmark antes que escribir migraciones caseras.

### sqlc

Herramienta: `github.com/sqlc-dev/sqlc`

Encaje en Nexgou:

- Generar repositorios tipados para samples grandes.
- Validar consultas SQL en CI.
- Reducir bugs de scan manual.

Por que aporta:

- Genera codigo Go type-safe desde SQL.
- No fuerza ORM ni modelo activo.
- Encaja mejor que un ORM con el objetivo performance/control.

Decision recomendada:

Usarlo como herramienta de desarrollo en samples/benchmark o modulos opcionales. No convertirlo en dependencia runtime del core.

### pgx

Paquete: `github.com/jackc/pgx/v5`

Encaje en Nexgou:

- Modulo opcional de PostgreSQL.
- Pool de conexiones.
- COPY, LISTEN/NOTIFY, batches, prepared statements y tipos PostgreSQL.

Por que aporta:

- Es el driver/toolkit Go de referencia para PostgreSQL de alto rendimiento.
- Tiene soporte para `database/sql`, pero su API nativa es mas potente.

Decision recomendada:

Adoptar solo cuando exista modulo Postgres o sample Postgres. Para el benchmark SQLite actual no aporta.

### go-redis

Paquete: `github.com/redis/go-redis/v9`

Encaje en Nexgou:

- Cache module.
- Backend distribuido para rate limit.
- Pub/Sub para eventos simples.
- Store de sesiones si se agrega auth avanzada.

Por que aporta:

- Cliente oficial Redis para Go.
- Pooling, pipelines, cluster, sentinel y hooks de OpenTelemetry.

Decision recomendada:

Integrar como modulo opcional cuando haya un caso real: cache, rate limit distribuido o pub/sub. No agregarlo al core por defecto.

## Testing e infraestructura

### Testcontainers for Go

Paquete: `github.com/testcontainers/testcontainers-go`

Encaje en Nexgou:

- Tests reales de Postgres, Redis, Prometheus o servicios externos.
- Smoke tests reproducibles sin depender de servicios instalados localmente.

Por que aporta:

- Crea y limpia contenedores durante tests.
- Reduce falsos positivos de mocks para integraciones de infraestructura.

Riesgos:

- Requiere Docker disponible.
- Puede hacer lenta la suite si no se separa por build tags o paquetes especificos.

Decision recomendada:

Adoptar solo en tests de integracion, con build tag o paquete separado. No usar en unit tests.

## Librerias que no conviene meter al core

### ORMs completos

Ejemplos: GORM, ent, Bun.

No son malas librerias, pero para Nexgou conviene evitar que el framework base imponga ORM. El core debe funcionar igual con `database/sql`, `pgx`, SQL generado, repositorios manuales o cualquier capa del usuario.

Recomendacion:

- No incluir en core.
- Documentar ejemplos externos si usuarios los piden.
- Para samples oficiales, preferir `database/sql`, `pgx` o `sqlc`.

### Routers y frameworks HTTP externos

Ejemplos: Gin, Chi, Echo, Fiber.

Nexgou ya tiene router, pipeline y adapters. Integrar otro router como dependencia central duplicaria responsabilidades y confundiria la API.

Recomendacion:

- Usarlos solo como referencia competitiva.
- No integrarlos al core.

### Config loaders pesados

Ejemplos: Viper.

Pueden ser utiles en aplicaciones, pero para el framework base suelen traer mas superficie de la necesaria. Nexgou puede resolver config con un servicio pequeno y permitir que el usuario cargue archivos/env como prefiera.

Recomendacion:

- Mantener config propia y simple.
- Evaluar `envconfig` o similar solo si el `ConfigService` necesita binding tipado real.

## Plan sugerido

1. Estabilizar la suite base actual antes de adoptar mas dependencias productivas.
2. Cerrar la evaluacion de Sonic con `-count=5` y tests de compatibilidad.
3. Crear modulo opcional de OpenTelemetry con middleware/interceptor HTTP.
4. Formalizar `ValidationPipe` sobre `go-playground/validator`.
5. Formalizar `JwtModule` y `JwtGuard` sobre `golang-jwt/jwt`.
6. Evaluar `zerolog` contra el logger actual y decidir si reemplaza el backend interno.
7. Agregar `goose` a samples/benchmark si la persistencia sigue creciendo.
8. Usar Testcontainers solo para integraciones que de verdad necesiten servicios externos.

## Comandos de validacion

```txt
go test ./test/common ./test/logger
go test ./test/adapters/fasthttp
go test ./test/common -run ^$ -bench . -benchmem -count=5
go test ./test/logger -run ^$ -bench . -benchmem -count=5
go test ./test/...
```

Si `go test ./test/...` falla por cambios de migracion no relacionados, no adoptar nuevas dependencias productivas hasta separar esos fallos de la evaluacion.

## Referencias

- Sonic: https://github.com/bytedance/sonic
- OpenTelemetry Go: https://github.com/open-telemetry/opentelemetry-go
- Prometheus Go client: https://github.com/prometheus/client_golang
- go-playground/validator: https://github.com/go-playground/validator
- golang-jwt/jwt: https://github.com/golang-jwt/jwt
- zerolog: https://github.com/rs/zerolog
- zap: https://github.com/uber-go/zap
- goose: https://github.com/pressly/goose
- sqlc: https://github.com/sqlc-dev/sqlc
- pgx: https://github.com/jackc/pgx
- go-redis: https://github.com/redis/go-redis
- Testcontainers for Go: https://github.com/testcontainers/testcontainers-go
