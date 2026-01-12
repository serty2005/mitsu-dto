package monitor

import (
	"context"
	"log"
	"sync"
	"time"

	"mitsuscanner/driver"
	"mitsuscanner/internal/models"
)

// Service реализует сервис мониторинга состояния ККТ
type Service struct {
	driver         driver.Driver
	config         Config
	status         models.KktStatus
	ctx            context.Context
	cancel         context.CancelFunc
	mutex          sync.Mutex
	isPaused       bool
	updateCallback func(*models.KktStatus)
}

// Config содержит конфигурацию опроса
type Config struct {
	PollInterval time.Duration    // Интервал опроса
	InitialState models.KktStatus // Начальное состояние
}

// NewService создает новый экземпляр сервиса мониторинга
func NewService(drv driver.Driver, cfg Config) *Service {
	return &Service{
		driver: drv,
		config: cfg,
		status: cfg.InitialState,
	}
}

// Start запускает мониторинг ККТ
func (s *Service) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Если уже запущен - остановим
	if s.cancel != nil {
		s.cancel()
	}

	// Инициализация контекста
	s.ctx, s.cancel = context.WithCancel(context.Background())

	// Установка начального состояния
	s.status = s.config.InitialState
	s.status.LastUpdate = time.Now()

	// Сразу вызываем callback с начальными данными
	if s.updateCallback != nil {
		statusCopy := s.getStatusCopy()
		go s.updateCallback(&statusCopy)
	}

	// Запуск горутины мониторинга
	go s.monitorRoutine()
	log.Printf("[MONITOR] Мониторинг ККТ запущен")
}

// Stop останавливает мониторинг
func (s *Service) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
		log.Printf("[MONITOR] Мониторинг ККТ остановлен")
	}
}

// Pause приостанавливает мониторинг
func (s *Service) Pause() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.isPaused = true
}

// Resume возобновляет мониторинг
func (s *Service) Resume() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.isPaused = false
}

// SetUpdateCallback устанавливает callback для обновления UI
func (s *Service) SetUpdateCallback(fn func(*models.KktStatus)) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.updateCallback = fn
}

// GetCurrentStatus возвращает текущее состояние (потокобезопасно)
func (s *Service) GetCurrentStatus() models.KktStatus {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.getStatusCopy()
}

// monitorRoutine - основная горутина мониторинга
func (s *Service) monitorRoutine() {
	// Небольшая пауза перед началом опроса
	time.Sleep(2 * time.Second)

	ticker := time.NewTicker(s.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return

		case <-ticker.C:
			s.mutex.Lock()
			if s.isPaused {
				s.mutex.Unlock()
				continue
			}
			s.mutex.Unlock()

			// Опрашиваем ТОЛЬКО флаг питания
			s.checkPowerFlag()
		}
	}
}

// checkPowerFlag проверяет только флаг питания
func (s *Service) checkPowerFlag() {
	// Получаем текущий флаг питания из ККТ
	// Если ККТ перезагрузилась, флаг сбросится в 0 (false).
	// Если все ок, он останется 1 (true), который мы установили при подключении.
	powerFlag, err := s.driver.GetPowerFlag()
	if err != nil {
		// Ошибки связи не логируем в основной лог, чтобы не спамить
		// Можно добавить счетчик ошибок, если нужно
		return
	}

	s.mutex.Lock()
	changed := powerFlag != s.status.PowerFlag
	s.status.PowerFlag = powerFlag
	s.status.LastUpdate = time.Now()
	s.mutex.Unlock()

	// Если состояние изменилось, обновляем UI
	if changed && s.updateCallback != nil {
		s.mutex.Lock()
		statusCopy := s.getStatusCopy()
		s.mutex.Unlock()

		// Логируем только событие смены статуса
		if !statusCopy.PowerFlag {
			log.Printf("[MONITOR] ВНИМАНИЕ: Обнаружена перезагрузка ККТ! (Флаг сброшен)")
		} else {
			log.Printf("[MONITOR] Питание восстановлено/подтверждено.")
		}

		go s.updateCallback(&statusCopy)
	}
}

// getStatusCopy возвращает копию текущего состояния (потокобезопасно)
func (s *Service) getStatusCopy() models.KktStatus {
	return models.KktStatus{
		ModelName:    s.status.ModelName,
		SerialNumber: s.status.SerialNumber,
		UnsentDocs:   s.status.UnsentDocs,
		PowerFlag:    s.status.PowerFlag,
		LastUpdate:   s.status.LastUpdate,
	}
}
