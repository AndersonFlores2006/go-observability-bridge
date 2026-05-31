package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AndersonFlores2006/go-observability-bridge/internal/config"
	"github.com/AndersonFlores2006/go-observability-bridge/internal/handler"
	"github.com/AndersonFlores2006/go-observability-bridge/internal/repository"
	"github.com/AndersonFlores2006/go-observability-bridge/internal/service"
	"github.com/AndersonFlores2006/go-observability-bridge/pkg/logger"
)

func main() {
	// Cargar configuración
	cfg := config.Load()

	// Inicializar logger
	log := logger.New(cfg.LogLevel)
	log.Info("Iniciando Observability Bridge...")

	// Inicializar repositorio
	repo, err := repository.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Error conectando a base de datos: %v", err)
	}
	defer repo.Close()

	// Inicializar servicios
	metricsService := service.NewMetricsService(repo, log)

	// Inicializar handlers
	router := handler.NewRouter(metricsService, log)

	// Configurar servidor HTTP
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Canal para graceful shutdown
	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Goroutine para graceful shutdown
	go func() {
		<-quit
		log.Info("Servidor deteniéndose...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Error en shutdown: %v", err)
		}
		close(done)
	}()

	log.Info("Servidor corriendo en http://localhost:%s", cfg.Port)
	log.Info("Endpoints disponibles:")
	log.Info("  POST /api/v1/metrics       - Enviar métricas")
	log.Info("  POST /api/v1/logs          - Enviar logs")
	log.Info("  GET  /api/v1/metrics/query - Consultar métricas")
	log.Info("  GET  /api/v1/health        - Health check")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Error iniciando servidor: %v", err)
	}

	<-done
	log.Info("Servidor detenido correctamente")
}
