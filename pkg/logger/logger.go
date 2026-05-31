package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger interface para logging estructurado
type Logger struct {
	level string
}

// New crea un nuevo logger
func New(level string) *Logger {
	return &Logger{level: level}
}

// Info logea mensajes informativos
func (l *Logger) Info(format string, args ...interface{}) {
	l.log("INFO", format, args...)
}

// Error logea mensajes de error
func (l *Logger) Error(format string, args ...interface{}) {
	l.log("ERROR", format, args...)
}

// Fatal logea error y termina el programa
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log("FATAL", format, args...)
	os.Exit(1)
}

func (l *Logger) log(level, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	log.Printf("[%s] %s: %s", timestamp, level, message)
}
