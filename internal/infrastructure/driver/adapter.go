package driver

import (
	"time"

	"mitsuscanner/internal/domain/models"
	"mitsuscanner/internal/domain/ports"
	legacy "mitsuscanner/pkg/mitsudriver"
)

// MitsuDriverAdapter адаптирует legacy-драйвер mitsudriver.MitsuDriver к интерфейсу ports.Driver.
type MitsuDriverAdapter struct {
	driver legacy.Driver
}

// NewMitsuDriverAdapter создает новый экземпляр MitsuDriverAdapter.
func NewMitsuDriverAdapter(driver legacy.Driver) ports.Driver {
	return &MitsuDriverAdapter{
		driver: driver,
	}
}

// SetConnectionSettings обновляет конфигурацию внутреннего драйвера
func (a *MitsuDriverAdapter) SetConnectionSettings(p models.ConnectionProfile) error {
	cfg := legacy.Config{
		ConnectionType: int32(p.ConnectionType),
		ComName:        p.ComName,
		BaudRate:       int32(p.BaudRate),
		IPAddress:      p.IPAddress,
		TCPPort:        int32(p.TCPPort),
		Timeout:        3000, // Дефолтный таймаут
		// Важно: нужно сохранить логгер, если он был
		// Но в текущей архитектуре логгер внутри legacy config не критичен,
		// так как мы логируем снаружи. Если нужно, можно прокинуть.
	}

	// Пересоздаем экземпляр старого драйвера с новым конфигом
	a.driver = legacy.NewMitsuDriver(cfg)
	return nil
}

// Connect устанавливает соединение с ККТ.
func (a *MitsuDriverAdapter) Connect() error {
	return a.driver.Connect()
}

// Disconnect разрывает соединение с ККТ.
func (a *MitsuDriverAdapter) Disconnect() error {
	return a.driver.Disconnect()
}

// GetFiscalInfo получает фискальную информацию.
func (a *MitsuDriverAdapter) GetFiscalInfo() (*models.FiscalInfo, error) {
	info, err := a.driver.GetFiscalInfo()
	if err != nil {
		return nil, err
	}
	return ConvertDriverToDomainFiscalInfo(info), nil
}

// GetModel получает модель ККТ.
func (a *MitsuDriverAdapter) GetModel() (string, error) {
	return a.driver.GetModel()
}

// GetVersion получает версию устройства.
func (a *MitsuDriverAdapter) GetVersion() (string, string, string, error) {
	return a.driver.GetVersion()
}

// GetDateTime получает дату и время ККТ.
func (a *MitsuDriverAdapter) GetDateTime() (time.Time, error) {
	return a.driver.GetDateTime()
}

// GetCashier получает данные кассира.
func (a *MitsuDriverAdapter) GetCashier() (string, string, error) {
	return a.driver.GetCashier()
}

// GetPrinterSettings получает настройки принтера.
func (a *MitsuDriverAdapter) GetPrinterSettings() (*models.PrinterSettings, error) {
	settings, err := a.driver.GetPrinterSettings()
	if err != nil {
		return nil, err
	}
	return &models.PrinterSettings{
		Model:    settings.Model,
		BaudRate: settings.BaudRate,
		Paper:    settings.Paper,
		Font:     settings.Font,
		Width:    settings.Width,
		Length:   settings.Length,
	}, nil
}

// GetMoneyDrawerSettings получает настройки денежного ящика.
func (a *MitsuDriverAdapter) GetMoneyDrawerSettings() (*models.DrawerSettings, error) {
	settings, err := a.driver.GetMoneyDrawerSettings()
	if err != nil {
		return nil, err
	}
	return &models.DrawerSettings{
		Pin:  settings.Pin,
		Rise: settings.Rise,
		Fall: settings.Fall,
	}, nil
}

// GetComSettings получает настройки COM-порта.
func (a *MitsuDriverAdapter) GetComSettings() (int32, error) {
	return a.driver.GetComSettings()
}

// GetHeader получает заголовок (клише) по номеру.
func (a *MitsuDriverAdapter) GetHeader(headerNum int) ([]models.ClicheLineData, error) {
	lines, err := a.driver.GetHeader(headerNum)
	if err != nil {
		return nil, err
	}
	return ConvertDriverToDomainClicheLineData(lines), nil
}

// GetLanSettings получает настройки сети.
func (a *MitsuDriverAdapter) GetLanSettings() (*models.LanSettings, error) {
	settings, err := a.driver.GetLanSettings()
	if err != nil {
		return nil, err
	}
	return &models.LanSettings{
		Addr: settings.Addr,
		Port: settings.Port,
		Mask: settings.Mask,
		Dns:  settings.Dns,
		Gw:   settings.Gw,
	}, nil
}

// GetOfdSettings получает настройки ОФД.
func (a *MitsuDriverAdapter) GetOfdSettings() (*models.OfdSettings, error) {
	settings, err := a.driver.GetOfdSettings()
	if err != nil {
		return nil, err
	}
	return &models.OfdSettings{
		Addr:     settings.Addr,
		Port:     settings.Port,
		Client:   settings.Client,
		TimerFN:  settings.TimerFN,
		TimerOFD: settings.TimerOFD,
	}, nil
}

// GetOismSettings получает настройки сервера OISM.
func (a *MitsuDriverAdapter) GetOismSettings() (*models.OismSettings, error) {
	settings, err := a.driver.GetOismSettings()
	if err != nil {
		return nil, err
	}
	return &models.OismSettings{
		Addr: settings.Addr,
		Port: settings.Port,
	}, nil
}

// GetOkpSettings получает настройки сервера OKP.
func (a *MitsuDriverAdapter) GetOkpSettings() (*models.ServerSettings, error) {
	settings, err := a.driver.GetOkpSettings()
	if err != nil {
		return nil, err
	}
	return &models.ServerSettings{
		Addr: settings.Addr,
		Okp:  settings.Okp,
		Port: settings.Port,
	}, nil
}

// GetTaxRates получает налоговые ставки.
func (a *MitsuDriverAdapter) GetTaxRates() (*models.TaxRates, error) {
	rates, err := a.driver.GetTaxRates()
	if err != nil {
		return nil, err
	}
	return ConvertDriverToDomainTaxRates(rates), nil
}

// GetRegistrationData получает данные о регистрации ККТ.
func (a *MitsuDriverAdapter) GetRegistrationData() (*models.RegData, error) {
	data, err := a.driver.GetRegistrationData()
	if err != nil {
		return nil, err
	}
	return ConvertDriverToDomainRegData(data), nil
}

// GetShiftStatus получает статус текущей смены.
func (a *MitsuDriverAdapter) GetShiftStatus() (*models.ShiftStatus, error) {
	status, err := a.driver.GetShiftStatus()
	if err != nil {
		return nil, err
	}
	return ConvertDriverToDomainShiftStatus(status), nil
}

// GetShiftTotals получает итоги смены.
func (a *MitsuDriverAdapter) GetShiftTotals() (*models.ShiftTotals, error) {
	totals, err := a.driver.GetShiftTotals()
	if err != nil {
		return nil, err
	}
	return ConvertDriverToDomainShiftTotals(totals), nil
}

// GetFnStatus получает статус фискального накопителя.
func (a *MitsuDriverAdapter) GetFnStatus() (*models.FnStatus, error) {
	status, err := a.driver.GetFnStatus()
	if err != nil {
		return nil, err
	}
	return ConvertDriverToDomainFnStatus(status), nil
}

// GetOfdExchangeStatus получает статус обмена с ОФД.
func (a *MitsuDriverAdapter) GetOfdExchangeStatus() (*models.OfdExchangeStatus, error) {
	status, err := a.driver.GetOfdExchangeStatus()
	if err != nil {
		return nil, err
	}
	return ConvertDriverToDomainOfdExchangeStatus(status), nil
}

// GetMarkingStatus получает статус маркировки.
func (a *MitsuDriverAdapter) GetMarkingStatus() (*models.MarkingStatus, error) {
	status, err := a.driver.GetMarkingStatus()
	if err != nil {
		return nil, err
	}
	return ConvertDriverToDomainMarkingStatus(status), nil
}

// GetTimezone получает временную зону.
func (a *MitsuDriverAdapter) GetTimezone() (int, error) {
	return a.driver.GetTimezone()
}

// GetPowerStatus получает статус питания.
func (a *MitsuDriverAdapter) GetPowerStatus() (int, error) {
	return a.driver.GetPowerStatus()
}

// GetPowerFlag получает состояние флага питания ФН.
func (a *MitsuDriverAdapter) GetPowerFlag() (bool, error) {
	return a.driver.GetPowerFlag()
}

// GetOptions получает настройки устройства (b0-b9).
func (a *MitsuDriverAdapter) GetOptions() (*models.DeviceOptions, error) {
	options, err := a.driver.GetOptions()
	if err != nil {
		return nil, err
	}
	result := ConvertDriverToDomainDeviceOptions(*options)
	return &result, nil
}

// GetCurrentDocumentType получает тип текущего документа.
func (a *MitsuDriverAdapter) GetCurrentDocumentType() (int, error) {
	return a.driver.GetCurrentDocumentType()
}

// GetDocumentXMLFromFN получает XML-документ из ФН по номеру.
func (a *MitsuDriverAdapter) GetDocumentXMLFromFN(fd int) (string, error) {
	return a.driver.GetDocumentXMLFromFN(fd)
}

// SetPowerFlag устанавливает состояние флага питания ФН.
func (a *MitsuDriverAdapter) SetPowerFlag(value int) error {
	return a.driver.SetPowerFlag(value)
}

// SetDateTime устанавливает дату и время ККТ.
func (a *MitsuDriverAdapter) SetDateTime(t time.Time) error {
	return a.driver.SetDateTime(t)
}

// SetCashier устанавливает данные кассира.
func (a *MitsuDriverAdapter) SetCashier(name string, inn string) error {
	return a.driver.SetCashier(name, inn)
}

// SetComSettings устанавливает настройки COM-порта.
func (a *MitsuDriverAdapter) SetComSettings(speed int32) error {
	return a.driver.SetComSettings(speed)
}

// SetPrinterSettings устанавливает настройки принтера.
func (a *MitsuDriverAdapter) SetPrinterSettings(settings models.PrinterSettings) error {
	driverSettings := ConvertDomainToDriverPrinterSettings(settings)
	return a.driver.SetPrinterSettings(driverSettings)
}

// SetMoneyDrawerSettings устанавливает настройки денежного ящика.
func (a *MitsuDriverAdapter) SetMoneyDrawerSettings(settings models.DrawerSettings) error {
	driverSettings := ConvertDomainToDriverDrawerSettings(settings)
	return a.driver.SetMoneyDrawerSettings(driverSettings)
}

// SetHeader устанавливает заголовок (клише) по номеру.
func (a *MitsuDriverAdapter) SetHeader(headerNum int, lines []models.ClicheLineData) error {
	driverLines := ConvertDomainToDriverClicheLineData(lines)
	return a.driver.SetHeader(headerNum, driverLines)
}

// SetHeaderLine устанавливает отдельную строку заголовка (клише).
func (a *MitsuDriverAdapter) SetHeaderLine(headerNum int, lineNum int, text string, format string) error {
	return a.driver.SetHeaderLine(headerNum, lineNum, text, format)
}

// SetLanSettings устанавливает настройки сети.
func (a *MitsuDriverAdapter) SetLanSettings(settings models.LanSettings) error {
	driverSettings := ConvertDomainToDriverLanSettings(settings)
	return a.driver.SetLanSettings(driverSettings)
}

// SetOfdSettings устанавливает настройки ОФД.
func (a *MitsuDriverAdapter) SetOfdSettings(settings models.OfdSettings) error {
	driverSettings := ConvertDomainToDriverOfdSettings(settings)
	return a.driver.SetOfdSettings(driverSettings)
}

// SetOismSettings устанавливает настройки сервера OISM.
func (a *MitsuDriverAdapter) SetOismSettings(settings models.OismSettings) error {
	driverSettings := ConvertDomainToDriverOismSettings(settings)
	return a.driver.SetOismSettings(driverSettings)
}

// SetOkpSettings устанавливает настройки сервера OKP.
func (a *MitsuDriverAdapter) SetOkpSettings(settings models.ServerSettings) error {
	driverSettings := ConvertDomainToDriverOkpSettings(settings)
	return a.driver.SetOkpSettings(driverSettings)
}

// SetOption устанавливает отдельную опцию устройства.
func (a *MitsuDriverAdapter) SetOption(optionNum int, value int) error {
	return a.driver.SetOption(optionNum, value)
}

// SetTimezone устанавливает временную зону.
func (a *MitsuDriverAdapter) SetTimezone(value int) error {
	return a.driver.SetTimezone(value)
}

// Register регистрирует ККТ.
func (a *MitsuDriverAdapter) Register(req models.RegistrationRequest) (*models.RegResponse, error) {
	driverReq := ConvertDomainToDriverRegistrationRequest(req)
	resp, err := a.driver.Register(driverReq)
	if err != nil {
		return nil, err
	}
	return ConvertDriverToDomainRegResponse(resp), nil
}

// Reregister перерегистрирует ККТ.
func (a *MitsuDriverAdapter) Reregister(req models.RegistrationRequest, reasons []int) (*models.RegResponse, error) {
	driverReq := ConvertDomainToDriverRegistrationRequest(req)
	resp, err := a.driver.Reregister(driverReq, reasons)
	if err != nil {
		return nil, err
	}
	return ConvertDriverToDomainRegResponse(resp), nil
}

// CloseFiscalArchive закрывает фискальный архив.
func (a *MitsuDriverAdapter) CloseFiscalArchive() (*models.CloseFnResult, error) {
	result, err := a.driver.CloseFiscalArchive()
	if err != nil {
		return nil, err
	}
	return ConvertDriverToDomainCloseFnResult(result), nil
}

// ResetMGM сбрасывает МГМ (Модуль Грузки Маркировки).
func (a *MitsuDriverAdapter) ResetMGM() error {
	return a.driver.ResetMGM()
}

// OpenShift открывает смену.
func (a *MitsuDriverAdapter) OpenShift(operator string) error {
	return a.driver.OpenShift(operator)
}

// CloseShift закрывает смену.
func (a *MitsuDriverAdapter) CloseShift(operator string) error {
	return a.driver.CloseShift(operator)
}

// PrintXReport печатает X-отчет.
func (a *MitsuDriverAdapter) PrintXReport() error {
	return a.driver.PrintXReport()
}

// PrintZReport печатает Z-отчет.
func (a *MitsuDriverAdapter) PrintZReport() error {
	return a.driver.PrintZReport()
}

// OpenCheck открывает чек.
func (a *MitsuDriverAdapter) OpenCheck(checkType int, taxSystem int) error {
	return a.driver.OpenCheck(checkType, taxSystem)
}

// AddPosition добавляет позицию в чек.
func (a *MitsuDriverAdapter) AddPosition(pos models.ItemPosition) error {
	driverPos := ConvertDomainToDriverItemPosition(pos)
	return a.driver.AddPosition(driverPos)
}

// Subtotal выполняет промежуточный итог.
func (a *MitsuDriverAdapter) Subtotal() error {
	return a.driver.Subtotal()
}

// Payment выполняет оплату.
func (a *MitsuDriverAdapter) Payment(pay models.PaymentInfo) error {
	driverPay := ConvertDomainToDriverPaymentInfo(pay)
	return a.driver.Payment(driverPay)
}

// CloseCheck закрывает чек.
func (a *MitsuDriverAdapter) CloseCheck() error {
	return a.driver.CloseCheck()
}

// CancelCheck отменяет чек.
func (a *MitsuDriverAdapter) CancelCheck() error {
	return a.driver.CancelCheck()
}

// OpenCorrectionCheck открывает корректировочный чек.
func (a *MitsuDriverAdapter) OpenCorrectionCheck(checkType int, taxSystem int) error {
	return a.driver.OpenCorrectionCheck(checkType, taxSystem)
}

// RebootDevice перезагружает устройство.
func (a *MitsuDriverAdapter) RebootDevice() error {
	return a.driver.RebootDevice()
}

// PrintDiagnostics печатает диагностику.
func (a *MitsuDriverAdapter) PrintDiagnostics() error {
	return a.driver.PrintDiagnostics()
}

// DeviceJob выполняет устройственную операцию.
func (a *MitsuDriverAdapter) DeviceJob(job int) error {
	return a.driver.DeviceJob(job)
}

// TechReset выполняет технологическое обнуление.
func (a *MitsuDriverAdapter) TechReset() error {
	return a.driver.TechReset()
}

// Feed пропускает бумагу на указанное количество строк.
func (a *MitsuDriverAdapter) Feed(lines int) error {
	return a.driver.Feed(lines)
}

// Cut выполняет резку бумаги.
func (a *MitsuDriverAdapter) Cut() error {
	return a.driver.Cut()
}

// PrintLastDocument печатает последний документ.
func (a *MitsuDriverAdapter) PrintLastDocument() error {
	return a.driver.PrintLastDocument()
}

// UploadImage загружает изображение в память ККТ.
func (a *MitsuDriverAdapter) UploadImage(index int, data []byte) error {
	return a.driver.UploadImage(index, data)
}

// OfdBeginRead начинает чтение данных для обмена с ОФД.
func (a *MitsuDriverAdapter) OfdBeginRead() (int, error) {
	return a.driver.OfdBeginRead()
}

// OfdReadBlock читает блок данных для обмена с ОФД.
func (a *MitsuDriverAdapter) OfdReadBlock(offset, length int) ([]byte, int, error) {
	return a.driver.OfdReadBlock(offset, length)
}

// OfdEndRead завершает чтение данных для обмена с ОФД.
func (a *MitsuDriverAdapter) OfdEndRead() error {
	return a.driver.OfdEndRead()
}

// OfdLoadReceipt загружает чек для обмена с ОФД.
func (a *MitsuDriverAdapter) OfdLoadReceipt(receipt []byte) error {
	return a.driver.OfdLoadReceipt(receipt)
}

// OfdCancelRead отменяет чтение данных для обмена с ОФД.
func (a *MitsuDriverAdapter) OfdCancelRead() error {
	return a.driver.OfdCancelRead()
}

// OfdReadFullDocument читает полный документ для обмена с ОФД.
func (a *MitsuDriverAdapter) OfdReadFullDocument() ([]byte, error) {
	return a.driver.OfdReadFullDocument()
}
