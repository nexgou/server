> **[English](#contributing-to-nexgou)** | **[Español](#contribuir-a-nexgou)**

---

# Contributing to Nexgou

Thank you for your interest in contributing to Nexgou. Contributions of all kinds are welcome — bug fixes, new features, documentation improvements, and sample applications.

> Nexgou is under active development. The API is not yet stable, so please open an issue first before starting any significant work.

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Running Tests](#running-tests)
- [Commit Conventions](#commit-conventions)
- [Submitting a Pull Request](#submitting-a-pull-request)
- [Acceptance Criteria](#acceptance-criteria)

---

## Prerequisites

| Tool | Minimum version |
|------|----------------|
| Go | 1.21 |
| Git | 2.34 |

---

## Project Structure

```
.
├── nexgou.go              # Public API — re-exports from src/
├── src/                   # Internal packages
│   ├── app/               # App bootstrap (CreateApp, Listen, ListenGRPC)
│   ├── common/            # Shared types (Context, Route, exceptions, …)
│   ├── core/              # IoC container + Module system
│   ├── router/            # HTTP router
│   ├── middleware/        # Built-in middleware (CORS, rate limit, …)
│   ├── guard/             # Guard execution pipeline
│   ├── interceptor/       # Interceptor pipeline
│   ├── filter/            # Exception filters
│   ├── pipe/              # Pipes (ParseInt, ParseUUID, …)
│   ├── config/            # ConfigService
│   ├── logger/            # LoggerService
│   ├── websocket/         # WebSocket support
│   ├── sse/               # Server-Sent Events support
│   └── grpc/              # gRPC support
├── test/                  # All tests (mirrors src/ structure)
├── samples/               # Runnable example applications
│   ├── api/               # REST + WebSocket + SSE sample
│   ├── chat/              # WebSocket chat sample
│   ├── sse/               # SSE metrics sample
│   └── grpc/              # gRPC greeter sample
└── docs/                  # Extended documentation
```

---

## Getting Started

### 1. Fork and clone

```bash
git clone https://github.com/<your-username>/server.git
cd server
```

### 2. Install git hooks

The repository includes commit-message hooks that enforce Conventional Commits locally before you push.

**Linux / macOS:**
```bash
./scripts/install-hooks.sh
```

**Windows (PowerShell):**
```powershell
./scripts/install-hooks.ps1
```

### 3. Install dependencies

```bash
go mod download
```

### 4. Verify everything is working

```bash
go build ./...
go test ./test/...
```

---

## Development Workflow

1. **Open an issue** — describe the bug or feature you want to work on. This avoids duplicate work and lets us align on the approach before you invest time.
2. **Create a branch** from `main`:
   ```bash
   git checkout -b feat/my-feature
   # or
   git checkout -b fix/the-bug
   ```
3. **Make your changes** and add tests in `test/<package>/`.
4. **Run the full suite** locally (see [Running Tests](#running-tests)).
5. **Commit** following the [Commit Conventions](#commit-conventions).
6. **Open a Pull Request** against `main`.

---

## Running Tests

All tests live in `test/` and are fully isolated from the source packages (external `_test` packages).

```bash
# Run all tests
go test ./test/...

# Run with race detector (required for CI)
go test -race ./test/...

# Run a specific package
go test ./test/router/...

# Run with coverage report
go test -race -coverprofile=coverage.out -covermode=atomic ./test/...
go tool cover -html=coverage.out
```

> **Coverage threshold:** CI enforces a minimum of **80%** total coverage. New code should include tests that keep coverage at or above this threshold.

```bash
# Check current total coverage
go test -coverprofile=coverage.out -covermode=atomic ./test/...
go tool cover -func=coverage.out | grep total
```

---

## Commit Conventions

This project uses [Conventional Commits](https://www.conventionalcommits.org/). The CI pipeline validates every commit message in a PR and the PR title itself.

### Format

```
<type>(<scope>): <short description>

[optional body]

[optional footer: BREAKING CHANGE: ...]
```

### Allowed types

| Type | When to use |
|------|------------|
| `feat` | A new feature |
| `fix` | A bug fix |
| `docs` | Documentation changes only |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `perf` | Performance improvement |
| `test` | Adding or updating tests |
| `build` | Build system or dependency changes |
| `ci` | CI configuration changes |
| `chore` | Maintenance (updating go.sum, tooling, …) |
| `revert` | Reverts a previous commit |
| `style` | Formatting, missing semicolons, etc. |

### Examples

```
feat(grpc): add server-streaming RPC support
fix(router): resolve param extraction on trailing slashes
docs(sse): add reconnect example to README
test(middleware): add race-condition test for rate limiter
feat!: rename ListenGRPC to StartGRPC
```

For **breaking changes**, append `!` after the type/scope or add a `BREAKING CHANGE:` footer.

---

## Submitting a Pull Request

1. Make sure CI is green on your branch before requesting review.
2. Fill in the pull request template — every checkbox is there for a reason.
3. Keep PRs focused: one feature or bug fix per PR. Large refactors should be discussed in an issue first.
4. Update documentation (`docs/`, `README.md`) if your change affects the public API or behavior.
5. Add or update the relevant sample in `samples/` if you are adding a new feature.

A maintainer will review your PR and may request changes. Once approved, it will be squash-merged; the **PR title** becomes the squash commit message, so make sure it follows the commit convention.

---

## Acceptance Criteria

A PR is ready to merge when all of the following are true:

- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `golangci-lint run` passes (run locally with `make lint` or install from [golangci-lint.run](https://golangci-lint.run))
- [ ] `go test -race ./test/...` passes
- [ ] Total coverage does not drop below **80%**
- [ ] Commit messages and PR title follow Conventional Commits
- [ ] Public API changes are reflected in `README.md` and/or `docs/`
- [ ] No new external dependencies added without prior discussion

---

---

# Contribuir a Nexgou

Gracias por tu interés en contribuir a Nexgou. Son bienvenidas contribuciones de todo tipo — corrección de bugs, nuevas funcionalidades, mejoras de documentación y aplicaciones de ejemplo.

> Nexgou está en desarrollo activo. La API todavía no es estable, así que por favor abre un issue antes de comenzar cualquier trabajo significativo.

---

## Tabla de contenidos

- [Requisitos previos](#requisitos-previos)
- [Estructura del proyecto](#estructura-del-proyecto)
- [Primeros pasos](#primeros-pasos)
- [Flujo de desarrollo](#flujo-de-desarrollo)
- [Ejecutar los tests](#ejecutar-los-tests)
- [Convenciones de commits](#convenciones-de-commits)
- [Enviar un Pull Request](#enviar-un-pull-request)
- [Criterios de aceptación](#criterios-de-aceptación)

---

## Requisitos previos

| Herramienta | Versión mínima |
|-------------|---------------|
| Go | 1.21 |
| Git | 2.34 |

---

## Estructura del proyecto

```
.
├── nexgou.go              # API pública — re-exporta desde src/
├── src/                   # Paquetes internos
│   ├── app/               # Bootstrap (CreateApp, Listen, ListenGRPC)
│   ├── common/            # Tipos compartidos (Context, Route, exceptions, …)
│   ├── core/              # Contenedor IoC + sistema de módulos
│   ├── router/            # Router HTTP
│   ├── middleware/        # Middleware integrado (CORS, rate limit, …)
│   ├── guard/             # Pipeline de guards
│   ├── interceptor/       # Pipeline de interceptores
│   ├── filter/            # Filtros de excepciones
│   ├── pipe/              # Pipes (ParseInt, ParseUUID, …)
│   ├── config/            # ConfigService
│   ├── logger/            # LoggerService
│   ├── websocket/         # Soporte WebSocket
│   ├── sse/               # Soporte Server-Sent Events
│   └── grpc/              # Soporte gRPC
├── test/                  # Todos los tests (espeja la estructura de src/)
├── samples/               # Aplicaciones de ejemplo ejecutables
│   ├── api/               # Ejemplo REST + WebSocket + SSE
│   ├── chat/              # Ejemplo chat WebSocket
│   ├── sse/               # Ejemplo métricas SSE
│   └── grpc/              # Ejemplo gRPC greeter
└── docs/                  # Documentación extendida
```

---

## Primeros pasos

### 1. Fork y clona el repositorio

```bash
git clone https://github.com/<tu-usuario>/server.git
cd server
```

### 2. Instala los git hooks

El repositorio incluye hooks de mensajes de commit que refuerzan Conventional Commits localmente antes de hacer push.

**Linux / macOS:**
```bash
./scripts/install-hooks.sh
```

**Windows (PowerShell):**
```powershell
./scripts/install-hooks.ps1
```

### 3. Instala las dependencias

```bash
go mod download
```

### 4. Verifica que todo funciona

```bash
go build ./...
go test ./test/...
```

---

## Flujo de desarrollo

1. **Abre un issue** — describe el bug o feature en el que quieres trabajar. Esto evita trabajo duplicado y permite alinear el enfoque antes de invertir tiempo.
2. **Crea una rama** desde `main`:
   ```bash
   git checkout -b feat/mi-funcionalidad
   # o
   git checkout -b fix/el-bug
   ```
3. **Realiza tus cambios** y añade tests en `test/<paquete>/`.
4. **Ejecuta la suite completa** localmente (ver [Ejecutar los tests](#ejecutar-los-tests)).
5. **Confirma los cambios** siguiendo las [Convenciones de commits](#convenciones-de-commits).
6. **Abre un Pull Request** contra `main`.

---

## Ejecutar los tests

Todos los tests residen en `test/` y están completamente aislados de los paquetes fuente (paquetes externos con sufijo `_test`).

```bash
# Ejecutar todos los tests
go test ./test/...

# Ejecutar con detector de carreras (requerido por CI)
go test -race ./test/...

# Ejecutar un paquete específico
go test ./test/router/...

# Ejecutar con informe de cobertura
go test -race -coverprofile=coverage.out -covermode=atomic ./test/...
go tool cover -html=coverage.out
```

> **Umbral de cobertura:** el CI impone un mínimo del **80%** de cobertura total. El código nuevo debe incluir tests que mantengan la cobertura en ese umbral o por encima.

```bash
# Verificar cobertura total actual
go test -coverprofile=coverage.out -covermode=atomic ./test/...
go tool cover -func=coverage.out | grep total
```

---

## Convenciones de commits

Este proyecto usa [Conventional Commits](https://www.conventionalcommits.org/es/v1.0.0/). El pipeline de CI valida cada mensaje de commit en un PR y el propio título del PR.

### Formato

```
<tipo>(<alcance>): <descripción corta>

[cuerpo opcional]

[pie opcional: BREAKING CHANGE: ...]
```

### Tipos permitidos

| Tipo | Cuándo usarlo |
|------|--------------|
| `feat` | Una nueva funcionalidad |
| `fix` | Corrección de un bug |
| `docs` | Solo cambios en documentación |
| `refactor` | Cambio de código que no corrige un bug ni añade funcionalidad |
| `perf` | Mejora de rendimiento |
| `test` | Añadir o actualizar tests |
| `build` | Cambios en el sistema de build o dependencias |
| `ci` | Cambios en la configuración de CI |
| `chore` | Mantenimiento (actualizar go.sum, herramientas, …) |
| `revert` | Revierte un commit anterior |
| `style` | Formato, punto y coma olvidado, etc. |

### Ejemplos

```
feat(grpc): añadir soporte para RPCs server-streaming
fix(router): resolver extracción de params con trailing slash
docs(sse): añadir ejemplo de reconexión al README
test(middleware): añadir test de condición de carrera en rate limiter
feat!: renombrar ListenGRPC a StartGRPC
```

Para **cambios con ruptura de compatibilidad**, añade `!` tras el tipo/alcance o incluye un pie `BREAKING CHANGE:`.

---

## Enviar un Pull Request

1. Asegúrate de que el CI está en verde en tu rama antes de pedir revisión.
2. Rellena la plantilla del pull request — cada casilla tiene su razón de ser.
3. Mantén los PRs enfocados: una funcionalidad o corrección de bug por PR. Las refactorizaciones grandes deben discutirse en un issue primero.
4. Actualiza la documentación (`docs/`, `README.md`) si tu cambio afecta a la API pública o al comportamiento.
5. Añade o actualiza el ejemplo correspondiente en `samples/` si estás añadiendo una nueva funcionalidad.

Un mantenedor revisará tu PR y puede solicitar cambios. Una vez aprobado, se hará squash-merge; el **título del PR** se convierte en el mensaje del commit final, así que asegúrate de que sigue la convención.

---

## Criterios de aceptación

Un PR está listo para fusionar cuando todo lo siguiente es verdad:

- [ ] `go build ./...` pasa
- [ ] `go vet ./...` pasa
- [ ] `golangci-lint run` pasa (ejecutar localmente con `make lint` o instalar desde [golangci-lint.run](https://golangci-lint.run))
- [ ] `go test -race ./test/...` pasa
- [ ] La cobertura total no cae por debajo del **80%**
- [ ] Los mensajes de commit y el título del PR siguen Conventional Commits
- [ ] Los cambios en la API pública se reflejan en `README.md` y/o `docs/`
- [ ] No se añaden nuevas dependencias externas sin discusión previa
