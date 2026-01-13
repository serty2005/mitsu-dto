package gui

import (
	"mitsuscanner/internal/service"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
)

// RunDiffDialog открывает диалог подтверждения изменений.
func RunDiffDialog(owner walk.Form, changes []service.Change) (bool, []service.Change) {
	var dlg *walk.Dialog
	var tv *walk.TableView
	var acceptPB *walk.PushButton

	// Копия изменений
	localChanges := make([]service.Change, len(changes))
	copy(localChanges, changes)

	model := NewDiffModel(localChanges)

	err := d.Dialog{
		AssignTo:      &dlg,
		Title:         "Подтверждение изменений",
		MinSize:       d.Size{Width: 650, Height: 400},
		Layout:        d.VBox{},
		DefaultButton: &acceptPB,
		Children: []d.Widget{
			d.Label{Text: "Следующие настройки будут записаны в ККТ:"},
			d.TableView{
				AssignTo:         &tv,
				AlternatingRowBG: true,
				Model:            model,
				Columns: []d.TableViewColumn{
					{Title: "Параметр", Width: 220},
					{Title: "Было", Width: 180},
					{Title: "Стало", Width: 180},
				},
				// Контекстное меню убрали, как просили
			},
			d.Composite{
				Layout: d.HBox{},
				Children: []d.Widget{
					d.Label{Text: "* Правый клик по строке для отмены (удаления) изменения", Font: d.Font{PointSize: 8}, TextColor: walk.RGB(100, 100, 100)},
					d.HSpacer{},
					d.PushButton{
						AssignTo:  &acceptPB,
						Text:      "Применить",
						OnClicked: func() { dlg.Accept() },
					},
					d.PushButton{
						Text:      "Отмена",
						OnClicked: func() { dlg.Cancel() },
					},
				},
			},
		},
	}.Create(owner)

	if err != nil {
		walk.MsgBox(owner, "Ошибка", err.Error(), walk.MsgBoxIconError)
		return false, nil
	}

	// ХАК: Обработка правого клика вручную для удаления
	// walk.TableView не имеет OnItemRightClicked в декларативном виде,
	// используем MouseUp
	tv.MouseUp().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.MouseButton. {
			// Определяем индекс строки под курсором
			// К сожалению, walk не дает простого способа получить строку по координатам в public API без WinAPI.
			// НО: Обычно правый клик выделяет строку. Попробуем взять CurrentIndex.
			idx := tv.CurrentIndex()
			if idx >= 0 {
				model.RemoveItem(idx)
			}
		}
	})

	if dlg.Run() == walk.DlgCmdOK {
		// Возвращаем true и список (даже если он пуст)
		return true, model.items
	}

	return false, nil
}

// DiffModel модель для таблицы изменений
type DiffModel struct {
	walk.TableModelBase
	items []service.Change
}

func NewDiffModel(items []service.Change) *DiffModel {
	return &DiffModel{items: items}
}

func (m *DiffModel) RowCount() int {
	return len(m.items)
}

func (m *DiffModel) Value(row, col int) interface{} {
	if row >= len(m.items) {
		return ""
	}
	item := m.items[row]
	switch col {
	case 0:
		return item.Description
	case 1:
		return item.OldValue
	case 2:
		return item.NewValue
	}
	return ""
}

func (m *DiffModel) RemoveItem(index int) {
	if index < 0 || index >= len(m.items) {
		return
	}
	m.items = append(m.items[:index], m.items[index+1:]...)
	m.PublishRowsReset()
}
