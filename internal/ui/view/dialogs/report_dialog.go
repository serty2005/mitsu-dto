package dialogs

import (
	"os"
	"path/filepath"

	"mitsuscanner/internal/domain/models"
	"mitsuscanner/internal/ui/view/utils"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
)

// ShowReportDialog открывает модальное окно для показа текста отчета
func ShowReportDialog(owner walk.Form, title string, data interface{}, kind models.ReportKind) {
	var dlg *walk.Dialog
	var copyPB, savePB, closePB *walk.PushButton

	// Формируем текст отчета
	lines, err := utils.BuildReportLines(data, kind)
	if err != nil {
		walk.MsgBox(owner, "Ошибка", err.Error(), walk.MsgBoxIconError)
		return
	}
	text := utils.FormatKeyValueText(lines)

	err = d.Dialog{
		AssignTo:      &dlg,
		Title:         title,
		MinSize:       d.Size{Width: 500, Height: 400},
		MaxSize:       d.Size{Width: 500, Height: 400},
		Layout:        d.VBox{},
		DefaultButton: &copyPB,
		CancelButton:  &closePB,
		Children: []d.Widget{
			d.TextEdit{
				Text:     utils.ToWindowsText(text),
				ReadOnly: true,
				VScroll:  true,
				Font:     d.Font{Family: "Consolas", PointSize: 9},
			},
			d.Composite{
				Layout: d.HBox{Spacing: 6},
				Children: []d.Widget{
					d.HSpacer{},
					d.PushButton{
						AssignTo: &copyPB,
						Text:     "Копировать",
						OnClicked: func() {
							_ = walk.Clipboard().SetText(text)
						},
					},
					d.PushButton{
						AssignTo: &savePB,
						Text:     "Сохранить...",
						OnClicked: func() {
							saveReportWithDialog(dlg, data, text)
						},
					},
					d.PushButton{
						AssignTo: &closePB,
						Text:     "Закрыть",
						OnClicked: func() {
							dlg.Accept()
						},
					},
				},
			},
		},
	}.Create(owner)

	if err != nil {
		walk.MsgBox(owner, "Ошибка", err.Error(), walk.MsgBoxIconError)
		return
	}

	dlg.Run()
}

// saveReportWithDialog открывает системный диалог сохранения файла
func saveReportWithDialog(owner walk.Form, data interface{}, text string) {
	dlg := new(walk.FileDialog)

	// Формируем имя файла по умолчанию
	defaultName := utils.GenerateReportFileName(data)
	dlg.FilePath = defaultName
	dlg.Filter = "Text Files (*.txt)|*.txt|All Files (*.*)|*.*"
	dlg.Title = "Сохранить отчет"

	// Пытаемся открыть в "Документах"
	if home, err := os.UserHomeDir(); err == nil {
		dlg.InitialDirPath = filepath.Join(home, "Documents")
	}

	if ok, _ := dlg.ShowSave(owner); ok {
		if err := os.WriteFile(dlg.FilePath, []byte(text), 0644); err != nil {
			walk.MsgBox(owner, "Ошибка", "Не удалось сохранить файл:\n"+err.Error(), walk.MsgBoxIconError)
		} else {
			walk.MsgBox(owner, "Успех", "Файл успешно сохранен.", walk.MsgBoxIconInformation)
		}
	}
}
