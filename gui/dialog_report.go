package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"mitsuscanner/driver"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
)

/*
   ===============================
   REPORT MODAL - МОДАЛКА ДЛЯ ОТЧЕТОВ
   ===============================

   Архитектура:
   1. buildReportLines   — сбор данных отчета (БЕЗ форматирования)
   2. formatKeyValueText — выравнивание и формирование plain text
   3. buildFormattedText — тонкая склейка
   4. RunReportModal     — UI

   ВАЖНО:
   Итог — обычная строка (plain text).
   Она используется ОДИНАКОВО для:
   - отображения
   - копирования
   - сохранения
*/

// --------------------
// Модель строки
// --------------------

type kv struct {
	Key   string
	Value string
}

// --------------------
// 1. СБОР ДАННЫХ
// --------------------

func buildReportLines(data interface{}, kind driver.ReportKind) ([]kv, error) {
	switch kind {
	case driver.ReportKindRegistration:
		var rd driver.RegData
		if d, ok := data.(driver.RegData); ok {
			rd = d
		} else if d, ok := data.(*driver.RegData); ok {
			rd = *d
		} else {
			return nil, fmt.Errorf("invalid report data type: %T", data)
		}
		return buildRegistrationLines(rd), nil
	case driver.ReportKindCloseFn:
		var cd driver.ReportFnCloseData
		if d, ok := data.(driver.ReportFnCloseData); ok {
			cd = d
		} else if d, ok := data.(*driver.ReportFnCloseData); ok {
			cd = *d
		} else {
			return nil, fmt.Errorf("invalid report data type: %T", data)
		}
		return buildCloseFnLines(cd), nil
	default:
		return nil, fmt.Errorf("unknown report kind: %v", kind)
	}
}

func buildRegistrationLines(regData driver.RegData) []kv {
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

func buildCloseFnLines(closeData driver.ReportFnCloseData) []kv {
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

// --------------------
// 2. ФОРМАТИРОВАНИЕ
// --------------------

func formatKeyValueText(lines []kv) string {
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

func toWindowsText(s string) string {
	return strings.ReplaceAll(s, "\n", "\r\n")
}

// --------------------
// 3. СКЛЕЙКА
// --------------------

func buildFormattedReportText(meta driver.ReportMeta) (string, error) {
	if meta.Data == nil {
		return "", fmt.Errorf("report data is nil")
	}
	if meta.Text != "" {
		return meta.Text, nil
	}
	lines, err := buildReportLines(meta.Data, meta.Kind)
	if err != nil {
		return "", err
	}
	return formatKeyValueText(lines), nil
}

// --------------------
// 4. UI
// --------------------

func RunReportModal(owner walk.Form, meta driver.ReportMeta) {
	var dlg *walk.Dialog
	var copyPB, savePB, closePB *walk.PushButton

	text, err := buildFormattedReportText(meta)
	if err != nil {
		walk.MsgBox(owner, "Ошибка", err.Error(), walk.MsgBoxIconError)
		return
	}

	err = d.Dialog{
		AssignTo:      &dlg,
		Title:         meta.Title,
		MinSize:       d.Size{Width: 500, Height: 400},
		MaxSize:       d.Size{Width: 500, Height: 400},
		Layout:        d.VBox{},
		DefaultButton: &copyPB,
		CancelButton:  &closePB,
		Children: []d.Widget{
			d.TextEdit{
				Text:     toWindowsText(text),
				ReadOnly: true,
				VScroll:  true,
				Font:     d.Font{Family: "Consolas", PointSize: 9},
			},
			d.Composite{
				Layout: d.HBox{Spacing: 6},
				Children: []d.Widget{
					d.HSpacer{},
					d.PushButton{
						AssignTo: &copyPB,
						Text:     "Копировать",
						OnClicked: func() {
							_ = walk.Clipboard().SetText(text)
						},
					},
					d.PushButton{
						AssignTo: &savePB,
						Text:     "Сохранить",
						OnClicked: func() {
							if err := saveReportText(text, meta.Kind); err != nil {
								walk.MsgBox(dlg, "Ошибка", err.Error(), walk.MsgBoxIconError)
							}
						},
					},
					d.PushButton{
						AssignTo: &closePB,
						Text:     "Закрыть",
						OnClicked: func() {
							dlg.Accept()
						},
					},
				},
			},
		},
	}.Create(owner)

	if err != nil {
		panic(err)
	}

	dlg.Run()
}

// RunRegistrationTextDialog - обратная совместимость
func RunRegistrationTextDialog(owner walk.Form, regData driver.RegData) {
	meta := driver.ReportMeta{
		Kind:  driver.ReportKindRegistration,
		Title: "Расшифрованные данные регистрации",
		Data:  regData,
	}
	RunReportModal(owner, meta)
}

// saveReportText сохраняет текст отчета в соответствующую папку.
func saveReportText(text string, kind driver.ReportKind) error {
	var dir string
	var filename string
	switch kind {
	case driver.ReportKindRegistration:
		dir = "./registration_docs"
		filename = "registration_data.txt"
	case driver.ReportKindCloseFn:
		dir = "./close_fn_docs"
		filename = "close_fn_data.txt"
	default:
		dir = "./reports"
		filename = "report_data.txt"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	fullPath := filepath.Join(dir, filename)
	return os.WriteFile(fullPath, []byte(text), 0644)
}

// formatRNM форматирует РНМ с пробелами по 4 цифры.
func formatRNM(rnm string) string {
	if len(rnm) != 16 {
		return rnm
	}
	return fmt.Sprintf("%s %s %s %s", rnm[0:4], rnm[4:8], rnm[8:12], rnm[12:16])
}

// formatDateTime форматирует дату и время как дд.мм.гггг чч:мм.
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

// formatFFD заменяет 4 на 1.2.
func formatFFD(ffd string) string {
	if ffd == "4" {
		return "1.2"
	}
	return ffd
}

// formatTaxSystems расшифровывает системы налогообложения.
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
