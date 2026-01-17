package service

import (
	"mitsuscanner/internal/domain/models"
	"mitsuscanner/pkg/mitsudriver"
)

// Priority aliases для совместимости
const (
	PriorityNormal   = models.PriorityNormal
	PriorityCliche   = models.PriorityCliche
	PriorityNetwork  = models.PriorityNetwork
	PriorityCritical = models.PriorityCritical
)

// SettingsSnapshot представляет собой полный слепок настроек вкладки "Сервис".
type SettingsSnapshot struct {
	// Сеть и связь
	Ofd      mitsudriver.OfdSettings
	Oism     mitsudriver.OismSettings
	Lan      mitsudriver.LanSettings
	Timezone int

	// Оборудование
	Printer mitsudriver.PrinterSettings
	Drawer  mitsudriver.DrawerSettings

	// Опции (b0-b9)
	// Храним как плоскую структуру для удобства сравнения
	Options mitsudriver.DeviceOptions

	// Клише
	// Ключ map - номер типа клише (1..4). Значение - массив строк.
	Cliches map[int][]mitsudriver.ClicheLineData
}

// Change представляет одно атомарное (или групповое) изменение настроек.
type Change struct {
	ID          string                           // Уникальный ID поля (для подсветки в GUI)
	Description string                           // Человекочитаемое описание изменения
	OldValue    interface{}                      // Значение "Было" (для отображения)
	NewValue    interface{}                      // Значение "Стало" (для отображения)
	Priority    models.Priority                  // Приоритет выполнения
	ApplyFunc   func(d mitsudriver.Driver) error // ApplyFunc - замыкание, которое применит это изменение к драйверу.
}

// NewSettingsSnapshot создает пустой снапшот с инициализированной картой клише.
func NewSettingsSnapshot() *SettingsSnapshot {
	return &SettingsSnapshot{
		Cliches: make(map[int][]mitsudriver.ClicheLineData),
	}
}

// IsZero проверяет, пустой ли снапшот (используется для проверки инициализации).
func (s *SettingsSnapshot) IsZero() bool {
	return s == nil || len(s.Cliches) == 0
}

// Helper: deepEqual для срезов клише
func clichesEqual(a, b []mitsudriver.ClicheLineData) bool {
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
