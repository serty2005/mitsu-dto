package models

// SettingsSnapshot представляет собой полный слепок настроек вкладки "Сервис".
type SettingsSnapshot struct {
	// Сеть и связь
	Ofd      OfdSettings
	Oism     OismSettings
	Lan      LanSettings
	Timezone int

	// Оборудование
	Printer PrinterSettings
	Drawer  DrawerSettings

	// Опции (b0-b9)
	// Храним как плоскую структуру для удобства сравнения
	Options DeviceOptions

	// Клише
	// Ключ map - номер типа клише (1..4). Значение - массив строк.
	Cliches map[int][]ClicheLineData
}

// NewSettingsSnapshot создает пустой снапшот с инициализированной картой клише.
func NewSettingsSnapshot() *SettingsSnapshot {
	return &SettingsSnapshot{
		Cliches: make(map[int][]ClicheLineData),
	}
}

// IsZero проверяет, пустой ли снапшот (используется для проверки инициализации).
func (s *SettingsSnapshot) IsZero() bool {
	return s == nil || len(s.Cliches) == 0
}

// Helper: deepEqual для срезов клише
func clichesEqual(a, b []ClicheLineData) bool {
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
