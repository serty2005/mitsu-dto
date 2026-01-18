package mitsu

import (
	"context"
	"fmt"
	"time"
)

// Client интерфейс определяет базовые методы для взаимодействия с ККТ Mitsu
type Client interface {
	// Connect устанавливает соединение с устройством
	Connect() error

	// Disconnect разрывает соединение с устройством
	Disconnect() error

	// SendCommand отправляет XML-команду устройству и возвращает ответ
	SendCommand(ctx context.Context, xmlCmd string) ([]byte, error)

	// --- Получение информации (раздел 3) ---
	// GetDev запрашивает модель ККТ (См. п. 3.3)
	GetDev(ctx context.Context) (*DevResponse, error)
	// GetVer запрашивает версию ПО, серийный номер и MAC-адрес (См. п. 3.4)
	GetVer(ctx context.Context) (*VerResponse, error)
	// GetDateTime запрашивает текущую дату и время ККТ (См. п. 3.5)
	GetDateTime(ctx context.Context) (*DateTimeResponse, error)
	// GetCashier запрашивает данные кассира (См. п. 3.6)
	GetCashier(ctx context.Context) (*CashierResponse, error)
	// GetPrinterSettings запрашивает настройки принтера (См. п. 3.7)
	GetPrinterSettings(ctx context.Context) (*PrinterSettings, error)
	// GetMoneyDrawerSettings запрашивает настройки денежного ящика (См. п. 3.8)
	GetMoneyDrawerSettings(ctx context.Context) (*DrawerSettings, error)
	// GetComSettings запрашивает настройки COM-порта (См. п. 3.9)
	GetComSettings(ctx context.Context) (*ComSettingsResponse, error)
	// GetHeader запрашивает заголовок (См. п. 3.10)
	GetHeader(ctx context.Context, n int) (*HeaderResponse, error)
	// GetLanSettings запрашивает настройки LAN (См. п. 3.11)
	GetLanSettings(ctx context.Context) (*LanSettings, error)
	// GetOfdSettings запрашивает настройки ОФД (См. п. 3.12)
	GetOfdSettings(ctx context.Context) (*OfdSettings, error)
	// GetOismSettings запрашивает настройки OISM (См. п. 3.13)
	GetOismSettings(ctx context.Context) (*OismSettings, error)
	// GetOkpSettings запрашивает настройки OKP (См. п. 3.14)
	GetOkpSettings(ctx context.Context) (*ServerSettings, error)
	// GetTaxRates запрашивает налоговые ставки (См. п. 3.15)
	GetTaxRates(ctx context.Context) (*TaxRates, error)
	// GetRegistrationData запрашивает данные о регистрации ККТ (См. п. 3.16)
	GetRegistrationData(ctx context.Context) (*RegData, error)
	// GetStatus получает статус ККТ (INFO='0')
	GetStatus(ctx context.Context) (*ShiftStatus, error)
	// GetShiftTotals получает итоги смены (INFO='1')
	GetShiftTotals(ctx context.Context) (*ShiftTotals, error)
	// GetFnStatus получает статус фискального накопителя (INFO='F')
	GetFnStatus(ctx context.Context) (*FnStatus, error)
	// GetOfdExchangeStatus получает статус обмена с ОФД (INFO='O')
	GetOfdExchangeStatus(ctx context.Context) (*OfdExchangeStatus, error)
	// GetMarkingStatus получает статус маркировки (INFO='M')
	GetMarkingStatus(ctx context.Context) (*MarkingStatus, error)
	// GetPowerStatus запрашивает статус питания (См. п. 3.33)
	GetPowerStatus(ctx context.Context) (*PowerStatusResponse, error)
	// GetTimezone запрашивает часовой пояс (См. п. 3.35)
	GetTimezone(ctx context.Context) (*TimezoneResponse, error)
	// GetOptions запрашивает все опции устройства (См. п. 4.13)
	GetOptions(ctx context.Context) (*DeviceOptions, error)
	// GetCurrentDocumentType получает тип текущего документа
	GetCurrentDocumentType(ctx context.Context) (*CurrentDocumentTypeResponse, error)
	// GetDocumentInfoFromFN получает информацию о документе (OFFSET и LENGTH) по номеру FD
	GetDocumentInfoFromFN(ctx context.Context, fd int) (*DocumentInfoResponse, error)
	// ReadBlock читает блок данных из памяти ФН по OFFSET и LENGTH
	ReadBlock(ctx context.Context, offset int64, length int) (*ReadBlockResponse, error)

	// --- Установка настроек (раздел 4) ---
	// SetDateTime устанавливает дату и время ККТ (См. п. 4.3)
	SetDateTime(ctx context.Context, t time.Time) error
	// SetCashier устанавливает данные кассира (См. п. 4.4)
	SetCashier(ctx context.Context, name string, inn string) error
	// SetComSettings устанавливает настройки COM-порта (См. п. 4.5)
	SetComSettings(ctx context.Context, speed int32) error
	// SetPrinterSettings устанавливает настройки принтера (См. п. 4.6)
	SetPrinterSettings(ctx context.Context, s PrinterSettings) error
	// SetMoneyDrawerSettings устанавливает настройки денежного ящика (См. п. 4.7)
	SetMoneyDrawerSettings(ctx context.Context, s DrawerSettings) error
	// SetHeader устанавливает клише и подвала (См. п. 4.8)
	SetHeader(ctx context.Context, headerNum int, lines []ClicheLineData) error
	// SetHeaderLine устанавливает одну строку клише (См. п. 4.8)
	SetHeaderLine(ctx context.Context, headerNum int, lineNum int, text string, format string) error
	// SetLanSettings устанавливает настройки LAN (См. п. 4.9)
	SetLanSettings(ctx context.Context, s LanSettings) error
	// SetOfdSettings устанавливает настройки ОФД (См. п. 4.10)
	SetOfdSettings(ctx context.Context, s OfdSettings) error
	// SetOismSettings устанавливает настройки OISM (См. п. 4.11)
	SetOismSettings(ctx context.Context, s ServerSettings) error
	// SetOkpSettings устанавливает настройки OKP (См. п. 4.12)
	SetOkpSettings(ctx context.Context, s ServerSettings) error
	// SetOption устанавливает одну опцию устройства (См. п. 4.13)
	SetOption(ctx context.Context, optionNum int, value int) error
	// SetPowerFlag устанавливает флаг питания (См. п. 4.14)
	SetPowerFlag(ctx context.Context, value int) error
	// SetTimezone устанавливает часовой пояс (Добавлено в FW 1.2.18)
	SetTimezone(ctx context.Context, value int) error
	// TechReset выполняет технологическое обнуление устройства
	TechReset(ctx context.Context) error

	// --- Регистрация (раздел 5) ---
	// Register выполняет первичную регистрацию ККТ (5.1)
	Register(ctx context.Context, req RegistrationRequest) (*RegResponse, error)
	// Reregister выполняет перерегистрацию ККТ (5.2)
	Reregister(ctx context.Context, req RegistrationRequest, reasons []int) (*RegResponse, error)
	// CloseFiscalArchive закрывает фискальный режим (5.4)
	CloseFiscalArchive(ctx context.Context) (*CloseFnResult, error)

	// --- Смена (раздел 6) ---
	// OpenShift открывает смену
	OpenShift(ctx context.Context, operator string) error
	// CloseShift закрывает смену
	CloseShift(ctx context.Context, operator string) error
	// PrintXReport печатает X-отчет
	PrintXReport(ctx context.Context) error
	// PrintZReport печатает отчет по расчетам (не закрывает смену!)
	PrintZReport(ctx context.Context) error

	// --- Чеки (раздел 7) ---
	// OpenCheck открывает чек
	OpenCheck(ctx context.Context, checkType int, taxSystem int) error
	// AddPosition добавляет позицию в чек
	AddPosition(ctx context.Context, pos ItemPosition) error
	// Subtotal рассчитывает промежуточный итог
	Subtotal(ctx context.Context) error
	// Payment производит оплату
	Payment(ctx context.Context, pay PaymentInfo) error
	// CloseCheck закрывает чек
	CloseCheck(ctx context.Context) error
	// CancelCheck отменяет чек
	CancelCheck(ctx context.Context) error
	// OpenCorrectionCheck открывает чек коррекции
	OpenCorrectionCheck(ctx context.Context, checkType int, taxSystem int) error

	// --- Другие операции ---
	// RebootDevice перезапускает устройство
	RebootDevice(ctx context.Context) error
	// PrintDiagnostics печатает диагностическую информацию
	PrintDiagnostics(ctx context.Context) error
	// DeviceJob выполняет задачу устройства
	DeviceJob(ctx context.Context, job int) error
	// Feed проматывает бумагу на указанное количество строк
	Feed(ctx context.Context, lines int) error
	// Cut выполняет отрезку чека
	Cut(ctx context.Context) error
	// PrintLastDocument печатает последний сформированный документ (копию)
	PrintLastDocument(ctx context.Context) error
	// ResetMGM сбрасывает флаг МГМ
	ResetMGM(ctx context.Context) error

	// --- ОФД ---
	// OfdBeginRead начинает процедуру чтения первого непереданного документа для отправки в ОФД
	OfdBeginRead(ctx context.Context) (int, error)
	// OfdReadBlock считывает блок сообщения заданной длины, начиная с заданной позиции
	OfdReadBlock(ctx context.Context, offset, length int) ([]byte, int, error)
	// OfdEndRead завершает чтение документа
	OfdEndRead(ctx context.Context) error
	// OfdLoadReceipt записывает квитанцию от ОФД в ФН
	OfdLoadReceipt(ctx context.Context, receipt []byte) error
	// OfdCancelRead отменяет чтение документа
	OfdCancelRead(ctx context.Context) error
	// OfdReadFullDocument читает полный документ для отправки в ОФД
	OfdReadFullDocument(ctx context.Context) ([]byte, error)
}

// mitsuClient реализует интерфейс Client для работы с ККТ Mitsu
type mitsuClient struct {
	transport *Transport
	config    Config
}

// NewClient создаёт новый экземпляр Client с заданной конфигурацией подключения
func NewClient(config Config) Client {
	transport := NewTransport(config)
	return &mitsuClient{
		transport: transport,
		config:    config,
	}
}

// Connect устанавливает соединение с устройством
func (c *mitsuClient) Connect() error {
	if c.transport == nil {
		return fmt.Errorf("транспорт не инициализирован")
	}
	return c.transport.Connect()
}

// Disconnect разрывает соединение с устройством
func (c *mitsuClient) Disconnect() error {
	if c.transport == nil {
		return fmt.Errorf("транспорт не инициализирован")
	}
	return c.transport.Disconnect()
}

// SendCommand отправляет XML-команду устройству и возвращает ответ
func (c *mitsuClient) SendCommand(ctx context.Context, xmlCmd string) ([]byte, error) {
	// Проверка контекста на отмену
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if c.transport == nil {
		return nil, fmt.Errorf("транспорт не инициализирован")
	}

	// Отправка команды через транспорт
	resp, err := c.transport.Send(xmlCmd, true)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки команды: %w", err)
	}

	return resp, nil
}
