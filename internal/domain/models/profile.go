package models

import "time"

// ConnectionProfile представляет профиль подключения к ККТ
type ConnectionProfile struct {
	SerialNumber   string    // Серийный номер ККТ (уникальный идентификатор)
	ConnectionType int       // 0 = COM, 6 = TCP
	ComName        string    // Например "COM9"
	BaudRate       int       // Например 115200
	IPAddress      string    // Например "192.168.1.100"
	TCPPort        int       // Например 8200
	FirmwareVer    string    // Версия прошивки ККТ
	ModelName      string    // Модель ККТ
	LastUsed       time.Time // Время последнего успешного подключения
}
