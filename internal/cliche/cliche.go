package cliche

import (
	"fmt"
	"strconv"
)

// Props описывает свойства форматирования строки клише.
// Формат F='xxxxxx' (6 цифр).
type Props struct {
	Invert    bool // 1 цифра: 0-нет, 1-да
	Width     int  // 2 цифра: 0-8
	Height    int  // 3 цифра: 0-8
	Font      int  // 4 цифра: 0-настройка, 1-A, 2-B
	Underline int  // 5 цифра: 0-нет, 1-текст, 2-вся строка
	Align     int  // 6 цифра: 0-лево, 1-центр, 2-право
}

// Line представляет одну строку клише с текстом и настройками.
type Line struct {
	Text   string
	Format string // Сырая строка формата "xxxxxx"
	Props  Props  // Распаршенные свойства
}

// ParseFormat разбирает строку формата "xxxxxx" в структуру Props.
func ParseFormat(fmtStr string) Props {
	p := Props{}
	runes := []rune(fmtStr)

	// Добиваем нулями если строка короче 6 символов или пустая
	for len(runes) < 6 {
		runes = append(runes, '0')
	}

	p.Invert = (runes[0] == '1')
	p.Width = parseInt(runes[1])
	p.Height = parseInt(runes[2])
	p.Font = parseInt(runes[3])
	p.Underline = parseInt(runes[4])
	p.Align = parseInt(runes[5])

	return p
}

// BuildFormat собирает строку формата "xxxxxx" из структуры Props.
func BuildFormat(p Props) string {
	inv := 0
	if p.Invert {
		inv = 1
	}

	// Валидация диапазонов согласно документации
	w := clamp(p.Width, 0, 8)
	h := clamp(p.Height, 0, 8)
	f := clamp(p.Font, 0, 2)
	u := clamp(p.Underline, 0, 2)
	a := clamp(p.Align, 0, 2)

	return fmt.Sprintf("%d%d%d%d%d%d", inv, w, h, f, u, a)
}

// DefaultProps возвращает настройки по умолчанию ("000000").
func DefaultProps() Props {
	return Props{}
}

// --- Helpers ---

func parseInt(r rune) int {
	val, err := strconv.Atoi(string(r))
	if err != nil {
		return 0
	}
	return val
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
