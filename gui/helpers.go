package gui

import (
	"mitsuscanner/driver"
)

// =============================================================================
// ХЕЛПЕРЫ ДЛЯ СБОРКИ ДАННЫХ ИЗ НЕСКОЛЬКИХ АТОМАРНЫХ ОПЕРАЦИЙ ДРАЙВЕРА
// =============================================================================

// GetFullRegistrationData собирает полные данные регистрации для отображения в GUI.
// Выполняет последовательность команд:
//   - GET REG='?' (основные данные регистрации)
//   - GET INFO='F' (данные ФН: серийный номер, исполнение)
//   - GET VER='?' (серийный номер принтера)
//
// Используется в: onReadRegistration, onRegister, onReregister
func GetFullRegistrationData(drv driver.Driver) (*driver.RegData, error) {
	// 1. Базовые данные регистрации
	regData, err := drv.GetRegistrationData()
	if err != nil {
		return nil, err
	}

	// 2. Данные ФН (серийный номер, исполнение)
	// Ошибки игнорируем — поля останутся пустыми
	if fnStatus, err := drv.GetFnStatus(); err == nil {
		regData.FnSerial = fnStatus.Serial
		regData.FnEdition = fnStatus.Edition
	}

	// 3. Серийный номер принтера
	// Ошибки игнорируем — поле останется пустым
	if _, serial, _, err := drv.GetVersion(); err == nil {
		regData.PrinterSerial = serial
	}

	return regData, nil
}

// GetCloseFnReportData собирает данные для отчёта о закрытии фискального накопителя.
// Выполняет последовательность команд:
//   - GET REG='?' (RNM, адрес, место)
//   - GET INFO='F' (серийный номер ФН)
//   - GET VER='?' (серийный номер принтера)
//   - GET DOC='X:fd' + READ (XML документа для извлечения даты-времени)
//
// Параметры:
//   - drv: активный драйвер
//   - fd: номер фискального документа (из результата CloseFiscalArchive)
//   - fp: фискальный признак (из результата CloseFiscalArchive)
//
// Используется в: onCloseFn
func GetCloseFnReportData(drv driver.Driver, fd int, fp string) (*driver.ReportFnCloseData, error) {
	// 1. Базовые данные регистрации (RNM, адрес, место)
	regData, err := drv.GetRegistrationData()
	if err != nil {
		return nil, err
	}

	// 2. Серийный номер ФН
	fnSerial := ""
	if fnStatus, err := drv.GetFnStatus(); err == nil {
		fnSerial = fnStatus.Serial
	}

	// 3. Серийный номер принтера (ЗН ККТ)
	printerSerial := ""
	if _, serial, _, err := drv.GetVersion(); err == nil {
		printerSerial = serial
	}

	// 4. Дата-время из XML документа в ФН
	dateTimeStr := "—"
	if xmlDoc, err := drv.GetDocumentXMLFromFN(fd); err == nil {
		if dt, err := driver.ExtractDocDateTime(xmlDoc); err == nil {
			dateTimeStr = dt
		}
	}

	return &driver.ReportFnCloseData{
		FD:        fd,
		FP:        fp,
		DateTime:  dateTimeStr,
		RNM:       regData.RNM,
		FNNumber:  fnSerial,
		KKTNumber: printerSerial,
		Address:   regData.Address,
		Place:     regData.Place,
	}, nil
}
