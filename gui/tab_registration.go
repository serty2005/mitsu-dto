package main

import (
	"strings"

	"mitsuscanner/mitsu"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
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

	OFDName string
	OFDINN  string

	// Checkboxes (Settings)
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
func GetRegistrationTab() TabPage {
	regModel = &RegViewModel{FFD: "1.2"} // Default

	return TabPage{
		Title:  "Регистрация",
		Layout: VBox{},
		Children: []Widget{
			// Верхняя панель
			Composite{
				Layout: Grid{Columns: 4},
				Children: []Widget{
					Label{Text: "Регистрационный номер ККТ (РНМ):"},
					LineEdit{Text: Bind("RNM"), MinSize: Size{Width: 150}},
					PushButton{Text: "Вычислить (CRC)", OnClicked: func() { walk.MsgBox(mw, "Info", "Тут будет расчет КПК", walk.MsgBoxIconInformation) }},
					Label{Text: ""},

					Label{Text: "Версия ФФД:"},
					ComboBox{
						Value:   Bind("FFD"),
						Model:   []string{"1.05", "1.1", "1.2"},
						MinSize: Size{Width: 100},
					},
				},
			},

			// Основной контент
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					// Левая колонка
					Composite{
						Layout: VBox{},
						Children: []Widget{
							GroupBox{
								Title:  "Реквизиты организации",
								Layout: Grid{Columns: 2},
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
								Title:  "Настройки ККТ",
								Layout: Grid{Columns: 2},
								Children: []Widget{
									CheckBox{Text: "Автономный режим", Checked: Bind("ModeAutonomous")},
									CheckBox{Text: "Только БСО", Checked: Bind("ModeBSO")},
									CheckBox{Text: "Шифрование данных", Checked: Bind("ModeEncryption")},
									CheckBox{Text: "Подакцизные товары", Checked: Bind("ModeExcise")},
									CheckBox{Text: "Расчеты за услуги", Checked: Bind("ModeService")},
									CheckBox{Text: "Проведение азартных игр", Checked: Bind("ModeGambling")},
									CheckBox{Text: "Расчеты в Интернет", Checked: Bind("ModeInternet")},
									CheckBox{Text: "Проведение лотерей", Checked: Bind("ModeLottery")},
									CheckBox{Text: "Принтер в автомате", Checked: Bind("ModeAutomat")},
									CheckBox{Text: "Маркированные товары", Checked: Bind("ModeMarking")},
								},
							},
						},
					},

					// Правая колонка
					Composite{
						Layout: VBox{},
						Children: []Widget{
							GroupBox{
								Title:  "Системы налогообложения",
								Layout: VBox{},
								Children: []Widget{
									CheckBox{Text: "ОСН", Checked: Bind("TaxOSN")},
									CheckBox{Text: "УСН доход", Checked: Bind("TaxUSN")},
									CheckBox{Text: "УСН доход - расход", Checked: Bind("TaxUSN_M")},
									CheckBox{Text: "ЕСХН", Checked: Bind("TaxESHN")},
									CheckBox{Text: "Патент", Checked: Bind("TaxPat")},
								},
							},
							GroupBox{
								Title:  "Сферы деятельности (Доп)",
								Layout: VBox{},
								Children: []Widget{
									CheckBox{Text: "Ломбард", Checked: Bind("ModePawn")},
									CheckBox{Text: "Страхование", Checked: Bind("ModeInsurance")},
									CheckBox{Text: "Общепит", Checked: Bind("ModeCatering")},
									CheckBox{Text: "Оптовая торговля", Checked: Bind("ModeWholesale")},
									CheckBox{Text: "Вендинг", Checked: Bind("ModeVending")},
								},
							},
						},
					},
				},
			},

			// ОФД
			GroupBox{
				Title:  "Оператор фискальных данных",
				Layout: Grid{Columns: 4},
				Children: []Widget{
					Label{Text: "ИНН ОФД:"},
					LineEdit{Text: Bind("OFDINN")},
					Label{Text: "Наименование ОФД:"},
					LineEdit{Text: Bind("OFDName")},
				},
			},

			// Кнопки
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{Text: "Считать из ККТ", OnClicked: onReadRegistration},
					HSpacer{},
					PushButton{Text: "Закрытие ФН", OnClicked: onCloseFn},
					PushButton{Text: "Замена ФН / Перерегистрация", OnClicked: onReregister},
					PushButton{Text: "Регистрация", OnClicked: onRegister},
				},
			},
		},
		DataBinder: DataBinder{
			AssignTo:       &regBinder,
			DataSource:     regModel,
			ErrorPresenter: ToolTipErrorPresenter{},
		},
	}
}

func onReadRegistration() {
	if driver == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения к ККТ", walk.MsgBoxIconError)
		return
	}

	go func() {
		regData, err := driver.GetRegistrationData()
		if err != nil {
			mw.Synchronize(func() { walk.MsgBox(mw, "Ошибка", err.Error(), walk.MsgBoxIconError) })
			return
		}

		mw.Synchronize(func() {
			regModel.RNM = regData.RNM
			regModel.INN = regData.Inn
			regModel.OrgName = regData.OrgName
			regModel.Address = regData.Address
			regModel.Place = regData.Place
			regModel.Email = regData.EmailSender
			regModel.Site = regData.Site
			regModel.FFD = regData.FfdVer
			regModel.OFDINN = regData.OfdInn
			regModel.OFDName = regData.OfdName

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

			// Парсинг СНО (ИСПРАВЛЕНИЕ: используем T1062 из RegData)
			// Строка вида "0,1,5"
			regModel.TaxOSN = false
			regModel.TaxUSN = false
			regModel.TaxUSN_M = false
			regModel.TaxENVD = false
			regModel.TaxESHN = false
			regModel.TaxPat = false

			taxParts := strings.Split(regData.TaxSystems, ",")
			for _, t := range taxParts {
				switch strings.TrimSpace(t) {
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

			regBinder.Reset()
			logMsg("Данные регистрации считаны успешно.")
		})
	}()
}

func onRegister() {
	if driver == nil {
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
		if err := driver.SetCashier("Администратор", ""); err != nil {
			logMsg("Ошибка установки кассира: %v", err)
			return
		}
		if err := driver.Register(req); err != nil {
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
	if driver == nil {
		return
	}
	if err := regBinder.Submit(); err != nil {
		return
	}

	res := walk.MsgBox(mw, "Причина", "Да - Замена ФН\nНет - Смена реквизитов", walk.MsgBoxYesNoCancel)
	var reasons []int
	if res == walk.DlgCmdYes {
		reasons = []int{1}
	} else if res == walk.DlgCmdNo {
		reasons = []int{3}
	} else {
		return
	}

	req := fillRequestFromModel()

	go func() {
		if err := driver.SetCashier("Администратор", ""); err != nil {
			logMsg("Ошибка установки кассира: %v", err)
			return
		}
		if err := driver.Reregister(req, reasons); err != nil {
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

func onCloseFn() {
	if driver == nil {
		return
	}
	if walk.MsgBox(mw, "Подтверждение", "Вы действительно хотите закрыть фискальный архив?\nЭто необратимая операция!", walk.MsgBoxYesNo|walk.MsgBoxIconWarning) != walk.DlgCmdYes {
		return
	}
	go func() {
		if err := driver.CloseFiscalArchive(); err != nil {
			mw.Synchronize(func() { walk.MsgBox(mw, "Ошибка", err.Error(), walk.MsgBoxIconError) })
		} else {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Успех", "Фискальный архив закрыт.", walk.MsgBoxIconInformation)
			})
		}
	}()
}

func fillRequestFromModel() mitsu.RegistrationRequest {
	req := mitsu.RegistrationRequest{
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
