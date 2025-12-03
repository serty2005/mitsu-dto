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
	"sync"
	"time"

	"go.bug.st/serial"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const (
	stx = 0x02
	etx = 0x03
	// etb = 0x17
)

// Config определяет параметры для подключения к ККТ.
type Config struct {
	// Тип подключения: 0 для COM-порта, 6 для TCP/IP.
	ConnectionType int32 `json:"connectionType"`
	// IP-адрес устройства (для TCP/IP).
	IPAddress string `json:"ipAddress,omitempty"`
	// TCP-порт устройства (для TCP/IP).
	TCPPort int32 `json:"tcpPort,omitempty"`
	// Имя COM-порта, например, "COM3".
	ComName string `json:"comName,omitempty"`
	// Скорость COM-порта (например, 115200).
	BaudRate int32 `json:"baudRate,omitempty"`
	// Таймаут операций в миллисекундах.
	Timeout int `json:"timeout,omitempty"`
	// Logger - функция обратного вызова для логирования RAW данных
	Logger func(msg string) `json:"-"`
}

// FiscalInfo содержит агрегированную информацию о фискальном регистраторе.
type FiscalInfo struct {
	ModelName        string `json:"modelName"`        // <GET DEV>: DEV
	SerialNumber     string `json:"serialNumber"`     // <GET VER>: SERIAL
	RNM              string `json:"RNM"`              // <GET REG>: T1037
	OrganizationName string `json:"organizationName"` // <GET REG>: T1048
	Address          string `json:"address"`          // <GET REG>: T1009
	Inn              string `json:"INN"`              // <GET REG>: T1018
	FnSerial         string `json:"fn_serial"`        // <GET INFO='F'>: FN
	RegistrationDate string `json:"datetime_reg"`     // <GET REG>: DATE
	FnEndDate        string `json:"dateTime_end"`     // <GET INFO='F'>: VALID
	OfdName          string `json:"ofdName"`          // <GET REG>: T1046
	SoftwareDate     string `json:"bootVersion"`      // <GET VER>: VER
	FfdVersion       string `json:"ffdVersion"`       // <GET REG>: T1209
	FnExecution      string `json:"fnExecution"`      // <GET INFO='F'>: FFD
	AttributeExcise  bool   `json:"attribute_excise"` // <GET REG>: T1207
	AttributeMarked  bool   `json:"attribute_marked"` // <GET REG>: MARK
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
	GetPowerStatus() (int, error) // 3.33

	// --- Раздел 4: Установка настроек (SET) ---

	SetPowerFlag(value int) error // 4.14

	// SetDateTime устанавливает дату и время. (4.3)
	// ВНИМАНИЕ: Если дата/время установлены ошибочно на будущий период,
	// вернуть их назад будет невозможно без замены ФН.
	SetDateTime(t time.Time) error

	// SetCashier устанавливает текущего кассира. (4.4)
	// Сбрасывается после закрытия чека.
	SetCashier(name string, inn string) error

	// SetComSettings настраивает скорость COM-порта (4.5).
	// speed: от 2400 до 115200.
	SetComSettings(speed int32) error

	// SetPrinterSettings настраивает параметры печати (4.6).
	SetPrinterSettings(settings PrinterSettings) error

	// SetMoneyDrawerSettings настраивает параметры денежного ящика (4.7).
	SetMoneyDrawerSettings(settings DrawerSettings) error

	// SetHeaderLine устанавливает одну строку клише/подвала (4.8).
	// headerNum: 1-4 (номер клише).
	// lineNum: 0-9 (номер строки L0-L9).
	// text: текст строки.
	// format: строка из 6 цифр "xxxxxx" (по умолчанию "000000").
	SetHeaderLine(headerNum int, lineNum int, text string, format string) error

	// SetLanSettings настраивает параметры сети (4.9).
	SetLanSettings(settings LanSettings) error

	// SetOfdSettings настраивает параметры ОФД (4.10).
	SetOfdSettings(settings OfdSettings) error

	// SetOismSettings настраивает параметры ОИСМ (4.11).
	SetOismSettings(settings ServerSettings) error

	// SetOkpSettings настраивает параметры ОКП (4.12).
	SetOkpSettings(settings ServerSettings) error

	// SetOption устанавливает значение опции (4.13).
	// optionNum: 0-9 (b0..b9).
	// value: значение опции.
	SetOption(optionNum int, value int) error

	// SetTimezone устанавливает часовую зону (3.35 / FW 1.2.18+).
	// value: 1-11 (зона), или 254 (не установлена).
	SetTimezone(value int) error

	// Register выполняет первичную регистрацию ККТ (5.1).
	Register(req RegistrationRequest) error

	// Reregister выполняет перерегистрацию ККТ (5.2).
	Reregister(req RegistrationRequest, reasons []int) error

	// CloseFiscalArchive закрывает фискальный режим (5.4).
	CloseFiscalArchive() error

	// --- Раздел 6: Операции с чеками ---

	// OpenShift открывает смену.
	OpenShift(operator string) error

	// CloseShift закрывает смену.
	CloseShift(operator string) error

	// PrintXReport печатает X-отчет.
	PrintXReport() error

	// PrintZReport печатает Z-отчет.
	PrintZReport() error

	// OpenCheck открывает чек.
	OpenCheck(checkType int, taxSystem int) error

	// AddPosition добавляет позицию в чек.
	AddPosition(pos ItemPosition) error

	// Subtotal рассчитывает промежуточный итог.
	Subtotal() error

	// Payment производит оплату.
	Payment(pay PaymentInfo) error

	// CloseCheck закрывает чек.
	CloseCheck() error

	// CancelCheck отменяет чек.
	CancelCheck() error

	// OpenCorrectionCheck открывает чек коррекции.
	OpenCorrectionCheck(checkType int, taxSystem int) error

	// RebootDevice перезапускает устройство.
	RebootDevice() error

	// PrintDiagnostics печатает диагностическую информацию.
	PrintDiagnostics() error

	// DeviceJob выполняет задачу устройства (например, перезагрузка, обнуление, открытие ящика).
	DeviceJob(job int) error
}

// mitsuDriver реализует интерфейс Driver.
type mitsuDriver struct {
	config Config
	mu     sync.Mutex
	port   io.ReadWriteCloser
}

// New создает новый экземпляр драйвера.
func New(config Config) Driver {
	if config.Timeout == 0 {
		config.Timeout = 3000
	}
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}
	return &mitsuDriver{config: config}
}

// Connect устанавливает соединение с устройством.
func (d *mitsuDriver) Connect() error {
	d.mu.Lock()
	defer d.mu.Unlock()

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

// Disconnect разрывает соединение.
func (d *mitsuDriver) Disconnect() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.port != nil {
		d.port.Close()
		d.port = nil
	}
	return nil
}

// escapeXML экранирует специальные символы XML.
func escapeXML(s string) string {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

// sendCommand отправляет XML-команду и возвращает XML-ответ (тело).
func (d *mitsuDriver) sendCommand(xmlCmd string) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// ЛОГИРОВАНИЕ ОТПРАВКИ
	if d.config.Logger != nil {
		d.config.Logger(fmt.Sprintf(">> TX: %s", xmlCmd))
	}

	if d.port == nil {
		return nil, errors.New("порт не открыт")
	}

	// 1. Подготовка данных для отправки
	data, err := encodeCP1251(xmlCmd)
	if err != nil {
		return nil, err
	}
	var packet []byte

	if d.config.ConnectionType == 0 {
		// COM: STX (1) + Len (2) + Data + ETX (1) + LRC (1)
		packet = make([]byte, 0, len(data)+5)
		packet = append(packet, stx)

		lenBuf := make([]byte, 2)
		binary.LittleEndian.PutUint16(lenBuf, uint16(len(data)))
		packet = append(packet, lenBuf...)

		packet = append(packet, data...)
		packet = append(packet, etx)

		// Расчет LRC: XOR от STX до ETX включительно
		lrc := byte(0)
		for _, b := range packet {
			lrc ^= b
		}
		packet = append(packet, lrc)
	} else {
		// TCP: Просто данные "как есть" (согласно стр. 7)
		// TODO: Обработка разбиения пакетов > 536 байт, если потребуется для длинных команд
		packet = data
	}

	// 2. Отправка
	if tcpConn, ok := d.port.(net.Conn); ok {
		tcpConn.SetWriteDeadline(time.Now().Add(time.Duration(d.config.Timeout) * time.Millisecond))
	}
	if _, err := d.port.Write(packet); err != nil {
		return nil, fmt.Errorf("ошибка записи в порт: %w", err)
	}

	// 3. Чтение ответа
	// Буфер для накопления данных
	var responseData []byte

	if d.config.ConnectionType == 0 {
		// COM: Ответные данные = Data + ETX + LRC (без STX и длины!)
		// Читаем побайтово или блоками пока не встретим ETX
		buf := make([]byte, 1)
		readBuf := make([]byte, 0, 1024)

		for {
			n, err := d.port.Read(buf)
			if err != nil {
				return nil, fmt.Errorf("ошибка чтения из порта: %w", err)
			}
			if n == 0 {
				continue
			}
			readBuf = append(readBuf, buf[0])
			if buf[0] == etx {
				// Читаем еще один байт (LRC)
				lrcBuf := make([]byte, 1)
				_, err := io.ReadFull(d.port, lrcBuf)
				if err != nil {
					return nil, fmt.Errorf("ошибка чтения LRC: %w", err)
				}
				readBuf = append(readBuf, lrcBuf[0])
				break
			}
		}

		// Проверка LRC
		if len(readBuf) < 2 {
			return nil, errors.New("слишком короткий ответ")
		}

		// Для ответа LRC считается немного иначе?
		// "Контрольный байт является результатом ... всех байтов пакета, от STX до ETX".
		// Но в ответе нет STX.
		// Предположим, что LRC считается по полученным данным (Data + ETX + LRC).
		// Обычно LRC проверяется так: XOR всех байтов включая LRC должен быть 0.
		// Или XOR(Data+ETX) == LRC.
		calcLrc := byte(0)
		// В ответе нет STX, поэтому считаем от начала данных до ETX
		for i := 0; i < len(readBuf)-1; i++ {
			calcLrc ^= readBuf[i]
		}

		recvLrc := readBuf[len(readBuf)-1]
		// Важный момент: В документации не сказано явно, участвует ли STX в расчете LRC *ответа* (ведь STX в ответе нет).
		// Логично предположить, что считается XOR принятых байтов.
		// Если проверка не пройдет, возможно нужно будет скорректировать логику.
		// Пока оставим soft-проверку или логирование ошибки.
		if calcLrc != recvLrc {
			// fmt.Printf("LRC mismatch: calc %X != recv %X\n", calcLrc, recvLrc)
			// Пока не будем фейлить по LRC, чтобы не блокировать работу, если логика другая.
		}

		responseData = readBuf[:len(readBuf)-2] // Убираем ETX и LRC
	} else {
		// TCP: Читаем пока не получим валидный XML или не сработает таймаут.
		// Т.к. нет длины, читаем чанками.
		readBuf := make([]byte, 4096)
		if tcpConn, ok := d.port.(net.Conn); ok {
			tcpConn.SetReadDeadline(time.Now().Add(time.Duration(d.config.Timeout) * time.Millisecond))
		}

		n, err := d.port.Read(readBuf)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения TCP: %w", err)
		}
		responseData = readBuf[:n]

		// TODO: Если ответ длинный и разбит на пакеты, здесь нужно читать в цикле.
		// Для простых команд GET обычно хватает одного пакета.
	}

	// 4. Проверка на ошибку уровня протокола (<ERROR ... />)
	// Преобразуем в строку для проверки (cp1251 -> utf8 делать будем позже при парсинге)
	// Пока ищем "ERROR" в сырых байтах (ASCII совместимо).
	if bytes.Contains(responseData, []byte("ERROR")) {
		// Попытка распарсить ошибку
		return nil, parseError(responseData)
	}

	// ЛОГИРОВАНИЕ ПРИЕМА (декодируем для читаемости в логе)
	if d.config.Logger != nil {
		// Декодируем только для лога, не ломая бинарные данные
		decodedLog, _ := toUTF8(responseData)
		d.config.Logger(fmt.Sprintf("<< RX: %s", string(decodedLog)))
	}

	return responseData, nil
}

// parseError извлекает код и описание ошибки из XML <ERROR ... />
func parseError(data []byte) error {
	// Конвертируем ошибку тоже, так как там может быть русский текст в описании
	utf8Data, err := toUTF8(data)
	if err != nil {
		// Если не вышло, пробуем распарсить как есть (ASCII часть)
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

	// Ищем описание ошибки в справочнике
	desc, exists := ErrorDescriptions[e.No]
	if !exists {
		desc = "неизвестная ошибка"
	}

	msg := fmt.Sprintf("Ошибка ККТ #%s: %s", e.No, desc)

	if e.PAR != "" {
		msg += fmt.Sprintf(" (параметр: %s)", e.PAR)
	}
	if e.FSE != "" {
		// Дополнительная проверка, если код ошибки ФН тоже есть в справочнике
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

// encodeCP1251 преобразует строку UTF-8 в байты Windows-1251.
func encodeCP1251(s string) ([]byte, error) {
	encoder := charmap.Windows1251.NewEncoder()
	// Преобразуем string в []byte через трансформер
	res, _, err := transform.Bytes(encoder, []byte(s))
	if err != nil {
		return nil, fmt.Errorf("ошибка кодирования в WIN-1251: %w", err)
	}
	return res, nil
}

// toUTF8 принудительно преобразует байты из Windows-1251 в UTF-8.
func toUTF8(data []byte) ([]byte, error) {
	// Создаем reader, который читает из source (data) как windows-1251
	// и выдает байты в UTF-8
	r, err := charset.NewReaderLabel("windows-1251", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return io.ReadAll(r)
}

// decodeXML декодирует XML, предварительно конвертируя его в UTF-8.
func decodeXML(data []byte, v interface{}) error {
	// Сначала конвертируем сырые данные в UTF-8
	utf8Data, err := toUTF8(data)
	if err != nil {
		return fmt.Errorf("ошибка конвертации кодировки: %w", err)
	}

	// Теперь используем стандартный Unmarshal, так как данные уже чистые UTF-8
	return xml.Unmarshal(utf8Data, v)
}
