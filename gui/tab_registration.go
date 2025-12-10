package gui

import (
	"fmt"
	"mitsuscanner/driver"
	"strconv"
	"strings"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
)

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
	ModeVending    bool // ДОБАВЛЕНО

	// SNO (Taxation)
	TaxOSN   bool // 0
	TaxUSN   bool // 1
	TaxUSN_M bool // 2
	TaxENVD  bool // 3
	TaxESHN  bool // 4
	TaxPat   bool // 5
}

var regModel *RegViewModel
var regBinder *walk.DataBinder

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
					d.PushButton{Text: "Вычислить (CRC)", OnClicked: func() { walk.MsgBox(mw, "Info", "Тут будет расчет КПК", walk.MsgBoxIconInformation) }},
				},
			},

			// Основной контент
			d.Composite{
				Layout: d.VBox{Margins: d.Margins{Left: 8, Top: 8, Right: 8, Bottom: 2}, Spacing: 5},
				Children: []d.Widget{
					d.Composite{
						// Используем HBox для расположения блоков по горизонтали
						// Alignment: AlignTop прижмет маленький блок к верху
						Layout: d.HBox{Margins: d.Margins{Left: 8, Top: 2, Right: 8, Bottom: 2}, Spacing: 5, Alignment: d.Alignment2D(d.AlignDefault)},
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
								StretchFactor: 1,
								// MinSize можно задать, чтобы чекбоксы не сплющивались, если окно сузят
								MinSize: d.Size{Width: 150},
								Layout:  d.VBox{Margins: d.Margins{Left: 3, Top: 1, Right: 3, Bottom: 1}, Spacing: 1},
								Children: []d.Widget{
									d.CheckBox{Text: "ОСН", Checked: d.Bind("TaxOSN"), Alignment: d.Alignment2D(d.AlignNear)},
									d.CheckBox{Text: "УСН доход", Checked: d.Bind("TaxUSN"), Alignment: d.Alignment2D(d.AlignNear)},
									d.CheckBox{Text: "УСН доход - расход", Checked: d.Bind("TaxUSN_M"), Alignment: d.Alignment2D(d.AlignNear)},
									d.CheckBox{Text: "ЕСХН", Checked: d.Bind("TaxESHN"), Alignment: d.Alignment2D(d.AlignNear)},
									d.CheckBox{Text: "Патент", Checked: d.Bind("TaxPat"), Alignment: d.Alignment2D(d.AlignNear)},
								},
							},
						},
					},
					d.GroupBox{
						Title:  "Фискальные признаки",
						Layout: d.Grid{Columns: 4, Margins: d.Margins{Left: 3, Top: 1, Right: 3, Bottom: 1}, Spacing: 1},
						Children: []d.Widget{
							d.CheckBox{Text: "Автономный режим", Checked: d.Bind("ModeAutonomous")},
							d.CheckBox{Text: "Шифрование данных", Checked: d.Bind("ModeEncryption")},
							d.CheckBox{Text: "Расчеты за услуги", Checked: d.Bind("ModeService")},
							d.CheckBox{Text: "Расчеты в Интернет", Checked: d.Bind("ModeInternet")},
							d.CheckBox{Text: "Принтер в автомате", Checked: d.Bind("ModeAutomat")},
							d.CheckBox{Text: "Только БСО", Checked: d.Bind("ModeBSO")},
							d.CheckBox{Text: "Подакцизные товары", Checked: d.Bind("ModeExcise")},
							d.CheckBox{Text: "Проведение азартных игр", Checked: d.Bind("ModeGambling")},
							d.CheckBox{Text: "Проведение лотерей", Checked: d.Bind("ModeLottery")},
							d.CheckBox{Text: "Маркированные товары", Checked: d.Bind("ModeMarking")},
							d.CheckBox{Text: "Ломбард", Checked: d.Bind("ModePawn")},
							d.CheckBox{Text: "Страхование", Checked: d.Bind("ModeInsurance")},
							d.CheckBox{Text: "Общепит", Checked: d.Bind("ModeCatering")},
							d.CheckBox{Text: "Оптовая торговля", Checked: d.Bind("ModeWholesale")},
							d.CheckBox{Text: "Вендинг", Checked: d.Bind("ModeVending")},
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

func onReadRegistration() {
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		logMsg("=== НАЧАЛО ЧТЕНИЯ РЕГИСТРАЦИОННЫХ ДАННЫХ ===")
		regData, err := drv.GetRegistrationData()
		if err != nil {
			logMsg("ОШИБКА получения данных: %v", err)
			mw.Synchronize(func() { walk.MsgBox(mw, "Ошибка", err.Error(), walk.MsgBoxIconError) })
			return
		}

		logMsg("=== ДАННЫЕ ПОЛУЧЕНЫ ОТ ДРАЙВЕРА ===")
		logMsg("RNM: '%s'", regData.RNM)
		logMsg("INN: '%s'", regData.Inn)
		logMsg("OrgName: '%s'", regData.OrgName)
		logMsg("Address: '%s'", regData.Address)
		logMsg("Place: '%s'", regData.Place)
		logMsg("EmailSender: '%s'", regData.EmailSender)
		logMsg("Site: '%s'", regData.Site)
		logMsg("OfdName: '%s'", regData.OfdName)
		logMsg("OfdInn: '%s'", regData.OfdInn)

		logMsg("=== ФЛАГИ РЕЖИМОВ РАБОТЫ ===")
		logMsg("MarkAttr: '%s' -> %v", regData.MarkAttr, regData.MarkAttr == "1")
		logMsg("ExciseAttr: '%s' -> %v", regData.ExciseAttr, regData.ExciseAttr == "1")
		logMsg("InternetAttr: '%s' -> %v", regData.InternetAttr, regData.InternetAttr == "1")
		logMsg("ServiceAttr: '%s' -> %v", regData.ServiceAttr, regData.ServiceAttr == "1")
		logMsg("BsoAttr: '%s' -> %v", regData.BsoAttr, regData.BsoAttr == "1")
		logMsg("LotteryAttr: '%s' -> %v", regData.LotteryAttr, regData.LotteryAttr == "1")
		logMsg("GamblingAttr: '%s' -> %v", regData.GamblingAttr, regData.GamblingAttr == "1")
		logMsg("PawnAttr: '%s' -> %v", regData.PawnAttr, regData.PawnAttr == "1")
		logMsg("InsAttr: '%s' -> %v", regData.InsAttr, regData.InsAttr == "1")
		logMsg("DineAttr: '%s' -> %v", regData.DineAttr, regData.DineAttr == "1")
		logMsg("OptAttr: '%s' -> %v", regData.OptAttr, regData.OptAttr == "1")
		logMsg("VendAttr: '%s' -> %v", regData.VendAttr, regData.VendAttr == "1")
		logMsg("AutoModeAttr: '%s' -> %v", regData.AutoModeAttr, regData.AutoModeAttr == "1")
		logMsg("AutonomAttr: '%s' -> %v", regData.AutonomAttr, regData.AutonomAttr == "1")
		logMsg("EncryptAttr: '%s' -> %v", regData.EncryptAttr, regData.EncryptAttr == "1")
		logMsg("PrintAutoAttr: '%s' -> %v", regData.PrintAutoAttr, regData.PrintAutoAttr == "1")

		logMsg("=== СИСТЕМЫ НАЛОГООБЛОЖЕНИЯ ===")
		logMsg("TaxSystems: '%s'", regData.TaxSystems)

		mw.Synchronize(func() {
			logMsg("=== НАЧАЛО ЗАПОЛНЕНИЯ ПОЛЕЙ В GUI ===")

			regModel.RNM = regData.RNM
			regModel.INN = regData.Inn
			regModel.OrgName = regData.OrgName
			regModel.Address = regData.Address
			regModel.Place = regData.Place
			regModel.Email = regData.EmailSender
			regModel.Site = regData.Site
			regModel.OFDINN = regData.OfdInn
			regModel.OFDName = regData.OfdName

			logMsg("Основные поля заполнены:")
			logMsg("  RNM: '%s'", regModel.RNM)
			logMsg("  INN: '%s'", regModel.INN)
			logMsg("  OrgName: '%s'", regModel.OrgName)
			logMsg("  Address: '%s'", regModel.Address)
			logMsg("  Place: '%s'", regModel.Place)
			logMsg("  Email: '%s'", regModel.Email)
			logMsg("  Site: '%s'", regModel.Site)
			logMsg("  OFDINN: '%s'", regModel.OFDINN)
			logMsg("  OFDName: '%s'", regModel.OFDName)

			// Парсинг флагов из RegData
			regModel.ModeMarking = (regData.MarkAttr == "1")
			regModel.ModeExcise = (regData.ExciseAttr == "1")
			regModel.ModeInternet = (regData.InternetAttr == "1")
			regModel.ModeService = (regData.ServiceAttr == "1")
			regModel.ModeBSO = (regData.BsoAttr == "1")
			regModel.ModeLottery = (regData.LotteryAttr == "1")
			regModel.ModeGambling = (regData.GamblingAttr == "1")
			regModel.ModePawn = (regData.PawnAttr == "1")
			regModel.ModeInsurance = (regData.InsAttr == "1")
			regModel.ModeCatering = (regData.DineAttr == "1")
			regModel.ModeWholesale = (regData.OptAttr == "1")
			regModel.ModeVending = (regData.VendAttr == "1")
			regModel.ModeAutomat = (regData.PrintAutoAttr == "1" || regData.AutoModeAttr == "1")
			regModel.ModeAutonomous = (regData.AutonomAttr == "1")
			regModel.ModeEncryption = (regData.EncryptAttr == "1")

			logMsg("Флаги режимов установлены:")
			logMsg("  ModeMarking: %v (из '%s')", regModel.ModeMarking, regData.MarkAttr)
			logMsg("  ModeExcise: %v (из '%s')", regModel.ModeExcise, regData.ExciseAttr)
			logMsg("  ModeInternet: %v (из '%s')", regModel.ModeInternet, regData.InternetAttr)
			logMsg("  ModeService: %v (из '%s')", regModel.ModeService, regData.ServiceAttr)
			logMsg("  ModeBSO: %v (из '%s')", regModel.ModeBSO, regData.BsoAttr)
			logMsg("  ModeLottery: %v (из '%s')", regModel.ModeLottery, regData.LotteryAttr)
			logMsg("  ModeGambling: %v (из '%s')", regModel.ModeGambling, regData.GamblingAttr)
			logMsg("  ModePawn: %v (из '%s')", regModel.ModePawn, regData.PawnAttr)
			logMsg("  ModeInsurance: %v (из '%s')", regModel.ModeInsurance, regData.InsAttr)
			logMsg("  ModeCatering: %v (из '%s')", regModel.ModeCatering, regData.DineAttr)
			logMsg("  ModeWholesale: %v (из '%s')", regModel.ModeWholesale, regData.OptAttr)
			logMsg("  ModeVending: %v (из '%s')", regModel.ModeVending, regData.VendAttr)
			logMsg("  ModeAutomat: %v (из '%s' или '%s')", regModel.ModeAutomat, regData.PrintAutoAttr, regData.AutoModeAttr)
			logMsg("  ModeAutonomous: %v (из '%s')", regModel.ModeAutonomous, regData.AutonomAttr)
			logMsg("  ModeEncryption: %v (из '%s')", regModel.ModeEncryption, regData.EncryptAttr)

			// Парсинг СНО (ИСПРАВЛЕНИЕ: используем T1062 из RegData)
			// Строка вида "0,1,5"
			regModel.TaxOSN = false
			regModel.TaxUSN = false
			regModel.TaxUSN_M = false
			regModel.TaxENVD = false
			regModel.TaxESHN = false
			regModel.TaxPat = false

			taxParts := strings.Split(regData.TaxSystems, ",")
			logMsg("Парсинг систем налогообложения из '%s':", regData.TaxSystems)
			for _, t := range taxParts {
				trimmedT := strings.TrimSpace(t)
				switch trimmedT {
				case "0":
					regModel.TaxOSN = true
					logMsg("  Установлен TaxOSN = true")
				case "1":
					regModel.TaxUSN = true
					logMsg("  Установлен TaxUSN = true")
				case "2":
					regModel.TaxUSN_M = true
					logMsg("  Установлен TaxUSN_M = true")
				case "3":
					regModel.TaxENVD = true
					logMsg("  Установлен TaxENVD = true")
				case "4":
					regModel.TaxESHN = true
					logMsg("  Установлен TaxESHN = true")
				case "5":
					regModel.TaxPat = true
					logMsg("  Установлен TaxPat = true")
				default:
					logMsg("  Неизвестная система налогообложения: '%s'", trimmedT)
				}
			}

			logMsg("Системы налогообложения после парсинга:")
			logMsg("  TaxOSN: %v", regModel.TaxOSN)
			logMsg("  TaxUSN: %v", regModel.TaxUSN)
			logMsg("  TaxUSN_M: %v", regModel.TaxUSN_M)
			logMsg("  TaxENVD: %v", regModel.TaxENVD)
			logMsg("  TaxESHN: %v", regModel.TaxESHN)
			logMsg("  TaxPat: %v", regModel.TaxPat)

			// Используем Reset() для обновления UI из модели, а не Submit() который перезаписывает модель из UI
			logMsg("Вызов regBinder.Reset() для обновления UI...")
			if err := regBinder.Reset(); err != nil {
				logMsg("ОШИБКА обновления биндинга при загрузке данных: %v", err)
				walk.MsgBox(mw, "Ошибка биндинга", fmt.Sprintf("Ошибка обновления UI: %v", err), walk.MsgBoxIconError)
			} else {
				logMsg("regBinder.Reset() выполнен успешно!")
			}

			// Логи после Submit для проверки значений модели
			logMsg("Проверка значений модели ПОСЛЕ regBinder.Submit():")
			logMsg("  RNM: '%s'", regModel.RNM)
			logMsg("  INN: '%s'", regModel.INN)
			logMsg("  OrgName: '%s'", regModel.OrgName)
			logMsg("  Address: '%s'", regModel.Address)
			logMsg("  Place: '%s'", regModel.Place)
			logMsg("  Email: '%s'", regModel.Email)
			logMsg("  Site: '%s'", regModel.Site)
			logMsg("  OFDINN: '%s'", regModel.OFDINN)
			logMsg("  OFDName: '%s'", regModel.OFDName)
			logMsg("  ModeAutonomous: %v", regModel.ModeAutonomous)
			logMsg("  ModeEncryption: %v", regModel.ModeEncryption)
			logMsg("  TaxOSN: %v", regModel.TaxOSN)
			logMsg("  TaxUSN: %v", regModel.TaxUSN)

			logMsg("=== ДАННЫЕ РЕГИСТРАЦИИ ОБРАБОТАНЫ ===")
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

	if len(regModel.INN) != 10 && len(regModel.INN) != 12 {
		walk.MsgBox(mw, "Ошибка", "Некорректная длина ИНН", walk.MsgBoxIconError)
		return
	}

	req := fillRequestFromModel()

	go func() {
		if err := drv.SetCashier("Администратор", ""); err != nil {
			logMsg("Ошибка установки кассира: %v", err)
			return
		}
		if err := drv.Register(req); err != nil {
			mw.Synchronize(func() { walk.MsgBox(mw, "Ошибка регистрации", err.Error(), walk.MsgBoxIconError) })
		} else {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Успех", "ККТ успешно зарегистрирована!", walk.MsgBoxIconInformation)
			})
			logMsg("Регистрация выполнена успешно.")
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

	req := fillRequestFromModel()

	go func() {
		if err := drv.SetCashier("Администратор", ""); err != nil {
			logMsg("Ошибка установки кассира: %v", err)
			return
		}
		if err := drv.Reregister(req, reasons); err != nil {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Ошибка перерегистрации", err.Error(), walk.MsgBoxIconError)
			})
		} else {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Успех", "ККТ перерегистрирована!", walk.MsgBoxIconInformation)
			})
		}
	}()
}

func onSelectReasons() {
	// 1. Передаем текущее значение внутрь
	reasons, ok := RunReasonDialog(mw, regModel.Reasons)
	if ok {
		// 2. Обновляем модель
		regModel.Reasons = reasons

		// 3. Принудительно обновляем UI через Binder
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

	req := fillRequestFromModel()
	reasons := []int{1}

	go func() {
		if err := drv.SetCashier("Администратор", ""); err != nil {
			logMsg("Ошибка установки кассира: %v", err)
			return
		}
		if err := drv.Reregister(req, reasons); err != nil {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Ошибка замены ФН", err.Error(), walk.MsgBoxIconError)
			})
		} else {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Успех", "ФН заменен!", walk.MsgBoxIconInformation)
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
		if err := drv.CloseFiscalArchive(); err != nil {
			mw.Synchronize(func() { walk.MsgBox(mw, "Ошибка", err.Error(), walk.MsgBoxIconError) })
		} else {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Успех", "Фискальный архив закрыт.", walk.MsgBoxIconInformation)
			})
		}
	}()
}

func fillRequestFromModel() driver.RegistrationRequest {
	req := driver.RegistrationRequest{
		RNM:            regModel.RNM,
		Inn:            regModel.INN,
		OrgName:        regModel.OrgName,
		Address:        regModel.Address,
		Place:          regModel.Place,
		SenderEmail:    regModel.Email,
		FnsSite:        regModel.Site,
		FfdVer:         regModel.FFD,
		OfdName:        regModel.OFDName,
		OfdInn:         regModel.OFDINN,
		AutonomousMode: regModel.ModeAutonomous,
		Encryption:     regModel.ModeEncryption,
		Service:        regModel.ModeService,
		InternetCalc:   regModel.ModeInternet,
		BSO:            regModel.ModeBSO,
		Gambling:       regModel.ModeGambling,
		Lottery:        regModel.ModeLottery,
		Excise:         regModel.ModeExcise,
		Marking:        regModel.ModeMarking,
		PawnShop:       regModel.ModePawn,
		Insurance:      regModel.ModeInsurance,
		Catering:       regModel.ModeCatering,
		Wholesale:      regModel.ModeWholesale,
		Vending:        regModel.ModeVending,
		PrinterAutomat: regModel.ModeAutomat,
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
