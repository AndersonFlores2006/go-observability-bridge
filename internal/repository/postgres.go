package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AndersonFlores2006/go-observability-bridge/internal/model"
	_ "github.com/lib/pq"
)

type Repository interface {
	SaveMetric(ctx context.Context, metric *model.Metric) error
	SaveLog(ctx context.Context, log *model.LogEntry) error
	QueryMetrics(ctx context.Context, req *model.QueryRequest) (*model.QueryResponse, error)
	ListLogs(ctx context.Context, limit int) ([]model.LogEntry, error)
	Close() error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgres(connectionURL string) (Repository, error) {
	db, err := sql.Open("postgres", connectionURL)
	if err != nil {
		return nil, fmt.Errorf("error abriendo conexión: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error conectando a base de datos: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	repo := &PostgresRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("error ejecutando migración: %w", err)
	}

	return repo, nil
}

func (r *PostgresRepository) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS metrics (
		id BIGSERIAL PRIMARY KEY,
		metric_name VARCHAR(255) NOT NULL,
		value DOUBLE PRECISION NOT NULL,
		labels JSONB DEFAULT '{}',
		timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_metrics_name_timestamp ON metrics(metric_name, timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_metrics_labels ON metrics USING GIN(labels);

	CREATE TABLE IF NOT EXISTS logs (
		id BIGSERIAL PRIMARY KEY,
		level VARCHAR(50) NOT NULL,
		message TEXT NOT NULL,
		service VARCHAR(255) NOT NULL,
		metadata JSONB DEFAULT '{}',
		timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_logs_service_timestamp ON logs(service, timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
	CREATE INDEX IF NOT EXISTS idx_logs_metadata ON logs USING GIN(metadata);
	`

	_, err := r.db.Exec(schema)
	return err
}

func (r *PostgresRepository) SaveMetric(ctx context.Context, metric *model.Metric) error {
	labelsJSON, err := json.Marshal(metric.Labels)
	if err != nil {
		return fmt.Errorf("error serializando labels: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO metrics (metric_name, value, labels, timestamp)
		 VALUES ($1, $2, $3, $4)`,
		metric.Name, metric.Value, labelsJSON, metric.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("error guardando métrica: %w", err)
	}
	return nil
}

func (r *PostgresRepository) SaveLog(ctx context.Context, log *model.LogEntry) error {
	metadataJSON, err := json.Marshal(log.Metadata)
	if err != nil {
		return fmt.Errorf("error serializando metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO logs (level, message, service, metadata, timestamp)
		 VALUES ($1, $2, $3, $4, $5)`,
		log.Level, log.Message, log.Service, metadataJSON, log.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("error guardando log: %w", err)
	}
	return nil
}

func (r *PostgresRepository) QueryMetrics(ctx context.Context, req *model.QueryRequest) (*model.QueryResponse, error) {
	var (
		query string
		args  []interface{}
	)

	agg := strings.ToLower(req.Aggregation)
	paramIdx := 1

	args = append(args, req.MetricName)

	where := fmt.Sprintf("metric_name = $%d", paramIdx)
	paramIdx++

	if !req.From.IsZero() {
		where += fmt.Sprintf(" AND timestamp >= $%d", paramIdx)
		args = append(args, req.From)
		paramIdx++
	}
	if !req.To.IsZero() {
		where += fmt.Sprintf(" AND timestamp <= $%d", paramIdx)
		args = append(args, req.To)
		paramIdx++
	}
	if len(req.Labels) > 0 {
		for k, v := range req.Labels {
			cond := fmt.Sprintf(" AND labels @> $%d", paramIdx)
			where += cond
			filter := map[string]string{k: v}
			filterJSON, _ := json.Marshal(filter)
			args = append(args, string(filterJSON))
			paramIdx++
		}
	}

	switch agg {
	case "avg", "sum", "max", "min":
		query = fmt.Sprintf(`
			SELECT date_trunc('hour', timestamp) AS bucket,
			       %s(value) AS agg_value
			FROM metrics
			WHERE %s
			GROUP BY bucket
			ORDER BY bucket ASC`, agg, where)
	case "count":
		query = fmt.Sprintf(`
			SELECT date_trunc('hour', timestamp) AS bucket,
			       COUNT(*) AS agg_value
			FROM metrics
			WHERE %s
			GROUP BY bucket
			ORDER BY bucket ASC`, where)
	default:
		query = fmt.Sprintf(`
			SELECT timestamp, value
			FROM metrics
			WHERE %s
			ORDER BY timestamp ASC`, where)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error consultando métricas: %w", err)
	}
	defer rows.Close()

	var points []model.DataPoint
	for rows.Next() {
		var dp model.DataPoint
		if agg != "" {
			var bucket time.Time
			var val float64
			if err := rows.Scan(&bucket, &val); err != nil {
				return nil, fmt.Errorf("error escaneando fila: %w", err)
			}
			dp = model.DataPoint{Timestamp: bucket, Value: val}
		} else {
			if err := rows.Scan(&dp.Timestamp, &dp.Value); err != nil {
				return nil, fmt.Errorf("error escaneando fila: %w", err)
			}
		}
		points = append(points, dp)
	}

	return &model.QueryResponse{
		MetricName:  req.MetricName,
		Values:      points,
		Aggregation: agg,
	}, nil
}

func (r *PostgresRepository) ListLogs(ctx context.Context, limit int) ([]model.LogEntry, error) {
	if limit < 1 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT level, message, service, metadata, timestamp
		 FROM logs ORDER BY timestamp DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("error listando logs: %w", err)
	}
	defer rows.Close()

	var logs []model.LogEntry
	for rows.Next() {
		var l model.LogEntry
		var metaJSON []byte
		if err := rows.Scan(&l.Level, &l.Message, &l.Service, &metaJSON, &l.Timestamp); err != nil {
			return nil, fmt.Errorf("error escaneando log: %w", err)
		}
		json.Unmarshal(metaJSON, &l.Metadata)
		if l.Metadata == nil {
			l.Metadata = make(map[string]string)
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}
