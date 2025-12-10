package driver

import (
	"errors"
	"fmt"
	"time"
)

// GetFiscalInfo собирает полную информацию о ККТ, последовательно вызывая команды протокола.
func (d *mitsuDriver) GetFiscalInfo() (*FiscalInfo, error) {
	info := &FiscalInfo{}

	// 1. Получение модели (<GET DEV='?'/>)
	// Стр. 9
	if resp, err := d.sendCommand("<GET DEV='?'/>"); err != nil {
		return nil, fmt.Errorf("ошибка получения модели: %w", err)
	} else {
		var r struct {
			Dev string `xml:"DEV,attr"`
		}
		if err := decodeXML(resp, &r); err == nil {
			info.ModelName = r.Dev
		}
	}

	// 2. Получение версии и серийного номера (<GET VER='?'/>)
	// Стр. 10
	if resp, err := d.sendCommand("<GET VER='?'/>"); err != nil {
		return nil, fmt.Errorf("ошибка получения версии: %w", err)
	} else {
		var r struct {
			Serial string `xml:"SERIAL,attr"`
			Ver    string `xml:"VER,attr"`
		}
		if err := decodeXML(resp, &r); err == nil {
			info.SerialNumber = r.Serial
			info.SoftwareDate = r.Ver // В Mitsu версия строковая (напр. "1.2.02")
		}
	}

	// 3. Получение регистрационных данных (<GET REG='?'/>)
	// Стр. 11-12
	// Эта команда возвращает основные параметры регистрации.
	if resp, err := d.sendCommand("<GET REG='?'/>"); err != nil {
		return nil, fmt.Errorf("ошибка получения рег. данных: %w", err)
	} else {
		var r struct {
			// Атрибуты
			Rnm        string `xml:"T1037,attr"`
			Inn        string `xml:"T1018,attr"`
			FfdVer     string `xml:"T1209,attr"`
			RegDate    string `xml:"DATE,attr"`
			MarkAttr   string `xml:"MARK,attr"`  // Признак маркировки ('1' - да)
			ExciseAttr string `xml:"T1207,attr"` // Признак подакцизных ('1' - да)
			// Вложенные теги
			OrgName string `xml:"T1048"`
			Address string `xml:"T1009"`
			OfdName string `xml:"T1046"`
		}
		if err := decodeXML(resp, &r); err != nil {
			return nil, fmt.Errorf("ошибка разбора рег. данных: %w", err)
		}

		info.RNM = r.Rnm
		info.Inn = r.Inn
		info.FfdVersion = r.FfdVer
		info.RegistrationDate = r.RegDate
		info.OrganizationName = r.OrgName
		info.Address = r.Address
		info.OfdName = r.OfdName

		// Обработка флагов (в XML они приходят как "1" или "0")
		info.AttributeMarked = r.MarkAttr == "1"
		info.AttributeExcise = r.ExciseAttr == "1"
	}

	// 4. Получение статуса ФН (<GET INFO='F'/>)
	// Стр. 15. Команда может принимать вид INFO='F' или INFO='FN'
	if resp, err := d.sendCommand("<GET INFO='F'/>"); err != nil {
		return nil, fmt.Errorf("ошибка получения статуса ФН: %w", err)
	} else {
		var r struct {
			FnSerial string `xml:"FN,attr"`    // Заводской номер ФН (атрибут FN, см пример стр 15)
			FnValid  string `xml:"VALID,attr"` // Срок действия
			FnFfd    string `xml:"FFD,attr"`   // Версия ФФД ФН
			Edition  string `xml:"EDITION,attr"`
		}
		if err := decodeXML(resp, &r); err != nil {
			return nil, fmt.Errorf("ошибка разбора статуса ФН: %w", err)
		}
		info.FnSerial = r.FnSerial
		info.FnEndDate = r.FnValid
		info.FnExecution = r.FnFfd
		info.FnEdition = r.Edition
	}

	return info, nil
}

// --- Реализация методов GET ---

// GetModel (3.3)
func (d *mitsuDriver) GetModel() (string, error) {
	resp, err := d.sendCommand("<GET DEV='?'/>")
	if err != nil {
		return "", err
	}
	var r struct {
		Dev string `xml:"DEV,attr"`
	}
	if err := decodeXML(resp, &r); err != nil {
		return "", err
	}
	return r.Dev, nil
}

// GetVersion (3.4)
func (d *mitsuDriver) GetVersion() (string, string, string, error) {
	resp, err := d.sendCommand("<GET VER='?'/>")
	if err != nil {
		return "", "", "", err
	}
	var r struct {
		Serial string `xml:"SERIAL,attr"`
		Ver    string `xml:"VER,attr"`
		Mac    string `xml:"MAC,attr"`
	}
	if err := decodeXML(resp, &r); err != nil {
		return "", "", "", err
	}
	return r.Ver, r.Serial, r.Mac, nil
}

// GetDateTime (3.5)
func (d *mitsuDriver) GetDateTime() (time.Time, error) {
	// Запрашиваем дату и время одной командой
	resp, err := d.sendCommand("<GET DATE='?' TIME='?'/>")
	if err != nil {
		return time.Time{}, err
	}
	var r struct {
		Date string `xml:"DATE,attr"` // гггг-мм-дд
		Time string `xml:"TIME,attr"` // чч:мм:сс
	}
	if err := decodeXML(resp, &r); err != nil {
		return time.Time{}, err
	}

	// Парсинг
	fullTime := fmt.Sprintf("%sT%s", r.Date, r.Time)
	return time.Parse("2006-01-02T15:04:05", fullTime)
}

// GetCashier (3.6)
func (d *mitsuDriver) GetCashier() (string, string, error) {
	resp, err := d.sendCommand("<GET CASHIER='?'/>")
	if err != nil {
		return "", "", err
	}
	var r struct {
		Name string `xml:"CASHIER,attr"`
		Inn  string `xml:"INN,attr"`
	}
	if err := decodeXML(resp, &r); err != nil {
		return "", "", err
	}
	return r.Name, r.Inn, nil
}

// GetPrinterSettings (3.7)
func (d *mitsuDriver) GetPrinterSettings() (*PrinterSettings, error) {
	resp, err := d.sendCommand("<GET PRINTER='?'/>")
	if err != nil {
		return nil, err
	}
	var s PrinterSettings
	if err := decodeXML(resp, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetMoneyDrawerSettings (3.8)
func (d *mitsuDriver) GetMoneyDrawerSettings() (*DrawerSettings, error) {
	resp, err := d.sendCommand("<GET CD='?'/>")
	if err != nil {
		return nil, err
	}
	// Тут нужно быть внимательным с XML.
	// Ответ: <OK CD='контакт' RISE='фронт' FALL='спад' />
	// Но пример: <OK CD:PIN='5' ... /> - это опечатка в доке или реальный формат?
	// Судя по разделу 3.8: <OK CD='контакт' ...>
	// Попробуем распарсить стандартно.
	var s DrawerSettings
	if err := decodeXML(resp, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetComSettings (3.9)
func (d *mitsuDriver) GetComSettings() (int32, error) {
	resp, err := d.sendCommand("<GET COM='?'/>")
	if err != nil {
		return 0, err
	}
	var r struct {
		Speed int32 `xml:"COM,attr"`
	}
	if err := decodeXML(resp, &r); err != nil {
		return 0, err
	}
	return r.Speed, nil
}

// GetHeader (3.10)
func (d *mitsuDriver) GetHeader(n int) ([]string, error) {
	cmd := fmt.Sprintf("<GET HEADER='%d'/>", n)
	resp, err := d.sendCommand(cmd)
	if err != nil {
		return nil, err
	}

	// Структура ответа сложная: <OK ...><L0 ...>Текст</L0><L1 ...>Текст</L1></OK>
	// Используем вспомогательную структуру для парсинга вложенных тегов
	type Line struct {
		Text string `xml:",chardata"`
	}
	type HeaderResp struct {
		L0 Line `xml:"L0"`
		L1 Line `xml:"L1"`
		L2 Line `xml:"L2"`
		L3 Line `xml:"L3"`
		L4 Line `xml:"L4"`
		L5 Line `xml:"L5"`
		L6 Line `xml:"L6"`
		L7 Line `xml:"L7"`
		L8 Line `xml:"L8"`
		L9 Line `xml:"L9"`
	}
	var r HeaderResp
	if err := decodeXML(resp, &r); err != nil {
		return nil, err
	}

	// Собираем непустые строки
	var lines []string
	if r.L0.Text != "" {
		lines = append(lines, r.L0.Text)
	}
	if r.L1.Text != "" {
		lines = append(lines, r.L1.Text)
	}
	if r.L2.Text != "" {
		lines = append(lines, r.L2.Text)
	}
	if r.L3.Text != "" {
		lines = append(lines, r.L3.Text)
	}
	if r.L4.Text != "" {
		lines = append(lines, r.L4.Text)
	}
	if r.L5.Text != "" {
		lines = append(lines, r.L5.Text)
	}
	if r.L6.Text != "" {
		lines = append(lines, r.L6.Text)
	}
	if r.L7.Text != "" {
		lines = append(lines, r.L7.Text)
	}
	if r.L8.Text != "" {
		lines = append(lines, r.L8.Text)
	}
	if r.L9.Text != "" {
		lines = append(lines, r.L9.Text)
	}

	return lines, nil
}

// GetLanSettings (3.11)
func (d *mitsuDriver) GetLanSettings() (*LanSettings, error) {
	resp, err := d.sendCommand("<GET LAN='?'/>")
	if err != nil {
		return nil, err
	}
	var s LanSettings
	if err := decodeXML(resp, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetOfdSettings (3.12)
func (d *mitsuDriver) GetOfdSettings() (*OfdSettings, error) {
	resp, err := d.sendCommand("<GET OFD='?'/>")
	if err != nil {
		return nil, err
	}
	var s OfdSettings
	if err := decodeXML(resp, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetOismSettings (3.13)
func (d *mitsuDriver) GetOismSettings() (*ServerSettings, error) {
	resp, err := d.sendCommand("<GET OISM='?'/>")
	if err != nil {
		return nil, err
	}
	var s ServerSettings
	// OISM возвращает ADDR
	if err := decodeXML(resp, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetOkpSettings (3.14)
func (d *mitsuDriver) GetOkpSettings() (*ServerSettings, error) {
	resp, err := d.sendCommand("<GET OKP='?'/>")
	if err != nil {
		return nil, err
	}
	var s ServerSettings
	// OKP возвращает атрибут OKP вместо ADDR (стр. 11: <OK OKP='IP/URL...' ...>)
	// Наша структура ServerSettings имеет тег OKP
	if err := decodeXML(resp, &s); err != nil {
		return nil, err
	}
	// Унификация: если заполнено поле Okp, переносим в Addr
	if s.Okp != "" && s.Addr == "" {
		s.Addr = s.Okp
	}
	return &s, nil
}

// GetTaxRates (3.15)
func (d *mitsuDriver) GetTaxRates() (*TaxRates, error) {
	resp, err := d.sendCommand("<GET TAX='?'/>")
	if err != nil {
		return nil, err
	}
	var t TaxRates
	if err := decodeXML(resp, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// GetRegistrationData (3.16)
func (d *mitsuDriver) GetRegistrationData() (*RegData, error) {
	resp, err := d.sendCommand("<GET REG='?'/>")
	if err != nil {
		return nil, err
	}
	var r RegData
	if err := decodeXML(resp, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// GetShiftStatus (3.17)
func (d *mitsuDriver) GetShiftStatus() (*ShiftStatus, error) {
	resp, err := d.sendCommand("<GET INFO='0'/>")
	if err != nil {
		return nil, err
	}
	var s ShiftStatus
	if err := decodeXML(resp, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetShiftTotals (3.18)
func (d *mitsuDriver) GetShiftTotals() (*ShiftTotals, error) {
	resp, err := d.sendCommand("<GET INFO='1'/>")
	if err != nil {
		return nil, err
	}
	var s ShiftTotals
	if err := decodeXML(resp, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetFnStatus (3.22)
func (d *mitsuDriver) GetFnStatus() (*FnStatus, error) {
	resp, err := d.sendCommand("<GET INFO='F'/>")
	if err != nil {
		return nil, err
	}
	var f FnStatus
	if err := decodeXML(resp, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// GetOfdExchangeStatus (3.23)
func (d *mitsuDriver) GetOfdExchangeStatus() (*OfdExchangeStatus, error) {
	resp, err := d.sendCommand("<GET INFO='O'/>")
	if err != nil {
		return nil, err
	}
	var s OfdExchangeStatus
	if err := decodeXML(resp, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetMarkingStatus (3.24)
func (d *mitsuDriver) GetMarkingStatus() (*MarkingStatus, error) {
	resp, err := d.sendCommand("<GET INFO='M'/>")
	if err != nil {
		return nil, err
	}
	var m MarkingStatus
	if err := decodeXML(resp, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// GetPowerStatus (3.33)
// Возвращает 1 (был сбой) или 0 (ок/сброшен).
func (d *mitsuDriver) GetPowerStatus() (int, error) {
	resp, err := d.sendCommand("<GET POWER='?'/>")
	if err != nil {
		return 0, err
	}
	var r struct {
		Val int `xml:"POWER,attr"`
	}
	if err := decodeXML(resp, &r); err != nil {
		return 0, err
	}
	return r.Val, nil
}

// GetTimezone (3.35)
func (d *mitsuDriver) GetTimezone() (int, error) {
	resp, err := d.sendCommand("<GET TIMEZONE='?'/>")
	if err != nil {
		return 0, err
	}
	var r struct {
		Tz string `xml:"TIMEZONE,attr"`
	}
	if err := decodeXML(resp, &r); err != nil {
		return 0, err
	}
	if r.Tz == "" {
		return 0, errors.New("часовая зона не возвращена")
	}
	var tz int
	_, err = fmt.Sscanf(r.Tz, "%d", &tz)
	return tz, err
}
