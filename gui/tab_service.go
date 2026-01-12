package gui

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"

	"mitsuscanner/driver"
	"mitsuscanner/internal/app"
	"mitsuscanner/internal/models"
	"mitsuscanner/internal/service/settings"
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
		{"Без принтера", "0"}, {"RP-809", "1"}, {"F80", "2"},
	}
	listFonts = []*NV{
		{"А", "0"}, {"B", "1"},
	}
	listPaper = []*NV{
		{"57 мм", "57"}, {"80 мм", "80"},
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
	listUnderline = []*NV{
		{"Нет", "0"}, {"Текст", "1"}, {"Вся строка", "2"},
	}
)

// -----------------------------
// МОДЕЛИ ДАННЫХ КЛИШЕ
// -----------------------------

// ClicheModel - модель для TableView.
// Использует models.ClicheItem вместо локальной структуры.
type ClicheModel struct {
	walk.TableModelBase
	Items []*models.ClicheItem
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

	// --- Клише ---
	SelectedClicheType string               // "1".."4"
	ClicheItems        []*models.ClicheItem // 10 строк
	CurrentClicheLine  *models.ClicheItem   // Указатель на редактируемую строку

	// --- Снимок для сравнения ---
	OriginalSnapshot *models.Snapshot
}

// CreateSnapshot создает глубокую копию текущих настроек для последующего сравнения
func (s *ServiceViewModel) CreateSnapshot() *models.Snapshot {
	return models.CreateSnapshotFromModel(
		s.OfdString, s.OfdClient,
		s.TimerFN, s.TimerOFD,
		s.OismString,
		s.LanAddr, s.LanPort, s.LanMask, s.LanDns, s.LanGw,
		s.PrintModel, s.PrintBaud, s.PrintPaper, s.PrintFont,
		s.OptTimezone, s.OptCut, s.OptAutoTest, s.OptNearEnd, s.OptTextQR, s.OptCountInCheck,
		s.OptQRPos, s.OptRounding, s.OptDrawerTrig, s.OptB9,
		s.DrawerPin, s.DrawerRise, s.DrawerFall,
		s.ClicheItems,
	)
}

// -----------------------------
// SERVICE TAB (КОНТРОЛЛЕР)
// -----------------------------

type ServiceTab struct {
	app         *app.App
	viewModel   *ServiceViewModel
	binder      *walk.DataBinder
	clicheModel *ClicheModel

	// Элементы для прямого доступа (Время)
	kktTimeLabel *walk.Label
	pcTimeLabel  *walk.Label

	// Таймеры
	pcTicker  *time.Ticker
	kktTicker *time.Ticker

	// Флаг загрузки (блокировка событий)
	isLoadingData bool

	// --- ActionButton ---
	actionButton *walk.PushButton
	isWriteMode  bool
	normalFont   *walk.Font
	boldFont     *walk.Font

	// --- Labels for highlighting ---
	ofdStringLabel       *walk.Label
	oismStringLabel      *walk.Label
	lanAddrLabel         *walk.Label
	lanMaskLabel         *walk.Label
	lanDnsLabel          *walk.Label
	lanGwLabel           *walk.Label
	lanPortLabel         *walk.Label
	ofdClientLabel       *walk.Label
	timerFNLabel         *walk.Label
	timerOFDLabel        *walk.Label
	printModelLabel      *walk.Label
	printBaudLabel       *walk.Label
	printPaperLabel      *walk.Label
	printFontLabel       *walk.Label
	optTimezoneLabel     *walk.Label
	optCutLabel          *walk.Label
	optNearEndLabel      *walk.Label
	optAutoTestLabel     *walk.Label
	optQRPosLabel        *walk.Label
	optRoundingLabel     *walk.Label
	optDrawerTrigLabel   *walk.Label
	optTextQRLabel       *walk.Label
	optCountInCheckLabel *walk.Label
	optB9Label           *walk.Label
	drawerPinLabel       *walk.Label
	drawerRiseLabel      *walk.Label
	drawerFallLabel      *walk.Label

	// --- Элементы редактора Клише ---
	clicheTable        *walk.TableView
	clicheEditorGroup  *walk.GroupBox
	clicheEditorBinder *walk.DataBinder // Биндинг панели редактирования
}

// NewServiceTab создает новый экземпляр контроллера ServiceTab
func NewServiceTab(a *app.App) *ServiceTab {
	st := &ServiceTab{
		app: a,
		viewModel: &ServiceViewModel{
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
			CurrentClicheLine:  &models.ClicheItem{}, // Заглушка
		},
	}

	// Инициализация строк клише (10 пустых)
	st.viewModel.ClicheItems = make([]*models.ClicheItem, 10)
	for i := 0; i < 10; i++ {
		item := &models.ClicheItem{Index: i, Format: "000000"}
		item.ParseFormatString()
		st.viewModel.ClicheItems[i] = item
	}
	st.clicheModel = &ClicheModel{Items: st.viewModel.ClicheItems}

	return st
}

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

func (st *ServiceTab) loadServiceInitial() {
	// Ждем инициализации окна
	go func() {
		for i := 0; i < 20; i++ {
			if st.app.MainWindow != nil && st.app.MainWindow.Handle() != 0 {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if st.app.MainWindow == nil {
			log.Printf("[GUI] loadServiceInitial: MainWindow is nil after wait")
			return
		}
		log.Printf("[GUI] loadServiceInitial: MainWindow ready, Handle: %d", st.app.MainWindow.Handle())
		if st.app.MainWindow.Handle() == 0 {
			log.Printf("[GUI] loadServiceInitial: Handle is 0, window not ready, skipping")
			return
		}

		drv := st.app.GetDriver()
		if drv == nil {
			log.Printf("[GUI] loadServiceInitial: Driver is nil, calling Synchronize")
			st.app.MainWindow.Synchronize(func() {
				log.Printf("[GUI] loadServiceInitial: Inside Synchronize for no driver")
				st.viewModel.KktTimeStr = "Нет подключения"
				st.viewModel.PcTimeStr = time.Now().Format("02.01.2006 15:04:05")
				st.binder.Reset()
			})
			return
		}

		// 1. Время
		t, err := drv.GetDateTime()
		log.Printf("[GUI] loadServiceInitial: Calling Synchronize for time update")
		st.app.MainWindow.Synchronize(func() {
			log.Printf("[GUI] loadServiceInitial: Inside Synchronize for time")
			if err == nil {
				st.viewModel.KktTimeStr = t.Format("02.01.2006 15:04:05")
			} else {
				st.viewModel.KktTimeStr = "Ошибка"
			}
			st.binder.Reset()
		})

		// 2. Остальные настройки
		log.Printf("[GUI] loadServiceInitial: Calling Synchronize for loading start")
		st.app.MainWindow.Synchronize(func() { st.isLoadingData = true })
		st.readNetworkSettings()
		st.readAllParameters()
		// Клише загружается отдельно по кнопке
		log.Printf("[GUI] loadServiceInitial: Calling Synchronize for loading end")
		st.app.MainWindow.Synchronize(func() {
			st.isLoadingData = false
			st.viewModel.OriginalSnapshot = st.viewModel.CreateSnapshot()
		})
	}()
}

func (st *ServiceTab) readNetworkSettings() {
	drv := st.app.GetDriver()
	if drv == nil {
		return
	}

	ofd, _ := drv.GetOfdSettings()
	oism, _ := drv.GetOismSettings()
	lan, _ := drv.GetLanSettings()

	st.app.MainWindow.Synchronize(func() {
		if ofd != nil {
			st.viewModel.OfdString = joinHostPort(ofd.Addr, ofd.Port)
			st.viewModel.OfdClient = ofd.Client
			st.viewModel.TimerFN = ofd.TimerFN
			st.viewModel.TimerOFD = ofd.TimerOFD
		}
		if oism != nil {
			st.viewModel.OismString = joinHostPort(oism.Addr, oism.Port)
		}
		if lan != nil {
			st.viewModel.LanAddr = lan.Addr
			st.viewModel.LanPort = lan.Port
			st.viewModel.LanMask = lan.Mask
			st.viewModel.LanDns = lan.Dns
			st.viewModel.LanGw = lan.Gw
		}
		st.binder.Reset()
	})
}

func (st *ServiceTab) readAllParameters() {
	drv := st.app.GetDriver()
	if drv == nil {
		return
	}

	prn, _ := drv.GetPrinterSettings()
	cd, _ := drv.GetMoneyDrawerSettings()
	opts, _ := drv.GetOptions()
	tz, _ := drv.GetTimezone()

	st.app.MainWindow.Synchronize(func() {
		// Принтер
		if prn != nil {
			st.viewModel.PrintModel = prn.Model
			st.viewModel.PrintBaud = strconv.Itoa(prn.BaudRate)
			st.viewModel.PrintPaper = prn.Paper
			st.viewModel.PrintFont = prn.Font
		}

		// Ящик
		if cd != nil {
			st.viewModel.DrawerPin = cd.Pin
			st.viewModel.DrawerRise = cd.Rise
			st.viewModel.DrawerFall = cd.Fall
		}

		// Часовой пояс
		st.viewModel.OptTimezone = strconv.Itoa(tz)

		// Опции
		if opts != nil {
			st.viewModel.OptQRPos = fmt.Sprintf("%d", opts.B1)
			st.viewModel.OptRounding = fmt.Sprintf("%d", opts.B2)
			st.viewModel.OptCut = (opts.B3 == 1)
			st.viewModel.OptAutoTest = (opts.B4 == 1)
			st.viewModel.OptDrawerTrig = fmt.Sprintf("%d", opts.B5)
			st.viewModel.OptNearEnd = (opts.B6 == 1)
			st.viewModel.OptTextQR = (opts.B7 == 1)
			st.viewModel.OptCountInCheck = (opts.B8 == 1)
			st.viewModel.OptB9 = fmt.Sprintf("%d", opts.B9)
		}

		st.binder.Reset()
	})
}

// -----------------------------
// ЛОГИКА ВРЕМЕНИ
// -----------------------------

func (st *ServiceTab) startClocks() {
	if st.pcTicker == nil {
		st.pcTicker = time.NewTicker(time.Second)
		go func() {
			for range st.pcTicker.C {
				if st.app.MainWindow == nil || st.app.MainWindow.Handle() == 0 {
					continue
				}
				st.app.MainWindow.Synchronize(func() {
					now := time.Now()
					st.viewModel.PcTimeStr = now.Format("02.01.2006 15:04:05")
					st.checkTimeDifference(now)
					if st.pcTimeLabel != nil {
						st.pcTimeLabel.SetText(st.viewModel.PcTimeStr)
					}
				})
			}
		}()
	}
	if st.kktTicker == nil {
		st.kktTicker = time.NewTicker(time.Second)
		go func() {
			for range st.kktTicker.C {
				if st.app.MainWindow == nil || st.app.MainWindow.Handle() == 0 {
					continue
				}
				st.app.MainWindow.Synchronize(func() {
					if len(st.viewModel.KktTimeStr) > 10 && st.viewModel.KktTimeStr != "Нет подключения" && st.viewModel.KktTimeStr != "Ошибка" {
						t, err := time.Parse("02.01.2006 15:04:05", st.viewModel.KktTimeStr)
						if err == nil {
							t = t.Add(time.Second)
							st.viewModel.KktTimeStr = t.Format("02.01.2006 15:04:05")
							if st.kktTimeLabel != nil {
								st.kktTimeLabel.SetText(st.viewModel.KktTimeStr)
							}
						}
					}
				})
			}
		}()
	}
}

func (st *ServiceTab) checkTimeDifference(pcTime time.Time) {
	if st.kktTimeLabel == nil {
		return
	}
	kktTime, err := time.Parse("02.01.2006 15:04:05", st.viewModel.KktTimeStr)
	if err != nil {
		st.kktTimeLabel.SetTextColor(walk.RGB(0, 0, 0))
		return
	}
	diff := pcTime.Sub(kktTime)
	if diff < 0 {
		diff = -diff
	}
	if diff > 5*time.Minute {
		st.kktTimeLabel.SetTextColor(walk.RGB(255, 0, 0))
	} else {
		st.kktTimeLabel.SetTextColor(walk.RGB(0, 0, 0))
	}
}

// -----------------------------
// UI: CREATE
// -----------------------------

func (st *ServiceTab) Create() d.TabPage {
	st.loadServiceInitial()
	st.startClocks()

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
									d.Label{AssignTo: &st.kktTimeLabel, Text: d.Bind("KktTimeStr"), Font: d.Font{PointSize: 9, Bold: true}},
									d.Label{Text: "ПК:", Font: d.Font{PointSize: 8}},
									d.Label{AssignTo: &st.pcTimeLabel, Text: d.Bind("PcTimeStr"), Font: d.Font{PointSize: 9, Bold: true}},
								},
							},
							d.VSpacer{Size: 2},
							d.PushButton{Text: "Синхронизировать", OnClicked: st.onSyncTime},
						},
					},
					d.GroupBox{
						Title: "Операции", StretchFactor: 1,
						Layout: d.Grid{Columns: 2, Margins: d.Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 4},
						Children: []d.Widget{
							d.PushButton{Text: "Прогон/Отрезка", OnClicked: st.onFeedAndCut, MinSize: d.Size{Width: 90}},
							d.PushButton{Text: "Тех. сброс", OnClicked: st.onTechReset, MinSize: d.Size{Width: 90}},
							d.PushButton{Text: "Ден. ящик", OnClicked: st.onOpenDrawer, MinSize: d.Size{Width: 90}},
							d.PushButton{Text: "X-отчёт", OnClicked: st.onPrintXReport, MinSize: d.Size{Width: 90}},
							d.PushButton{Text: "Сброс МГМ", OnClicked: st.onMGMReset, MinSize: d.Size{Width: 90}},
							d.PushButton{AssignTo: &st.actionButton, Text: "Прочитать всё", OnClicked: st.onActionButtonClicked, MinSize: d.Size{Width: 140}},
						},
					},
				},
			},

			// ТАБЫ ПОДКАТЕГОРИЙ
			d.TabWidget{
				MinSize: d.Size{Height: 250, Width: 300},
				Pages: []d.TabPage{

					// 1. ПАРАМЕТРЫ
					{
						Title:  "Параметры",
						Layout: d.VBox{Margins: d.Margins{Left: 6, Top: 6, Right: 6, Bottom: 6}, Spacing: 3},
						Children: []d.Widget{
							// --- Блок 1: ОФД и Сеть ---
							d.GroupBox{
								Title:  "ОФД и Часовой пояс",
								Layout: d.HBox{Alignment: d.AlignHNearVNear, Spacing: 3},
								Children: []d.Widget{
									// Колонка 1: Адреса ОФД и Часовой пояс
									d.Composite{
										Layout: d.Grid{Columns: 2, Spacing: 3},
										Children: []d.Widget{
											d.Label{AssignTo: &st.ofdStringLabel, Text: "Адрес ОФД:", ToolTipText: "Адрес ОФД в формате URL:PORT"},
											d.LineEdit{Text: d.Bind("OfdString"), MaxSize: d.Size{Width: 80}, OnTextChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.oismStringLabel, Text: "Адрес ОИСМ:", ToolTipText: "Адрес ОИСМ в формате URL:PORT"},
											d.LineEdit{Text: d.Bind("OismString"), MaxSize: d.Size{Width: 80}, OnTextChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.optTimezoneLabel, Text: "Часовой пояс:"},
											d.ComboBox{
												Value: d.Bind("OptTimezone"), BindingMember: "Code", DisplayMember: "Name", Model: listTimezones,
												MinSize: d.Size{Width: 80}, OnCurrentIndexChanged: st.updateChangeTracking,
											},
										},
									},
									// Колонка 2: Клиент и Таймеры
									d.Composite{
										Layout: d.Grid{Columns: 2, Spacing: 3},
										Children: []d.Widget{
											d.Label{AssignTo: &st.ofdClientLabel, Text: "Режим клиента ОФД:"},
											d.ComboBox{
												Value: d.Bind("OfdClient"), BindingMember: "Code", DisplayMember: "Name", Model: listClients,
												MinSize: d.Size{Width: 100}, OnCurrentIndexChanged: func() { st.checkOfdClientChange(); st.updateChangeTracking() },
											},

											d.Label{AssignTo: &st.timerFNLabel, Text: "Таймер ФН:"},
											d.NumberEdit{Value: d.Bind("TimerFN"), MaxSize: d.Size{Width: 60}, OnValueChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.timerOFDLabel, Text: "Таймер ОФД:"},
											d.NumberEdit{Value: d.Bind("TimerOFD"), MaxSize: d.Size{Width: 60}, OnValueChanged: st.updateChangeTracking},
										},
									},
									// Колонка 3: LAN (Справа)
									d.Composite{
										Layout: d.Grid{Columns: 2, Spacing: 3},
										Children: []d.Widget{
											d.Label{AssignTo: &st.lanAddrLabel, Text: "IP LAN:"},
											d.LineEdit{Text: d.Bind("LanAddr"), MinSize: d.Size{Width: 100}, OnTextChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.lanMaskLabel, Text: "Mask LAN:"},
											d.LineEdit{Text: d.Bind("LanMask"), MinSize: d.Size{Width: 100}, OnTextChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.lanGwLabel, Text: "GW LAN:"},
											d.LineEdit{Text: d.Bind("LanGw"), MinSize: d.Size{Width: 100}, OnTextChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.lanPortLabel, Text: "Port LAN:"},
											d.NumberEdit{Value: d.Bind("LanPort"), MaxSize: d.Size{Width: 60}, OnValueChanged: st.updateChangeTracking},
										},
									},
								},
							},

							// --- Блок 2: Настройки устройства ---
							d.GroupBox{
								Title:  "Настройки устройства",
								Layout: d.HBox{Alignment: d.AlignHNearVNear, Spacing: 2},
								Children: []d.Widget{
									// Колонка 1: Принтер
									d.GroupBox{
										Title:  "Принтер",
										Layout: d.Grid{Columns: 4, Spacing: 3},
										Children: []d.Widget{
											d.Label{AssignTo: &st.printModelLabel, Text: "Printer:", ToolTipText: "Модель принтера"},
											d.ComboBox{Value: d.Bind("PrintModel"), MaxSize: d.Size{Width: 60}, BindingMember: "Code", DisplayMember: "Name", Model: listModels, OnCurrentIndexChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.optCutLabel, Text: "Отрезчик:", ToolTipText: "Включить/выключить отрезку чеков"},
											d.CheckBox{Checked: d.Bind("OptCut"), OnCheckStateChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.printBaudLabel, Text: "Baud:", ToolTipText: "Скорость принтера"},
											d.ComboBox{Value: d.Bind("PrintBaud"), MaxSize: d.Size{Width: 60}, BindingMember: "Code", DisplayMember: "Name", Model: listBaud, OnCurrentIndexChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.optNearEndLabel, Text: "Near-end:", ToolTipText: "Сигнал скорого окончания бумаги"},
											d.CheckBox{Checked: d.Bind("OptNearEnd"), OnCheckStateChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.printPaperLabel, Text: "Бумага:", ToolTipText: "Размер бумаги"},
											d.ComboBox{Value: d.Bind("PrintPaper"), MaxSize: d.Size{Width: 60}, BindingMember: "Code", DisplayMember: "Name", Model: listPaper, OnCurrentIndexChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.optAutoTestLabel, Text: "Авто-тест:", ToolTipText: "Печать тестирования при запуске"},
											d.CheckBox{Checked: d.Bind("OptAutoTest"), OnCheckStateChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.printFontLabel, Text: "Шрифт:", ToolTipText: "Номер шрифта. A - стандартный, B - компактный"},
											d.ComboBox{Value: d.Bind("PrintFont"), MaxSize: d.Size{Width: 60}, BindingMember: "Code", DisplayMember: "Name", Model: listFonts, OnCurrentIndexChanged: st.updateChangeTracking},
										},
									},

									// Колонка 2: Денежный ящик
									d.GroupBox{
										Title:  "Денежный ящик",
										Layout: d.Grid{Columns: 2, Spacing: 3},
										Children: []d.Widget{

											d.Label{AssignTo: &st.optDrawerTrigLabel, Text: "Ящик:"},
											d.ComboBox{Value: d.Bind("OptDrawerTrig"), MaxSize: d.Size{Width: 70}, BindingMember: "Code", DisplayMember: "Name", Model: listDrawerTrig, OnCurrentIndexChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.drawerPinLabel, Text: "PIN ящика:"},
											d.NumberEdit{Value: d.Bind("DrawerPin"), MaxSize: d.Size{Width: 70}, OnValueChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.drawerRiseLabel, Text: "Rise ящика (ms):"},
											d.NumberEdit{Value: d.Bind("DrawerRise"), MaxSize: d.Size{Width: 70}, OnValueChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.drawerFallLabel, Text: "Fall ящика (ms):"},
											d.NumberEdit{Value: d.Bind("DrawerFall"), MaxSize: d.Size{Width: 70}, OnValueChanged: st.updateChangeTracking},
										},
									},

									// Колонка 3: Форматирование и QR
									d.Composite{
										Layout: d.Grid{Columns: 1, Spacing: 1},
										Children: []d.Widget{
											d.Label{AssignTo: &st.optQRPosLabel, Text: "QR Позиция:"},
											d.ComboBox{Value: d.Bind("OptQRPos"), BindingMember: "Code", DisplayMember: "Name", Model: listQRPos, OnCurrentIndexChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.optTextQRLabel, Text: "Текст у QR:"},
											d.CheckBox{Checked: d.Bind("OptTextQR"), OnCheckStateChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.optCountInCheckLabel, Text: "Количество покупок:"},
											d.CheckBox{Checked: d.Bind("OptCountInCheck"), OnCheckStateChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.optRoundingLabel, Text: "Округление:"},
											d.ComboBox{Value: d.Bind("OptRounding"), BindingMember: "Code", DisplayMember: "Name", Model: listRounding, OnCurrentIndexChanged: st.updateChangeTracking},

											d.Label{AssignTo: &st.optB9Label, Text: "Опция b9 (СНО):"},
											d.LineEdit{Text: d.Bind("OptB9"), MaxLength: 3, MaxSize: d.Size{Width: 60}, ToolTipText: "Сумма: СНО(1-8) + X-отчет(16)", OnTextChanged: st.updateChangeTracking},
										},
									},
								},
							},
						},
					},

					// 2. КЛИШЕ
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
									d.PushButton{Text: "Считать", OnClicked: st.onReadCliche},
									d.PushButton{Text: "Записать", OnClicked: st.onWriteCliche},
								},
							},

							// Рабочая область
							d.Composite{
								Layout: d.HBox{MarginsZero: true, Spacing: 10},
								Children: []d.Widget{
									// Таблица (Список строк)
									d.TableView{
										AssignTo:         &st.clicheTable,
										Model:            st.clicheModel,
										AlternatingRowBG: true,
										Columns: []d.TableViewColumn{
											{Title: "#", Width: 30},
											{Title: "Fmt", Width: 60},
											{Title: "Текст", Width: 300},
										},
										MinSize:               d.Size{Width: 400, Height: 200},
										OnCurrentIndexChanged: st.onClicheSelectionChanged,
									},

									// Редактор выбранной строки
									d.GroupBox{
										AssignTo: &st.clicheEditorGroup,
										Title:    "Настройки строки",
										Layout:   d.VBox{Margins: d.Margins{Left: 10, Top: 10, Right: 10, Bottom: 10}, Spacing: 8},
										Enabled:  false,
										DataBinder: d.DataBinder{
											AssignTo:   &st.clicheEditorBinder,
											DataSource: st.viewModel.CurrentClicheLine,
											AutoSubmit: true,
										},
										Children: []d.Widget{
											d.Label{Text: "Текст:"},
											d.LineEdit{
												Text:          d.Bind("Text"),
												OnTextChanged: func() { st.onClicheItemChanged(); st.updateChangeTracking() },
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
														OnCurrentIndexChanged: func() { st.onClicheItemChanged(); st.updateChangeTracking() },
													},

													d.Label{Text: "Шрифт:"},
													d.ComboBox{
														Value:                 d.Bind("Font"),
														Model:                 listFonts,
														BindingMember:         "Code",
														DisplayMember:         "Name",
														OnCurrentIndexChanged: func() { st.onClicheItemChanged(); st.updateChangeTracking() },
													},

													d.Label{Text: "Подчеркивание:"},
													d.ComboBox{
														Value:                 d.Bind("Underline"),
														Model:                 listUnderline,
														BindingMember:         "Code",
														DisplayMember:         "Name",
														OnCurrentIndexChanged: func() { st.onClicheItemChanged(); st.updateChangeTracking() },
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
														OnValueChanged: func() { st.onClicheItemChanged(); st.updateChangeTracking() },
													},
													d.Label{Text: "Высота:"},
													d.NumberEdit{
														Value:          d.Bind("Height"),
														MinValue:       0,
														MaxValue:       8,
														MaxSize:        d.Size{Width: 40},
														OnValueChanged: func() { st.onClicheItemChanged(); st.updateChangeTracking() },
													},
												},
											},
											d.CheckBox{
												Text:                "Инверсия (Белым по черному)",
												Checked:             d.Bind("Invert"),
												OnCheckStateChanged: func() { st.onClicheItemChanged(); st.updateChangeTracking() },
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
			AssignTo:       &st.binder,
			DataSource:     st.viewModel,
			ErrorPresenter: d.ToolTipErrorPresenter{},
		},
	}
}

// -----------------------------
// ЛОГИКА ИНТЕРФЕЙСА (Event Handlers)
// -----------------------------

func (st *ServiceTab) checkOfdClientChange() {
	if st.isLoadingData {
		return
	}
	if st.viewModel.OfdClient == "0" {
		res := walk.MsgBox(st.app.MainWindow, "Подтверждение",
			"Для использования встроенного клиента ОФД требуется подключение ФР к локальной сети (LAN).\n\nПодтверждаете переключение?",
			walk.MsgBoxYesNo|walk.MsgBoxIconQuestion)

		if res != walk.DlgCmdYes {
			// Откат значения
			st.viewModel.OfdClient = "1"
			// Принудительно обновляем UI
			if st.binder != nil {
				st.binder.Reset()
			}
		}
	}
}

func (st *ServiceTab) onTechReset() {
	drv := st.app.GetDriver()
	if drv == nil {
		return
	}

	if walk.MsgBox(st.app.MainWindow, "ВНИМАНИЕ",
		"Выполнить ТЕХНОЛОГИЧЕСКОЕ ОБНУЛЕНИЕ?\nЭто может привести к сбросу настроек и потере данных в ОЗУ.",
		walk.MsgBoxYesNo|walk.MsgBoxIconWarning) != walk.DlgCmdYes {
		return
	}

	go func() {
		err := drv.TechReset()
		st.app.MainWindow.Synchronize(func() {
			if err != nil {
				walk.MsgBox(st.app.MainWindow, "Ошибка", "Сбой тех. обнуления: "+err.Error(), walk.MsgBoxIconError)
			} else {
				walk.MsgBox(st.app.MainWindow, "Успех", "Технологическое обнуление выполнено.", walk.MsgBoxIconInformation)
				// Перечитываем настройки, так как они сбросились
				st.isLoadingData = true
				st.readAllParameters()
				st.readNetworkSettings()
				st.isLoadingData = false
			}
		})
	}()
}

func (st *ServiceTab) onSyncTime() {
	drv := st.app.GetDriver()
	if drv == nil {
		return
	}
	now := time.Now()
	go func() {
		err := drv.SetDateTime(now)
		st.app.MainWindow.Synchronize(func() {
			if err != nil {
				walk.MsgBox(st.app.MainWindow, "Ошибка", "Ошибка синхронизации: "+err.Error(), walk.MsgBoxIconError)
			} else {
				st.viewModel.KktTimeStr = now.Format("02.01.2006 15:04:05")
				st.binder.Reset()
				walk.MsgBox(st.app.MainWindow, "Успех", "Время синхронизировано.", walk.MsgBoxIconInformation)
			}
		})
	}()
}

func (st *ServiceTab) onOpenDrawer() {
	drv := st.app.GetDriver()
	if drv == nil {
		return
	}
	go func() { drv.DeviceJob(2) }()
}

func (st *ServiceTab) onPrintXReport() {
	drv := st.app.GetDriver()
	if drv == nil {
		return
	}
	go func() { drv.PrintXReport() }()
}

func (st *ServiceTab) onMGMReset() {
	drv := st.app.GetDriver()
	if drv == nil {
		return
	}
	go func() { drv.ResetMGM() }()
}

func (st *ServiceTab) onFeedAndCut() {
	drv := st.app.GetDriver()
	if drv == nil {
		return
	}
	go func() {
		drv.Feed(20)
		drv.Cut()
	}()
}

func (st *ServiceTab) onReadCliche() {
	drv := st.app.GetDriver()
	if drv == nil {
		walk.MsgBox(st.app.MainWindow, "Ошибка", "Нет подключения", walk.MsgBoxIconError)
		return
	}

	typeID, _ := strconv.Atoi(st.viewModel.SelectedClicheType)

	go func() {
		lines, err := drv.GetHeader(typeID)
		if err != nil {
			st.app.MainWindow.Synchronize(func() {
				walk.MsgBox(st.app.MainWindow, "Ошибка", fmt.Sprintf("Ошибка чтения клише: %v", err), walk.MsgBoxIconError)
			})
			return
		}

		st.app.MainWindow.Synchronize(func() {
			// Заполняем модель
			for i := 0; i < 10; i++ {
				if i < len(lines) {
					st.viewModel.ClicheItems[i].Text = lines[i].Text
					st.viewModel.ClicheItems[i].Format = lines[i].Format
				} else {
					st.viewModel.ClicheItems[i].Text = ""
					st.viewModel.ClicheItems[i].Format = "000000"
				}
				st.viewModel.ClicheItems[i].ParseFormatString()
			}
			st.clicheModel.PublishRowsReset()
			idx := st.clicheTable.CurrentIndex()
			if idx >= 0 {
				st.reloadEditor(idx)
			}
		})
	}()
}

func (st *ServiceTab) onWriteCliche() {
	drv := st.app.GetDriver()
	if drv == nil {
		return
	}

	typeID, _ := strconv.Atoi(st.viewModel.SelectedClicheType)

	var linesToWrite []struct{ txt, fmt string }
	for _, item := range st.viewModel.ClicheItems {
		item.UpdateFormatString()
		linesToWrite = append(linesToWrite, struct{ txt, fmt string }{item.Text, item.Format})
	}

	go func() {
		for i, l := range linesToWrite {
			if err := drv.SetHeaderLine(typeID, i, l.txt, l.fmt); err != nil {
				fmt.Printf("Error writing line %d: %v\n", i, err)
			}
		}
		st.app.MainWindow.Synchronize(func() {
			walk.MsgBox(st.app.MainWindow, "Успех", "Клише записано", walk.MsgBoxIconInformation)
		})
	}()
}

func (st *ServiceTab) onClicheSelectionChanged() {
	idx := st.clicheTable.CurrentIndex()
	if idx < 0 {
		st.clicheEditorGroup.SetEnabled(false)
		return
	}
	st.reloadEditor(idx)
}

func (st *ServiceTab) reloadEditor(idx int) {
	item := st.viewModel.ClicheItems[idx]
	st.clicheEditorBinder.SetDataSource(item)
	st.clicheEditorBinder.Reset()

	st.clicheEditorGroup.SetEnabled(true)
	st.clicheEditorGroup.SetTitle(fmt.Sprintf("Настройки строки №%d", idx+1))
}

func (st *ServiceTab) onClicheItemChanged() {
	idx := st.clicheTable.CurrentIndex()
	if idx >= 0 {
		item := st.viewModel.ClicheItems[idx]
		item.UpdateFormatString()
		st.clicheModel.PublishRowChanged(idx)
	}
}

func (st *ServiceTab) onActionButtonClicked() {
	if st.viewModel.OriginalSnapshot == nil {
		st.app.MainWindow.Synchronize(func() { st.isLoadingData = true })
		st.readNetworkSettings()
		st.readAllParameters()
		st.app.MainWindow.Synchronize(func() {
			st.isLoadingData = false
			st.viewModel.OriginalSnapshot = st.viewModel.CreateSnapshot()
			st.isWriteMode = false
			st.actionButton.SetText("Прочитать всё")
		})
		return
	}

	if !st.isWriteMode {
		st.app.MainWindow.Synchronize(func() { st.isLoadingData = true })
		st.readNetworkSettings()
		st.readAllParameters()
		st.app.MainWindow.Synchronize(func() {
			st.isLoadingData = false
			st.viewModel.OriginalSnapshot = st.viewModel.CreateSnapshot()
			st.isWriteMode = false
			st.actionButton.SetText("Прочитать всё")
		})
	} else {
		newSnap := st.viewModel.CreateSnapshot()
		changes := settings.CompareSnapshots(st.viewModel.OriginalSnapshot, newSnap)
		if len(changes) == 0 {
			return
		}
		confirmed, ok := RunSummaryDialog(st.app.MainWindow, changes)
		if !ok {
			return
		}
		go func() {
			if st.performBatchWrite(confirmed) {
				st.app.MainWindow.Synchronize(func() {
					st.viewModel.OriginalSnapshot = st.viewModel.CreateSnapshot()
					st.updateChangeTracking()
				})
			}
		}()
	}
}

func (st *ServiceTab) performBatchWrite(confirmed []*models.ChangeItem) bool {
	drv := st.app.GetDriver()
	if drv == nil {
		st.app.MainWindow.Synchronize(func() {
			walk.MsgBox(st.app.MainWindow, "Ошибка", "Нет подключения к устройству", walk.MsgBoxIconError)
		})
		return false
	}
	categoryChanges := make(map[string][]*models.ChangeItem)
	for _, ch := range confirmed {
		categoryChanges[ch.Category] = append(categoryChanges[ch.Category], ch)
	}
	var errors []string
	if _, ok := categoryChanges["ОФД"]; ok {
		h, p := splitHostPort(st.viewModel.OfdString)
		err := drv.SetOfdSettings(driver.OfdSettings{
			Addr: h, Port: p, Client: st.viewModel.OfdClient,
			TimerFN: st.viewModel.TimerFN, TimerOFD: st.viewModel.TimerOFD,
		})
		if err != nil {
			errors = append(errors, fmt.Sprintf("Ошибка записи ОФД: %v", err))
		}
	}
	if _, ok := categoryChanges["ОИСМ"]; ok {
		h, p := splitHostPort(st.viewModel.OismString)
		err := drv.SetOismSettings(driver.ServerSettings{Addr: h, Port: p})
		if err != nil {
			errors = append(errors, fmt.Sprintf("Ошибка записи ОИСМ: %v", err))
		}
	}
	if _, ok := categoryChanges["LAN"]; ok {
		err := drv.SetLanSettings(driver.LanSettings{
			Addr: st.viewModel.LanAddr, Mask: st.viewModel.LanMask,
			Port: st.viewModel.LanPort, Dns: st.viewModel.LanDns, Gw: st.viewModel.LanGw,
		})
		if err != nil {
			errors = append(errors, fmt.Sprintf("Ошибка записи LAN: %v", err))
		}
	}
	if _, ok := categoryChanges["Принтер"]; ok {
		baudRate, _ := strconv.Atoi(st.viewModel.PrintBaud)
		err := drv.SetPrinterSettings(driver.PrinterSettings{
			Model: st.viewModel.PrintModel, BaudRate: baudRate,
			Paper: st.viewModel.PrintPaper, Font: st.viewModel.PrintFont,
		})
		if err != nil {
			errors = append(errors, fmt.Sprintf("Ошибка записи принтера: %v", err))
		}
	}
	if changes, ok := categoryChanges["Опции"]; ok {
		for _, ch := range changes {
			switch ch.Name {
			case "Часовой пояс":
				tz, _ := strconv.Atoi(st.viewModel.OptTimezone)
				err := drv.SetTimezone(tz)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Ошибка записи часового пояса: %v", err))
				}
			case "Автоотрез":
				value := 0
				if st.viewModel.OptCut {
					value = 1
				}
				err := drv.SetOption(3, value)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Ошибка записи автоотреза: %v", err))
				}
			case "Звук бумаги":
				value := 0
				if st.viewModel.OptNearEnd {
					value = 1
				}
				err := drv.SetOption(6, value)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Ошибка записи звука бумаги: %v", err))
				}
			case "Автотест":
				value := 0
				if st.viewModel.OptAutoTest {
					value = 1
				}
				err := drv.SetOption(4, value)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Ошибка записи автотеста: %v", err))
				}
			case "Позиция QR":
				value, _ := strconv.Atoi(st.viewModel.OptQRPos)
				err := drv.SetOption(1, value)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Ошибка записи позиции QR: %v", err))
				}
			case "Округление":
				value, _ := strconv.Atoi(st.viewModel.OptRounding)
				err := drv.SetOption(2, value)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Ошибка записи округления: %v", err))
				}
			case "Триггер ящика":
				value, _ := strconv.Atoi(st.viewModel.OptDrawerTrig)
				err := drv.SetOption(5, value)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Ошибка записи триггера ящика: %v", err))
				}
			case "Текст у QR":
				value := 0
				if st.viewModel.OptTextQR {
					value = 1
				}
				err := drv.SetOption(7, value)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Ошибка записи текста у QR: %v", err))
				}
			case "Количество покупок":
				value := 0
				if st.viewModel.OptCountInCheck {
					value = 1
				}
				err := drv.SetOption(8, value)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Ошибка записи количества покупок: %v", err))
				}
			case "Опция b9":
				value, _ := strconv.Atoi(st.viewModel.OptB9)
				err := drv.SetOption(9, value)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Ошибка записи опции b9: %v", err))
				}
			}
		}
	}
	if _, ok := categoryChanges["Денежный ящик"]; ok {
		err := drv.SetMoneyDrawerSettings(driver.DrawerSettings{
			Pin: st.viewModel.DrawerPin, Rise: st.viewModel.DrawerRise, Fall: st.viewModel.DrawerFall,
		})
		if err != nil {
			errors = append(errors, fmt.Sprintf("Ошибка записи денежного ящика: %v", err))
		}
	}
	if changes, ok := categoryChanges["Клише"]; ok {
		typeID, _ := strconv.Atoi(st.viewModel.SelectedClicheType)
		for _, ch := range changes {
			var lineNum int
			n, err := fmt.Sscanf(ch.Name, "Строка %d", &lineNum)
			if err != nil || n != 1 {
				continue
			}
			lineNum--
			if lineNum < 0 || lineNum >= len(st.viewModel.ClicheItems) {
				continue
			}
			item := st.viewModel.ClicheItems[lineNum]
			item.UpdateFormatString()
			err = drv.SetHeaderLine(typeID, lineNum, item.Text, item.Format)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Ошибка записи клише строки %d: %v", lineNum+1, err))
			}
		}
	}
	if len(errors) > 0 {
		errorMsg := "Ошибки при записи:\n" + strings.Join(errors, "\n")
		st.app.MainWindow.Synchronize(func() {
			walk.MsgBox(st.app.MainWindow, "Ошибка", errorMsg, walk.MsgBoxIconError)
		})
		return false
	}
	st.app.MainWindow.Synchronize(func() {
		walk.MsgBox(st.app.MainWindow, "Успех", "Изменения записаны успешно", walk.MsgBoxIconInformation)
	})
	return true
}

func (st *ServiceTab) updateChangeTracking() {
	if st.isLoadingData || st.viewModel.OriginalSnapshot == nil {
		return
	}
	if st.normalFont == nil {
		var err error
		st.normalFont, err = walk.NewFont("Segoe UI", 9, 0)
		if err != nil {
			return
		}
		st.boldFont, err = walk.NewFont("Segoe UI", 9, walk.FontBold)
		if err != nil {
			return
		}
	}
	newSnap := st.viewModel.CreateSnapshot()
	changes := settings.CompareSnapshots(st.viewModel.OriginalSnapshot, newSnap)
	changedFields := make(map[string]bool)
	for _, ch := range changes {
		switch ch.Category {
		case "ОФД":
			if ch.Name == "Сервер" {
				changedFields["OfdString"] = true
			} else if ch.Name == "Режим клиента" {
				changedFields["OfdClient"] = true
			} else if ch.Name == "Таймер ФН" {
				changedFields["TimerFN"] = true
			} else if ch.Name == "Таймер ОФД" {
				changedFields["TimerOFD"] = true
			}
		case "ОИСМ":
			changedFields["OismString"] = true
		case "LAN":
			if ch.Name == "IP адрес" {
				changedFields["LanAddr"] = true
			} else if ch.Name == "Порт" {
				changedFields["LanPort"] = true
			} else if ch.Name == "Маска" {
				changedFields["LanMask"] = true
			} else if ch.Name == "DNS" {
				changedFields["LanDns"] = true
			} else if ch.Name == "Шлюз" {
				changedFields["LanGw"] = true
			}
		case "Принтер":
			if ch.Name == "Модель" {
				changedFields["PrintModel"] = true
			} else if ch.Name == "Скорость" {
				changedFields["PrintBaud"] = true
			} else if ch.Name == "Ширина бумаги" {
				changedFields["PrintPaper"] = true
			} else if ch.Name == "Шрифт" {
				changedFields["PrintFont"] = true
			}
		case "Опции":
			if ch.Name == "Часовой пояс" {
				changedFields["OptTimezone"] = true
			} else if ch.Name == "Автоотрез" {
				changedFields["OptCut"] = true
			} else if ch.Name == "Звук бумаги" {
				changedFields["OptNearEnd"] = true
			} else if ch.Name == "Автотест" {
				changedFields["OptAutoTest"] = true
			} else if ch.Name == "Позиция QR" {
				changedFields["OptQRPos"] = true
			} else if ch.Name == "Округление" {
				changedFields["OptRounding"] = true
			} else if ch.Name == "Триггер ящика" {
				changedFields["OptDrawerTrig"] = true
			} else if ch.Name == "Текст у QR" {
				changedFields["OptTextQR"] = true
			} else if ch.Name == "Количество покупок" {
				changedFields["OptCountInCheck"] = true
			} else if ch.Name == "Опция b9" {
				changedFields["OptB9"] = true
			}
		case "Денежный ящик":
			if ch.Name == "PIN" {
				changedFields["DrawerPin"] = true
			} else if ch.Name == "Rise (ms)" {
				changedFields["DrawerRise"] = true
			} else if ch.Name == "Fall (ms)" {
				changedFields["DrawerFall"] = true
			}
		case "Клише":
			changedFields["Cliche"] = true
		}
	}
	if len(changes) > 0 {
		st.isWriteMode = true
		st.actionButton.SetText("Записать изменения")
		st.actionButton.SetFont(st.boldFont)
	} else {
		st.isWriteMode = false
		st.actionButton.SetText("Прочитать всё")
		st.actionButton.SetFont(st.normalFont)
	}
	// Highlight labels
	if st.ofdStringLabel != nil {
		if changedFields["OfdString"] {
			st.ofdStringLabel.SetFont(st.boldFont)
		} else {
			st.ofdStringLabel.SetFont(st.normalFont)
		}
	}
	if st.oismStringLabel != nil {
		if changedFields["OismString"] {
			st.oismStringLabel.SetFont(st.boldFont)
		} else {
			st.oismStringLabel.SetFont(st.normalFont)
		}
	}
	if st.lanAddrLabel != nil {
		if changedFields["LanAddr"] {
			st.lanAddrLabel.SetFont(st.boldFont)
		} else {
			st.lanAddrLabel.SetFont(st.normalFont)
		}
	}
	if st.lanMaskLabel != nil {
		if changedFields["LanMask"] {
			st.lanMaskLabel.SetFont(st.boldFont)
		} else {
			st.lanMaskLabel.SetFont(st.normalFont)
		}
	}
	if st.lanDnsLabel != nil {
		if changedFields["LanDns"] {
			st.lanDnsLabel.SetFont(st.boldFont)
		} else {
			st.lanDnsLabel.SetFont(st.normalFont)
		}
	}
	if st.lanGwLabel != nil {
		if changedFields["LanGw"] {
			st.lanGwLabel.SetFont(st.boldFont)
		} else {
			st.lanGwLabel.SetFont(st.normalFont)
		}
	}
	if st.lanPortLabel != nil {
		if changedFields["LanPort"] {
			st.lanPortLabel.SetFont(st.boldFont)
		} else {
			st.lanPortLabel.SetFont(st.normalFont)
		}
	}
	if st.ofdClientLabel != nil {
		if changedFields["OfdClient"] {
			st.ofdClientLabel.SetFont(st.boldFont)
		} else {
			st.ofdClientLabel.SetFont(st.normalFont)
		}
	}
	if st.timerFNLabel != nil {
		if changedFields["TimerFN"] {
			st.timerFNLabel.SetFont(st.boldFont)
		} else {
			st.timerFNLabel.SetFont(st.normalFont)
		}
	}
	if st.timerOFDLabel != nil {
		if changedFields["TimerOFD"] {
			st.timerOFDLabel.SetFont(st.boldFont)
		} else {
			st.timerOFDLabel.SetFont(st.normalFont)
		}
	}
	if st.printModelLabel != nil {
		if changedFields["PrintModel"] {
			st.printModelLabel.SetFont(st.boldFont)
		} else {
			st.printModelLabel.SetFont(st.normalFont)
		}
	}
	if st.printBaudLabel != nil {
		if changedFields["PrintBaud"] {
			st.printBaudLabel.SetFont(st.boldFont)
		} else {
			st.printBaudLabel.SetFont(st.normalFont)
		}
	}
	if st.printPaperLabel != nil {
		if changedFields["PrintPaper"] {
			st.printPaperLabel.SetFont(st.boldFont)
		} else {
			st.printPaperLabel.SetFont(st.normalFont)
		}
	}
	if st.printFontLabel != nil {
		if changedFields["PrintFont"] {
			st.printFontLabel.SetFont(st.boldFont)
		} else {
			st.printFontLabel.SetFont(st.normalFont)
		}
	}
	if st.optTimezoneLabel != nil {
		if changedFields["OptTimezone"] {
			st.optTimezoneLabel.SetFont(st.boldFont)
		} else {
			st.optTimezoneLabel.SetFont(st.normalFont)
		}
	}
	if st.optCutLabel != nil {
		if changedFields["OptCut"] {
			st.optCutLabel.SetFont(st.boldFont)
		} else {
			st.optCutLabel.SetFont(st.normalFont)
		}
	}
	if st.optNearEndLabel != nil {
		if changedFields["OptNearEnd"] {
			st.optNearEndLabel.SetFont(st.boldFont)
		} else {
			st.optNearEndLabel.SetFont(st.normalFont)
		}
	}
	if st.optAutoTestLabel != nil {
		if changedFields["OptAutoTest"] {
			st.optAutoTestLabel.SetFont(st.boldFont)
		} else {
			st.optAutoTestLabel.SetFont(st.normalFont)
		}
	}
	if st.optQRPosLabel != nil {
		if changedFields["OptQRPos"] {
			st.optQRPosLabel.SetFont(st.boldFont)
		} else {
			st.optQRPosLabel.SetFont(st.normalFont)
		}
	}
	if st.optRoundingLabel != nil {
		if changedFields["OptRounding"] {
			st.optRoundingLabel.SetFont(st.boldFont)
		} else {
			st.optRoundingLabel.SetFont(st.normalFont)
		}
	}
	if st.optDrawerTrigLabel != nil {
		if changedFields["OptDrawerTrig"] {
			st.optDrawerTrigLabel.SetFont(st.boldFont)
		} else {
			st.optDrawerTrigLabel.SetFont(st.normalFont)
		}
	}
	if st.optTextQRLabel != nil {
		if changedFields["OptTextQR"] {
			st.optTextQRLabel.SetFont(st.boldFont)
		} else {
			st.optTextQRLabel.SetFont(st.normalFont)
		}
	}
	if st.optCountInCheckLabel != nil {
		if changedFields["OptCountInCheck"] {
			st.optCountInCheckLabel.SetFont(st.boldFont)
		} else {
			st.optCountInCheckLabel.SetFont(st.normalFont)
		}
	}
	if st.optB9Label != nil {
		if changedFields["OptB9"] {
			st.optB9Label.SetFont(st.boldFont)
		} else {
			st.optB9Label.SetFont(st.normalFont)
		}
	}
	if st.drawerPinLabel != nil {
		if changedFields["DrawerPin"] {
			st.drawerPinLabel.SetFont(st.boldFont)
		} else {
			st.drawerPinLabel.SetFont(st.normalFont)
		}
	}
	if st.drawerRiseLabel != nil {
		if changedFields["DrawerRise"] {
			st.drawerRiseLabel.SetFont(st.boldFont)
		} else {
			st.drawerRiseLabel.SetFont(st.normalFont)
		}
	}
	if st.drawerFallLabel != nil {
		if changedFields["DrawerFall"] {
			st.drawerFallLabel.SetFont(st.boldFont)
		} else {
			st.drawerFallLabel.SetFont(st.normalFont)
		}
	}
}
