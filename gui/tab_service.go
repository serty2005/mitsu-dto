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
		{"Внешний", "1"},
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
	listPapers = []*NV{
		{"80мм", "0"}, {"57мм", "1"},
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
		{"UTC+2 (Клд)", "2"}, {"UTC+3 (Мск)", "3"}, {"UTC+4 (Смр)", "4"},
		{"UTC+5 (Екб)", "5"}, {"UTC+6 (Омск)", "6"}, {"UTC+7 (Крсн)", "7"},
		{"UTC+8 (Ирк)", "8"}, {"UTC+9 (Якт)", "9"}, {"UTC+10 (Влд)", "10"},
		{"UTC+11 (Маг)", "11"}, {"UTC+12 (Кам)", "12"},
	}

	// --- Для КЛИШЕ ---
	listClicheTypes = []*NV{
		{"1 - Заголовок (Клише)", "1"},
		{"2 - После пользователя", "2"},
		{"3 - Подвал (Реклама)", "3"},
		{"4 - Конец чека", "4"},
	}
	listAlign = []*NV{
		{"Слева", "0"}, {"Центр", "1"}, {"Справа", "2"},
	}
	listFonts = []*NV{
		{"A", "0"}, {"B", "1"},
	}
	listUnderline = []*NV{
		{"Нет", "0"}, {"Текст", "1"}, {"Вся строка", "2"},
	}
)

// -----------------------------
// МОДЕЛИ ДАННЫХ КЛИШЕ
// -----------------------------

// ClicheItem представляет одну строку клише для GUI.
type ClicheItem struct {
	Index  int
	Text   string
	Format string // Сырой формат "xxxxxx"

	// Поля для редактирования (разбор Format)
	Invert    bool
	Width     int
	Height    int
	Font      string // "0", "1", "2"
	Underline string // "0", "1", "2"
	Align     string // "0", "1", "2"
}

// ParseFormatString разбирает строку формата "xxxxxx" в поля структуры.
func (c *ClicheItem) ParseFormatString() {
	runes := []rune(c.Format)
	// Добиваем нулями если короткая
	for len(runes) < 6 {
		runes = append(runes, '0')
	}

	c.Invert = (runes[0] == '1')
	c.Width, _ = strconv.Atoi(string(runes[1]))
	c.Height, _ = strconv.Atoi(string(runes[2]))
	c.Font = string(runes[3])
	c.Underline = string(runes[4])
	c.Align = string(runes[5])
}

// UpdateFormatString собирает строку формата "xxxxxx" из полей структуры.
func (c *ClicheItem) UpdateFormatString() {
	inv := "0"
	if c.Invert {
		inv = "1"
	}

	// Лимиты размеров 0..8 (хотя протокол позволяет 1-8, 0-дефолт)
	w := c.Width
	if w < 0 {
		w = 0
	}
	if w > 8 {
		w = 8
	}

	h := c.Height
	if h < 0 {
		h = 0
	}
	if h > 8 {
		h = 8
	}

	c.Format = fmt.Sprintf("%s%d%d%s%s%s",
		inv, w, h,
		ensureChar(c.Font),
		ensureChar(c.Underline),
		ensureChar(c.Align))
}

func ensureChar(s string) string {
	if len(s) == 0 {
		return "0"
	}
	return string(s[0])
}

// ClicheModel - модель для TableView.
type ClicheModel struct {
	walk.TableModelBase
	Items []*ClicheItem
}

func (m *ClicheModel) RowCount() int {
	return len(m.Items)
}

func (m *ClicheModel) Value(row, col int) interface{} {
	item := m.Items[row]
	switch col {
	case 0:
		return item.Index + 1 // Номер строки 1..10
	case 1:
		return item.Format
	case 2:
		return item.Text
	}
	return ""
}

// -----------------------------
// VIEW MODEL (ГЛАВНАЯ)
// -----------------------------

type ServiceViewModel struct {
	// --- Время ---
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
	PrintModel string // "1", "2"
	PrintBaud  string // "115200"
	PrintPaper int    // 57, 80
	PrintFont  int    // 0, 1

	// Опции (b0..b9)
	OptTimezone     string
	OptCut          bool
	OptAutoTest     bool
	OptNearEnd      bool
	OptTextQR       bool
	OptCountInCheck bool
	OptQRPos        string
	OptRounding     string
	OptDrawerTrig   string
	OptB9           string

	// Денежный ящик
	DrawerPin  int
	DrawerRise int
	DrawerFall int

	// --- Клише (Новая структура) ---
	SelectedClicheType string        // "1".."4"
	ClicheItems        []*ClicheItem // 10 строк
	CurrentClicheLine  *ClicheItem   // Указатель на редактируемую строку
}

var (
	serviceModel  *ServiceViewModel
	serviceBinder *walk.DataBinder // Биндинг основной вкладки

	// Элементы для прямого доступа (Время)
	kktTimeLabel *walk.Label
	pcTimeLabel  *walk.Label

	// Таймеры
	pcTicker  *time.Ticker
	kktTicker *time.Ticker

	// Флаг загрузки (блокировка событий)
	isLoadingData bool

	// --- Элементы редактора Клише ---
	clicheTable        *walk.TableView
	clicheModel        *ClicheModel
	clicheEditorGroup  *walk.GroupBox
	clicheEditorBinder *walk.DataBinder // Биндинг панели редактирования
)

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
	// Ждем инициализации окна
	go func() {
		for i := 0; i < 20; i++ {
			if mw != nil && mw.Handle() != 0 {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if mw == nil {
			return
		}

		drv := driver.Active
		if drv == nil {
			mw.Synchronize(func() {
				serviceModel.KktTimeStr = "Нет подключения"
				serviceModel.PcTimeStr = time.Now().Format("02.01.2006 15:04:05")
				serviceBinder.Reset()
			})
			return
		}

		// 1. Время
		t, err := drv.GetDateTime()
		mw.Synchronize(func() {
			if err == nil {
				serviceModel.KktTimeStr = t.Format("02.01.2006 15:04:05")
			} else {
				serviceModel.KktTimeStr = "Ошибка"
			}
			serviceBinder.Reset()
		})

		// 2. Остальные настройки
		mw.Synchronize(func() { isLoadingData = true })
		readNetworkSettings()
		readAllParameters()
		// Клише загружается отдельно по кнопке, но можно инициализировать модель пустыми
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
			serviceModel.OfdClient = ofd.Client
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
			serviceModel.OptB9 = fmt.Sprintf("%d", opts.B9)
		}

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
	// Инициализация модели
	serviceModel = &ServiceViewModel{
		PrintModel:         "1",
		PrintBaud:          "115200",
		PrintPaper:         80,
		PrintFont:          0,
		DrawerPin:          5,
		DrawerRise:         100,
		DrawerFall:         100,
		OptTimezone:        "3",
		OptQRPos:           "1",
		OptRounding:        "0",
		OptDrawerTrig:      "1",
		OptCut:             true,
		OptB9:              "0",
		OfdClient:          "1", // По умолчанию внешний
		SelectedClicheType: "1",
		CurrentClicheLine:  &ClicheItem{}, // Заглушка, чтобы binder не падал
	}

	// Инициализация строк клише (10 пустых)
	serviceModel.ClicheItems = make([]*ClicheItem, 10)
	for i := 0; i < 10; i++ {
		item := &ClicheItem{Index: i, Format: "000000"}
		item.ParseFormatString()
		serviceModel.ClicheItems[i] = item
	}
	clicheModel = &ClicheModel{Items: serviceModel.ClicheItems}

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
							d.PushButton{Text: "Прогон/Отрезка", OnClicked: onFeedAndCutService, MinSize: d.Size{Width: 90}},
							d.PushButton{Text: "Тех. сброс", OnClicked: onTechReset, MinSize: d.Size{Width: 90}},
							d.PushButton{Text: "Ден. ящик", OnClicked: onOpenDrawer, MinSize: d.Size{Width: 90}},
							d.PushButton{Text: "X-отчёт", OnClicked: onPrintXReport, MinSize: d.Size{Width: 90}},
							d.PushButton{Text: "Сброс МГМ", OnClicked: onMGMReset, MinSize: d.Size{Width: 90}},
							d.PushButton{Text: "Прочитать всё", OnClicked: func() { readNetworkSettings(); readAllParameters() }, MinSize: d.Size{Width: 90}},
							d.PushButton{Text: "Записать всё", OnClicked: onWriteAllParameters, MinSize: d.Size{Width: 90}},
						},
					},
				},
			},

			// ТАБЫ ПОДКАТЕГОРИЙ
			d.TabWidget{
				MinSize: d.Size{Height: 350},
				Pages: []d.TabPage{

					// 1. ПАРАМЕТРЫ (Объединенная вкладка)
					{
						Title: "Параметры",
						// Нулевые отступы для вкладки
						Layout: d.VBox{MarginsZero: true, Spacing: 0, Alignment: d.AlignHNearVNear},
						Children: []d.Widget{
							d.Composite{
								Layout: d.HBox{Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}, Spacing: 4, Alignment: d.AlignHNearVNear},
								Children: []d.Widget{

									// КОЛОНКА 1
									d.Composite{
										Layout: d.VBox{MarginsZero: true, Spacing: 4},
										Children: []d.Widget{
											// Group 1: ОФД и ОИСМ
											d.GroupBox{
												Title:  "ОФД и ОИСМ",
												Layout: d.Grid{Columns: 2, Spacing: 4, Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
												Children: []d.Widget{
													d.Label{Text: "ОФД:"}, d.LineEdit{Text: d.Bind("OfdString"), MinSize: d.Size{Width: 110}, MaxSize: d.Size{Width: 120}},
													d.Label{Text: "ОИСМ:"}, d.LineEdit{Text: d.Bind("OismString"), MinSize: d.Size{Width: 110}, MaxSize: d.Size{Width: 120}},
													d.Label{Text: "Пояс:"}, d.ComboBox{Value: d.Bind("OptTimezone"), BindingMember: "Code", DisplayMember: "Name", Model: listTimezones, MinSize: d.Size{Width: 110}, MaxSize: d.Size{Width: 120}},
												},
											},
											// Group 2: Принтер
											d.GroupBox{
												Title:  "Принтер и Бумага",
												Layout: d.Grid{Columns: 4, Spacing: 4, Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
												Children: []d.Widget{
													d.Label{Text: "Модель:"}, d.ComboBox{Value: d.Bind("PrintModel"), BindingMember: "Code", DisplayMember: "Name", Model: listModels, MaxSize: d.Size{Width: 70}},
													d.Label{Text: "Отрез:"}, d.CheckBox{Checked: d.Bind("OptCut")},
													d.Label{Text: "Бод:"}, d.ComboBox{Value: d.Bind("PrintBaud"), BindingMember: "Code", DisplayMember: "Name", Model: listBaud, MaxSize: d.Size{Width: 70}},
													d.Label{Text: "Звук:"}, d.CheckBox{Checked: d.Bind("OptNearEnd")},
													d.Label{Text: "Ширина:"}, d.ComboBox{Value: d.Bind("PrintPaper"), BindingMember: "Code", DisplayMember: "Name", Model: listPapers, MaxSize: d.Size{Width: 70}},
													d.Label{Text: "Тест:"}, d.CheckBox{Checked: d.Bind("OptAutoTest")},
													d.Label{Text: "Шрифт:"}, d.ComboBox{Value: d.Bind("PrintFont"), BindingMember: "Code", DisplayMember: "Name", Model: listFonts, MaxSize: d.Size{Width: 70}, ToolTipText: "A-стандратный, B-компактный"},
												},
											},
										},
									},

									// КОЛОНКА 2
									d.Composite{
										Layout: d.VBox{MarginsZero: true, Spacing: 4},
										Children: []d.Widget{
											// Group 3: Клиент
											d.GroupBox{
												Title:  "Настройки клиента",
												Layout: d.Grid{Columns: 2, Spacing: 4, Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
												Children: []d.Widget{
													d.Label{Text: "Режим:"}, d.ComboBox{Value: d.Bind("OfdClient"), BindingMember: "Code", DisplayMember: "Name", Model: listClients, OnCurrentIndexChanged: checkOfdClientChange, MaxSize: d.Size{Width: 100}},
													d.Label{Text: "Т. ФН:"}, d.NumberEdit{Value: d.Bind("TimerFN"), MaxSize: d.Size{Width: 40}},
													d.Label{Text: "Т. ОФД:"}, d.NumberEdit{Value: d.Bind("TimerOFD"), MaxSize: d.Size{Width: 40}},
												},
											},
											// Group 4: Ящик
											d.GroupBox{
												Title:  "Денежный ящик",
												Layout: d.Grid{Columns: 2, Spacing: 4, Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
												Children: []d.Widget{
													d.Label{Text: "Триггер:"}, d.ComboBox{Value: d.Bind("OptDrawerTrig"), BindingMember: "Code", DisplayMember: "Name", Model: listDrawerTrig, MaxSize: d.Size{Width: 80}},
													d.Label{Text: "PIN:"}, d.NumberEdit{Value: d.Bind("DrawerPin"), MaxSize: d.Size{Width: 40}},
													d.Label{Text: "Rise:"}, d.NumberEdit{Value: d.Bind("DrawerRise"), MaxSize: d.Size{Width: 40}},
													d.Label{Text: "Fall:"}, d.NumberEdit{Value: d.Bind("DrawerFall"), MaxSize: d.Size{Width: 40}},
												},
											},
										},
									},

									// КОЛОНКА 3
									d.Composite{
										Layout: d.VBox{MarginsZero: true, Spacing: 4},
										Children: []d.Widget{
											// Group 5: LAN
											d.GroupBox{
												Title:  "Сеть (LAN)",
												Layout: d.Grid{Columns: 2, Spacing: 4, Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
												Children: []d.Widget{
													d.Label{Text: "IP:"}, d.LineEdit{Text: d.Bind("LanAddr"), MinSize: d.Size{Width: 90}, MaxSize: d.Size{Width: 100}},
													d.Label{Text: "Mask:"}, d.LineEdit{Text: d.Bind("LanMask"), MinSize: d.Size{Width: 90}, MaxSize: d.Size{Width: 100}},
													d.Label{Text: "GW:"}, d.LineEdit{Text: d.Bind("LanGw"), MinSize: d.Size{Width: 90}, MaxSize: d.Size{Width: 100}},
													d.Label{Text: "Port:"}, d.NumberEdit{Value: d.Bind("LanPort"), MaxSize: d.Size{Width: 60}},
												},
											},
											// Group 6: Чеки
											d.GroupBox{
												Title:  "Вид чека и Опции",
												Layout: d.Grid{Columns: 2, Spacing: 4, Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
												Children: []d.Widget{
													d.Label{Text: "QR:"}, d.ComboBox{Value: d.Bind("OptQRPos"), BindingMember: "Code", DisplayMember: "Name", Model: listQRPos, MaxSize: d.Size{Width: 80}},
													d.Label{Text: "Текст QR:"}, d.CheckBox{Checked: d.Bind("OptTextQR")},
													d.Label{Text: "Покупок:"}, d.CheckBox{Checked: d.Bind("OptCountInCheck")},
													d.Label{Text: "Округл.:"}, d.ComboBox{Value: d.Bind("OptRounding"), BindingMember: "Code", DisplayMember: "Name", Model: listRounding, MaxSize: d.Size{Width: 60}},
													d.Label{Text: "b9:"}, d.LineEdit{Text: d.Bind("OptB9"), MaxLength: 3, MaxSize: d.Size{Width: 30}, ToolTipText: "Сумма: СНО(1-8) + X-отчет(16)"},
												},
											},
										},
									},
								},
							},
						},
					},

					// 2. КЛИШЕ (Master-Detail Редактор)
					{
						Title:  "Клише",
						Layout: d.VBox{Margins: d.Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 5},
						Children: []d.Widget{
							// Панель выбора
							d.Composite{
								Layout: d.HBox{MarginsZero: true, Alignment: d.AlignHNearVCenter},
								Children: []d.Widget{
									d.Label{Text: "Редактировать:"},
									d.ComboBox{
										Value:         d.Bind("SelectedClicheType"),
										Model:         listClicheTypes,
										BindingMember: "Code", DisplayMember: "Name",
										MinSize: d.Size{Width: 200},
									},
									d.PushButton{Text: "Считать", OnClicked: onReadCliche},
									d.PushButton{Text: "Записать", OnClicked: onWriteCliche},
								},
							},

							// Рабочая область
							d.Composite{
								Layout: d.HBox{MarginsZero: true, Spacing: 10},
								Children: []d.Widget{
									// Таблица (Список строк)
									d.TableView{
										AssignTo:         &clicheTable,
										Model:            clicheModel,
										AlternatingRowBG: true,
										Columns: []d.TableViewColumn{
											{Title: "#", Width: 30},
											{Title: "Fmt", Width: 60},
											{Title: "Текст", Width: 300},
										},
										MinSize:               d.Size{Width: 400, Height: 200},
										OnCurrentIndexChanged: onClicheSelectionChanged,
									},

									// Редактор выбранной строки
									d.GroupBox{
										AssignTo: &clicheEditorGroup,
										Title:    "Настройки строки",
										Layout:   d.VBox{Margins: d.Margins{Left: 10, Top: 10, Right: 10, Bottom: 10}, Spacing: 8},
										Enabled:  false, // По умолчанию выключено
										DataBinder: d.DataBinder{
											AssignTo:   &clicheEditorBinder,
											DataSource: serviceModel.CurrentClicheLine,
											// OnChange удален, так как его не существует в d.DataBinder
											AutoSubmit: true,
										},
										Children: []d.Widget{
											d.Label{Text: "Текст:"},
											d.LineEdit{
												Text:          d.Bind("Text"),
												OnTextChanged: onClicheItemChanged, // Отслеживаем изменение текста
											},

											d.Composite{
												Layout: d.Grid{Columns: 2, Spacing: 10},
												Children: []d.Widget{
													d.Label{Text: "Выравнивание:"},
													d.ComboBox{
														Value:                 d.Bind("Align"),
														Model:                 listAlign,
														BindingMember:         "Code",
														DisplayMember:         "Name",
														OnCurrentIndexChanged: onClicheItemChanged, // Отслеживаем выбор
													},

													d.Label{Text: "Шрифт:"},
													d.ComboBox{
														Value:                 d.Bind("Font"),
														Model:                 listFonts,
														BindingMember:         "Code",
														DisplayMember:         "Name",
														OnCurrentIndexChanged: onClicheItemChanged, // Отслеживаем выбор
													},

													d.Label{Text: "Подчеркивание:"},
													d.ComboBox{
														Value:                 d.Bind("Underline"),
														Model:                 listUnderline,
														BindingMember:         "Code",
														DisplayMember:         "Name",
														OnCurrentIndexChanged: onClicheItemChanged, // Отслеживаем выбор
													},
												},
											},

											d.GroupBox{
												Title:  "Масштабирование",
												Layout: d.Grid{Columns: 4},
												Children: []d.Widget{
													d.Label{Text: "Ширина:"},
													d.NumberEdit{
														Value:          d.Bind("Width"),
														MinValue:       0,
														MaxValue:       8,
														MaxSize:        d.Size{Width: 40},
														OnValueChanged: onClicheItemChanged, // Отслеживаем число
													},
													d.Label{Text: "Высота:"},
													d.NumberEdit{
														Value:          d.Bind("Height"),
														MinValue:       0,
														MaxValue:       8,
														MaxSize:        d.Size{Width: 40},
														OnValueChanged: onClicheItemChanged, // Отслеживаем число
													},
												},
											},
											d.CheckBox{
												Text:                "Инверсия (Белым по черному)",
												Checked:             d.Bind("Invert"),
												OnCheckStateChanged: func() { onClicheItemChanged() }, // Отслеживаем галочку
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

	// Подготовка данных ОФД/LAN
	ofdHost, ofdPort := splitHostPort(serviceModel.OfdString)
	oismHost, oismPort := splitHostPort(serviceModel.OismString)

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
		if v, err := strconv.Atoi(serviceModel.OptB9); err == nil {
			drv.SetOption(9, v)
		}

		// 5. ОФД
		drv.SetOfdSettings(driver.OfdSettings{
			Addr:     ofdHost,
			Port:     ofdPort,
			Client:   serviceModel.OfdClient,
			TimerFN:  serviceModel.TimerFN,
			TimerOFD: serviceModel.TimerOFD,
		})

		// 6. ОИСМ
		drv.SetOismSettings(driver.ServerSettings{Addr: oismHost, Port: oismPort})

		// 7. LAN
		drv.SetLanSettings(driver.LanSettings{
			Addr: serviceModel.LanAddr, Mask: serviceModel.LanMask, Port: serviceModel.LanPort,
			Dns: serviceModel.LanDns, Gw: serviceModel.LanGw,
		})

		mw.Synchronize(func() {
			walk.MsgBox(mw, "Успех", "Все параметры отправлены в ККТ", walk.MsgBoxIconInformation)
		})
	}()
}

// Удалены onWriteOfdSettings, onWriteOismSettings, onWriteLanSettings за ненадобностью

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

func onFeedAndCutService() {
	drv := driver.Active
	if drv == nil {
		return
	}
	go func() {
		drv.Feed(5)
		drv.Cut()
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

func onMGMReset() {
	drv := driver.Active
	if drv == nil {
		return
	}
	go func() { drv.ResetMGM() }()
}

func onReadOfdSettings() {
	mw.Synchronize(func() { isLoadingData = true })
	readNetworkSettings()
	mw.Synchronize(func() { isLoadingData = false })
}
func onReadOismSettings() { onReadOfdSettings() }
func onReadLanSettings()  { onReadOfdSettings() }

// -----------------------------
// ЛОГИКА КЛИШЕ (НОВАЯ)
// -----------------------------

// onReadCliche читает выбранный тип клише из ККТ
func onReadCliche() {
	drv := driver.Active
	if drv == nil {
		walk.MsgBox(mw, "Ошибка", "Нет подключения", walk.MsgBoxIconError)
		return
	}

	typeID, _ := strconv.Atoi(serviceModel.SelectedClicheType)

	go func() {
		lines, err := drv.GetHeader(typeID)
		if err != nil {
			mw.Synchronize(func() {
				walk.MsgBox(mw, "Ошибка", fmt.Sprintf("Ошибка чтения клише: %v", err), walk.MsgBoxIconError)
			})
			return
		}

		mw.Synchronize(func() {
			// Заполняем модель
			for i := 0; i < 10; i++ {
				if i < len(lines) {
					serviceModel.ClicheItems[i].Text = lines[i].Text
					serviceModel.ClicheItems[i].Format = lines[i].Format
				} else {
					serviceModel.ClicheItems[i].Text = ""
					serviceModel.ClicheItems[i].Format = "000000"
				}
				// Парсим формат для редактора
				serviceModel.ClicheItems[i].ParseFormatString()
			}
			// Обновляем таблицу
			clicheModel.PublishRowsReset()
			// Если что-то выбрано, обновляем редактор
			idx := clicheTable.CurrentIndex()
			if idx >= 0 {
				reloadEditor(idx)
			}
		})
	}()
}

// onWriteCliche записывает текущее состояние таблицы в ККТ
func onWriteCliche() {
	drv := driver.Active
	if drv == nil {
		return
	}

	typeID, _ := strconv.Atoi(serviceModel.SelectedClicheType)

	// Подготавливаем данные
	var linesToWrite []struct{ txt, fmt string }
	for _, item := range serviceModel.ClicheItems {
		// Убедимся, что формат актуален
		item.UpdateFormatString()
		linesToWrite = append(linesToWrite, struct{ txt, fmt string }{item.Text, item.Format})
	}

	go func() {
		for i, l := range linesToWrite {
			// i = номер строки (0..9)
			// l.fmt уже "xxxxxx"
			if err := drv.SetHeaderLine(typeID, i, l.txt, l.fmt); err != nil {
				// Логируем или игнорируем
				fmt.Printf("Error writing line %d: %v\n", i, err)
			}
		}
		mw.Synchronize(func() {
			walk.MsgBox(mw, "Успех", "Клише записано", walk.MsgBoxIconInformation)
		})
	}()
}

// onClicheSelectionChanged вызывается при клике на строку таблицы
func onClicheSelectionChanged() {
	idx := clicheTable.CurrentIndex()
	if idx < 0 {
		clicheEditorGroup.SetEnabled(false)
		return
	}
	reloadEditor(idx)
}

// reloadEditor перепривязывает редактор к выбранной строке
func reloadEditor(idx int) {
	// 1. Берем указатель на реальный объект из списка
	item := serviceModel.ClicheItems[idx]

	// 2. Подменяем DataSource у биндера редактора
	// ВАЖНО: Мы меняем источник данных для binder'а на лету
	clicheEditorBinder.SetDataSource(item)
	clicheEditorBinder.Reset()

	clicheEditorGroup.SetEnabled(true)
	clicheEditorGroup.SetTitle(fmt.Sprintf("Настройки строки №%d", idx+1))
}

// onClicheItemChanged вызывается при любом изменении в полях редактора
func onClicheItemChanged() {
	// Принудительно обновляем форматную строку на основе полей
	idx := clicheTable.CurrentIndex()
	if idx >= 0 {
		item := serviceModel.ClicheItems[idx]
		item.UpdateFormatString()
		// Уведомляем таблицу, что данные в этой строке изменились
		clicheModel.PublishRowChanged(idx)
	}
}
