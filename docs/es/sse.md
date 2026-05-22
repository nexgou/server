# Server-Sent Events (SSE)

Nexgou tiene soporte de primera clase para SSE a través del sub-paquete `nexgousse`. Los handlers SSE se integran perfectamente con el sistema de módulos, el contenedor IoC, los Guards y la pipeline de middleware completa — sin servidor separado ni actualización de protocolo.

---

## Tabla de Contenidos

- [Cómo funciona](#cómo-funciona)
- [API de SSEContext](#api-de-ssecontext)
- [Registrar una ruta SSE](#registrar-una-ruta-sse)
  - [Usar el helper nexgou.SSE (recomendado)](#usar-el-helper-nexgousse-recomendado)
  - [Usar ToHTTPHandler directamente](#usar-tohttphandler-directamente)
- [Detectar desconexión del cliente](#detectar-desconexión-del-cliente)
- [Eventos con nombre](#eventos-con-nombre)
- [IDs de evento y reconexión](#ids-de-evento-y-reconexión)
- [Parámetros de URL](#parámetros-de-url)
- [Guards en rutas SSE](#guards-en-rutas-sse)
- [Versionado](#versionado)
- [Ejemplo completo — streaming de métricas](#ejemplo-completo--streaming-de-métricas)
- [Probar con curl](#probar-con-curl)
- [Probar con Postman](#probar-con-postman)
- [Probar desde un navegador](#probar-desde-un-navegador)
- [SSE vs WebSocket — cuándo usar cada uno](#sse-vs-websocket--cuándo-usar-cada-uno)

---

## Cómo funciona

Una ruta SSE es un handler HTTP `GET` estándar que mantiene la conexión abierta y escribe eventos en el formato `text/event-stream`. El framework:

1. Hace coincidir la solicitud con la tabla de rutas HTTP normal.
2. Ejecuta la pipeline de middleware global y cualquier Guard de ruta.
3. Llama a `ToHTTPHandler` que escribe las cabeceras de respuesta SSE requeridas y crea un `SSEContext`.
4. Invoca la función handler SSE durante el tiempo de vida de la conexión.
5. Retorna cuando el handler retorna (desconexión del cliente o retorno explícito).

No tiene lugar ninguna actualización de protocolo — SSE funciona sobre HTTP/1.1 puro y pasa a través de cada proxy, balanceador de carga y CDN sin configuración especial.

### Cabeceras de respuesta escritas automáticamente

| Cabecera | Valor |
|--------|-------|
| `Content-Type` | `text/event-stream` |
| `Cache-Control` | `no-cache` |
| `Connection` | `keep-alive` |
| `X-Accel-Buffering` | `no` *(deshabilita el buffering de nginx/proxy)* |

---

## API de SSEContext

`SSEContext` envuelve el `http.ResponseWriter` y expone una API limpia para escribir eventos SSE.

### Métodos de escritura

| Método | Salida SSE | Descripción |
|--------|-----------|-------------|
| `Send(data string) error` | `data: <data>\n\n` | Enviar un evento sin nombre de texto plano |
| `SendNamed(event, data string) error` | `event: <name>\ndata: <data>\n\n` | Enviar un evento con nombre |
| `SendJSON(v any) error` | `data: <json>\n\n` | Serializar `v` como JSON y enviar |
| `SendNamedJSON(event string, v any) error` | `event: <name>\ndata: <json>\n\n` | Serializar `v` como JSON y enviar como evento con nombre |
| `SendComment(comment string) error` | `: <comment>\n\n` | Enviar un comentario keepalive (invisible para `EventSource`) |
| `SetRetry(ms int) error` | `retry: <ms>\n\n` | Establecer el delay de reconexión del navegador en milisegundos |
| `SetID(id string) error` | `id: <id>\n\n` | Establecer el ID del último evento |

> Los valores de datos multi-línea se dividen automáticamente: cada `\n` en el string de datos produce una línea `data:` separada, según lo requiere la especificación SSE.

### Helpers de solicitud

| Método | Descripción |
|--------|-------------|
| `Param(key string) string` | Leer un parámetro de ruta URL (p. ej. `:topic` → `"topic"`) |
| `Header(key string) string` | Leer una cabecera de la solicitud original |
| `LastEventID() string` | Valor de la cabecera `Last-Event-ID` (enviada en reconexión del navegador) |
| `RemoteAddr() string` | Dirección de red del cliente |
| `Done() <-chan struct{}` | Se cierra cuando el cliente se desconecta |
| `Request *http.Request` | La solicitud HTTP original (solo lectura) |

---

## Registrar una ruta SSE

### Usar el helper nexgou.SSE (recomendado)

```go
func (c *NotificationsController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.SSE("/notifications", c.Stream),
    }
}
```

`nexgou.SSE` devuelve un `nexgou.Route` estándar para que puedas encadenar `.Guard()`, `.Version()` y `.Intercept()` exactamente como cualquier ruta HTTP:

```go
nexgou.SSE("/notifications", c.Stream).
    Guard(&AuthGuard{}).
    Version("v1")
// path efectivo: /v1/notifications
```

### Usar ToHTTPHandler directamente

Si prefieres control explícito o necesitas mezclar SSE y HTTP en la misma lista `Register()` sin el helper:

```go
import nexgousse "github.com/nexgou/server/src/sse"

func (c *NotificationsController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.Get("/notifications", nexgousse.ToHTTPHandler(c.Stream)),
    }
}
```

---

## Detectar desconexión del cliente

Usar `ctx.Done()` para detener el bucle de streaming limpiamente cuando el cliente cierra la conexión. Esto previene fugas de goroutines.

```go
func (c *NotificationsController) Stream(ctx *nexgou.SSEContext) error {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            // Cliente desconectado — devolver nil para una salida limpia.
            return nil
        case <-ticker.C:
            if err := ctx.SendJSON(nexgou.H{"ts": time.Now().Unix()}); err != nil {
                return err
            }
        }
    }
}
```

`ctx.Done()` está respaldado por `http.Request.Context().Done()`, que Go's HTTP server cierra tan pronto como se detecta que la conexión TCP del cliente está cerrada.

---

## Eventos con nombre

Los eventos con nombre permiten al navegador registrar listeners específicos con `EventSource.addEventListener`:

```go
// Servidor
ctx.SendNamedJSON("user.created", nexgou.H{"id": 42, "name": "Alice"})
ctx.SendNamedJSON("order.shipped", nexgou.H{"order": "ORD-001"})
```

```js
// Navegador
const es = new EventSource('/notifications')

es.addEventListener('user.created', (e) => {
    const user = JSON.parse(e.data)
    console.log('Nuevo usuario:', user.name)
})

es.addEventListener('order.shipped', (e) => {
    const order = JSON.parse(e.data)
    console.log('Enviado:', order.order)
})
```

Los eventos sin nombre (`Send`, `SendJSON`) son recibidos por `es.onmessage`.

---

## IDs de evento y reconexión

Establecer un ID en cada evento para que el navegador pueda reanudar desde donde lo dejó después de una reconexión:

```go
func (c *FeedController) Stream(ctx *nexgou.SSEContext) error {
    ctx.SetRetry(3000)

    // Reanudar desde el último evento visto si el cliente se está reconectando.
    lastID := ctx.LastEventID()
    events := c.service.EventsSince(lastID)

    for _, ev := range events {
        ctx.SetID(ev.ID)
        ctx.SendNamedJSON("feed", ev)
    }

    // Continuar con eventos en vivo...
}
```

En la reconexión, el navegador envía `Last-Event-ID: <last-id>` automáticamente.

---

## Parámetros de URL

Define parámetros con un prefijo de dos puntos en el path, luego léelos con `ctx.Param`:

```go
nexgou.SSE("/metrics/live/:topic", c.StreamTopic)

func (c *MetricsController) StreamTopic(ctx *nexgou.SSEContext) error {
    topic := ctx.Param("topic") // "cpu", "mem", etc.
    // ...
}
```

---

## Guards en rutas SSE

Los guards se ejecutan como middleware HTTP normal antes de que la conexión se entregue al handler SSE. Un guard fallido responde con HTTP `403` y el stream nunca comienza.

```go
nexgou.SSE("/notifications", c.Stream).Guard(&AuthGuard{})
```

```go
type AuthGuard struct{}

func (g *AuthGuard) CanActivate(ctx *nexgou.Context) (bool, error) {
    return ctx.Header("Authorization") != "", nil
}
```

Los múltiples guards se evalúan en orden; el primer fallo hace un cortocircuito:

```go
nexgou.SSE("/admin/events", c.Stream).Guard(&AuthGuard{}, &AdminRoleGuard{})
```

---

## Versionado

```go
nexgou.SSE("/notifications", c.Stream).Version("v1")
// path efectivo: /v1/notifications
```

---

## Ejemplo completo — streaming de métricas

### `metrics/metrics.controller.go`

```go
package metrics

import (
    "math/rand"
    "time"

    nexgou "github.com/nexgou/server"
)

type MetricsController struct{}

func NewMetricsController() *MetricsController { return &MetricsController{} }

func (c *MetricsController) Register() []nexgou.Route {
    return []nexgou.Route{
        // Stream completo: todas las métricas cada segundo.
        nexgou.SSE("/metrics/live", c.StreamAll),
        // Stream filtrado: un solo topic (cpu | mem | goroutines).
        nexgou.SSE("/metrics/live/:topic", c.StreamTopic),
    }
}

func (c *MetricsController) StreamAll(ctx *nexgou.SSEContext) error {
    ctx.SetRetry(5000)
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return nil
        case <-ticker.C:
            ctx.SendNamedJSON("metrics", nexgou.H{
                "cpu":        20 + rand.Float64()*60,
                "mem":        30 + rand.Float64()*50,
                "goroutines": 10 + rand.Intn(90),
                "ts":         time.Now().UTC().Format(time.RFC3339),
            })
        }
    }
}

func (c *MetricsController) StreamTopic(ctx *nexgou.SSEContext) error {
    topic := ctx.Param("topic")
    if topic != "cpu" && topic != "mem" && topic != "goroutines" {
        return nexgou.BadRequestException("topic desconocido: " + topic)
    }
    ctx.SetRetry(5000)
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return nil
        case <-ticker.C:
            ctx.SendNamedJSON(topic, nexgou.H{
                "value": 20 + rand.Float64()*60,
                "ts":    time.Now().UTC().Format(time.RFC3339),
            })
        }
    }
}
```

### `metrics/metrics.module.go`

```go
package metrics

import nexgou "github.com/nexgou/server"

var Module = nexgou.Module(nexgou.ModuleOptions{
    Controllers: []any{NewMetricsController},
})
```

### Registrar en `app.module.go`

```go
var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{
        nexgou.ConfigModule,
        nexgou.LogModule,
        metrics.Module,
        // ...
    },
})
```

---

## Probar con curl

```bash
# Stream completo de métricas
curl -N http://localhost:3000/metrics/live

# Topic único
curl -N http://localhost:3000/metrics/live/cpu
```

El flag `-N` (`--no-buffer`) deshabilita el buffering de salida de curl para que los eventos aparezcan inmediatamente. Salida esperada:

```
event: metrics
data: {"cpu":47.23,"goroutines":34,"mem":61.05,"ts":"2026-05-22T10:00:01Z"}

event: metrics
data: {"cpu":52.11,"goroutines":41,"mem":58.99,"ts":"2026-05-22T10:00:02Z"}
```

---

## Probar con Postman

1. Abrir Postman → **New → HTTP**
2. Método: `GET`, URL: `http://localhost:3000/metrics/live`
3. Si la ruta tiene un Guard, añadir la cabecera requerida en la pestaña **Headers**
4. Hacer clic en **Send**
5. Postman muestra cada evento a medida que llega en el panel de respuesta

> Postman renderiza respuestas SSE en tiempo real a partir de v10+. Si ves la respuesta completa solo después de cerrar el stream, actualiza Postman.

---

## Probar desde un navegador

Abrir la consola del navegador y ejecutar:

```js
const es = new EventSource('http://localhost:3000/metrics/live')

// Eventos con nombre
es.addEventListener('metrics', (e) => {
    console.log(JSON.parse(e.data))
})

// Error / desconexión
es.onerror = () => console.warn('SSE desconectado — el navegador reconectará automáticamente')
```

Para detener el stream:

```js
es.close()
```

---

## SSE vs WebSocket — cuándo usar cada uno

| | SSE | WebSocket |
|---|---|---|
| Dirección | Solo servidor → cliente | Bidireccional |
| Protocolo | HTTP puro | Actualización HTTP a WS |
| Reconexión automática | Sí — integrada en `EventSource` | No — debe implementarse manualmente |
| Soporte de proxy / CDN | Funciona de inmediato | Requiere proxy compatible con WS |
| Complejidad | Simple — handler HTTP estándar | Mayor — ciclo de vida de conexión persistente |
| **Mejor para** | Notificaciones, feeds en vivo, progreso, dashboards | Chat, edición colaborativa, gaming |

**Regla general:** usar SSE cuando el servidor habla y el cliente escucha. Usar WebSocket cuando ambos lados necesitan enviar mensajes en cualquier momento.
