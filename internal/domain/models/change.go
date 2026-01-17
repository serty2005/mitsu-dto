package models

// Priority определяет порядок применения настроек.
type Priority int

const (
	PriorityNormal   Priority = 0 // Обычные настройки (Опции, Принтер)
	PriorityCliche   Priority = 1 // Клише (много данных, лучше отдельно)
	PriorityNetwork  Priority = 2 // Сеть (может разорвать соединение, строго в конце)
	PriorityCritical Priority = 3 // Критические операции (если появятся)
)

// Change представляет одно атомарное (или групповое) изменение настроек.
type Change struct {
	ID          string      // Уникальный ID поля (для подсветки в GUI)
	Description string      // Человекочитаемое описание изменения
	OldValue    interface{} // Значение "Было" (для отображения)
	NewValue    interface{} // Значение "Стало" (для отображения)
	Priority    Priority    // Приоритет выполнения
}
