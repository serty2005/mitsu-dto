package gui

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"

	"mitsuscanner/driver"
	"mitsuscanner/internal/service"
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
	// Значения должны быть "80" и "57", а не индексы "0"/"1"
	listPapers = []*NV{
		{"80мм", "80"}, {"57мм", "57"},
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
	// ВАЖНО: Все поля для ComboBox должны быть string, иначе DataBinder не выберет значение
	PrintModel string // "1", "2"
	PrintBaud  string // "115200"
	PrintPaper string // "80", "57" (БЫЛО int, стало string)
	PrintFont  string // "0", "1"   (БЫЛО int, стало string)

	// Опции (b0-b9)
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

	// Денежный ящик (остаются int, т.к. используются в NumberEdit)
	DrawerPin  int
	DrawerRise int
	DrawerFall int

	// --- Клише ---
	SelectedClicheType string        // "1".."4"
	ClicheItems        []*ClicheItem // 10 строк
	CurrentClicheLine  *ClicheItem   // Указатель на редактируемую строку
}

// ServiceLabels хранит ссылки на лейблы для выделения жирным шрифтом
type ServiceLabels struct {
	// Printer
	PrinterModel, PrinterBaud, PrinterPaper, PrinterFont, PrinterCut, PrinterSound, PrinterTest *walk.Label
	// Drawer
	DrawerPin, DrawerRise, DrawerFall, DrawerTrig *walk.Label
	// Network
	OfdAddr, OfdClient, TimerFN, TimerOFD, OismAddr, Timezone *walk.Label
	LanIp, LanMask, LanGw, LanPort                            *walk.Label
	// Options
	OptQRPos, OptTextQR, OptCount, OptRounding, OptB9 *walk.Label
	// Cliche
	ClicheHeader *walk.Label
}

var (
	serviceModel  *ServiceViewModel
	serviceBinder *walk.DataBinder
	sLabels       ServiceLabels

	// Глобальные снапшоты состояния
	initialSnapshot *service.SettingsSnapshot
	currentSnapshot *service.SettingsSnapshot
	currentChanges  []service.Change

	// Элементы управления
	btnServiceAction *walk.PushButton

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
	clicheEditorBinder *walk.DataBinder
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
// ЛОГИКА СНАПШОТОВ И СРАВНЕНИЯ
// -----------------------------

// viewModelToSnapshot конвертирует текущую ViewModel в структуру Snapshot.
// viewModelToSnapshot конвертирует текущую ViewModel в структуру Snapshot.
func viewModelToSnapshot(vm *ServiceViewModel) *service.SettingsSnapshot {
	s := service.NewSettingsSnapshot()

	// 1. ОФД
	host, port := splitHostPort(vm.OfdString)
	s.Ofd = driver.OfdSettings{
		Addr:     host,
		Port:     port,
		Client:   vm.OfdClient,
		TimerFN:  vm.TimerFN,
		TimerOFD: vm.TimerOFD,
	}

	// 2. ОИСМ
	hostO, portO := splitHostPort(vm.OismString)
	s.Oism = driver.ServerSettings{Addr: hostO, Port: portO}

	// 3. LAN
	s.Lan = driver.LanSettings{
		Addr: vm.LanAddr, Mask: vm.LanMask, Port: vm.LanPort,
		Dns: vm.LanDns, Gw: vm.LanGw,
	}

	// 4. Timezone
	s.Timezone, _ = strconv.Atoi(vm.OptTimezone)

	// 5. Printer
	// Преобразуем строки из UI обратно в int для драйвера
	baud, _ := strconv.Atoi(vm.PrintBaud)
	paper, _ := strconv.Atoi(vm.PrintPaper)
	font, _ := strconv.Atoi(vm.PrintFont)

	s.Printer = driver.PrinterSettings{
		Model:    vm.PrintModel,
		BaudRate: baud,
		Paper:    paper,
		Font:     font,
	}

	// 6. Drawer
	s.Drawer = driver.DrawerSettings{
		Pin:  vm.DrawerPin,
		Rise: vm.DrawerRise,
		Fall: vm.DrawerFall,
	}

	// 7. Options
	opts := driver.DeviceOptions{}
	opts.B1, _ = strconv.Atoi(vm.OptQRPos)
	opts.B2, _ = strconv.Atoi(vm.OptRounding)
	opts.B3 = boolToInt(vm.OptCut)
	opts.B4 = boolToInt(vm.OptAutoTest)
	opts.B5, _ = strconv.Atoi(vm.OptDrawerTrig)
	opts.B6 = boolToInt(vm.OptNearEnd)
	opts.B7 = boolToInt(vm.OptTextQR)
	opts.B8 = boolToInt(vm.OptCountInCheck)
	opts.B9, _ = strconv.Atoi(vm.OptB9)
	s.Options = opts

	// 8. Cliches
	if currentSnapshot != nil {
		for k, v := range currentSnapshot.Cliches {
			dst := make([]driver.ClicheLineData, len(v))
			copy(dst, v)
			s.Cliches[k] = dst
		}
	}

	curType, _ := strconv.Atoi(vm.SelectedClicheType)
	var lines []driver.ClicheLineData
	for _, item := range vm.ClicheItems {
		item.UpdateFormatString()
		lines = append(lines, driver.ClicheLineData{
			Text:   item.Text,
			Format: item.Format,
		})
	}
	s.Cliches[curType] = lines

	return s
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func onServiceAction() {
	if len(currentChanges) > 0 {
		onWriteAllParameters()
	} else {
		onReadAllSettings()
	}
}

// recalcChanges вызывается при любом изменении в UI
func recalcChanges() {
	if isLoadingData || serviceBinder == nil || initialSnapshot == nil {
		return
	}

	// 1. Submit данных из UI в ViewModel
	if err := serviceBinder.Submit(); err != nil {
		// Ошибка валидации или биндинга, пока игнорируем
		return
	}

	// 2. Обновляем текущий снапшот на основе VM
	currentSnapshot = viewModelToSnapshot(serviceModel)

	// 3. Сравниваем с начальным
	currentChanges = service.Compare(initialSnapshot, currentSnapshot)

	// 4. Обновляем UI кнопки
	mw.Synchronize(func() {
		count := len(currentChanges)
		if count > 0 {
			btnServiceAction.SetText(fmt.Sprintf("Записать (%d)", count))
			// Можно добавить визуальный акцент, если нужно
		} else {
			btnServiceAction.SetText("Считать")
		}
		// Кнопка всегда активна, если мы не в процессе загрузки (isLoadingData)
		btnServiceAction.SetEnabled(!isLoadingData)

		// 5. Подсветка лейблов
		highlightLabels(currentChanges)
	})
}

// highlightLabels проходит по списку изменений и делает начертание лейблов жирным.
func highlightLabels(changes []service.Change) {
	// Карта соответствия атомарных ID из компаратора к лейблам в UI.
	// Используем **walk.Label, так как переменные в sLabels инициализируются декларативно.
	labelMap := map[string][]**walk.Label{
		"Printer_Model":    {&sLabels.PrinterModel},
		"Printer_Baud":     {&sLabels.PrinterBaud},
		"Printer_Paper":    {&sLabels.PrinterPaper},
		"Printer_Font":     {&sLabels.PrinterFont},
		"Drawer_Settings":  {&sLabels.DrawerPin, &sLabels.DrawerRise, &sLabels.DrawerFall},
		"Timezone":         {&sLabels.Timezone},
		"Opt_QRPos":        {&sLabels.OptQRPos},
		"Opt_Rounding":     {&sLabels.OptRounding},
		"Opt_Cut":          {&sLabels.PrinterCut},
		"Opt_AutoTest":     {&sLabels.PrinterTest},
		"Opt_DrawerTrig":   {&sLabels.DrawerTrig},
		"Opt_NearEnd":      {&sLabels.PrinterSound},
		"Opt_TextQR":       {&sLabels.OptTextQR},
		"Opt_CountInCheck": {&sLabels.OptCount},
		"Opt_B9":           {&sLabels.OptB9},
		"Ofd_Addr":         {&sLabels.OfdAddr},
		"Ofd_Client":       {&sLabels.OfdClient},
		"Ofd_Timers":       {&sLabels.TimerFN, &sLabels.TimerOFD},
		"Oism_Addr":        {&sLabels.OismAddr},
		"Lan_Settings":     {&sLabels.LanIp, &sLabels.LanMask, &sLabels.LanGw, &sLabels.LanPort},
	}

	// Функция-помощник для смены начертания без изменения размера и шрифта
	setBold := func(lbPtr **walk.Label, bold bool) {
		if lbPtr == nil || *lbPtr == nil {
			return
		}
		lbl := *lbPtr
		f := lbl.Font()
		style := walk.FontStyle(0)
		if bold {
			style = walk.FontBold
		}

		// Создаем новый шрифт, сохраняя Family и PointSize
		newFont, err := walk.NewFont(f.Family(), f.PointSize(), style)
		if err == nil {
			lbl.SetFont(newFont)
		}
	}

	// 1. Сначала сбрасываем все лейблы в обычное начертание
	for _, labels := range labelMap {
		for _, lblPtr := range labels {
			setBold(lblPtr, false)
		}
	}
	if sLabels.ClicheHeader != nil {
		setBold(&sLabels.ClicheHeader, false)
	}

	// 2. Устанавливаем жирное начертание для измененных параметров
	for _, ch := range changes {
		// Обработка клише (динамический ID)
		if len(ch.ID) > 7 && ch.ID[:7] == "Cliche_" {
			if sLabels.ClicheHeader != nil {
				setBold(&sLabels.ClicheHeader, true)
			}
			continue
		}

		// Остальные параметры по карте
		if labels, ok := labelMap[ch.ID]; ok {
			for _, lblPtr := range labels {
				setBold(lblPtr, true)
			}
		}
	}
}

// -----------------------------
// ЗАГРУЗКА ДАННЫХ
// -----------------------------

func loadServiceInitial() {
	go func() {
		// Ждем инициализации окна
		for i := 0; i < 20; i++ {
			if mw != nil && mw.Handle() != 0 {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if mw == nil {
			return
		}
		onReadAllSettings()
	}()
}

func onReadAllSettings() {
	drv := driver.Active
	if drv == nil {
		mw.Synchronize(func() {
			serviceModel.KktTimeStr = "Нет подключения"
			serviceModel.PcTimeStr = time.Now().Format("02.01.2006 15:04:05")
			serviceBinder.Reset()
			btnServiceAction.SetText("Считать настройки")
		})
		return
	}

	mw.Synchronize(func() {
		// Блокируем триггеры изменений на время загрузки
		isLoadingData = true
		btnServiceAction.SetEnabled(false)
		btnServiceAction.SetText("Чтение...")
	})

	go func() {
		// Читаем всё из ККТ (время, настройки, клише)
		t, _ := drv.GetDateTime()
		ofd, _ := drv.GetOfdSettings()
		oism, _ := drv.GetOismSettings()
		lan, _ := drv.GetLanSettings()
		prn, _ := drv.GetPrinterSettings()
		cd, _ := drv.GetMoneyDrawerSettings()
		opts, _ := drv.GetOptions()
		tz, _ := drv.GetTimezone()

		allCliches := make(map[int][]driver.ClicheLineData)
		for i := 1; i <= 4; i++ {
			lines, err := drv.GetHeader(i)
			if err == nil {
				allCliches[i] = lines
			}
		}

		mw.Synchronize(func() {
			// 1. Обновляем модель времени
			if !t.IsZero() {
				serviceModel.KktTimeStr = t.Format("02.01.2006 15:04:05")
			}

			// 2. Обновляем основные поля в ViewModel
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
			if prn != nil {
				serviceModel.PrintModel = prn.Model
				serviceModel.PrintBaud = strconv.Itoa(prn.BaudRate)
				serviceModel.PrintPaper = strconv.Itoa(prn.Paper)
				serviceModel.PrintFont = strconv.Itoa(prn.Font)
			}
			if cd != nil {
				serviceModel.DrawerPin = cd.Pin
				serviceModel.DrawerRise = cd.Rise
				serviceModel.DrawerFall = cd.Fall
			}
			serviceModel.OptTimezone = strconv.Itoa(tz)
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

			// 3. Обновляем клише (текущий выбранный тип)
			curType, _ := strconv.Atoi(serviceModel.SelectedClicheType)
			lines := allCliches[curType]
			for i := 0; i < 10; i++ {
				if i < len(lines) {
					serviceModel.ClicheItems[i].Text = lines[i].Text
					serviceModel.ClicheItems[i].Format = lines[i].Format
				} else {
					serviceModel.ClicheItems[i].Text = ""
					serviceModel.ClicheItems[i].Format = "000000"
				}
				serviceModel.ClicheItems[i].ParseFormatString()
			}
			clicheModel.PublishRowsReset()

			// 4. Синхронизируем UI с обновленной моделью
			serviceBinder.Reset()

			// 5. СОЗДАЕМ ИДЕНТИЧНЫЕ СНАПШОТЫ
			// Сначала формируем Snapshot из того, что только что записали в VM
			tempSnap := viewModelToSnapshot(serviceModel)
			// Добавляем в него все вычитанные типы клише (не только текущий)
			tempSnap.Cliches = allCliches

			initialSnapshot = tempSnap
			currentSnapshot = tempSnap
			currentChanges = nil

			// 6. Снимаем блокировку и обновляем состояние кнопки
			isLoadingData = false
			btnServiceAction.SetEnabled(true)
			btnServiceAction.SetText("Считать настройки")

			// Принудительно сбрасываем подсветку
			highlightLabels(nil)
		})
	}()
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
		PrintPaper:         "80",
		PrintFont:          "0",
		DrawerPin:          5,
		DrawerRise:         100,
		DrawerFall:         100,
		OptTimezone:        "3",
		OptQRPos:           "1",
		OptRounding:        "0",
		OptDrawerTrig:      "1",
		OptCut:             true,
		OptB9:              "0",
		OfdClient:          "1",
		SelectedClicheType: "1",
		CurrentClicheLine:  &ClicheItem{},
	}

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
							// ОСНОВНЫЕ КНОПКИ УПРАВЛЕНИЯ НАСТРОЙКАМИ
							d.PushButton{
								AssignTo:  &btnServiceAction,
								Text:      "Считать",
								OnClicked: onServiceAction,
								MinSize:   d.Size{Width: 150},
							},
						},
					},
				},
			},

			// ТАБЫ ПОДКАТЕГОРИЙ
			d.TabWidget{
				MinSize: d.Size{Height: 300},
				Pages: []d.TabPage{

					// 1. ПАРАМЕТРЫ
					{
						Title:  "Параметры",
						Layout: d.VBox{MarginsZero: true, Spacing: 0, Alignment: d.AlignHNearVNear},
						Children: []d.Widget{
							d.Composite{
								Layout: d.HBox{Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}, Spacing: 4, Alignment: d.AlignHNearVNear},
								Children: []d.Widget{

									// КОЛОНКА 1
									d.Composite{
										Layout: d.VBox{MarginsZero: true, Spacing: 4},
										Children: []d.Widget{
											d.GroupBox{
												Title:  "ОФД и ОИСМ",
												Layout: d.Grid{Columns: 2, Spacing: 4, Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
												Children: []d.Widget{
													d.Label{AssignTo: &sLabels.OfdAddr, Text: "ОФД:"}, d.LineEdit{Text: d.Bind("OfdString"), MinSize: d.Size{Width: 110}, MaxSize: d.Size{Width: 120}, OnTextChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.OismAddr, Text: "ОИСМ:"}, d.LineEdit{Text: d.Bind("OismString"), MinSize: d.Size{Width: 110}, MaxSize: d.Size{Width: 120}, OnTextChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.Timezone, Text: "Пояс:"}, d.ComboBox{Value: d.Bind("OptTimezone"), BindingMember: "Code", DisplayMember: "Name", Model: listTimezones, MinSize: d.Size{Width: 110}, MaxSize: d.Size{Width: 120}, OnCurrentIndexChanged: recalcChanges},
												},
											},
											d.GroupBox{
												Title:  "Принтер и Бумага",
												Layout: d.Grid{Columns: 4, Spacing: 4, Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
												Children: []d.Widget{
													d.Label{AssignTo: &sLabels.PrinterModel, Text: "Модель:"}, d.ComboBox{Value: d.Bind("PrintModel"), BindingMember: "Code", DisplayMember: "Name", Model: listModels, MaxSize: d.Size{Width: 70}, OnCurrentIndexChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.PrinterCut, Text: "Отрез:"}, d.CheckBox{Checked: d.Bind("OptCut"), OnCheckStateChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.PrinterBaud, Text: "Бод:"}, d.ComboBox{Value: d.Bind("PrintBaud"), BindingMember: "Code", DisplayMember: "Name", Model: listBaud, MaxSize: d.Size{Width: 70}, OnCurrentIndexChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.PrinterSound, Text: "Звук:"}, d.CheckBox{Checked: d.Bind("OptNearEnd"), OnCheckStateChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.PrinterPaper, Text: "Ширина:"}, d.ComboBox{Value: d.Bind("PrintPaper"), BindingMember: "Code", DisplayMember: "Name", Model: listPapers, MaxSize: d.Size{Width: 70}, OnCurrentIndexChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.PrinterTest, Text: "Тест:"}, d.CheckBox{Checked: d.Bind("OptAutoTest"), OnCheckStateChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.PrinterFont, Text: "Шрифт:"}, d.ComboBox{Value: d.Bind("PrintFont"), BindingMember: "Code", DisplayMember: "Name", Model: listFonts, MaxSize: d.Size{Width: 70}, ToolTipText: "A-стандратный, B-компактный", OnCurrentIndexChanged: recalcChanges},
												},
											},
											d.GroupBox{
												Title:  "Сеть (LAN)",
												Layout: d.Grid{Columns: 2, Spacing: 4, Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
												Children: []d.Widget{
													d.Label{AssignTo: &sLabels.LanIp, Text: "IP:"}, d.LineEdit{Text: d.Bind("LanAddr"), MinSize: d.Size{Width: 90}, MaxSize: d.Size{Width: 100}, OnTextChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.LanMask, Text: "Mask:"}, d.LineEdit{Text: d.Bind("LanMask"), MinSize: d.Size{Width: 90}, MaxSize: d.Size{Width: 100}, OnTextChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.LanGw, Text: "GW:"}, d.LineEdit{Text: d.Bind("LanGw"), MinSize: d.Size{Width: 90}, MaxSize: d.Size{Width: 100}, OnTextChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.LanPort, Text: "Port:"}, d.NumberEdit{Value: d.Bind("LanPort"), MaxSize: d.Size{Width: 60}, OnValueChanged: recalcChanges},
												},
											},
										},
									},

									// КОЛОНКА 2
									d.Composite{
										Layout: d.VBox{MarginsZero: true, Spacing: 4},
										Children: []d.Widget{
											d.GroupBox{
												Title:  "Настройки клиента",
												Layout: d.Grid{Columns: 2, Spacing: 4, Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
												Children: []d.Widget{
													d.Label{AssignTo: &sLabels.OfdClient, Text: "Режим:"}, d.ComboBox{Value: d.Bind("OfdClient"), BindingMember: "Code", DisplayMember: "Name", Model: listClients, OnCurrentIndexChanged: checkOfdClientChange, MaxSize: d.Size{Width: 100}},
													d.Label{AssignTo: &sLabels.TimerFN, Text: "Т. ФН:"}, d.NumberEdit{Value: d.Bind("TimerFN"), MaxSize: d.Size{Width: 40}, OnValueChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.TimerOFD, Text: "Т. ОФД:"}, d.NumberEdit{Value: d.Bind("TimerOFD"), MaxSize: d.Size{Width: 40}, OnValueChanged: recalcChanges},
												},
											},
											d.GroupBox{
												Title:  "Денежный ящик",
												Layout: d.Grid{Columns: 2, Spacing: 4, Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
												Children: []d.Widget{
													d.Label{AssignTo: &sLabels.DrawerTrig, Text: "Триггер:"}, d.ComboBox{Value: d.Bind("OptDrawerTrig"), BindingMember: "Code", DisplayMember: "Name", Model: listDrawerTrig, MaxSize: d.Size{Width: 80}, OnCurrentIndexChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.DrawerPin, Text: "PIN:"}, d.NumberEdit{Value: d.Bind("DrawerPin"), MaxSize: d.Size{Width: 40}, OnValueChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.DrawerRise, Text: "Rise:"}, d.NumberEdit{Value: d.Bind("DrawerRise"), MaxSize: d.Size{Width: 40}, OnValueChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.DrawerFall, Text: "Fall:"}, d.NumberEdit{Value: d.Bind("DrawerFall"), MaxSize: d.Size{Width: 40}, OnValueChanged: recalcChanges},
												},
											},
											d.GroupBox{
												Title:  "Вид чека и Опции",
												Layout: d.Grid{Columns: 2, Spacing: 4, Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
												Children: []d.Widget{
													d.Label{AssignTo: &sLabels.OptQRPos, Text: "QR:"}, d.ComboBox{Value: d.Bind("OptQRPos"), BindingMember: "Code", DisplayMember: "Name", Model: listQRPos, MaxSize: d.Size{Width: 80}, OnCurrentIndexChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.OptTextQR, Text: "Текст QR:"}, d.CheckBox{Checked: d.Bind("OptTextQR"), OnCheckStateChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.OptCount, Text: "Покупок:"}, d.CheckBox{Checked: d.Bind("OptCountInCheck"), OnCheckStateChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.OptRounding, Text: "Округл.:"}, d.ComboBox{Value: d.Bind("OptRounding"), BindingMember: "Code", DisplayMember: "Name", Model: listRounding, MaxSize: d.Size{Width: 60}, OnCurrentIndexChanged: recalcChanges},
													d.Label{AssignTo: &sLabels.OptB9, Text: "b9:"}, d.LineEdit{Text: d.Bind("OptB9"), MaxLength: 3, MaxSize: d.Size{Width: 30}, ToolTipText: "Сумма: СНО(1-8) + X-отчет(16)", OnTextChanged: recalcChanges},
												},
											},
										},
									},

									// КОЛОНКА 3
									// d.Composite{
									// 	Layout: d.VBox{MarginsZero: true, Spacing: 4},
									// 	Children: []d.Widget{
									// 		d.GroupBox{
									// 			Title:  "Сеть (LAN)",
									// 			Layout: d.Grid{Columns: 2, Spacing: 4, Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
									// 			Children: []d.Widget{
									// 				d.Label{AssignTo: &sLabels.LanIp, Text: "IP:"}, d.LineEdit{Text: d.Bind("LanAddr"), MinSize: d.Size{Width: 90}, MaxSize: d.Size{Width: 100}, OnTextChanged: recalcChanges},
									// 				d.Label{AssignTo: &sLabels.LanMask, Text: "Mask:"}, d.LineEdit{Text: d.Bind("LanMask"), MinSize: d.Size{Width: 90}, MaxSize: d.Size{Width: 100}, OnTextChanged: recalcChanges},
									// 				d.Label{AssignTo: &sLabels.LanGw, Text: "GW:"}, d.LineEdit{Text: d.Bind("LanGw"), MinSize: d.Size{Width: 90}, MaxSize: d.Size{Width: 100}, OnTextChanged: recalcChanges},
									// 				d.Label{AssignTo: &sLabels.LanPort, Text: "Port:"}, d.NumberEdit{Value: d.Bind("LanPort"), MaxSize: d.Size{Width: 60}, OnValueChanged: recalcChanges},
									// 			},
									// 		},
									// 		d.GroupBox{
									// 			Title:  "Вид чека и Опции",
									// 			Layout: d.Grid{Columns: 2, Spacing: 4, Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
									// 			Children: []d.Widget{
									// 				d.Label{AssignTo: &sLabels.OptQRPos, Text: "QR:"}, d.ComboBox{Value: d.Bind("OptQRPos"), BindingMember: "Code", DisplayMember: "Name", Model: listQRPos, MaxSize: d.Size{Width: 80}, OnCurrentIndexChanged: recalcChanges},
									// 				d.Label{AssignTo: &sLabels.OptTextQR, Text: "Текст QR:"}, d.CheckBox{Checked: d.Bind("OptTextQR"), OnCheckStateChanged: recalcChanges},
									// 				d.Label{AssignTo: &sLabels.OptCount, Text: "Покупок:"}, d.CheckBox{Checked: d.Bind("OptCountInCheck"), OnCheckStateChanged: recalcChanges},
									// 				d.Label{AssignTo: &sLabels.OptRounding, Text: "Округл.:"}, d.ComboBox{Value: d.Bind("OptRounding"), BindingMember: "Code", DisplayMember: "Name", Model: listRounding, MaxSize: d.Size{Width: 60}, OnCurrentIndexChanged: recalcChanges},
									// 				d.Label{AssignTo: &sLabels.OptB9, Text: "b9:"}, d.LineEdit{Text: d.Bind("OptB9"), MaxLength: 3, MaxSize: d.Size{Width: 30}, ToolTipText: "Сумма: СНО(1-8) + X-отчет(16)", OnTextChanged: recalcChanges},
									// 			},
									// 		},
									// 	},

								},
							},
						},
					},

					// 2. КЛИШЕ
					{
						Title:  "Клише",
						Layout: d.VBox{Margins: d.Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}, Spacing: 5},
						Children: []d.Widget{
							d.Composite{
								Layout: d.HBox{MarginsZero: true, Alignment: d.AlignHNearVCenter},
								Children: []d.Widget{
									d.Label{AssignTo: &sLabels.ClicheHeader, Text: "Редактировать:"},
									d.ComboBox{
										Value:         d.Bind("SelectedClicheType"),
										Model:         listClicheTypes,
										BindingMember: "Code", DisplayMember: "Name",
										MinSize:               d.Size{Width: 200},
										OnCurrentIndexChanged: onClicheTypeSwitched, // Специальный обработчик
									},
								},
							},
							d.Composite{
								Layout: d.HBox{MarginsZero: true, Spacing: 10},
								Children: []d.Widget{
									d.TableView{
										AssignTo:         &clicheTable,
										Model:            clicheModel,
										AlternatingRowBG: true,
										Columns: []d.TableViewColumn{
											{Title: "#", Width: 30},
											{Title: "Fmt", Width: 60},
											{Title: "Текст", Width: 200},
										},
										MinSize:               d.Size{Width: 300, Height: 200},
										MaxSize:               d.Size{Width: 300, Height: 200},
										OnCurrentIndexChanged: onClicheSelectionChanged,
									},
									d.GroupBox{
										AssignTo: &clicheEditorGroup,
										Title:    "Настройки строки",
										Layout:   d.VBox{Margins: d.Margins{Left: 10, Top: 10, Right: 10, Bottom: 10}, Spacing: 8},
										Enabled:  false,
										MaxSize:  d.Size{Width: 300, Height: 200},
										DataBinder: d.DataBinder{
											AssignTo:   &clicheEditorBinder,
											DataSource: serviceModel.CurrentClicheLine,
											AutoSubmit: true,
										},
										Children: []d.Widget{
											d.Label{Text: "Текст:"},
											d.LineEdit{
												Text:          d.Bind("Text"),
												OnTextChanged: onClicheItemChanged,
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
														OnCurrentIndexChanged: onClicheItemChanged,
													},
													d.Label{Text: "Шрифт:"},
													d.ComboBox{
														Value:                 d.Bind("Font"),
														Model:                 listFonts,
														BindingMember:         "Code",
														DisplayMember:         "Name",
														OnCurrentIndexChanged: onClicheItemChanged,
													},
													d.Label{Text: "Подчеркивание:"},
													d.ComboBox{
														Value:                 d.Bind("Underline"),
														Model:                 listUnderline,
														BindingMember:         "Code",
														DisplayMember:         "Name",
														OnCurrentIndexChanged: onClicheItemChanged,
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
														OnValueChanged: onClicheItemChanged,
													},
													d.Label{Text: "Высота:"},
													d.NumberEdit{
														Value:          d.Bind("Height"),
														MinValue:       0,
														MaxValue:       8,
														MaxSize:        d.Size{Width: 40},
														OnValueChanged: onClicheItemChanged,
													},
												},
											},
											d.CheckBox{
												Text:                "Инверсия (Белым по черному)",
												Checked:             d.Bind("Invert"),
												OnCheckStateChanged: func() { onClicheItemChanged() },
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
	recalcChanges() // Сразу вызываем пересчет

	if serviceModel.OfdClient == "0" {
		res := walk.MsgBox(mw, "Подтверждение",
			"Для использования встроенного клиента ОФД требуется подключение ФР к локальной сети (LAN).\n\nПодтверждаете переключение?",
			walk.MsgBoxYesNo|walk.MsgBoxIconQuestion)

		if res != walk.DlgCmdYes {
			serviceModel.OfdClient = "1"
			if serviceBinder != nil {
				serviceBinder.Reset()
			}
			recalcChanges()
		}
	}
}

func onClicheTypeSwitched() {
	if isLoadingData || initialSnapshot == nil {
		return
	}
	// Сохраняем состояние предыдущего выбора в currentSnapshot (это делает viewModelToSnapshot)
	// и загружаем новый выбор в UI.
	// 1. Сначала коммитим текущие изменения в currentSnapshot (через recalc)
	recalcChanges()

	// 2. Теперь загружаем данные для НОВОГО типа из currentSnapshot в ViewModel
	newType, _ := strconv.Atoi(serviceModel.SelectedClicheType)

	// В currentSnapshot уже хранятся актуальные данные для всех типов (т.к. мы делали DeepCopy при инициализации и обновляем в viewModelToSnapshot)
	lines := currentSnapshot.Cliches[newType]

	// Заполняем ViewModel данными из снапшота
	for i := 0; i < 10; i++ {
		if i < len(lines) {
			serviceModel.ClicheItems[i].Text = lines[i].Text
			serviceModel.ClicheItems[i].Format = lines[i].Format
		} else {
			serviceModel.ClicheItems[i].Text = ""
			serviceModel.ClicheItems[i].Format = "000000"
		}
		serviceModel.ClicheItems[i].ParseFormatString()
	}

	clicheModel.PublishRowsReset()
	if idx := clicheTable.CurrentIndex(); idx >= 0 {
		reloadEditor(idx)
	}
}

// -----------------------------
// ЛОГИКА ЗАПИСИ
// -----------------------------

func onWriteAllParameters() {
	// Если нет изменений, выходим (кнопка должна быть disabled, но на всякий случай)
	if len(currentChanges) == 0 {
		return
	}

	// 1. Показываем диалог со списком изменений
	confirmed, finalChanges := RunDiffDialog(mw, currentChanges)
	if !confirmed {
		return
	}

	if len(finalChanges) == 0 {
		return // Пользователь удалил все строки и нажал ОК
	}

	// 2. Запускаем пайплайн применения
	ApplyChangesPipeline(finalChanges)
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
				onReadAllSettings() // Перечитываем всё
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

// -----------------------------
// ЛОГИКА КЛИШЕ (НОВАЯ)
// -----------------------------

func onClicheSelectionChanged() {
	idx := clicheTable.CurrentIndex()
	if idx < 0 {
		clicheEditorGroup.SetEnabled(false)
		return
	}
	reloadEditor(idx)
}

func reloadEditor(idx int) {
	item := serviceModel.ClicheItems[idx]
	clicheEditorBinder.SetDataSource(item)
	clicheEditorBinder.Reset()

	clicheEditorGroup.SetEnabled(true)
	clicheEditorGroup.SetTitle(fmt.Sprintf("Настройки строки №%d", idx+1))
}

func onClicheItemChanged() {
	// При любом изменении в редакторе пересчитываем формат и изменения
	idx := clicheTable.CurrentIndex()
	if idx >= 0 {
		item := serviceModel.ClicheItems[idx]
		item.UpdateFormatString()
		clicheModel.PublishRowChanged(idx)
		recalcChanges()
	}
}
