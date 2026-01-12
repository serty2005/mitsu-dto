package models

import (
	"fmt"
	"time"
)

// ConnectionProfile представляет профиль подключения к ККТ
type ConnectionProfile struct {
	SerialNumber   string    `json:"serial_number"`   // Серийный номер ККТ (уникальный идентификатор)
	ConnectionType int       `json:"connection_type"` // 0 = COM, 6 = TCP
	ComName        string    `json:"com_name"`        // Например "COM9"
	BaudRate       int       `json:"baud_rate"`       // Например 115200
	IPAddress      string    `json:"ip_address"`      // Например "192.168.1.100"
	TCPPort        int       `json:"tcp_port"`        // Например 8200
	FirmwareVer    string    `json:"firmware_ver"`    // Версия прошивки ККТ
	ModelName      string    `json:"model_name"`      // Модель ККТ
	LastUsed       time.Time `json:"last_used"`       // Время последнего успешного подключения
}

// DisplayString возвращает строку для UI
func (p *ConnectionProfile) DisplayString() string {
	var connInfo string
	if p.ConnectionType == 0 {
		connInfo = p.ComName
	} else {
		connInfo = fmt.Sprintf("%s:%d", p.IPAddress, p.TCPPort)
	}

	fwVer := p.FirmwareVer
	if fwVer == "" {
		fwVer = "—"
	} else {
		fwVer = "v" + fwVer
	}

	return fmt.Sprintf("SN%s - %s - %s", p.SerialNumber, connInfo, fwVer)
}
