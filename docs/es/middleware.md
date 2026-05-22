# Middleware

> **[← Volver al README](../../README.es.md)**

---

## Tabla de Contenidos

- [Cómo funciona el middleware](#cómo-funciona-el-middleware)
- [Orden recomendado de la pipeline](#orden-recomendado-de-la-pipeline)
- [Recovery](#recovery)
- [Logger](#logger)
- [CORS](#cors)
- [Cabeceras de seguridad](#cabeceras-de-seguridad)
- [Limitación de tasa](#limitación-de-tasa)
- [Timeout](#timeout)
- [Límite de tamaño del cuerpo](#límite-de-tamaño-del-cuerpo)
- [Middleware personalizado](#middleware-personalizado)
- [Global vs por ruta](#global-vs-por-ruta)

> Para detalles de configuración enfocados en seguridad (valores de cabeceras, opciones CORS, cabeceras de límite de tasa, etc.), ver [Seguridad](security.md).

---

## Cómo funciona el middleware

El middleware es una función que envuelve un handler. Recibe el siguiente handler en la cadena y puede ejecutar código antes, después, o en lugar de él.

```go
type MiddlewareFunc func(HandlerFunc) HandlerFunc
type HandlerFunc    func(*Context) error
```

El middleware se registra en el `App` y se ejecuta **en el orden de registro** en cada solicitud:

```go
app.Use(middleware.Recovery())   // se ejecuta primero
app.Use(middleware.Logger())     // se ejecuta segundo
// ...el handler se ejecuta al final
```

El flujo de ejecución para una solicitud con dos middlewares:

```
Solicitud
  → Recovery.antes
    → Logger.antes
      → Handler
    → Logger.después
  → Recovery.después
Respuesta
```

Si algún middleware o handler devuelve un error, este se propaga hacia arriba en la cadena. El filtro global de excepciones (si está configurado) lo captura en el nivel superior.

---

## Orden recomendado de la pipeline

```go
app.Use(middleware.Recovery())          // 1. Capturar panics primero — nada por encima
app.Use(middleware.SecurityHeaders())   // 2. Cabeceras de seguridad en cada respuesta
app.Use(middleware.CorsWithOptions(...))// 3. CORS — debe ejecutarse antes de los handlers reales
app.Use(middleware.RateLimit(...))      // 4. Rechazar abuso temprano
app.Use(middleware.Timeout(...))        // 5. Establecer un deadline
app.Use(middleware.BodyLimit(...))      // 6. Limitar el cuerpo antes de leerlo
app.Use(middleware.Logger())           // 7. Registrar después de los filtros anteriores (métricas más limpias)
```

---

## Recovery

**Paquete:** `github.com/nexgou/server/src/middleware`

Recupera de panics de Go en cualquier lugar de la cadena de handlers y los convierte en una respuesta `500 Internal Server Error`. **Siempre regístrate primero** para cubrir todo el middleware posterior.

```go
app.Use(middleware.Recovery())
```

Comportamiento:
- Llama a `recover()` en una función diferida
- Devuelve `500` con el mensaje del panic (texto plano, no JSON — combinar con `HttpExceptionFilter` para JSON)
- Registra el panic en stderr con un stack trace

---

## Logger

**Paquete:** `github.com/nexgou/server/src/middleware`

Registra cada solicitud HTTP con método, path, código de estado y duración.

```go
app.Use(middleware.Logger())
```

Salida de ejemplo (coloreada en terminales):

```
[Nexgou] GET /users 200 1.23ms
[Nexgou] POST /users 201 456µs
[Nexgou] GET /users/999 404 89µs
```

El logger incorporado usa el escritor colorizado interno del framework. Si necesitas logging JSON estructurado, usa [LogModule](logger.md) junto con un middleware personalizado.

---

## CORS

**Paquete:** `github.com/nexgou/server/src/middleware`

### CORS simple (permitir todo)

```go
app.Use(middleware.Cors())
// Establece: Access-Control-Allow-Origin: *
```

### CORS configurable

```go
app.Use(middleware.CorsWithOptions(middleware.CorsOptions{
    AllowedOrigins:   []string{"https://app.example.com", "https://admin.example.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
    AllowedHeaders:   []string{"Authorization", "Content-Type"},
    ExposedHeaders:   []string{"X-Request-ID"},
    AllowCredentials: true,
    MaxAge:           86400, // caché de preflight: 24h
}))
```

| Opción | Tipo | Por defecto | Descripción |
|:---|:---|:---|:---|
| `AllowedOrigins` | `[]string` | `["*"]` | Valores de `Origin` permitidos |
| `AllowedMethods` | `[]string` | verbos comunes | Métodos HTTP permitidos |
| `AllowedHeaders` | `[]string` | `["*"]` | Cabeceras de solicitud permitidas |
| `ExposedHeaders` | `[]string` | `[]` | Cabeceras de respuesta expuestas al navegador |
| `AllowCredentials` | `bool` | `false` | Permitir `credentials: include` en solicitudes |
| `MaxAge` | `int` | `0` | Caché de resultado de preflight en segundos |

Las solicitudes preflight `OPTIONS` se manejan automáticamente con `204 No Content`.

---

## Cabeceras de seguridad

Establece 7 cabeceras HTTP de seguridad en cada respuesta.

```go
app.Use(middleware.SecurityHeaders())
```

Valores por defecto:

| Cabecera | Valor por defecto |
|:---|:---|
| `Content-Security-Policy` | `default-src 'self'` |
| `X-Frame-Options` | `DENY` |
| `X-Content-Type-Options` | `nosniff` |
| `X-XSS-Protection` | `1; mode=block` |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` |
| `Referrer-Policy` | `strict-origin-when-cross-origin` |
| `Permissions-Policy` | `geolocation=(), microphone=(), camera=()` |

Sobrescribir o deshabilitar cabeceras individuales:

```go
app.Use(middleware.SecurityHeaders(middleware.SecurityOptions{
    ContentSecurityPolicy: "default-src 'self'; script-src 'self' cdn.example.com",
    XFrameOptions:         "-",  // "-" deshabilita la cabecera completamente
}))
```

Ver [Seguridad](security.md) para detalles completos.

---

## Limitación de tasa

### Límite de tasa global

```go
app.Use(middleware.RateLimit(100, time.Minute))
// 100 solicitudes por minuto por dirección IP, globalmente
```

Devuelve `429 Too Many Requests` con cabecera `Retry-After` cuando se supera el límite.

### Guard de límite de tasa por ruta

```go
nexgou.Post("/auth/login", c.Login).
    Guard(&middleware.RateLimitGuard{Max: 5, Window: time.Minute})
```

Establece las cabeceras `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset` en cada respuesta.

Tanto los límites globales como los por ruta se aplican de forma independiente — el más restrictivo gana.

Ver [Seguridad](security.md) para la referencia de cabeceras de respuesta.

---

## Timeout

### Timeout global

```go
app.Use(middleware.Timeout(30 * time.Second))
```

Devuelve `408 Request Timeout` si el handler no completa dentro del deadline.

### Interceptor de timeout por ruta

```go
nexgou.Post("/slow-operation", c.Process).
    Intercept(&middleware.TimeoutInterceptor{Duration: 120 * time.Second})
```

El interceptor por ruta reemplaza el deadline efectivo para esa ruta específica.

---

## Límite de tamaño del cuerpo

### Límite global del cuerpo

```go
app.Use(middleware.BodyLimit(1 << 20)) // 1 MB
```

Devuelve `413 Content Too Large` si el cuerpo de la solicitud supera el límite.

### Interceptor de límite del cuerpo por ruta

```go
nexgou.Post("/uploads", c.Upload).
    Intercept(&middleware.BodyLimitInterceptor{MaxBytes: 50 << 20}) // 50 MB
```

Útil para permitir cargas grandes en rutas específicas manteniendo un límite global estricto.

---

## Middleware personalizado

Cualquier función que coincida con `func(nexgou.HandlerFunc) nexgou.HandlerFunc` es middleware válido.

```go
func RequestIDMiddleware(next nexgou.HandlerFunc) nexgou.HandlerFunc {
    return func(ctx *nexgou.Context) error {
        id := uuid.New().String()
        ctx.Writer.Header().Set("X-Request-ID", id)
        ctx.Request.Header.Set("X-Request-ID", id)
        return next(ctx)
    }
}

app.Use(RequestIDMiddleware)
```

### Acceder a los objetos http subyacentes

```go
func MyMiddleware(next nexgou.HandlerFunc) nexgou.HandlerFunc {
    return func(ctx *nexgou.Context) error {
        // Acceder a la solicitud raw
        r := ctx.Request
        w := ctx.Writer

        // Establecer cabeceras de respuesta antes de que se ejecute el handler
        w.Header().Set("X-Powered-By", "Nexgou")

        return next(ctx)
    }
}
```

---

## Global vs por ruta

| Mecanismo | Registro | Alcance |
|:---|:---|:---|
| `app.Use(mw)` | En el `App` | Cada solicitud |
| `.Guard(g)` en ruta | En una `Route` | Solo esa ruta, se ejecuta antes del handler |
| `.Intercept(i)` en ruta | En una `Route` | Solo esa ruta, envuelve el handler |

El middleware global y los guards/interceptores por ruta son **aditivos**. Si se establece un límite de tasa global de 100 req/min Y un límite por ruta de 5 req/min en `/login`, ambos se aplican de forma independiente — llegar a 5 en `/login` activa el guard de nivel de ruta aunque el contador global esté aún en 4.
