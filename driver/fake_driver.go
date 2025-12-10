package driver

import "time"

func NewFakeDriver() Driver {
	return &fakeDriver{}
}

type fakeDriver struct{}

func (d *fakeDriver) Connect() error {
	return nil
}

func (d *fakeDriver) Disconnect() error {
	return nil
}

func (d *fakeDriver) GetFiscalInfo() (*FiscalInfo, error) {
	return &FiscalInfo{}, nil
}

func (d *fakeDriver) GetModel() (string, error) {
	return "FAKE", nil
}

func (d *fakeDriver) GetVersion() (string, string, string, error) {
	return "1", "2", "3", nil
}

func (d *fakeDriver) GetDateTime() (time.Time, error) {
	return time.Now(), nil
}

func (d *fakeDriver) GetCashier() (string, string, error) {
	return "Иванов", "1234567890", nil
}

func (d *fakeDriver) GetPrinterSettings() (*PrinterSettings, error) {
	return &PrinterSettings{}, nil
}

func (d *fakeDriver) GetMoneyDrawerSettings() (*DrawerSettings, error) {
	return &DrawerSettings{}, nil
}

func (d *fakeDriver) GetComSettings() (int32, error) {
	return 115200, nil
}

func (d *fakeDriver) GetHeader(int) ([]string, error) {
	return []string{"Header1", "Header2"}, nil
}

func (d *fakeDriver) GetLanSettings() (*LanSettings, error) {
	return &LanSettings{
		Addr: "192.168.1.1",
		Port: 8200,
		Mask: "255.255.255.0",
		Dns:  "8.8.8.8",
		Gw:   "192.168.1.254",
	}, nil
}

func (d *fakeDriver) GetOfdSettings() (*OfdSettings, error) {
	return &OfdSettings{
		Addr:     "ofd.example.com",
		Port:     8200,
		Client:   "Client123",
		TimerFN:  30,
		TimerOFD: 60,
	}, nil
}

func (d *fakeDriver) GetOismSettings() (*ServerSettings, error) {
	return &ServerSettings{
		Addr: "oism.example.com",
		Port: 8200,
	}, nil
}

func (d *fakeDriver) GetOkpSettings() (*ServerSettings, error) {
	return &ServerSettings{}, nil
}

func (d *fakeDriver) GetTaxRates() (*TaxRates, error) {
	return &TaxRates{}, nil
}

func (d *fakeDriver) GetRegistrationData() (*RegData, error) {
	return &RegData{}, nil
}

func (d *fakeDriver) GetShiftStatus() (*ShiftStatus, error) {
	return &ShiftStatus{}, nil
}

func (d *fakeDriver) GetShiftTotals() (*ShiftTotals, error) {
	return &ShiftTotals{}, nil
}

func (d *fakeDriver) GetFnStatus() (*FnStatus, error) {
	return &FnStatus{}, nil
}

func (d *fakeDriver) GetOfdExchangeStatus() (*OfdExchangeStatus, error) {
	return &OfdExchangeStatus{}, nil
}

func (d *fakeDriver) GetMarkingStatus() (*MarkingStatus, error) {
	return &MarkingStatus{}, nil
}

func (d *fakeDriver) GetTimezone() (int, error) {
	return 3, nil
}

func (d *fakeDriver) GetPowerStatus() (int, error) {
	return 0, nil
}

func (d *fakeDriver) SetPowerFlag(value int) error {
	return nil
}

func (d *fakeDriver) SetDateTime(t time.Time) error {
	return nil
}

func (d *fakeDriver) SetCashier(name string, inn string) error {
	return nil
}

func (d *fakeDriver) SetComSettings(speed int32) error {
	return nil
}

func (d *fakeDriver) SetPrinterSettings(settings PrinterSettings) error {
	return nil
}

func (d *fakeDriver) SetMoneyDrawerSettings(settings DrawerSettings) error {
	return nil
}

func (d *fakeDriver) SetHeaderLine(headerNum int, lineNum int, text string, format string) error {
	return nil
}

func (d *fakeDriver) SetLanSettings(settings LanSettings) error {
	return nil
}

func (d *fakeDriver) SetOfdSettings(settings OfdSettings) error {
	return nil
}

func (d *fakeDriver) SetOismSettings(settings ServerSettings) error {
	return nil
}

func (d *fakeDriver) SetOkpSettings(settings ServerSettings) error {
	return nil
}

func (d *fakeDriver) SetOption(optionNum int, value int) error {
	return nil
}

func (d *fakeDriver) SetTimezone(value int) error {
	return nil
}

func (d *fakeDriver) Register(req RegistrationRequest) error {
	return nil
}

func (d *fakeDriver) Reregister(req RegistrationRequest, reasons []int) error {
	return nil
}

func (d *fakeDriver) CloseFiscalArchive() error {
	return nil
}

func (d *fakeDriver) OpenShift(operator string) error {
	return nil
}

func (d *fakeDriver) CloseShift(operator string) error {
	return nil
}

func (d *fakeDriver) PrintXReport() error {
	return nil
}

func (d *fakeDriver) PrintZReport() error {
	return nil
}

func (d *fakeDriver) OpenCheck(checkType int, taxSystem int) error {
	return nil
}

func (d *fakeDriver) AddPosition(pos ItemPosition) error {
	return nil
}

func (d *fakeDriver) Subtotal() error {
	return nil
}

func (d *fakeDriver) Payment(pay PaymentInfo) error {
	return nil
}

func (d *fakeDriver) CloseCheck() error {
	return nil
}

func (d *fakeDriver) CancelCheck() error {
	return nil
}

func (d *fakeDriver) OpenCorrectionCheck(checkType int, taxSystem int) error {
	return nil
}

func (d *fakeDriver) RebootDevice() error {
	return nil
}

func (d *fakeDriver) PrintDiagnostics() error {
	return nil
}

func (d *fakeDriver) DeviceJob(job int) error {
	return nil
}

func (d *fakeDriver) Feed(lines int) error {
	return nil
}

func (d *fakeDriver) Cut() error {
	return nil
}

func (d *fakeDriver) PrintLastDocument() error {
	return nil
}
