package gui

import (
	"fmt"
	"html"
	"mitsuscanner/driver"
	"regexp"
	"strconv"
	"strings"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
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

var regModel *RegViewModel
var regBinder *walk.DataBinder
var fnPhaseLabel *walk.Label

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

// GetRegistrationTab возвращает описание вкладки "Регистрация"
func GetRegistrationTab() d.TabPage {
	regModel = &RegViewModel{} // Default

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
					d.PushButton{Text: "Вычислить (CRC)", OnClicked: onCalcRNM},
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
											d.Label{AssignTo: &fnPhaseLabel, Text: d.Bind("FnPhaseText")},
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
												OnClicked:   onSendToOfd,
												MinSize:     d.Size{Width: 110},
												ToolTipText: "Отправить первый неотправленный документ в ОФД",
											},
											d.PushButton{
												Text:        "↻", // Unicode символ обновления
												OnClicked:   onRefreshFnInfo,
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
										OnClicked: onSelectReasons,
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
					d.PushButton{Text: "Считать из ККТ", OnClicked: onReadRegistration},
					d.HSpacer{},
					d.PushButton{Text: "Закрытие ФН", OnClicked: onCloseFn},
					d.PushButton{Text: "Замена ФН", OnClicked: onReplaceFn},
					d.PushButton{Text: "Перерегистрация", OnClicked: onReregister},
					d.PushButton{Text: "Регистрация", OnClicked: onRegister},
				},
			},
		},
		DataBinder: d.DataBinder{
			AssignTo:       &regBinder,
			DataSource:     regModel,
			ErrorPresenter: d.ToolTipErrorPresenter{},
		},
	}
}

// --- Обработчики событий ---

// onCalcRNM открывает диалог генерации РНМ
func onCalcRNM() {
	// Сначала забираем данные из формы в модель
	if err := regBinder.Submit(); err != nil {
		return
	}

	inn := strings.TrimSpace(regModel.INN)
	if len(inn) != 10 && len(inn) != 12 {
		walk.MsgBox(mw, "Ошибка", "Для расчета РНМ заполните корректный ИНН пользователя (10 или 12 цифр).", walk.MsgBoxIconError)
		return
	}

	// Пытаемся получить заводской номер из драйвера
	serial := ""
	if driver.Active != nil {
		info, err := driver.Active.GetFiscalInfo()
		if err == nil && info != nil {
			serial = info.SerialNumber
		}
	}

	// Запускаем диалог
	runRnmGenerationDialog(inn, serial)
}

func onReadRegistration() {
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		regData, err := drv.GetRegistrationData()
		if err != nil {
			mw.Synchronize(func() { walk.MsgBox(mw, "Ошибка", err.Error(), walk.MsgBoxIconError) })
			return
		}

		fnStatus, fnErr := drv.GetFnStatus()

		mw.Synchronize(func() {

			regModel.RNM = regData.RNM
			regModel.INN = regData.Inn
			regModel.OrgName = regData.OrgName
			regModel.Address = regData.Address
			regModel.Place = regData.Place
			regModel.Email = regData.EmailSender
			regModel.Site = regData.Site
			regModel.OFDINN = regData.OfdInn
			regModel.OFDName = html.UnescapeString(regData.OfdName)

			// --- Парсинг атрибутов режимов работы ---

			// Парсинг MODE и ExtMODE с использованием hasBit
			modeInt := int(regData.ModeMask)
			extModeInt := int(regData.ExtModeMask)

			regModel.ModeEncryption = hasBit(modeInt, 0) // Шифрование
			regModel.ModeAutonomous = hasBit(modeInt, 1) // Автономный режим
			regModel.ModeService = hasBit(modeInt, 3)    // Услуги
			regModel.ModeBSO = hasBit(modeInt, 4)        // БСО
			regModel.ModeInternet = hasBit(modeInt, 5)   // Интернет
			regModel.ModeCatering = hasBit(modeInt, 6)   // Общепит
			regModel.ModeWholesale = hasBit(modeInt, 7)  // Опт

			regModel.ModeExcise = hasBit(extModeInt, 0)    // Подакцизные
			regModel.ModeGambling = hasBit(extModeInt, 1)  // Азартные
			regModel.ModeLottery = hasBit(extModeInt, 2)   // Лотереи
			regModel.ModeAutomat = hasBit(extModeInt, 3)   // Принтер в автомате
			regModel.ModeMarking = hasBit(extModeInt, 4)   // Маркированные товары
			regModel.ModePawn = hasBit(extModeInt, 5)      // Ломбард
			regModel.ModeInsurance = hasBit(extModeInt, 6) // Страхование
			regModel.ModeVending = hasBit(extModeInt, 7)   // Вендинг

			// Парсинг СНО
			regModel.TaxOSN = false
			regModel.TaxUSN = false
			regModel.TaxUSN_M = false
			regModel.TaxENVD = false
			regModel.TaxESHN = false
			regModel.TaxPat = false

			taxParts := strings.Split(regData.TaxSystems, ",")
			for _, t := range taxParts {
				trimmedT := strings.TrimSpace(t)
				switch trimmedT {
				case "0":
					regModel.TaxOSN = true
				case "1":
					regModel.TaxUSN = true
				case "2":
					regModel.TaxUSN_M = true
				case "3":
					regModel.TaxENVD = true
				case "4":
					regModel.TaxESHN = true
				case "5":
					regModel.TaxPat = true
				}
			}

			// Установка TaxSystemBase на первую зарегистрированную СНО по умолчанию
			if regModel.TaxOSN {
				regModel.TaxSystemBase = "0"
			} else if regModel.TaxUSN {
				regModel.TaxSystemBase = "1"
			} else if regModel.TaxUSN_M {
				regModel.TaxSystemBase = "2"
			} else if regModel.TaxESHN {
				regModel.TaxSystemBase = "4"
			} else if regModel.TaxPat {
				regModel.TaxSystemBase = "5"
			}

			// Если есть базовая СНО, установить её
			if regData.TaxBase != "" {
				regModel.TaxSystemBase = regData.TaxBase
			}

			if err := regBinder.Reset(); err != nil {
				walk.MsgBox(mw, "Ошибка биндинга", fmt.Sprintf("Ошибка обновления UI: %v", err), walk.MsgBoxIconError)
			}

			// Читаем информацию о ФН
			if fnErr == nil {
				regModel.FnNumber = fnStatus.Serial
				regModel.FnValidDate = fnStatus.Valid
				regModel.FnPhase = fnStatus.Phase

				// Декодируем фазу
				phaseText, phaseColor := decodeFnPhase(fnStatus.Phase)
				regModel.FnPhaseText = phaseText
				regModel.FnPhaseColor = phaseColor
			} else {
				regModel.FnNumber = "Ошибка чтения"
				regModel.FnPhaseText = "—"
				regModel.FnValidDate = "—"
			}

			if err := regBinder.Reset(); err != nil {
				walk.MsgBox(mw, "Ошибка биндинга", fmt.Sprintf("Ошибка обновления UI: %v", err), walk.MsgBoxIconError)
			}

			if fnErr == nil && fnPhaseLabel != nil {
				_, phaseColor := decodeFnPhase(fnStatus.Phase)
				fnPhaseLabel.SetTextColor(phaseColor)
			}
		})
	}()
}

func onRegister() {
	drv := driver.Active
	if drv == nil {
		return
	}
	if err := regBinder.Submit(); err != nil {
		return
	}

	if !regexp.MustCompile(`^\d+$`).MatchString(regModel.INN) || (len(regModel.INN) != 10 && len(regModel.INN) != 12) {
		walk.MsgBox(mw, "Ошибка", "ИНН должен состоять только из цифр и иметь длину 10 или 12 символов.", walk.MsgBoxIconError)
		return
	}

	if strings.TrimSpace(regModel.OrgName) == "" {
		walk.MsgBox(mw, "Ошибка", "Поле 'Наименование' обязательно для заполнения.", walk.MsgBoxIconError)
		return
	}

	if strings.TrimSpace(regModel.Address) == "" {
		walk.MsgBox(mw, "Ошибка", "Поле 'Адрес расчетов' обязательно для заполнения.", walk.MsgBoxIconError)
		return
	}

	if strings.TrimSpace(regModel.Place) == "" {
		walk.MsgBox(mw, "Ошибка", "Поле 'Место расчетов' обязательно для заполнения.", walk.MsgBoxIconError)
		return
	}

	req := fillRequestFromModel(false)

	go func() {
		if err := drv.SetCashier("Администратор", ""); err != nil {
			return
		}
		resp, err := drv.Register(req)
		if err != nil {
			mw.Synchronize(func() { walk.MsgBox(mw, "Ошибка регистрации", err.Error(), walk.MsgBoxIconError) })
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
			mw.Synchronize(func() { walk.MsgBox(mw, "Ошибка", err.Error(), walk.MsgBoxIconError) })
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
			mw.Synchronize(func() { RunReportModal(mw, meta) })
		}
	}()
}

func onReregister() {
	drv := driver.Active
	if drv == nil {
		return
	}
	if err := regBinder.Submit(); err != nil {
		return
	}

	if regModel.Reasons == "" {
		walk.MsgBox(mw, "Ошибка", "Не выбраны причины перерегистрации", walk.MsgBoxIconError)
		return
	}

	var reasons []int
	parts := strings.Split(regModel.Reasons, ",")
	for _, p := range parts {
		code, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			walk.MsgBox(mw, "Ошибка", "Некорректный формат причин", walk.MsgBoxIconError)
			return
		}
		reasons = append(reasons, code)
	}

	if !regexp.MustCompile(`^\d+$`).MatchString(regModel.INN) || (len(regModel.INN) != 10 && len(regModel.INN) != 12) {
		walk.MsgBox(mw, "Ошибка", "ИНН должен состоять только из цифр и иметь длину 10 или 12 символов.", walk.MsgBoxIconError)
		return
	}

	if strings.TrimSpace(regModel.OrgName) == "" {
		walk.MsgBox(mw, "Ошибка", "Поле 'Наименование' обязательно для заполнения.", walk.MsgBoxIconError)
		return
	}

	if strings.TrimSpace(regModel.Address) == "" {
		walk.MsgBox(mw, "Ошибка", "Поле 'Адрес расчетов' обязательно для заполнения.", walk.MsgBoxIconError)
		return
	}

	if strings.TrimSpace(regModel.Place) == "" {
		walk.MsgBox(mw, "Ошибка", "Поле 'Место расчетов' обязательно для заполнения.", walk.MsgBoxIconError)
		return
	}

	req := fillRequestFromModel(true)

	go func() {
		if err := drv.SetCashier("Администратор", ""); err != nil {
			return
		}
		_, err := drv.Reregister(req, reasons)
		if err != nil {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Ошибка перерегистрации", err.Error(), walk.MsgBoxIconError)
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
			mw.Synchronize(func() { walk.MsgBox(mw, "Ошибка", err.Error(), walk.MsgBoxIconError) })
			return
		}
		meta.Data = regData
		if meta.Kind == driver.ReportReg || meta.Kind == driver.ReportRereg {
			mw.Synchronize(func() { RunReportModal(mw, meta) })
		}
	}()
}

func onSelectReasons() {
	if err := regBinder.Submit(); err != nil {
		return
	}
	reasons, ok := RunReasonDialog(mw, regModel.Reasons)
	if ok {
		regModel.Reasons = reasons
		if err := regBinder.Reset(); err != nil {
			fmt.Println("Binder reset error:", err)
		}
	}
}

func onReplaceFn() {
	drv := driver.Active
	if drv == nil {
		return
	}
	if err := regBinder.Submit(); err != nil {
		return
	}

	req := fillRequestFromModel(false)
	reasons := []int{1}

	go func() {
		if err := drv.SetCashier("Администратор", ""); err != nil {
			return
		}
		resp, err := drv.Reregister(req, reasons)
		if err != nil {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Ошибка замены ФН", err.Error(), walk.MsgBoxIconError)
			})
		} else {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Успех", fmt.Sprintf("ФН заменен!\nФД: %s\nФП: %s", resp.FdNumber, resp.FpNumber), walk.MsgBoxIconInformation)
			})
		}
	}()
}

func onCloseFn() {
	drv := driver.Active
	if drv == nil {
		return
	}
	if walk.MsgBox(mw, "Подтверждение", "Вы действительно хотите закрыть фискальный архив?\nЭто необратимая операция!", walk.MsgBoxYesNo|walk.MsgBoxIconWarning) != walk.DlgCmdYes {
		return
	}
	go func() {
		// 1. Закрытие ФН (включает PRINT)
		result, err := drv.CloseFiscalArchive()
		if err != nil {
			mw.Synchronize(func() { walk.MsgBox(mw, "Ошибка", err.Error(), walk.MsgBoxIconError) })
			return
		}

		// 2. Получение типа документа для метаданных отчёта
		typeCode, _ := drv.GetCurrentDocumentType()
		meta := driver.GetReportMeta(typeCode)

		// 3. Сбор данных для отчёта через хелпер
		reportData, err := GetCloseFnReportData(drv, result.FD, result.FP)
		if err != nil {
			mw.Synchronize(func() { walk.MsgBox(mw, "Ошибка", err.Error(), walk.MsgBoxIconError) })
			return
		}

		// 4. Отображение отчёта
		meta.Data = reportData
		mw.Synchronize(func() { RunReportModal(mw, meta) })
	}()
}

// onSendToOfd отправляет первый неотправленный документ в ОФД
func onSendToOfd() {
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		result, err := SendFirstUnsentDocument(drv)
		mw.Synchronize(func() {
			if err != nil {
				walk.MsgBox(mw, "Ошибка", err.Error(), walk.MsgBoxIconError)
				return
			}

			if result.Success {
				walk.MsgBox(mw, "Успех",
					fmt.Sprintf("Отправлено документов: %d", result.DocumentsSent),
					walk.MsgBoxIconInformation)
			} else {
				walk.MsgBox(mw, "Информация", result.ErrorMessage, walk.MsgBoxIconWarning)
			}

			// Обновляем информацию о ФН
			onRefreshFnInfo()
		})
	}()
}

// onRefreshFnInfo обновляет информацию о ФН
func onRefreshFnInfo() {
	drv := driver.Active
	if drv == nil {
		return
	}

	go func() {
		err := RefreshFnInfo(drv)
		if err != nil {
			mw.Synchronize(func() {
				logMsg("Ошибка обновления ФН: %v", err)
			})
		}
	}()
}

func fillRequestFromModel(isRereg bool) driver.RegistrationRequest {
	req := driver.RegistrationRequest{
		IsReregistration: isRereg,
		RNM:              regModel.RNM,
		Inn:              regModel.INN,
		OrgName:          regModel.OrgName,
		Address:          regModel.Address,
		Place:            regModel.Place,
		SenderEmail:      regModel.Email,
		FnsSite:          regModel.Site,
		FfdVer:           regModel.FFD,
		OfdName:          regModel.OFDName,
		OfdInn:           regModel.OFDINN,
		AutonomousMode:   regModel.ModeAutonomous,
		Encryption:       regModel.ModeEncryption,
		Service:          regModel.ModeService,
		InternetCalc:     regModel.ModeInternet,
		BSO:              regModel.ModeBSO,
		Gambling:         regModel.ModeGambling,
		Lottery:          regModel.ModeLottery,
		Excise:           regModel.ModeExcise,
		Marking:          regModel.ModeMarking,
		PawnShop:         regModel.ModePawn,
		Insurance:        regModel.ModeInsurance,
		Catering:         regModel.ModeCatering,
		Wholesale:        regModel.ModeWholesale,
		Vending:          regModel.ModeVending,
		PrinterAutomat:   regModel.ModeAutomat,
	}

	var taxCodes []string
	if regModel.TaxOSN {
		taxCodes = append(taxCodes, "0")
	}
	if regModel.TaxUSN {
		taxCodes = append(taxCodes, "1")
	}
	if regModel.TaxUSN_M {
		taxCodes = append(taxCodes, "2")
	}
	if regModel.TaxENVD {
		taxCodes = append(taxCodes, "3")
	}
	if regModel.TaxESHN {
		taxCodes = append(taxCodes, "4")
	}
	if regModel.TaxPat {
		taxCodes = append(taxCodes, "5")
	}
	req.TaxSystems = strings.Join(taxCodes, ",")

	return req
}

// --- Логика генерации РНМ (CRC) ---

// runRnmGenerationDialog запускает диалоговое окно для расчета РНМ
func runRnmGenerationDialog(inn, serial string) {
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

							// Расчет
							rnm, err := calculateRNM(dlgModel.OrderNum, inn, dlgModel.Serial)
							if err != nil {
								walk.MsgBox(dlg, "Ошибка расчета", err.Error(), walk.MsgBoxIconError)
								return
							}

							// Применяем результат
							regModel.RNM = rnm
							if err := regBinder.Reset(); err != nil {
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
	}.Create(mw)

	if err != nil {
		walk.MsgBox(mw, "Error", err.Error(), walk.MsgBoxIconError)
		return
	}

	dlg.Run()
}

// calculateRNM выполняет расчет РНМ по алгоритму CRC16-CCITT.
// Формат входных данных для CRC:
// Pad(Order, 10) + Pad(INN, 12) + Pad(Serial, 20)
// Результат: Pad(Order, 10) + Pad(CRC, 6)
func calculateRNM(orderNum, inn, serial string) (string, error) {
	// 1. Формируем строку для расчета
	paddedOrder := padLeft(orderNum, 10, '0')
	paddedInn := padLeft(inn, 12, '0')
	paddedSerial := padLeft(serial, 20, '0')

	// Строка: 0000000001 + 007804437548 + 00000000000000000156 (пример)
	calcString := paddedOrder + paddedInn + paddedSerial

	// 2. Считаем CRC
	crc := crc16ccitt([]byte(calcString))

	// 3. Формируем хвост (CRC дополненный до 6 цифр нулями)
	// Пример: CRC 33271 -> "033271"
	crcStr := fmt.Sprintf("%d", crc)
	paddedCrc := padLeft(crcStr, 6, '0')

	// 4. Итоговый РНМ
	finalRnm := paddedOrder + paddedCrc

	return finalRnm, nil
}

// crc16ccitt вычисляет CRC-16 (CCITT False)
// Poly: 0x1021, Init: 0xFFFF
func crc16ccitt(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if (crc & 0x8000) != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// padLeft дополняет строку символом padChar слева до длины length
func padLeft(s string, length int, padChar byte) string {
	if len(s) >= length {
		return s // Или обрезать, если требуется строго length
	}
	padding := make([]byte, length-len(s))
	for i := range padding {
		padding[i] = padChar
	}
	return string(padding) + s
}
