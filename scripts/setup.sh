#!/bin/bash
set -e

echo "================================"
echo "Setup: Go Observability Bridge"
echo "================================"

# Verificar Go
if ! command -v go &> /dev/null; then
    echo "❌ Go no está instalado. Instalar desde https://go.dev/dl/"
    exit 1
fi

echo "✅ Go version: $(go version)"

# Verificar Podman
if ! command -v podman &> /dev/null; then
    echo "⚠️ Podman no encontrado. Puedes usar Docker o instalar Podman."
else
    echo "✅ Podman version: $(podman --version)"
fi

# Inicializar proyecto
echo -e "\n Inicializando módulo Go..."
go mod tidy

# Crear .env si no existe
if [ ! -f .env ]; then
    echo " Creando .env desde ejemplo..."
    cp .env.example .env
    echo " Edita .env con tu DATABASE_URL antes de ejecutar"
fi

echo -e "\n✅ Setup completado!"
echo ""
echo "Próximos pasos:"
echo "1. Edita .env con tu configuración de PostgreSQL"
echo "2. Ejecuta: make dev"
echo "3. Verifica: curl http://localhost:8080/api/v1/health"
