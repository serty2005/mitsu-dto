package gui

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"mitsuscanner/driver"
	"mitsuscanner/pkg/ofdclient"
	"time"
)

// OfdTransferResult содержит результат передачи документа
type OfdTransferResult struct {
	Success       bool
	DocumentsSent int
	ErrorMessage  string
}

// SendFirstUnsentDocument отправляет первый неотправленный документ в ОФД.
func SendFirstUnsentDocument(drv driver.Driver) (*OfdTransferResult, error) {
	result := &OfdTransferResult{}

	// 1. Проверяем статус обмена с ОФД
	ofdStatus, err := drv.GetOfdExchangeStatus()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения статуса ОФД: %w", err)
	}
	log.Printf("[OFD] Статус ОФД: %d неотправленных документов", ofdStatus.Count)

	if ofdStatus.Count == 0 {
		result.Success = true
		result.DocumentsSent = 0
		result.ErrorMessage = "Нет неотправленных документов"
		return result, nil
	}

	// 2. Настройки ОФД
	ofdSettings, err := drv.GetOfdSettings()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения настроек ОФД: %w", err)
	}
	log.Printf("[OFD] Настройки ОФД: адрес %s, порт %d", ofdSettings.Addr, ofdSettings.Port)

	// Гарантируем внешний клиент (восстановление настроек)
	originalClient := ofdSettings.Client
	needRestore := false
	if ofdSettings.Client != "1" {
		ofdSettings.Client = "1"
		if err := drv.SetOfdSettings(*ofdSettings); err != nil {
			return nil, fmt.Errorf("ошибка установки режима внешнего клиента: %w", err)
		}
		needRestore = true
	}
	defer func() {
		if needRestore {
			ofdSettings.Client = originalClient
			drv.SetOfdSettings(*ofdSettings)
		}
	}()

	// 3. Инфо о ФН
	fnStatus, err := drv.GetFnStatus()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения статуса ФН: %w", err)
	}
	fnSerial := fnStatus.Serial
	log.Printf("[OFD] ФН: %s", fnSerial)

	// 4. Читаем документ из ККТ
	docData, err := drv.OfdReadFullDocument()
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения документа: %w", err)
	}
	log.Printf("[OFD] Прочитано байт из ККТ: %d", len(docData))

	// Инициализация клиента
	client := ofdclient.New(ofdclient.Config{
		Timeout:       300 * time.Second,
		RetryCount:    3,
		RetryInterval: 5 * time.Second,
		Logger: func(msg string) {
			log.Printf("[OFDClient] %s", msg)
		},
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	var resp *ofdclient.SendResponse
	ofdAddr := fmt.Sprintf("%s:%d", ofdSettings.Addr, ofdSettings.Port)

	// 5. ПРОВЕРКА НА ГОТОВОЕ СООБЩЕНИЕ
	// Сигнатура ОФД: 2A 08 41 0A
	ofdSignature := []byte{0x2A, 0x08, 0x41, 0x0A}

	if bytes.HasPrefix(docData, ofdSignature) {
		log.Printf("[OFD] Обнаружена сигнатура готового сообщения ОФД. Отправляем RAW.")
		// ККТ вернула полный пакет (Заголовок + Контейнер), отправляем как есть
		resp, err = client.SendRaw(ctx, ofdAddr, docData)
	} else {
		log.Printf("[OFD] Сигнатура не найдена. Формируем сообщение вручную.")
		// ККТ вернула только TLV, нужно упаковать

		ffdVersion := resolveFFDVersion(fnStatus.Ffd)

		// Создаем заголовок контейнера (Тип A5, Версия 1 -> 32 байта)
		contHeader := ofdclient.CreateContainerHeader(0xA5, 0, 1)
		containerData, err := ofdclient.SerializeContainer(contHeader, docData)
		if err != nil {
			return nil, fmt.Errorf("ошибка упаковки контейнера: %w", err)
		}

		req := ofdclient.SendRequest{
			OfdAddress: ofdAddr,
			FnNumber:   fnSerial,
			FFDVersion: ffdVersion,
			Container:  containerData,
		}
		resp, err = client.Send(ctx, req)
	}

	if err != nil {
		log.Printf("[OFD] Ошибка отправки: %v", err)
		return nil, fmt.Errorf("ошибка отправки в ОФД: %w", err)
	}
	log.Printf("[OFD] Успешная отправка. Получен ответ длиной %d байт", len(resp.RawMessage))

	// 6. Записываем квитанцию в ФН
	// ФН требует ПОЛНОЕ сообщение (RawMessage) включая заголовок 30 байт,
	// а не только тело квитанции (Receipt).
	if err := drv.OfdLoadReceipt(resp.RawMessage); err != nil {
		return nil, fmt.Errorf("ошибка записи квитанции: %w", err)
	}
	log.Printf("[OFD] Квитанция записана в ФН (размер: %d байт)", len(resp.RawMessage))

	result.Success = true
	result.DocumentsSent = 1
	result.ErrorMessage = fmt.Sprintf(
		"Документ успешно отправлен в ОФД.\nФН: %s\nКвитанция: %d байт",
		fnSerial, len(resp.Receipt))

	// Обновляем информацию о неотправленных
	sh, err := drv.GetShiftStatus()
	if err == nil {
		// Обновление счётчика неотправленных документов в панели статуса
		if unsentDocsLabel != nil {
			unsentDocsLabel.SetText(fmt.Sprintf("ОФД: %d", sh.Ofd.Count))
		}
	}

	return result, nil
}

// resolveFFDVersion приводит код версии из ККТ к стандарту 1.0/1.05/1.1/1.2
func resolveFFDVersion(code string) string {
	switch code {
	case "4", "1.2", "1.20":
		return "1.2"
	case "3", "1.1", "1.10":
		return "1.1"
	case "2", "1.05":
		return "1.05"
	default:
		return "1.0"
	}
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
