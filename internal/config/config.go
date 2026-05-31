package config

import (
	"os"
)

// Config contiene toda la configuración de la aplicación
type Config struct {
	Port        string
	DatabaseURL string
	LogLevel    string
	WebhookURL  string
}

// Load carga la configuración desde variables de entorno
func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://user:pass@localhost:5432/observability?sslmode=disable"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		WebhookURL:  getEnv("WEBHOOK_URL", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
