package viewmodel

import (
	"mitsuscanner/internal/cliche"
	"strconv"
	"time"
)

// NameValue используется для заполнения выпадающих списков (ComboBox) в UI.
type NameValue struct {
	Name string
	Code string
}

// -----------------------------
// СПИСКИ ЗНАЧЕНИЙ ДЛЯ ВЫПАДАЮЩИХ МЕНЮ
// -----------------------------

var (
	// Список клиентов ОФД
	ListOfdClients = []*NameValue{
		{"Внешний", "1"},
		{"Встроенный (LAN)", "0"},
	}

	// Список скоростей передачи данных
	ListBaudRates = []*NameValue{
		{"9600", "9600"}, {"19200", "19200"}, {"38400", "38400"}, {"57600", "57600"}, {"115200", "115200"},
	}

	// Список моделей принтеров
	ListPrinterModels = []*NameValue{
		{"RP-809", "1"}, {"F80", "2"},
	}

	// Список ширины бумаги (значения в мм)
	ListPaperWidths = []*NameValue{
		{"80мм", "80"}, {"57мм", "57"},
	}

	// Список позиций QR-кода на чеке
	ListQRPositions = []*NameValue{
		{"Слева", "0"}, {"По центру", "1"}, {"Справа", "2"},
	}

	// Список вариантов округления суммы
	ListRoundingOptions = []*NameValue{
		{"Нет", "0"}, {"0.10", "1"}, {"0.50", "2"}, {"1.00", "3"},
	}

	// Список условий автоматического открытия денежного ящика
	ListDrawerTriggers = []*NameValue{
		{"Нет", "0"}, {"Наличные", "1"}, {"Безнал", "2"}, {"Всегда", "3"},
	}

	// Список часовых поясов
	ListTimezones = []*NameValue{
		{"UTC+2 (Клд)", "2"}, {"UTC+3 (Мск)", "3"}, {"UTC+4 (Смр)", "4"},
		{"UTC+5 (Екб)", "5"}, {"UTC+6 (Омск)", "6"}, {"UTC+7 (Крсн)", "7"},
		{"UTC+8 (Ирк)", "8"}, {"UTC+9 (Якт)", "9"}, {"UTC+10 (Влд)", "10"},
		{"UTC+11 (Маг)", "11"}, {"UTC+12 (Кам)", "12"}, {"Не настроено", "254"},
	}

	// Список типов клише
	ListClicheTypes = []*NameValue{
		{"1 - Заголовок (Клише)", "1"},
		{"2 - После пользователя", "2"},
		{"3 - Подвал (Реклама)", "3"},
		{"4 - Конец чека", "4"},
	}

	// Список вариантов выравнивания текста
	ListAlignments = []*NameValue{
		{"Слева", "0"}, {"Центр", "1"}, {"Справа", "2"},
	}

	// Список шрифтов
	ListFonts = []*NameValue{
		{"A", "0"}, {"B", "1"},
	}

	// Список вариантов подчеркивания
	ListUnderlineOptions = []*NameValue{
		{"Нет", "0"}, {"Текст", "1"}, {"Вся строка", "2"},
	}
)

// ClicheItemWrapper обертка над internal/cliche.Line для UI.
// Используется для биндинга данных в TableView.
type ClicheItemWrapper struct {
	Index int
	Line  cliche.Line
}

// Геттеры/Сеттеры для DataBinding

func (c *ClicheItemWrapper) Text() string     { return c.Line.Text }
func (c *ClicheItemWrapper) SetText(v string) { c.Line.Text = v }

func (c *ClicheItemWrapper) Format() string { return c.Line.Format }

func (c *ClicheItemWrapper) Invert() bool     { return c.Line.Props.Invert }
func (c *ClicheItemWrapper) SetInvert(v bool) { c.Line.Props.Invert = v; c.updateFormat() }

func (c *ClicheItemWrapper) Width() int     { return c.Line.Props.Width }
func (c *ClicheItemWrapper) SetWidth(v int) { c.Line.Props.Width = v; c.updateFormat() }

func (c *ClicheItemWrapper) Height() int     { return c.Line.Props.Height }
func (c *ClicheItemWrapper) SetHeight(v int) { c.Line.Props.Height = v; c.updateFormat() }

func (c *ClicheItemWrapper) Font() string { return strconv.Itoa(c.Line.Props.Font) }
func (c *ClicheItemWrapper) SetFont(v string) {
	if f, err := strconv.Atoi(v); err == nil {
		c.Line.Props.Font = f
		c.updateFormat()
	}
}

func (c *ClicheItemWrapper) Underline() string { return strconv.Itoa(c.Line.Props.Underline) }
func (c *ClicheItemWrapper) SetUnderline(v string) {
	if u, err := strconv.Atoi(v); err == nil {
		c.Line.Props.Underline = u
		c.updateFormat()
	}
}

func (c *ClicheItemWrapper) Align() string { return strconv.Itoa(c.Line.Props.Align) }
func (c *ClicheItemWrapper) SetAlign(v string) {
	if a, err := strconv.Atoi(v); err == nil {
		c.Line.Props.Align = a
		c.updateFormat()
	}
}

func (c *ClicheItemWrapper) updateFormat() {
	c.Line.Format = cliche.BuildFormat(c.Line.Props)
}

// ServiceViewModel отвечает за отображение всех настроек на вкладке "Сервис".
type ServiceViewModel struct {
	// --- Время ---
	KktTimeStr    string // Время ККТ (для отображения)
	TargetTimeStr string // Желаемое время для установки
	AutoSyncPC    bool   // Автоматическая синхронизация с временем ПК

	// --- Связь и ОФД ---
	OfdString string // Адрес:Порт ОФД
	OfdClient string // Тип клиента ОФД ("0" или "1")
	TimerFN   int    // Таймер ФН (секунды)
	TimerOFD  int    // Таймер ОФД (секунды)

	OismString string // Адрес:Порт ОИСМ

	// LAN
	LanAddr string // IP-адрес
	LanPort int    // Порт
	LanMask string // Маска подсети
	LanDns  string // DNS-сервер
	LanGw   string // Шлюз

	// --- Параметры (Оборудование и Опции) ---
	// Для ComboBox используем string, чтобы обеспечить корректный биндинг
	PrintModel string // Модель принтера ("1", "2")
	PrintBaud  string // Скорость передачи ("115200")
	PrintPaper string // Ширина бумаги ("80", "57")
	PrintFont  string // Шрифт ("0", "1")

	// Опции устройства
	OptTimezone     string // Часовой пояс
	OptCut          bool   // Флаг отрезчика
	OptAutoTest     bool   // Автоматический тест при запуске
	OptNearEnd      bool   // Датчик окончания бумаги
	OptTextQR       bool   // Текстовое представление QR
	OptCountInCheck bool   // Подсчет покупок в чеке
	OptQRPos        string // Позиция QR-кода
	OptRounding     string // Округление
	OptDrawerTrig   string // Условие открытия денежного ящика

	// Опция b9 (разделена на СНО и Флаг полного Х-отчета)
	OptB9_BaseTax string       // Базовая СНО ("0", "1"... )
	OptB9_FullX   bool         // Флаг полного X-отчета
	OptB9_SNO     []*NameValue // Доступные системы налогообложения

	// Денежный ящик
	DrawerPin  int // PIN-код
	DrawerRise int // Время нарастания импульса (мс)
	DrawerFall int // Время спада импульса (мс)

	// --- Клише ---
	SelectedClicheType string               // Выбранный тип клише ("1".."4")
	LastSelectedType   string               // Предыдущий выбранный тип
	ClicheItems        []*ClicheItemWrapper // 10 строк клише
	CurrentClicheLine  *ClicheItemWrapper   // Редактируемая строка
	TempClicheLine     *ClicheItemWrapper   // Временный объект для редактирования
}

// NewServiceViewModel создаёт новый экземпляр ServiceViewModel с дефолтными значениями.
func NewServiceViewModel() *ServiceViewModel {
	vm := &ServiceViewModel{
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
		OptB9_SNO:          []*NameValue{{Name: "Не выбрано", Code: "0"}},
		OfdClient:          "1",
		SelectedClicheType: "1",
		LastSelectedType:   "1",
		CurrentClicheLine:  &ClicheItemWrapper{},
		TempClicheLine:     &ClicheItemWrapper{Line: cliche.Line{Format: "000000", Props: cliche.DefaultProps()}},
		KktTimeStr:         "Нет подключения",
		TargetTimeStr:      formatTime(time.Now()),
		AutoSyncPC:         true,
	}

	// Инициализация строк клише
	vm.ClicheItems = make([]*ClicheItemWrapper, 10)
	for i := 0; i < 10; i++ {
		vm.ClicheItems[i] = &ClicheItemWrapper{
			Index: i,
			Line: cliche.Line{
				Format: "000000",
				Props:  cliche.DefaultProps(),
			},
		}
	}

	return vm
}

// formatTime форматирует время для отображения в UI (ДД.ММ.ГГГГ ЧЧ:ММ:СС).
func formatTime(t time.Time) string {
	return t.Format("02.01.2006 15:04:05")
}
