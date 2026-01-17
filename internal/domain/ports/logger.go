package ports

// Logger определяет интерфейс для абстракции логирования.
// Позволяет использовать любой реализации логирования (например, стандартный log, zap, slog).
type Logger interface {
	// Debug выводит отладочную информацию
	Debug(msg string, args ...interface{})

	// Info выводит информационные сообщения
	Info(msg string, args ...interface{})

	// Warn выводит предупреждения
	Warn(msg string, args ...interface{})

	// Error выводит ошибки
	Error(msg string, args ...interface{})

	// Fatal выводит критические ошибки и завершает программу
	Fatal(msg string, args ...interface{})

	// Printf форматированный вывод (для совместимости)
	Printf(format string, args ...interface{})
}
