# API Documentation: Observability Bridge

## Base URL

```
Desarrollo: http://localhost:8080
Producción: https://observability-bridge-xxxx-uc.a.run.app
```

## Autenticación

Actualmente sin autenticación (para MVP). Futuro: API Keys en header `X-API-Key`.

## Endpoints

### 1. Health Check

```http
GET /api/v1/health
```

**Response 200 OK:**
```json
{
  "status": "ok",
  "timestamp": "2025-01-15T10:30:00Z",
  "service": "go-observability-bridge",
  "version": "1.0.0"
}
```

---

### 2. Ingesta de Métricas

```http
POST /api/v1/metrics
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "cpu_usage",
  "value": 78.5,
  "labels": {
    "host": "server-01",
    "region": "us-east",
    "environment": "production"
  },
  "timestamp": "2025-01-15T10:30:00Z"
}
```

**Campos:**
- `name` (required): Nombre de la métrica
- `value` (required): Valor numérico
- `labels` (optional): Tags para filtrado
- `timestamp` (optional): ISO 8601, default: now()

**Response 201 Created:**
```json
{
  "status": "success",
  "id": "metric_1705315800000000000"
}
```

**Response 400 Bad Request:**
```json
{
  "error": "Campo 'name' requerido",
  "status": 400,
  "success": false
}
```

---

### 3. Ingesta de Logs

```http
POST /api/v1/logs
Content-Type: application/json
```

**Request Body:**
```json
{
  "level": "ERROR",
  "message": "Database connection timeout",
  "service": "api-gateway",
  "metadata": {
    "request_id": "abc-123",
    "duration_ms": "5000",
    "endpoint": "/api/v1/users"
  },
  "timestamp": "2025-01-15T10:30:00Z"
}
```

**Campos:**
- `level` (required): DEBUG, INFO, WARN, ERROR, FATAL
- `message` (required): Mensaje del log
- `service` (required): Nombre del servicio origen
- `metadata` (optional): Campos adicionales
- `timestamp` (optional): ISO 8601, default: now()

**Response 201 Created:**
```json
{
  "status": "success",
  "id": "log_1705315800000000000"
}
```

---

### 4. Consulta de Métricas

```http
GET /api/v1/metrics/query?metric=cpu_usage&from=2025-01-01T00:00:00Z&to=2025-01-15T23:59:59Z&aggregation=avg
```

**Query Parameters:**
- `metric` (required): Nombre de la métrica
- `from` (optional): Inicio rango temporal (ISO 8601)
- `to` (optional): Fin rango temporal (ISO 8601)
- `aggregation` (optional): avg, sum, count, max, min
- `label.*` (optional): Filtrar por labels (ej: `label.host=server-01`)

**Response 200 OK:**
```json
{
  "metric_name": "cpu_usage",
  "aggregation": "avg",
  "values": [
    {
      "timestamp": "2025-01-15T10:00:00Z",
      "value": 65.3
    },
    {
      "timestamp": "2025-01-15T11:00:00Z",
      "value": 78.5
    }
  ]
}
```

---

## Códigos de Error

| Código | Significado |
|--------|-------------|
| 400 | Bad Request - JSON inválido o parámetros faltantes |
| 404 | Not Found - Métrica no encontrada |
| 405 | Method Not Allowed - Usar método HTTP correcto |
| 500 | Internal Server Error - Error del servidor |

## Rate Limiting

Futuro: 1000 requests/min por API key.

## Ejemplos con curl

### Script de prueba completo

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"

# Health check
echo "=== Health Check ==="
curl -s $BASE_URL/api/v1/health | jq

# Enviar métrica
echo -e "\n=== Enviar Métrica ==="
curl -s -X POST $BASE_URL/api/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "name": "memory_usage",
    "value": 512.5,
    "labels": {"host": "server-01", "unit": "MB"}
  }' | jq

# Enviar log
echo -e "\n=== Enviar Log ==="
curl -s -X POST $BASE_URL/api/v1/logs \
  -H "Content-Type: application/json" \
  -d '{
    "level": "INFO",
    "message": "User login successful",
    "service": "auth-service",
    "metadata": {"user_id": "12345"}
  }' | jq

# Consultar métricas
echo -e "\n=== Consultar Métricas ==="
curl -s "$BASE_URL/api/v1/metrics/query?metric=memory_usage" | jq
```
