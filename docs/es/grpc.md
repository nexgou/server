# gRPC

> **[← Volver al README](../../README.es.md)**

---

## Tabla de Contenidos

- [Descripción general](#descripción-general)
- [Cómo funciona (sin .proto requerido)](#cómo-funciona-sin-proto-requerido)
- [Tipos de mensajes](#tipos-de-mensajes)
- [Interfaz y descriptor de servicio](#interfaz-y-descriptor-de-servicio)
- [Implementación del controlador](#implementación-del-controlador)
- [RPCs con streaming del servidor](#rpcs-con-streaming-del-servidor)
- [Registro del módulo](#registro-del-módulo)
- [Iniciar ambos servidores](#iniciar-ambos-servidores)
- [Guards en rutas gRPC](#guards-en-rutas-grpc)
- [GRPCContext](#grpccontext)
- [Combinar HTTP + gRPC](#combinar-http--grpc)
- [Probar con grpcurl](#probar-con-grpcurl)
- [Referencia de la API](#referencia-de-la-api)

---

## Descripción general

Nexgou proporciona soporte de primera clase para gRPC sin requerir archivos `.proto` ni generación de código con `protoc`. Los descriptores de servicio se escriben en Go puro usando `grpc.ServiceDesc`, haciéndolos **compatibles a nivel de wire** con cualquier cliente gRPC estándar. Los Guards funcionan en RPCs unarias de forma transparente, recibiendo metadatos gRPC mapeados a cabeceras HTTP.

---

## Cómo funciona (sin `.proto` requerido)

Normalmente, `protoc` genera:
1. Tipos de struct de mensajes desde tu archivo `.proto`
2. Un struct `ServiceDesc` describiendo tu servicio

Nexgou te permite escribir ambos a mano. El formato wire resultante es idéntico al código generado por protoc porque usa los mismos paquetes `google.golang.org/grpc` y `google.golang.org/protobuf` bajo el capó.

**Estructura por servicio gRPC:**

```
myservice/
├── service.go      ← tipos de mensajes + ServiceDesc + interfaz handler
├── controller.go   ← implementa la interfaz handler + Register()/RegisterGRPC()
└── module.go       ← nexgou.Module(...)
```

---

## Tipos de mensajes

Cada mensaje protobuf necesita cuatro elementos para ser compatible a nivel de wire:

```go
import "google.golang.org/protobuf/runtime/protoimpl"

type HelloRequest struct {
    state         protoimpl.MessageState `protogen:"open.v1"`
    Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
    unknownFields protoimpl.UnknownFields
    sizeCache     protoimpl.SizeCache
}

// Requerido por la interfaz proto.Message
func (x *HelloRequest) ProtoMessage() {}
func (x *HelloRequest) Reset()        { *x = HelloRequest{} }
func (x *HelloRequest) String() string { return x.Name }
```

El struct tag `protobuf:"..."` refleja lo que protoc generaría para un campo proto3:

| Componente del tag | Significado |
|:---|:---|
| `bytes` | Tipo wire (bytes para strings, mensajes) |
| `1` | Número de campo (debe coincidir con la definición `.proto` si se interopera) |
| `opt` | Campo opcional |
| `name=name` | Nombre del campo tal como aparece en la definición proto |
| `proto3` | Usa semántica proto3 (omisión del valor cero) |

---

## Interfaz y descriptor de servicio

Define una interfaz que tu controlador implementará, luego conéctala en un `grpc.ServiceDesc`:

### RPC unaria

```go
// Firma de bajo nivel del handler unario
func _Greeter_SayHello_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
    in := new(HelloRequest)
    if err := dec(in); err != nil {
        return nil, err
    }
    if interceptor == nil {
        return srv.(GreeterHandler).SayHello(ctx, in)
    }
    info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/greeter.Greeter/SayHello"}
    handler := func(ctx context.Context, req any) (any, error) {
        return srv.(GreeterHandler).SayHello(ctx, req.(*HelloRequest))
    }
    return interceptor(ctx, in, info, handler)
}

type GreeterHandler interface {
    SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error)
}

var Greeter_ServiceDesc = grpc.ServiceDesc{
    ServiceName: "greeter.Greeter",     // debe ser "paquete.NombreServicio"
    HandlerType: (*GreeterHandler)(nil),
    Methods: []grpc.MethodDesc{
        {
            MethodName: "SayHello",
            Handler:    _Greeter_SayHello_Handler,
        },
    },
    Streams: []grpc.StreamDesc{},
}
```

---

## RPCs con streaming del servidor

Para streaming del lado del servidor, usar `grpc.StreamDesc` y `grpc.ServerStream`:

```go
// Handler de stream de bajo nivel
func _Greeter_SayHelloStream_Handler(srv any, stream grpc.ServerStream) error {
    in := new(HelloRequest)
    if err := stream.RecvMsg(in); err != nil {
        return err
    }
    return srv.(StreamingGreeterHandler).SayHelloStream(in, stream)
}

type StreamingGreeterHandler interface {
    SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error)
    SayHelloStream(req *HelloRequest, stream grpc.ServerStream) error
}

var Greeter_ServiceDesc = grpc.ServiceDesc{
    ServiceName: "greeter.Greeter",
    HandlerType: (*StreamingGreeterHandler)(nil),
    Methods: []grpc.MethodDesc{
        {MethodName: "SayHello", Handler: _Greeter_SayHello_Handler},
    },
    Streams: []grpc.StreamDesc{
        {
            StreamName:    "SayHelloStream",
            Handler:       _Greeter_SayHelloStream_Handler,
            ServerStreams: true,
        },
    },
}
```

Implementar el handler de streaming en el controlador:

```go
func (c *GreeterController) SayHelloStream(req *HelloRequest, stream grpc.ServerStream) error {
    for i := 0; i < 5; i++ {
        reply := &HelloReply{Message: fmt.Sprintf("Hola %s (%d)", req.Name, i+1)}
        if err := stream.SendMsg(reply); err != nil {
            return err
        }
        time.Sleep(500 * time.Millisecond)
    }
    return nil
}
```

---

## Implementación del controlador

```go
// greeter/controller.go
package greeter

import (
    "context"
    "fmt"

    nexgou "github.com/nexgou/server"
)

type GreeterController struct {
    log *nexgou.ScopedLogger
}

func NewGreeterController(log *nexgou.LoggerService) *GreeterController {
    return &GreeterController{log: log.WithContext("GreeterController")}
}

// Verificación de salud HTTP (ruta compañera opcional)
func (c *GreeterController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.Get("/health", c.Health),
    }
}

func (c *GreeterController) Health(ctx *nexgou.Context) error {
    return ctx.JSON(200, nexgou.H{"status": "ok"})
}

// Rutas gRPC
func (c *GreeterController) RegisterGRPC() []nexgou.GRPCRoute {
    return []nexgou.GRPCRoute{
        nexgou.GRPC(Greeter_ServiceDesc, c).
            Guard(&AuthGuard{}).
            Version("v1"),
    }
}

// Implementación de RPC unaria
func (c *GreeterController) SayHello(_ context.Context, req *HelloRequest) (*HelloReply, error) {
    c.log.Info("SayHello llamado", "name", req.Name)
    return &HelloReply{Message: fmt.Sprintf("Hola, %s!", req.Name)}, nil
}
```

---

## Registro del módulo

```go
// greeter/module.go
package greeter

import nexgou "github.com/nexgou/server"

var Module = nexgou.Module(nexgou.ModuleOptions{
    Controllers: []any{NewGreeterController},
})
```

El framework auto-detecta `GRPCController` comprobando la existencia del método `RegisterGRPC()`.

---

## Iniciar ambos servidores

Ejecutar gRPC y HTTP en puertos separados de forma concurrente:

```go
// main.go
func main() {
    app := nexgou.CreateApp(AppModule)

    app.Use(middleware.Recovery())
    app.Use(middleware.Logger())
    app.SetFilter(&filter.HttpExceptionFilter{})

    // gRPC en :50051 (no bloqueante)
    go func() {
        if err := app.ListenGRPC(50051); err != nil {
            log.Fatalf("error del servidor gRPC: %v", err)
        }
    }()

    // HTTP en :3000 (bloqueante)
    log.Fatal(app.Listen(3000))
}
```

Ambos servidores comparten el mismo árbol de módulos y contenedor DI — los servicios se instancian una vez y se comparten.

---

## Guards en rutas gRPC

Los guards asociados a un `GRPCRoute` se ejecutan en cada RPC **unaria** de ese servicio. Reciben un `*nexgou.Context` con:

- `ctx.Header("key")` → lee el valor de metadatos entrantes gRPC para `key`
- `ctx.Method()` → devuelve el path completo del método gRPC (p. ej. `/greeter.Greeter/SayHello`)

Esto significa que el **mismo guard** funciona tanto para HTTP como para gRPC sin modificaciones:

```go
type AuthGuard struct{}

func (g *AuthGuard) CanActivate(ctx *nexgou.Context) (bool, error) {
    token := ctx.Header("authorization") // funciona para metadatos HTTP y gRPC
    return validateToken(token), nil
}

// Funciona en HTTP:
nexgou.Get("/users", c.List).Guard(&AuthGuard{})

// Funciona en gRPC:
nexgou.GRPC(MyService_ServiceDesc, c).Guard(&AuthGuard{})
```

> Nota: Los guards en `GRPCRoute` solo aplican a RPCs **unarias**. Las RPCs de streaming no son interceptadas actualmente por guards.

---

## GRPCContext

Para casos de uso avanzados (p. ej. middleware que necesita inspeccionar metadatos gRPC), usar `nexgou.GRPCContext`:

```go
func (c *MyController) MyRPC(ctx context.Context, req *MyRequest) (*MyResponse, error) {
    // El context.Context estándar lleva metadatos gRPC
    md, ok := metadata.FromIncomingContext(ctx)
    if ok {
        tokens := md.Get("authorization")
    }
    return &MyResponse{}, nil
}
```

Cuando los guards se ejecutan, Nexgou envuelve internamente el contexto gRPC en un `*nexgou.Context`. No necesitas usar `GRPCContext` directamente en la mayoría de los casos.

---

## Combinar HTTP + gRPC

Un solo controlador puede implementar ambos:

```go
func (c *GreeterController) Register() []nexgou.Route {
    return []nexgou.Route{
        nexgou.Get("/health", c.Health),
        nexgou.Get("/greet/:name", c.HTTPGreet),
    }
}

func (c *GreeterController) RegisterGRPC() []nexgou.GRPCRoute {
    return []nexgou.GRPCRoute{
        nexgou.GRPC(Greeter_ServiceDesc, c),
    }
}
```

Ambos se registran desde la misma entrada `Controllers` en el módulo.

---

## Probar con grpcurl

[grpcurl](https://github.com/fullstorydev/grpcurl) es la forma más fácil de probar servicios gRPC desde la línea de comandos.

```bash
# Listar servicios (requiere reflexión del servidor — añadir grpc/reflection al servidor para esto)
grpcurl -plaintext localhost:50051 list

# Llamar a una RPC unaria
grpcurl -plaintext -d '{"name": "World"}' localhost:50051 greeter.Greeter/SayHello

# Con metadatos (se mapean a metadatos gRPC, leídos por guards via ctx.Header)
grpcurl -plaintext \
  -H 'authorization: Bearer my-token' \
  -d '{"name": "Alice"}' \
  localhost:50051 greeter.Greeter/SayHello
```

---

## Referencia de la API

### `nexgou.GRPC(desc, impl)`

```go
func GRPC(desc GRPCServiceDesc, impl any) GRPCRoute
```

Crea un `GRPCRoute` vinculando un descriptor de servicio a su implementación.

### Métodos de `GRPCRoute`

| Método | Devuelve | Descripción |
|:---|:---|:---|
| `.Guard(guards ...Guard)` | `GRPCRoute` | Asociar guards a todas las RPCs unarias (fluido) |
| `.Version(v string)` | `GRPCRoute` | Establecer una etiqueta de versión (informativo, fluido) |
| `.Ver()` | `string` | Devolver la etiqueta de versión |

### Métodos de `App`

| Método | Descripción |
|:---|:---|
| `app.ListenGRPC(port int, ip ...string) error` | Iniciar el servidor gRPC (bloqueante) |

### Alias de tipos

| Alias | Tipo real | Descripción |
|:---|:---|:---|
| `nexgou.GRPCServiceDesc` | `grpc.ServiceDesc` | Descriptor de servicio |
| `nexgou.GRPCServerStream` | `grpc.ServerStream` | Interfaz de streaming del lado del servidor |
| `nexgou.GRPCServer` | `*nexgougrpc.GRPCServer` | El wrapper del servidor gRPC subyacente |
