package ofdclient

import (
	"fmt"
	"time"
)

// Константы сигнатур — байты в порядке нумерации согласно спецификации
// Сигнатура ОФД: '2A08410A'h → байты [0x2A, 0x08, 0x41, 0x0A]
var SignatureOFDBytes = [4]byte{0x2A, 0x08, 0x41, 0x0A}

// Сигнатура ОИСМ: 'DD80CAA1'h → байты [0xDD, 0x80, 0xCA, 0xA1]
var SignatureOISMBytes = [4]byte{0xDD, 0x80, 0xCA, 0xA1}

// Константы версий S-протокола — байты в порядке нумерации
// '81A2'h → байты [0x81, 0xA2]
var SProtoVersionOFDBytes = [2]byte{0x81, 0xA2}

// '82FB'h → байты [0x82, 0xFB] — тестовые посылки
var SProtoVersionOFDTestBytes = [2]byte{0x82, 0xFB}

// '82A2'h → байты [0x82, 0xA2]
var SProtoVersionOISMBytes = [2]byte{0x82, 0xA2}

// Версии P-протокола (ФФД) — байты в порядке нумерации
// ФФД 1.0:  '0100'h → [0x01, 0x00]
// ФФД 1.05: '0105'h → [0x01, 0x05]
// ФФД 1.1:  '0110'h → [0x01, 0x10]
// ФФД 1.2:  '0120'h → [0x01, 0x20]
var (
	PProtoFFD10Bytes  = [2]byte{0x01, 0x00}
	PProtoFFD105Bytes = [2]byte{0x01, 0x05}
	PProtoFFD11Bytes  = [2]byte{0x01, 0x10}
	PProtoFFD12Bytes  = [2]byte{0x01, 0x20}
)

// Флаги сообщения (Little Endian при сериализации)
type MessageFlags uint16

const (
	FlagCRCNone        MessageFlags = 0b00     // Биты 0-1: без CRC
	FlagCRCHeader      MessageFlags = 0b01     // CRC только заголовка
	FlagCRCFull        MessageFlags = 0b10     // CRC заголовка + тела
	FlagHasContainer   MessageFlags = 0b0100   // Бит 2: есть контейнер
	FlagExpectResponse MessageFlags = 0b010000 // Биты 4-5: ожидаем ответ
)

// MessageHeader — заголовок сообщения протокола (30 байт)
// Порядок байт:
//   - Signature, SProtoVersion, PProtoVersion, FnNumber — порядок нумерации (как массивы байт)
//   - BodySize, Flags, CRC — Little Endian
type MessageHeader struct {
	Signature     [4]byte      // 4 байта, порядок нумерации
	SProtoVersion [2]byte      // 2 байта, порядок нумерации
	PProtoVersion [2]byte      // 2 байта, порядок нумерации
	FnNumber      [16]byte     // 16 байт ASCII
	BodySize      uint16       // 2 байта, Little Endian
	Flags         MessageFlags // 2 байта, Little Endian
	CRC           uint16       // 2 байта, Little Endian
}

// FFDVersionToBytes конвертирует строковую версию ФФД в байты протокола
func FFDVersionToBytes(version string) ([2]byte, error) {
	switch version {
	case "1.0", "1.00", "1":
		return PProtoFFD10Bytes, nil
	case "1.05":
		return PProtoFFD105Bytes, nil
	case "1.1", "1.10":
		return PProtoFFD11Bytes, nil
	case "1.2", "1.20":
		return PProtoFFD12Bytes, nil
	default:
		return [2]byte{}, fmt.Errorf("%w: %s", ErrInvalidFFDVersion, version)
	}
}

// FnFFDCodeToVersion конвертирует код ФФД из статуса ФН в строку версии
// Коды из ФН: "2" = 1.05, "3" = 1.1, "4" = 1.2
func FnFFDCodeToVersion(ffdCode string) string {
	switch ffdCode {
	case "2":
		return "1.05"
	case "3":
		return "1.1"
	case "4":
		return "1.2"
	default:
		return "1.0"
	}
}

// ContainerHeader — заголовок контейнера (минимум 7 байт)
// Все поля в Little Endian
type ContainerHeader struct {
	Length        uint16 // Little Endian
	CRC           uint16 // Little Endian
	ContainerType byte
	DataType      byte
	FormatVersion byte
}

// SendRequest — запрос на отправку документа
type SendRequest struct {
	OfdAddress string // host:port
	FnNumber   string // 16-значный номер ФН
	FFDVersion string // "1.0", "1.05", "1.1", "1.2"
	Container  []byte // Контейнер с документом (получен от ККТ)
}

// SendResponse — ответ от ОФД
type SendResponse struct {
	Receipt    []byte // Квитанция для записи в ФН
	RawMessage []byte // Полное сырое сообщение (для отладки)
}

// Config — конфигурация клиента
type Config struct {
	Timeout       time.Duration // Таймаут ожидания ответа (по умолчанию 300с)
	RetryCount    int           // Количество попыток переподключения
	RetryInterval time.Duration // Интервал между попытками
	Logger        func(string)  // Опциональный логгер
}
