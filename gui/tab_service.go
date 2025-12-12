package gui

import (
	"fmt"
	"mitsuscanner/driver"
	"net"
	"strconv"
	"time"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
)

// -----------------------------
// ВСПОМОГАТЕЛЬНЫЕ СТРУКТУРЫ
// -----------------------------

// NV (Name-Value) используется для заполнения ComboBox
type NV struct {
	Name string
	Code string
}

// Списки значений для выпадающих меню
var (
	// ОФД Клиент
	listClients = []*NV{
		{"Внешний (ПК/Служба)", "1"},
		{"Встроенный (LAN)", "0"},
	}
	// Скорость
	listBaud = []*NV{
		{"9600", "9600"}, {"19200", "19200"}, {"38400", "38400"}, {"57600", "57600"}, {"115200", "115200"},
	}
	// Модель принтера
	listModels = []*NV{
		{"RP-809", "1"}, {"F80", "2"},
	}
	// Позиция QR (b1)
	listQRPos = []*NV{
		{"Слева", "0"}, {"По центру", "1"}, {"Справа", "2"},
	}
	// Округление (b2)
	listRounding = []*NV{
		{"Нет", "0"}, {"0.10", "1"}, {"0.50", "2"}, {"1.00", "3"},
	}
	// Триггер ящика (b5)
	listDrawerTrig = []*NV{
		{"Нет", "0"}, {"Наличные", "1"}, {"Безнал", "2"}, {"Всегда", "3"},
	}
	// Часовые пояса (упрощенно)
	listTimezones = []*NV{
		{"UTC+2 (Калининград)", "2"}, {"UTC+3 (Москва)", "3"}, {"UTC+4 (Самара)", "4"},
		{"UTC+5 (Екатеринбург)", "5"}, {"UTC+6 (Омск)", "6"}, {"UTC+7 (Красноярск)", "7"},
		{"UTC+8 (Иркутск)", "8"}, {"UTC+9 (Якутск)", "9"}, {"UTC+10 (Владивосток)", "10"},
		{"UTC+11 (Магадан)", "11"}, {"UTC+12 (Камчатка)", "12"},
	}
)

// -----------------------------
// МОДЕЛИ ТАБЛИЦ (Клише)
// -----------------------------

type HeaderModel struct {
	walk.TableModelBase
	items []map[string]interface{}
}

func (m *HeaderModel) RowCount() int {
	return len(m.items)
}

func (m *HeaderModel) Value(row, col int) interface{} {
	return m.items[row]["Text"]
}

func (m *HeaderModel) SetValue(row, col int, value interface{}) error {
	m.items[row]["Text"] = value.(string)
	return nil
}

type FooterModel struct {
	walk.TableModelBase
	items []map[string]interface{}
}

func (m *FooterModel) RowCount() int {
	return len(m.items)
}

func (m *FooterModel) Value(row, col int) interface{} {
	return m.items[row]["Text"]
}

func (m *FooterModel) SetValue(row, col int, value interface{}) error {
	m.items[row]["Text"] = value.(string)
	return nil
}

// -----------------------------
// МОДЕЛЬ ВЬЮ (ViewModel)
// -----------------------------

type ServiceViewModel struct {
	// Время
	KktTimeStr string
	PcTimeStr  string

	// --- Связь и ОФД ---
	OfdString string // Адрес:Порт
	OfdClient string // "0" или "1"
	TimerFN   int
	TimerOFD  int

	OismString string // Адрес:Порт

	// LAN
	LanAddr string
	LanPort int
	LanMask string
	LanDns  string
	LanGw   string

	// --- Параметры (Оборудование и Опции) ---

	// Принтер
	PrintModel string // "1", "2"
	PrintBaud  string // "115200"
	PrintPaper int    // 57, 80
	PrintFont  int    // 0, 1

	// Опции (b0..b9) и прочее
	OptTimezone     string // Строка для маппинга
	OptCut          bool   // b3
	OptAutoTest     bool   // b4
	OptNearEnd      bool   // b6
	OptTextQR       bool   // b7
	OptCountInCheck bool   // b8

	OptQRPos      string // b1
	OptRounding   string // b2
	OptDrawerTrig string // b5

	// Денежный ящик (Настройки физики)
	DrawerPin  int
	DrawerRise int
	DrawerFall int

	// Клише
	HeaderLines []map[string]interface{}
	FooterLines []map[string]interface{}

	HeaderModel *HeaderModel
	FooterModel *FooterModel

	// Изображения
	ImgIndex int
	ImgPath  string
}

var serviceModel *ServiceViewModel
var serviceBinder *walk.DataBinder

// Виджеты для прямого доступа
var kktTimeLabel *walk.Label
var pcTimeLabel *walk.Label

// Таймеры
var pcTicker *time.Ticker
var kktTicker *time.Ticker

// Флаг, предотвращающий срабатывание событий при программной загрузке данных
var isLoadingData bool

// -----------------------------
// УТИЛИТЫ
// -----------------------------

func splitHostPort(full string) (string, int) {
	if full == "" {
		return "", 0
	}
	host, portStr, err := net.SplitHostPort(full)
	if err != nil {
		return full, 0
	}
	port, _ := strconv.Atoi(portStr)
	return host, port
}

func joinHostPort(host string, port int) string {
	if port == 0 {
		return host
	}
	return fmt.Sprintf("%s:%d", host, port)
}

// -----------------------------
// ЗАГРУЗКА ДАННЫХ
// -----------------------------

func loadServiceInitial() {
	waitForWindow := func() bool {
		for i := 0; i < 20; i++ {
			if mw != nil && mw.Handle() != 0 {
				return true
			}
			time.Sleep(100 * time.Millisecond)
		}
		return false
	}

	drv := driver.Active
	if drv == nil {
		serviceModel.KktTimeStr = "Нет подключения"
		serviceModel.PcTimeStr = time.Now().Format("02.01.2006 15:04:05")
		return
	}

	// 1. Время
	go func() {
		if !waitForWindow() {
			return
		}
		t, err := drv.GetDateTime()
		mw.Synchronize(func() {
			if err == nil {
				serviceModel.KktTimeStr = t.Format("02.01.2006 15:04:05")
			} else {
				serviceModel.KktTimeStr = "Ошибка"
			}
			serviceBinder.Reset()
		})
	}()

	// 2. Остальные настройки
	go func() {
		if !waitForWindow() {
			return
		}
		mw.Synchronize(func() { isLoadingData = true })
		readNetworkSettings()
		readAllParameters()
		readCliche()
		mw.Synchronize(func() { isLoadingData = false })
	}()
}

func readNetworkSettings() {
	drv := driver.Active
	if drv == nil {
		return
	}

	ofd, _ := drv.GetOfdSettings()
	oism, _ := drv.GetOismSettings()
	lan, _ := drv.GetLanSettings()

	mw.Synchronize(func() {
		if ofd != nil {
			serviceModel.OfdString = joinHostPort(ofd.Addr, ofd.Port)
			serviceModel.OfdClient = ofd.Client // "0" или "1"
			serviceModel.TimerFN = ofd.TimerFN
			serviceModel.TimerOFD = ofd.TimerOFD
		}
		if oism != nil {
			serviceModel.OismString = joinHostPort(oism.Addr, oism.Port)
		}
		if lan != nil {
			serviceModel.LanAddr = lan.Addr
			serviceModel.LanPort = lan.Port
			serviceModel.LanMask = lan.Mask
			serviceModel.LanDns = lan.Dns
			serviceModel.LanGw = lan.Gw
		}
		serviceBinder.Reset()
	})
}

func readAllParameters() {
	drv := driver.Active
	if drv == nil {
		return
	}

	prn, _ := drv.GetPrinterSettings()
	cd, _ := drv.GetMoneyDrawerSettings()
	opts, _ := drv.GetOptions()
	tz, _ := drv.GetTimezone()

	mw.Synchronize(func() {
		// Принтер
		if prn != nil {
			serviceModel.PrintModel = prn.Model
			serviceModel.PrintBaud = strconv.Itoa(prn.BaudRate)
			serviceModel.PrintPaper = prn.Paper
			serviceModel.PrintFont = prn.Font
		}

		// Ящик
		if cd != nil {
			serviceModel.DrawerPin = cd.Pin
			serviceModel.DrawerRise = cd.Rise
			serviceModel.DrawerFall = cd.Fall
		}

		// Часовой пояс
		serviceModel.OptTimezone = strconv.Itoa(tz)

		// Опции
		if opts != nil {
			serviceModel.OptQRPos = fmt.Sprintf("%d", opts.B1)
			serviceModel.OptRounding = fmt.Sprintf("%d", opts.B2)
			serviceModel.OptCut = (opts.B3 == 1)
			serviceModel.OptAutoTest = (opts.B4 == 1)
			serviceModel.OptDrawerTrig = fmt.Sprintf("%d", opts.B5)
			serviceModel.OptNearEnd = (opts.B6 == 1)
			serviceModel.OptTextQR = (opts.B7 == 1)
			serviceModel.OptCountInCheck = (opts.B8 == 1)
		}

		serviceBinder.Reset()
	})
}

func readCliche() {
	drv := driver.Active
	if drv == nil {
		return
	}

	h, _ := drv.GetHeader(1)
	f, _ := drv.GetHeader(3)

	mw.Synchronize(func() {
		for i := range serviceModel.HeaderLines {
			if i < len(h) {
				serviceModel.HeaderLines[i]["Text"] = h[i]
			} else {
				serviceModel.HeaderLines[i]["Text"] = ""
			}
		}
		serviceModel.HeaderModel.PublishRowsReset()

		for i := range serviceModel.FooterLines {
			if i < len(f) {
				serviceModel.FooterLines[i]["Text"] = f[i]
			} else {
				serviceModel.FooterLines[i]["Text"] = ""
			}
		}
		serviceModel.FooterModel.PublishRowsReset()
		serviceBinder.Reset()
	})
}

// -----------------------------
// ЛОГИКА ВРЕМЕНИ
// -----------------------------

func startClocks() {
	if pcTicker == nil {
		pcTicker = time.NewTicker(time.Second)
		go func() {
			for range pcTicker.C {
				if mw == nil || mw.Handle() == 0 {
					continue
				}
				mw.Synchronize(func() {
					now := time.Now()
					serviceModel.PcTimeStr = now.Format("02.01.2006 15:04:05")
					checkTimeDifference(now)
					if pcTimeLabel != nil {
						pcTimeLabel.SetText(serviceModel.PcTimeStr)
					}
				})
			}
		}()
	}
	if kktTicker == nil {
		kktTicker = time.NewTicker(time.Second)
		go func() {
			for range kktTicker.C {
				if mw == nil || mw.Handle() == 0 {
					continue
				}
				mw.Synchronize(func() {
					if len(serviceModel.KktTimeStr) > 10 && serviceModel.KktTimeStr != "Нет подключения" && serviceModel.KktTimeStr != "Ошибка" {
						t, err := time.Parse("02.01.2006 15:04:05", serviceModel.KktTimeStr)
						if err == nil {
							t = t.Add(time.Second)
							serviceModel.KktTimeStr = t.Format("02.01.2006 15:04:05")
							if kktTimeLabel != nil {
								kktTimeLabel.SetText(serviceModel.KktTimeStr)
							}
						}
					}
				})
			}
		}()
	}
}

func checkTimeDifference(pcTime time.Time) {
	if kktTimeLabel == nil {
		return
	}
	kktTime, err := time.Parse("02.01.2006 15:04:05", serviceModel.KktTimeStr)
	if err != nil {
		kktTimeLabel.SetTextColor(walk.RGB(0, 0, 0))
		return
	}
	diff := pcTime.Sub(kktTime)
	if diff < 0 {
		diff = -diff
	}
	if diff > 5*time.Minute {
		kktTimeLabel.SetTextColor(walk.RGB(255, 0, 0))
	} else {
		kktTimeLabel.SetTextColor(walk.RGB(0, 0, 0))
	}
}

// -----------------------------
// UI: GET SERVICE TAB
// -----------------------------

func GetServiceTab() d.TabPage {
	serviceModel = &ServiceViewModel{
		HeaderLines:   make([]map[string]interface{}, 5),
		FooterLines:   make([]map[string]interface{}, 5),
		PrintModel:    "1",
		PrintBaud:     "115200",
		PrintPaper:    80,
		PrintFont:     0,
		DrawerPin:     5,
		DrawerRise:    100,
		DrawerFall:    100,
		OptTimezone:   "3",
		OptQRPos:      "1",
		OptRounding:   "0",
		OptDrawerTrig: "1",
		OptCut:        true,
		OfdClient:     "1", // По умолчанию внешний
		ImgIndex:      0,   // 0 - Логотип
	}

	for i := range serviceModel.HeaderLines {
		serviceModel.HeaderLines[i] = map[string]interface{}{"Text": ""}
	}
	for i := range serviceModel.FooterLines {
		serviceModel.FooterLines[i] = map[string]interface{}{"Text": ""}
	}

	serviceModel.HeaderModel = &HeaderModel{items: serviceModel.HeaderLines}
	serviceModel.FooterModel = &FooterModel{items: serviceModel.FooterLines}

	loadServiceInitial()
	startClocks()

	return d.TabPage{
		Title:  "Сервис",
		Layout: d.VBox{Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}, Spacing: 5},
		Children: []d.Widget{

			// ВЕРХ: Время и Операции
			d.Composite{
				Layout: d.HBox{MarginsZero: true, Spacing: 6},
				Children: []d.Widget{
					d.GroupBox{
						Title: "Синхронизация", StretchFactor: 1,
						Layout: d.VBox{Margins: d.Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 4},
						Children: []d.Widget{
							d.Composite{
								Layout: d.Grid{Columns: 2, Spacing: 4},
								Children: []d.Widget{
									d.Label{Text: "ККТ:", Font: d.Font{PointSize: 8}},
									d.Label{AssignTo: &kktTimeLabel, Text: d.Bind("KktTimeStr"), Font: d.Font{PointSize: 9, Bold: true}},
									d.Label{Text: "ПК:", Font: d.Font{PointSize: 8}},
									d.Label{AssignTo: &pcTimeLabel, Text: d.Bind("PcTimeStr"), Font: d.Font{PointSize: 9, Bold: true}},
								},
							},
							d.VSpacer{Size: 2},
							d.PushButton{Text: "Синхронизировать", OnClicked: onSyncTime},
						},
					},
					d.GroupBox{
						Title: "Операции", StretchFactor: 1,
						Layout: d.Grid{Columns: 2, Margins: d.Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 4},
						Children: []d.Widget{
							d.PushButton{Text: "Перезагрузка", OnClicked: onRebootDevice, MinSize: d.Size{Width: 90}},
							d.PushButton{Text: "Тех. сброс", OnClicked: onTechReset, MinSize: d.Size{Width: 90}},
							d.PushButton{Text: "Ден. ящик", OnClicked: onOpenDrawer, MinSize: d.Size{Width: 90}},
							d.PushButton{Text: "X-отчёт", OnClicked: onPrintXReport, MinSize: d.Size{Width: 90}},
						},
					},
				},
			},

			// ТАБЫ
			d.TabWidget{
				MinSize: d.Size{Height: 320},
				Pages: []d.TabPage{

					// 1. СВЯЗЬ И ОФД
					{
						Title:  "Связь и ОФД",
						Layout: d.VBox{Margins: d.Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 8},
						Children: []d.Widget{
							d.GroupBox{
								Title:  "ОФД и ОИСМ (Адрес:Порт)",
								Layout: d.Grid{Columns: 2, Spacing: 8},
								Children: []d.Widget{
									d.Label{Text: "Сервер ОФД:"}, d.LineEdit{Text: d.Bind("OfdString")},
									d.Label{Text: "Сервер ОИСМ:"}, d.LineEdit{Text: d.Bind("OismString")},
								},
							},
							d.Composite{
								Layout: d.HBox{MarginsZero: true, Spacing: 10, Alignment: d.AlignHNearVNear},
								Children: []d.Widget{
									// ОФД Доп (Слева)
									d.GroupBox{
										Title:         "Настройки клиента",
										StretchFactor: 1,
										Layout:        d.Grid{Columns: 2, Spacing: 6},
										Children: []d.Widget{
											d.Label{Text: "Режим:"},
											d.ComboBox{
												Value:         d.Bind("OfdClient"),
												BindingMember: "Code", DisplayMember: "Name", Model: listClients,
												OnCurrentIndexChanged: checkOfdClientChange,
											},
											d.Label{Text: "Имя клиента:"}, d.LineEdit{Text: d.Bind("OfdClient"), ReadOnly: true, ToolTipText: "Задается автоматически режимом"},
											d.Label{Text: "Таймер ФН:"}, d.NumberEdit{Value: d.Bind("TimerFN"), MaxSize: d.Size{Width: 50}},
											d.Label{Text: "Таймер ОФД:"}, d.NumberEdit{Value: d.Bind("TimerOFD"), MaxSize: d.Size{Width: 50}},
										},
									},
									// LAN (Справа)
									d.GroupBox{
										Title:         "Настройки сети (LAN)",
										StretchFactor: 1,
										Layout:        d.Grid{Columns: 2, Spacing: 6},
										Children: []d.Widget{
											d.Label{Text: "IP:"}, d.LineEdit{Text: d.Bind("LanAddr")},
											d.Label{Text: "Mask:"}, d.LineEdit{Text: d.Bind("LanMask")},
											d.Label{Text: "GW:"}, d.LineEdit{Text: d.Bind("LanGw")},
											d.Label{Text: "Port:"}, d.NumberEdit{Value: d.Bind("LanPort")},
										},
									},
								},
							},
							d.Composite{
								Layout: d.HBox{Alignment: d.AlignHCenterVCenter},
								Children: []d.Widget{
									d.PushButton{Text: "Прочитать настройки", OnClicked: func() { onReadOfdSettings(); onReadOismSettings(); onReadLanSettings() }},
									d.PushButton{Text: "Записать настройки", OnClicked: func() { onWriteOfdSettings(); onWriteOismSettings(); onWriteLanSettings() }},
								},
							},
						},
					},

					// 2. ПАРАМЕТРЫ
					{
						Title:  "Параметры",
						Layout: d.VBox{Margins: d.Margins{Left: 6, Top: 6, Right: 6, Bottom: 6}, Spacing: 6},
						Children: []d.Widget{
							d.Composite{
								Layout: d.HBox{MarginsZero: true, Spacing: 6, Alignment: d.AlignHNearVNear},
								Children: []d.Widget{

									// Группа 1: Железо
									d.GroupBox{
										Title:         "Принтер и Бумага",
										StretchFactor: 1,
										Layout:        d.Grid{Columns: 2, Spacing: 4},
										Children: []d.Widget{
											d.Label{Text: "Скорость:"},
											d.ComboBox{Value: d.Bind("PrintBaud"), BindingMember: "Code", DisplayMember: "Name", Model: listBaud},
											d.Label{Text: "Ширина:"}, d.NumberEdit{Value: d.Bind("PrintPaper")},
											d.Label{Text: "Шрифт:"}, d.NumberEdit{Value: d.Bind("PrintFont")},
											d.Label{Text: "Час. пояс:"},
											d.ComboBox{Value: d.Bind("OptTimezone"), BindingMember: "Code", DisplayMember: "Name", Model: listTimezones},
											d.Label{Text: "Авто-отрез:", MinSize: d.Size{Width: 60}}, d.CheckBox{Checked: d.Bind("OptCut")},
											d.Label{Text: "Звук бумаги:"}, d.CheckBox{Checked: d.Bind("OptNearEnd")},
											d.Label{Text: "Авто-тест:"}, d.CheckBox{Checked: d.Bind("OptAutoTest")},
										},
									},

									// Группа 2: Чеки и Опции
									d.GroupBox{
										Title:         "Вид чека и Опции",
										StretchFactor: 1,
										Layout:        d.Grid{Columns: 2, Spacing: 4},
										Children: []d.Widget{
											d.Label{Text: "QR Позиция:"},
											d.ComboBox{Value: d.Bind("OptQRPos"), BindingMember: "Code", DisplayMember: "Name", Model: listQRPos},

											d.Label{Text: "Округление:"},
											d.ComboBox{Value: d.Bind("OptRounding"), BindingMember: "Code", DisplayMember: "Name", Model: listRounding},

											d.Label{Text: "Ящик при:"},
											d.ComboBox{Value: d.Bind("OptDrawerTrig"), BindingMember: "Code", DisplayMember: "Name", Model: listDrawerTrig},

											d.Label{Text: "Текст у QR:"}, d.CheckBox{Checked: d.Bind("OptTextQR")},
											d.Label{Text: "Кол. покупок:"}, d.CheckBox{Checked: d.Bind("OptCountInCheck")},
										},
									},
								},
							},

							// Группа 3: Денежный ящик (Импульсы)
							d.GroupBox{
								Title:  "Настройки Денежного Ящика (Импульс)",
								Layout: d.HBox{Spacing: 10},
								Children: []d.Widget{
									d.Label{Text: "PIN:"}, d.NumberEdit{Value: d.Bind("DrawerPin"), MaxSize: d.Size{Width: 40}},
									d.Label{Text: "Rise (ms):"}, d.NumberEdit{Value: d.Bind("DrawerRise"), MaxSize: d.Size{Width: 50}},
									d.Label{Text: "Fall (ms):"}, d.NumberEdit{Value: d.Bind("DrawerFall"), MaxSize: d.Size{Width: 50}},
								},
							},

							d.Composite{
								Layout: d.HBox{Alignment: d.AlignHCenterVCenter},
								Children: []d.Widget{
									d.PushButton{Text: "Прочитать всё", OnClicked: readAllParameters},
									d.PushButton{Text: "Записать всё", OnClicked: onWriteAllParameters},
								},
							},
						},
					},

					// 3. КЛИШЕ
					{
						Title:  "Клише",
						Layout: d.HBox{Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}, Spacing: 5},
						Children: []d.Widget{
							d.Composite{
								Layout:        d.VBox{Spacing: 2},
								StretchFactor: 1,
								Children: []d.Widget{
									d.Label{Text: "Заголовок:", Font: d.Font{Bold: true}},
									d.TableView{
										MinSize: d.Size{Width: 100, Height: 100},
										Columns: []d.TableViewColumn{{Title: "Текст"}},
										Model:   serviceModel.HeaderModel,
									},
									d.Composite{Layout: d.HBox{}, Children: []d.Widget{
										d.PushButton{Text: "Read", OnClicked: func() { onReadHeaderSingle(1) }},
										d.PushButton{Text: "Write", OnClicked: func() { onWriteHeaderSingle(1) }},
									}},
								},
							},
							d.Composite{
								Layout:        d.VBox{Spacing: 2},
								StretchFactor: 1,
								Children: []d.Widget{
									d.Label{Text: "Подвал:", Font: d.Font{Bold: true}},
									d.TableView{
										MinSize: d.Size{Width: 100, Height: 100},
										Columns: []d.TableViewColumn{{Title: "Текст"}},
										Model:   serviceModel.FooterModel,
									},
									d.Composite{Layout: d.HBox{}, Children: []d.Widget{
										d.PushButton{Text: "Read", OnClicked: func() { onReadHeaderSingle(3) }},
										d.PushButton{Text: "Write", OnClicked: func() { onWriteHeaderSingle(3) }},
									}},
								},
							},
						},
					},
					// 4. ГРАФИКА
					// {
					// 	Title:  "Графика",
					// 	Layout: d.VBox{Margins: d.Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 10},
					// 	Children: []d.Widget{
					// 		d.GroupBox{
					// 			Title:  "Загрузка изображений (Логотип)",
					// 			Layout: d.Grid{Columns: 3, Spacing: 10},
					// 			Children: []d.Widget{
					// 				d.Label{Text: "Номер (0=Лого):"},
					// 				d.NumberEdit{
					// 					Value:    d.Bind("ImgIndex"),
					// 					MinValue: 0,
					// 					MaxValue: 20,
					// 					MaxSize:  d.Size{Width: 50},
					// 				},
					// 				d.Label{Text: "Индекс 0 - печатается в заголовке чека."},
					// 			},
					// 		},
					// 		d.GroupBox{
					// 			Title:  "Файл",
					// 			Layout: d.VBox{Spacing: 5},
					// 			Children: []d.Widget{
					// 				d.Composite{
					// 					Layout: d.HBox{MarginsZero: true},
					// 					Children: []d.Widget{
					// 						d.LineEdit{Text: d.Bind("ImgPath"), ReadOnly: true},
					// 						d.PushButton{Text: "Выбрать...", OnClicked: onSelectImageFile},
					// 					},
					// 				},
					// 				d.VSpacer{Size: 5},
					// 				d.PushButton{
					// 					Text:      "Загрузить в ККТ",
					// 					OnClicked: onUploadImage,
					// 					MinSize:   d.Size{Height: 40},
					// 				},
					// 			},
					// 		},
					// 		d.Label{
					// 			Text:          "Поддерживаются: JPG, PNG, BMP.\nКартинка будет автоматически преобразована в ч/б BMP\nи уменьшена до ширины печати.",
					// 			TextAlignment: d.AlignCenter,
					// 			MinSize:       d.Size{Height: 60},
					// 		},
					// 	},
					// },
				},
			},
		},
		DataBinder: d.DataBinder{
			AssignTo:       &serviceBinder,
			DataSource:     serviceModel,
			ErrorPresenter: d.ToolTipErrorPresenter{},
		},
	}
}

// -----------------------------
// ЛОГИКА ИНТЕРФЕЙСА (Event Handlers)
// -----------------------------

func checkOfdClientChange() {
	if isLoadingData {
		return
	}
	if serviceModel.OfdClient == "0" {
		res := walk.MsgBox(mw, "Подтверждение",
			"Для использования встроенного клиента ОФД требуется подключение ФР к локальной сети (LAN).\n\nПодтверждаете переключение?",
			walk.MsgBoxYesNo|walk.MsgBoxIconQuestion)

		if res != walk.DlgCmdYes {
			// Откат значения
			serviceModel.OfdClient = "1"
			// Принудительно обновляем UI
			if serviceBinder != nil {
				serviceBinder.Reset()
			}
		}
	}
}

// -----------------------------
// ЛОГИКА ЗАПИСИ
// -----------------------------

func onWriteAllParameters() {
	drv := driver.Active
	if drv == nil {
		return
	}
	if err := serviceBinder.Submit(); err != nil {
		return
	}

	baudRate, _ := strconv.Atoi(serviceModel.PrintBaud)
	tz, _ := strconv.Atoi(serviceModel.OptTimezone)

	go func() {
		// 1. Принтер
		drv.SetPrinterSettings(driver.PrinterSettings{
			Model:    serviceModel.PrintModel,
			BaudRate: baudRate,
			Paper:    serviceModel.PrintPaper,
			Font:     serviceModel.PrintFont,
		})

		// 2. Ящик
		drv.SetMoneyDrawerSettings(driver.DrawerSettings{
			Pin:  serviceModel.DrawerPin,
			Rise: serviceModel.DrawerRise,
			Fall: serviceModel.DrawerFall,
		})

		// 3. Timezone
		drv.SetTimezone(tz)

		// 4. Опции
		if v, err := strconv.Atoi(serviceModel.OptQRPos); err == nil {
			drv.SetOption(1, v)
		}
		if v, err := strconv.Atoi(serviceModel.OptRounding); err == nil {
			drv.SetOption(2, v)
		}

		if serviceModel.OptCut {
			drv.SetOption(3, 1)
		} else {
			drv.SetOption(3, 0)
		}
		if serviceModel.OptAutoTest {
			drv.SetOption(4, 1)
		} else {
			drv.SetOption(4, 0)
		}

		if v, err := strconv.Atoi(serviceModel.OptDrawerTrig); err == nil {
			drv.SetOption(5, v)
		}

		if serviceModel.OptNearEnd {
			drv.SetOption(6, 1)
		} else {
			drv.SetOption(6, 0)
		}
		if serviceModel.OptTextQR {
			drv.SetOption(7, 1)
		} else {
			drv.SetOption(7, 0)
		}
		if serviceModel.OptCountInCheck {
			drv.SetOption(8, 1)
		} else {
			drv.SetOption(8, 0)
		}

		mw.Synchronize(func() {
			walk.MsgBox(mw, "Успех", "Параметры отправлены", walk.MsgBoxIconInformation)
		})
	}()
}

func onWriteOfdSettings() {
	drv := driver.Active
	if drv == nil {
		return
	}
	serviceBinder.Submit()

	h, p := splitHostPort(serviceModel.OfdString)

	go func() {
		err := drv.SetOfdSettings(driver.OfdSettings{
			Addr:     h,
			Port:     p,
			Client:   serviceModel.OfdClient,
			TimerFN:  serviceModel.TimerFN,
			TimerOFD: serviceModel.TimerOFD,
		})
		if err == nil {
			mw.Synchronize(func() { walk.MsgBox(mw, "OK", "ОФД записан", walk.MsgBoxIconInformation) })
		}
	}()
}

func onWriteOismSettings() {
	drv := driver.Active
	if drv == nil {
		return
	}
	serviceBinder.Submit()

	h, p := splitHostPort(serviceModel.OismString)

	go func() {
		err := drv.SetOismSettings(driver.ServerSettings{Addr: h, Port: p})
		if err == nil {
			mw.Synchronize(func() { walk.MsgBox(mw, "OK", "ОИСМ записан", walk.MsgBoxIconInformation) })
		}
	}()
}

func onWriteLanSettings() {
	drv := driver.Active
	if drv == nil {
		return
	}
	serviceBinder.Submit()
	go func() {
		err := drv.SetLanSettings(driver.LanSettings{
			Addr: serviceModel.LanAddr, Mask: serviceModel.LanMask, Port: serviceModel.LanPort,
			Dns: serviceModel.LanDns, Gw: serviceModel.LanGw,
		})
		if err == nil {
			mw.Synchronize(func() { walk.MsgBox(mw, "OK", "LAN записан", walk.MsgBoxIconInformation) })
		}
	}()
}

func onTechReset() {
	drv := driver.Active
	if drv == nil {
		return
	}

	if walk.MsgBox(mw, "ВНИМАНИЕ",
		"Выполнить ТЕХНОЛОГИЧЕСКОЕ ОБНУЛЕНИЕ?\nЭто может привести к сбросу настроек и потере данных в ОЗУ.",
		walk.MsgBoxYesNo|walk.MsgBoxIconWarning) != walk.DlgCmdYes {
		return
	}

	go func() {
		err := drv.TechReset()
		mw.Synchronize(func() {
			if err != nil {
				walk.MsgBox(mw, "Ошибка", "Сбой тех. обнуления: "+err.Error(), walk.MsgBoxIconError)
			} else {
				walk.MsgBox(mw, "Успех", "Технологическое обнуление выполнено.", walk.MsgBoxIconInformation)
				// Перечитываем настройки, так как они сбросились
				isLoadingData = true
				readAllParameters()
				readNetworkSettings()
				isLoadingData = false
			}
		})
	}()
}

// Стандартные функции (не менялись, но нужны для компиляции)
func onQueryTime() {
	drv := driver.Active
	if drv == nil {
		return
	}
	go func() {
		t, err := drv.GetDateTime()
		mw.Synchronize(func() {
			if err != nil {
				walk.MsgBox(mw, "Ошибка", "Ошибка времени: "+err.Error(), walk.MsgBoxIconError)
			} else {
				serviceModel.KktTimeStr = t.Format("02.01.2006 15:04:05")
				serviceBinder.Reset()
			}
		})
	}()
}

func onSyncTime() {
	drv := driver.Active
	if drv == nil {
		return
	}
	now := time.Now()
	go func() {
		err := drv.SetDateTime(now)
		mw.Synchronize(func() {
			if err != nil {
				walk.MsgBox(mw, "Ошибка", "Ошибка синхронизации: "+err.Error(), walk.MsgBoxIconError)
			} else {
				serviceModel.KktTimeStr = now.Format("02.01.2006 15:04:05")
				serviceBinder.Reset()
				walk.MsgBox(mw, "Успех", "Время синхронизировано.", walk.MsgBoxIconInformation)
			}
		})
	}()
}

func onRebootDevice() {
	drv := driver.Active
	if drv == nil {
		return
	}
	go func() {
		drv.RebootDevice()
		mw.Synchronize(func() {
			walk.MsgBox(mw, "Инфо", "Команда перезагрузки отправлена", walk.MsgBoxIconInformation)
		})
	}()
}

func onOpenDrawer() {
	drv := driver.Active
	if drv == nil {
		return
	}
	go func() { drv.DeviceJob(2) }()
}

func onPrintXReport() {
	drv := driver.Active
	if drv == nil {
		return
	}
	go func() { drv.PrintXReport() }()
}

func onReadOfdSettings() {
	mw.Synchronize(func() { isLoadingData = true })
	readNetworkSettings()
	mw.Synchronize(func() { isLoadingData = false })
}
func onReadOismSettings() { onReadOfdSettings() }
func onReadLanSettings()  { onReadOfdSettings() }

func onReadHeaderSingle(mode int) {
	drv := driver.Active
	if drv == nil {
		return
	}
	go func() {
		rows, err := drv.GetHeader(mode)
		mw.Synchronize(func() {
			if err != nil {
				return
			}
			target := serviceModel.HeaderLines
			if mode == 3 {
				target = serviceModel.FooterLines
			}
			for i := range target {
				if i < len(rows) {
					target[i]["Text"] = rows[i]
				} else {
					target[i]["Text"] = ""
				}
			}
			if mode == 1 {
				serviceModel.HeaderModel.PublishRowsReset()
			} else {
				serviceModel.FooterModel.PublishRowsReset()
			}
		})
	}()
}

func onWriteHeaderSingle(mode int) {
	drv := driver.Active
	if drv == nil {
		return
	}
	var lines []string
	src := serviceModel.HeaderLines
	if mode == 3 {
		src = serviceModel.FooterLines
	}
	for _, m := range src {
		lines = append(lines, fmt.Sprintf("%v", m["Text"]))
	}

	go func() {
		for i, txt := range lines {
			drv.SetHeaderLine(mode, i, txt, "")
		}
		mw.Synchronize(func() { walk.MsgBox(mw, "OK", "Записано", walk.MsgBoxIconInformation) })
	}()
}

// Оставлено для попыток реализовать загрузку картинок в логотип
func onSelectImageFile() {
	dlg := new(walk.FileDialog)
	dlg.Title = "Выбор изображения"
	dlg.Filter = "Изображения (*.png;*.jpg;*.bmp)|*.png;*.jpg;*.bmp|Все файлы (*.*)|*.*"

	if ok, _ := dlg.ShowOpen(mw); ok {
		serviceModel.ImgPath = dlg.FilePath
		serviceBinder.Reset()
	}
}

func onUploadImage() {
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения", walk.MsgBoxIconError)
		return
	}
	if serviceModel.ImgPath == "" {
		walk.MsgBox(mw, "Внимание", "Выберите файл", walk.MsgBoxIconWarning)
		return
	}

	// Блокируем интерфейс (опционально можно добавить блокировку)

	go func() {
		// 1. Подготовка картинки (макс ширина 384 точки, безопасное значение)
		// Для F80 может быть 576, но 384 работает везде.
		bmpData, err := PrepareImageForKKT(serviceModel.ImgPath, 384)
		if err != nil {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Ошибка конвертации", err.Error(), walk.MsgBoxIconError)
			})
			return
		}

		logMsg("Размер BMP для загрузки: %d байт", len(bmpData))

		// 2. Загрузка
		if err := drv.UploadImage(serviceModel.ImgIndex, bmpData); err != nil {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Ошибка загрузки", err.Error(), walk.MsgBoxIconError)
			})
			return
		}

		mw.Synchronize(func() {
			walk.MsgBox(mw, "Успех", fmt.Sprintf("Картинка #%d успешно загружена!", serviceModel.ImgIndex), walk.MsgBoxIconInformation)
		})
		logMsg("Загрузка картинки #%d завершена.", serviceModel.ImgIndex)
	}()
}
