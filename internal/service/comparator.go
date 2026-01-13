package service

import (
	"fmt"
	"mitsuscanner/driver"
)

// Compare сравнивает два снапшота и возвращает список атомарных изменений.
func Compare(initial, current *SettingsSnapshot) []Change {
	var changes []Change

	if initial == nil || current == nil {
		return changes
	}

	// --- 1. ПРИНТЕР (Атомарно) ---
	pOld, pNew := initial.Printer, current.Printer
	if pOld.Model != pNew.Model {
		changes = append(changes, Change{
			ID: "Printer_Model", Description: "Модель принтера",
			OldValue: pOld.Model, NewValue: pNew.Model, Priority: PriorityNormal,
			ApplyFunc: func(d driver.Driver) error { return d.SetPrinterSettings(current.Printer) },
		})
	}
	if pOld.BaudRate != pNew.BaudRate {
		changes = append(changes, Change{
			ID: "Printer_Baud", Description: "Скорость печати",
			OldValue: pOld.BaudRate, NewValue: pNew.BaudRate, Priority: PriorityNormal,
			ApplyFunc: func(d driver.Driver) error { return d.SetPrinterSettings(current.Printer) },
		})
	}
	if pOld.Paper != pNew.Paper {
		changes = append(changes, Change{
			ID: "Printer_Paper", Description: "Ширина ленты",
			OldValue: fmt.Sprintf("%dмм", pOld.Paper), NewValue: fmt.Sprintf("%dмм", pNew.Paper), Priority: PriorityNormal,
			ApplyFunc: func(d driver.Driver) error { return d.SetPrinterSettings(current.Printer) },
		})
	}
	if pOld.Font != pNew.Font {
		changes = append(changes, Change{
			ID: "Printer_Font", Description: "Шрифт принтера",
			OldValue: pOld.Font, NewValue: pNew.Font, Priority: PriorityNormal,
			ApplyFunc: func(d driver.Driver) error { return d.SetPrinterSettings(current.Printer) },
		})
	}

	// --- 2. ДЕНЕЖНЫЙ ЯЩИК (Атомарно) ---
	dOld, dNew := initial.Drawer, current.Drawer
	if dOld.Pin != dNew.Pin || dOld.Rise != dNew.Rise || dOld.Fall != dNew.Fall {
		// Для ящика оставим одну группу, так как команда SetMoneyDrawerSettings принимает все 3 параметра сразу
		changes = append(changes, Change{
			ID: "Drawer_Settings", Description: "Параметры ден. ящика",
			OldValue: fmt.Sprintf("Pin:%d, R:%d, F:%d", dOld.Pin, dOld.Rise, dOld.Fall),
			NewValue: fmt.Sprintf("Pin:%d, R:%d, F:%d", dNew.Pin, dNew.Rise, dNew.Fall),
			Priority: PriorityNormal,
			ApplyFunc: func(d driver.Driver) error { return d.SetMoneyDrawerSettings(current.Drawer) },
		})
	}

	// --- 3. ЧАСОВОЙ ПОЯС ---
	if initial.Timezone != current.Timezone {
		changes = append(changes, Change{
			ID: "Timezone", Description: "Часовой пояс",
			OldValue: initial.Timezone, NewValue: current.Timezone, Priority: PriorityNormal,
			ApplyFunc: func(d driver.Driver) error { return d.SetTimezone(current.Timezone) },
		})
	}

	// --- 4. ОПЦИИ (b1-b9) ---
	checkOption := func(id, name string, optNum int, oldV, newV int) {
		if oldV != newV {
			changes = append(changes, Change{
				ID: id, Description: "Опция: " + name,
				OldValue: oldV, NewValue: newV, Priority: PriorityNormal,
				ApplyFunc: func(d driver.Driver) error { return d.SetOption(optNum, newV) },
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

	// --- 5. ОФД (Атомарно) ---
	ofdO, ofdN := initial.Ofd, current.Ofd
	if ofdO.Addr != ofdN.Addr || ofdO.Port != ofdN.Port {
		changes = append(changes, Change{
			ID: "Ofd_Addr", Description: "Адрес сервера ОФД",
			OldValue: fmt.Sprintf("%s:%d", ofdO.Addr, ofdO.Port),
			NewValue: fmt.Sprintf("%s:%d", ofdN.Addr, ofdN.Port),
			Priority: PriorityNormal,
			ApplyFunc: func(d driver.Driver) error { return d.SetOfdSettings(current.Ofd) },
		})
	}
	if ofdO.Client != ofdN.Client {
		changes = append(changes, Change{
			ID: "Ofd_Client", Description: "Режим клиента ОФД",
			OldValue: ofdO.Client, NewValue: ofdN.Client, Priority: PriorityNormal,
			ApplyFunc: func(d driver.Driver) error { return d.SetOfdSettings(current.Ofd) },
		})
	}
	if ofdO.TimerFN != ofdN.TimerFN || ofdO.TimerOFD != ofdN.TimerOFD {
		changes = append(changes, Change{
			ID: "Ofd_Timers", Description: "Таймеры ОФД/ФН",
			OldValue: fmt.Sprintf("ФН:%d, ОФД:%d", ofdO.TimerFN, ofdO.TimerOFD),
			NewValue: fmt.Sprintf("ФН:%d, ОФД:%d", ofdN.TimerFN, ofdN.TimerOFD),
			Priority: PriorityNormal,
			ApplyFunc: func(d driver.Driver) error { return d.SetOfdSettings(current.Ofd) },
		})
	}

	// --- 6. ОИСМ ---
	if initial.Oism.Addr != current.Oism.Addr || initial.Oism.Port != current.Oism.Port {
		changes = append(changes, Change{
			ID: "Oism_Addr", Description: "Адрес сервера ОИСМ",
			OldValue: fmt.Sprintf("%s:%d", initial.Oism.Addr, initial.Oism.Port),
			NewValue: fmt.Sprintf("%s:%d", current.Oism.Addr, current.Oism.Port),
			Priority: PriorityNormal,
			ApplyFunc: func(d driver.Driver) error { return d.SetOismSettings(current.Oism) },
		})
	}

	// --- 7. LAN (Атомарно, приоритет Network) ---
	lO, lN := initial.Lan, current.Lan
	if lO.Addr != lN.Addr || lO.Port != lN.Port || lO.Mask != lN.Mask || lO.Gw != lN.Gw || lO.Dns != lN.Dns {
		changes = append(changes, Change{
			ID: "Lan_Settings", Description: "Сетевые настройки LAN",
			OldValue: fmt.Sprintf("IP:%s, P:%d", lO.Addr, lO.Port),
			NewValue: fmt.Sprintf("IP:%s, P:%d", lN.Addr, lN.Port),
			Priority: PriorityNetwork,
			ApplyFunc: func(d driver.Driver) error { return d.SetLanSettings(current.Lan) },
		})
	}

	// --- 8. КЛИШЕ (По типам) ---
	clicheNames := map[int]string{1: "Заголовок", 2: "После пользователя", 3: "Подвал", 4: "Конец чека"}
	for typeID := 1; typeID <= 4; typeID++ {
		if !clichesEqual(initial.Cliches[typeID], current.Cliches[typeID]) {
			tid, lNew := typeID, current.Cliches[typeID]
			changes = append(changes, Change{
				ID: fmt.Sprintf("Cliche_%d", tid), Description: "Клише: " + clicheNames[tid],
				OldValue: "(изменено)", NewValue: "(новые данные)", Priority: PriorityCliche,
				ApplyFunc: func(d driver.Driver) error {
					for i, line := range lNew {
						if err := d.SetHeaderLine(tid, i, line.Text, line.Format); err != nil {
							return err
						}
					}
					return nil
				},
			})
		}
	}

	return changes
}
