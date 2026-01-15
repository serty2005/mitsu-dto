package service

import (
	"math"
	"time"
)

const TimeLayout = "02.01.2006 15:04:05"

// TimeStatus описывает состояние синхронизации времени.
type TimeStatus int

const (
	TimeStatusOk       TimeStatus = iota // Разница <= 5 минут
	TimeStatusCritical                   // Разница > 5 минут
	TimeStatusError                      // Ошибка парсинга или нет данных
)

// TimeLogic содержит методы для работы со временем ККТ.
type TimeLogic struct{}

// NewTimeLogic создает экземпляр сервиса.
func NewTimeLogic() *TimeLogic {
	return &TimeLogic{}
}

// ParseTime парсит строку времени из интерфейса.
func (s *TimeLogic) ParseTime(val string) (time.Time, error) {
	return time.Parse(TimeLayout, val)
}

// FormatTime форматирует время для отображения.
func (s *TimeLogic) FormatTime(t time.Time) string {
	return t.Format(TimeLayout)
}

// CompareTimes возвращает разницу между временем ККТ и целевым временем, а также статус.
func (s *TimeLogic) CompareTimes(kktTimeStr string, targetTime time.Time) (time.Duration, TimeStatus) {
	kktTime, err := s.ParseTime(kktTimeStr)
	if err != nil {
		return 0, TimeStatusError
	}

	diff := targetTime.Sub(kktTime)
	// Берем модуль разницы
	absDiff := time.Duration(math.Abs(float64(diff)))

	if absDiff > 5*time.Minute {
		return absDiff, TimeStatusCritical
	}

	return absDiff, TimeStatusOk
}
