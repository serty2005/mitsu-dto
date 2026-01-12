package gui

import (
	"time"

	"mitsuscanner/driver"
	"mitsuscanner/internal/models"
	"mitsuscanner/internal/service/monitor"
)

// KktPanelStatus содержит данные состояния для верхней панели
// Используется для совместимости с существующим кодом GUI
type KktPanelStatus struct {
	ModelName    string    // Модель ККТ
	SerialNumber string    // Серийный номер
	UnsentDocs   int       // Количество неотправленных документов в ОФД
	PowerFlag    bool      // Текущее значение флага питания (TRUE=OK, FALSE=Reboot)
	LastUpdate   time.Time // Время последнего обновления
}

// Глобальная переменная сервиса мониторинга
var monitorService *monitor.Service

// StartMonitor запускает мониторинг через сервис
func StartMonitor(drv driver.Driver, model, serial string, unsentDocs int) {
	if monitorService != nil {
		monitorService.Stop()
	}

	// Создаем начальное состояние
	initialState := models.KktStatus{
		ModelName:    model,
		SerialNumber: serial,
		UnsentDocs:   unsentDocs,
		PowerFlag:    true, // Начальное состояние - OK
		LastUpdate:   time.Now(),
	}

	// Создаем конфиг
	config := monitor.Config{
		PollInterval: 3 * time.Second,
		InitialState: initialState,
	}

	// Инициализируем сервис
	monitorService = monitor.NewService(drv, config)

	// Запускаем мониторинг
	monitorService.Start()
}

// StopMonitor останавливает мониторинг
func StopMonitor() {
	if monitorService != nil {
		monitorService.Stop()
	}
}

// PauseMonitor приостанавливает мониторинг
func PauseMonitor() {
	if monitorService != nil {
		monitorService.Pause()
	}
}

// ResumeMonitor возобновляет мониторинг
func ResumeMonitor() {
	if monitorService != nil {
		monitorService.Resume()
	}
}

// SetUpdateCallback устанавливает callback для обновления UI
func SetUpdateCallback(fn func(*models.KktStatus)) {
	if monitorService != nil {
		monitorService.SetUpdateCallback(fn)
	}
}
