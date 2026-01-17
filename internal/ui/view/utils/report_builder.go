package utils

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"mitsuscanner/internal/domain/models"
)

// kv представляет пару ключ-значение для построения отчета
type kv struct {
	Key   string
	Value string
}

// BuildReportLines собирает данные для отчета в зависимости от типа
func BuildReportLines(data interface{}, kind models.ReportKind) ([]kv, error) {
	switch kind {
	case models.ReportKindRegistration:
		var rd models.RegData
		if d, ok := data.(models.RegData); ok {
			rd = d
		} else if d, ok := data.(*models.RegData); ok {
			rd = *d
		} else {
			return nil, fmt.Errorf("invalid report data type: %T", data)
		}
		return buildRegistrationLines(rd), nil
	case models.ReportKindCloseFn:
		var cd models.ReportFnCloseData
		if d, ok := data.(models.ReportFnCloseData); ok {
			cd = d
		} else if d, ok := data.(*models.ReportFnCloseData); ok {
			cd = *d
		} else {
			return nil, fmt.Errorf("invalid report data type: %T", data)
		}
		return buildCloseFnLines(cd), nil
	default:
		return nil, fmt.Errorf("unknown report kind: %v", kind)
	}
}

func buildRegistrationLines(regData models.RegData) []kv {
	var lines []kv

	lines = append(lines, kv{"ЗН ФР", regData.PrinterSerial})
	lines = append(lines, kv{"Номер ФД", regData.FdNumber})
	lines = append(lines, kv{"ФП", regData.FpNumber})
	lines = append(lines, kv{"Дата и время регистрации", formatDateTime(regData.RegDate, regData.RegTime)})
	lines = append(lines, kv{"Номер ФН", regData.FnSerial})
	lines = append(lines, kv{"Исполнение ФН", regData.FnEdition})
	lines = append(lines, kv{"РНМ", formatRNM(regData.RNM)})
	lines = append(lines, kv{"ИНН организации", regData.Inn})
	lines = append(lines, kv{"Наименование организации", regData.OrgName})
	lines = append(lines, kv{"Адрес расчетов", regData.Address})
	lines = append(lines, kv{"Место расчетов", regData.Place})

	lines = append(lines, kv{"Версия ФФД", formatFFD(regData.FfdVer)})
	lines = append(lines, kv{"Системы налогообложения (СНО)", formatTaxSystems(regData.TaxSystems)})
	lines = append(lines, kv{"Базовая система налогообложения", regData.TaxBase})

	lines = append(lines, kv{"ИНН ОФД", regData.OfdInn})
	lines = append(lines, kv{"Наименование ОФД", regData.OfdName})
	lines = append(lines, kv{"Коды причин регистрации", regData.Base})

	// Флаги — добавляем только если есть
	appendIfNotEmpty(&lines, "Маркировка товаров", regData.MarkAttr)
	appendIfNotEmpty(&lines, "Подакцизные товары", regData.ExciseAttr)
	appendIfNotEmpty(&lines, "Расчеты в сети Интернет", regData.InternetAttr)
	appendIfNotEmpty(&lines, "Услуги", regData.ServiceAttr)
	appendIfNotEmpty(&lines, "БСО", regData.BsoAttr)
	appendIfNotEmpty(&lines, "Лотерея", regData.LotteryAttr)
	appendIfNotEmpty(&lines, "Азартные игры", regData.GamblingAttr)
	appendIfNotEmpty(&lines, "Ломбард", regData.PawnAttr)
	appendIfNotEmpty(&lines, "Страхование", regData.InsAttr)
	appendIfNotEmpty(&lines, "Общепит", regData.DineAttr)
	appendIfNotEmpty(&lines, "Оптовая торговля", regData.OptAttr)
	appendIfNotEmpty(&lines, "Вендинг", regData.VendAttr)
	appendIfNotEmpty(&lines, "Автоматический режим", regData.AutoModeAttr)
	appendIfNotEmpty(&lines, "Номер автомата", regData.AutoNumAttr)

	return lines
}

func buildCloseFnLines(closeData models.ReportFnCloseData) []kv {
	var lines []kv

	lines = append(lines, kv{"Дата и время закрытия", closeData.DateTime})
	lines = append(lines, kv{"ФП", closeData.FP})
	lines = append(lines, kv{"Номер ФД", fmt.Sprintf("%d", closeData.FD)})
	lines = append(lines, kv{"РНМ", formatRNM(closeData.RNM)})
	lines = append(lines, kv{"Номер ФН", closeData.FNNumber})
	lines = append(lines, kv{"ЗН ККТ", closeData.KKTNumber})
	lines = append(lines, kv{"Адрес расчетов", closeData.Address})
	lines = append(lines, kv{"Место расчетов", closeData.Place})

	return lines
}

func appendIfNotEmpty(lines *[]kv, key, value string) {
	if value != "" {
		*lines = append(*lines, kv{key, value})
	}
}

// FormatKeyValueText форматирует данные отчета в текстовый вид с выравниванием
func FormatKeyValueText(lines []kv) string {
	maxKeyLen := 0
	for _, l := range lines {
		keyLen := utf8.RuneCountInString(l.Key)
		if keyLen > maxKeyLen {
			maxKeyLen = keyLen
		}
	}

	var b strings.Builder

	for _, l := range lines {
		keyPad := maxKeyLen - utf8.RuneCountInString(l.Key)
		prefix := l.Key + strings.Repeat(" ", keyPad) + ": "

		value := normalizeNewlines(l.Value)
		valueLines := strings.Split(value, "\n")

		// Первая строка
		b.WriteString(prefix)
		b.WriteString(valueLines[0])
		b.WriteString("\n")

		// Остальные строки (если есть)
		for i := 1; i < len(valueLines); i++ {
			b.WriteString(strings.Repeat(" ", utf8.RuneCountInString(prefix)))
			b.WriteString(valueLines[i])
			b.WriteString("\n")
		}
	}

	return b.String()
}

func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

// ToWindowsText конвертирует новые строки в Windows-формат
func ToWindowsText(s string) string {
	return strings.ReplaceAll(s, "\n", "\r\n")
}

// formatRNM форматирует РНМ с пробелами по 4 цифры
func formatRNM(rnm string) string {
	if len(rnm) != 16 {
		return rnm
	}
	return fmt.Sprintf("%s %s %s %s", rnm[0:4], rnm[4:8], rnm[8:12], rnm[12:16])
}

// formatDateTime форматирует дату и время как дд.мм.гггг чч:мм
func formatDateTime(date, time string) string {
	if len(date) != 10 || len(time) < 5 {
		return date + " " + time
	}
	// date: гггг-мм-дд
	// time: чч:мм:сс
	year := date[0:4]
	month := date[5:7]
	day := date[8:10]
	hour := time[0:2]
	min := time[3:5]
	return fmt.Sprintf("%s.%s.%s %s:%s", day, month, year, hour, min)
}

// formatFFD заменяет 4 на 1.2
func formatFFD(ffd string) string {
	if ffd == "4" {
		return "1.2"
	}
	return ffd
}

// formatTaxSystems расшифровывает системы налогообложения
func formatTaxSystems(tax string) string {
	taxMap := map[string]string{
		"0": "ОСН",
		"1": "УСН доход",
		"2": "УСН доход-расход",
		"3": "ЕНВД",
		"4": "ЕСХН",
		"5": "Патент",
	}
	parts := strings.Split(tax, ",")
	var names []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if name, ok := taxMap[p]; ok {
			names = append(names, name)
		} else {
			names = append(names, p)
		}
	}
	return strings.Join(names, ", ")
}

// GenerateReportFileName генерирует имя файла для отчета вида "РНМ_YYYYMMDD.txt"
func GenerateReportFileName(data interface{}) string {
	rnm := ""

	switch d := data.(type) {
	case models.RegData:
		rnm = d.RNM
	case *models.RegData:
		rnm = d.RNM
	case models.ReportFnCloseData:
		rnm = d.RNM
	case *models.ReportFnCloseData:
		rnm = d.RNM
	}

	// Очистка РНМ от пробелов
	rnm = strings.ReplaceAll(rnm, " ", "")
	if rnm == "" {
		rnm = "Report"
	}

	dateStr := time.Now().Format("20060102") // YYYYMMDD
	return fmt.Sprintf("%s_%s.txt", rnm, dateStr)
}
