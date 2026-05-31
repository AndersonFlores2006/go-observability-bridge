#!/bin/bash

echo "================================"
echo "Pruebas: Observability Bridge"
echo "================================"

BASE_URL="http://localhost:8080"

echo -e "\n1. Health Check"
curl -s $BASE_URL/api/v1/health | jq . || echo "⚠️ Instala jq para mejor formato"

echo -e "\n2. Enviar métrica (CPU)"
curl -s -X POST $BASE_URL/api/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cpu_usage",
    "value": 78.5,
    "labels": {"host": "server-01", "region": "us-east"}
  }' | jq .

echo -e "\n3. Enviar métrica (Memoria)"
curl -s -X POST $BASE_URL/api/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "name": "memory_usage",
    "value": 512.5,
    "labels": {"host": "server-01", "unit": "MB"}
  }' | jq .

echo -e "\n4. Enviar log (ERROR)"
curl -s -X POST $BASE_URL/api/v1/logs \
  -H "Content-Type: application/json" \
  -d '{
    "level": "ERROR",
    "message": "Database connection timeout after 5000ms",
    "service": "api-gateway",
    "metadata": {"request_id": "abc-123", "endpoint": "/api/v1/users"}
  }' | jq .

echo -e "\n5. Consultar métrica"
curl -s "$BASE_URL/api/v1/metrics/query?metric=cpu_usage" | jq .

echo -e "\n✅ Pruebas completadas"
