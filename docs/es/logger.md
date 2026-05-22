# Logger

> **[← Volver al README](../../README.es.md)**

---

## Tabla de Contenidos

- [Descripción general](#descripción-general)
- [Habilitar LogModule](#habilitar-logmodule)
- [Inyectar LoggerService](#inyectar-loggerservice)
- [Niveles de log](#niveles-de-log)
- [Loggers con alcance](#loggers-con-alcance)
- [Campos estructurados](#campos-estructurados)
- [Formatos de salida](#formatos-de-salida)
- [Variables de entorno](#variables-de-entorno)
- [Referencia de la API](#referencia-de-la-api)

---

## Descripción general

`LoggerService` es un logger estructurado con dos modos de salida: texto coloreado (desarrollo) y JSON (producción). Es provisto por `nexgou.LogModule` e inyectable en cualquier lugar de la aplicación a través del contenedor IoC.

---

## Habilitar LogModule

Importar `nexgou.LogModule` en el módulo raíz:

```go
var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{
        nexgou.ConfigModule,
        nexgou.LogModule, // <-- habilita la inyección de LoggerService
        UserModule,
    },
})
```

---

## Inyectar LoggerService

Declarar `*nexgou.LoggerService` como parámetro del constructor:

```go
type UserService struct {
    log *nexgou.ScopedLogger
}

func NewUserService(logger *nexgou.LoggerService) *UserService {
    return &UserService{
        log: logger.WithContext("UserService"),
    }
}
```

Siempre llamar a `.WithContext(name)` inmediatamente en el constructor para vincular el logger al nombre del servicio.

---

## Niveles de log

Los cuatro niveles estándar, en orden de severidad:

| Nivel | Constante | Cuándo usar |
|:---|:---|:---|
| DEBUG | `nexgou.LevelDebug` | Diagnósticos verbosos de desarrollo |
| INFO | `nexgou.LevelInfo` | Eventos operativos normales |
| WARN | `nexgou.LevelWarn` | Problemas recuperables, comportamiento degradado |
| ERROR | `nexgou.LevelError` | Errores que requieren atención |

```go
log.Debug("cache miss", "key", "user:42")
log.Info("usuario creado", "id", "42", "name", "Alice")
log.Warn("límite de tasa próximo", "ip", "1.2.3.4", "remaining", 5)
log.Error("conexión a base de datos fallida", "err", err)
```

---

## Loggers con alcance

`logger.WithContext(name)` devuelve un `*nexgou.ScopedLogger` que antepone `[name]` a cada línea de log. Esto hace inmediatamente obvio qué componente produjo cada entrada de log.

```go
// En cada servicio, definir el alcance del logger en el momento de construcción
func NewOrderService(logger *nexgou.LoggerService) *OrderService {
    return &OrderService{log: logger.WithContext("OrderService")}
}

func NewPaymentService(logger *nexgou.LoggerService) *PaymentService {
    return &PaymentService{log: logger.WithContext("PaymentService")}
}
```

Salida (formato texto):

```
[INFO]  [OrderService]   pedido creado  id=ORD-001 total=99.99
[INFO]  [PaymentService] pago en cola   id=PAY-001 amount=99.99
[WARN]  [PaymentService] reintento 1/3  id=PAY-001
[ERROR] [PaymentService] pago fallido   id=PAY-001 err=timeout
```

---

## Campos estructurados

Los métodos de log aceptan pares clave-valor después del mensaje:

```go
func (s *UserService) CreateUser(name, email string) (*User, error) {
    s.log.Info("creando usuario", "name", name, "email", email)

    user, err := s.repo.Insert(name, email)
    if err != nil {
        s.log.Error("fallo al crear usuario", "name", name, "err", err)
        return nil, err
    }

    s.log.Info("usuario creado", "id", user.ID, "name", user.Name)
    return user, nil
}
```

Las claves y valores son pares: cada clave debe ir seguida de su valor. El número de argumentos después del mensaje debe ser par.

```go
// Correcto
log.Info("evento", "key1", val1, "key2", val2)

// Incorrecto — número impar de argumentos extra (el último val no tiene clave)
log.Info("evento", "key1", val1, val2)
```

---

## Formatos de salida

### Formato texto (por defecto / desarrollo)

Salida colorizada y legible para humanos en stdout:

```
[INFO]  [UserService] usuario creado  id=42 name=Alice
[WARN]  [AuthService] validación de token lenta  duration=250ms
[ERROR] [DBService]   conexión fallida  err=dial tcp refused
```

### Formato JSON (producción)

JSON legible por máquina, un objeto por línea, adecuado para agregadores de logs (Datadog, Loki, CloudWatch):

```json
{"level":"INFO","context":"UserService","msg":"usuario creado","id":"42","name":"Alice","ts":"2026-05-22T10:30:00Z"}
{"level":"WARN","context":"AuthService","msg":"validación de token lenta","duration":"250ms","ts":"2026-05-22T10:30:01Z"}
```

---

## Variables de entorno

| Variable | Valores | Por defecto | Descripción |
|:---|:---|:---|:---|
| `LOG_LEVEL` | `debug`, `info`, `warn`, `error` | `info` | Nivel mínimo para producir salida |
| `LOG_FORMAT` | `json`, `text` | `text` | Formato de salida |

```bash
LOG_LEVEL=debug LOG_FORMAT=json go run .
```

El logger lee estas variables una vez en el momento de construcción (`NewLoggerService()` se llama durante el arranque de la aplicación).

---

## Referencia de la API

### `LoggerService`

```go
func (l *LoggerService) WithContext(name string) *ScopedLogger
func (l *LoggerService) Debug(msg string, args ...any)
func (l *LoggerService) Info(msg string, args ...any)
func (l *LoggerService) Warn(msg string, args ...any)
func (l *LoggerService) Error(msg string, args ...any)
```

### `ScopedLogger`

```go
func (s *ScopedLogger) Debug(msg string, args ...any)
func (s *ScopedLogger) Info(msg string, args ...any)
func (s *ScopedLogger) Warn(msg string, args ...any)
func (s *ScopedLogger) Error(msg string, args ...any)
```

### Constantes de nivel de log

```go
nexgou.LevelDebug  // 0
nexgou.LevelInfo   // 1
nexgou.LevelWarn   // 2
nexgou.LevelError  // 3
```
