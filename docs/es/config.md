# Config

> **[← Volver al README](../../README.es.md)**

---

## Tabla de Contenidos

- [Descripción general](#descripción-general)
- [Habilitar ConfigModule](#habilitar-configmodule)
- [Inyectar ConfigService](#inyectar-configservice)
- [Leer valores](#leer-valores)
- [MustGet (variables requeridas)](#mustget-variables-requeridas)
- [Configuración específica por entorno](#configuración-específica-por-entorno)
- [Referencia de la API](#referencia-de-la-api)

---

## Descripción general

`ConfigService` proporciona acceso tipado y seguro a las variables de entorno. Es provisto por `nexgou.ConfigModule` — un módulo listo para usar que importas en tu módulo raíz. Una vez importado, `*nexgou.ConfigService` se vuelve inyectable en cualquier proveedor o controlador.

---

## Habilitar ConfigModule

Importar `nexgou.ConfigModule` en el módulo raíz:

```go
var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{
        nexgou.ConfigModule, // <-- habilita la inyección de ConfigService en todos lados
        nexgou.LogModule,
        UserModule,
    },
})
```

Eso es todo. No se necesita archivo de configuración — los valores se leen del entorno del SO.

---

## Inyectar ConfigService

Declarar `*nexgou.ConfigService` como parámetro en cualquier constructor:

```go
type DatabaseService struct {
    dsn  string
    pool int
}

func NewDatabaseService(cfg *nexgou.ConfigService) *DatabaseService {
    return &DatabaseService{
        dsn:  cfg.MustGet("DATABASE_URL"),
        pool: cfg.GetInt("DB_POOL_SIZE", 10),
    }
}
```

El contenedor IoC inyecta la instancia singleton de `ConfigService` automáticamente.

---

## Leer valores

### `Get(key string) string`

Devuelve el valor de la variable de entorno `key`. Devuelve un string vacío si la variable no está establecida.

```go
region := cfg.Get("AWS_REGION") // "" si no está establecida
```

### `GetOrDefault(key, fallback string) string`

Devuelve el valor si está establecido, o `fallback` si la variable no está establecida o está vacía.

```go
host := cfg.GetOrDefault("HOST", "localhost")
```

### `GetInt(key string, fallback int) int`

Parsea el valor como un entero. Devuelve `fallback` si la variable no está establecida o no puede parsearse.

```go
port    := cfg.GetInt("PORT", 3000)
workers := cfg.GetInt("WORKER_COUNT", 4)
```

### `GetBool(key string, fallback bool) bool`

Parsea el valor como un booleano. Valores verdaderos: `"1"`, `"true"`, `"yes"` (sin distinción de mayúsculas). Falso: `"0"`, `"false"`, `"no"`. Devuelve `fallback` si no está establecido o no se reconoce.

```go
debug   := cfg.GetBool("DEBUG", false)
verbose := cfg.GetBool("VERBOSE", false)
```

### `MustGet(key string) string`

Devuelve el valor. **Provoca panic en el arranque** si la variable no está establecida. Usar esto para configuración requerida que no tiene sentido tener un valor por defecto.

```go
jwtSecret   := cfg.MustGet("JWT_SECRET")
databaseURL := cfg.MustGet("DATABASE_URL")
```

---

## Configuración específica por entorno

Usar un cargador de `.env` (p. ej. [`godotenv`](https://github.com/joho/godotenv)) antes de iniciar la aplicación para cargar un archivo `.env`:

```go
// main.go
import "github.com/joho/godotenv"

func main() {
    _ = godotenv.Load() // carga .env si existe, ignora el error en producción

    app := nexgou.CreateApp(AppModule)
    // ...
}
```

Archivo `.env`:

```dotenv
PORT=3000
DATABASE_URL=postgres://user:pass@localhost/mydb
JWT_SECRET=super-secret
DEBUG=true
LOG_LEVEL=info
LOG_FORMAT=json
```

---

## Referencia de la API

```go
// Paquete: github.com/nexgou/server (accesible como nexgou.ConfigService)

func (c *ConfigService) Get(key string) string
func (c *ConfigService) GetOrDefault(key, fallback string) string
func (c *ConfigService) GetInt(key string, fallback int) int
func (c *ConfigService) GetBool(key string, fallback bool) bool
func (c *ConfigService) MustGet(key string) string
```

| Método | ¿Provoca panic? | Devuelve si falta |
|:---|:---:|:---|
| `Get` | No | `""` |
| `GetOrDefault` | No | `fallback` |
| `GetInt` | No | `fallback` |
| `GetBool` | No | `fallback` |
| `MustGet` | **Sí** | — |
