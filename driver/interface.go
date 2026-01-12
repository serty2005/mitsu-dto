package driver

import "time"

// Driver определяет интерфейс для работы с фискальными регистраторами.
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
	GetHeader(int) ([]ClicheLineData, error)
	GetLanSettings() (*LanSettings, error)
	GetOfdSettings() (*OfdSettings, error)
	GetOismSettings() (*OismSettings, error)
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
	GetPowerFlag() (bool, error) // Получить состояние флага питания
	GetOptions() (*DeviceOptions, error)
	GetCurrentDocumentType() (int, error)
	GetDocumentXMLFromFN(fd int) (string, error)

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

	Register(req RegistrationRequest) (*RegResponse, error)
	Reregister(req RegistrationRequest, reasons []int) (*RegResponse, error)
	CloseFiscalArchive() (*CloseFnResult, error)
	ResetMGM() error

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
	// TechReset выполняет технологическое обнуление (<SET FACTORY=''/>).
	TechReset() error

	Feed(lines int) error
	Cut() error
	PrintLastDocument() error

	// UploadImage загружает изображение в память ККТ.
	// index: 0 - логотип, 1-20 - пользовательские картинки.
	// data: бинарные данные BMP файла.
	UploadImage(index int, data []byte) error

	// OFD Exchange (раздел 13 документации)
	OfdBeginRead() (int, error)
	OfdReadBlock(offset, length int) ([]byte, int, error)
	OfdEndRead() error
	OfdLoadReceipt(receipt []byte) error
	OfdCancelRead() error
	OfdReadFullDocument() ([]byte, error)
}
