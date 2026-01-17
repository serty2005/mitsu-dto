package view

import (
	"fmt"
	"mitsuscanner/internal/ui/controller"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
)

// MainWindowView отвечает за отображение главного окна приложения и взаимодействие с пользователем.
// View слой следуя Clean Architecture должен только заниматься рендерингом и обработкой пользовательских событий.
type MainWindowView struct {
	mw               *walk.MainWindow
	mainCtrl         *controller.MainController
	serviceCtrl      *controller.ServiceController
	registrationCtrl *controller.RegistrationController
	addrCombo        *walk.ComboBox
	actionBtn        *walk.PushButton
	clearProfilesBtn *walk.PushButton
	kktInfoComposite *walk.Composite
	modelLabel       *walk.Label
	serialLabel      *walk.Label
	unsentDocsLabel  *walk.Label
	rebootIndicator  *walk.Label
	logView          *walk.TextEdit
	logGroupBox      *walk.GroupBox
	collapsedLogComp *walk.Composite
	logPreviewLabel  *walk.Label
	isLogExpanded    bool
}

// NewMainWindowView создает новый экземпляр MainWindowView с переданными контроллерами.
func NewMainWindowView(mainCtrl *controller.MainController, serviceCtrl *controller.ServiceController, registrationCtrl *controller.RegistrationController) *MainWindowView {
	return &MainWindowView{
		mainCtrl:         mainCtrl,
		serviceCtrl:      serviceCtrl,
		registrationCtrl: registrationCtrl,
		isLogExpanded:    true,
	}
}

// Create создает и инициализирует главное окно приложения.
func (w *MainWindowView) Create() error {
	// Устанавливаем callback для обновления UI
	w.mainCtrl.SetOnUpdate(w.updateUI)

	// Подготавливаем начальные данные

	// Создаем главное окно
	err := d.MainWindow{
		AssignTo: &w.mw,
		Title:    "Mitsu Scanner",
		Size:     d.Size{Width: 400, Height: 600},
		MinSize:  d.Size{Width: 400, Height: 600},
		MaxSize:  d.Size{Width: 400, Height: 600},
		Layout:   d.VBox{MarginsZero: true, Spacing: 5},
		Children: []d.Widget{
			// --- Верхняя панель (Подключение + Инфо) ---
			d.GroupBox{
				Layout: d.HBox{Margins: d.Margins{Left: 5, Top: 5, Right: 5, Bottom: 5}, Spacing: 10},
				Children: []d.Widget{
					// ЛЕВАЯ ЧАСТЬ: Управление подключением
					d.Composite{
						Layout: d.HBox{MarginsZero: true, Spacing: 5},
						Children: []d.Widget{
							d.ComboBox{
								AssignTo:              &w.addrCombo,
								Editable:              true,
								Value:                 d.Bind("ConnectionString"),
								MinSize:               d.Size{Width: 220, Height: 0},
								ToolTipText:           "Введите COMx:Baud или IP:Port. Примеры: COM9:115200, 192.168.1.50:8200",
								OnCurrentIndexChanged: w.onDeviceSelectionChanged,
								OnTextChanged:         w.onDeviceTextChanged,
							},
							d.PushButton{
								AssignTo:  &w.actionBtn,
								Text:      "Подключить",
								OnClicked: w.onActionBtnClicked,
								MinSize:   d.Size{Width: 80},
							},
							d.PushButton{
								AssignTo:    &w.clearProfilesBtn,
								Text:        "🗑️",
								MaxSize:     d.Size{Width: 30},
								ToolTipText: "Очистить сохранённые профили",
							},
						},
					},
					// ПРАВАЯ ЧАСТЬ: Инфо о ККТ (Model, SN, Reboot status)
					d.Composite{
						AssignTo: &w.kktInfoComposite,
						Visible:  false, // Скрыт до подключения
						Layout:   d.HBox{MarginsZero: true, Spacing: 8, Alignment: d.AlignHNearVCenter},
						Children: []d.Widget{
							d.Label{AssignTo: &w.modelLabel, Text: "Mitsu", Font: d.Font{Bold: true}},
							d.Label{AssignTo: &w.serialLabel, Text: "SN: ..."},
							d.Label{AssignTo: &w.unsentDocsLabel, Text: "ОФД: 0"},
							d.Label{Text: "|"},
							d.Label{
								AssignTo:    &w.rebootIndicator,
								Text:        "⦿", // Кружок
								Font:        d.Font{PointSize: 14, Bold: true},
								TextColor:   walk.RGB(0, 200, 0), // Зеленый
								ToolTipText: "ON: Норма (Флаг=1)\nOFF: Был сбой питания (Флаг=0)",
							},
						},
					},
				},
			},
			// --- Вкладки ---
			d.TabWidget{
				MinSize: d.Size{Height: 500},
				MaxSize: d.Size{Height: 500},
				Pages: []d.TabPage{
					// 1. Информация
					{
						Title:  "Информация",
						Layout: d.VBox{Margins: d.Margins{Left: 6, Top: 6, Right: 6, Bottom: 6}, Spacing: 5},
						Children: []d.Widget{
							d.PushButton{Text: "Обновить данные"},
							d.TextEdit{
								AssignTo: &w.logView,
								ReadOnly: true,
								VScroll:  true,
								Font:     d.Font{Family: "Consolas", PointSize: 9},
								MinSize:  d.Size{Height: 400},
								MaxSize:  d.Size{Height: 400},
							},
						},
					},
					// 2. Регистрация
					NewRegistrationTab(w.registrationCtrl, w.mw).Create(),
					// 3. Сервис
					NewServiceTab(w.serviceCtrl, w.mw).Create(),
				},
			},
			// --- Лог (Сворачиваемый) ---
			d.Composite{
				Layout: d.VBox{MarginsZero: true},
				Children: []d.Widget{
					// Развернутый вид
					d.GroupBox{
						AssignTo: &w.logGroupBox,
						Title:    "Лог",
						Layout:   d.VBox{MarginsZero: true},
						MinSize:  d.Size{Height: 150},
						MaxSize:  d.Size{Height: 150},
						Children: []d.Widget{
							d.Composite{
								Layout: d.HBox{MarginsZero: true},
								Children: []d.Widget{
									d.PushButton{Text: "🔽 Свернуть", OnClicked: w.toggleLog, MaxSize: d.Size{Width: 80}},
								},
							},
							d.TextEdit{
								AssignTo: &w.logView,
								ReadOnly: true,
								VScroll:  true,
								HScroll:  true,
							},
						},
					},
					// Свернутый вид
					d.Composite{
						AssignTo: &w.collapsedLogComp,
						Visible:  false, // Скрыт по умолчанию
						Layout:   d.HBox{Margins: d.Margins{Left: 5, Top: 2, Right: 5, Bottom: 2}},
						Children: []d.Widget{
							d.PushButton{Text: "🔼 Лог", OnClicked: w.toggleLog, MaxSize: d.Size{Width: 60}},
							d.Label{
								AssignTo:      &w.logPreviewLabel,
								Text:          "...",
								TextAlignment: d.AlignNear,
								EllipsisMode:  d.EllipsisEnd,
								MaxSize:       d.Size{Width: 550},
							},
						},
					},
				},
			},
		},
	}.Create()

	if err != nil {
		return err
	}

	w.mainCtrl.Initialize()

	// Подключаем обработчик закрытия окна
	w.mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if w.mainCtrl != nil {
			// Попытка разорвать соединение при закрытии окна
			_ = w.mainCtrl.Disconnect()
		}
	})

	return nil
}

// Run запускает главный цикл обработки сообщений окна.
func (w *MainWindowView) Run() {
	w.mw.Run()
}

// updateUI обновляет состояние интерфейса в зависимости от данных из ViewModel.
func (w *MainWindowView) updateUI() {
	w.mw.Synchronize(func() {
		vm := w.mainCtrl.ViewModel()

		if len(vm.ConnectionList) > 0 {
			// walk требует, чтобы модель была []string для ComboBox, если мы хотим просто строки
			// Но SetModel принимает interface{}.
			// Важный нюанс: если мы меняем модель, мы можем потерять текущий текст.
			// Поэтому сохраняем и восстанавливаем.
			currentText := w.addrCombo.Text()
			w.addrCombo.SetModel(vm.ConnectionList)
			// Если текст был пустой (первый запуск), ставим первый элемент
			if currentText == "" && len(vm.ConnectionList) > 0 {
				w.addrCombo.SetText(vm.ConnectionList[0])
			} else {
				w.addrCombo.SetText(currentText)
			}
		}
		// Обновляем состояние элементов управления
		w.actionBtn.SetText(vm.ActionButtonText)
		w.actionBtn.SetEnabled(vm.ActionButtonEnabled)
		w.addrCombo.SetEnabled(vm.ConnectionStringEnabled)
		w.clearProfilesBtn.SetEnabled(vm.ClearProfilesButtonEnabled)
		w.kktInfoComposite.SetVisible(vm.KKTInfoVisible)

		// Обновляем информацию о ККТ
		w.modelLabel.SetText(vm.ModelName)
		w.serialLabel.SetText("SN: " + vm.SerialNumber)
		w.unsentDocsLabel.SetText("ОФД: " + fmt.Sprintf("%d", vm.UnsentDocsCount))

		// Обновляем индикатор питания
		if vm.PowerFlag {
			w.rebootIndicator.SetText("⦿")
			w.rebootIndicator.SetTextColor(walk.RGB(0, 200, 0)) // Зеленый
			w.rebootIndicator.SetToolTipText("Питание в норме")
		} else {
			w.rebootIndicator.SetText("○")
			w.rebootIndicator.SetTextColor(walk.RGB(255, 0, 0)) // Красный
			w.rebootIndicator.SetToolTipText("ВНИМАНИЕ: Произошла перезагрузка ККТ!")
		}
	})
}

// onActionBtnClicked обрабатывает событие нажатия на кнопку действия.
func (w *MainWindowView) onActionBtnClicked() {
	// 1. Подготовка данных
	w.syncConnectionString()
	vm := w.mainCtrl.ViewModel()

	// 2. Логика действия
	if vm.IsConnected {
		if err := w.mainCtrl.Disconnect(); err != nil {
			walk.MsgBox(w.mw, "Ошибка", err.Error(), walk.MsgBoxIconError)
		}
	} else if vm.ActionButtonText == "Искать" {
		if err := w.mainCtrl.SearchDevice(); err != nil {
			walk.MsgBox(w.mw, "Ошибка", err.Error(), walk.MsgBoxIconError)
		}
	} else {
		// Подключение
		if err := w.mainCtrl.Connect(); err != nil {
			walk.MsgBox(w.mw, "Ошибка", err.Error(), walk.MsgBoxIconError)
		}
	}
}

// toggleLog переключает видимость лога.
func (w *MainWindowView) toggleLog() {
	w.isLogExpanded = !w.isLogExpanded

	w.mw.SetSuspended(true)
	defer w.mw.SetSuspended(false)

	if w.isLogExpanded {
		w.logGroupBox.SetVisible(true)
		w.collapsedLogComp.SetVisible(false)
	} else {
		w.logGroupBox.SetVisible(false)
		w.collapsedLogComp.SetVisible(true)
	}
}

func (w *MainWindowView) onDeviceSelectionChanged() {
	// При выборе из списка сразу синхронизируем текст в VM
	w.syncConnectionString()
}

func (w *MainWindowView) onDeviceTextChanged() {
	// При ручном вводе тоже синхронизируем
	w.syncConnectionString()
}

// syncConnectionString вручную переносит текст из виджета в ViewModel.
// Это надежнее, чем DataBinder().Submit() при каждом чихе.
func (w *MainWindowView) syncConnectionString() {
	text := w.addrCombo.Text()
	w.mainCtrl.ViewModel().ConnectionString = text

	// Теперь, когда в VM актуальный текст, обновляем состояние кнопок (Искать/Подключить)
	w.mainCtrl.ViewModel().UpdateUIState()
	w.updateUI()
}

func (w *MainWindowView) onClearProfiles() {
	if walk.MsgBox(w.mw, "Подтверждение", "Очистить все профили?", walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) != walk.DlgCmdYes {
		return
	}
	// Вызываем контроллер (он сам обновит список и UI)
	if err := w.mainCtrl.ClearProfiles(); err != nil {
		walk.MsgBox(w.mw, "Ошибка", err.Error(), walk.MsgBoxIconError)
	}
}
