# Arquitectura: Observability Bridge

## Descripcion

Sistema de ingestion y consulta de metricas/logs para servicios distribuidos con dashboard web incluido. Alternativa ligera a Datadog/New Relic.

## Diagrama de Arquitectura

```
                    CLIENTES
  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
  │ Microservicio│  │   Aplicacion │  │   Servidor   │
  │     Go       │  │    Python    │  │     Node     │
  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
         │                │                │
         └────────────────┼────────────────┘
                          │ POST /api/v1/metrics
                          │ POST /api/v1/logs
                          │ GET  /api/v1/metrics/query
                          │ GET  /api/v1/logs
                          │
                  ┌───────┴──────────────────────┐
                  │   Observability Bridge       │
                  │                               │
                  │  ┌─────────────────────────┐  │
                  │  │      Router (net/http)  │  │
                  │  │  /api/v1/metrics        │  │
                  │  │  /api/v1/logs           │  │
                  │  │  /api/v1/metrics/query  │  │
                  │  │  /api/v1/health         │  │
                  │  │  / → static/index.html  │  │
                  │  └─────────────────────────┘  │
                  │  ┌─────────────────────────┐  │
                  │  │    MetricsService       │  │
                  │  │  - IngestMetric         │  │
                  │  │  - QueryMetrics         │  │
                  │  │  - IngestLog            │  │
                  │  │  - ListLogs             │  │
                  │  └─────────────────────────┘  │
                  └──────┬──────────────────────┘
                         │
                  ┌──────┴──────────┐
                  │   PostgreSQL    │
                  │                 │
                  │  - metrics      │
                  │  - logs         │
                  │  - alerts       │
                  └────────────────┘
```

## Estructura de Carpetas

```
go-observability-bridge/
  cmd/api/              # Punto de entrada
    main.go             # Inicializacion y servidor HTTP
  internal/
    config/             # Variables de entorno
    handler/            # Handlers HTTP
    model/              # Estructuras de dominio
    repository/         # Acceso a PostgreSQL
    service/            # Logica de negocio
  pkg/
    logger/             # Logger estructurado
  static/               # Dashboard web
    index.html          # UI interactiva (sin dependencias JS)
  deploy/
    Containerfile       # Imagen Podman/Docker
    podman-compose.yml  # Stack completo
    schema.sql          # Esquema SQL
  documentacion/
    ARQUITECTURA.md
    API.md
    EJECUCION.md
```

## Capas

### 1. Handler (Infraestructura)
- Recibe requests HTTP
- Valida input basico
- Delega a Service
- Formatea response JSON
- Sirve el dashboard estatico

### 2. Service (Dominio)
- Logica de negocio
- Transformacion de datos
- Coordinacion entre repositorios

### 3. Repository (Persistencia)
- Abstraccion de base de datos
- Queries SQL parametrizadas
- Migracion automatica al iniciar
- Manejo de conexiones

### 4. Model (Entidades)
- Estructuras puras
- Sin dependencias externas
- Contratos entre capas

## Flujo de Datos

### Ingesta de Metrica
1. Cliente HTTP POST a `/api/v1/metrics`
2. Handler valida JSON y deserializa
3. Service asigna timestamp e ID
4. Repository inserta en PostgreSQL
5. Handler responde 201 Created

### Consulta de Metricas
1. Cliente HTTP GET a `/api/v1/metrics/query`
2. Handler parsea query params (metric, aggregation, from, to, label.*)
3. Service construye query SQL
4. Repository ejecuta agregacion
5. Handler responde 200 OK con datos

### Dashboard Web
1. Navegador solicita GET a `/`
2. Servidor responde con `static/index.html`
3. JavaScript del dashboard consulta la API REST
4. Graficos se dibujan con Canvas API (sin librerias externas)

## Decisiones de Diseno

1. **Sin ORM**: SQL nativo para control total y performance
2. **Migracion automatica**: Las tablas se crean al iniciar el servidor
3. **Sin dependencias JS**: El dashboard usa Canvas API nativa
4. **Un solo binario**: El servidor sirve API y dashboard en el mismo puerto
5. **PostgreSQL**: Compatible con TimescaleDB para series temporales
6. **Context con timeouts**: Graceful shutdown y cancelacion
