package gui

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"

	"mitsuscanner/driver"
	"mitsuscanner/internal/app"
	"mitsuscanner/internal/service/registration"
)

// hasBit проверяет, установлен ли бит в целом числе
func hasBit(value int, bit int) bool {
	return (value & (1 << bit)) != 0
}

// RegViewModel - модель данных для формы регистрации
type RegViewModel struct {
	RNM           string
	INN           string
	OrgName       string
	Address       string
	Place         string
	Email         string
	Site          string
	AutomatNumber string
	FFD           string
	Reasons       string

	OFDName string
	OFDINN  string

	// d.CheckBoxes (Settings)
	ModeAutonomous bool
	ModeEncryption bool
	ModeService    bool
	ModeInternet   bool
	ModeAutomat    bool
	ModeBSO        bool
	ModeExcise     bool
	ModeGambling   bool
	ModeLottery    bool
	ModeMarking    bool
	ModePawn       bool
	ModeInsurance  bool
	ModeCatering   bool
	ModeWholesale  bool
	ModeVending    bool

	// SNO (Taxation)
	TaxOSN        bool // 0
	TaxUSN        bool // 1
	TaxUSN_M      bool // 2
	TaxENVD       bool // 3
	TaxESHN       bool // 4
	TaxPat        bool // 5
	TaxSystemBase string

	// Информация о ФН
	FnNumber     string
	FnPhase      string
	FnPhaseText  string
	FnPhaseColor walk.Color
	FnValidDate  string
}

// RegistrationTab - контроллер вкладки регистрации
type RegistrationTab struct {
	app          *app.App
	viewModel    *RegViewModel
	binder       *walk.DataBinder
	fnPhaseLabel *walk.Label
}

// NewRegistrationTab создает новый экземпляр контроллера
func NewRegistrationTab(a *app.App) *RegistrationTab {
	return &RegistrationTab{
		app:       a,
		viewModel: &RegViewModel{},
	}
}

// decodeFnPhase возвращает текст и цвет для фазы ФН.
// PHASE: 0x01=Готов к фискализации (Синий), 0x03=Боевой режим (Зелёный),
// 0x07=ФН закрыт, отчёт не отправлен (Красный), 0x0F=ФР в архиве (Синий)
func decodeFnPhase(phase string) (text string, color walk.Color) {
	phase = strings.TrimPrefix(strings.ToLower(phase), "0x")
	val, err := strconv.ParseInt(phase, 16, 32)
	if err != nil {
		return "Неизвестно", walk.RGB(0, 0, 0) // Чёрный по умолчанию
	}

	switch val {
	case 0x01:
		return "Готов к фискализации", walk.RGB(0, 0, 255) // Синий
	case 0x03:
		return "Боевой режим", walk.RGB(0, 128, 0) // Зелёный
	case 0x07:
		return "ФН закрыт", walk.RGB(255, 0, 0) // Красный
	case 0x0F:
		return "ФР в архиве", walk.RGB(0, 0, 255) // Синий
	default:
		return fmt.Sprintf("Неизвестная фаза (%s)", phase), walk.RGB(128, 128, 128)
	}
}

// Create возвращает описание вкладки "Регистрация"
func (rt *RegistrationTab) Create() d.TabPage {
	return d.TabPage{
		Title:  "Регистрация",
		Layout: d.VBox{},
		Children: []d.Widget{
			// Верхняя панель
			d.Composite{
				Layout: d.Grid{Columns: 3, Margins: d.Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 5},
				Children: []d.Widget{
					d.Label{Text: "Регистрационный номер ККТ (РНМ):"},
					d.LineEdit{Text: d.Bind("RNM"), MinSize: d.Size{Width: 150}},
					d.PushButton{Text: "Вычислить (CRC)", OnClicked: rt.onCalcRNM},
				},
			},

			// Основной контент
			d.Composite{
				Layout: d.VBox{Margins: d.Margins{Left: 8, Top: 8, Right: 8, Bottom: 2}, Spacing: 3},
				Children: []d.Widget{
					d.Composite{
						// Используем HBox для расположения блоков по горизонтали
						// Alignment: AlignTop прижмет маленький блок к верху
						Layout: d.HBox{Margins: d.Margins{Left: 8, Top: 2, Right: 8, Bottom: 2}, Spacing: 3, Alignment: d.Alignment2D(d.AlignDefault)},
						Children: []d.Widget{
							d.GroupBox{
								Title: "Реквизиты организации",
								// StretchFactor: 3 означает, что этот блок будет занимать ~75% ширины (3 части)
								// StretchFactor: 3,
								Layout: d.Grid{Columns: 2, Margins: d.Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 5},
								Children: []d.Widget{
									d.Label{Text: "Наименование:"},
									d.LineEdit{Text: d.Bind("OrgName")},
									d.Label{Text: "ИНН:"},
									d.LineEdit{Text: d.Bind("INN")},
									d.Label{Text: "Адрес расчетов:"},
									d.LineEdit{Text: d.Bind("Address")},
									d.Label{Text: "Место расчетов:"},
									d.LineEdit{Text: d.Bind("Place")},
									d.Label{Text: "E-mail отправителя:"},
									d.LineEdit{Text: d.Bind("Email")},
									d.Label{Text: "Сайт ФНС:"},
									d.LineEdit{Text: d.Bind("Site")},
								},
							},
							d.GroupBox{
								Title: "Системы налогообложения",
								// StretchFactor: 1 означает, что этот блок будет занимать ~25% ширины (1 часть)
								// StretchFactor: 1,
								// MinSize можно задать, чтобы чекбоксы не сплющивались, если окно сузят
								MinSize: d.Size{Width: 150},
								Layout:  d.VBox{Margins: d.Margins{Left: 3, Top: 1, Right: 3, Bottom: 1}, Spacing: 1},
								Children: []d.Widget{
									d.CheckBox{Text: "ОСН", Checked: d.Bind("TaxOSN"), Alignment: d.Alignment2D(d.AlignNear)},
									d.CheckBox{Text: "УСН доход", Checked: d.Bind("TaxUSN"), Alignment: d.Alignment2D(d.AlignNear)},
									d.CheckBox{Text: "УСН доход - расход", Checked: d.Bind("TaxUSN_M"), Alignment: d.Alignment2D(d.AlignNear)},
									d.CheckBox{Text: "ЕСХН", Checked: d.Bind("TaxESHN"), Alignment: d.Alignment2D(d.AlignNear)},
									d.CheckBox{Text: "Патент", Checked: d.Bind("TaxPat"), Alignment: d.Alignment2D(d.AlignNear)},
									d.Label{Text: "СНО по умолчанию:", Font: d.Font{PointSize: 7}},
									d.ComboBox{
										Value:         d.Bind("TaxSystemBase"),
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
					d.Composite{
						Layout: d.HBox{Margins: d.Margins{Left: 3, Top: 1, Right: 3, Bottom: 1}, Spacing: 5},
						Children: []d.Widget{
							d.GroupBox{
								Title:  "Фискальные признаки",
								Layout: d.Grid{Columns: 3, Margins: d.Margins{Left: 3, Top: 1, Right: 3, Bottom: 1}, Spacing: 1},
								Children: []d.Widget{
									// Колонка 1
									d.CheckBox{Text: "Автономный режим", Checked: d.Bind("ModeAutonomous")},
									d.CheckBox{Text: "Шифрование данных", Checked: d.Bind("ModeEncryption")},
									d.CheckBox{Text: "Расчеты за услуги", Checked: d.Bind("ModeService")},
									d.CheckBox{Text: "Расчеты в Интернет", Checked: d.Bind("ModeInternet")},
									d.CheckBox{Text: "Принтер в автомате", Checked: d.Bind("ModeAutomat")},
									// Колонка 2
									d.CheckBox{Text: "Только БСО", Checked: d.Bind("ModeBSO")},
									d.CheckBox{Text: "Подакцизные товары", Checked: d.Bind("ModeExcise")},
									d.CheckBox{Text: "Проведение азартных игр", Checked: d.Bind("ModeGambling")},
									d.CheckBox{Text: "Проведение лотерей", Checked: d.Bind("ModeLottery")},
									d.CheckBox{Text: "Маркированные товары", Checked: d.Bind("ModeMarking")},
									// Колонка 3
									d.CheckBox{Text: "Ломбард", Checked: d.Bind("ModePawn")},
									d.CheckBox{Text: "Страхование", Checked: d.Bind("ModeInsurance")},
									d.CheckBox{Text: "Общепит", Checked: d.Bind("ModeCatering")},
									d.CheckBox{Text: "Оптовая торговля", Checked: d.Bind("ModeWholesale")},
									d.CheckBox{Text: "Вендинг", Checked: d.Bind("ModeVending")},
								},
							},
							d.GroupBox{
								Title:   "Информация о ФН",
								Layout:  d.VBox{MarginsZero: true, Spacing: 3},
								MinSize: d.Size{Width: 200},
								Children: []d.Widget{
									// Информационные поля
									d.Composite{
										Layout: d.Grid{Columns: 2, Spacing: 5},
										Children: []d.Widget{
											d.Label{Text: "№:"},
											d.Label{Text: d.Bind("FnNumber"), Font: d.Font{Bold: true}},
											d.Label{Text: "Фаза:"},
											d.Label{AssignTo: &rt.fnPhaseLabel, Text: d.Bind("FnPhaseText")},
											d.Label{Text: "До:"},
											d.Label{Text: d.Bind("FnValidDate"), Font: d.Font{Bold: true}},
										},
									},
									// Кнопки управления
									d.Composite{
										Layout: d.HBox{MarginsZero: true, Spacing: 5, Alignment: d.AlignHCenterVCenter},
										Children: []d.Widget{
											d.PushButton{
												Text:        "Отправить в ОФД",
												OnClicked:   rt.onSendToOfd,
												MinSize:     d.Size{Width: 110},
												ToolTipText: "Отправить первый неотправленный документ в ОФД",
											},
											d.PushButton{
												Text:        "↻", // Unicode символ обновления
												OnClicked:   rt.onRefreshFnInfo,
												MaxSize:     d.Size{Width: 30, Height: 30},
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

			// --- 3. Блок: Причины (слева) + ОФД (справа) ---
			d.Composite{
				// Выравнивание по верху, чтобы блоки не разъезжались по высоте
				Layout: d.HBox{Margins: d.Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 10, Alignment: d.Alignment2D(d.AlignNear)},
				Children: []d.Widget{

					// ПРАВАЯ ЧАСТЬ: ОФД
					d.GroupBox{
						Title:         "Оператор фискальных данных",
						StretchFactor: 2, // Занимает 2 части ширины (шире, чем причины)
						Layout:        d.VBox{Margins: d.Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 5},
						Children: []d.Widget{
							d.Composite{
								Layout: d.HBox{Margins: d.Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 5},
								Children: []d.Widget{
									d.Label{Text: "ИНН ОФД:", TextAlignment: d.AlignFar},
									d.LineEdit{Text: d.Bind("OFDINN")},
								},
							},
							d.Composite{
								Layout: d.HBox{Margins: d.Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 5},
								Children: []d.Widget{
									d.Label{Text: "Наименование ОФД:", TextAlignment: d.AlignFar},
									d.LineEdit{Text: d.Bind("OFDName")},
								},
							},
						},
					},

					// ЛЕВАЯ ЧАСТЬ: Причины перерегистрации
					d.GroupBox{
						Title:         "Причины перерегистрации",
						StretchFactor: 1, // Занимает 1 часть ширины
						Layout:        d.VBox{Margins: d.Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 10},
						Children: []d.Widget{
							// Поле ввода кодов
							d.LineEdit{Text: d.Bind("Reasons"), ReadOnly: true},

							// Контейнер для центрирования кнопки
							d.Composite{
								Layout: d.HBox{MarginsZero: true, Alignment: d.AlignHCenterVCenter},
								Children: []d.Widget{
									d.PushButton{
										Text:      "Выбрать...",
										OnClicked: rt.onSelectReasons,
										MinSize:   d.Size{Width: 100}, // Фиксируем размер кнопки
										MaxSize:   d.Size{Width: 100},
									},
								},
							},
						},
					},
				},
			},

			// Кнопки
			d.Composite{
				Layout: d.HBox{Margins: d.Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 5},
				Children: []d.Widget{
					d.PushButton{Text: "Считать из ККТ", OnClicked: rt.onReadRegistration},
					d.HSpacer{},
					d.PushButton{Text: "Закрытие ФН", OnClicked: rt.onCloseFn},
					d.PushButton{Text: "Замена ФН", OnClicked: rt.onReplaceFn},
					d.PushButton{Text: "Перерегистрация", OnClicked: rt.onReregister},
					d.PushButton{Text: "Регистрация", OnClicked: rt.onRegister},
				},
			},
		},
		DataBinder: d.DataBinder{
			AssignTo:       &rt.binder,
			DataSource:     rt.viewModel,
			ErrorPresenter: d.ToolTipErrorPresenter{},
		},
	}
}

// --- Обработчики событий ---

// onCalcRNM открывает диалог генерации РНМ
func (rt *RegistrationTab) onCalcRNM() {
	// Сначала забираем данные из формы в модель
	if err := rt.binder.Submit(); err != nil {
		return
	}

	inn := strings.TrimSpace(rt.viewModel.INN)
	if len(inn) != 10 && len(inn) != 12 {
		walk.MsgBox(rt.app.MainWindow, "Ошибка", "Для расчета РНМ заполните корректный ИНН пользователя (10 или 12 цифр).", walk.MsgBoxIconError)
		return
	}

	// Пытаемся получить заводской номер из драйвера
	serial := ""
	if rt.app.GetDriver() != nil {
		info, err := rt.app.GetDriver().GetFiscalInfo()
		if err == nil && info != nil {
			serial = info.SerialNumber
		}
	}

	// Запускаем диалог
	rt.runRnmGenerationDialog(inn, serial)
}

func (rt *RegistrationTab) onReadRegistration() {
	drv := rt.app.GetDriver()
	if drv == nil {
		walk.MsgBox(rt.app.MainWindow, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		regData, err := drv.GetRegistrationData()
		if err != nil {
			rt.app.MainWindow.Synchronize(func() { walk.MsgBox(rt.app.MainWindow, "Ошибка", err.Error(), walk.MsgBoxIconError) })
			return
		}

		fnStatus, fnErr := drv.GetFnStatus()

		rt.app.MainWindow.Synchronize(func() {

			rt.viewModel.RNM = regData.RNM
			rt.viewModel.INN = regData.Inn
			rt.viewModel.OrgName = regData.OrgName
			rt.viewModel.Address = regData.Address
			rt.viewModel.Place = regData.Place
			rt.viewModel.Email = regData.EmailSender
			rt.viewModel.Site = regData.Site
			rt.viewModel.OFDINN = regData.OfdInn
			rt.viewModel.OFDName = html.UnescapeString(regData.OfdName)

			// --- Парсинг атрибутов режимов работы ---

			// Парсинг MODE и ExtMODE с использованием hasBit
			modeInt := int(regData.ModeMask)
			extModeInt := int(regData.ExtModeMask)

			rt.viewModel.ModeEncryption = hasBit(modeInt, 0) // Шифрование
			rt.viewModel.ModeAutonomous = hasBit(modeInt, 1) // Автономный режим
			rt.viewModel.ModeService = hasBit(modeInt, 3)    // Услуги
			rt.viewModel.ModeBSO = hasBit(modeInt, 4)        // БСО
			rt.viewModel.ModeInternet = hasBit(modeInt, 5)   // Интернет
			rt.viewModel.ModeCatering = hasBit(modeInt, 6)   // Общепит
			rt.viewModel.ModeWholesale = hasBit(modeInt, 7)  // Опт

			rt.viewModel.ModeExcise = hasBit(extModeInt, 0)    // Подакцизные
			rt.viewModel.ModeGambling = hasBit(extModeInt, 1)  // Азартные
			rt.viewModel.ModeLottery = hasBit(extModeInt, 2)   // Лотереи
			rt.viewModel.ModeAutomat = hasBit(extModeInt, 3)   // Принтер в автомате
			rt.viewModel.ModeMarking = hasBit(extModeInt, 4)   // Маркированные товары
			rt.viewModel.ModePawn = hasBit(extModeInt, 5)      // Ломбард
			rt.viewModel.ModeInsurance = hasBit(extModeInt, 6) // Страхование
			rt.viewModel.ModeVending = hasBit(extModeInt, 7)   // Вендинг

			// Парсинг СНО
			rt.viewModel.TaxOSN = false
			rt.viewModel.TaxUSN = false
			rt.viewModel.TaxUSN_M = false
			rt.viewModel.TaxENVD = false
			rt.viewModel.TaxESHN = false
			rt.viewModel.TaxPat = false

			taxParts := strings.Split(regData.TaxSystems, ",")
			for _, t := range taxParts {
				trimmedT := strings.TrimSpace(t)
				switch trimmedT {
				case "0":
					rt.viewModel.TaxOSN = true
				case "1":
					rt.viewModel.TaxUSN = true
				case "2":
					rt.viewModel.TaxUSN_M = true
				case "3":
					rt.viewModel.TaxENVD = true
				case "4":
					rt.viewModel.TaxESHN = true
				case "5":
					rt.viewModel.TaxPat = true
				}
			}

			// Установка TaxSystemBase на первую зарегистрированную СНО по умолчанию
			rt.viewModel.TaxSystemBase = regData.TaxBase

			if err := rt.binder.Reset(); err != nil {
				walk.MsgBox(rt.app.MainWindow, "Ошибка биндинга", fmt.Sprintf("Ошибка обновления UI: %v", err), walk.MsgBoxIconError)
			}

			// Читаем информацию о ФН
			if fnErr == nil {
				rt.viewModel.FnNumber = fnStatus.Serial
				rt.viewModel.FnValidDate = fnStatus.Valid
				rt.viewModel.FnPhase = fnStatus.Phase

				// Декодируем фазу
				phaseText, phaseColor := decodeFnPhase(fnStatus.Phase)
				rt.viewModel.FnPhaseText = phaseText
				rt.viewModel.FnPhaseColor = phaseColor
			} else {
				rt.viewModel.FnNumber = "Ошибка чтения"
				rt.viewModel.FnPhaseText = "—"
				rt.viewModel.FnValidDate = "—"
			}

			if err := rt.binder.Reset(); err != nil {
				walk.MsgBox(rt.app.MainWindow, "Ошибка биндинга", fmt.Sprintf("Ошибка обновления UI: %v", err), walk.MsgBoxIconError)
			}

			if fnErr == nil && rt.fnPhaseLabel != nil {
				_, phaseColor := decodeFnPhase(fnStatus.Phase)
				rt.fnPhaseLabel.SetTextColor(phaseColor)
			}
		})
	}()
}

func (rt *RegistrationTab) onRegister() {
	drv := rt.app.GetDriver()
	if drv == nil {
		return
	}
	if err := rt.binder.Submit(); err != nil {
		return
	}

	if !regexp.MustCompile(`^\d+$`).MatchString(rt.viewModel.INN) || (len(rt.viewModel.INN) != 10 && len(rt.viewModel.INN) != 12) {
		walk.MsgBox(rt.app.MainWindow, "Ошибка", "ИНН должен состоять только из цифр и иметь длину 10 или 12 символов.", walk.MsgBoxIconError)
		return
	}

	if strings.TrimSpace(rt.viewModel.OrgName) == "" {
		walk.MsgBox(rt.app.MainWindow, "Ошибка", "Поле 'Наименование' обязательно для заполнения.", walk.MsgBoxIconError)
		return
	}

	if strings.TrimSpace(rt.viewModel.Address) == "" {
		walk.MsgBox(rt.app.MainWindow, "Ошибка", "Поле 'Адрес расчетов' обязательно для заполнения.", walk.MsgBoxIconError)
		return
	}

	if strings.TrimSpace(rt.viewModel.Place) == "" {
		walk.MsgBox(rt.app.MainWindow, "Ошибка", "Поле 'Место расчетов' обязательно для заполнения.", walk.MsgBoxIconError)
		return
	}

	req := rt.fillRequestFromModel(false)

	go func() {
		if err := drv.SetCashier("Администратор", ""); err != nil {
			return
		}
		resp, err := drv.Register(req)
		if err != nil {
			rt.app.MainWindow.Synchronize(func() {
				walk.MsgBox(rt.app.MainWindow, "Ошибка регистрации", err.Error(), walk.MsgBoxIconError)
			})
			return
		}
		if err := drv.PrintLastDocument(); err != nil {
		}
		typeCode, err := drv.GetCurrentDocumentType()
		if err != nil {
			// log.Printf("[DRIVER] Ошибка получения типа документа: %v", err)
		}
		meta := driver.GetReportMeta(typeCode)
		regData, err := GetFullRegistrationData(drv)
		if err != nil {
			rt.app.MainWindow.Synchronize(func() { walk.MsgBox(rt.app.MainWindow, "Ошибка", err.Error(), walk.MsgBoxIconError) })
			return
		}
		if resp.FdNumber != "" {
			regData.FdNumber = resp.FdNumber
		}
		if resp.FpNumber != "" {
			regData.FpNumber = resp.FpNumber
		}
		meta.Data = regData
		if meta.Kind == driver.ReportReg || meta.Kind == driver.ReportRereg {
			rt.app.MainWindow.Synchronize(func() { RunReportModal(rt.app.MainWindow, meta) })
		}
	}()
}

func (rt *RegistrationTab) onReregister() {
	drv := rt.app.GetDriver()
	if drv == nil {
		return
	}
	if err := rt.binder.Submit(); err != nil {
		return
	}

	if rt.viewModel.Reasons == "" {
		walk.MsgBox(rt.app.MainWindow, "Ошибка", "Не выбраны причины перерегистрации", walk.MsgBoxIconError)
		return
	}

	var reasons []int
	parts := strings.Split(rt.viewModel.Reasons, ",")
	for _, p := range parts {
		code, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			walk.MsgBox(rt.app.MainWindow, "Ошибка", "Некорректный формат причин", walk.MsgBoxIconError)
			return
		}
		reasons = append(reasons, code)
	}

	if !regexp.MustCompile(`^\d+$`).MatchString(rt.viewModel.INN) || (len(rt.viewModel.INN) != 10 && len(rt.viewModel.INN) != 12) {
		walk.MsgBox(rt.app.MainWindow, "Ошибка", "ИНН должен состоять только из цифр и иметь длину 10 или 12 символов.", walk.MsgBoxIconError)
		return
	}

	if strings.TrimSpace(rt.viewModel.OrgName) == "" {
		walk.MsgBox(rt.app.MainWindow, "Ошибка", "Поле 'Наименование' обязательно для заполнения.", walk.MsgBoxIconError)
		return
	}

	if strings.TrimSpace(rt.viewModel.Address) == "" {
		walk.MsgBox(rt.app.MainWindow, "Ошибка", "Поле 'Адрес расчетов' обязательно для заполнения.", walk.MsgBoxIconError)
		return
	}

	if strings.TrimSpace(rt.viewModel.Place) == "" {
		walk.MsgBox(rt.app.MainWindow, "Ошибка", "Поле 'Место расчетов' обязательно для заполнения.", walk.MsgBoxIconError)
		return
	}

	req := rt.fillRequestFromModel(true)

	go func() {
		if err := drv.SetCashier("Администратор", ""); err != nil {
			return
		}
		_, err := drv.Reregister(req, reasons)
		if err != nil {
			rt.app.MainWindow.Synchronize(func() {
				walk.MsgBox(rt.app.MainWindow, "Ошибка перерегистрации", err.Error(), walk.MsgBoxIconError)
			})
			return
		}
		if err := drv.PrintLastDocument(); err != nil {
		}
		typeCode, err := drv.GetCurrentDocumentType()
		if err != nil {
			// log.Printf("[DRIVER] Ошибка получения типа документа: %v", err)
		}
		meta := driver.GetReportMeta(typeCode)
		regData, err := GetFullRegistrationData(drv)
		if err != nil {
			rt.app.MainWindow.Synchronize(func() { walk.MsgBox(rt.app.MainWindow, "Ошибка", err.Error(), walk.MsgBoxIconError) })
			return
		}
		meta.Data = regData
		if meta.Kind == driver.ReportReg || meta.Kind == driver.ReportRereg {
			rt.app.MainWindow.Synchronize(func() { RunReportModal(rt.app.MainWindow, meta) })
		}
	}()
}

func (rt *RegistrationTab) onSelectReasons() {
	if err := rt.binder.Submit(); err != nil {
		return
	}
	reasons, ok := RunReasonDialog(rt.app.MainWindow, rt.viewModel.Reasons)
	if ok {
		rt.viewModel.Reasons = reasons
		if err := rt.binder.Reset(); err != nil {
			fmt.Println("Binder reset error:", err)
		}
	}
}

func (rt *RegistrationTab) onReplaceFn() {
	drv := rt.app.GetDriver()
	if drv == nil {
		return
	}
	if err := rt.binder.Submit(); err != nil {
		return
	}

	req := rt.fillRequestFromModel(false)
	reasons := []int{1}

	go func() {
		if err := drv.SetCashier("Администратор", ""); err != nil {
			return
		}
		resp, err := drv.Reregister(req, reasons)
		if err != nil {
			rt.app.MainWindow.Synchronize(func() {
				walk.MsgBox(rt.app.MainWindow, "Ошибка замены ФН", err.Error(), walk.MsgBoxIconError)
			})
		} else {
			rt.app.MainWindow.Synchronize(func() {
				walk.MsgBox(rt.app.MainWindow, "Успех", fmt.Sprintf("ФН заменен!\nФД: %s\nФП: %s", resp.FdNumber, resp.FpNumber), walk.MsgBoxIconInformation)
			})
		}
	}()
}

func (rt *RegistrationTab) onCloseFn() {
	drv := rt.app.GetDriver()
	if drv == nil {
		return
	}
	if walk.MsgBox(rt.app.MainWindow, "Подтверждение", "Вы действительно хотите закрыть фискальный архив?\nЭто необратимая операция!", walk.MsgBoxYesNo|walk.MsgBoxIconWarning) != walk.DlgCmdYes {
		return
	}
	go func() {
		// 1. Закрытие ФН (включает PRINT)
		result, err := drv.CloseFiscalArchive()
		if err != nil {
			rt.app.MainWindow.Synchronize(func() { walk.MsgBox(rt.app.MainWindow, "Ошибка", err.Error(), walk.MsgBoxIconError) })
			return
		}

		// 2. Получение типа документа для метаданных отчёта
		typeCode, _ := drv.GetCurrentDocumentType()
		meta := driver.GetReportMeta(typeCode)

		// 3. Сбор данных для отчёта через хелпер
		reportData, err := GetCloseFnReportData(drv, result.FD, result.FP)
		if err != nil {
			rt.app.MainWindow.Synchronize(func() { walk.MsgBox(rt.app.MainWindow, "Ошибка", err.Error(), walk.MsgBoxIconError) })
			return
		}

		// 4. Отображение отчёта
		meta.Data = reportData
		rt.app.MainWindow.Synchronize(func() { RunReportModal(rt.app.MainWindow, meta) })
	}()
}

// onSendToOfd отправляет первый неотправленный документ в ОФД
func (rt *RegistrationTab) onSendToOfd() {
	drv := rt.app.GetDriver()
	if drv == nil {
		walk.MsgBox(rt.app.MainWindow, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		result, err := SendFirstUnsentDocument(drv)
		rt.app.MainWindow.Synchronize(func() {
			if err != nil {
				walk.MsgBox(rt.app.MainWindow, "Ошибка", err.Error(), walk.MsgBoxIconError)
				return
			}

			if result.Success {
				walk.MsgBox(rt.app.MainWindow, "Успех",
					fmt.Sprintf("Отправлено документов: %d", result.DocumentsSent),
					walk.MsgBoxIconInformation)
			} else {
				walk.MsgBox(rt.app.MainWindow, "Информация", result.ErrorMessage, walk.MsgBoxIconWarning)
			}

			// Обновляем информацию о ФН
			rt.onRefreshFnInfo()
		})
	}()
}

// onRefreshFnInfo обновляет информацию о ФН
func (rt *RegistrationTab) onRefreshFnInfo() {
	drv := rt.app.GetDriver()
	if drv == nil {
		return
	}

	go func() {
		// Используем локальный метод обновления модели, так как глобальный helper RefreshFnInfo
		// теперь не имеет доступа к rt.viewModel и rt.binder
		fnStatus, err := drv.GetFnStatus()
		if err != nil {
			rt.app.MainWindow.Synchronize(func() {
				logMsg("Ошибка обновления ФН: %v", err)
			})
			return
		}

		rt.app.MainWindow.Synchronize(func() {
			rt.viewModel.FnNumber = fnStatus.Serial
			rt.viewModel.FnValidDate = fnStatus.Valid
			rt.viewModel.FnPhase = fnStatus.Phase

			phaseText, phaseColor := decodeFnPhase(fnStatus.Phase)
			rt.viewModel.FnPhaseText = phaseText
			rt.viewModel.FnPhaseColor = phaseColor

			if rt.binder != nil {
				rt.binder.Reset()
			}
			if rt.fnPhaseLabel != nil {
				rt.fnPhaseLabel.SetTextColor(phaseColor)
			}
		})
	}()
}

func (rt *RegistrationTab) fillRequestFromModel(isRereg bool) driver.RegistrationRequest {
	req := driver.RegistrationRequest{
		IsReregistration: isRereg,
		RNM:              rt.viewModel.RNM,
		Inn:              rt.viewModel.INN,
		OrgName:          rt.viewModel.OrgName,
		Address:          rt.viewModel.Address,
		Place:            rt.viewModel.Place,
		SenderEmail:      rt.viewModel.Email,
		FnsSite:          rt.viewModel.Site,
		FfdVer:           rt.viewModel.FFD,
		OfdName:          rt.viewModel.OFDName,
		OfdInn:           rt.viewModel.OFDINN,
		AutonomousMode:   rt.viewModel.ModeAutonomous,
		Encryption:       rt.viewModel.ModeEncryption,
		Service:          rt.viewModel.ModeService,
		InternetCalc:     rt.viewModel.ModeInternet,
		BSO:              rt.viewModel.ModeBSO,
		Gambling:         rt.viewModel.ModeGambling,
		Lottery:          rt.viewModel.ModeLottery,
		Excise:           rt.viewModel.ModeExcise,
		Marking:          rt.viewModel.ModeMarking,
		PawnShop:         rt.viewModel.ModePawn,
		Insurance:        rt.viewModel.ModeInsurance,
		Catering:         rt.viewModel.ModeCatering,
		Wholesale:        rt.viewModel.ModeWholesale,
		Vending:          rt.viewModel.ModeVending,
		PrinterAutomat:   rt.viewModel.ModeAutomat,
		TaxSystemBase:    rt.viewModel.TaxSystemBase,
	}

	var taxCodes []string
	if rt.viewModel.TaxOSN {
		taxCodes = append(taxCodes, "0")
	}
	if rt.viewModel.TaxUSN {
		taxCodes = append(taxCodes, "1")
	}
	if rt.viewModel.TaxUSN_M {
		taxCodes = append(taxCodes, "2")
	}
	if rt.viewModel.TaxENVD {
		taxCodes = append(taxCodes, "3")
	}
	if rt.viewModel.TaxESHN {
		taxCodes = append(taxCodes, "4")
	}
	if rt.viewModel.TaxPat {
		taxCodes = append(taxCodes, "5")
	}
	req.TaxSystems = strings.Join(taxCodes, ",")

	return req
}

// --- Логика генерации РНМ (CRC) ---

// runRnmGenerationDialog запускает диалоговое окно для расчета РНМ
func (rt *RegistrationTab) runRnmGenerationDialog(inn, serial string) {
	var dlg *walk.Dialog
	var acceptPB, cancelPB *walk.PushButton
	var db *walk.DataBinder

	// Модель для диалога
	dlgModel := struct {
		Serial   string
		OrderNum string
	}{
		Serial:   serial,
		OrderNum: "1", // По умолчанию 1
	}

	err := d.Dialog{
		AssignTo:      &dlg,
		Title:         "Генерация РНМ (CRC16)",
		MinSize:       d.Size{Width: 350, Height: 200},
		Layout:        d.VBox{},
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		DataBinder: d.DataBinder{
			AssignTo:   &db,
			DataSource: &dlgModel,
		},
		Children: []d.Widget{
			d.Label{Text: "ИНН Пользователя (из формы):"},
			d.LineEdit{Text: inn, ReadOnly: true}, // ИНН только для отображения

			d.Label{Text: "Заводской номер ККТ (20 знаков):"},
			d.LineEdit{Text: d.Bind("Serial"), MaxLength: 20},

			d.Label{Text: "Порядковый номер регистрации (обычно 1):"},
			d.LineEdit{Text: d.Bind("OrderNum"), MaxLength: 10},

			d.Composite{
				Layout: d.HBox{},
				Children: []d.Widget{
					d.HSpacer{},
					d.PushButton{
						AssignTo: &acceptPB,
						Text:     "Сгенерировать",
						OnClicked: func() {
							if err := db.Submit(); err != nil {
								return
							}
							if dlgModel.Serial == "" {
								walk.MsgBox(dlg, "Ошибка", "Введите заводской номер", walk.MsgBoxIconError)
								return
							}

							// Расчет с использованием нового сервиса
							rnm, err := registration.CalculateRNM(dlgModel.OrderNum, inn, dlgModel.Serial)
							if err != nil {
								walk.MsgBox(dlg, "Ошибка расчета", err.Error(), walk.MsgBoxIconError)
								return
							}

							// Применяем результат
							rt.viewModel.RNM = rnm
							if err := rt.binder.Reset(); err != nil {
								fmt.Println("Binder reset error:", err)
							}
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
	}.Create(rt.app.MainWindow)

	if err != nil {
		walk.MsgBox(rt.app.MainWindow, "Error", err.Error(), walk.MsgBoxIconError)
		return
	}

	dlg.Run()
}
