package gui

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"

	"mitsuscanner/driver"
	"mitsuscanner/internal/cliche"
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
		{"UTC+11 (Маг)", "11"}, {"UTC+12 (Кам)", "12"}, {"Не настроено", "254"},
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

// ClicheItemWrapper обертка над internal/cliche.Line для GUI.
// Walk требует конкретных типов полей для DataBinding.
type ClicheItemWrapper struct {
	Index int
	Line  cliche.Line
}

// Геттеры/Сеттеры для DataBinder, чтобы мапить поля формы на cliche.Props

func (c *ClicheItemWrapper) Text() string     { return c.Line.Text }
func (c *ClicheItemWrapper) SetText(v string) { c.Line.Text = v }

func (c *ClicheItemWrapper) Format() string { return c.Line.Format }

func (c *ClicheItemWrapper) Invert() bool     { return c.Line.Props.Invert }
func (c *ClicheItemWrapper) SetInvert(v bool) { c.Line.Props.Invert = v; c.updateFormat() }

func (c *ClicheItemWrapper) Width() int     { return c.Line.Props.Width }
func (c *ClicheItemWrapper) SetWidth(v int) { c.Line.Props.Width = v; c.updateFormat() }

func (c *ClicheItemWrapper) Height() int     { return c.Line.Props.Height }
func (c *ClicheItemWrapper) SetHeight(v int) { c.Line.Props.Height = v; c.updateFormat() }

// Для ComboBox используем string
func (c *ClicheItemWrapper) Font() string { return strconv.Itoa(c.Line.Props.Font) }
func (c *ClicheItemWrapper) SetFont(v string) {
	c.Line.Props.Font, _ = strconv.Atoi(v)
	c.updateFormat()
}

func (c *ClicheItemWrapper) Underline() string { return strconv.Itoa(c.Line.Props.Underline) }
func (c *ClicheItemWrapper) SetUnderline(v string) {
	c.Line.Props.Underline, _ = strconv.Atoi(v)
	c.updateFormat()
}

func (c *ClicheItemWrapper) Align() string { return strconv.Itoa(c.Line.Props.Align) }
func (c *ClicheItemWrapper) SetAlign(v string) {
	c.Line.Props.Align, _ = strconv.Atoi(v)
	c.updateFormat()
}

func (c *ClicheItemWrapper) updateFormat() {
	c.Line.Format = cliche.BuildFormat(c.Line.Props)
}

// ClicheModel - модель для TableView.
type ClicheModel struct {
	walk.TableModelBase
	Items []*ClicheItemWrapper
}

func (m *ClicheModel) RowCount() int {
	return len(m.Items)
}

func (m *ClicheModel) Value(row, col int) interface{} {
	if row >= len(m.Items) {
		return ""
	}
	item := m.Items[row]
	switch col {
	case 0:
		return item.Index + 1 // Номер строки 1..10
	case 1:
		return item.Line.Format
	case 2:
		return item.Line.Text
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

	// Опция b9 (разделена на СНО и Флаг)
	OptB9_BaseTax string // "0", "1"... (Value member)
	OptB9_FullX   bool

	// Доступные СНО для b9 (храним в модели, но обновляем виджет вручную)
	OptB9_SNO []*NV

	// Денежный ящик (остаются int, т.к. используются в NumberEdit)
	DrawerPin  int
	DrawerRise int
	DrawerFall int

	// --- Клише ---
	SelectedClicheType string               // "1".."4"
	ClicheItems        []*ClicheItemWrapper // 10 строк
	CurrentClicheLine  *ClicheItemWrapper   // Указатель на редактируемую строку
	// TempClicheLine - временный объект для редактирования.
	// Данные из него попадают в ClicheItems только по кнопке "Применить".
	TempClicheLine *ClicheItemWrapper
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
	OptQRPos, OptTextQR, OptCount, OptRounding *walk.Label
	// B9
	OptB9_BaseTax, OptB9_FullX *walk.Label
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

	// Специфичный контрол для b9, чтобы менять его Model вручную
	b9ComboBox *walk.ComboBox

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

	ceText                   *walk.LineEdit
	ceAlign, ceFont, ceUnder *walk.ComboBox
	ceWidth, ceHeight        *walk.NumberEdit
	ceInvert                 *walk.CheckBox
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

	// В опции B9 содержится код базовой СНО + флаг полного Х-отчет (+16).
	// Код СНО в b9 = 1..8.
	b9Val := 0
	if vm.OptB9_BaseTax != "" {
		if v, err := strconv.Atoi(vm.OptB9_BaseTax); err == nil {
			b9Val += v
		}
	}
	if vm.OptB9_FullX {
		b9Val += 16
	}
	opts.B9 = b9Val
	s.Options = opts

	// 8. Cliches
	// Копируем остальные типы из текущего снапшота, чтобы не потерять их
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
		// Обновляем формат в строке перед сохранением
		item.updateFormat()
		lines = append(lines, driver.ClicheLineData{
			Text:   item.Line.Text,
			Format: item.Line.Format,
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
		"Opt_B9":           {&sLabels.OptB9_BaseTax, &sLabels.OptB9_FullX},
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
		// Читаем всё из ККТ (время, настройки, клише, регистрацию)
		t, _ := drv.GetDateTime()
		ofd, _ := drv.GetOfdSettings()
		oism, _ := drv.GetOismSettings()
		lan, _ := drv.GetLanSettings()
		prn, _ := drv.GetPrinterSettings()
		cd, _ := drv.GetMoneyDrawerSettings()
		opts, _ := drv.GetOptions()
		tz, _ := drv.GetTimezone()
		regData, _ := drv.GetRegistrationData()

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

			// --- Формирование списка доступных СНО для b9 ---
			var taxList []*NV
			// Добавляем пустой вариант
			taxList = append(taxList, &NV{Name: "Не выбрано", Code: "0"})

			if regData != nil && regData.TaxSystems != "" {
				parts := strings.Split(regData.TaxSystems, ",")
				taxNameMap := map[string]string{
					"0": "ОСН", "1": "УСН доход", "2": "УСН д-р",
					"3": "ЕНВД", "4": "ЕСХН", "5": "Патент",
				}
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}

					name, ok := taxNameMap[p]
					if !ok {
						name = "СНО #" + p
					}

					// Конвертация: Code(1062) -> Code(b9) = Code(1062) + 1
					if codeInt, err := strconv.Atoi(p); err == nil {
						b9Code := strconv.Itoa(codeInt + 1)
						taxList = append(taxList, &NV{Name: name, Code: b9Code})
					}
				}
			}
			serviceModel.OptB9_SNO = taxList

			// Обновляем виджет ComboBox вручную
			if b9ComboBox != nil {
				b9ComboBox.SetModel(taxList)
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

				// Парсинг b9
				serviceModel.OptB9_FullX = (opts.B9 & 16) != 0
				taxVal := opts.B9 & 0x0F
				serviceModel.OptB9_BaseTax = strconv.Itoa(taxVal)
			}

			// 3. Обновляем клише (ОБНОВЛЕНО)
			curType, _ := strconv.Atoi(serviceModel.SelectedClicheType)
			lines := allCliches[curType]

			for i := 0; i < 10; i++ {
				var text, format string
				if i < len(lines) {
					text = lines[i].Text
					format = lines[i].Format
				} else {
					text = ""
					format = "000000"
				}

				// Используем пакет cliche для парсинга
				props := cliche.ParseFormat(format)

				serviceModel.ClicheItems[i].Line = cliche.Line{
					Text:   text,
					Format: format,
					Props:  props,
				}
			}

			// 4. Сначала обновляем СНАПШОТЫ, чтобы StyleCell имел доступ к данным
			tempSnap := viewModelToSnapshot(serviceModel)
			tempSnap.Cliches = allCliches

			initialSnapshot = tempSnap
			currentSnapshot = tempSnap
			currentChanges = nil

			// 5. Теперь безопасно обновлять таблицу и принудительно перерисовывать
			clicheModel.PublishRowsReset()
			// ИСПРАВЛЕНО: Принудительная перерисовка таблицы
			if clicheTable != nil {
				clicheTable.Invalidate()
			}

			// 6. Синхронизируем UI с обновленной моделью
			serviceBinder.Reset()

			// 7. Снимаем блокировку и обновляем состояние кнопки
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
		OptB9_FullX:        false,
		OptB9_BaseTax:      "0",
		OptB9_SNO:          []*NV{{Name: "Не выбрано", Code: "0"}},
		OfdClient:          "1",
		SelectedClicheType: "1",
		CurrentClicheLine:  &ClicheItemWrapper{},
		TempClicheLine:     &ClicheItemWrapper{Line: cliche.Line{Format: "000000", Props: cliche.DefaultProps()}},
	}

	serviceModel.ClicheItems = make([]*ClicheItemWrapper, 10)
	for i := 0; i < 10; i++ {
		// Инициализация строк с дефолтными пропсами
		wrapper := &ClicheItemWrapper{
			Index: i,
			Line: cliche.Line{
				Format: "000000",
				Props:  cliche.DefaultProps(),
			},
		}
		serviceModel.ClicheItems[i] = wrapper
	}
	clicheModel = &ClicheModel{Items: serviceModel.ClicheItems}

	loadServiceInitial()
	startClocks()

	return d.TabPage{
		Title:  "Сервис",
		Layout: d.VBox{MarginsZero: true, Spacing: 5},
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
								Layout: d.HBox{Margins: d.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}, Spacing: 4, Alignment: d.AlignHCenterVCenter},
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

													// ИЗМЕНЕНО: Новые контролы для b9
													d.Label{AssignTo: &sLabels.OptB9_FullX, Text: "X-Отчет:"},
													d.CheckBox{Text: "Полный", Checked: d.Bind("OptB9_FullX"), ToolTipText: "Печатать полный X-отчет (b9 & 16)", OnCheckStateChanged: recalcChanges},

													d.Label{AssignTo: &sLabels.OptB9_BaseTax, Text: "Баз. СНО:"},
													d.ComboBox{
														AssignTo:              &b9ComboBox, // Прямая ссылка для ручного обновления модели
														Value:                 d.Bind("OptB9_BaseTax"),
														BindingMember:         "Code",
														DisplayMember:         "Name",
														Model:                 serviceModel.OptB9_SNO, // Статическая инициализация
														MinSize:               d.Size{Width: 100},
														ToolTipText:           "Система налогообложения по умолчанию",
														OnCurrentIndexChanged: recalcChanges,
													},
												},
											},
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
							d.Composite{
								Layout: d.HBox{MarginsZero: true, Alignment: d.AlignHCenterVCenter},
								Children: []d.Widget{
									d.Label{AssignTo: &sLabels.ClicheHeader, Text: "Редактировать:"},
									d.ComboBox{
										Value:         d.Bind("SelectedClicheType"),
										Model:         listClicheTypes,
										BindingMember: "Code", DisplayMember: "Name",
										MinSize:               d.Size{Width: 100},
										OnCurrentIndexChanged: onClicheTypeSwitched,
									},
								},
							},
							d.Composite{
								Layout: d.HBox{MarginsZero: true, Spacing: 5},
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
										StyleCell: func(style *walk.CellStyle) {
											// Задаем черный цвет по умолчанию
											style.TextColor = walk.RGB(0, 0, 0)

											if initialSnapshot == nil {
												return
											}
											row := style.Row()
											if row < 0 || row >= len(serviceModel.ClicheItems) {
												return
											}

											curType, _ := strconv.Atoi(serviceModel.SelectedClicheType)
											if initialSnapshot.Cliches == nil {
												return
											}
											initialLines := initialSnapshot.Cliches[curType]

											var initialFormat, initialText string
											if row < len(initialLines) {
												initialFormat = initialLines[row].Format
												initialText = initialLines[row].Text
											} else {
												initialFormat = "000000"
												initialText = ""
											}

											currentItem := serviceModel.ClicheItems[row]

											// Если есть отличия - выделяем жирным
											if currentItem.Line.Text != initialText || currentItem.Line.Format != initialFormat {
												// БЕЗОПАСНОЕ ПОЛУЧЕНИЕ ШРИФТА
												var family string
												var size int

												if style.Font != nil {
													family = style.Font.Family()
													size = style.Font.PointSize()
												} else if clicheTable != nil && clicheTable.Font() != nil {
													// Берем шрифт самой таблицы
													family = clicheTable.Font().Family()
													size = clicheTable.Font().PointSize()
												} else {
													// Дефолтные значения (если совсем всё плохо)
													family = "Segoe UI"
													size = 9
												}

												// Создаем жирный шрифт
												if f, err := walk.NewFont(family, size, walk.FontBold); err == nil {
													style.Font = f
												}
											}
										},
									},
									d.GroupBox{
										AssignTo: &clicheEditorGroup,
										Title:    "Настройки строки",
										Layout:   d.VBox{Margins: d.Margins{Left: 10, Top: 10, Right: 10, Bottom: 10}, Spacing: 8},
										Enabled:  false,
										MaxSize:  d.Size{Width: 300, Height: 250},
										DataBinder: d.DataBinder{
											AssignTo:   &clicheEditorBinder,
											DataSource: serviceModel.TempClicheLine,
											// AutoSubmit не нужен, мы читаем вручную
										},
										Children: []d.Widget{
											d.Label{Text: "Текст:"},
											d.LineEdit{
												AssignTo: &ceText,
												Text:     d.Bind("Text"),
											},
											d.Composite{
												Layout: d.Grid{Columns: 2, Spacing: 10},
												Children: []d.Widget{
													d.Label{Text: "Выравнивание:"},
													d.ComboBox{
														AssignTo:      &ceAlign,
														Value:         d.Bind("Align"),
														Model:         listAlign,
														BindingMember: "Code", DisplayMember: "Name",
													},
													d.Label{Text: "Шрифт:"},
													d.ComboBox{
														AssignTo:      &ceFont,
														Value:         d.Bind("Font"),
														Model:         listFonts,
														BindingMember: "Code", DisplayMember: "Name",
													},
													d.Label{Text: "Подчеркивание:"},
													d.ComboBox{
														AssignTo:      &ceUnder,
														Value:         d.Bind("Underline"),
														Model:         listUnderline,
														BindingMember: "Code", DisplayMember: "Name",
													},
												},
											},
											d.GroupBox{
												Title:  "Масштабирование",
												Layout: d.Grid{Columns: 4},
												Children: []d.Widget{
													d.Label{Text: "Ширина:"},
													d.NumberEdit{
														AssignTo: &ceWidth,
														Value:    d.Bind("Width"),
														MinValue: 0, MaxValue: 8, MaxSize: d.Size{Width: 40},
													},
													d.Label{Text: "Высота:"},
													d.NumberEdit{
														AssignTo: &ceHeight,
														Value:    d.Bind("Height"),
														MinValue: 0, MaxValue: 8, MaxSize: d.Size{Width: 40},
													},
												},
											},
											d.CheckBox{
												AssignTo: &ceInvert, // <--- ДОБАВЛЕНО
												Text:     "Инверсия (Белым по черному)",
												Checked:  d.Bind("Invert"),
											},
											d.VSpacer{Size: 5},
											d.PushButton{
												Text:      "Применить изменения строки",
												OnClicked: onApplyClicheLine,
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

// restoreViewFromSnapshot восстанавливает значения в ViewModel из предоставленного снапшота.
func restoreViewFromSnapshot(vm *ServiceViewModel, snap *service.SettingsSnapshot) {
	if snap == nil {
		return
	}

	// Блокируем триггеры recalcChanges
	isLoadingData = true
	defer func() {
		isLoadingData = false
		recalcChanges() // Пересчитываем, чтобы сбросить флаги изменений и кнопки
	}()

	// 1. ОФД
	vm.OfdString = joinHostPort(snap.Ofd.Addr, snap.Ofd.Port)
	vm.OfdClient = snap.Ofd.Client
	vm.TimerFN = snap.Ofd.TimerFN
	vm.TimerOFD = snap.Ofd.TimerOFD

	// 2. ОИСМ
	vm.OismString = joinHostPort(snap.Oism.Addr, snap.Oism.Port)

	// 3. LAN
	vm.LanAddr = snap.Lan.Addr
	vm.LanPort = snap.Lan.Port
	vm.LanMask = snap.Lan.Mask
	vm.LanDns = snap.Lan.Dns
	vm.LanGw = snap.Lan.Gw

	// 4. Timezone
	vm.OptTimezone = strconv.Itoa(snap.Timezone)

	// 5. Принтер
	vm.PrintModel = snap.Printer.Model
	vm.PrintBaud = strconv.Itoa(snap.Printer.BaudRate)
	vm.PrintPaper = strconv.Itoa(snap.Printer.Paper)
	vm.PrintFont = strconv.Itoa(snap.Printer.Font)

	// 6. Денежный ящик
	vm.DrawerPin = snap.Drawer.Pin
	vm.DrawerRise = snap.Drawer.Rise
	vm.DrawerFall = snap.Drawer.Fall

	// 7. Опции
	vm.OptQRPos = strconv.Itoa(snap.Options.B1)
	vm.OptRounding = strconv.Itoa(snap.Options.B2)
	vm.OptCut = (snap.Options.B3 == 1)
	vm.OptAutoTest = (snap.Options.B4 == 1)
	vm.OptDrawerTrig = strconv.Itoa(snap.Options.B5)
	vm.OptNearEnd = (snap.Options.B6 == 1)
	vm.OptTextQR = (snap.Options.B7 == 1)
	vm.OptCountInCheck = (snap.Options.B8 == 1)

	// Восстановление b9
	vm.OptB9_FullX = (snap.Options.B9 & 16) != 0
	taxVal := snap.Options.B9 & 0x0F
	vm.OptB9_BaseTax = strconv.Itoa(taxVal)

	// 8. Клише
	curType, _ := strconv.Atoi(vm.SelectedClicheType)
	lines := snap.Cliches[curType]

	for i := 0; i < 10; i++ {
		var text, format string
		if i < len(lines) {
			text = lines[i].Text
			format = lines[i].Format
		} else {
			text = ""
			format = "000000"
		}

		// Заполняем структуру Line внутри Wrapper, используя новый парсер
		vm.ClicheItems[i].Line = cliche.Line{
			Text:   text,
			Format: format,
			Props:  cliche.ParseFormat(format),
		}
	}

	// Обновляем визуальное состояние
	if serviceBinder != nil {
		serviceBinder.Reset()
	}
	if clicheModel != nil {
		clicheModel.PublishRowsReset()
	}
}

func checkOfdClientChange() {
	if isLoadingData {
		return
	}
	recalcChanges()

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
	recalcChanges()

	newType, _ := strconv.Atoi(serviceModel.SelectedClicheType)
	lines := currentSnapshot.Cliches[newType]

	for i := 0; i < 10; i++ {
		var text, format string
		if i < len(lines) {
			text = lines[i].Text
			format = lines[i].Format
		} else {
			text = ""
			format = "000000"
		}

		// Заполняем структуру Line внутри Wrapper, используя парсер
		serviceModel.ClicheItems[i].Line = cliche.Line{
			Text:   text,
			Format: format,
			Props:  cliche.ParseFormat(format),
		}
	}

	clicheModel.PublishRowsReset()
	// Если была выбрана строка для редактирования, перезагружаем редактор
	if idx := clicheTable.CurrentIndex(); idx >= 0 {
		reloadEditor(idx)
	}
}

func onWriteAllParameters() {
	if len(currentChanges) == 0 {
		return
	}
	confirmed, finalChanges := RunDiffDialog(mw, currentChanges)
	if !confirmed {
		return
	}
	ApplyChangesPipeline(finalChanges)
}

func onTechReset() {
	drv := driver.Active
	if drv == nil {
		return
	}
	if walk.MsgBox(mw, "ВНИМАНИЕ",
		"Выполнить ТЕХНОЛОГИЧЕСКОЕ ОБНУЛЕНИЕ?\nЭто полностью очистит настройки ККТ.",
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
				onReadAllSettings()
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

// reloadEditor загружает данные из выбранной строки списка во временный объект редактора.
func reloadEditor(idx int) {
	srcItem := serviceModel.ClicheItems[idx]

	// Копируем данные в TempClicheLine
	// Важно скопировать значения, а не ссылку, чтобы редактирование не меняло список сразу
	serviceModel.TempClicheLine.Index = srcItem.Index
	serviceModel.TempClicheLine.Line = srcItem.Line // cliche.Line - это struct (value type), копируется по значению

	if clicheEditorBinder != nil {
		clicheEditorBinder.SetDataSource(serviceModel.TempClicheLine)
		clicheEditorBinder.Reset()
	}

	// Обновляем DataBinder редактора
	clicheEditorBinder.Reset()

	clicheEditorGroup.SetEnabled(true)
	clicheEditorGroup.SetTitle(fmt.Sprintf("Настройки строки №%d", idx+1))
}

// onApplyClicheLine вызывается при нажатии кнопки "Применить" в редакторе клише.
func onApplyClicheLine() {
	idx := clicheTable.CurrentIndex()
	if idx < 0 {
		return
	}

	// 1. Ручное чтение значений из виджетов
	newText := ceText.Text()

	// Хелпер для чтения значения из ComboBox через Model и CurrentIndex
	getComboVal := func(cb *walk.ComboBox) int {
		idx := cb.CurrentIndex()
		if idx < 0 {
			return 0
		}

		// Получаем модель виджета и приводим к типу []*NV, который мы использовали при инициализации
		if items, ok := cb.Model().([]*NV); ok {
			if idx < len(items) {
				if i, err := strconv.Atoi(items[idx].Code); err == nil {
					return i
				}
			}
		}
		return 0
	}

	newAlign := getComboVal(ceAlign)
	newFont := getComboVal(ceFont)
	newUnder := getComboVal(ceUnder)

	newWidth := int(ceWidth.Value())
	newHeight := int(ceHeight.Value())
	newInvert := ceInvert.Checked()

	// 2. Обновляем структуру TempClicheLine вручную
	serviceModel.TempClicheLine.Line.Text = newText
	serviceModel.TempClicheLine.Line.Props.Align = newAlign
	serviceModel.TempClicheLine.Line.Props.Font = newFont
	serviceModel.TempClicheLine.Line.Props.Underline = newUnder
	serviceModel.TempClicheLine.Line.Props.Width = newWidth
	serviceModel.TempClicheLine.Line.Props.Height = newHeight
	serviceModel.TempClicheLine.Line.Props.Invert = newInvert

	// Пересчитываем формат (строку "xxxxxx")
	serviceModel.TempClicheLine.updateFormat()

	// 3. Сравниваем с текущим значением в списке
	originalItem := serviceModel.ClicheItems[idx]

	hasChanges := false
	if serviceModel.TempClicheLine.Line.Text != originalItem.Line.Text {
		hasChanges = true
	}
	if serviceModel.TempClicheLine.Line.Format != originalItem.Line.Format {
		hasChanges = true
	}

	if !hasChanges {
		logMsg("Нет изменений в строке %d", idx+1)
		return
	}

	// 4. Применяем изменения в основной список
	originalItem.Line = serviceModel.TempClicheLine.Line

	// 5. Обновляем таблицу и запускаем пересчет изменений
	clicheModel.PublishRowChanged(idx)
	if clicheTable != nil {
		clicheTable.Invalidate()
	}
	recalcChanges()

	logMsg("Строка %d обновлена в памяти: %s (Format: %s)", idx+1, newText, serviceModel.TempClicheLine.Line.Format)
}

func onClicheSelectionChanged() {
	idx := clicheTable.CurrentIndex()
	if idx < 0 {
		clicheEditorGroup.SetEnabled(false)
		return
	}
	reloadEditor(idx)
}
