# API publica NexGou 2.0.0

La API publica estable de `2.0.0` vive en el import raiz:

```go
import nexgou "github.com/nexgou/server"
```

No se requiere importar paquetes bajo `src` para crear aplicaciones HTTP.

## Alcance

Incluido en `2.0.0`:

- app lifecycle;
- modulos e inyeccion de dependencias;
- rutas HTTP y versionado;
- middleware global;
- guards, interceptors, pipes y filters;
- config y logger;
- adapter `fasthttp` por medio de `ListenAndServe`;
- samples HTTP y benchmark NexGou.

Fuera de alcance en `2.0.0`:

- WebSocket;
- Server-Sent Events;
- gRPC.

## Aplicacion

```go
app := nexgou.CreateApp(AppModule)
app.Use(nexgou.Recovery())
app.SetFilter(&nexgou.HttpExceptionFilter{})
err := nexgou.ListenAndServe(":3000", app)
```

Tipos y funciones:

- `App`;
- `CreateApp(root IModule) *App`;
- `ListenAndServe(address string, app *App) error`;
- `PrintBanner(config BannerConfig)`;
- `PrintRoutes(app *App)`;
- `app.PrintBanner(config BannerConfig)`;
- `app.PrintRoutes()`;
- `app.WriteRoutes(writer io.Writer)`.

`ListenAndServe` imprime automaticamente el banner de arranque y las rutas registradas antes de iniciar `fasthttp`. El banner usa `NEXGOU_APP_NAME`, `SERVICE_NAME`, `NEXGOU_APP_VERSION`, `SERVICE_VERSION`, `NEXGOU_ENV`, `APP_ENV` y valores por defecto cuando esas variables no estan definidas.

## Modulos

```go
var UserModule = nexgou.Module(nexgou.ModuleOptions{
    Controllers: []any{NewUserController},
    Providers: []any{NewUserService},
})
```

Tipos y funciones:

- `IModule`;
- `ModuleOptions`;
- `Module(options ModuleOptions) IModule`.

## Rutas

```go
nexgou.Get("/users/:id", controller.FindOne).Version("v1")
nexgou.Post("/users", controller.Create).Guard(&AuthGuard{})
```

Helpers:

- `Get`;
- `Post`;
- `Put`;
- `Patch`;
- `Delete`.

Tipos:

- `Route`;
- `Controller`;
- `HandlerFunc`;
- `Context`;
- `H`.

## Pipeline HTTP

Orden de ejecucion:

```txt
middleware global
guard
pipe
interceptor before
handler
interceptor after
filter on error
```

Tipos:

- `MiddlewareFunc`;
- `Guard`;
- `Interceptor`;
- `Pipe`;
- `ExceptionFilter`.

Built-ins:

- `Recovery()`;
- `Logger()`;
- `Cors()`;
- `CorsWithOptions(options CorsOptions)`;
- `SecurityHeaders(options ...SecurityOptions)`;
- `RateLimit(max int, window time.Duration)`;
- `Timeout(duration time.Duration)`;
- `BodyLimit(maxBytes int64)`;
- `HttpExceptionFilter`.

Pipes incluidos:

- `ParseIntPipe`;
- `ParseUUIDPipe`;
- `DefaultValuePipe`.

## Config y logger

Modulos:

- `ConfigModule`;
- `LogModule`.

Tipos:

- `ConfigService`;
- `LoggerService`;
- `ScopedLogger`;
- `LoggerOptions`;
- `LoggerLevel`;
- `LoggerFormat`.

Variables de entorno:

- `LOG_LEVEL`: `debug`, `info`, `warn`, `error`, `silent`;
- `LOG_FORMAT`: `text`, `json`.

## Errores

Helpers:

- `Exception(status, message)`;
- `BadRequestException(message)`;
- `UnauthorizedException(message)`;
- `ForbiddenException(message)`;
- `NotFoundException(message)`;
- `InternalServerErrorException(message)`.

Tipo:

- `HttpException`.
