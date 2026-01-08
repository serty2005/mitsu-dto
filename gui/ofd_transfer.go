package gui

import (
	"fmt"
	"mitsuscanner/driver"
)

// OfdTransferResult содержит результат передачи документа
type OfdTransferResult struct {
	Success       bool
	DocumentsSent int
	ErrorMessage  string
}

// SendFirstUnsentDocument отправляет первый неотправленный документ в ОФД.
// Это заглушка - в будущем здесь будет реальный клиент ОФД.
func SendFirstUnsentDocument(drv driver.Driver) (*OfdTransferResult, error) {
	result := &OfdTransferResult{}

	// 1. Проверяем статус обмена с ОФД
	ofdStatus, err := drv.GetOfdExchangeStatus()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения статуса ОФД: %w", err)
	}

	if ofdStatus.Count == 0 {
		result.Success = true
		result.DocumentsSent = 0
		result.ErrorMessage = "Нет неотправленных документов"
		return result, nil
	}

	// 2. Проверяем/устанавливаем режим внешнего клиента
	ofdSettings, err := drv.GetOfdSettings()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения настроек ОФД: %w", err)
	}

	originalClient := ofdSettings.Client
	needRestore := false

	if ofdSettings.Client != "1" {
		// Переключаем на внешний клиент
		ofdSettings.Client = "1"
		if err := drv.SetOfdSettings(*ofdSettings); err != nil {
			return nil, fmt.Errorf("ошибка установки режима внешнего клиента: %w", err)
		}
		needRestore = true
	}

	// Гарантируем восстановление настроек
	defer func() {
		if needRestore {
			ofdSettings.Client = originalClient
			drv.SetOfdSettings(*ofdSettings)
		}
	}()

	// 3. Получаем номер ФН для клиента ОФД
	fnStatus, err := drv.GetFnStatus()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения статуса ФН: %w", err)
	}
	fnSerial := fnStatus.Serial

	// 4. Читаем документ из ККТ
	docData, err := drv.OfdReadFullDocument()
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения документа: %w", err)
	}

	// 5. ЗАГЛУШКА: Отправка в ОФД
	// TODO: Интеграция с реальным клиентом ОФД
	// receipt, err := ofdClient.SendDocument(fnSerial, docData)
	// if err != nil {
	//     return nil, fmt.Errorf("ошибка отправки в ОФД: %w", err)
	// }

	// Пока используем заглушку
	_ = fnSerial
	_ = docData

	result.Success = false
	result.ErrorMessage = fmt.Sprintf(
		"Документ прочитан (%d байт). Клиент ОФД ещё не реализован.\n"+
			"ФН: %s\nНеотправленных: %d",
		len(docData), fnSerial, ofdStatus.Count)

	// 6. Записываем квитанцию в ФН (когда будет реальный клиент)
	// if err := drv.OfdLoadReceipt(receipt); err != nil {
	//     return nil, fmt.Errorf("ошибка записи квитанции: %w", err)
	// }

	// result.Success = true
	// result.DocumentsSent = 1

	return result, nil
}

// RefreshFnInfo обновляет информацию о ФН в модели регистрации
func RefreshFnInfo(drv driver.Driver) error {
	if drv == nil {
		return fmt.Errorf("нет подключения к ККТ")
	}

	fnStatus, err := drv.GetFnStatus()
	if err != nil {
		return fmt.Errorf("ошибка чтения статуса ФН: %w", err)
	}

	// Обновляем модель
	regModel.FnNumber = fnStatus.Serial
	regModel.FnValidDate = fnStatus.Valid
	regModel.FnPhase = fnStatus.Phase

	phaseText, phaseColor := decodeFnPhase(fnStatus.Phase)
	regModel.FnPhaseText = phaseText
	regModel.FnPhaseColor = phaseColor

	// Обновляем UI
	if regBinder != nil {
		regBinder.Reset()
	}
	if fnPhaseLabel != nil {
		fnPhaseLabel.SetTextColor(phaseColor)
	}

	return nil
}
