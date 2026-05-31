package model

import "time"

// Metric representa una métrica recibida
type Metric struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels"`
	Timestamp time.Time         `json:"timestamp"`
}

// LogEntry representa un log recibido
type LogEntry struct {
	ID        string            `json:"id"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Service   string            `json:"service"`
	Metadata  map[string]string `json:"metadata"`
	Timestamp time.Time         `json:"timestamp"`
}

// QueryRequest para consultar métricas
type QueryRequest struct {
	MetricName  string            `json:"metric_name"`
	From        time.Time         `json:"from"`
	To          time.Time         `json:"to"`
	Labels      map[string]string `json:"labels"`
	Aggregation string            `json:"aggregation"` // avg, sum, count, max, min
}

// QueryResponse resultado de consulta
type QueryResponse struct {
	MetricName  string      `json:"metric_name"`
	Values      []DataPoint `json:"values"`
	Aggregation string      `json:"aggregation"`
}

// DataPoint punto de datos para series temporales
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}
