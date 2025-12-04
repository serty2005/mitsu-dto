// Package mitsu предоставляет интерфейс для взаимодействия с фискальными
// регистраторами Mitsu 1-F через прямой протокол обмена (XML over COM/TCP).
package mitsu

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const (
	stx           = 0x02
	etx           = 0x03
	etb           = 0x17 // End of Transmission Block (для TCP пакетов)
	maxTcpPayload = 535  // Максимальный размер данных в одном TCP пакете
)

// Config определяет параметры для подключения к ККТ.
type Config struct {
	ConnectionType int32            `json:"connectionType"`      // 0 - COM, 6 - TCP
	IPAddress      string           `json:"ipAddress,omitempty"` // TCP IP
	TCPPort        int32            `json:"tcpPort,omitempty"`   // TCP Port
	ComName        string           `json:"comName,omitempty"`   // COM Port Name
	BaudRate       int32            `json:"baudRate,omitempty"`  // COM Speed
	Timeout        int              `json:"timeout,omitempty"`   // Timeout ms
	Logger         func(msg string) `json:"-"`
}

// FiscalInfo содержит агрегированную информацию о фискальном регистраторе.
type FiscalInfo struct {
	ModelName        string `json:"modelName"`
	SerialNumber     string `json:"serialNumber"`
	RNM              string `json:"RNM"`
	OrganizationName string `json:"organizationName"`
	Address          string `json:"address"`
	Inn              string `json:"INN"`
	FnSerial         string `json:"fn_serial"`
	RegistrationDate string `json:"datetime_reg"`
	FnEndDate        string `json:"dateTime_end"`
	OfdName          string `json:"ofdName"`
	SoftwareDate     string `json:"bootVersion"`
	FfdVersion       string `json:"ffdVersion"`
	FnExecution      string `json:"fnExecution"`
	AttributeExcise  bool   `json:"attribute_excise"`
	AttributeMarked  bool   `json:"attribute_marked"`
}

// Driver определяет основной интерфейс для работы с ККТ.
type Driver interface {
	Connect() error
	Disconnect() error
	GetFiscalInfo() (*FiscalInfo, error)
	GetModel() (string, error)
	GetVersion() (string, string, string, error)
	GetDateTime() (time.Time, error)
	GetCashier() (string, string, error)
	GetPrinterSettings() (*PrinterSettings, error)
	GetMoneyDrawerSettings() (*DrawerSettings, error)
	GetComSettings() (int32, error)
	GetHeader(int) ([]string, error)
	GetLanSettings() (*LanSettings, error)
	GetOfdSettings() (*OfdSettings, error)
	GetOismSettings() (*ServerSettings, error)
	GetOkpSettings() (*ServerSettings, error)
	GetTaxRates() (*TaxRates, error)
	GetRegistrationData() (*RegData, error)
	GetShiftStatus() (*ShiftStatus, error)
	GetShiftTotals() (*ShiftTotals, error)
	GetFnStatus() (*FnStatus, error)
	GetOfdExchangeStatus() (*OfdExchangeStatus, error)
	GetMarkingStatus() (*MarkingStatus, error)
	GetTimezone() (int, error)
	GetPowerStatus() (int, error)

	SetPowerFlag(value int) error
	SetDateTime(t time.Time) error
	SetCashier(name string, inn string) error
	SetComSettings(speed int32) error
	SetPrinterSettings(settings PrinterSettings) error
	SetMoneyDrawerSettings(settings DrawerSettings) error
	SetHeaderLine(headerNum int, lineNum int, text string, format string) error
	SetLanSettings(settings LanSettings) error
	SetOfdSettings(settings OfdSettings) error
	SetOismSettings(settings ServerSettings) error
	SetOkpSettings(settings ServerSettings) error
	SetOption(optionNum int, value int) error
	SetTimezone(value int) error

	Register(req RegistrationRequest) error
	Reregister(req RegistrationRequest, reasons []int) error
	CloseFiscalArchive() error

	OpenShift(operator string) error
	CloseShift(operator string) error
	PrintXReport() error
	PrintZReport() error
	OpenCheck(checkType int, taxSystem int) error
	AddPosition(pos ItemPosition) error
	Subtotal() error
	Payment(pay PaymentInfo) error
	CloseCheck() error
	CancelCheck() error
	OpenCorrectionCheck(checkType int, taxSystem int) error
	RebootDevice() error
	PrintDiagnostics() error
	DeviceJob(job int) error
}

type mitsuDriver struct {
	config Config
	mu     sync.Mutex
	port   io.ReadWriteCloser
}

func New(config Config) Driver {
	if config.Timeout == 0 {
		config.Timeout = 3000
	}
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}
	return &mitsuDriver{config: config}
}

func (d *mitsuDriver) Connect() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.connectLocked()
}

// connectLocked - внутренняя функция подключения (без мьютекса)
func (d *mitsuDriver) connectLocked() error {
	if d.port != nil {
		return nil
	}

	var err error
	switch d.config.ConnectionType {
	case 0: // COM
		mode := &serial.Mode{
			BaudRate: int(d.config.BaudRate),
			DataBits: 8,
			Parity:   serial.NoParity,
			StopBits: serial.OneStopBit,
		}
		d.port, err = serial.Open(d.config.ComName, mode)
		if err != nil {
			return fmt.Errorf("ошибка открытия COM-порта: %w", err)
		}
		if p, ok := d.port.(serial.Port); ok {
			p.SetReadTimeout(time.Duration(d.config.Timeout) * time.Millisecond)
		}

	case 6: // TCP/IP
		addr := net.JoinHostPort(d.config.IPAddress, strconv.Itoa(int(d.config.TCPPort)))
		// Для TCP используем таймаут на подключение
		conn, err := net.DialTimeout("tcp", addr, time.Duration(d.config.Timeout)*time.Millisecond)
		if err != nil {
			return fmt.Errorf("ошибка подключения TCP: %w", err)
		}
		d.port = conn

	default:
		return fmt.Errorf("неизвестный тип подключения: %d", d.config.ConnectionType)
	}

	return nil
}

func (d *mitsuDriver) Disconnect() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.disconnectLocked()
}

func (d *mitsuDriver) disconnectLocked() error {
	if d.port != nil {
		d.port.Close()
		d.port = nil
	}
	return nil
}

func escapeXML(s string) string {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

// sendCommand отправляет команду с механизмом повтора (Reconnect) для TCP
func (d *mitsuDriver) sendCommand(xmlCmd string) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Максимальное количество попыток для TCP. Для COM - 1.
	attempts := 1
	if d.config.ConnectionType == 6 {
		attempts = 2
	}

	var lastErr error

	for i := 0; i < attempts; i++ {
		// Если порт закрыт, пробуем открыть (актуально для реконнекта)
		if d.port == nil {
			if err := d.connectLocked(); err != nil {
				lastErr = err
				continue
			}
		}

		resp, err := d.performExchange(xmlCmd)
		if err == nil {
			return resp, nil
		}

		// Если ошибка, и это TCP, и мы еще можем повторить...
		if d.config.ConnectionType == 6 && i < attempts-1 {
			// Логируем факт реконнекта
			if d.config.Logger != nil {
				d.config.Logger(fmt.Sprintf("TCP Error (%v). Reconnecting...", err))
			}
			// Принудительно закрываем порт, чтобы connectLocked открыл новый
			d.disconnectLocked()
			lastErr = err
			continue
		}

		return nil, err
	}

	return nil, lastErr
}

// performExchange выполняет один цикл отправки-приема (без логики реконнекта)
func (d *mitsuDriver) performExchange(xmlCmd string) ([]byte, error) {
	// ЛОГИРОВАНИЕ ОТПРАВКИ
	if d.config.Logger != nil {
		d.config.Logger(fmt.Sprintf(">> TX: %s", xmlCmd))
	}

	// 1. Подготовка данных
	data, err := encodeCP1251(xmlCmd)
	if err != nil {
		return nil, err
	}

	// 2. Отправка
	if d.config.ConnectionType == 0 {
		// --- COM ---
		packet := make([]byte, 0, len(data)+5)
		packet = append(packet, stx)
		lenBuf := make([]byte, 2)
		binary.LittleEndian.PutUint16(lenBuf, uint16(len(data)))
		packet = append(packet, lenBuf...)
		packet = append(packet, data...)
		packet = append(packet, etx)
		lrc := byte(0)
		for _, b := range packet {
			lrc ^= b
		}
		packet = append(packet, lrc)

		if _, err := d.port.Write(packet); err != nil {
			return nil, err
		}
	} else {
		// --- TCP ---
		if tcpConn, ok := d.port.(net.Conn); ok {
			tcpConn.SetWriteDeadline(time.Now().Add(time.Duration(d.config.Timeout) * time.Millisecond))
		}

		offset := 0
		totalLen := len(data)
		for offset < totalLen {
			remaining := totalLen - offset
			chunkSize := remaining
			if chunkSize > maxTcpPayload {
				chunkSize = maxTcpPayload
			}
			chunk := data[offset : offset+chunkSize]
			if _, err := d.port.Write(chunk); err != nil {
				return nil, err
			}
			offset += chunkSize
			if offset < totalLen {
				if _, err := d.port.Write([]byte{etb}); err != nil {
					return nil, err
				}
			}
		}
	}

	// 3. Чтение
	var responseData []byte

	if d.config.ConnectionType == 0 {
		// --- COM ---
		buf := make([]byte, 1)
		readBuf := make([]byte, 0, 1024)
		for {
			n, err := d.port.Read(buf)
			if err != nil {
				return nil, err
			}
			if n == 0 {
				continue
			}
			readBuf = append(readBuf, buf[0])
			if buf[0] == etx {
				lrcBuf := make([]byte, 1)
				_, err := io.ReadFull(d.port, lrcBuf)
				if err != nil {
					return nil, err
				}
				readBuf = append(readBuf, lrcBuf[0])
				break
			}
		}
		if len(readBuf) < 2 {
			return nil, errors.New("short response")
		}
		responseData = readBuf[:len(readBuf)-2]
	} else {
		// --- TCP ---
		if tcpConn, ok := d.port.(net.Conn); ok {
			tcpConn.SetReadDeadline(time.Now().Add(time.Duration(d.config.Timeout) * time.Millisecond))
		}

		accumulated := make([]byte, 0, 4096)
		tempBuf := make([]byte, 1024)

		for {
			n, err := d.port.Read(tempBuf)
			if err != nil {
				// Если EOF, но данные уже есть - это нормальное завершение для устройств,
				// закрывающих сокет после ответа.
				if err == io.EOF && len(accumulated) > 0 {
					break
				}
				// Если Timeout, но данные есть - тоже считаем успехом
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() && len(accumulated) > 0 {
					break
				}
				return nil, err
			}
			if n == 0 {
				continue
			}

			chunk := tempBuf[:n]
			hasEtb := false
			if len(chunk) > 0 && chunk[len(chunk)-1] == etb {
				hasEtb = true
				chunk = chunk[:len(chunk)-1]
			}
			accumulated = append(accumulated, chunk...)

			// Если был ETB, значит ждем еще. Если нет - проверяем конец XML.
			if !hasEtb {
				// Простая проверка конца пакета
				// Проверяем последние байты на наличие закрывающих тегов
				// Для надежности декодируем хвост в строку (ASCII символы >, / совпадают)
				tailLen := 20
				if len(accumulated) < 20 {
					tailLen = len(accumulated)
				}
				tail := string(accumulated[len(accumulated)-tailLen:])

				if strings.Contains(tail, "/>") ||
					strings.Contains(tail, "</OK>") ||
					strings.Contains(tail, "</ERROR>") ||
					strings.Contains(tail, "</ANS>") ||
					strings.Contains(tail, "</Do>") {
					break
				}
			}
		}
		responseData = accumulated
	}

	// 4. Проверка на ошибку уровня протокола
	if bytes.Contains(responseData, []byte("ERROR")) {
		return nil, parseError(responseData)
	}

	if d.config.Logger != nil {
		decodedLog, _ := toUTF8(responseData)
		d.config.Logger(fmt.Sprintf("<< RX: %s", string(decodedLog)))
	}

	return responseData, nil
}

func parseError(data []byte) error {
	utf8Data, err := toUTF8(data)
	if err != nil {
		utf8Data = data
	}

	type ErrorResp struct {
		No  string `xml:"No,attr"`
		FSE string `xml:"FSE,attr"`
		TAG string `xml:"TAG,attr"`
		PAR string `xml:"PAR,attr"`
	}
	var e ErrorResp

	if err := xml.Unmarshal(utf8Data, &e); err != nil {
		return fmt.Errorf("ошибка ККТ (нераспознанная): %s", string(data))
	}

	desc, exists := ErrorDescriptions[e.No]
	if !exists {
		desc = "неизвестная ошибка"
	}

	msg := fmt.Sprintf("Ошибка ККТ #%s: %s", e.No, desc)

	if e.PAR != "" {
		msg += fmt.Sprintf(" (параметр: %s)", e.PAR)
	}
	if e.FSE != "" {
		fnDesc, fnExists := ErrorDescriptions[e.FSE]
		if fnExists {
			msg += fmt.Sprintf(", ошибка ФН #%s: %s", e.FSE, fnDesc)
		} else {
			msg += fmt.Sprintf(", ошибка ФН: %s", e.FSE)
		}
	}
	if e.TAG != "" {
		msg += fmt.Sprintf(" [TAG: %s]", e.TAG)
	}

	return errors.New(msg)
}

func encodeCP1251(s string) ([]byte, error) {
	encoder := charmap.Windows1251.NewEncoder()
	res, _, err := transform.Bytes(encoder, []byte(s))
	if err != nil {
		return nil, fmt.Errorf("ошибка кодирования в WIN-1251: %w", err)
	}
	return res, nil
}

func toUTF8(data []byte) ([]byte, error) {
	r, err := charset.NewReaderLabel("windows-1251", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return io.ReadAll(r)
}

func decodeXML(data []byte, v interface{}) error {
	utf8Data, err := toUTF8(data)
	if err != nil {
		return fmt.Errorf("ошибка конвертации кодировки: %w", err)
	}
	return xml.Unmarshal(utf8Data, v)
}
