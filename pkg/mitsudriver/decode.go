package mitsudriver

import (
	"fmt"
	"regexp"
	"time"
)

// DecodeMode декодирует битовую маску MODE в список установленных битов.
func DecodeMode(mask uint32) []string {
	var res []string
	for i := 0; i < 32; i++ {
		if mask&(1<<i) != 0 {
			res = append(res, fmt.Sprintf("MODE бит %d", i))
		}
	}
	return res
}

// DecodeExtMode декодирует битовую маску ExtMODE в список установленных битов.
func DecodeExtMode(mask uint32) []string {
	var res []string
	for i := 0; i < 32; i++ {
		if mask&(1<<i) != 0 {
			res = append(res, fmt.Sprintf("ExtMODE бит %d", i))
		}
	}
	return res
}

// GetReportMeta возвращает метаданные отчета по типу документа.
func GetReportMeta(typeCode int) ReportMeta {
	switch typeCode {
	case 1: // Отчёт о регистрации
		return ReportMeta{
			Kind:  ReportReg,
			Title: "Отчет о регистрации",
		}
	case 11: // Отчет о (пере) регистрации
		return ReportMeta{
			Kind:  ReportRereg,
			Title: "Отчет о перерегистрации",
		}
	case 6: // Отчёт о закрытии ФН
		return ReportMeta{
			Kind:  ReportKindCloseFn,
			Title: "Отчет о закрытии фискального архива",
		}
	default:
		return ReportMeta{
			Kind:  "",
			Title: "Неизвестный отчет",
		}
	}
}

// ExtractDocDateTime парсит XML строку, находит содержимое тега <T1012> и возвращает дату-время в формате "02.01.2006 15:04".
// Поддерживает layout'ы: "02-01-06T15:04", "02-01-06T15:04:05", "2006-01-02T15:04", "2006-01-02T15:04:05".
// Если тег отсутствует или парсинг не удался, возвращает ошибку.
func ExtractDocDateTime(xmlStr string) (string, error) {
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
