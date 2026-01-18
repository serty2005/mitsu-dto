package mitsu

import (
	"context"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"strings"
)

// OfdBeginRead начинает процедуру чтения первого непереданного документа для отправки в ОФД.
// Возвращает размер сообщения в байтах.
// Команда: <Do OFD='BEGIN'/>
// Ответ: <OK LENGTH='размер'/>
func (c *mitsuClient) OfdBeginRead(ctx context.Context) (int, error) {
	resp, err := c.SendCommand(ctx, "<Do OFD='BEGIN'/>")
	if err != nil {
		return 0, fmt.Errorf("ошибка начала чтения документа ОФД: %w", err)
	}

	var r struct {
		Length int `xml:"LENGTH,attr"`
	}
	if err := xml.Unmarshal(resp, &r); err != nil {
		return 0, fmt.Errorf("ошибка разбора ответа OFD BEGIN: %w", err)
	}

	return r.Length, nil
}

// OfdReadBlock считывает блок сообщения заданной длины, начиная с заданной позиции.
// offset - смещение в байтах от начала сообщения
// length - число байт блока данных, не более 1000 байт
// Возвращает: фактический размер прочитанного блока и данные в бинарном виде.
// Команда: <Do OFD='READ' OFFSET='позиция' LENGTH='размер'/>
// Ответ: <OK LENGTH='размер'>БЛОК ДАННЫХ В HEX</OK>
func (c *mitsuClient) OfdReadBlock(ctx context.Context, offset, length int) ([]byte, int, error) {
	if length > 1000 {
		length = 1000
	}

	cmd := fmt.Sprintf("<Do OFD='READ' OFFSET='%d' LENGTH='%d'/>", offset, length)
	resp, err := c.SendCommand(ctx, cmd)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка чтения блока OFD offset=%d length=%d: %w", offset, length, err)
	}

	var r struct {
		Length int    `xml:"LENGTH,attr"`
		Data   string `xml:",innerxml"`
	}
	if err := xml.Unmarshal(resp, &r); err != nil {
		return nil, 0, fmt.Errorf("ошибка разбора ответа OFD READ: %w", err)
	}

	data, err := hex.DecodeString(r.Data)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка декодирования HEX данных OFD: %w", err)
	}

	return data, r.Length, nil
}

// OfdEndRead завершает чтение документа.
// Команда: <Do OFD='END'/>
func (c *mitsuClient) OfdEndRead(ctx context.Context) error {
	_, err := c.SendCommand(ctx, "<Do OFD='END'/>")
	if err != nil {
		return fmt.Errorf("ошибка завершения чтения документа ОФД: %w", err)
	}
	return nil
}

// OfdLoadReceipt записывает квитанцию от ОФД в ФН.
// receipt - бинарные данные квитанции
// Команда: <Do OFD='LOAD' LENGTH='размер'>КВИТАНЦИЯ В HEX</OK>
func (c *mitsuClient) OfdLoadReceipt(ctx context.Context, receipt []byte) error {
	hexData := strings.ToUpper(hex.EncodeToString(receipt))
	cmd := fmt.Sprintf("<DO OFD='LOAD' LENGTH='%d'>%s</DO>", len(receipt), hexData)

	_, err := c.SendCommand(ctx, cmd)
	if err != nil {
		return fmt.Errorf("ошибка записи квитанции ОФД: %w", err)
	}
	return nil
}

// OfdCancelRead отменяет чтение документа.
// Команда: <Do OFD='CANCEL'/>
func (c *mitsuClient) OfdCancelRead(ctx context.Context) error {
	_, err := c.SendCommand(ctx, "<Do OFD='CANCEL'/>")
	if err != nil {
		return fmt.Errorf("ошибка отмены чтения документа ОФД: %w", err)
	}
	return nil
}

// OfdReadFullDocument читает полный документ для отправки в ОФД.
// Возвращает бинарные данные документа (с обёрткой для ОФД).
func (c *mitsuClient) OfdReadFullDocument(ctx context.Context) ([]byte, error) {
	totalLength, err := c.OfdBeginRead(ctx)
	if err != nil {
		return nil, err
	}

	if totalLength == 0 {
		c.OfdEndRead(ctx)
		return nil, fmt.Errorf("документ пуст или отсутствует")
	}

	const blockSize = 1000
	var fullData []byte
	offset := 0

	for offset < totalLength {
		remaining := totalLength - offset
		chunkSize := blockSize
		if chunkSize > remaining {
			chunkSize = remaining
		}

		data, actualLen, err := c.OfdReadBlock(ctx, offset, chunkSize)
		if err != nil {
			c.OfdCancelRead(ctx)
			return nil, err
		}

		fullData = append(fullData, data...)
		offset += actualLen

		if actualLen == 0 {
			break
		}
	}

	if err := c.OfdEndRead(ctx); err != nil {
		return nil, err
	}

	return fullData, nil
}
