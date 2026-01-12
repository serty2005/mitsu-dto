package settings

import (
	"fmt"
	"mitsuscanner/internal/models"
	"strconv"
)

// CompareSnapshots сравнивает два снимка настроек и возвращает список изменений
// Возвращает срез указателей на ChangeItem, где каждое изменение описывает
// разницу между старым и новым значением параметра
func CompareSnapshots(old, new *models.Snapshot) []*models.ChangeItem {
	var changes []*models.ChangeItem

	// Сравнение сетевых настроек
	if new.OfdString != old.OfdString {
		changes = append(changes, &models.ChangeItem{
			Category: "ОФД",
			Name:     "Сервер",
			OldVal:   old.OfdString,
			NewVal:   new.OfdString,
		})
	}
	if new.OfdClient != old.OfdClient {
		oldVal := "Внешний"
		if old.OfdClient == "0" {
			oldVal = "Встроенный"
		}
		newVal := "Внешний"
		if new.OfdClient == "0" {
			newVal = "Встроенный"
		}
		changes = append(changes, &models.ChangeItem{
			Category: "ОФД",
			Name:     "Режим клиента",
			OldVal:   oldVal,
			NewVal:   newVal,
		})
	}
	if new.TimerFN != old.TimerFN {
		changes = append(changes, &models.ChangeItem{
			Category: "ОФД",
			Name:     "Таймер ФН",
			OldVal:   strconv.Itoa(old.TimerFN),
			NewVal:   strconv.Itoa(new.TimerFN),
		})
	}
	if new.TimerOFD != old.TimerOFD {
		changes = append(changes, &models.ChangeItem{
			Category: "ОФД",
			Name:     "Таймер ОФД",
			OldVal:   strconv.Itoa(old.TimerOFD),
			NewVal:   strconv.Itoa(new.TimerOFD),
		})
	}
	if new.OismString != old.OismString {
		changes = append(changes, &models.ChangeItem{
			Category: "ОИСМ",
			Name:     "Сервер",
			OldVal:   old.OismString,
			NewVal:   new.OismString,
		})
	}
	if new.LanAddr != old.LanAddr {
		changes = append(changes, &models.ChangeItem{
			Category: "LAN",
			Name:     "IP адрес",
			OldVal:   old.LanAddr,
			NewVal:   new.LanAddr,
		})
	}
	if new.LanPort != old.LanPort {
		changes = append(changes, &models.ChangeItem{
			Category: "LAN",
			Name:     "Порт",
			OldVal:   strconv.Itoa(old.LanPort),
			NewVal:   strconv.Itoa(new.LanPort),
		})
	}
	if new.LanMask != old.LanMask {
		changes = append(changes, &models.ChangeItem{
			Category: "LAN",
			Name:     "Маска",
			OldVal:   old.LanMask,
			NewVal:   new.LanMask,
		})
	}
	if new.LanDns != old.LanDns {
		changes = append(changes, &models.ChangeItem{
			Category: "LAN",
			Name:     "DNS",
			OldVal:   old.LanDns,
			NewVal:   new.LanDns,
		})
	}
	if new.LanGw != old.LanGw {
		changes = append(changes, &models.ChangeItem{
			Category: "LAN",
			Name:     "Шлюз",
			OldVal:   old.LanGw,
			NewVal:   new.LanGw,
		})
	}

	// Параметры принтера
	if new.PrintModel != old.PrintModel {
		oldVal := "RP-809"
		if old.PrintModel == "2" {
			oldVal = "F80"
		}
		newVal := "RP-809"
		if new.PrintModel == "2" {
			newVal = "F80"
		}
		changes = append(changes, &models.ChangeItem{
			Category: "Принтер",
			Name:     "Модель",
			OldVal:   oldVal,
			NewVal:   newVal,
		})
	}
	if new.PrintBaud != old.PrintBaud {
		changes = append(changes, &models.ChangeItem{
			Category: "Принтер",
			Name:     "Скорость",
			OldVal:   old.PrintBaud,
			NewVal:   new.PrintBaud,
		})
	}
	if new.PrintPaper != old.PrintPaper {
		changes = append(changes, &models.ChangeItem{
			Category: "Принтер",
			Name:     "Ширина бумаги",
			OldVal:   strconv.Itoa(old.PrintPaper),
			NewVal:   strconv.Itoa(new.PrintPaper),
		})
	}
	if new.PrintFont != old.PrintFont {
		changes = append(changes, &models.ChangeItem{
			Category: "Принтер",
			Name:     "Шрифт",
			OldVal:   strconv.Itoa(old.PrintFont),
			NewVal:   strconv.Itoa(new.PrintFont),
		})
	}

	// Опции
	if new.OptTimezone != old.OptTimezone {
		changes = append(changes, &models.ChangeItem{
			Category: "Опции",
			Name:     "Часовой пояс",
			OldVal:   old.OptTimezone,
			NewVal:   new.OptTimezone,
		})
	}
	if new.OptCut != old.OptCut {
		oldVal := "Нет"
		if old.OptCut {
			oldVal = "Да"
		}
		newVal := "Нет"
		if new.OptCut {
			newVal = "Да"
		}
		changes = append(changes, &models.ChangeItem{
			Category: "Опции",
			Name:     "Автоотрез",
			OldVal:   oldVal,
			NewVal:   newVal,
		})
	}
	if new.OptAutoTest != old.OptAutoTest {
		oldVal := "Нет"
		if old.OptAutoTest {
			oldVal = "Да"
		}
		newVal := "Нет"
		if new.OptAutoTest {
			newVal = "Да"
		}
		changes = append(changes, &models.ChangeItem{
			Category: "Опции",
			Name:     "Автотест",
			OldVal:   oldVal,
			NewVal:   newVal,
		})
	}
	if new.OptNearEnd != old.OptNearEnd {
		oldVal := "Нет"
		if old.OptNearEnd {
			oldVal = "Да"
		}
		newVal := "Нет"
		if new.OptNearEnd {
			newVal = "Да"
		}
		changes = append(changes, &models.ChangeItem{
			Category: "Опции",
			Name:     "Звук бумаги",
			OldVal:   oldVal,
			NewVal:   newVal,
		})
	}
	if new.OptTextQR != old.OptTextQR {
		oldVal := "Нет"
		if old.OptTextQR {
			oldVal = "Да"
		}
		newVal := "Нет"
		if new.OptTextQR {
			newVal = "Да"
		}
		changes = append(changes, &models.ChangeItem{
			Category: "Опции",
			Name:     "Текст у QR",
			OldVal:   oldVal,
			NewVal:   newVal,
		})
	}
	if new.OptCountInCheck != old.OptCountInCheck {
		oldVal := "Нет"
		if old.OptCountInCheck {
			oldVal = "Да"
		}
		newVal := "Нет"
		if new.OptCountInCheck {
			newVal = "Да"
		}
		changes = append(changes, &models.ChangeItem{
			Category: "Опции",
			Name:     "Количество покупок",
			OldVal:   oldVal,
			NewVal:   newVal,
		})
	}
	if new.OptQRPos != old.OptQRPos {
		oldVal := "Слева"
		switch old.OptQRPos {
		case "1":
			oldVal = "По центру"
		case "2":
			oldVal = "Справа"
		}
		newVal := "Слева"
		switch new.OptQRPos {
		case "1":
			newVal = "По центру"
		case "2":
			newVal = "Справа"
		}
		changes = append(changes, &models.ChangeItem{
			Category: "Опции",
			Name:     "Позиция QR",
			OldVal:   oldVal,
			NewVal:   newVal,
		})
	}
	if new.OptRounding != old.OptRounding {
		oldVal := "Нет"
		switch old.OptRounding {
		case "1":
			oldVal = "0.10"
		case "2":
			oldVal = "0.50"
		case "3":
			oldVal = "1.00"
		}
		newVal := "Нет"
		switch new.OptRounding {
		case "1":
			newVal = "0.10"
		case "2":
			newVal = "0.50"
		case "3":
			newVal = "1.00"
		}
		changes = append(changes, &models.ChangeItem{
			Category: "Опции",
			Name:     "Округление",
			OldVal:   oldVal,
			NewVal:   newVal,
		})
	}
	if new.OptDrawerTrig != old.OptDrawerTrig {
		oldVal := "Нет"
		switch old.OptDrawerTrig {
		case "1":
			oldVal = "Наличные"
		case "2":
			oldVal = "Безнал"
		case "3":
			oldVal = "Всегда"
		}
		newVal := "Нет"
		switch new.OptDrawerTrig {
		case "1":
			newVal = "Наличные"
		case "2":
			newVal = "Безнал"
		case "3":
			newVal = "Всегда"
		}
		changes = append(changes, &models.ChangeItem{
			Category: "Опции",
			Name:     "Триггер ящика",
			OldVal:   oldVal,
			NewVal:   newVal,
		})
	}
	if new.OptB9 != old.OptB9 {
		changes = append(changes, &models.ChangeItem{
			Category: "Опции",
			Name:     "Опция b9",
			OldVal:   old.OptB9,
			NewVal:   new.OptB9,
		})
	}

	// Денежный ящик
	if new.DrawerPin != old.DrawerPin {
		changes = append(changes, &models.ChangeItem{
			Category: "Денежный ящик",
			Name:     "PIN",
			OldVal:   strconv.Itoa(old.DrawerPin),
			NewVal:   strconv.Itoa(new.DrawerPin),
		})
	}
	if new.DrawerRise != old.DrawerRise {
		changes = append(changes, &models.ChangeItem{
			Category: "Денежный ящик",
			Name:     "Rise (ms)",
			OldVal:   strconv.Itoa(old.DrawerRise),
			NewVal:   strconv.Itoa(new.DrawerRise),
		})
	}
	if new.DrawerFall != old.DrawerFall {
		changes = append(changes, &models.ChangeItem{
			Category: "Денежный ящик",
			Name:     "Fall (ms)",
			OldVal:   strconv.Itoa(old.DrawerFall),
			NewVal:   strconv.Itoa(new.DrawerFall),
		})
	}

	// Клише
	for i, current := range new.ClicheItems {
		if i >= len(old.ClicheItems) {
			continue
		}
		old := old.ClicheItems[i]
		if current.Text != old.Text {
			changes = append(changes, &models.ChangeItem{
				Category: "Клише",
				Name:     fmt.Sprintf("Строка %d", i+1),
				OldVal:   old.Text,
				NewVal:   current.Text,
			})
		}
	}

	return changes
}
