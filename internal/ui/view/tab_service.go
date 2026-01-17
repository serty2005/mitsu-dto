package view

import (
	"fmt"

	"mitsuscanner/internal/ui/controller"
	"mitsuscanner/internal/ui/view/models"
	"mitsuscanner/internal/ui/viewmodel"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// ServiceTab представляет вкладку "Сервис" с управляющим контроллером
type ServiceTab struct {
	controller *controller.ServiceController
	owner      walk.Form
}

// NewServiceTab создает новый экземпляр вкладки "Сервис"
func NewServiceTab(controller *controller.ServiceController, owner walk.Form) *ServiceTab {
	return &ServiceTab{
		controller: controller,
		owner:      owner,
	}
}

// Create возвращает компонент вкладки с UI
func (t *ServiceTab) Create() TabPage {
	var (
		clicheTable        *walk.TableView
		clicheModel        *models.ClicheTableModel
		readBtn            *walk.PushButton
		writeBtn           *walk.PushButton
		kktTimeLabel       *walk.Label
		targetTimeEdit     *walk.LineEdit
		autoSyncCheck      *walk.CheckBox
		b9ComboBox         *walk.ComboBox
		clicheEditorGroup  *walk.GroupBox
		clicheEditorBinder *walk.DataBinder
		ceText             *walk.LineEdit
		ceAlign            *walk.ComboBox
		ceFont             *walk.ComboBox
		ceUnder            *walk.ComboBox
		ceWidth            *walk.NumberEdit
		ceHeight           *walk.NumberEdit
		ceInvert           *walk.CheckBox

		// Модель данных
		vm = t.controller.ViewModel()
	)

	clicheModel = models.NewClicheTableModel(vm.ClicheItems)

	return TabPage{
		Title:  "Сервис",
		Layout: VBox{MarginsZero: true, Spacing: 5},
		DataBinder: DataBinder{
			Name:       "serviceVM",
			DataSource: vm,
		},
		Children: []Widget{
			// Верх: Время и Операции
			Composite{
				Layout: HBox{MarginsZero: true, Spacing: 6},
				Children: []Widget{
					GroupBox{
						Title: "Синхронизация времени", StretchFactor: 1,
						Layout: VBox{Margins: Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 4},
						Children: []Widget{
							Composite{
								Layout: Grid{Columns: 2, Spacing: 4},
								Children: []Widget{
									Label{Text: "Время ККТ:", Font: Font{PointSize: 8}},
									Label{
										AssignTo: &kktTimeLabel,
										Text:     Bind("KktTimeStr"),
										Font:     Font{PointSize: 9, Bold: true},
									},

									Label{Text: "Установить:", Font: Font{PointSize: 8}},
									LineEdit{
										AssignTo: &targetTimeEdit,
										Text:     Bind("TargetTimeStr"),
										ReadOnly: false,
									},
								},
							},
							Composite{
								Layout: HBox{MarginsZero: true},
								Children: []Widget{
									CheckBox{
										AssignTo: &autoSyncCheck,
										Text:     "Авто (ПК)",
										Checked:  Bind("AutoSyncPC"),
										OnCheckStateChanged: func() {
											if targetTimeEdit != nil {
												targetTimeEdit.SetReadOnly(autoSyncCheck.Checked())
											}
										},
									},
									HSpacer{},
									PushButton{Text: "Установить", OnClicked: func() {
										if err := t.controller.SyncTime(); err != nil {
											walk.MsgBox(t.owner, "Ошибка", err.Error(), walk.MsgBoxIconError)
											return
										}
										walk.MsgBox(t.owner, "Успех", "Время успешно синхронизировано", walk.MsgBoxIconInformation)
									}},
								},
							},
						},
					},
					GroupBox{
						Title: "Операции", StretchFactor: 1,
						Layout: Grid{Columns: 2, Margins: Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 4},
						Children: []Widget{
							PushButton{Text: "Прогон/Отрезка", MinSize: Size{Width: 90}},
							PushButton{Text: "Тех. сброс", MinSize: Size{Width: 90}},
							PushButton{Text: "Ден. ящик", MinSize: Size{Width: 90}},
							PushButton{Text: "X-отчёт", MinSize: Size{Width: 90}},
							PushButton{Text: "Сброс МГМ", MinSize: Size{Width: 90}},
							PushButton{
								AssignTo: &readBtn,
								Text:     "Считать",
								OnClicked: func() {
									if err := t.controller.ReadSettings(); err != nil {
										walk.MsgBox(t.owner, "Ошибка", err.Error(), walk.MsgBoxIconError)
										return
									}
									clicheModel.UpdateItems(vm.ClicheItems)
								},
								MinSize: Size{Width: 150},
							},
							PushButton{
								AssignTo: &writeBtn,
								Text:     "Записать",
								OnClicked: func() {
									if err := t.controller.WriteSettings(t.owner); err != nil {
										walk.MsgBox(t.owner, "Ошибка", err.Error(), walk.MsgBoxIconError)
										return
									}
									walk.MsgBox(t.owner, "Успех", "Настройки успешно записаны", walk.MsgBoxIconInformation)
								},
								MinSize: Size{Width: 150},
							},
						},
					},
				},
			},

			// Табы подкатегорий
			TabWidget{
				MinSize: Size{Height: 300},
				Pages: []TabPage{
					{
						Title:  "Параметры",
						Layout: VBox{MarginsZero: true, Spacing: 0, Alignment: AlignHNearVNear},
						Children: []Widget{
							Composite{
								Layout: HBox{Margins: Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}, Spacing: 4, Alignment: AlignHCenterVCenter},
								Children: []Widget{
									Composite{
										Layout: VBox{MarginsZero: true, Spacing: 4},
										Children: []Widget{
											GroupBox{
												Title:  "ОФД и ОИСМ",
												Layout: Grid{Columns: 4, Spacing: 4, Margins: Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
												Children: []Widget{
													Label{Text: "ОФД:"}, LineEdit{Text: Bind("OfdString"), MinSize: Size{Width: 130}, MaxSize: Size{Width: 130}},
													Label{Text: "Клиент:"}, ComboBox{Value: Bind("OfdClient"), BindingMember: "Code", DisplayMember: "Name", Model: viewmodel.ListOfdClients, MaxSize: Size{Width: 100}},
													Label{Text: "ОИСМ:"}, LineEdit{Text: Bind("OismString"), MinSize: Size{Width: 130}, MaxSize: Size{Width: 130}},
													Label{Text: "Т. ФН:"}, NumberEdit{Value: Bind("TimerFN"), MaxSize: Size{Width: 40}},
													Label{Text: "Пояс:"}, ComboBox{Value: Bind("OptTimezone"), BindingMember: "Code", DisplayMember: "Name", Model: viewmodel.ListTimezones, MinSize: Size{Width: 110}, MaxSize: Size{Width: 120}},
													Label{Text: "Т. ОФД:"}, NumberEdit{Value: Bind("TimerOFD"), MaxSize: Size{Width: 40}},
												},
											},
											GroupBox{
												Title:  "Принтер и Бумага",
												Layout: Grid{Columns: 6, Spacing: 2, Margins: Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
												Children: []Widget{
													Label{Text: "Модель:"}, ComboBox{Value: Bind("PrintModel"), BindingMember: "Code", DisplayMember: "Name", Model: viewmodel.ListPrinterModels, MaxSize: Size{Width: 70}},
													Label{Text: "Отрезчик:"}, CheckBox{Checked: Bind("OptCut")},
													Label{Text: "Ящик:", ToolTipText: "Условие автоматического открытия денежного ящика"}, ComboBox{Value: Bind("OptDrawerTrig"), BindingMember: "Code", DisplayMember: "Name", Model: viewmodel.ListDrawerTriggers, MaxSize: Size{Width: 80}},
													Label{Text: "Ширина:", ToolTipText: "Ширина бумаги"}, ComboBox{Value: Bind("PrintPaper"), BindingMember: "Code", DisplayMember: "Name", Model: viewmodel.ListPaperWidths, MaxSize: Size{Width: 70}},
													Label{Text: "Звук:", ToolTipText: "Звук датчика окончания бумаги"}, CheckBox{Checked: Bind("OptNearEnd")},
													Label{Text: "PIN:", ToolTipText: "Пин денежного ящика"}, NumberEdit{Value: Bind("DrawerPin"), MaxSize: Size{Width: 40}},
													Label{Text: "Шрифт:", ToolTipText: "Шрифт А - стандартный, B - компактный"}, ComboBox{Value: Bind("PrintFont"), BindingMember: "Code", DisplayMember: "Name", Model: viewmodel.ListFonts, MaxSize: Size{Width: 70}, ToolTipText: "A-стандратный, B-компактный"},
													Label{Text: "Тест:", ToolTipText: "Тест страница при запуске"}, CheckBox{Checked: Bind("OptAutoTest")},
													Label{Text: "Rise:", ToolTipText: "Время нарастания импульса открывания в миллисекундах."}, NumberEdit{Value: Bind("DrawerRise"), MaxSize: Size{Width: 40}},
													Label{Text: "Бод:"}, ComboBox{Value: Bind("PrintBaud"), BindingMember: "Code", DisplayMember: "Name", Model: viewmodel.ListBaudRates, MaxSize: Size{Width: 70}},
													HSpacer{ColumnSpan: 2},
													Label{Text: "Fall:", ToolTipText: "Время спада импульса в миллисекундах"}, NumberEdit{Value: Bind("DrawerFall"), MaxSize: Size{Width: 40}},
												},
											},
										},
									},

									Composite{
										Layout: VBox{MarginsZero: true, Spacing: 4},
										Children: []Widget{
											GroupBox{
												Title:  "Сеть (LAN)",
												Layout: Grid{Columns: 2, Spacing: 4, Margins: Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
												Children: []Widget{
													Label{Text: "IP:"}, LineEdit{Text: Bind("LanAddr"), MinSize: Size{Width: 90}, MaxSize: Size{Width: 100}},
													Label{Text: "Mask:"}, LineEdit{Text: Bind("LanMask"), MinSize: Size{Width: 90}, MaxSize: Size{Width: 100}},
													Label{Text: "GW:"}, LineEdit{Text: Bind("LanGw"), MinSize: Size{Width: 90}, MaxSize: Size{Width: 100}},
													Label{Text: "Port:"}, NumberEdit{Value: Bind("LanPort"), MaxSize: Size{Width: 60}},
												},
											},
											GroupBox{
												Title:  "Вид чека и Опции",
												Layout: Grid{Columns: 4, Spacing: 4, Margins: Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
												Children: []Widget{
													Label{Text: "QR:", ToolTipText: "Позиция QR-кода"},
													ComboBox{Value: Bind("OptQRPos"), BindingMember: "Code", DisplayMember: "Name", Model: viewmodel.ListQRPositions, MaxSize: Size{Width: 40}},
													Label{Text: "Текст QR:"},
													CheckBox{Checked: Bind("OptTextQR")},
													Label{Text: "Покупок:"},
													CheckBox{Checked: Bind("OptCountInCheck")},
													Label{Text: "Округл.:"},
													ComboBox{Value: Bind("OptRounding"), BindingMember: "Code", DisplayMember: "Name", Model: viewmodel.ListRoundingOptions, MaxSize: Size{Width: 40}},
													Label{Text: "X-Отчет:"},
													CheckBox{Text: "Полный", Checked: Bind("OptB9_FullX"), ToolTipText: "Печатать полный X-отчет"},
													Label{Text: "Баз. СНО:"},
													ComboBox{
														AssignTo:      &b9ComboBox,
														Value:         Bind("OptB9_BaseTax"),
														BindingMember: "Code",
														DisplayMember: "Name",
														Model:         vm.OptB9_SNO,
														MinSize:       Size{Width: 40},
														ToolTipText:   "Система налогообложения по умолчанию",
													},
												},
											},
										},
									},
								},
							},
						},
					},

					{
						Title:  "Клише",
						Layout: VBox{Margins: Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 5},
						Children: []Widget{
							Composite{
								Layout: HBox{MarginsZero: true, Alignment: AlignHCenterVCenter},
								Children: []Widget{
									Label{Text: "Редактировать:"},
									ComboBox{
										Value:         Bind("SelectedClicheType"),
										Model:         viewmodel.ListClicheTypes,
										BindingMember: "Code", DisplayMember: "Name",
										MinSize: Size{Width: 100},
									},
								},
							},
							Composite{
								Layout: HBox{MarginsZero: true, Spacing: 5},
								Children: []Widget{
									TableView{
										AssignTo:         &clicheTable,
										Model:            clicheModel,
										AlternatingRowBG: true,
										Columns: []TableViewColumn{
											{Title: "#", Width: 30},
											{Title: "Fmt", Width: 60},
											{Title: "Текст", Width: 200},
										},
										MinSize: Size{Width: 300, Height: 200},
										MaxSize: Size{Width: 300, Height: 200},
										OnCurrentIndexChanged: func() {
											idx := clicheTable.CurrentIndex()
											if idx < 0 {
												clicheEditorGroup.SetEnabled(false)
												return
											}

											srcItem := vm.ClicheItems[idx]
											vm.TempClicheLine.Index = srcItem.Index
											vm.TempClicheLine.Line = srcItem.Line

											if clicheEditorBinder != nil {
												clicheEditorBinder.SetDataSource(vm.TempClicheLine)
												clicheEditorBinder.Reset()
											}

											clicheEditorGroup.SetEnabled(true)
											clicheEditorGroup.SetTitle(fmt.Sprintf("Настройки строки №%d", idx+1))
										},
									},
									GroupBox{
										AssignTo: &clicheEditorGroup,
										Title:    "Настройки строки",
										Layout:   VBox{Margins: Margins{Left: 10, Top: 10, Right: 10, Bottom: 10}, Spacing: 8},
										Enabled:  false,
										MaxSize:  Size{Width: 300, Height: 250},
										DataBinder: DataBinder{
											AssignTo:   &clicheEditorBinder,
											DataSource: vm.TempClicheLine,
										},
										Children: []Widget{
											Label{Text: "Текст:"},
											LineEdit{
												AssignTo: &ceText,
												Text:     Bind("Text"),
											},
											Composite{
												Layout: Grid{Columns: 2, Spacing: 10},
												Children: []Widget{
													Label{Text: "Выравнивание:"},
													ComboBox{
														AssignTo:      &ceAlign,
														Value:         Bind("Align"),
														Model:         viewmodel.ListAlignments,
														BindingMember: "Code", DisplayMember: "Name",
													},
													Label{Text: "Шрифт:"},
													ComboBox{
														AssignTo:      &ceFont,
														Value:         Bind("Font"),
														Model:         viewmodel.ListFonts,
														BindingMember: "Code", DisplayMember: "Name",
													},
													Label{Text: "Подчеркивание:"},
													ComboBox{
														AssignTo:      &ceUnder,
														Value:         Bind("Underline"),
														Model:         viewmodel.ListUnderlineOptions,
														BindingMember: "Code", DisplayMember: "Name",
													},
												},
											},
											GroupBox{
												Title:  "Масштабирование",
												Layout: Grid{Columns: 4},
												Children: []Widget{
													Label{Text: "Ширина:"},
													NumberEdit{
														AssignTo: &ceWidth,
														Value:    Bind("Width"),
														MinValue: 0, MaxValue: 8, MaxSize: Size{Width: 40},
													},
													Label{Text: "Высота:"},
													NumberEdit{
														AssignTo: &ceHeight,
														Value:    Bind("Height"),
														MinValue: 0, MaxValue: 8, MaxSize: Size{Width: 40},
													},
												},
											},
											CheckBox{
												AssignTo: &ceInvert,
												Text:     "Инверсия (Белым по черному)",
												Checked:  Bind("Invert"),
											},
											VSpacer{Size: 5},
											PushButton{
												Text: "Применить изменения строки",
												OnClicked: func() {
													idx := clicheTable.CurrentIndex()
													if idx < 0 {
														return
													}

													originalItem := vm.ClicheItems[idx]
													originalItem.Line = vm.TempClicheLine.Line
													clicheModel.UpdateRow(idx)
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
