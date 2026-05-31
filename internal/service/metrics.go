package service

import (
	"context"
	"fmt"
	"time"

	"github.com/AndersonFlores2006/go-observability-bridge/internal/model"
	"github.com/AndersonFlores2006/go-observability-bridge/internal/repository"
	"github.com/AndersonFlores2006/go-observability-bridge/pkg/logger"
)

// MetricsService maneja la lógica de métricas
type MetricsService struct {
	repo repository.Repository
	log  *logger.Logger
}

// NewMetricsService crea nuevo servicio de métricas
func NewMetricsService(repo repository.Repository, log *logger.Logger) *MetricsService {
	return &MetricsService{repo: repo, log: log}
}

// IngestMetric procesa una métrica recibida
func (s *MetricsService) IngestMetric(ctx context.Context, metric *model.Metric) error {
	if metric.ID == "" {
		metric.ID = generateID()
	}
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now().UTC()
	}

	s.log.Info("Ingestando métrica: %s = %.2f", metric.Name, metric.Value)
	return s.repo.SaveMetric(ctx, metric)
}

// QueryMetrics consulta métricas almacenadas
func (s *MetricsService) QueryMetrics(ctx context.Context, req *model.QueryRequest) (*model.QueryResponse, error) {
	s.log.Info("Consultando métrica: %s", req.MetricName)
	return s.repo.QueryMetrics(ctx, req)
}

// IngestLog procesa un log recibido
func (s *MetricsService) IngestLog(ctx context.Context, log *model.LogEntry) error {
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now().UTC()
	}
	log.ID = generateID()
	return s.repo.SaveLog(ctx, log)
}

// ListLogs recupera logs recientes
func (s *MetricsService) ListLogs(ctx context.Context, limit int) ([]model.LogEntry, error) {
	return s.repo.ListLogs(ctx, limit)
}

func generateID() string {
	return fmt.Sprintf("metric_%d", time.Now().UnixNano())
}
