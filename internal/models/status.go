package models

import "time"

// KktStatus содержит данные состояния для верхней панели
// Используется для отображения текущего состояния ККТ в интерфейсе пользователя
type KktStatus struct {
	ModelName    string    // Модель ККТ
	SerialNumber string    // Серийный номер
	UnsentDocs   int       // Количество неотправленных документов в ОФД
	PowerFlag    bool      // Текущее значение флага питания (TRUE=OK, FALSE=Reboot)
	LastUpdate   time.Time // Время последнего обновления
}
