package mitsudriver

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// escapeXMLText экранирует только спецсимволы XML, сохраняя кавычки.
// Это исправляет проблему, когда кавычки в названии ОФД превращались в &#34;
func escapeXMLText(s string) string {
	r := strings.NewReplacer(
		"&", "&",
		"<", "<",
		">", ">",
	)
	return r.Replace(s)
}

func decodeXML(data []byte, v interface{}) error {
	utf8Data, err := toUTF8(data)
	if err != nil {
		return fmt.Errorf("ошибка конвертации кодировки: %w", err)
	}
	return xml.Unmarshal(utf8Data, v)
}

func toUTF8(data []byte) ([]byte, error) {
	r, err := charset.NewReaderLabel("windows-1251", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return io.ReadAll(r)
}

func encodeCP1251(s string) ([]byte, error) {
	encoder := charmap.Windows1251.NewEncoder()
	res, _, err := transform.Bytes(encoder, []byte(s))
	if err != nil {
		return nil, fmt.Errorf("ошибка кодирования в WIN-1251: %w", err)
	}
	return res, nil
}

func parseError(data []byte) error {
	utf8Data, err := toUTF8(data)
	if err != nil {
		utf8Data = data
	}

	type ErrorResp struct {
		No  string `xml:"No,attr"`
		FSE string `xml:"FSE,attr"`
		TAG string `xml:"TAG,attr"`
		PAR string `xml:"PAR,attr"`
	}
	var e ErrorResp

	if err := xml.Unmarshal(utf8Data, &e); err != nil {
		return fmt.Errorf("ошибка ККТ (нераспознанная): %s", string(data))
	}

	desc, exists := ErrorDescriptions[e.No]
	if !exists {
		desc = "неизвестная ошибка"
	}

	msg := fmt.Sprintf("Ошибка ККТ #%s: %s", e.No, desc)

	if e.PAR != "" {
		msg += fmt.Sprintf(" (параметр: %s)", e.PAR)
	}
	if e.FSE != "" {
		fnDesc, fnExists := ErrorDescriptions[e.FSE]
		if fnExists {
			msg += fmt.Sprintf(", ошибка ФН #%s: %s", e.FSE, fnDesc)
		} else {
			msg += fmt.Sprintf(", ошибка ФН: %s", e.FSE)
		}
	}
	if e.TAG != "" {
		msg += fmt.Sprintf(" [TAG: %s]", e.TAG)
	}

	return errors.New(msg)
}
