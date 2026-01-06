package driver

import (
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"strconv"
	"testing"
)

func TestDecodeMode(t *testing.T) {
	if len(DecodeMode(0)) != 0 {
		t.Error("expected empty")
	}
	res := DecodeMode(1)
	if len(res) != 1 || res[0] != "MODE бит 0" {
		t.Error("expected MODE бит 0")
	}
	res = DecodeMode(3)
	if len(res) != 2 || res[0] != "MODE бит 0" || res[1] != "MODE бит 1" {
		t.Error("expected two bits")
	}
}

func TestDecodeExtMode(t *testing.T) {
	if len(DecodeExtMode(0)) != 0 {
		t.Error("expected empty")
	}
	res := DecodeExtMode(1)
	if len(res) != 1 || res[0] != "ExtMODE бит 0" {
		t.Error("expected ExtMODE бит 0")
	}
	res = DecodeExtMode(3)
	if len(res) != 2 || res[0] != "ExtMODE бит 0" || res[1] != "ExtMODE бит 1" {
		t.Error("expected two bits")
	}
}

func TestExtractDocDateTime(t *testing.T) {
	tests := []struct {
		name     string
		xmlStr   string
		expected string
		hasError bool
	}{
		{
			name:     "layout 02-01-06T15:04",
			xmlStr:   `<DocXML FORM="2"><T1012>01-05-23T01:35</T1012></DocXML>`,
			expected: "01.05.2023 01:35",
			hasError: false,
		},
		{
			name:     "layout 02-01-06T15:04:05",
			xmlStr:   `<DocXML FORM="2"><T1012>01-05-23T01:35:45</T1012></DocXML>`,
			expected: "01.05.2023 01:35",
			hasError: false,
		},
		{
			name:     "layout 2006-01-02T15:04",
			xmlStr:   `<DocXML FORM="2"><T1012>2023-05-01T01:35</T1012></DocXML>`,
			expected: "01.05.2023 01:35",
			hasError: false,
		},
		{
			name:     "layout 2006-01-02T15:04:05",
			xmlStr:   `<DocXML FORM="2"><T1012>2023-05-01T01:35:45</T1012></DocXML>`,
			expected: "01.05.2023 01:35",
			hasError: false,
		},
		{
			name:     "layout 02-01-06T15:04 with 4-digit year equivalent",
			xmlStr:   `<DocXML FORM="2"><T1012>01-05-24T01:35</T1012></DocXML>`,
			expected: "01.05.2024 01:35",
			hasError: false,
		},
		{
			name:     "missing T1012 tag",
			xmlStr:   `<DocXML FORM="2"></DocXML>`,
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractDocDateTime(tt.xmlStr)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestRegDataParsing(t *testing.T) {
	xmlStr := `<OK T1077="1234567890"></OK>`
	var reg RegData
	err := xml.Unmarshal([]byte(xmlStr), &reg)
	if err != nil {
		t.Fatalf("failed to unmarshal XML: %v", err)
	}
	if reg.FpNumber != "1234567890" {
		t.Errorf("expected FpNumber '1234567890', got '%s'", reg.FpNumber)
	}
}

// parseDocInfo парсит ответ на <GET DOC='X:fd'/>
func parseDocInfo(data []byte) (offset int64, length int, err error) {
	var docInfo struct {
		Offset string `xml:"OFFSET,attr"`
		Length int    `xml:"LENGTH,attr"`
	}
	if err := decodeXML(data, &docInfo); err != nil {
		return 0, 0, fmt.Errorf("ошибка парсинга информации о документе: %w", err)
	}
	offset, err = strconv.ParseInt(docInfo.Offset, 16, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("ошибка парсинга OFFSET как hex: %w", err)
	}
	return offset, docInfo.Length, nil
}

// decodeHexBlock декодирует ответ на <READ OFFSET='...' LENGTH='...'/>
func decodeHexBlock(data []byte, expectedLength int) ([]byte, error) {
	var blockResp struct {
		Length int    `xml:"LENGTH,attr"`
		Data   string `xml:",innerxml"`
	}
	if err := decodeXML(data, &blockResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга блока: %w", err)
	}
	chunk, err := hex.DecodeString(blockResp.Data)
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования HEX блока: %w", err)
	}
	if len(chunk) != expectedLength {
		return nil, fmt.Errorf("ожидалось %d байт, декодировано %d", expectedLength, len(chunk))
	}
	return chunk, nil
}

func TestParseDocInfo(t *testing.T) {
	tests := []struct {
		name           string
		xmlStr         string
		expectedOffset int64
		expectedLength int
		hasError       bool
	}{
		{
			name:           "valid response",
			xmlStr:         `<OK OFFSET='60010000' LENGTH='385'/>`,
			expectedOffset: 0x60010000,
			expectedLength: 385,
			hasError:       false,
		},
		{
			name:           "invalid offset",
			xmlStr:         `<OK OFFSET='invalid' LENGTH='385'/>`,
			expectedOffset: 0,
			expectedLength: 0,
			hasError:       true,
		},
		{
			name:           "missing attributes",
			xmlStr:         `<OK/>`,
			expectedOffset: 0,
			expectedLength: 0,
			hasError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset, length, err := parseDocInfo([]byte(tt.xmlStr))
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if offset != tt.expectedOffset {
					t.Errorf("expected offset %X, got %X", tt.expectedOffset, offset)
				}
				if length != tt.expectedLength {
					t.Errorf("expected length %d, got %d", tt.expectedLength, length)
				}
			}
		})
	}
}

func TestDecodeHexBlock(t *testing.T) {
	tests := []struct {
		name           string
		xmlStr         string
		expectedLength int
		expectedData   string
		hasError       bool
	}{
		{
			name:           "valid hex data",
			xmlStr:         `<OK LENGTH='4'>48656C6C</OK>`, // "Hell"
			expectedLength: 4,
			expectedData:   "Hell",
			hasError:       false,
		},
		{
			name:           "invalid hex",
			xmlStr:         `<OK LENGTH='4'>invalid</OK>`,
			expectedLength: 4,
			expectedData:   "",
			hasError:       true,
		},
		{
			name:           "length mismatch",
			xmlStr:         `<OK LENGTH='4'>48656C</OK>`, // 3 bytes
			expectedLength: 4,
			expectedData:   "",
			hasError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := decodeHexBlock([]byte(tt.xmlStr), tt.expectedLength)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if string(data) != tt.expectedData {
					t.Errorf("expected data %q, got %q", tt.expectedData, string(data))
				}
			}
		})
	}
}
