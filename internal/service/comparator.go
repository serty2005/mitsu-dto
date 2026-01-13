package service

import (
	"fmt"
	"mitsuscanner/driver"
)

// Compare сравнивает два снапшота и возвращает список изменений.
// initial - состояние "Считать" (база).
// current - текущее состояние UI.
func Compare(initial, current *SettingsSnapshot) []Change {
	var changes []Change

	if initial == nil || current == nil {
		return changes
	}

	// 1. Сравнение настроек Принтера
	if initial.Printer != current.Printer {
		changes = append(changes, Change{
			ID:          "Printer",
			Description: "Настройки принтера (Модель/Скорость/Бумага/Шрифт)",
			OldValue:    fmt.Sprintf("%s, %d, %dмм", initial.Printer.Model, initial.Printer.BaudRate, initial.Printer.Width),
			NewValue:    fmt.Sprintf("%s, %d, %dмм", current.Printer.Model, current.Printer.BaudRate, current.Printer.Width),
			Priority:    PriorityNormal,
			ApplyFunc: func(d driver.Driver) error {
				return d.SetPrinterSettings(current.Printer)
			},
		})
	}

	// 2. Сравнение Денежного ящика
	if initial.Drawer != current.Drawer {
		changes = append(changes, Change{
			ID:          "Drawer",
			Description: "Настройки денежного ящика",
			OldValue:    fmt.Sprintf("Pin:%d Rise:%d Fall:%d", initial.Drawer.Pin, initial.Drawer.Rise, initial.Drawer.Fall),
			NewValue:    fmt.Sprintf("Pin:%d Rise:%d Fall:%d", current.Drawer.Pin, current.Drawer.Rise, current.Drawer.Fall),
			Priority:    PriorityNormal,
			ApplyFunc: func(d driver.Driver) error {
				return d.SetMoneyDrawerSettings(current.Drawer)
			},
		})
	}

	// 3. Сравнение Часового пояса
	if initial.Timezone != current.Timezone {
		changes = append(changes, Change{
			ID:          "Timezone",
			Description: "Часовой пояс",
			OldValue:    initial.Timezone,
			NewValue:    current.Timezone,
			Priority:    PriorityNormal,
			ApplyFunc: func(d driver.Driver) error {
				return d.SetTimezone(current.Timezone)
			},
		})
	}

	// 4. Сравнение Опций (по одной)
	// Helper для добавления опции
	checkOption := func(id string, name string, optNum int, oldVal, newVal int) {
		if oldVal != newVal {
			changes = append(changes, Change{
				ID:          id,
				Description: fmt.Sprintf("Опция: %s", name),
				OldValue:    oldVal,
				NewValue:    newVal,
				Priority:    PriorityNormal,
				ApplyFunc: func(d driver.Driver) error {
					return d.SetOption(optNum, newVal)
				},
			})
		}
	}

	optsOld := initial.Options
	optsNew := current.Options

	checkOption("OptQRPos", "Позиция QR (b1)", 1, optsOld.B1, optsNew.B1)
	checkOption("OptRounding", "Округление (b2)", 2, optsOld.B2, optsNew.B2)
	checkOption("OptCut", "Автоотрез (b3)", 3, optsOld.B3, optsNew.B3)
	checkOption("OptAutoTest", "Автотест (b4)", 4, optsOld.B4, optsNew.B4)
	checkOption("OptDrawerTrig", "Триггер ящика (b5)", 5, optsOld.B5, optsNew.B5)
	checkOption("OptNearEnd", "Датчик бумаги (b6)", 6, optsOld.B6, optsNew.B6)
	checkOption("OptTextQR", "Текст у QR (b7)", 7, optsOld.B7, optsNew.B7)
	checkOption("OptCountInCheck", "Счетчик покупок (b8)", 8, optsOld.B8, optsNew.B8)
	checkOption("OptB9", "Опция b9", 9, optsOld.B9, optsNew.B9)

	// 5. Сравнение ОФД (группа)
	if initial.Ofd != current.Ofd {
		changes = append(changes, Change{
			ID:          "OFD",
			Description: "Настройки ОФД",
			OldValue:    fmt.Sprintf("%s:%d (Cl:%s)", initial.Ofd.Addr, initial.Ofd.Port, initial.Ofd.Client),
			NewValue:    fmt.Sprintf("%s:%d (Cl:%s)", current.Ofd.Addr, current.Ofd.Port, current.Ofd.Client),
			Priority:    PriorityNormal, // ОФД настройки не рвут связь, если мы не меняем режим LAN клиента глобально
			ApplyFunc: func(d driver.Driver) error {
				return d.SetOfdSettings(current.Ofd)
			},
		})
	}

	// 6. Сравнение ОИСМ
	if initial.Oism != current.Oism {
		changes = append(changes, Change{
			ID:          "OISM",
			Description: "Настройки ОИСМ",
			OldValue:    fmt.Sprintf("%s:%d", initial.Oism.Addr, initial.Oism.Port),
			NewValue:    fmt.Sprintf("%s:%d", current.Oism.Addr, current.Oism.Port),
			Priority:    PriorityNormal,
			ApplyFunc: func(d driver.Driver) error {
				return d.SetOismSettings(current.Oism)
			},
		})
	}

	// 7. Сравнение LAN (Высокий приоритет риска / Низкий приоритет выполнения)
	if initial.Lan != current.Lan {
		changes = append(changes, Change{
			ID:          "LAN",
			Description: "Сетевые настройки (LAN) - Потребуется перезагрузка!",
			OldValue:    fmt.Sprintf("%s / %s", initial.Lan.Addr, initial.Lan.Port),
			NewValue:    fmt.Sprintf("%s / %s", current.Lan.Addr, current.Lan.Port),
			Priority:    PriorityNetwork, // Выполняется последним
			ApplyFunc: func(d driver.Driver) error {
				return d.SetLanSettings(current.Lan)
			},
		})
	}

	// 8. Сравнение Клише (по типам)
	// Мы сравниваем целиком массивы строк для каждого типа.
	clicheNames := map[int]string{
		1: "Заголовок",
		2: "После пользователя",
		3: "Подвал",
		4: "Конец чека",
	}

	for typeID := 1; typeID <= 4; typeID++ {
		linesOld := initial.Cliches[typeID]
		linesNew := current.Cliches[typeID]

		if !clichesEqual(linesOld, linesNew) {
			// Замыкание переменных для ApplyFunc
			tid := typeID
			lNew := linesNew

			changes = append(changes, Change{
				ID:          fmt.Sprintf("Cliche_%d", tid),
				Description: fmt.Sprintf("Клише: %s", clicheNames[tid]),
				OldValue:    "(изменено)",
				NewValue:    "(новые данные)",
				Priority:    PriorityCliche,
				ApplyFunc: func(d driver.Driver) error {
					// Записываем все строки данного типа
					for i, line := range lNew {
						if err := d.SetHeaderLine(tid, i, line.Text, line.Format); err != nil {
							return fmt.Errorf("ошибка записи строки %d: %w", i, err)
						}
					}
					return nil
				},
			})
		}
	}

	return changes
}
