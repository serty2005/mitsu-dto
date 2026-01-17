package ports

import (
	"time"

	"mitsuscanner/internal/domain/models"
)

// Driver определяет интерфейс для работы с фискальными регистраторами.
// Все методы используют чистые доменные типы, не зависящие от внешних библиотек.
type Driver interface {
	SetConnectionSettings(profile models.ConnectionProfile) error
	Connect() error
	Disconnect() error
	GetFiscalInfo() (*models.FiscalInfo, error)
	GetModel() (string, error)
	GetVersion() (string, string, string, error)
	GetDateTime() (time.Time, error)
	GetCashier() (string, string, error)
	GetPrinterSettings() (*models.PrinterSettings, error)
	GetMoneyDrawerSettings() (*models.DrawerSettings, error)
	GetComSettings() (int32, error)
	GetHeader(int) ([]models.ClicheLineData, error)
	GetLanSettings() (*models.LanSettings, error)
	GetOfdSettings() (*models.OfdSettings, error)
	GetOismSettings() (*models.OismSettings, error)
	GetOkpSettings() (*models.ServerSettings, error)
	GetTaxRates() (*models.TaxRates, error)
	GetRegistrationData() (*models.RegData, error)
	GetShiftStatus() (*models.ShiftStatus, error)
	GetShiftTotals() (*models.ShiftTotals, error)
	GetFnStatus() (*models.FnStatus, error)
	GetOfdExchangeStatus() (*models.OfdExchangeStatus, error)
	GetMarkingStatus() (*models.MarkingStatus, error)
	GetTimezone() (int, error)
	GetPowerStatus() (int, error)
	GetPowerFlag() (bool, error) // Получить состояние флага питания ФН
	GetOptions() (*models.DeviceOptions, error)
	GetCurrentDocumentType() (int, error)
	GetDocumentXMLFromFN(fd int) (string, error)

	SetPowerFlag(value int) error
	SetDateTime(t time.Time) error
	SetCashier(name string, inn string) error
	SetComSettings(speed int32) error
	SetPrinterSettings(settings models.PrinterSettings) error
	SetMoneyDrawerSettings(settings models.DrawerSettings) error
	SetHeader(headerNum int, lines []models.ClicheLineData) error
	SetHeaderLine(headerNum int, lineNum int, text string, format string) error
	SetLanSettings(settings models.LanSettings) error
	SetOfdSettings(settings models.OfdSettings) error
	SetOismSettings(settings models.OismSettings) error
	SetOkpSettings(settings models.ServerSettings) error
	SetOption(optionNum int, value int) error
	SetTimezone(value int) error

	Register(req models.RegistrationRequest) (*models.RegResponse, error)
	Reregister(req models.RegistrationRequest, reasons []int) (*models.RegResponse, error)
	CloseFiscalArchive() (*models.CloseFnResult, error)
	ResetMGM() error

	OpenShift(operator string) error
	CloseShift(operator string) error
	PrintXReport() error
	PrintZReport() error
	OpenCheck(checkType int, taxSystem int) error
	AddPosition(pos models.ItemPosition) error
	Subtotal() error
	Payment(pay models.PaymentInfo) error
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
