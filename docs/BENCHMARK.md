# Benchmark competitivo HTTP con k6, Grafana, Docker y SQLite

Documento de referencia para crear un entorno reproducible de benchmark y validar un framework HTTP propio contra competidores de varios lenguajes.

El objetivo no es solo medir `req/s`, sino validar si el producto es competitivo en:

- throughput;
- latencia p50/p95/p99;
- consumo de CPU;
- consumo de memoria;
- estabilidad bajo carga;
- errores;
- rendimiento de CRUD real;
- comportamiento con SQLite;
- facilidad de despliegue;
- observabilidad;
- DX para crear APIs.

---

## Objetivo del benchmark

Crear un laboratorio reproducible donde cada competidor exponga el mismo CRUD HTTP sobre SQLite.

Todos los servicios deben implementar el mismo contrato:

```txt
GET    /health
POST   /users
GET    /users/:id
GET    /users
PUT    /users/:id
DELETE /users/:id
```

Y todos deben cumplir:

```txt
- Mismo payload JSON
- Misma base SQLite
- Misma estructura de tabla
- Mismos índices
- Misma política de errores
- Mismo entorno Docker
- Misma prueba k6
- Mismo dashboard Grafana
- Mismo informe final
```

---

## Stack del benchmark

```txt
Load testing:       k6
Visualización:      Grafana
Métricas:           Prometheus
Contenedores:       Docker Compose
Base de datos:      SQLite
Servicios:          Docker por competidor
Informe:            Markdown + JSON summary + capturas Grafana
```

---

## Competidores iniciales

La selección debe cubrir distintos lenguajes y estilos de framework.

|   # | Competidor                    | Lenguaje              | Runtime/base              | Motivo                                        |
| --: | ----------------------------- | --------------------- | ------------------------- | --------------------------------------------- |
|   1 | Tu framework                  | Go                    | net/http o adapter propio | Producto a validar                            |
|   2 | Fastify                       | TypeScript/JavaScript | Node.js                   | DX, plugins, schemas                          |
|   3 | Gin                           | Go                    | net/http                  | Framework Go popular                          |
|   4 | Fiber                         | Go                    | fasthttp                  | Performance Go estilo Express                 |
|   5 | ASP.NET Core Minimal API      | C#                    | Kestrel                   | Referencia fuerte de arquitectura/performance |
|   6 | Actix Web                     | Rust                  | Tokio                     | Alto rendimiento y buen diseño                |
|   7 | FastAPI                       | Python                | Uvicorn                   | OpenAPI, productividad, typing                |
|   8 | Spring Boot WebFlux o Quarkus | Java                  | JVM                       | Ecosistema enterprise                         |
|   9 | Phoenix                       | Elixir                | Cowboy/BEAM               | Concurrencia y resiliencia                    |
|  10 | Laravel Octane / Swoole       | PHP                   | Swoole                    | Runtime persistente PHP                       |

Para una primera fase, empieza con 4:

```txt
1. NexGou
2. Gin
3. Fastify
4. ASP.NET Core Minimal API
```

Después añade Rust, Python, Java, Elixir y PHP.

### Estado implementado en este repositorio

La primera preparacion del laboratorio deja implementados los directorios existentes bajo `benchmark/`:

| Servicio | Stack | Puerto |
| -------- | ----- | ------ |
| `nexgou` | Go + NexGou | 3001 |
| `fastify` | Node.js + Fastify | 3002 |
| `asp-kestrel` | ASP.NET Core Minimal API | 3003 |
| `actix-web` | Rust + Actix Web | 3004 |
| `hyper` | Rust + Hyper | 3005 |
| `vert-x` | Java + Vert.x Web | 3006 |
| `ajax-php` | PHP + PDO SQLite | 3007 |

Competidores documentados pero no implementados en esta fase, como Gin, Fiber, FastAPI, Phoenix, Quarkus/Spring o Laravel Octane, quedan para una iteracion posterior.

---

## Estructura recomendada del repositorio

En este repositorio, la unidad base del benchmark debe quedar definida dentro de `benchmark/[type_server]`. La idea es separar cada implementación por carpeta y dejar lo compartido bajo un directorio común dentro de `benchmark/`.

```txt
server/
  benchmark/
    _shared/
      docker-compose.yml
      docker-compose.observability.yml
      .env.example

      k6/
        scenarios/
          smoke.js
          crud-read-heavy.js
          crud-write-heavy.js
          crud-mixed.js
          spike.js
          soak.js
        lib/
          client.js
          payloads.js
          checks.js
        results/
          .gitkeep

      observability/
        prometheus/
          prometheus.yml
        grafana/
          provisioning/
            datasources/
              prometheus.yml
            dashboards/
              dashboards.yml
          dashboards/
            k6-http-benchmark.json

      reports/
        template.md
        generated/
          .gitkeep

      scripts/
        run-smoke.sh
        run-crud-mixed.sh
        run-all.sh
        generate-report.sh
        clean.sh

    nexgou/
      Dockerfile
      src/
      data/
        .gitkeep

    go-gin/
      Dockerfile
      go.mod
      main.go
      data/
        .gitkeep

    fastify/
      Dockerfile
      package.json
      src/
        server.ts
      data/
        .gitkeep

    asp-kestrel/
      Dockerfile
      Program.cs
      asp-kestrel.csproj
      data/
        .gitkeep

    actix-web/
      Dockerfile
      Cargo.toml
      src/
        main.rs
      data/
        .gitkeep

    hyper/
      Dockerfile
      src/
        main.rs
      data/
        .gitkeep

    vert-x/
      Dockerfile
      pom.xml
      src/
        main/
          java/
      data/
        .gitkeep

  docs/
    BENCHMARK.md
  src/
  test/
```

---

## Contrato HTTP común

Estado NexGou `2.0.0`: `benchmark/nexgou` ya implementa este contrato con SQLite y adapter `fasthttp` a traves de la API publica raiz.

Validacion local ejecutada:

```txt
go test ./test/benchmark ./benchmark/nexgou/... OK
k6 run -e BASE_URL=http://localhost:3001 --summary-export benchmark/_shared/k6/results/nexgou-smoke.json benchmark/_shared/k6/scenarios/smoke.js OK
k6 run -e BASE_URL=http://localhost:3001 -e VUS=2 -e DURATION=3s --summary-export benchmark/_shared/k6/results/nexgou-crud-mixed.json benchmark/_shared/k6/scenarios/crud-mixed.js OK
```

Resultados NexGou de la corrida corta:

| Escenario             |    Throughput |       p50 |       p95 | Error rate |      Checks |
| --------------------- | ------------: | --------: | --------: | ---------: | ----------: |
| smoke                 |  214.64 req/s | 814.95 us |   1.68 ms |      0.00% |       12/12 |
| crud-mixed, 2 VUs, 3s | 3884.95 req/s |  526.5 us | 735.02 us |      0.00% | 23368/23368 |

### Health check

```http
GET /health
```

Respuesta:

```json
{
	"status": "ok",
	"service": "go-gin",
	"version": "1.0.0"
}
```

---

### Crear usuario

```http
POST /users
Content-Type: application/json
```

Body:

```json
{
	"name": "Sergio Gonzalez",
	"email": "sergio@example.com",
	"age": 34
}
```

Respuesta esperada:

```json
{
	"id": 1,
	"name": "Sergio Gonzalez",
	"email": "sergio@example.com",
	"age": 34,
	"createdAt": "2026-05-25T20:00:00Z",
	"updatedAt": "2026-05-25T20:00:00Z"
}
```

---

### Obtener usuario

```http
GET /users/:id
```

Respuesta:

```json
{
	"id": 1,
	"name": "Sergio Gonzalez",
	"email": "sergio@example.com",
	"age": 34,
	"createdAt": "2026-05-25T20:00:00Z",
	"updatedAt": "2026-05-25T20:00:00Z"
}
```

---

### Listar usuarios

```http
GET /users?limit=20&offset=0
```

Respuesta:

```json
{
	"items": [],
	"limit": 20,
	"offset": 0,
	"total": 0
}
```

---

### Actualizar usuario

```http
PUT /users/:id
Content-Type: application/json
```

Body:

```json
{
	"name": "Sergio Updated",
	"email": "sergio.updated@example.com",
	"age": 35
}
```

---

### Eliminar usuario

```http
DELETE /users/:id
```

Respuesta:

```json
{
	"deleted": true
}
```

---

## Esquema SQLite común

Todos los competidores deben crear exactamente esta tabla.

```sql
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    age INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
```

### Reglas de SQLite

Para que el benchmark sea justo:

```txt
- Un archivo SQLite por contenedor.
- No compartir el mismo archivo SQLite entre competidores.
- Volumen Docker separado por competidor.
- Mismo PRAGMA en todos los servicios.
- Mismo schema.
- Mismo seed inicial.
```

PRAGMA recomendado:

```sql
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA temp_store = MEMORY;
PRAGMA busy_timeout = 5000;
PRAGMA foreign_keys = ON;
```

---

## Docker Compose principal

Archivo: `benchmark/_shared/docker-compose.yml`

```yaml
services:
    nexgou:
        build:
            context: ../nexgou
        container_name: bench-nexgou
        environment:
            PORT: 3001
            DB_PATH: /app/data/db.sqlite
            SERVICE_NAME: nexgou
        volumes:
            - nexgou_data:/app/data
        ports:
            - '3001:3001'
        deploy:
            resources:
                limits:
                    cpus: '1.00'
                    memory: 512M

    go-gin:
        build:
            context: ../go-gin
        container_name: bench-go-gin
        environment:
            PORT: 3002
            DB_PATH: /app/data/db.sqlite
            SERVICE_NAME: go-gin
        volumes:
            - go_gin_data:/app/data
        ports:
            - '3002:3002'
        deploy:
            resources:
                limits:
                    cpus: '1.00'
                    memory: 512M

    fastify:
        build:
            context: ../fastify
        container_name: bench-fastify
        environment:
            PORT: 3003
            DB_PATH: /app/data/db.sqlite
            SERVICE_NAME: fastify
        volumes:
            - fastify_data:/app/data
        ports:
            - '3003:3003'
        deploy:
            resources:
                limits:
                    cpus: '1.00'
                    memory: 512M

    asp-kestrel:
        build:
            context: ../asp-kestrel
        container_name: bench-asp-kestrel
        environment:
            PORT: 3004
            DB_PATH: /app/data/db.sqlite
            SERVICE_NAME: asp-kestrel
            ASPNETCORE_URLS: http://+:3004
        volumes:
            - asp_kestrel_data:/app/data
        ports:
            - '3004:3004'
        deploy:
            resources:
                limits:
                    cpus: '1.00'
                    memory: 512M

volumes:
    nexgou_data:
    go_gin_data:
    fastify_data:
    asp_kestrel_data:
```

> Nota: `deploy.resources` funciona de forma completa en Swarm. Para Docker Compose local moderno puedes usarlo como referencia, pero conviene validar límites reales con `docker stats` o usar flags equivalentes al ejecutar contenedores.

---

## Observabilidad con Prometheus y Grafana

Archivo: `benchmark/_shared/docker-compose.observability.yml`

```yaml
services:
    prometheus:
        image: prom/prometheus:latest
        container_name: bench-prometheus
        command:
            - '--config.file=/etc/prometheus/prometheus.yml'
            - '--web.enable-remote-write-receiver'
            - '--storage.tsdb.retention.time=15d'
        volumes:
            - ./observability/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
            - prometheus_data:/prometheus
        ports:
            - '9090:9090'

    grafana:
        image: grafana/grafana:latest
        container_name: bench-grafana
        environment:
            GF_SECURITY_ADMIN_USER: admin
            GF_SECURITY_ADMIN_PASSWORD: admin
            GF_USERS_ALLOW_SIGN_UP: 'false'
        volumes:
            - grafana_data:/var/lib/grafana
            - ./observability/grafana/provisioning:/etc/grafana/provisioning
            - ./observability/grafana/dashboards:/var/lib/grafana/dashboards
        ports:
            - '3000:3000'
        depends_on:
            - prometheus

volumes:
    prometheus_data:
    grafana_data:
```

---

## Prometheus config

Archivo: `benchmark/_shared/observability/prometheus/prometheus.yml`

```yaml
global:
    scrape_interval: 5s

scrape_configs:
    - job_name: 'prometheus'
      static_configs:
          - targets: ['prometheus:9090']
```

Para k6, usaremos Prometheus Remote Write:

```bash
K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9090/api/v1/write \
k6 run -o experimental-prometheus-rw benchmark/_shared/k6/scenarios/crud-mixed.js
```

---

## Grafana datasource

Archivo: `benchmark/_shared/observability/grafana/provisioning/datasources/prometheus.yml`

```yaml
apiVersion: 1

datasources:
    - name: Prometheus
      type: prometheus
      access: proxy
      url: http://prometheus:9090
      isDefault: true
```

---

## Grafana dashboards provisioning

Archivo: `benchmark/_shared/observability/grafana/provisioning/dashboards/dashboards.yml`

```yaml
apiVersion: 1

providers:
    - name: 'benchmark-dashboards'
      orgId: 1
      folder: 'HTTP Benchmarks'
      type: file
      disableDeletion: false
      updateIntervalSeconds: 10
      allowUiUpdates: true
      options:
          path: /var/lib/grafana/dashboards
```

---

## Variables de entorno

Archivo: `benchmark/_shared/.env.example`

```env
K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9090/api/v1/write

NEXGOU_URL=http://localhost:3001
GO_GIN_URL=http://localhost:3002
FASTIFY_URL=http://localhost:3003
ASP_KESTREL_URL=http://localhost:3004

K6_VUS=50
K6_DURATION=1m
K6_RAMP_UP=30s
K6_RAMP_DOWN=30s
```

---

## Script base k6

Archivo: `benchmark/_shared/k6/lib/payloads.js`

```js
export function userPayload(index) {
	return {
		name: `User ${index}`,
		email: `user.${Date.now()}.${index}.${Math.random()}@example.com`,
		age: 20 + (index % 40),
	};
}
```

Archivo: `benchmark/_shared/k6/lib/checks.js`

```js
import { check } from 'k6';

export function checkJsonResponse(response, expectedStatus = 200) {
	return check(response, {
		[`status is ${expectedStatus}`]: (r) => r.status === expectedStatus,
		'content-type is json': (r) =>
			String(r.headers['Content-Type'] || '').includes(
				'application/json',
			),
		'body is not empty': (r) => r.body && r.body.length > 0,
	});
}
```

Archivo: `benchmark/_shared/k6/lib/client.js`

```js
import http from 'k6/http';
import { checkJsonResponse } from './checks.js';

export class ApiClient {
	constructor(baseUrl, serviceName) {
		this.baseUrl = baseUrl;
		this.serviceName = serviceName;
		this.headers = {
			'Content-Type': 'application/json',
			'X-Benchmark-Service': serviceName,
		};
	}

	health() {
		const res = http.get(`${this.baseUrl}/health`, {
			tags: { endpoint: 'health', service: this.serviceName },
		});

		checkJsonResponse(res, 200);
		return res;
	}

	createUser(payload) {
		const res = http.post(
			`${this.baseUrl}/users`,
			JSON.stringify(payload),
			{
				headers: this.headers,
				tags: { endpoint: 'create_user', service: this.serviceName },
			},
		);

		checkJsonResponse(res, 201);
		return res;
	}

	getUser(id) {
		const res = http.get(`${this.baseUrl}/users/${id}`, {
			tags: { endpoint: 'get_user', service: this.serviceName },
		});

		checkJsonResponse(res, 200);
		return res;
	}

	listUsers() {
		const res = http.get(`${this.baseUrl}/users?limit=20&offset=0`, {
			tags: { endpoint: 'list_users', service: this.serviceName },
		});

		checkJsonResponse(res, 200);
		return res;
	}

	updateUser(id, payload) {
		const res = http.put(
			`${this.baseUrl}/users/${id}`,
			JSON.stringify(payload),
			{
				headers: this.headers,
				tags: { endpoint: 'update_user', service: this.serviceName },
			},
		);

		checkJsonResponse(res, 200);
		return res;
	}

	deleteUser(id) {
		const res = http.del(`${this.baseUrl}/users/${id}`, null, {
			tags: { endpoint: 'delete_user', service: this.serviceName },
		});

		checkJsonResponse(res, 200);
		return res;
	}
}
```

---

## Escenario smoke test

Archivo: `benchmark/_shared/k6/scenarios/smoke.js`

```js
import { sleep } from 'k6';
import { ApiClient } from '../lib/client.js';
import { userPayload } from '../lib/payloads.js';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3001';
const SERVICE_NAME = __ENV.SERVICE_NAME || 'unknown';

export const options = {
	vus: 1,
	duration: '10s',
	thresholds: {
		http_req_failed: ['rate<0.01'],
		http_req_duration: ['p(95)<500'],
	},
};

export default function () {
	const client = new ApiClient(BASE_URL, SERVICE_NAME);

	client.health();

	const createRes = client.createUser(userPayload(__ITER));
	const user = createRes.json();

	client.getUser(user.id);
	client.listUsers();

	client.updateUser(user.id, {
		name: `Updated ${__ITER}`,
		email: user.email,
		age: 30,
	});

	client.deleteUser(user.id);

	sleep(1);
}
```

---

## Escenario CRUD mixto

Archivo: `benchmark/_shared/k6/scenarios/crud-mixed.js`

```js
import { sleep } from 'k6';
import { ApiClient } from '../lib/client.js';
import { userPayload } from '../lib/payloads.js';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3001';
const SERVICE_NAME = __ENV.SERVICE_NAME || 'unknown';

export const options = {
	scenarios: {
		crud_mixed: {
			executor: 'ramping-vus',
			stages: [
				{
					duration: __ENV.K6_RAMP_UP || '30s',
					target: Number(__ENV.K6_VUS || 50),
				},
				{
					duration: __ENV.K6_DURATION || '1m',
					target: Number(__ENV.K6_VUS || 50),
				},
				{ duration: __ENV.K6_RAMP_DOWN || '30s', target: 0 },
			],
		},
	},
	thresholds: {
		http_req_failed: ['rate<0.01'],
		http_req_duration: ['p(95)<300', 'p(99)<800'],
		checks: ['rate>0.99'],
	},
};

export default function () {
	const client = new ApiClient(BASE_URL, SERVICE_NAME);

	const createRes = client.createUser(userPayload(__ITER));
	const user = createRes.json();

	client.getUser(user.id);

	if (__ITER % 3 === 0) {
		client.listUsers();
	}

	if (__ITER % 4 === 0) {
		client.updateUser(user.id, {
			name: `Updated ${__ITER}`,
			email: user.email,
			age: 35,
		});
	}

	if (__ITER % 2 === 0) {
		client.deleteUser(user.id);
	}

	sleep(0.1);
}
```

---

## Escenario read-heavy

Archivo: `benchmark/_shared/k6/scenarios/crud-read-heavy.js`

```js
import { sleep } from 'k6';
import { ApiClient } from '../lib/client.js';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3001';
const SERVICE_NAME = __ENV.SERVICE_NAME || 'unknown';

export const options = {
	scenarios: {
		read_heavy: {
			executor: 'constant-vus',
			vus: Number(__ENV.K6_VUS || 100),
			duration: __ENV.K6_DURATION || '2m',
		},
	},
	thresholds: {
		http_req_failed: ['rate<0.01'],
		http_req_duration: ['p(95)<200', 'p(99)<600'],
		checks: ['rate>0.99'],
	},
};

export default function () {
	const client = new ApiClient(BASE_URL, SERVICE_NAME);

	client.health();
	client.listUsers();

	const id = 1 + (__ITER % 1000);
	client.getUser(id);

	sleep(0.05);
}
```

> Para este escenario necesitas un seed inicial de, por ejemplo, 1.000 usuarios.

---

## Escenario write-heavy

Archivo: `benchmark/_shared/k6/scenarios/crud-write-heavy.js`

```js
import { sleep } from 'k6';
import { ApiClient } from '../lib/client.js';
import { userPayload } from '../lib/payloads.js';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3001';
const SERVICE_NAME = __ENV.SERVICE_NAME || 'unknown';

export const options = {
	scenarios: {
		write_heavy: {
			executor: 'constant-vus',
			vus: Number(__ENV.K6_VUS || 50),
			duration: __ENV.K6_DURATION || '2m',
		},
	},
	thresholds: {
		http_req_failed: ['rate<0.02'],
		http_req_duration: ['p(95)<500', 'p(99)<1200'],
		checks: ['rate>0.98'],
	},
};

export default function () {
	const client = new ApiClient(BASE_URL, SERVICE_NAME);

	const createRes = client.createUser(userPayload(__ITER));
	const user = createRes.json();

	client.updateUser(user.id, {
		name: `Updated ${__ITER}`,
		email: user.email,
		age: 40,
	});

	sleep(0.05);
}
```

---

## Escenario spike

Archivo: `benchmark/_shared/k6/scenarios/spike.js`

```js
import { sleep } from 'k6';
import { ApiClient } from '../lib/client.js';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3001';
const SERVICE_NAME = __ENV.SERVICE_NAME || 'unknown';

export const options = {
	stages: [
		{ duration: '30s', target: 50 },
		{ duration: '15s', target: 500 },
		{ duration: '30s', target: 500 },
		{ duration: '15s', target: 50 },
		{ duration: '30s', target: 0 },
	],
	thresholds: {
		http_req_failed: ['rate<0.05'],
		http_req_duration: ['p(95)<1000', 'p(99)<2500'],
	},
};

export default function () {
	const client = new ApiClient(BASE_URL, SERVICE_NAME);
	client.health();
	client.listUsers();
	sleep(0.05);
}
```

---

## Escenario soak

Archivo: `benchmark/_shared/k6/scenarios/soak.js`

```js
import { sleep } from 'k6';
import { ApiClient } from '../lib/client.js';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3001';
const SERVICE_NAME = __ENV.SERVICE_NAME || 'unknown';

export const options = {
	scenarios: {
		soak: {
			executor: 'constant-vus',
			vus: Number(__ENV.K6_VUS || 50),
			duration: __ENV.K6_DURATION || '30m',
		},
	},
	thresholds: {
		http_req_failed: ['rate<0.01'],
		http_req_duration: ['p(95)<400', 'p(99)<1000'],
		checks: ['rate>0.99'],
	},
};

export default function () {
	const client = new ApiClient(BASE_URL, SERVICE_NAME);
	client.health();
	client.listUsers();
	sleep(0.2);
}
```

---

## Comandos de ejecución

### Levantar servicios

```bash
docker compose -f benchmark/_shared/docker-compose.yml up -d --build
```

### Levantar observabilidad

```bash
docker compose -f benchmark/_shared/docker-compose.observability.yml up -d
```

Grafana:

```txt
http://localhost:3000
user: admin
password: admin
```

Prometheus:

```txt
http://localhost:9090
```

---

## Ejecutar benchmark contra un servicio

### Smoke test

```bash
BASE_URL=http://localhost:3001 \
SERVICE_NAME=nexgou \
k6 run benchmark/_shared/k6/scenarios/smoke.js
```

### CRUD mixto con Prometheus Remote Write

```bash
K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9090/api/v1/write \
BASE_URL=http://localhost:3001 \
SERVICE_NAME=nexgou \
K6_VUS=100 \
K6_DURATION=3m \
k6 run -o experimental-prometheus-rw benchmark/_shared/k6/scenarios/crud-mixed.js
```

### Comparar Fastify

```bash
K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9090/api/v1/write \
BASE_URL=http://localhost:3003 \
SERVICE_NAME=fastify \
K6_VUS=100 \
K6_DURATION=3m \
k6 run -o experimental-prometheus-rw benchmark/_shared/k6/scenarios/crud-mixed.js
```

---

## Script para ejecutar todos los competidores

Archivo: `benchmark/_shared/scripts/run-all.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

SCENARIO="${1:-crud-mixed}"
VUS="${K6_VUS:-100}"
DURATION="${K6_DURATION:-3m}"
PROM_URL="${K6_PROMETHEUS_RW_SERVER_URL:-http://localhost:9090/api/v1/write}"

declare -A SERVICES=(
  ["nexgou"]="http://localhost:3001"
  ["go-gin"]="http://localhost:3002"
  ["fastify"]="http://localhost:3003"
  ["asp-kestrel"]="http://localhost:3004"
)

mkdir -p benchmark/_shared/k6/results

for SERVICE in "${!SERVICES[@]}"; do
  BASE_URL="${SERVICES[$SERVICE]}"
  echo "Running $SCENARIO for $SERVICE at $BASE_URL"

  K6_PROMETHEUS_RW_SERVER_URL="$PROM_URL" \
  BASE_URL="$BASE_URL" \
  SERVICE_NAME="$SERVICE" \
  K6_VUS="$VUS" \
  K6_DURATION="$DURATION" \
  k6 run \
    -o experimental-prometheus-rw \
    --summary-export "benchmark/_shared/k6/results/${SERVICE}-${SCENARIO}.json" \
    "benchmark/_shared/k6/scenarios/${SCENARIO}.js"

  sleep 10
done
```

Permisos:

```bash
chmod +x benchmark/_shared/scripts/run-all.sh
```

Ejecutar:

```bash
K6_VUS=100 K6_DURATION=3m ./benchmark/_shared/scripts/run-all.sh crud-mixed
```

---

## Métricas que debes comparar

### Métricas k6 principales

| Métrica               | Qué indica                        |
| --------------------- | --------------------------------- |
| `http_reqs`           | Total de requests                 |
| `http_req_duration`   | Duración total de request         |
| `http_req_waiting`    | Time to first byte                |
| `http_req_connecting` | Tiempo de conexión                |
| `http_req_blocked`    | Tiempo bloqueado antes de request |
| `http_req_failed`     | Ratio de errores                  |
| `vus`                 | Virtual users activos             |
| `iterations`          | Iteraciones completadas           |
| `checks`              | Validaciones correctas/fallidas   |

### Métricas por endpoint

Usa tags:

```txt
service=nexgou
endpoint=create_user
endpoint=get_user
endpoint=list_users
endpoint=update_user
endpoint=delete_user
```

Comparar:

```txt
- p50 por endpoint
- p95 por endpoint
- p99 por endpoint
- errores por endpoint
- throughput por endpoint
```

### Métricas Docker

Complementa k6 con:

```bash
docker stats
```

Registra:

```txt
- CPU %
- memoria
- net I/O
- block I/O
- pids
```

---

## Dashboard Grafana recomendado

Paneles mínimos:

```txt
1. Requests per second by service
2. HTTP request duration p50 by service
3. HTTP request duration p95 by service
4. HTTP request duration p99 by service
5. Error rate by service
6. Checks success rate by service
7. Requests by endpoint
8. p95 latency by endpoint
9. Active VUs
10. Iterations per second
```

PromQL orientativo:

### Requests per second

```promql
sum by (service) (rate(k6_http_reqs_total[1m]))
```

### Error rate

```promql
sum by (service) (rate(k6_http_req_failed_total[1m]))
```

### p95 latency

```promql
histogram_quantile(
  0.95,
  sum by (service, le) (rate(k6_http_req_duration_seconds_bucket[1m]))
)
```

### p99 latency

```promql
histogram_quantile(
  0.99,
  sum by (service, le) (rate(k6_http_req_duration_seconds_bucket[1m]))
)
```

> Los nombres exactos pueden cambiar según la salida de k6 y la versión. Valida los nombres en Prometheus antes de cerrar el dashboard.

---

## Informe final

Archivo: `benchmark/_shared/reports/template.md`

````md
# Informe de benchmark HTTP

Fecha: YYYY-MM-DD  
Máquina:  
CPU:  
RAM:  
Sistema operativo:  
Docker version:  
k6 version:

---

## Objetivo

Validar el rendimiento y estabilidad del framework propio frente a competidores HTTP.

---

## Competidores evaluados

| Servicio    | Lenguaje   | Runtime           | Puerto | Versión |
| ----------- | ---------- | ----------------- | -----: | ------- |
| nexgou      | Go         | net/http/fasthttp |   3001 |         |
| go-gin      | Go         | net/http          |   3002 |         |
| fastify     | TypeScript | Node.js           |   3003 |         |
| asp-kestrel | C#         | Kestrel           |   3004 |         |

---

## Escenario ejecutado

| Campo         | Valor      |
| ------------- | ---------- |
| Escenario     | crud-mixed |
| VUs           | 100        |
| Duración      | 3m         |
| Ramp-up       | 30s        |
| Ramp-down     | 30s        |
| Base de datos | SQLite WAL |
| CPU limit     | 1 CPU      |
| Memory limit  | 512 MB     |

---

## Resultados globales

| Servicio    | Req/s | p50 | p95 | p99 | Error rate | Checks OK | CPU avg | Mem avg |
| ----------- | ----: | --: | --: | --: | ---------: | --------: | ------: | ------: |
| nexgou      |       |     |     |     |            |           |         |         |
| go-gin      |       |     |     |     |            |           |         |         |
| fastify     |       |     |     |     |            |           |         |         |
| asp-kestrel |       |     |     |     |            |           |         |         |

---

## Resultados por endpoint

| Servicio | Endpoint    | Req/s | p50 | p95 | p99 | Error rate |
| -------- | ----------- | ----: | --: | --: | --: | ---------: |
| nexgou   | create_user |       |     |     |     |            |
| nexgou   | get_user    |       |     |     |     |            |
| nexgou   | list_users  |       |     |     |     |            |
| nexgou   | update_user |       |     |     |     |            |
| nexgou   | delete_user |       |     |     |     |            |

---

## Análisis

### Ganador en throughput

Pendiente.

### Ganador en latencia p95

Pendiente.

### Ganador en estabilidad

Pendiente.

### Ganador en consumo de recursos

Pendiente.

### Problemas detectados

Pendiente.

---

## Decisión

```txt
Aprobado / No aprobado
```
````

### Criterios mínimos para validar el producto

| Criterio                                              | Resultado |
| ----------------------------------------------------- | --------- |
| Error rate < 1%                                       |           |
| p95 < 300ms en CRUD mixto                             |           |
| p99 < 800ms en CRUD mixto                             |           |
| Sin memory leak en soak test                          |           |
| No degradación grave en spike test                    |           |
| Rendimiento igual o superior a Gin en escenario mixto |           |
| DX superior a Gin/Fastify en implementación CRUD      |           |

---

## Conclusión

Pendiente.

````

---

## Criterios de aceptación del producto

Para considerar que tu framework es competitivo en una primera versión:

```txt
Mínimo:
- Error rate < 1%
- p95 < 300ms en CRUD mixto
- p99 < 800ms en CRUD mixto
- Sin crashes
- Sin pérdida de datos
- Sin memory leak evidente en soak test
- Rendimiento cercano a Gin o Fastify

Bueno:
- Supera a Gin en throughput
- Mantiene p95 estable bajo spike
- Consume menos memoria que Fastify
- Tiene mejor DX que Fiber/Gin

Excelente:
- Se acerca a fasthttp/Fiber en read-heavy
- Tiene arquitectura más limpia que Gin/Fiber
- Tiene OpenAPI/validation integrada
- Soporta adapter net/http y fasthttp
- Métricas y tracing integrables desde el inicio
````

---

## Reglas para evitar benchmark falso

No hagas trampas, porque después el producto fallará en producción.

```txt
No comparar endpoints vacíos contra CRUD real.
No permitir que un framework use in-memory DB y otro SQLite.
No activar logs en uno y desactivarlos en otro.
No usar payloads distintos.
No usar pool de conexiones diferente sin documentarlo.
No mezclar debug mode con production mode.
No comparar desde host si un servicio corre fuera de Docker y otro dentro.
No ejecutar todos los servicios a la vez si compiten por CPU, salvo que sea intencionado.
No publicar solo req/s: incluye p95, p99, errores, CPU y memoria.
```

---

## Modos de benchmark

### Modo 1: servicio aislado

Más justo para medir rendimiento individual.

```txt
Levantas 1 competidor + observabilidad + k6
```

### Modo 2: todos los servicios levantados

Útil para comparar operativa, pero puede contaminar resultados por CPU/memoria.

```txt
Levantas todos los competidores a la vez
Ejecutas k6 contra uno cada vez
```

### Modo 3: CI/CD

Útil para evitar regresiones.

```txt
Cada PR ejecuta smoke + benchmark corto
Main ejecuta benchmark completo
```

---

## Integración CI/CD recomendada

Pipeline mínimo:

```txt
1. Build Docker images
2. Start observability
3. Start competitor under test
4. Run smoke test
5. Run crud-mixed short test
6. Export k6 JSON summary
7. Upload report artifact
8. Fail if thresholds fail
```

Ejemplo GitHub Actions simplificado:

```yaml
name: benchmark

on:
    pull_request:
    push:
        branches:
            - main

jobs:
    benchmark:
        runs-on: ubuntu-latest

        steps:
            - uses: actions/checkout@v4

            - name: Start observability
              run: docker compose -f benchmark/_shared/docker-compose.observability.yml up -d

            - name: Start services
              run: docker compose -f benchmark/_shared/docker-compose.yml up -d --build nexgou

            - name: Install k6
              run: |
                  sudo gpg -k
                  curl -s https://dl.k6.io/key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/k6-archive-keyring.gpg
                  echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
                  sudo apt-get update
                  sudo apt-get install -y k6

            - name: Run smoke benchmark
              run: |
                  BASE_URL=http://localhost:3001 \
                  SERVICE_NAME=nexgou \
                  k6 run --summary-export benchmark/_shared/k6/results/smoke.json benchmark/_shared/k6/scenarios/smoke.js

            - name: Upload benchmark results
              uses: actions/upload-artifact@v4
              with:
                  name: benchmark-results
                  path: benchmark/_shared/k6/results/
```

---

## Implementación mínima de un competidor Go Gin

Archivo: `benchmark/go-gin/Dockerfile`

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev sqlite-dev

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN go build -o server main.go

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache sqlite-libs

COPY --from=builder /app/server /app/server

EXPOSE 3002

CMD ["/app/server"]
```

Archivo: `benchmark/go-gin/go.mod`

```go
module benchmark/go-gin

go 1.25

require (
    github.com/gin-gonic/gin v1.10.0
    github.com/mattn/go-sqlite3 v1.14.24
)
```

> Puedes usar `modernc.org/sqlite` si quieres evitar CGO, pero para comparar rendimiento real, documenta claramente qué driver usa cada servicio.

---

## Implementación mínima de Fastify

Archivo: `benchmark/fastify/Dockerfile`

```dockerfile
FROM node:22-alpine

WORKDIR /app

RUN apk add --no-cache sqlite

COPY package.json package-lock.json* ./
RUN npm install

COPY . .

EXPOSE 3003

CMD ["npm", "run", "start"]
```

Archivo: `benchmark/fastify/package.json`

```json
{
	"name": "benchmark-fastify",
	"type": "module",
	"scripts": {
		"start": "node src/server.js"
	},
	"dependencies": {
		"better-sqlite3": "^11.8.1",
		"fastify": "^5.2.1"
	}
}
```

---

## Implementación mínima ASP.NET Core

Archivo: `benchmark/asp-kestrel/Dockerfile`

```dockerfile
FROM mcr.microsoft.com/dotnet/sdk:9.0 AS build

WORKDIR /src

COPY . .

RUN dotnet publish -c Release -o /app/publish

FROM mcr.microsoft.com/dotnet/aspnet:9.0

WORKDIR /app

COPY --from=build /app/publish .

EXPOSE 3004

ENTRYPOINT ["dotnet", "asp-kestrel.dll"]
```

---

## Plan de trabajo recomendado

### Día 1

```txt
- Crear repo
- Añadir Docker Compose
- Añadir observabilidad
- Añadir k6 smoke
- Implementar tu framework + Gin
```

### Día 2

```txt
- Añadir Fastify
- Añadir ASP.NET Core
- Añadir CRUD mixto
- Añadir reports template
```

### Día 3

```txt
- Añadir dashboards Grafana
- Añadir script run-all
- Ejecutar primer informe
- Detectar cuellos de botella
```

### Día 4+

```txt
- Añadir Rust Actix
- Añadir FastAPI
- Añadir Java/Quarkus o Vert.x
- Añadir CI
- Añadir benchmark de regresión
```

---

## Roadmap de validación de tu framework

### Fase 1: correctness

```txt
- Todos los endpoints funcionan
- SQLite persiste correctamente
- Respuestas JSON homogéneas
- Errores normalizados
- k6 checks > 99%
```

### Fase 2: baseline

```txt
- Comparar contra Gin
- Comparar contra Fastify
- Comparar contra ASP.NET Core
```

### Fase 3: tuning

```txt
- Reducir allocations
- Optimizar router
- Mejorar JSON encoding
- Revisar middleware overhead
- Revisar SQLite contention
```

### Fase 4: producto

```txt
- Añadir CLI
- Añadir OpenAPI
- Añadir validation
- Añadir logger
- Añadir tracing
- Añadir docs
```

---

## Qué conclusiones debe responder el informe

El informe final debe responder claramente:

```txt
1. ¿Tu framework es más rápido que Gin?
2. ¿Tu framework tiene mejor p95 que Fastify?
3. ¿Tu framework consume menos memoria que Fastify?
4. ¿Tu framework escala bien con más VUs?
5. ¿SQLite se convierte en cuello de botella?
6. ¿El router propio añade overhead?
7. ¿Los middlewares degradan mucho?
8. ¿El sistema de errores impacta?
9. ¿El adapter net/http es suficiente?
10. ¿Tiene sentido crear adapter fasthttp?
```

---

## Decisión final esperada

El producto estará validado si puedes demostrar algo como:

```txt
Tu framework:
- iguala o supera a Gin en CRUD mixto;
- se acerca a Fiber/fasthttp en read-heavy;
- tiene mejor DX y arquitectura que Gin/Fiber;
- mantiene p95/p99 estables;
- no aumenta errores bajo spike;
- no muestra memory leaks en soak;
- genera informes reproducibles con k6 + Grafana.
```

---

## Fuentes técnicas de referencia

- Grafana k6 documentation: https://grafana.com/docs/k6/latest/
- Running k6: https://grafana.com/docs/k6/latest/get-started/running-k6/
- k6 Docker image: https://hub.docker.com/r/grafana/k6
- k6 Prometheus Remote Write: https://grafana.com/docs/k6/latest/results-output/real-time/prometheus-remote-write/
- k6 thresholds: https://grafana.com/docs/k6/latest/using-k6/thresholds/
- k6 checks: https://grafana.com/docs/k6/latest/using-k6/checks/
- k6 built-in metrics: https://grafana.com/docs/k6/latest/using-k6/metrics/reference/
- Official k6 Prometheus dashboard: https://grafana.com/grafana/dashboards/19665-k6-prometheus/
