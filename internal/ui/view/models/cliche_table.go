package models

import (
	"mitsuscanner/internal/ui/viewmodel"

	"github.com/lxn/walk"
)

// ClicheTableModel реализует интерфейс walk.TableModel для отображения данных клише.
type ClicheTableModel struct {
	walk.TableModelBase
	items []*viewmodel.ClicheItemWrapper
}

// NewClicheTableModel создаёт новый экземпляр ClicheTableModel.
func NewClicheTableModel(items []*viewmodel.ClicheItemWrapper) *ClicheTableModel {
	return &ClicheTableModel{
		items: items,
	}
}

// RowCount возвращает количество строк в таблице.
func (m *ClicheTableModel) RowCount() int {
	return len(m.items)
}

// Value возвращает значение для отображения в заданной ячейке.
func (m *ClicheTableModel) Value(row, col int) interface{} {
	if row < 0 || row >= len(m.items) {
		return nil
	}

	item := m.items[row]

	switch col {
	case 0: // Строка клише (Index)
		return item.Index + 1 // Отображение номеров строк с 1
	case 1: // Формат
		return item.Format()
	case 2: // Текст
		return item.Text()
	default:
		return nil
	}
}

// ItemChanged реализует метод интерфейса walk.ListModel
func (m *ClicheTableModel) ItemChanged(index int) {
	if index >= 0 && index < len(m.items) {
		m.PublishRowChanged(index)
	}
}

// UpdateItems обновляет данные в модели и уведомляет об изменениях.
func (m *ClicheTableModel) UpdateItems(items []*viewmodel.ClicheItemWrapper) {
	m.items = items
	m.PublishRowsReset()
}

// UpdateRow уведомляет об изменении конкретной строки.
func (m *ClicheTableModel) UpdateRow(row int) {
	if row >= 0 && row < len(m.items) {
		m.PublishRowChanged(row)
	}
}

// Items возвращает ссылку на массив элементов для прямого доступа.
func (m *ClicheTableModel) Items() []*viewmodel.ClicheItemWrapper {
	return m.items
}

// ItemAtIndex возвращает элемент по индексу строки.
func (m *ClicheTableModel) ItemAtIndex(row int) *viewmodel.ClicheItemWrapper {
	if row < 0 || row >= len(m.items) {
		return nil
	}
	return m.items[row]
}
