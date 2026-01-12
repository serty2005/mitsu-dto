package gui

import (
	"fmt"
	"sort"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
	"mitsuscanner/internal/models"
)

// SummaryDialogModel модель для таблицы изменений
type SummaryDialogModel struct {
	walk.TableModelBase
	Items []*models.ChangeItem
}

func (m *SummaryDialogModel) RowCount() int {
	return len(m.Items)
}

func (m *SummaryDialogModel) Value(row, col int) interface{} {
	item := m.Items[row]
	switch col {
	case 0:
		return item.Remove
	case 1:
		return item.Name
	case 2:
		return item.OldVal
	case 3:
		return item.NewVal
	}
	return ""
}

func (m *SummaryDialogModel) Checked(row int) bool {
	return m.Items[row].Remove
}

func (m *SummaryDialogModel) SetChecked(row int, checked bool) error {
	m.Items[row].Remove = checked
	return nil
}

// RunSummaryDialog открывает диалог подтверждения изменений.
// Возвращает список только тех изменений, которые пользователь подтвердил.
func RunSummaryDialog(owner walk.Form, changes []*models.ChangeItem) ([]*models.ChangeItem, bool) {
	var dlg *walk.Dialog
	var tv *walk.TableView
	var acceptPB, cancelPB *walk.PushButton

	// По умолчанию все галочки сняты (не удалять)
	for _, c := range changes {
		c.Remove = false
	}

	// Сортировка по категориям
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Category == changes[j].Category {
			return changes[i].Name < changes[j].Name
		}
		return changes[i].Category < changes[j].Category
	})

	model := &SummaryDialogModel{Items: changes}

	err := d.Dialog{
		AssignTo:      &dlg,
		Title:         "Подтверждение изменений",
		MinSize:       d.Size{Width: 700, Height: 400},
		Layout:        d.VBox{},
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		Children: []d.Widget{
			d.Label{
				Text: fmt.Sprintf("Обнаружено изменений: %d. Отметьте параметры для исключения из записи:", len(changes)),
				Font: d.Font{Bold: true},
			},
			d.TableView{
				AssignTo:         &tv,
				Model:            model,
				CheckBoxes:       true, // Разрешаем чекбоксы для выбора
				AlternatingRowBG: true,
				Columns: []d.TableViewColumn{
					{Title: "Категория", Width: 120},
					{Title: "Параметр", Width: 180},
					{Title: "Было", Width: 130},
					{Title: "Стало", Width: 130},
				},
			},
			d.Composite{
				Layout: d.HBox{},
				Children: []d.Widget{
					d.HSpacer{},
					d.PushButton{
						AssignTo: &acceptPB,
						Text:     "Записать",
						OnClicked: func() {
							dlg.Accept()
						},
					},
					d.PushButton{
						AssignTo:  &cancelPB,
						Text:      "Отмена",
						OnClicked: func() { dlg.Cancel() },
					},
				},
			},
		},
	}.Create(owner)

	if err != nil {
		walk.MsgBox(owner, "Ошибка", err.Error(), walk.MsgBoxIconError)
		return nil, false
	}

	if dlg.Run() == walk.DlgCmdOK {
		// Фильтруем список, оставляем только неотмеченные (не удаленные)
		var confirmed []*models.ChangeItem
		for _, item := range model.Items {
			if !item.Remove {
				confirmed = append(confirmed, item)
			}
		}
		return confirmed, true
	}

	return nil, false
}
