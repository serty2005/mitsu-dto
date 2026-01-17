package settings

import (
	"fmt"

	"mitsuscanner/internal/domain/models"
	"mitsuscanner/internal/domain/ports"
)

// Compare сравнивает два снапшота настроек и возвращает список атомарных изменений
func Compare(initial, current *models.SettingsSnapshot) []Change {
	var changes []Change

	if initial == nil || current == nil {
		return changes
	}

	// --- 1. ПРИНТЕР ---
	pOld, pNew := initial.Printer, current.Printer
	if pOld.Model != pNew.Model {
		changes = append(changes, Change{
			ID: "Printer_Model", Description: "Модель принтера",
			OldValue: pOld.Model, NewValue: pNew.Model, Priority: models.PriorityNormal,
			ApplyFunc: func(d ports.Driver) error { return d.SetPrinterSettings(current.Printer) },
		})
	}
	if pOld.BaudRate != pNew.BaudRate {
		changes = append(changes, Change{
			ID: "Printer_Baud", Description: "Скорость печати",
			OldValue: pOld.BaudRate, NewValue: pNew.BaudRate, Priority: models.PriorityNormal,
			ApplyFunc: func(d ports.Driver) error { return d.SetPrinterSettings(current.Printer) },
		})
	}
	if pOld.Paper != pNew.Paper {
		changes = append(changes, Change{
			ID: "Printer_Paper", Description: "Ширина ленты",
			OldValue: fmt.Sprintf("%dмм", pOld.Paper), NewValue: fmt.Sprintf("%dмм", pNew.Paper), Priority: models.PriorityNormal,
			ApplyFunc: func(d ports.Driver) error { return d.SetPrinterSettings(current.Printer) },
		})
	}
	if pOld.Font != pNew.Font {
		changes = append(changes, Change{
			ID: "Printer_Font", Description: "Шрифт принтера",
			OldValue: pOld.Font, NewValue: pNew.Font, Priority: models.PriorityNormal,
			ApplyFunc: func(d ports.Driver) error { return d.SetPrinterSettings(current.Printer) },
		})
	}

	// --- 2. ДЕНЕЖНЫЙ ЯЩИК ---
	dOld, dNew := initial.Drawer, current.Drawer
	if dOld.Pin != dNew.Pin || dOld.Rise != dNew.Rise || dOld.Fall != dNew.Fall {
		changes = append(changes, Change{
			ID: "Drawer_Settings", Description: "Параметры ден. ящика",
			OldValue:  fmt.Sprintf("Pin:%d, R:%d, F:%d", dOld.Pin, dOld.Rise, dOld.Fall),
			NewValue:  fmt.Sprintf("Pin:%d, R:%d, F:%d", dNew.Pin, dNew.Rise, dNew.Fall),
			Priority:  models.PriorityNormal,
			ApplyFunc: func(d ports.Driver) error { return d.SetMoneyDrawerSettings(current.Drawer) },
		})
	}

	// --- 3. ЧАСОВОЙ ПОЯС ---
	if initial.Timezone != current.Timezone {
		changes = append(changes, Change{
			ID: "Timezone", Description: "Часовой пояс",
			OldValue: initial.Timezone, NewValue: current.Timezone, Priority: models.PriorityNormal,
			ApplyFunc: func(d ports.Driver) error { return d.SetTimezone(current.Timezone) },
		})
	}

	// --- 4. ОПЦИИ (b1-b9) ---
	checkOption := func(id, name string, optNum int, oldV, newV int) {
		if oldV != newV {
			changes = append(changes, Change{
				ID: id, Description: "Опция: " + name,
				OldValue: oldV, NewValue: newV, Priority: models.PriorityNormal,
				ApplyFunc: func(d ports.Driver) error { return d.SetOption(optNum, newV) },
			})
		}
	}
	oO, oN := initial.Options, current.Options
	checkOption("Opt_QRPos", "Позиция QR", 1, oO.B1, oN.B1)
	checkOption("Opt_Rounding", "Округление", 2, oO.B2, oN.B2)
	checkOption("Opt_Cut", "Автоотрез", 3, oO.B3, oN.B3)
	checkOption("Opt_AutoTest", "Автотест", 4, oO.B4, oN.B4)
	checkOption("Opt_DrawerTrig", "Триггер ящика", 5, oO.B5, oN.B5)
	checkOption("Opt_NearEnd", "Датчик бумаги", 6, oO.B6, oN.B6)
	checkOption("Opt_TextQR", "Текст у QR", 7, oO.B7, oN.B7)
	checkOption("Opt_CountInCheck", "Счетчик покупок", 8, oO.B8, oN.B8)
	checkOption("Opt_B9", "СНО по умолчанию (b9)", 9, oO.B9, oN.B9)

	// --- 5. ОФД ---
	ofdO, ofdN := initial.Ofd, current.Ofd
	if ofdO.Addr != ofdN.Addr || ofdO.Port != ofdN.Port {
		changes = append(changes, Change{
			ID: "Ofd_Addr", Description: "Адрес сервера ОФД",
			OldValue:  fmt.Sprintf("%s:%d", ofdO.Addr, ofdO.Port),
			NewValue:  fmt.Sprintf("%s:%d", ofdN.Addr, ofdN.Port),
			Priority:  models.PriorityNormal,
			ApplyFunc: func(d ports.Driver) error { return d.SetOfdSettings(current.Ofd) },
		})
	}
	if ofdO.Client != ofdN.Client {
		changes = append(changes, Change{
			ID: "Ofd_Client", Description: "Режим клиента ОФД",
			OldValue: ofdO.Client, NewValue: ofdN.Client, Priority: models.PriorityNormal,
			ApplyFunc: func(d ports.Driver) error { return d.SetOfdSettings(current.Ofd) },
		})
	}
	if ofdO.TimerFN != ofdN.TimerFN || ofdO.TimerOFD != ofdN.TimerOFD {
		changes = append(changes, Change{
			ID: "Ofd_Timers", Description: "Таймеры ОФД/ФН",
			OldValue:  fmt.Sprintf("ФН:%d, ОФД:%d", ofdO.TimerFN, ofdO.TimerOFD),
			NewValue:  fmt.Sprintf("ФН:%d, ОФД:%d", ofdN.TimerFN, ofdN.TimerOFD),
			Priority:  models.PriorityNormal,
			ApplyFunc: func(d ports.Driver) error { return d.SetOfdSettings(current.Ofd) },
		})
	}

	// --- 6. ОИСМ ---
	if initial.Oism.Addr != current.Oism.Addr || initial.Oism.Port != current.Oism.Port {
		changes = append(changes, Change{
			ID: "Oism_Addr", Description: "Адрес сервера ОИСМ",
			OldValue:  fmt.Sprintf("%s:%d", initial.Oism.Addr, initial.Oism.Port),
			NewValue:  fmt.Sprintf("%s:%d", current.Oism.Addr, current.Oism.Port),
			Priority:  models.PriorityNormal,
			ApplyFunc: func(d ports.Driver) error { return d.SetOismSettings(current.Oism) },
		})
	}

	// --- 7. LAN ---
	lO, lN := initial.Lan, current.Lan
	if lO.Addr != lN.Addr || lO.Port != lN.Port || lO.Mask != lN.Mask || lO.Gw != lN.Gw || lO.Dns != lN.Dns {
		changes = append(changes, Change{
			ID: "Lan_Settings", Description: "Сетевые настройки LAN",
			OldValue:  fmt.Sprintf("IP:%s, P:%d", lO.Addr, lO.Port),
			NewValue:  fmt.Sprintf("IP:%s, P:%d", lN.Addr, lN.Port),
			Priority:  models.PriorityNetwork,
			ApplyFunc: func(d ports.Driver) error { return d.SetLanSettings(current.Lan) },
		})
	}

	// --- 8. КЛИШЕ (Построчное сравнение и запись) ---
	clicheNames := map[int]string{1: "Заголовок", 2: "После пользователя", 3: "Подвал", 4: "Конец чека"}

	for typeID := 1; typeID <= 4; typeID++ {
		oldLines := initial.Cliches[typeID]
		newLines := current.Cliches[typeID]

		maxLen := len(oldLines)
		if len(newLines) > maxLen {
			maxLen = len(newLines)
		}

		for i := 0; i < maxLen; i++ {
			var oldL, newL models.ClicheLineData
			if i < len(oldLines) {
				oldL = oldLines[i]
			}
			if i < len(newLines) {
				newL = newLines[i]
			}

			// Пропускаем, только если обе строки идентичны (или обе пустые).
			// Изменение (включая очистку) будет обработано.
			if oldL.Text == newL.Text && oldL.Format == newL.Format {
				continue
			}

			// Если дошли сюда, значит есть изменение.
			tid := typeID
			lineNum := i
			finalText := newL.Text
			finalFormat := newL.Format

			oldVal := fmt.Sprintf("\"%s-%s\"", oldL.Format, oldL.Text)
			newVal := fmt.Sprintf("\"%s-%s\"", newL.Format, newL.Text)

			changes = append(changes, Change{
				ID:          fmt.Sprintf("Cliche_%d_%d", tid, lineNum),
				Description: fmt.Sprintf("Клише \"%s\", Строка %d", clicheNames[tid], lineNum+1),
				OldValue:    oldVal,
				NewValue:    newVal,
				Priority:    models.PriorityCliche,
				ApplyFunc: func(d ports.Driver) error {
					return d.SetHeaderLine(tid, lineNum, finalText, finalFormat)
				},
			})
		}
	}

	return changes
}
