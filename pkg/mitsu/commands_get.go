package mitsu

import (
	"context"
	"encoding/xml"
	"fmt"
)

// GetDev запрашивает модель ККТ (См. п. 3.3)
func (c *mitsuClient) GetDev(ctx context.Context) (*DevResponse, error) {
	resp, err := c.SendCommand(ctx, "<GET DEV='?'/>")
	if err != nil {
		return nil, err
	}

	var r DevResponse
	if err := xml.Unmarshal(resp, &r); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа DEV: %w", err)
	}

	return &r, nil
}

// GetVer запрашивает версию ПО, серийный номер и MAC-адрес (См. п. 3.4)
func (c *mitsuClient) GetVer(ctx context.Context) (*VerResponse, error) {
	resp, err := c.SendCommand(ctx, "<GET VER='?'/>")
	if err != nil {
		return nil, err
	}

	var r VerResponse
	if err := xml.Unmarshal(resp, &r); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа VER: %w", err)
	}

	return &r, nil
}

// GetDateTime запрашивает текущую дату и время ККТ (См. п. 3.5)
func (c *mitsuClient) GetDateTime(ctx context.Context) (*DateTimeResponse, error) {
	resp, err := c.SendCommand(ctx, "<GET DATE='?' TIME='?'/>")
	if err != nil {
		return nil, err
	}

	var r DateTimeResponse
	if err := xml.Unmarshal(resp, &r); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа DATE/TIME: %w", err)
	}

	return &r, nil
}

// GetCashier запрашивает данные кассира (См. п. 3.6)
func (c *mitsuClient) GetCashier(ctx context.Context) (*CashierResponse, error) {
	resp, err := c.SendCommand(ctx, "<GET CASHIER='?'/>")
	if err != nil {
		return nil, err
	}

	var r CashierResponse
	if err := xml.Unmarshal(resp, &r); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа CASHIER: %w", err)
	}

	return &r, nil
}

// GetPrinterSettings запрашивает настройки принтера (См. п. 3.7)
func (c *mitsuClient) GetPrinterSettings(ctx context.Context) (*PrinterSettings, error) {
	resp, err := c.SendCommand(ctx, "<GET PRINTER='?'/>")
	if err != nil {
		return nil, err
	}

	var s PrinterSettings
	if err := xml.Unmarshal(resp, &s); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа PRINTER: %w", err)
	}

	return &s, nil
}

// GetMoneyDrawerSettings запрашивает настройки денежного ящика (См. п. 3.8)
func (c *mitsuClient) GetMoneyDrawerSettings(ctx context.Context) (*DrawerSettings, error) {
	resp, err := c.SendCommand(ctx, "<GET CD='?'/>")
	if err != nil {
		return nil, err
	}

	var s DrawerSettings
	if err := xml.Unmarshal(resp, &s); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа CD: %w", err)
	}

	return &s, nil
}

// GetComSettings запрашивает настройки COM-порта (См. п. 3.9)
func (c *mitsuClient) GetComSettings(ctx context.Context) (*ComSettingsResponse, error) {
	resp, err := c.SendCommand(ctx, "<GET COM='?'/>")
	if err != nil {
		return nil, err
	}

	var r ComSettingsResponse
	if err := xml.Unmarshal(resp, &r); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа COM: %w", err)
	}

	return &r, nil
}

// GetHeader запрашивает заголовок (См. п. 3.10)
func (c *mitsuClient) GetHeader(ctx context.Context, n int) (*HeaderResponse, error) {
	cmd := fmt.Sprintf("<GET HEADER='%d'/>", n)
	resp, err := c.SendCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var r HeaderResponse
	if err := xml.Unmarshal(resp, &r); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа HEADER: %w", err)
	}

	return &r, nil
}

// GetLanSettings запрашивает настройки LAN (См. п. 3.11)
func (c *mitsuClient) GetLanSettings(ctx context.Context) (*LanSettings, error) {
	resp, err := c.SendCommand(ctx, "<GET LAN='?'/>")
	if err != nil {
		return nil, err
	}

	var s LanSettings
	if err := xml.Unmarshal(resp, &s); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа LAN: %w", err)
	}

	return &s, nil
}

// GetOfdSettings запрашивает настройки ОФД (См. п. 3.12)
func (c *mitsuClient) GetOfdSettings(ctx context.Context) (*OfdSettings, error) {
	resp, err := c.SendCommand(ctx, "<GET OFD='?'/>")
	if err != nil {
		return nil, err
	}

	var s OfdSettings
	if err := xml.Unmarshal(resp, &s); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа OFD: %w", err)
	}

	return &s, nil
}

// GetOismSettings запрашивает настройки OISM (См. п. 3.13)
func (c *mitsuClient) GetOismSettings(ctx context.Context) (*OismSettings, error) {
	resp, err := c.SendCommand(ctx, "<GET OISM='?'/>")
	if err != nil {
		return nil, err
	}

	var s OismSettings
	if err := xml.Unmarshal(resp, &s); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа OISM: %w", err)
	}

	return &s, nil
}

// GetOkpSettings запрашивает настройки OKP (См. п. 3.14)
func (c *mitsuClient) GetOkpSettings(ctx context.Context) (*ServerSettings, error) {
	resp, err := c.SendCommand(ctx, "<GET OKP='?'/>")
	if err != nil {
		return nil, err
	}

	var s ServerSettings
	if err := xml.Unmarshal(resp, &s); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа OKP: %w", err)
	}

	// Унификация: если заполнено поле Okp, переносим в Addr
	if s.Okp != "" && s.Addr == "" {
		s.Addr = s.Okp
	}

	return &s, nil
}

// GetTaxRates запрашивает налоговые ставки (См. п. 3.15)
func (c *mitsuClient) GetTaxRates(ctx context.Context) (*TaxRates, error) {
	resp, err := c.SendCommand(ctx, "<GET TAX='?'/>")
	if err != nil {
		return nil, err
	}

	var t TaxRates
	if err := xml.Unmarshal(resp, &t); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа TAX: %w", err)
	}

	return &t, nil
}

// GetRegistrationData запрашивает данные о регистрации ККТ (См. п. 3.16)
func (c *mitsuClient) GetRegistrationData(ctx context.Context) (*RegData, error) {
	resp, err := c.SendCommand(ctx, "<GET REG='?'/>")
	if err != nil {
		return nil, err
	}

	var r RegData
	if err := xml.Unmarshal(resp, &r); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа REG: %w", err)
	}

	return &r, nil
}

// GetStatus получает статус ККТ (INFO='0').
func (c *mitsuClient) GetStatus(ctx context.Context) (*ShiftStatus, error) {
	resp, err := c.SendCommand(ctx, "<GET INFO='0'/>")
	if err != nil {
		return nil, err
	}

	var s ShiftStatus
	if err := xml.Unmarshal(resp, &s); err != nil {
		return nil, fmt.Errorf("ошибка разбора статуса ККТ: %w", err)
	}

	return &s, nil
}

// GetShiftTotals получает итоги смены (INFO='1')
func (c *mitsuClient) GetShiftTotals(ctx context.Context) (*ShiftTotals, error) {
	resp, err := c.SendCommand(ctx, "<GET INFO='1'/>")
	if err != nil {
		return nil, err
	}

	var s ShiftTotals
	if err := xml.Unmarshal(resp, &s); err != nil {
		return nil, fmt.Errorf("ошибка разбора итогов смены: %w", err)
	}

	return &s, nil
}

// GetFnStatus получает статус фискального накопителя (INFO='F').
func (c *mitsuClient) GetFnStatus(ctx context.Context) (*FnStatus, error) {
	resp, err := c.SendCommand(ctx, "<GET INFO='F'/>")
	if err != nil {
		return nil, err
	}

	var f FnStatus
	if err := xml.Unmarshal(resp, &f); err != nil {
		return nil, fmt.Errorf("ошибка разбора статуса ФН: %w", err)
	}

	return &f, nil
}

// GetOfdExchangeStatus получает статус обмена с ОФД (INFO='O')
func (c *mitsuClient) GetOfdExchangeStatus(ctx context.Context) (*OfdExchangeStatus, error) {
	resp, err := c.SendCommand(ctx, "<GET INFO='O'/>")
	if err != nil {
		return nil, err
	}

	var s OfdExchangeStatus
	if err := xml.Unmarshal(resp, &s); err != nil {
		return nil, fmt.Errorf("ошибка разбора статуса обмена с ОФД: %w", err)
	}

	return &s, nil
}

// GetMarkingStatus получает статус маркировки (INFO='M')
func (c *mitsuClient) GetMarkingStatus(ctx context.Context) (*MarkingStatus, error) {
	resp, err := c.SendCommand(ctx, "<GET INFO='M'/>")
	if err != nil {
		return nil, err
	}

	var m MarkingStatus
	if err := xml.Unmarshal(resp, &m); err != nil {
		return nil, fmt.Errorf("ошибка разбора статуса маркировки: %w", err)
	}

	return &m, nil
}

// GetPowerStatus запрашивает статус питания (См. п. 3.33)
func (c *mitsuClient) GetPowerStatus(ctx context.Context) (*PowerStatusResponse, error) {
	resp, err := c.SendCommand(ctx, "<GET POWER='?'/>")
	if err != nil {
		return nil, err
	}

	var r PowerStatusResponse
	if err := xml.Unmarshal(resp, &r); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа POWER: %w", err)
	}

	return &r, nil
}

// GetTimezone запрашивает часовой пояс (См. п. 3.35)
func (c *mitsuClient) GetTimezone(ctx context.Context) (*TimezoneResponse, error) {
	resp, err := c.SendCommand(ctx, "<GET TIMEZONE='?'/>")
	if err != nil {
		return nil, err
	}

	var r TimezoneResponse
	if err := xml.Unmarshal(resp, &r); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа TIMEZONE: %w", err)
	}

	return &r, nil
}

// GetOptions запрашивает все опции устройства (См. п. 4.13)
func (c *mitsuClient) GetOptions(ctx context.Context) (*DeviceOptions, error) {
	resp, err := c.SendCommand(ctx, "<OPTION/>")
	if err != nil {
		return nil, err
	}

	var opts DeviceOptions
	if err := xml.Unmarshal(resp, &opts); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа OPTION: %w", err)
	}

	return &opts, nil
}

// GetCurrentDocumentType получает тип текущего документа
func (c *mitsuClient) GetCurrentDocumentType(ctx context.Context) (*CurrentDocumentTypeResponse, error) {
	resp, err := c.SendCommand(ctx, "<GET DOC='0'/>")
	if err != nil {
		return nil, err
	}

	var r CurrentDocumentTypeResponse
	if err := xml.Unmarshal(resp, &r); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа DOC='0': %w", err)
	}

	return &r, nil
}

// GetDocumentInfoFromFN получает информацию о документе (OFFSET и LENGTH) по номеру FD
func (c *mitsuClient) GetDocumentInfoFromFN(ctx context.Context, fd int) (*DocumentInfoResponse, error) {
	cmd := fmt.Sprintf("<GET DOC='X:%d'/>", fd)
	resp, err := c.SendCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var r DocumentInfoResponse
	if err := xml.Unmarshal(resp, &r); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа DOC='X:%d': %w", fd, err)
	}

	return &r, nil
}

// ReadBlock читает блок данных из памяти ФН по OFFSET и LENGTH
func (c *mitsuClient) ReadBlock(ctx context.Context, offset int64, length int) (*ReadBlockResponse, error) {
	cmd := fmt.Sprintf("<READ OFFSET='%X' LENGTH='%d'/>", offset, length)
	resp, err := c.SendCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var r ReadBlockResponse
	if err := xml.Unmarshal(resp, &r); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа READ: %w", err)
	}

	return &r, nil
}
