package models

import (
	"fmt"
	"strconv"
)

// ClicheItem представляет одну строку клише для GUI
// Используется в составе Snapshot для хранения состояния клише
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

// ParseFormatString разбирает строку формата "xxxxxx" в поля структуры
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

// UpdateFormatString собирает строку формата "xxxxxx" из полей структуры
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

// Snapshot представляет снимок настроек для сравнения изменений
// Используется для хранения состояния настроек ККТ в определенный момент времени
type Snapshot struct {
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
	ClicheItems []*ClicheItem // 10 строк
}

// CreateSnapshot создает глубокую копию текущих настроек для последующего сравнения
// Принимает текущую модель ServiceViewModel и возвращает Snapshot
func CreateSnapshotFromModel(
	ofdString, ofdClient string,
	timerFN, timerOFD int,
	oismString string,
	lanAddr string, lanPort int, lanMask, lanDns, lanGw string,
	printModel, printBaud string, printPaper, printFont int,
	optTimezone string, optCut, optAutoTest, optNearEnd, optTextQR, optCountInCheck bool,
	optQRPos, optRounding, optDrawerTrig, optB9 string,
	drawerPin, drawerRise, drawerFall int,
	clicheItems []*ClicheItem,
) *Snapshot {
	snap := &Snapshot{
		OfdString: ofdString,
		OfdClient: ofdClient,
		TimerFN:   timerFN,
		TimerOFD:  timerOFD,

		OismString: oismString,

		LanAddr: lanAddr,
		LanPort: lanPort,
		LanMask: lanMask,
		LanDns:  lanDns,
		LanGw:   lanGw,

		PrintModel: printModel,
		PrintBaud:  printBaud,
		PrintPaper: printPaper,
		PrintFont:  printFont,

		OptTimezone:     optTimezone,
		OptCut:          optCut,
		OptAutoTest:     optAutoTest,
		OptNearEnd:      optNearEnd,
		OptTextQR:       optTextQR,
		OptCountInCheck: optCountInCheck,
		OptQRPos:        optQRPos,
		OptRounding:     optRounding,
		OptDrawerTrig:   optDrawerTrig,
		OptB9:           optB9,

		DrawerPin:  drawerPin,
		DrawerRise: drawerRise,
		DrawerFall: drawerFall,
	}

	// Глубокая копия клише
	snap.ClicheItems = make([]*ClicheItem, len(clicheItems))
	for i, item := range clicheItems {
		snap.ClicheItems[i] = &ClicheItem{
			Index:     item.Index,
			Text:      item.Text,
			Format:    item.Format,
			Invert:    item.Invert,
			Width:     item.Width,
			Height:    item.Height,
			Font:      item.Font,
			Underline: item.Underline,
			Align:     item.Align,
		}
	}

	return snap
}
