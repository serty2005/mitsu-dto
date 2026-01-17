package view

import (
	"mitsuscanner/internal/domain/models"
	"mitsuscanner/internal/ui/controller"
	"mitsuscanner/internal/ui/view/dialogs"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// RegistrationTab представляет вкладку "Регистрация" с управляющим контроллером
type RegistrationTab struct {
	controller   *controller.RegistrationController
	owner        walk.Form
	fnPhaseLabel *walk.Label
}

// NewRegistrationTab создает новый экземпляр вкладки "Регистрация"
func NewRegistrationTab(controller *controller.RegistrationController, owner walk.Form) *RegistrationTab {
	return &RegistrationTab{
		controller: controller,
		owner:      owner,
	}
}

// Create возвращает компонент вкладки с UI
func (t *RegistrationTab) Create() TabPage {
	var (
		vm = t.controller.ViewModel()
	)

	return TabPage{
		Title:  "Регистрация",
		Layout: VBox{Margins: Margins{Left: 2, Top: 2, Right: 2, Bottom: 2}, Spacing: 3},
		DataBinder: DataBinder{
			Name:       "registrationVM",
			DataSource: vm,
		},
		Children: []Widget{
			// Верхняя панель
			Composite{
				Layout: Grid{Columns: 3, Margins: Margins{Left: 2, Top: 2, Right: 2, Bottom: 2}, Spacing: 3},
				Children: []Widget{
					Label{Text: "Регистрационный номер ККТ (РНМ):"},
					LineEdit{Text: Bind("RNM"), MinSize: Size{Width: 150}},
					PushButton{Text: "Вычислить (CRC)", OnClicked: t.onCalculateRNM},
				},
			},

			// Основной контент
			Composite{
				Layout: VBox{Margins: Margins{Left: 2, Top: 2, Right: 2, Bottom: 2}, Spacing: 3},
				Children: []Widget{
					Composite{
						Layout: HBox{Margins: Margins{Left: 2, Top: 2, Right: 2, Bottom: 2}, Spacing: 3},
						Children: []Widget{
							GroupBox{
								Title:  "Реквизиты организации",
								Layout: Grid{Columns: 2, Margins: Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 5},
								Children: []Widget{
									Label{Text: "Наименование:"},
									LineEdit{Text: Bind("OrgName")},
									Label{Text: "ИНН:"},
									LineEdit{Text: Bind("INN")},
									Label{Text: "Адрес расчетов:"},
									LineEdit{Text: Bind("Address")},
									Label{Text: "Место расчетов:"},
									LineEdit{Text: Bind("Place")},
									Label{Text: "E-mail отправителя:"},
									LineEdit{Text: Bind("Email")},
									Label{Text: "Сайт ФНС:"},
									LineEdit{Text: Bind("Site")},
								},
							},
							GroupBox{
								Title:   "Системы налогообложения",
								MinSize: Size{Width: 150},
								Layout:  VBox{Margins: Margins{Left: 3, Top: 1, Right: 3, Bottom: 1}, Spacing: 1},
								Children: []Widget{
									CheckBox{Text: "ОСН", Checked: Bind("TaxOSN")},
									CheckBox{Text: "УСН доход", Checked: Bind("TaxUSN")},
									CheckBox{Text: "УСН доход - расход", Checked: Bind("TaxUSN_M")},
									CheckBox{Text: "ЕСХН", Checked: Bind("TaxESHN")},
									CheckBox{Text: "Патент", Checked: Bind("TaxPat")},
									Label{Text: "СНО по умолчанию:", Font: Font{PointSize: 7}},
									ComboBox{
										Value:         Bind("TaxSystemBase"),
										BindingMember: "Code",
										DisplayMember: "Name",
										Model: []*NV{
											{Name: "", Code: ""},
											{Name: "ОСН", Code: "0"},
											{Name: "УСН доход", Code: "1"},
											{Name: "УСН доход - расход", Code: "2"},
											{Name: "ЕСХН", Code: "4"},
											{Name: "Патент", Code: "5"},
										},
									},
								},
							},
						},
					},
					Composite{
						Layout: HBox{Margins: Margins{Left: 3, Top: 1, Right: 3, Bottom: 1}, Spacing: 5},
						Children: []Widget{
							GroupBox{
								Title:  "Фискальные признаки",
								Layout: Grid{Columns: 3, Margins: Margins{Left: 3, Top: 1, Right: 3, Bottom: 1}, Spacing: 1},
								Children: []Widget{
									// Колонка 1
									CheckBox{Text: "Автономный режим", Checked: Bind("ModeAutonomous")},
									CheckBox{Text: "Шифрование данных", Checked: Bind("ModeEncryption")},
									CheckBox{Text: "Расчеты за услуги", Checked: Bind("ModeService")},
									CheckBox{Text: "Расчеты в Интернет", Checked: Bind("ModeInternet")},
									CheckBox{Text: "Принтер в автомате", Checked: Bind("ModeAutomat")},
									// Колонка 2
									CheckBox{Text: "Только БСО", Checked: Bind("ModeBSO")},
									CheckBox{Text: "Подакцизные товары", Checked: Bind("ModeExcise")},
									CheckBox{Text: "Проведение азартных игр", Checked: Bind("ModeGambling")},
									CheckBox{Text: "Проведение лотерей", Checked: Bind("ModeLottery")},
									CheckBox{Text: "Маркированные товары", Checked: Bind("ModeMarking")},
									// Колонка 3
									CheckBox{Text: "Ломбард", Checked: Bind("ModePawn")},
									CheckBox{Text: "Страхование", Checked: Bind("ModeInsurance")},
									CheckBox{Text: "Общепит", Checked: Bind("ModeCatering")},
									CheckBox{Text: "Оптовая торговля", Checked: Bind("ModeWholesale")},
									CheckBox{Text: "Вендинг", Checked: Bind("ModeVending")},
								},
							},
							GroupBox{
								Title:   "Информация о ФН",
								Layout:  VBox{MarginsZero: true, Spacing: 3},
								MinSize: Size{Width: 200},
								Children: []Widget{
									Composite{
										Layout: Grid{Columns: 2, Spacing: 5},
										Children: []Widget{
											Label{Text: "№:"},
											Label{Text: Bind("FnNumber"), Font: Font{Bold: true}},
											Label{Text: "Фаза:"},
											Label{AssignTo: &t.fnPhaseLabel, Text: Bind("FnPhaseText")},
											Label{Text: "До:"},
											Label{Text: Bind("FnValidDate"), Font: Font{Bold: true}},
										},
									},
									Composite{
										Layout: HBox{MarginsZero: true, Spacing: 5, Alignment: AlignHCenterVCenter},
										Children: []Widget{
											PushButton{
												Text:        "Отправить в ОФД",
												MinSize:     Size{Width: 110},
												ToolTipText: "Отправить первый неотправленный документ в ОФД",
											},
											PushButton{
												Text:        "↻",
												OnClicked:   t.onRefreshFnInfo,
												MaxSize:     Size{Width: 30, Height: 30},
												ToolTipText: "Обновить информацию о ФН",
											},
										},
									},
								},
							},
						},
					},
				},
			},

			// Блок: Причины + ОФД
			Composite{
				Layout: HBox{Margins: Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}, Spacing: 3},
				Children: []Widget{
					GroupBox{
						Title:         "Оператор фискальных данных",
						StretchFactor: 2,
						Layout:        VBox{Margins: Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}, Spacing: 3},
						Children: []Widget{
							Composite{
								Layout: HBox{Margins: Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}, Spacing: 3},
								Children: []Widget{
									Label{Text: "ИНН ОФД:", TextAlignment: AlignFar},
									LineEdit{Text: Bind("OFDINN")},
								},
							},
							Composite{
								Layout: HBox{Margins: Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}, Spacing: 3},
								Children: []Widget{
									Label{Text: "Наименование ОФД:", TextAlignment: AlignFar},
									LineEdit{Text: Bind("OFDName")},
								},
							},
						},
					},
					GroupBox{
						Title:         "Причины перерегистрации",
						StretchFactor: 1,
						Layout:        VBox{Margins: Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 10},
						Children: []Widget{
							LineEdit{Text: Bind("Reasons"), ReadOnly: true},
							Composite{
								Layout: HBox{MarginsZero: true, Alignment: AlignHCenterVCenter},
								Children: []Widget{
									PushButton{
										Text:      "Выбрать...",
										OnClicked: t.onSelectReasons,
										MinSize:   Size{Width: 100},
										MaxSize:   Size{Width: 100},
									},
								},
							},
						},
					},
				},
			},

			// Кнопки
			Composite{
				Layout: HBox{Margins: Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 5},
				Children: []Widget{
					PushButton{Text: "Считать из ККТ", OnClicked: t.onReadFromDevice},
					HSpacer{},
					PushButton{Text: "Закрытие ФН", OnClicked: t.onCloseFn},
					PushButton{Text: "Замена ФН", OnClicked: t.onReplaceFn},
					PushButton{Text: "Перерегистрация", OnClicked: t.onReregister},
					PushButton{Text: "Регистрация", OnClicked: t.onRegister},
				},
			},
		},
	}
}

// onCalculateRNM обрабатывает событие расчета РНМ
func (t *RegistrationTab) onCalculateRNM() {
	// TODO: Реализовать диалог для расчета РНМ
	walk.MsgBox(t.owner, "Информация", "Функция расчета РНМ будет реализована в следующей версии", walk.MsgBoxIconInformation)
}

// onReadFromDevice обрабатывает событие чтения данных из ККТ
func (t *RegistrationTab) onReadFromDevice() {
	if err := t.controller.ReadFromDevice(); err != nil {
		walk.MsgBox(t.owner, "Ошибка", err.Error(), walk.MsgBoxIconError)
		return
	}

	// Обновляем цвет фазы ФН
	t.updateFnPhaseColor()
}

// onSelectReasons обрабатывает событие выбора причин перерегистрации
func (t *RegistrationTab) onSelectReasons() {
	reasons, ok := dialogs.ShowReasonsDialog(t.owner, t.controller.ViewModel().Reasons)
	if ok {
		t.controller.ViewModel().Reasons = reasons
	}
}

// onRegister обрабатывает событие регистрации
func (t *RegistrationTab) onRegister() {
	resp, err := t.controller.Register()
	if err != nil {
		walk.MsgBox(t.owner, "Ошибка", err.Error(), walk.MsgBoxIconError)
		return
	}

	// TODO: Получить полные данные для отчета и показать его
	dialogs.ShowReportDialog(t.owner, "Расшифрованные данные регистрации", resp, models.ReportKindRegistration)
}

// onReregister обрабатывает событие перерегистрации
func (t *RegistrationTab) onReregister() {
	resp, err := t.controller.Reregister()
	if err != nil {
		walk.MsgBox(t.owner, "Ошибка", err.Error(), walk.MsgBoxIconError)
		return
	}

	// TODO: Получить полные данные для отчета и показать его
	dialogs.ShowReportDialog(t.owner, "Расшифрованные данные перерегистрации", resp, models.ReportKindRegistration)
}

// onCloseFn обрабатывает событие закрытия ФН
func (t *RegistrationTab) onCloseFn() {
	if walk.MsgBox(t.owner, "Подтверждение", "Вы действительно хотите закрыть фискальный архив?\nЭто необратимая операция!", walk.MsgBoxYesNo|walk.MsgBoxIconWarning) != walk.DlgCmdYes {
		return
	}

	data, err := t.controller.CloseFn()
	if err != nil {
		walk.MsgBox(t.owner, "Ошибка", err.Error(), walk.MsgBoxIconError)
		return
	}

	dialogs.ShowReportDialog(t.owner, "Отчет о закрытии фискального накопителя", data, models.ReportKindCloseFn)
}

// onReplaceFn обрабатывает событие замены ФН
func (t *RegistrationTab) onReplaceFn() {
	// Упрощенный флоу замены ФН - используем причину №1
	t.controller.ViewModel().Reasons = "1"

	resp, err := t.controller.Reregister()
	if err != nil {
		walk.MsgBox(t.owner, "Ошибка", err.Error(), walk.MsgBoxIconError)
		return
	}

	// TODO: Получить полные данные для отчета и показать его
	dialogs.ShowReportDialog(t.owner, "Расшифрованные данные перерегистрации", resp, models.ReportKindRegistration)
}

// onRefreshFnInfo обрабатывает событие обновления информации о ФН
func (t *RegistrationTab) onRefreshFnInfo() {
	if err := t.controller.RefreshFnInfo(); err != nil {
		walk.MsgBox(t.owner, "Ошибка", err.Error(), walk.MsgBoxIconError)
		return
	}

	// Обновляем цвет фазы ФН
	t.updateFnPhaseColor()
}

// updateFnPhaseColor обновляет цвет текста для фазы ФН
func (t *RegistrationTab) updateFnPhaseColor() {
	vm := t.controller.ViewModel()
	if t.fnPhaseLabel != nil {
		// Конвертируем hex-строку цвета в walk.Color
		switch vm.FnPhaseColor {
		case "#0000FF": // Синий
			t.fnPhaseLabel.SetTextColor(walk.RGB(0, 0, 255))
		case "#008000": // Зелёный
			t.fnPhaseLabel.SetTextColor(walk.RGB(0, 128, 0))
		case "#FF0000": // Красный
			t.fnPhaseLabel.SetTextColor(walk.RGB(255, 0, 0))
		case "#808080": // Серый
			t.fnPhaseLabel.SetTextColor(walk.RGB(128, 128, 128))
		default:
			t.fnPhaseLabel.SetTextColor(walk.RGB(0, 0, 0)) // Чёрный
		}
	}
}

// NV представляет пару имя-значение для ComboBox
type NV struct {
	Name string
	Code string
}
