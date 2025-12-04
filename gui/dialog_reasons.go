package main

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
		"1 - Изменение адреса или места установки",
		"2 - Смена оператора фискальных данных (ОФД)",
		"3 - Изменение сведений об автомате",
		"4 - Замена фискального накопителя (ФН)",
		"5 - Переход В режим передачи данных (Онлайн)",
		"6 - Переход ИЗ режима передачи данных (Автономный)",
		"7 - Изменение наименования (ФИО/Названия)",
		"8 - Иные причины",
	}

	// Создаем слайс для хранения указателей на чекбоксы
	checkBoxes := make([]*walk.CheckBox, len(reasons))
	var checkWidgets []d.Widget

	// 3. Генерация виджетов
	for i, text := range reasons {
		idx := i
		code := i + 1
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
									codes = append(codes, fmt.Sprintf("%d", i+1))
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
