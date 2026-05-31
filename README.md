# Observability Bridge

Sistema de observabilidad ligero para métricas y logs con dashboard web incluido. Alternativa open-source a Datadog/New Relic para equipos pequeños.

## Caracteristicas

- Recoleccion de metricas (valores numericos con labels)
- Centralizacion de logs estructurados
- Consulta de metricas con agregaciones (avg, sum, max, min, count)
- Dashboard web interactivo con graficos y tooltips
- Persistencia en PostgreSQL con migracion automatica
- Filtros por labels, rango de tiempo y agregaciones
- Graceful shutdown
- CORS habilitado para integraciones externas

## Arquitectura

```
Cliente HTTP / Dashboard Web
        |
    Router (net/http)
        |
    Service (logica de negocio)
        |
    Repository (PostgreSQL)
        |
    Base de Datos
```

## Estructura

```
go-observability-bridge/
  cmd/api/              # Punto de entrada del servidor
  internal/
    config/             # Variables de entorno
    handler/            # Handlers HTTP
    model/              # Estructuras de dominio
    repository/         # Acceso a PostgreSQL
    service/            # Logica de negocio
  pkg/logger/           # Logger estructurado
  static/               # Dashboard web (HTML+CSS+JS)
  deploy/
    Containerfile       # Imagen Podman/Docker
    podman-compose.yml  # Stack completo
    schema.sql          # Esquema SQL
  Makefile              # Comandos utiles
```

## Quick Start

### Requisitos

- Go 1.23+
- PostgreSQL 15+ (o Neon)

### 1. Configurar

```bash
cp .env.example .env
# Editar DATABASE_URL en .env
```

### 2. Ejecutar

```bash
make dev
```

### 3. Abrir

- **Dashboard**: http://localhost:8080
- **API**: http://localhost:8080/api/v1/health

## API

| Metodo | Ruta                    | Descripcion              |
|--------|-------------------------|--------------------------|
| GET    | /api/v1/health          | Health check             |
| POST   | /api/v1/metrics         | Enviar metrica           |
| POST   | /api/v1/logs            | Enviar log               |
| GET    | /api/v1/logs            | Recuperar logs recientes |
| GET    | /api/v1/metrics/query   | Consultar metricas       |

### Query params para /api/v1/metrics/query

| Parametro    | Ejemplo                          | Descripcion              |
|--------------|----------------------------------|--------------------------|
| metric       | cpu_usage                        | Nombre de la metrica     |
| aggregation  | avg, sum, max, min, count        | Agregacion               |
| from         | 2026-01-01T00:00:00Z             | Inicio del rango         |
| to           | 2026-01-02T00:00:00Z             | Fin del rango            |
| label.host   | server-01                        | Filtrar por label        |

### Ejemplos

```bash
# Health
curl http://localhost:8080/api/v1/health

# Enviar metrica
curl -X POST http://localhost:8080/api/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{"name":"cpu_usage","value":78.5,"labels":{"host":"server-01"}}'

# Enviar log
curl -X POST http://localhost:8080/api/v1/logs \
  -H "Content-Type: application/json" \
  -d '{"level":"ERROR","message":"DB timeout","service":"api"}'

# Consultar metricas
curl "http://localhost:8080/api/v1/metrics/query?metric=cpu_usage"

# Con agregacion
curl "http://localhost:8080/api/v1/metrics/query?metric=cpu_usage&aggregation=avg"

# Con filtro de label
curl "http://localhost:8080/api/v1/metrics/query?metric=cpu_usage&label.host=server-01"
```

## Comandos Make

| Comando        | Descripcion                     |
|----------------|---------------------------------|
| make dev       | Ejecutar en desarrollo          |
| make build     | Compilar binario                |
| make test      | Ejecutar tests                  |
| make fmt       | Formatear codigo                |
| make deps      | Descargar dependencias          |
| make check     | Verificar compilacion           |
| make compose-up| Iniciar stack con Podman Compose|
| make compose-down| Detener stack                 |

## Tecnologias

- **Go 1.24+**: Lenguaje principal
- **PostgreSQL 15+**: Base de datos
- **Podman**: Containerizacion
- **net/http**: Router HTTP nativo
- **lib/pq**: Driver PostgreSQL
- **Canvas API**: Graficos en el dashboard (sin dependencias JS)

## Roadmap

- [x] Estructura inicial y arquitectura
- [x] Endpoints completos con PostgreSQL
- [x] Dashboard web interactivo
- [x] Graficos con tooltips
- [x] Recuperacion de logs persistidos
- [ ] Autenticacion con API keys
- [ ] Sistema de alertas con thresholds
- [ ] Tests unitarios y de integracion
- [ ] CI/CD con GitHub Actions
- [ ] Deploy en GCP Cloud Run

## Licencia

MIT
