package gui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
)

// RunReasonDialog открывает диалог.
func RunReasonDialog(owner walk.Form, currentCodes string) (string, bool) {
	var dlg *walk.Dialog
	var acceptPB, cancelPB *walk.PushButton

	// Переменная для сохранения результата, пока окно еще открыто
	var resultString string

	// 1. Парсинг текущих кодов
	selectedMap := make(map[int]bool)
	if currentCodes != "" {
		parts := strings.Split(currentCodes, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if code, err := strconv.Atoi(p); err == nil {
				selectedMap[code] = true
			}
		}
	}

	// 2. Список причин
	reasons := []string{
		"1 - Замена ФН",
		"2 - Замена ОФД",
		"3 - Изменение реквизитов пользователя",
		"4 - Изменение адреса/места установки",
		"5 - Перевод кассы из автономного режима в режим передачи данных",
		"6 - Перевод кассы из режима передачи данных в автономный режим",
		"7 - Изменение версии модели ККТ",
		"8 - Изменение перечня СНО",
		"9 - Изменение номера автомата",
		"10 - Отключение автоматического режима (осуществление расчетов кассиром)",
		"11 - Включение автоматического режима",
		"12 - Включение режима БСО",
		"13 - Отключение режима БСО",
		"14 - Отключение режима расчетов в сети Интернет",
		"15 - Включение режима расчетов в Интернет (можно не печатать чек и БСО)",
		"18 - Отключение режима азартных игр",
		"19 - Включение режима азартных игр (прием ставок, выплата выигрыша)",
		"20 - Отключение режима лотерей",
		"21 - Включение режима лотерей (продажа билетов, выплата выигрышей)",
		"22 - Изменение версии ФФД",
		"32 - Иные причины",
	}

	// Создаем слайс для хранения указателей на чекбоксы и соответствующих кодов
	checkBoxes := make([]*walk.CheckBox, len(reasons))
	codesList := make([]int, len(reasons))
	var checkWidgets []d.Widget

	// 3. Генерация виджетов
	for i, text := range reasons {
		idx := i
		parts := strings.Split(text, " - ")
		code, _ := strconv.Atoi(parts[0])
		codesList[idx] = code
		isChecked := selectedMap[code]

		checkWidgets = append(checkWidgets, d.CheckBox{
			AssignTo: &checkBoxes[idx], // Привязываем Go-структуру
			Text:     text,
			Checked:  isChecked,
			MinSize:  d.Size{Width: 350},
		})
	}

	// 4. Описание диалога
	err := d.Dialog{
		AssignTo:      &dlg,
		Title:         "Причины перерегистрации (ФНС)",
		MinSize:       d.Size{Width: 360, Height: 320},
		MaxSize:       d.Size{Width: 360, Height: 320},
		FixedSize:     true,
		Layout:        d.VBox{},
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		Children: []d.Widget{
			d.ScrollView{
				Layout:          d.VBox{MarginsZero: true},
				HorizontalFixed: true,
				Children: []d.Widget{
					d.Composite{
						// Выравниваем слева сверху
						Layout:   d.VBox{MarginsZero: true, Spacing: 2, Alignment: d.AlignHNearVNear},
						Children: checkWidgets,
					},
				},
			},
			d.Composite{
				Layout: d.HBox{Margins: d.Margins{Top: 5}},
				Children: []d.Widget{
					d.HSpacer{},
					d.PushButton{
						AssignTo: &acceptPB,
						Text:     "OK",
						OnClicked: func() {
							// Считываем данные ЗДЕСЬ, пока окно живо
							var codes []string
							for i, cb := range checkBoxes {
								if cb.Checked() {
									codes = append(codes, fmt.Sprintf("%d", codesList[i]))
								}
							}
							resultString = strings.Join(codes, ",")

							// Закрываем окно с результатом OK
							dlg.Close(walk.DlgCmdOK)
						},
					},
					d.PushButton{
						AssignTo: &cancelPB,
						Text:     "Отмена",
						OnClicked: func() {
							dlg.Close(walk.DlgCmdCancel)
						},
					},
				},
			},
		},
	}.Create(owner)

	if err != nil {
		fmt.Println("Error creating dialog:", err)
		return "", false
	}

	// Запуск
	// Если вернулся OK, значит мы уже заполнили resultString внутри OnClicked
	if dlg.Run() == walk.DlgCmdOK {
		return resultString, true
	}

	return "", false
}
