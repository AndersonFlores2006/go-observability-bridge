# Guía de Ejecución: Go Observability Bridge

## Requisitos Previos

- [ ] Go 1.23+ instalado
- [ ] Podman (o Docker) instalado
- [ ] PostgreSQL 15+ (local o Neon)
- [ ] `make` (opcional, para shortcuts)

## Verificar Instalaciones

```bash
# Verificar Go
go version

# Verificar Podman
podman --version

# Verificar PostgreSQL (si local)
psql --version
```

## Configuración

### 1. Variables de Entorno

Copiar el archivo de ejemplo:

```bash
cp .env.example .env
```

Editar `.env` con tus valores:

```env
# Puerto del servidor
PORT=8080

# Base de datos PostgreSQL (Neon recomendado)
DATABASE_URL=postgresql://usuario:password@host.neon.tech/observability?sslmode=require

# Nivel de logs: debug, info, warn, error
LOG_LEVEL=info
```

### 2. Base de Datos

#### Opción A: Neon (Recomendado - Gratis)
1. Ir a https://neon.tech
2. Crear proyecto nuevo
3. Copiar "Connection string" y pegar en `.env`

#### Opción B: Local con Podman

```bash
# Iniciar PostgreSQL local
podman run -d \
  --name observability-db \
  -e POSTGRES_USER=admin \
  -e POSTGRES_PASSWORD=admin123 \
  -e POSTGRES_DB=observability \
  -p 5432:5432 \
  postgres:15-alpine

# URL para .env:
# DATABASE_URL=postgresql://admin:admin123@localhost:5432/observability?sslmode=disable
```

## Ejecución Local (Desarrollo)

### Paso 1: Instalar Dependencias

```bash
cd go-observability-bridge
go mod tidy
```

### Paso 2: Iniciar Servidor

```bash
# Usando variables de entorno del archivo .env
export $(cat .env | xargs)

# O exportar manualmente:
export PORT=8080
export DATABASE_URL="postgresql://admin:admin123@localhost:5432/observability?sslmode=disable"
export LOG_LEVEL=debug

# Ejecutar
go run cmd/api/main.go
```

Servidor iniciará en: http://localhost:8080

### Paso 3: Verificar Health Check

```bash
curl http://localhost:8080/api/v1/health
```

Respuesta esperada:
```json
{
  "status": "ok",
  "timestamp": "2025-01-XXTXX:XX:XXZ",
  "service": "go-observability-bridge",
  "version": "1.0.0"
}
```

## Ejecución con Podman (Producción Local)

### Paso 1: Construir Imagen

```bash
cd go-observability-bridge

# Construir con Podman
podman build -t observability-bridge:latest -f Containerfile .

# Verificar imagen creada
podman images | grep observability
```

### Paso 2: Ejecutar Contenedor

```bash
# Asegurar que PostgreSQL esté corriendo
podman run -d \
  --name observability-db \
  -e POSTGRES_USER=admin \
  -e POSTGRES_PASSWORD=admin123 \
  -e POSTGRES_DB=observability \
  -p 5432:5432 \
  postgres:15-alpine

# Esperar 10 segundos para que PostgreSQL inicie
sleep 10

# Ejecutar aplicación
podman run -d \
  --name observability-bridge \
  -p 8080:8080 \
  -e PORT=8080 \
  -e DATABASE_URL="postgresql://admin:admin123@host.containers.internal:5432/observability?sslmode=disable" \
  -e LOG_LEVEL=info \
  observability-bridge:latest
```

### Paso 3: Verificar Logs

```bash
podman logs -f observability-bridge
```

### Paso 4: Detener y Limpiar

```bash
# Detener contenedores
podman stop observability-bridge
podman stop observability-db

# Eliminar contenedores
podman rm observability-bridge
podman rm observability-db
```

## Testing de Endpoints

### Enviar Métrica

```bash
curl -X POST http://localhost:8080/api/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cpu_usage",
    "value": 78.5,
    "labels": {
      "host": "server-01",
      "region": "us-east"
    }
  }'
```

### Consultar Métricas

```bash
curl "http://localhost:8080/api/v1/metrics/query?metric=cpu_usage"
```

### Consultar Métricas con Agregación

```bash
curl "http://localhost:8080/api/v1/metrics/query?metric=cpu_usage&aggregation=avg"
```

### Recuperar Logs Recientes

```bash
curl "http://localhost:8080/api/v1/logs?limit=10"
```

### Dashboard Web

Abrir en el navegador: http://localhost:8080/

El dashboard permite:
- Ver metricas en grafico interactivo (con tooltip al hacer hover)
- Filtrar por nombre de metrica, agregacion, y rango de tiempo
- Ver logs recientes en tabla
- Enviar metricas de prueba desde el formulario
- Enviar logs de prueba desde el formulario

### Enviar Log

```bash
curl -X POST http://localhost:8080/api/v1/logs \
  -H "Content-Type: application/json" \
  -d '{
    "level": "ERROR",
    "message": "Database connection timeout",
    "service": "api-gateway",
    "metadata": {
      "request_id": "abc-123",
      "duration_ms": "5000"
    }
  }'
```

## Comandos Útiles

```bash
# Ver todos los contenedores corriendo
podman ps

# Ver logs en tiempo real
podman logs -f observability-bridge

# Entrar al contenedor (debug)
podman exec -it observability-bridge /bin/sh

# Reconstruir sin caché
podman build --no-cache -t observability-bridge:latest -f Containerfile .

# Ejecutar tests (cuando existan)
go test ./...

# Verificar conexión a DB desde app
podman exec -it observability-bridge /bin/sh -c "wget -qO- http://localhost:8080/api/v1/health"
```

## Troubleshooting

### Error: "connection refused" a PostgreSQL
- Verificar que PostgreSQL está corriendo: `podman ps`
- Verificar URL de conexión en `.env`
- Para conexión entre contenedores, usar `host.containers.internal` en lugar de `localhost`

### Error: "port already in use"
```bash
# Matar proceso usando el puerto
sudo lsof -ti:8080 | xargs kill -9
# o
sudo fuser -k 8080/tcp
```

### Error: "permission denied" en Podman
```bash
# Verificar permisos
podman unshare chown 0:0 -R .

# O ejecutar con sudo (no recomendado para dev)
sudo podman build ...
```

## Flujo de Trabajo Recomendado

1. **Desarrollo**: `go run cmd/api/main.go` con DB local
2. **Test integración**: Podman con PostgreSQL local
3. **Pre-producción**: Podman con Neon PostgreSQL
4. **Producción**: Cloud Run con Neon PostgreSQL
