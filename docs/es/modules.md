# Módulos

> **[← Volver al README](../../README.es.md)**

---

## Tabla de Contenidos

- [¿Qué es un módulo?](#qué-es-un-módulo)
- [Opciones de módulo](#opciones-de-módulo)
- [Módulo raíz](#módulo-raíz)
- [Módulos de funcionalidad](#módulos-de-funcionalidad)
- [Proveedores e inyección de dependencias](#proveedores-e-inyección-de-dependencias)
- [Importar y exportar proveedores](#importar-y-exportar-proveedores)
- [Módulos incorporados](#módulos-incorporados)
- [Árbol de módulos](#árbol-de-módulos)
- [Resolución del contenedor IoC](#resolución-del-contenedor-ioc)

---

## ¿Qué es un módulo?

Un **módulo** es la unidad organizativa fundamental de una aplicación Nexgou. Cada módulo encapsula una parte cohesiva de la aplicación — sus controladores, servicios (proveedores) y los sub-módulos de los que depende.

```
AppModule
├── ConfigModule   (incorporado)
├── LogModule      (incorporado)
├── UserModule
│   ├── UserController
│   └── UserService
└── OrderModule
    ├── OrderController
    ├── OrderService
    └── (importa UserModule.UserService vía Exports)
```

Toda aplicación tiene exactamente **un módulo raíz** que se pasa a `nexgou.CreateApp()`. Los módulos de funcionalidades forman un árbol con raíz en él.

---

## Opciones de módulo

```go
var MyModule = nexgou.Module(nexgou.ModuleOptions{
    Imports:     []nexgou.IModule{...},   // módulos cuyas exportaciones están disponibles aquí
    Controllers: []any{NewMyController},  // funciones constructoras para controladores
    Providers:   []any{NewMyService},     // funciones constructoras para servicios/repos/etc.
    Exports:     []any{NewMyService},     // subconjunto de Providers a exponer a módulos importadores
})
```

| Campo         | Tipo               | Descripción                                                                                         |
| :------------ | :----------------- | :-------------------------------------------------------------------------------------------------- |
| `Imports`     | `[]nexgou.IModule` | Otros módulos a importar. Sus proveedores exportados quedan disponibles para inyección aquí.        |
| `Controllers` | `[]any`            | **Funciones constructoras** (no instancias) para controladores HTTP.                                |
| `Providers`   | `[]any`            | **Funciones constructoras** para servicios, repositorios, factories, etc.                           |
| `Exports`     | `[]any`            | Subconjunto de constructores de `Providers` a poner a disposición de los módulos que importan este. |

> **Importante:** `Controllers` y `Providers` reciben **funciones constructoras**, no instancias de struct. El contenedor IoC las llama y resuelve sus parámetros automáticamente.

---

## Módulo raíz

El módulo raíz es el punto de entrada de la aplicación. Es el único módulo que se pasa a `nexgou.CreateApp()`.

```go
// app.module.go
package main

import (
    nexgou "github.com/nexgou/server"
    "myapp/catalog"
    "myapp/order"
    "myapp/user"
)

var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{
        nexgou.ConfigModule,
        nexgou.LogModule,
        user.UserModule,
        catalog.CatalogModule,
        order.OrderModule,
    },
})
```

El módulo raíz normalmente no tiene `Controllers` ni `Providers` propios — solo importa módulos de funcionalidades.

---

## Módulos de funcionalidad

Un **módulo de funcionalidad** agrupa todo lo relacionado con un dominio específico (usuarios, pedidos, productos, etc.).

```go
// user/user.module.go
package user

import nexgou "github.com/nexgou/server"

var UserModule = nexgou.Module(nexgou.ModuleOptions{
    Controllers: []any{NewUserController},
    Providers:   []any{NewUserService, NewUserRepository},
    Exports:     []any{NewUserService}, // UserService disponible para módulos que importen UserModule
})
```

---

## Proveedores e inyección de dependencias

Los **proveedores** son cualquier valor producido por una función constructora. Se registran en el contenedor IoC del módulo y se inyectan automáticamente donde su tipo aparece como parámetro.

### Reglas de la función constructora

1. Debe ser una función Go simple (cualquier nombre).
2. Sus parámetros son las dependencias a inyectar.
3. Su **primer valor de retorno** es el tipo provisto (utilizado como clave de registro).
4. Se permite un segundo valor de retorno opcional de tipo `error` (provoca panic si es no-nil en el arranque).

```go
// Dos proveedores, uno dependiendo del otro
func NewUserRepository(cfg *nexgou.ConfigService) *UserRepository {
    dsn := cfg.MustGet("DATABASE_URL")
    return &UserRepository{dsn: dsn}
}

func NewUserService(
    repo *UserRepository,
    log  *nexgou.LoggerService,
) *UserService {
    return &UserService{repo: repo, log: log.WithContext("UserService")}
}
```

### Inyección en controladores

```go
func NewUserController(
    svc *UserService,
    cfg *nexgou.ConfigService,
) *UserController {
    return &UserController{svc: svc, cfg: cfg}
}
```

El contenedor resuelve todo el grafo de dependencias — nunca se llama a los constructores manualmente.

---

## Importar y exportar proveedores

Para que un proveedor del módulo **A** esté disponible en el módulo **B**, el módulo A debe **exportarlo** y el módulo B debe **importar** el módulo A.

```
AppModule
└── OrderModule   (importa UserModule)
    └── UserModule (exporta UserService)
```

```go
// user/user.module.go
var UserModule = nexgou.Module(nexgou.ModuleOptions{
    Providers: []any{NewUserService},
    Exports:   []any{NewUserService}, // <-- hace UserService disponible a los importadores
})

// order/order.module.go
var OrderModule = nexgou.Module(nexgou.ModuleOptions{
    Imports:     []nexgou.IModule{UserModule}, // <-- obtiene UserService
    Controllers: []any{NewOrderController},
    Providers:   []any{NewOrderService},
})

// Ahora OrderService puede declarar *UserService como parámetro:
func NewOrderService(userSvc *UserService) *OrderService {
    return &OrderService{userSvc: userSvc}
}
```

Los proveedores que **no** están en `Exports` son privados a su módulo.

---

## Módulos incorporados

Nexgou incluye dos módulos listos para usar. Impórtalos en el módulo raíz para que sus proveedores estén disponibles en toda la aplicación.

### `nexgou.ConfigModule`

Provee: `*nexgou.ConfigService` (acceso tipado y seguro a variables de entorno).

```go
var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{nexgou.ConfigModule, ...},
})

// Inyectar en cualquier proveedor:
func NewMyService(cfg *nexgou.ConfigService) *MyService {
    port := cfg.GetInt("PORT", 3000)
    ...
}
```

Ver [Config](config.md) para la referencia completa.

### `nexgou.LogModule`

Provee: `*nexgou.LoggerService` (logger estructurado con salida JSON/texto).

```go
var AppModule = nexgou.Module(nexgou.ModuleOptions{
    Imports: []nexgou.IModule{nexgou.LogModule, ...},
})

// Inyectar y dar alcance a un contexto:
func NewMyService(log *nexgou.LoggerService) *MyService {
    return &MyService{log: log.WithContext("MyService")}
}
```

Ver [Logger](logger.md) para la referencia completa.

---

## Árbol de módulos

El framework recorre el árbol de módulos **en profundidad** antes de resolver cualquier dependencia. Esto significa que los `Imports` de un módulo son totalmente procesados antes de que el módulo mismo se inicialice — los proveedores importados siempre están listos cuando se ejecutan los constructores del módulo.

```
AppModule
├── ConfigModule   → registra *ConfigService
├── LogModule      → registra *LoggerService
└── UserModule
    ├── (tiene acceso a ConfigService, LogService vía Imports si se declara)
    ├── UserRepository  ← constructor llamado con *ConfigService inyectado
    └── UserController  ← constructor llamado con *UserService inyectado
```

---

## Resolución del contenedor IoC

Cada módulo tiene su propio contenedor. La resolución sigue estos pasos:

1. Buscar el tipo requerido en los proveedores propios del módulo.
2. Si no se encuentra, buscar en los proveedores exportados de todos los módulos importados.
3. Llamar a la función constructora, resolviendo recursivamente sus parámetros.
4. Almacenar el resultado en caché como **singleton** — cada tipo se instancia una vez por módulo.

Si una dependencia no puede resolverse, `CreateApp` produce un panic con un error descriptivo.

```go
// Esto provoca panic en el arranque si *DatabaseService no está registrado
// en ningún lugar del árbol de módulos visible para OrderModule
func NewOrderService(db *DatabaseService) *OrderService { ... }
```

Siempre asegúrate de que cada tipo de dependencia esté provisto localmente o exportado por un módulo importado.
