package gui

import (
	"context"
	"log"
	"sync"
	"time"

	"mitsuscanner/driver"
)

// KktPanelStatus содержит данные состояния для верхней панели
type KktPanelStatus struct {
	ModelName    string    // Модель ККТ
	SerialNumber string    // Серийный номер
	UnsentDocs   int       // Количество неотправленных документов в ОФД
	PowerFlag    bool      // Текущее значение флага питания (TRUE=OK, FALSE=Reboot)
	LastUpdate   time.Time // Время последнего обновления
}

// Глобальные переменные модуля
var (
	monitorCtx     context.Context
	monitorCancel  context.CancelFunc
	monitorMutex   sync.Mutex
	panelStatus    = &KktPanelStatus{}
	updateCallback func(*KktPanelStatus)
	isPaused       bool
)

// StartMonitor запускает мониторинг.
func StartMonitor(drv driver.Driver, model, serial string, unsentDocs int) {
	monitorMutex.Lock()
	defer monitorMutex.Unlock()

	// Если уже запущен - остановим
	if monitorCancel != nil {
		monitorCancel()
	}

	// Инициализация контекста
	monitorCtx, monitorCancel = context.WithCancel(context.Background())

	// Установка начального состояния
	panelStatus.ModelName = model
	panelStatus.SerialNumber = serial
	panelStatus.UnsentDocs = unsentDocs
	// Важно: Мы предполагаем, что при старте мониторинга мы уже установили флаг в 1 (TRUE).
	// Поэтому начальное состояние = true (OK).
	panelStatus.PowerFlag = true
	panelStatus.LastUpdate = time.Now()

	// Сразу обновляем UI начальными данными
	if updateCallback != nil {
		statusCopy := *panelStatus
		go updateCallback(&statusCopy)
	}

	// Запуск горутины мониторинга
	go monitorRoutine(drv)
	log.Printf("[MONITOR] Мониторинг ККТ запущен (Тихий режим)")
}

// StopMonitor останавливает мониторинг
func StopMonitor() {
	monitorMutex.Lock()
	defer monitorMutex.Unlock()

	if monitorCancel != nil {
		monitorCancel()
		monitorCancel = nil
		log.Printf("[MONITOR] Мониторинг ККТ остановлен")
	}
}

// PauseMonitor приостанавливает мониторинг (для рег-запросов)
func PauseMonitor() {
	monitorMutex.Lock()
	defer monitorMutex.Unlock()
	isPaused = true
}

// ResumeMonitor возобновляет мониторинг
func ResumeMonitor() {
	monitorMutex.Lock()
	defer monitorMutex.Unlock()
	isPaused = false
}

// SetUpdateCallback устанавливает callback для обновления UI
func SetUpdateCallback(fn func(*KktPanelStatus)) {
	monitorMutex.Lock()
	defer monitorMutex.Unlock()
	updateCallback = fn
}

// monitorRoutine - основная горутина мониторинга
func monitorRoutine(drv driver.Driver) {
	// Небольшая пауза перед началом опроса
	time.Sleep(2 * time.Second)

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-monitorCtx.Done():
			return

		case <-ticker.C:
			monitorMutex.Lock()
			if isPaused {
				monitorMutex.Unlock()
				continue
			}
			monitorMutex.Unlock()

			// Опрашиваем ТОЛЬКО флаг питания
			checkPowerFlag(drv)
		}
	}
}

// checkPowerFlag проверяет только флаг питания
func checkPowerFlag(drv driver.Driver) {
	// Получаем текущий флаг питания из ККТ
	// Если ККТ перезагрузилась, флаг сбросится в 0 (false).
	// Если все ок, он останется 1 (true), который мы установили при подключении.
	powerFlag, err := drv.GetPowerFlag()
	if err != nil {
		// Ошибки связи не логируем в основной лог, чтобы не спамить
		// Можно добавить счетчик ошибок, если нужно
		return
	}

	monitorMutex.Lock()
	changed := powerFlag != panelStatus.PowerFlag
	panelStatus.PowerFlag = powerFlag
	panelStatus.LastUpdate = time.Now()
	monitorMutex.Unlock()

	// Если состояние изменилось, обновляем UI
	if changed && updateCallback != nil {
		monitorMutex.Lock()
		statusCopy := *panelStatus
		monitorMutex.Unlock()

		// Логируем только событие смены статуса
		if !statusCopy.PowerFlag {
			log.Printf("[MONITOR] ВНИМАНИЕ: Обнаружена перезагрузка ККТ! (Флаг сброшен)")
		} else {
			log.Printf("[MONITOR] Питание восстановлено/подтверждено.")
		}

		go updateCallback(&statusCopy)
	}
}
