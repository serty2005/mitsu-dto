package service

import (
	"mitsuscanner/driver"
)

// Priority определяет порядок применения настроек.
type Priority int

const (
	PriorityNormal   Priority = 0 // Обычные настройки (Опции, Принтер)
	PriorityCliche   Priority = 1 // Клише (много данных, лучше отдельно)
	PriorityNetwork  Priority = 2 // Сеть (может разорвать соединение, строго в конце)
	PriorityCritical Priority = 3 // Критические операции (если появятся)
)

// SettingsSnapshot представляет собой полный слепок настроек вкладки "Сервис".
type SettingsSnapshot struct {
	// Сеть и связь
	Ofd      driver.OfdSettings
	Oism     driver.ServerSettings
	Lan      driver.LanSettings
	Timezone int

	// Оборудование
	Printer driver.PrinterSettings
	Drawer  driver.DrawerSettings

	// Опции (b0-b9)
	// Храним как плоскую структуру для удобства сравнения
	Options driver.DeviceOptions

	// Клише
	// Ключ map - номер типа клише (1..4). Значение - массив строк.
	Cliches map[int][]driver.ClicheLineData
}

// Change представляет одно атомарное (или групповое) изменение настроек.
type Change struct {
	ID          string      // Уникальный ID поля (для подсветки в GUI)
	Description string      // Человекочитаемое описание изменения
	OldValue    interface{} // Значение "Было" (для отображения)
	NewValue    interface{} // Значение "Стало" (для отображения)
	Priority    Priority    // Приоритет выполнения

	// ApplyFunc - замыкание, которое применит это изменение к драйверу.
	// Будет сформировано на этапе сравнения.
	ApplyFunc func(d driver.Driver) error
}

// NewSettingsSnapshot создает пустой снапшот с инициализированной картой клише.
func NewSettingsSnapshot() *SettingsSnapshot {
	return &SettingsSnapshot{
		Cliches: make(map[int][]driver.ClicheLineData),
	}
}

// IsZero проверяет, пустой ли снапшот (используется для проверки инициализации).
func (s *SettingsSnapshot) IsZero() bool {
	return s == nil || len(s.Cliches) == 0
}

// Helper: deepEqual для срезов клише
func clichesEqual(a, b []driver.ClicheLineData) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
