FROM golang:1.24-alpine AS builder

WORKDIR /app

# Instalar dependencias
RUN apk add --no-cache git

# Copiar go.mod y go.sum primero para cache
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar binario
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o observability-bridge ./cmd/api

# Imagen final mínima
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiar binario desde builder
COPY --from=builder /app/observability-bridge .

# Puerto expuesto
EXPOSE 8080

# Comando por defecto
CMD ["./observability-bridge"]
