package time

import (
	"math"
	"time"

	"mitsuscanner/internal/domain/ports"
)

const TimeLayout = "02.01.2006 15:04:05"

// TimeStatus описывает состояние синхронизации времени
type TimeStatus int

const (
	TimeStatusOk       TimeStatus = iota // Разница <= 5 минут
	TimeStatusCritical                   // Разница > 5 минут
	TimeStatusError                      // Ошибка парсинга или нет данных
)

// TimeService отвечает за логику работы со временем ККТ
type TimeService struct {
	driver ports.Driver
}

// NewTimeService создает новый экземпляр TimeService
func NewTimeService(driver ports.Driver) *TimeService {
	return &TimeService{
		driver: driver,
	}
}

// GetKKTTime возвращает текущее время ККТ
func (s *TimeService) GetKKTTime() (time.Time, error) {
	return s.driver.GetDateTime()
}

// GetCurrentTime возвращает текущее время системы
func (s *TimeService) GetCurrentTime() time.Time {
	return time.Now()
}

// SyncTime синхронизирует время ККТ с заданным временем
func (s *TimeService) SyncTime(t time.Time) error {
	return s.driver.SetDateTime(t)
}

// SyncWithSystemTime синхронизирует время ККТ с текущим временем системы
func (s *TimeService) SyncWithSystemTime() error {
	return s.SyncTime(time.Now())
}

// ParseTime парсит строку времени из интерфейса
func (s *TimeService) ParseTime(val string) (time.Time, error) {
	return time.Parse(TimeLayout, val)
}

// FormatTime форматирует время для отображения
func (s *TimeService) FormatTime(t time.Time) string {
	return t.Format(TimeLayout)
}

// CompareTimes возвращает разницу между временем ККТ и целевым временем, а также статус
func (s *TimeService) CompareTimes(kktTimeStr string, targetTime time.Time) (time.Duration, TimeStatus) {
	kktTime, err := s.ParseTime(kktTimeStr)
	if err != nil {
		return 0, TimeStatusError
	}

	diff := targetTime.Sub(kktTime)
	absDiff := time.Duration(math.Abs(float64(diff)))

	if absDiff > 5*time.Minute {
		return absDiff, TimeStatusCritical
	}

	return absDiff, TimeStatusOk
}
