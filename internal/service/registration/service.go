package registration

import (
	"bytes"
	"fmt"
	"regexp"
	"time"

	"mitsuscanner/internal/domain/models"
	"mitsuscanner/internal/domain/ports"
)

// RegistrationService реализует бизнес-логику регистрации ККТ.
type RegistrationService struct {
	driver ports.Driver
}

// NewRegistrationService создает новый экземпляр RegistrationService.
func NewRegistrationService(driver ports.Driver) *RegistrationService {
	return &RegistrationService{
		driver: driver,
	}
}

// GetFullRegistrationData собирает полные данные регистрации для отображения.
// Выполняет последовательность команд:
//   - GET REG='?' (основные данные регистрации)
//   - GET INFO='F' (данные ФН: серийный номер, исполнение)
//   - GET VER='?' (серийный номер принтера)
func (s *RegistrationService) GetFullRegistrationData() (*models.RegData, error) {
	// 1. Базовые данные регистрации
	regData, err := s.driver.GetRegistrationData()
	if err != nil {
		return nil, err
	}

	// 2. Данные ФН (серийный номер, исполнение) — ошибки игнорируем
	if fnStatus, err := s.driver.GetFnStatus(); err == nil {
		regData.FnSerial = fnStatus.Serial
		regData.FnEdition = fnStatus.Edition
	}

	// 3. Серийный номер принтера — ошибки игнорируем
	if _, serial, _, err := s.driver.GetVersion(); err == nil {
		regData.PrinterSerial = serial
	}

	return regData, nil
}

// Register выполняет регистрацию ККТ.
func (s *RegistrationService) Register(req models.RegistrationRequest) (*models.RegResponse, error) {
	return s.driver.Register(req)
}

// Reregister выполняет перерегистрацию ККТ.
func (s *RegistrationService) Reregister(req models.RegistrationRequest, reasons []int) (*models.RegResponse, error) {
	return s.driver.Reregister(req, reasons)
}

// CloseFiscalArchive закрывает фискальный архив и возвращает данные для отчета.
func (s *RegistrationService) CloseFiscalArchive() (*models.ReportFnCloseData, error) {
	// 1. Закрываем ФН
	result, err := s.driver.CloseFiscalArchive()
	if err != nil {
		return nil, err
	}

	// 2. Сбор данных для отчета
	reportData, err := s.getCloseFnReportData(result.FD, result.FP)
	if err != nil {
		return nil, err
	}

	return reportData, nil
}

// getCloseFnReportData собирает данные для отчета о закрытии фискального архива.
func (s *RegistrationService) getCloseFnReportData(fd int, fp string) (*models.ReportFnCloseData, error) {
	// 1. Базовые данные регистрации (RNM, адрес, место)
	regData, err := s.driver.GetRegistrationData()
	if err != nil {
		return nil, err
	}

	// 2. Серийный номер ФН
	fnSerial := ""
	if fnStatus, err := s.driver.GetFnStatus(); err == nil {
		fnSerial = fnStatus.Serial
	}

	// 3. Серийный номер принтера (ЗН ККТ)
	printerSerial := ""
	if _, serial, _, err := s.driver.GetVersion(); err == nil {
		printerSerial = serial
	}

	// 4. Дата-время из XML документа в ФН
	dateTimeStr := "—"
	if xmlDoc, err := s.driver.GetDocumentXMLFromFN(fd); err == nil {
		if dt, err := extractDocDateTime(xmlDoc); err == nil {
			dateTimeStr = dt
		}
	}

	return &models.ReportFnCloseData{
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

// extractDocDateTime парсит XML строку, находит содержимое тега <T1012> и возвращает дату-время в формате "02.01.2006 15:04".
// Поддерживает layout'ы: "02-01-06T15:04", "02-01-06T15:04:05", "2006-01-02T15:04", "2006-01-02T15:04:05".
func extractDocDateTime(xmlStr string) (string, error) {
	re := regexp.MustCompile(`<T1012>([^<]*)</T1012>`)
	matches := re.FindStringSubmatch(xmlStr)
	if len(matches) < 2 {
		return "", fmt.Errorf("тег T1012 не найден")
	}
	dateStr := matches[1]
	layouts := []string{"02-01-06T15:04", "02-01-06T15:04:05", "2006-01-02T15:04", "2006-01-02T15:04:05"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t.Format("02.01.2006 15:04"), nil
		}
	}
	return "", fmt.Errorf("не удалось распарсить дату-время: %s", dateStr)
}

// CalculateRNM выполняет расчет РНМ по алгоритму CRC16-CCITT.
// Формат входных данных для CRC:
// Pad(Order, 10) + Pad(INN, 12) + Pad(Serial, 20)
// Результат: Pad(Order, 10) + Pad(CRC, 6)
func (s *RegistrationService) CalculateRNM(orderNum, inn, serial string) (string, error) {
	// 1. Формируем строку для расчета
	paddedOrder := padLeft(orderNum, 10, '0')
	paddedInn := padLeft(inn, 12, '0')
	paddedSerial := padLeft(serial, 20, '0')

	calcString := paddedOrder + paddedInn + paddedSerial

	// 2. Считаем CRC
	crc := crc16ccitt([]byte(calcString))

	// 3. Формируем хвост (CRC дополненный до 6 цифр нулями)
	crcStr := fmt.Sprintf("%d", crc)
	paddedCrc := padLeft(crcStr, 6, '0')

	// 4. Итоговый РНМ
	finalRnm := paddedOrder + paddedCrc

	return finalRnm, nil
}

// crc16ccitt вычисляет CRC-16 (CCITT False)
// Poly: 0x1021, Init: 0xFFFF
func crc16ccitt(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if (crc & 0x8000) != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// padLeft дополняет строку символом padChar слева до длины length
func padLeft(s string, length int, padChar byte) string {
	if len(s) >= length {
		return s
	}
	padding := bytes.Repeat([]byte{padChar}, length-len(s))
	return string(padding) + s
}
