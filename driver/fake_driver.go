package driver

import (
	"fmt"
	"time"
)

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

// --- INFO & STATUS ---

func (d *fakeDriver) GetFiscalInfo() (*FiscalInfo, error) {
	return &FiscalInfo{
		ModelName:        "Mitsu 1-F (EMUL)",
		SerialNumber:     "00000000001",
		RNM:              "1234567890123456",
		OrganizationName: "ООО Ромашка",
		Inn:              "1234567890",
		Address:          "г. Москва, ул. Ленина, д. 1",
		OfdName:          "Яндекс.ОФД",
		RegistrationDate: "2023-01-01",
		FfdVersion:       "1.2",
		FnEndDate:        "2025-01-01",
		FnSerial:         "9999078900000001",
	}, nil
}

func (d *fakeDriver) GetModel() (string, error) {
	return "Mitsu 1-F", nil
}

func (d *fakeDriver) GetVersion() (string, string, string, error) {
	return "1.2.3", "001", "00:11:22:33:44:55", nil
}

func (d *fakeDriver) GetDateTime() (time.Time, error) {
	return time.Now(), nil
}

func (d *fakeDriver) GetCashier() (string, string, error) {
	return "Иванов И.И.", "123456789012", nil
}

func (d *fakeDriver) GetShiftStatus() (*ShiftStatus, error) {
	return &ShiftStatus{
		ShiftNum: 1,
		State:    "1", // Открыта
		Count:    5,
		FdNum:    100,
	}, nil
}

func (d *fakeDriver) GetShiftTotals() (*ShiftTotals, error) {
	return &ShiftTotals{}, nil
}

func (d *fakeDriver) GetFnStatus() (*FnStatus, error) {
	return &FnStatus{
		Serial: "9999078900000001",
		Valid:  "2025-01-01",
		Ffd:    "1.2",
	}, nil
}

func (d *fakeDriver) GetOfdExchangeStatus() (*OfdExchangeStatus, error) {
	return &OfdExchangeStatus{
		Count: 0, // Все отправлено
	}, nil
}

func (d *fakeDriver) GetMarkingStatus() (*MarkingStatus, error) {
	return &MarkingStatus{}, nil
}

func (d *fakeDriver) GetPowerStatus() (int, error) {
	return 0, nil
}

func (d *fakeDriver) GetTimezone() (int, error) {
	return 3, nil
}

// --- SETTINGS (GET) ---

func (d *fakeDriver) GetPrinterSettings() (*PrinterSettings, error) {
	return &PrinterSettings{
		Model:    "1", // RP-809
		BaudRate: 115200,
		Paper:    80,
		Font:     0,
	}, nil
}

func (d *fakeDriver) GetMoneyDrawerSettings() (*DrawerSettings, error) {
	return &DrawerSettings{
		Pin:  5,
		Rise: 100,
		Fall: 100,
	}, nil
}

func (d *fakeDriver) GetComSettings() (int32, error) {
	return 115200, nil
}

func (d *fakeDriver) GetHeader(n int) ([]ClicheLineData, error) {
	// Возвращаем заглушку на 10 строк
	res := make([]ClicheLineData, 10)
	for i := 0; i < 10; i++ {
		res[i] = ClicheLineData{
			Text:   fmt.Sprintf("Строка %d (Тип %d)", i+1, n),
			Format: "000000",
		}
	}
	if n == 1 {
		res[0].Text = "ООО \"РОМАШКА\""
		res[0].Format = "011001" // Пример: центр, увеличенный
		res[1].Text = "ДОБРО ПОЖАЛОВАТЬ"
	}
	return res, nil
}

func (d *fakeDriver) GetLanSettings() (*LanSettings, error) {
	return &LanSettings{
		Addr: "192.168.1.100",
		Port: 8200,
		Mask: "255.255.255.0",
		Dns:  "8.8.8.8",
		Gw:   "192.168.1.1",
	}, nil
}

func (d *fakeDriver) GetOfdSettings() (*OfdSettings, error) {
	return &OfdSettings{
		Addr:     "ofd.yandex.ru",
		Port:     21101,
		Client:   "MitsuClient",
		TimerFN:  30,
		TimerOFD: 60,
	}, nil
}

func (d *fakeDriver) GetOismSettings() (*ServerSettings, error) {
	return &ServerSettings{
		Addr: "oism.crpt.ru",
		Port: 80,
	}, nil
}

func (d *fakeDriver) GetOkpSettings() (*ServerSettings, error) {
	return &ServerSettings{
		Addr: "okp.example.com",
		Port: 80,
	}, nil
}

func (d *fakeDriver) GetTaxRates() (*TaxRates, error) {
	return &TaxRates{}, nil
}

func (d *fakeDriver) GetRegistrationData() (*RegData, error) {
	return &RegData{
		RNM:        "1234567890123456",
		Inn:        "7701010101",
		OrgName:    "MOCK ORG",
		Address:    "MOCK ADDRESS",
		OfdName:    "MOCK OFD",
		MarkAttr:   "1",
		TaxSystems: "1,2", // УСН
	}, nil
}

func (d *fakeDriver) GetOptions() (*DeviceOptions, error) {
	return &DeviceOptions{
		B0: 0, B1: 1, B2: 0, B3: 1, B4: 1,
		B5: 1, B6: 0, B7: 0, B8: 0, B9: 0,
	}, nil
}

// --- SETTINGS (SET) ---

func (d *fakeDriver) SetPowerFlag(value int) error                         { return nil }
func (d *fakeDriver) SetDateTime(t time.Time) error                        { return nil }
func (d *fakeDriver) SetCashier(name string, inn string) error             { return nil }
func (d *fakeDriver) SetComSettings(speed int32) error                     { return nil }
func (d *fakeDriver) SetPrinterSettings(settings PrinterSettings) error    { return nil }
func (d *fakeDriver) SetMoneyDrawerSettings(settings DrawerSettings) error { return nil }
func (d *fakeDriver) SetHeaderLine(headerNum int, lineNum int, text string, format string) error {
	return nil
}
func (d *fakeDriver) SetLanSettings(settings LanSettings) error     { return nil }
func (d *fakeDriver) SetOfdSettings(settings OfdSettings) error     { return nil }
func (d *fakeDriver) SetOismSettings(settings ServerSettings) error { return nil }
func (d *fakeDriver) SetOkpSettings(settings ServerSettings) error  { return nil }
func (d *fakeDriver) SetOption(optionNum int, value int) error      { return nil }
func (d *fakeDriver) SetTimezone(value int) error                   { return nil }

// --- OPERATIONS ---

func (d *fakeDriver) Register(req RegistrationRequest) error                  { return nil }
func (d *fakeDriver) Reregister(req RegistrationRequest, reasons []int) error { return nil }
func (d *fakeDriver) CloseFiscalArchive() error                               { return nil }

func (d *fakeDriver) OpenShift(operator string) error                        { return nil }
func (d *fakeDriver) CloseShift(operator string) error                       { return nil }
func (d *fakeDriver) PrintXReport() error                                    { return nil }
func (d *fakeDriver) PrintZReport() error                                    { return nil }
func (d *fakeDriver) OpenCheck(checkType int, taxSystem int) error           { return nil }
func (d *fakeDriver) AddPosition(pos ItemPosition) error                     { return nil }
func (d *fakeDriver) Subtotal() error                                        { return nil }
func (d *fakeDriver) Payment(pay PaymentInfo) error                          { return nil }
func (d *fakeDriver) CloseCheck() error                                      { return nil }
func (d *fakeDriver) CancelCheck() error                                     { return nil }
func (d *fakeDriver) OpenCorrectionCheck(checkType int, taxSystem int) error { return nil }
func (d *fakeDriver) RebootDevice() error                                    { return nil }
func (d *fakeDriver) PrintDiagnostics() error                                { return nil }
func (d *fakeDriver) DeviceJob(job int) error                                { return nil }
func (d *fakeDriver) Feed(lines int) error                                   { return nil }
func (d *fakeDriver) Cut() error                                             { return nil }
func (d *fakeDriver) PrintLastDocument() error                               { return nil }
func (d *fakeDriver) TechReset() error {
	fmt.Println("[MOCK] Tech Reset performed")
	return nil
}

func (d *fakeDriver) UploadImage(index int, data []byte) error {
	fmt.Printf("[MOCK] UploadImage Index=%d, Size=%d bytes\n", index, len(data))
	time.Sleep(500 * time.Millisecond) // Имитация задержки передачи
	return nil
}
