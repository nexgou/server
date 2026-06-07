# Middleware de Seguridad

Nexgou incluye cinco middleware y primitivas por ruta enfocados en seguridad, todos implementados usando la biblioteca estándar de Go — sin dependencias externas.

> Todo el middleware y guards residen en el paquete `github.com/nexgou/server/src/middleware`.

---

## Tabla de Contenidos

- [SecurityHeaders](#securityheaders)
- [CorsWithOptions](#corswithoptions)
- [RateLimit / RateLimitGuard](#ratelimit--ratelimitguard)
- [Timeout / TimeoutInterceptor](#timeout--timeoutinterceptor)
- [BodyLimit / BodyLimitInterceptor](#bodylimit--bodylimitinterceptor)
- [Orden recomendado de la pipeline](#orden-recomendado-de-la-pipeline)
- [Combinar límites globales y por ruta](#combinar-límites-globales-y-por-ruta)

---

## SecurityHeaders

Establece cabeceras HTTP seguras en cada solicitud.

### Uso

```go
// Por defecto — seguro para la mayoría de las aplicaciones
app.Use(middleware.SecurityHeaders())

// Personalizado — sobrescribir cabeceras individuales
app.Use(middleware.SecurityHeaders(middleware.SecurityOptions{
    ContentSecurityPolicy:   "default-src 'self'; img-src *; script-src 'self'",
    XFrameOptions:           "SAMEORIGIN",
    StrictTransportSecurity: "-", // "-" deshabilita la cabecera (p. ej. en desarrollo local)
}))
```

### Campos de SecurityOptions

| Campo                     | Por defecto                                | Descripción                                                |
| ------------------------- | ------------------------------------------ | ---------------------------------------------------------- |
| `ContentSecurityPolicy`   | `default-src 'self'`                       | Controla qué recursos puede cargar el navegador            |
| `XFrameOptions`           | `DENY`                                     | Previene clickjacking mediante incrustación con `<iframe>` |
| `XContentTypeOptions`     | `nosniff`                                  | Previene el sniffing de tipos MIME                         |
| `XXSSProtection`          | `1; mode=block`                            | Habilita el filtro XSS del navegador (navegadores legacy)  |
| `StrictTransportSecurity` | `max-age=31536000; includeSubDomains`      | Fuerza HTTPS durante 1 año                                 |
| `ReferrerPolicy`          | `strict-origin-when-cross-origin`          | Controla la cabecera `Referer`                             |
| `PermissionsPolicy`       | `geolocation=(), microphone=(), camera=()` | Restringe el acceso a funcionalidades del navegador        |

Establecer cualquier campo a `"-"` para omitir esa cabecera completamente.

---

## CorsWithOptions

Política CORS configurable con manejo automático de preflight (`OPTIONS`).

El helper original `Cors()` sigue disponible para casos de uso simples con `Access-Control-Allow-Origin: *`.

### Uso

```go
// API abierta — permitir cualquier origen
app.Use(middleware.CorsWithOptions(middleware.CorsOptions{
    AllowedOrigins: []string{"*"},
}))

// API restringida — orígenes específicos, con credenciales
app.Use(middleware.CorsWithOptions(middleware.CorsOptions{
    AllowedOrigins:   []string{"https://app.example.com", "https://admin.example.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Content-Type", "Authorization"},
    ExposedHeaders:   []string{"X-Request-Id"},
    AllowCredentials: true,
    MaxAge:           3600,
}))
```

### Campos de CorsOptions

| Campo              | Por defecto                                    | Descripción                                                                       |
| ------------------ | ---------------------------------------------- | --------------------------------------------------------------------------------- |
| `AllowedOrigins`   | `["*"]`                                        | Orígenes autorizados a hacer solicitudes cross-origin                             |
| `AllowedMethods`   | `GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS` | Métodos HTTP permitidos                                                           |
| `AllowedHeaders`   | `Content-Type, Authorization`                  | Cabeceras de solicitud permitidas                                                 |
| `ExposedHeaders`   | `[]`                                           | Cabeceras a las que el navegador puede acceder en la respuesta                    |
| `AllowCredentials` | `false`                                        | Permitir cookies / auth HTTP. Incompatible con origen `"*"`                       |
| `MaxAge`           | `600`                                          | Segundos para almacenar en caché la respuesta de preflight. Usar `-1` para omitir |

Las solicitudes preflight (`OPTIONS`) se manejan automáticamente: el middleware responde `204 No Content` y detiene la cadena.

Comportamiento de hardening para preflight:

- Si `Access-Control-Request-Method` está presente y no está en `AllowedMethods`, el middleware devuelve `204` sin cabeceras de permiso de preflight.
- Si `Access-Control-Request-Headers` incluye cabeceras fuera de `AllowedHeaders` (sin distinguir mayúsculas/minúsculas), el middleware devuelve `204` sin cabeceras de permiso de preflight.
- Si `AllowCredentials=true` y `AllowedOrigins=["*"]`, el middleware refleja el `Origin` entrante (en lugar de `*`) para mantener compatibilidad con el comportamiento esperado por navegadores.

---

## RateLimit / RateLimitGuard

Limitador de tasa de ventana fija por IP del cliente. Las solicitudes excedentes reciben `429 Too Many Requests`.

Cabeceras de respuesta establecidas en cada solicitud:

| Cabecera                | Descripción                                             |
| ----------------------- | ------------------------------------------------------- |
| `X-RateLimit-Limit`     | Máximo de solicitudes permitidas en la ventana          |
| `X-RateLimit-Remaining` | Solicitudes restantes en la ventana actual              |
| `Retry-After`           | Segundos hasta que se reinicia la ventana (solo en 429) |

La IP del cliente se resuelve en orden: `X-Forwarded-For` → `X-Real-IP` → `RemoteAddr`.

### Global — middleware

Se aplica a todas las rutas.

```go
// 100 solicitudes por IP por minuto
app.Use(middleware.RateLimit(100, time.Minute))
```

### Por ruta — RateLimitGuard

Implementa la interfaz `Guard`. Adjuntar con `.Guard(...)`.
Los límites por ruta son **independientes** del límite global — el cliente debe satisfacer ambos.

```go
nexgou.Post("/login", c.Login).
    Guard(&middleware.RateLimitGuard{Max: 5, Window: time.Minute})
```

### Campos de RateLimitGuard

| Campo    | Tipo            | Descripción                                         |
| -------- | --------------- | --------------------------------------------------- |
| `Max`    | `int`           | Máximo de solicitudes permitidas dentro de `Window` |
| `Window` | `time.Duration` | La ventana de tiempo para el contador               |

---

## Timeout / TimeoutInterceptor

Cancela el contexto de la solicitud después de la duración configurada.
Tiempo de espera agotado con `408 Request Timeout`.

### Global — middleware

```go
app.Use(middleware.Timeout(30 * time.Second))
```

### Por ruta — TimeoutInterceptor

Implementa la interfaz `Interceptor`. Adjuntar con `.Intercept(...)`.

```go
nexgou.Get("/report", c.HeavyReport).
    Intercept(&middleware.TimeoutInterceptor{Duration: 60 * time.Second})
```

### Campos de TimeoutInterceptor

| Campo      | Tipo            | Descripción                                            |
| ---------- | --------------- | ------------------------------------------------------ |
| `Duration` | `time.Duration` | Tiempo máximo que el handler puede tardar en responder |

> El timeout por ruta reemplaza el timeout global para esa ruta específica — el deadline más restrictivo gana (el que se dispara primero).

---

## BodyLimit / BodyLimitInterceptor

Limita el tamaño máximo del cuerpo de la solicitud usando `http.MaxBytesReader`.
Las solicitudes sobredimensionadas reciben `413 Payload Too Large`.

### Global — middleware

```go
app.Use(middleware.BodyLimit(1 << 20)) // 1 MB
```

### Por ruta — BodyLimitInterceptor

Implementa la interfaz `Interceptor`. Adjuntar con `.Intercept(...)`.

```go
nexgou.Post("/upload", c.Upload).
    Intercept(&middleware.BodyLimitInterceptor{MaxBytes: 50 << 20}) // 50 MB
```

### Campos de BodyLimitInterceptor

| Campo      | Tipo    | Descripción                                 |
| ---------- | ------- | ------------------------------------------- |
| `MaxBytes` | `int64` | Tamaño máximo permitido del cuerpo en bytes |

### Constantes de tamaño comunes

```go
1 << 10  =    1 KB
1 << 20  =    1 MB
10 << 20 =   10 MB
50 << 20 =   50 MB
```

---

## Orden recomendado de la pipeline

```go
app.Use(middleware.Recovery())        // 1. capturar panics de todo lo que sigue
app.Use(middleware.SecurityHeaders()) // 2. establecer cabeceras antes de cualquier escritura
app.Use(middleware.CorsWithOptions(middleware.CorsOptions{
    AllowedOrigins: []string{"*"},
}))                                   // 3. manejar preflight OPTIONS temprano
app.Use(middleware.RateLimit(100, time.Minute)) // 4. rechazar IPs abusivas temprano
app.Use(middleware.Timeout(30 * time.Second))   // 5. limitar todo el trabajo posterior
app.Use(middleware.BodyLimit(1 << 20))          // 6. limitar payload antes de que los handlers lean
app.Use(middleware.Logger())          // 7. registrar estado final (incl. 429, 408, 413)
```

---

## Combinar límites globales y por ruta

Los límites globales y por ruta son **independientes y aditivos**. Una solicitud debe pasar ambos.

```go
// Global: 100 req/min, timeout 30s, cuerpo 1 MB
app.Use(middleware.RateLimit(100, time.Minute))
app.Use(middleware.Timeout(30 * time.Second))
app.Use(middleware.BodyLimit(1 << 20))

// Ruta: adicionalmente 10 req/min, timeout 10s, cuerpo 64 KB
nexgou.Post("/users", c.Create).
    Guard(
        &AuthGuard{},
        &middleware.RateLimitGuard{Max: 10, Window: time.Minute},
    ).
    Intercept(
        &middleware.TimeoutInterceptor{Duration: 10 * time.Second},
        &middleware.BodyLimitInterceptor{MaxBytes: 64 << 10},
    ).
    Version("v1")
```

En el ejemplo anterior, una solicitud a `POST /users` está sujeta a:

- Límite de tasa: el contador que se agote primero (global o de nivel de ruta)
- Timeout: `min(30s, 10s)` = 10 segundos (el deadline por ruta se dispara primero)
- Cuerpo: `min(1 MB, 64 KB)` = 64 KB (el límite por ruta se dispara primero)
