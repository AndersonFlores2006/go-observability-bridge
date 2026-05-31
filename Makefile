# Ejecutar localmente
dev:
	@echo "Iniciando servidor de desarrollo..."
	export $$(grep -v '^#' .env | grep -v '^$$' | xargs) && go run cmd/api/main.go

# Construir binario
build:
	go build -o bin/observability-bridge cmd/api/main.go

# Ejecutar tests
test:
	go test -v ./...

# Construir imagen Podman
podman-build:
	podman build -t observability-bridge:latest -f Containerfile .

# Ejecutar con Podman
podman-run:
	podman run -d \
	  --name observability-bridge \
	  -p 8080:8080 \
	  --env-file .env \
	  observability-bridge:latest

# Iniciar stack completo con compose
compose-up:
	cd deploy && podman-compose up -d

# Detener stack
compose-down:
	cd deploy && podman-compose down

# Ver logs
logs:
	podman logs -f observability-bridge

# Formatear código
fmt:
	go fmt ./...

# Descargar dependencias
deps:
	go mod tidy

# Verificar compilación
check:
	go build ./...

.PHONY: dev build test podman-build podman-run compose-up compose-down logs fmt deps check
