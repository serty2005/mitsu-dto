package logger

import (
	"log"
	"os"

	"mitsuscanner/internal/domain/ports"
)

// StdLogger реализует интерфейс ports.Logger с использованием стандартной библиотеки log.
type StdLogger struct {
	logger *log.Logger
}

// NewStdLogger создает новый экземпляр StdLogger с заданным префиксом.
func NewStdLogger(prefix string) ports.Logger {
	return &StdLogger{
		logger: log.New(os.Stderr, prefix, log.LstdFlags),
	}
}

// Debug выводит отладочную информацию.
func (l *StdLogger) Debug(msg string, args ...interface{}) {
	l.logger.Printf("[DEBUG] "+msg, args...)
}

// Info выводит информационные сообщения.
func (l *StdLogger) Info(msg string, args ...interface{}) {
	l.logger.Printf("[INFO] "+msg, args...)
}

// Warn выводит предупреждения.
func (l *StdLogger) Warn(msg string, args ...interface{}) {
	l.logger.Printf("[WARN] "+msg, args...)
}

// Error выводит ошибки.
func (l *StdLogger) Error(msg string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+msg, args...)
}

// Fatal выводит критические ошибки и завершает программу.
func (l *StdLogger) Fatal(msg string, args ...interface{}) {
	l.logger.Fatalf("[FATAL] "+msg, args...)
}

// Printf форматированный вывод (для совместимости).
func (l *StdLogger) Printf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}
