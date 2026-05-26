# Módulos del Ecosistema Nexgou

Nexgou está organizado como un **ecosistema modular**. El framework principal (`github.com/nexgou/server`) incluye módulos integrados para las necesidades más comunes, y módulos adicionales de primera parte están publicados como paquetes independientes bajo la organización [`nexgou`](https://github.com/nexgou) en GitHub.

Cada módulo sigue la misma convención: impórtalo, agrégalo a `Imports` en cualquier `nexgou.Module(...)` y sus servicios quedan disponibles para Inyección de Dependencias.

---

## Tabla de contenido

| Módulo                          | Import path                       | Categoría     |
| ------------------------------- | --------------------------------- | ------------- |
| [server](#server)               | `github.com/nexgou/server`        | Core          |
| [validation](#validation)       | `github.com/nexgou/validation`    | Core          |
| [caching](#caching)             | `github.com/nexgou/caching`       | Datos         |
| [database](#database)           | `github.com/nexgou/database`      | Datos         |
| [mongo](#mongo)                 | `github.com/nexgou/mongo`         | Datos         |
| [postgres](#postgres)           | `github.com/nexgou/postgres`      | Datos         |
| [redis](#redis)                 | `github.com/nexgou/redis`         | Datos         |
| [sqlite](#sqlite)               | `github.com/nexgou/sqlite`        | Datos         |
| [compression](#compression)     | `github.com/nexgou/compression`   | HTTP          |
| [cookie](#cookie)               | `github.com/nexgou/cookie`        | HTTP          |
| [fileupload](#fileupload)       | `github.com/nexgou/fileupload`    | HTTP          |
| [graphql](#graphql)             | `github.com/nexgou/graphql`       | HTTP          |
| [streaming](#streaming)         | `github.com/nexgou/streaming`     | HTTP          |
| [jwt](#jwt)                     | `github.com/nexgou/jwt`           | Seguridad     |
| [cron](#cron)                   | `github.com/nexgou/cron`          | Planificación |
| [scheduler](#scheduler)         | `github.com/nexgou/scheduler`     | Planificación |
| [events](#events)               | `github.com/nexgou/events`        | Mensajería    |
| [mqtt](#mqtt)                   | `github.com/nexgou/mqtt`          | Mensajería    |
| [nats](#nats)                   | `github.com/nexgou/nats`          | Mensajería    |
| [queues](#queues)               | `github.com/nexgou/queues`        | Mensajería    |
| [rabbitmq](#rabbitmq)           | `github.com/nexgou/rabbitmq`      | Mensajería    |
| [serialization](#serialization) | `github.com/nexgou/serialization` | Mensajería    |
| [sqs](#sqs)                     | `github.com/nexgou/sqs`           | Mensajería    |

---

## Core

### server

**`github.com/nexgou/server`**

El paquete principal del framework. Provee el bootstrap de la aplicación, el motor HTTP, el sistema de módulos, el contenedor IoC, Inyección de Dependencias, Controllers, Guards, Interceptors, Pipes, Exception Filters, `ConfigModule`, `LogModule` y las utilidades de testing `nexgoutest`.

Es la única dependencia obligatoria — todos los demás módulos de esta lista son complementos opcionales.

```go
go get github.com/nexgou/server
```

---

### validation

**`github.com/nexgou/validation`**

Validación de structs mediante field tags. Integra `go-playground/validator/v10` y devuelve errores estructurados por campo, listos para respuestas de API.

Úsalo para:

- Validar cuerpos de peticiones (DTOs) antes de procesarlos.
- Aplicar reglas de negocio sobre structs con tags (`required`, `email`, `min`, `max`, `oneof`, `uuid`, `url`, …).
- Registrar funciones de validación personalizadas con nombre propio.

```go
go get github.com/nexgou/validation
```

```go
var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{nexgou.ValidationModule},
})
```

---

## Datos

### caching

**`github.com/nexgou/caching`**

Capa de caché genérica e independiente del backend. Provee un `CacheService` con operaciones `Get`, `Set`, `Delete` y `Clear`.

Úsalo para:

- Cachear cómputos costosos o resultados de consultas a base de datos.
- Almacenar datos de sesión con TTL.
- Cambiar de backend de caché (en memoria, Redis, …) sin modificar el código de la aplicación.

```go
go get github.com/nexgou/caching
```

---

### database

**`github.com/nexgou/database`**

Capa de abstracción de base de datos construida sobre `database/sql`. Provee pooling de conexiones, helpers de transacciones y query builders compatibles con cualquier driver `database/sql`.

Úsalo para:

- Gestionar conexiones SQL con un único `DatabaseService` inyectable.
- Ejecutar consultas y transacciones de forma consistente en distintos motores SQL.
- Cambiar entre SQLite, PostgreSQL, MySQL y otros solo cambiando el driver.

```go
go get github.com/nexgou/database
```

---

### mongo

**`github.com/nexgou/mongo`**

Módulo de integración con MongoDB. Envuelve el driver oficial `mongo-driver` y expone un `MongoService` listo para Inyección de Dependencias.

Úsalo para:

- Conectar a clusters MongoDB con configuración tipada.
- Acceder a colecciones a través de un servicio inyectable.
- Usar change streams y pipelines de agregación dentro del ciclo de vida de Nexgou.

```go
go get github.com/nexgou/mongo
```

---

### postgres

**`github.com/nexgou/postgres`**

Módulo de integración con PostgreSQL. Envuelve `pgx` y expone un `PostgresService` con pooling de conexiones, prepared statements y query helpers tipados.

Úsalo para:

- Conectar a PostgreSQL con el rendimiento completo de `pgx`.
- Ejecutar consultas parametrizadas, operaciones en batch y COPY FROM.
- Aprovechar características específicas de PostgreSQL (LISTEN/NOTIFY, advisory locks, JSONB).

```go
go get github.com/nexgou/postgres
```

---

### redis

**`github.com/nexgou/redis`**

Módulo de integración con Redis. Envuelve `go-redis` y expone un `RedisService` para operaciones clave-valor, pub/sub, streams y bloqueos distribuidos.

Úsalo para:

- Almacenar y recuperar valores cacheados con TTL.
- Implementar comunicación pub/sub entre servicios.
- Usar Redis Streams para event sourcing o colas de tareas.
- Bloqueos distribuidos y rate limiting respaldados por Redis.

```go
go get github.com/nexgou/redis
```

---

### sqlite

**`github.com/nexgou/sqlite`**

Módulo de integración con SQLite. Envuelve `modernc.org/sqlite` (Go puro, sin CGO) y expone un `SQLiteService` para bases de datos locales embebidas.

Úsalo para:

- Embeber una base de datos SQL completa sin dependencias externas.
- Ejecutar entornos de desarrollo o despliegues en el edge sin servidor de base de datos.
- Almacenar configuración, estado o logs de eventos locales en disco.

```go
go get github.com/nexgou/sqlite
```

---

## HTTP

### compression

**`github.com/nexgou/compression`**

Middleware y servicio de compresión de respuestas. Soporta codificación gzip, deflate y brotli con umbrales configurables y filtros por tipo MIME.

Úsalo para:

- Reducir el tamaño del payload de respuesta automáticamente.
- Configurar la compresión por ruta o de forma global.
- Controlar qué tipos de contenido y tamaños mínimos activan la compresión.

```go
go get github.com/nexgou/compression
```

---

### cookie

**`github.com/nexgou/cookie`**

Gestión tipada de cookies. Provee un `CookieService` para leer, escribir y eliminar cookies HTTP con soporte para valores firmados y cifrados.

Úsalo para:

- Establecer y leer cookies HTTP de forma type-safe.
- Firmar cookies para prevenir manipulación del lado del cliente.
- Cifrar valores de cookie para datos sensibles.
- Gestionar atributos de cookie (SameSite, HttpOnly, Secure, MaxAge).

```go
go get github.com/nexgou/cookie
```

---

### fileupload

**`github.com/nexgou/fileupload`**

Gestión de subida de archivos multipart. Provee un `FileUploadService` con límites de tamaño, validación de tipo MIME y backends de almacenamiento local y remoto.

Úsalo para:

- Aceptar subida de uno o varios archivos en handlers HTTP.
- Validar tipos y tamaños de archivo antes de procesarlos.
- Transmitir archivos directamente a almacenamiento de objetos (S3, GCS, disco local).
- Generar URLs de subida y metadatos.

```go
go get github.com/nexgou/fileupload
```

---

### graphql

**`github.com/nexgou/graphql`**

Integración de servidor GraphQL. Registra un `GraphQLController` que sirve un schema definido en Go (schema-first o code-first) y opcionalmente expone un playground GraphiQL.

Úsalo para:

- Construir una API GraphQL junto a (o en lugar de) endpoints REST.
- Definir queries, mutations y subscriptions.
- Reutilizar servicios y guards de Nexgou existentes en los resolvers.

```go
go get github.com/nexgou/graphql
```

---

### streaming

**`github.com/nexgou/streaming`**

Utilidades para streaming de respuestas HTTP por chunks. Provee helpers para transmitir grandes conjuntos de datos, NDJSON, CSV o flujos de bytes arbitrarios sin cargar todo en memoria.

Úsalo para:

- Transmitir resultados de consultas grandes al cliente de forma incremental.
- Exportar reportes o archivos CSV de gran tamaño al vuelo.
- Enviar NDJSON (JSON separado por líneas) para feeds de logs o eventos.
- Evitar picos de memoria al servir payloads grandes.

```go
go get github.com/nexgou/streaming
```

---

## Seguridad

### jwt

**`github.com/nexgou/jwt`**

Módulo de autenticación JWT (JSON Web Token). Provee un `JWTService` para firmar y verificar tokens, más un `JWTGuard` listo para usar que protege rutas.

Úsalo para:

- Emitir tokens de acceso y refresh con expiración configurable.
- Verificar y decodificar JWTs en guards o interceptors.
- Soportar algoritmos de firma RS256, ES256 y HS256.
- Adjuntar claims decodificados al `nexgou.Context` para handlers posteriores.

```go
go get github.com/nexgou/jwt
```

```go
// Proteger una ruta con JWTGuard
nexgou.Get("/perfil", c.Perfil).Guard(&jwt.JWTGuard{})
```

---

## Planificación

### cron

**`github.com/nexgou/cron`**

Planificador de tareas basado en expresiones cron. Provee un `CronService` para registrar funciones que se ejecutan según un schedule definido por una expresión cron estándar.

Úsalo para:

- Ejecutar trabajos de fondo recurrentes (reportes nocturnos, precalentamiento de caché).
- Definir schedules con expresiones cron estándar de 5 o 6 campos.
- Iniciar y detener trabajos individuales en tiempo de ejecución.

```go
go get github.com/nexgou/cron
```

```go
cron.Schedule("0 0 * * *", func() { /* trabajo nocturno */ })
```

---

### scheduler

**`github.com/nexgou/scheduler`**

Planificador de tareas flexible con soporte para trabajos por intervalo, retardados y de una sola ejecución. Complementa a `cron` cuando se necesita scheduling dinámico en runtime en lugar de expresiones cron estáticas.

Úsalo para:

- Planificar una tarea para ejecutarse tras un retardo (`RunAfter(5*time.Minute, fn)`).
- Ejecutar una tarea a intervalo fijo (`Every(10*time.Second, fn)`).
- Cancelar o reprogramar trabajos en tiempo de ejecución.
- Gestionar colas de tareas con límites de concurrencia.

```go
go get github.com/nexgou/scheduler
```

---

## Mensajería

### events

**`github.com/nexgou/events`**

Bus de eventos en proceso. Provee un `EventEmitter` para publicar y suscribirse a eventos tipados dentro de una sola instancia de aplicación.

Úsalo para:

- Desacoplar servicios con un patrón publish/subscribe.
- Emitir eventos de dominio desde un módulo y manejarlos en otro.
- Reemplazar llamadas directas entre servicios con notificaciones asíncronas fire-and-forget.

```go
go get github.com/nexgou/events
```

```go
emitter.Emit("usuario.creado", UsuarioCreadoEvent{ID: "123"})
emitter.On("usuario.creado", func(e UsuarioCreadoEvent) { /* handler */ })
```

---

### mqtt

**`github.com/nexgou/mqtt`**

Módulo de mensajería MQTT. Se conecta a cualquier broker MQTT (Mosquitto, HiveMQ, AWS IoT, …) y expone un `MQTTService` para publicar y suscribirse a topics.

Úsalo para:

- Conectar dispositivos IoT o nodos edge mediante MQTT.
- Publicar datos de sensores o comandos en topics.
- Suscribirse a topics con wildcards y procesar mensajes entrantes en handlers.
- Manejar niveles de QoS (0, 1, 2) y sesiones persistentes.

```go
go get github.com/nexgou/mqtt
```

---

### nats

**`github.com/nexgou/nats`**

Módulo de mensajería NATS. Envuelve el cliente oficial `nats.go` y provee un `NATSService` para pub/sub, request/reply y persistencia con JetStream.

Úsalo para:

- Difundir eventos entre múltiples instancias de servicio.
- Implementar comunicación request/reply entre microservicios sobre NATS.
- Usar JetStream para entrega de mensajes durable y at-least-once.
- Reemplazar llamadas HTTP entre servicios internos con mensajes NATS de alta velocidad.

```go
go get github.com/nexgou/nats
```

---

### queues

**`github.com/nexgou/queues`**

Abstracción de cola de trabajos independiente del backend. Provee una interfaz `QueueService` que funciona con múltiples backends (Redis, RabbitMQ, SQS, en memoria) a través de una API común.

Úsalo para:

- Delegar tareas lentas o pesadas a workers de fondo.
- Producir y consumir trabajos con reintentos y manejo de dead-letter.
- Cambiar de backend de cola sin modificar el código de productores o consumidores.
- Implementar fan-out, trabajos retardados y colas con prioridad.

```go
go get github.com/nexgou/queues
```

---

### rabbitmq

**`github.com/nexgou/rabbitmq`**

Módulo de integración con RabbitMQ. Envuelve `amqp091-go` y expone un `RabbitMQService` para publicar y consumir mensajes AMQP con gestión de topología de exchanges y colas.

Úsalo para:

- Publicar mensajes en exchanges (direct, fanout, topic, headers).
- Consumir mensajes con acknowledgment manual o automático.
- Declarar exchanges, colas y bindings mediante código.
- Implementar patrones RPC sobre RabbitMQ.

```go
go get github.com/nexgou/rabbitmq
```

---

### serialization

**`github.com/nexgou/serialization`**

Serialización y deserialización en múltiples formatos. Provee un `SerializationService` que soporta JSON, MessagePack, Protocol Buffers y CBOR a través de una interfaz unificada.

Úsalo para:

- Codificar/decodificar datos entre formatos sin cambiar la lógica de negocio.
- Serializar mensajes antes de publicarlos en una cola o broker.
- Elegir el formato wire más eficiente según el caso de uso (JSON para APIs, MessagePack para colas, Protocol Buffers para servicios internos).

```go
go get github.com/nexgou/serialization
```

---

### sqs

**`github.com/nexgou/sqs`**

Módulo de integración con AWS SQS. Envuelve el cliente SQS del AWS SDK v2 y expone un `SQSService` para enviar, recibir y eliminar mensajes de colas SQS.

Úsalo para:

- Enviar mensajes a colas SQS estándar o FIFO.
- Hacer polling y procesar mensajes en un loop de worker.
- Manejar visibility timeouts, reintentos y dead-letter queues.
- Integrar con triggers de AWS Lambda o workers en EC2 dentro del mismo codebase.

```go
go get github.com/nexgou/sqs
```

---

## Usando varios módulos juntos

```go
var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{
        nexgou.ConfigModule,
        nexgou.LogModule,
        nexgou.ValidationModule,   // github.com/nexgou/validation
        postgres.PostgresModule,   // github.com/nexgou/postgres
        redis.RedisModule,         // github.com/nexgou/redis
        jwt.JWTModule,             // github.com/nexgou/jwt
        cron.CronModule,           // github.com/nexgou/cron
        events.EventsModule,       // github.com/nexgou/events
        UsuariosModule,
        PedidosModule,
    },
})
```

Cada módulo expone sus servicios automáticamente a través de DI — inyéctalos en cualquier constructor de controller o servicio sin cableado manual.
