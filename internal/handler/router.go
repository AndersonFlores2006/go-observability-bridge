package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/AndersonFlores2006/go-observability-bridge/internal/model"
	"github.com/AndersonFlores2006/go-observability-bridge/internal/service"
	"github.com/AndersonFlores2006/go-observability-bridge/pkg/logger"
)

// Router configura todas las rutas HTTP
func NewRouter(metricsService *service.MetricsService, log *logger.Logger) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/api/v1/health", handleHealth)

	// Métricas
	mux.HandleFunc("/api/v1/metrics", func(w http.ResponseWriter, r *http.Request) {
		handleMetrics(w, r, metricsService, log)
	})

	// Logs
	mux.HandleFunc("/api/v1/logs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleListLogs(w, r, metricsService, log)
		} else {
			handleLogs(w, r, metricsService, log)
		}
	})

	// Query métricas
	mux.HandleFunc("/api/v1/metrics/query", func(w http.ResponseWriter, r *http.Request) {
		handleQuery(w, r, metricsService, log)
	})

	// Dashboard estático
	mux.Handle("/", http.FileServer(http.Dir("static")))

	return withCORS(withLogging(mux, log))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "go-observability-bridge",
		"version":   "1.0.0",
	})
}

func handleMetrics(w http.ResponseWriter, r *http.Request, service *service.MetricsService, log *logger.Logger) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	var metric model.Metric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if err := service.IngestMetric(r.Context(), &metric); err != nil {
		writeError(w, http.StatusInternalServerError, "Error procesando métrica")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"status": "success",
		"id":     metric.ID,
	})
}

func handleLogs(w http.ResponseWriter, r *http.Request, service *service.MetricsService, log *logger.Logger) {
	var logEntry model.LogEntry
	if err := json.NewDecoder(r.Body).Decode(&logEntry); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if err := service.IngestLog(r.Context(), &logEntry); err != nil {
		writeError(w, http.StatusInternalServerError, "Error guardando log")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"status": "success",
		"id":     logEntry.ID,
	})
}

func handleListLogs(w http.ResponseWriter, r *http.Request, service *service.MetricsService, log *logger.Logger) {
	logs, err := service.ListLogs(r.Context(), 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error listando logs")
		return
	}
	if logs == nil {
		logs = []model.LogEntry{}
	}
	writeJSON(w, http.StatusOK, logs)
}

func handleQuery(w http.ResponseWriter, r *http.Request, service *service.MetricsService, log *logger.Logger) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	// Parsear query params
	q := r.URL.Query()
	req := &model.QueryRequest{
		MetricName:  q.Get("metric"),
		Aggregation: q.Get("aggregation"),
	}

	if from := q.Get("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			req.From = t
		}
	}
	if to := q.Get("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			req.To = t
		}
	}
	for k, v := range q {
		if len(k) > 6 && k[:6] == "label." && len(v) > 0 {
			if req.Labels == nil {
				req.Labels = make(map[string]string)
			}
			req.Labels[k[6:]] = v[0]
		}
	}

	if req.MetricName == "" {
		writeError(w, http.StatusBadRequest, "Parámetro 'metric' requerido")
		return
	}

	resp, err := service.QueryMetrics(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error consultando métricas")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error":   message,
		"status":  status,
		"success": false,
	})
}

func withLogging(next http.Handler, log *logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Info("%s %s - %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
