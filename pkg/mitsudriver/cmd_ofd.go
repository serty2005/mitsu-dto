package mitsudriver

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// OfdBeginRead начинает процедуру чтения первого непереданного документа для отправки в ОФД.
// Возвращает размер сообщения в байтах.
// Команда: <Do OFD='BEGIN'/>
// Ответ: <OK LENGTH='размер'/>
func (d *mitsuDriver) OfdBeginRead() (int, error) {
	resp, err := d.sendCommand("<Do OFD='BEGIN'/>")
	if err != nil {
		return 0, fmt.Errorf("ошибка начала чтения документа ОФД: %w", err)
	}

	var r struct {
		Length int `xml:"LENGTH,attr"`
	}
	if err := decodeXML(resp, &r); err != nil {
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
func (d *mitsuDriver) OfdReadBlock(offset, length int) ([]byte, int, error) {
	if length > 1000 {
		length = 1000
	}

	cmd := fmt.Sprintf("<Do OFD='READ' OFFSET='%d' LENGTH='%d'/>", offset, length)
	resp, err := d.sendCommand(cmd)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка чтения блока OFD offset=%d length=%d: %w", offset, length, err)
	}

	var r struct {
		Length int    `xml:"LENGTH,attr"`
		Data   string `xml:",innerxml"`
	}
	if err := decodeXML(resp, &r); err != nil {
		return nil, 0, fmt.Errorf("ошибка разбора ответа OFD READ: %w", err)
	}

	// Декодируем HEX данные
	data, err := hex.DecodeString(r.Data)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка декодирования HEX данных OFD: %w", err)
	}

	return data, r.Length, nil
}

// OfdEndRead завершает чтение документа.
// Команда: <Do OFD='END'/>
func (d *mitsuDriver) OfdEndRead() error {
	_, err := d.sendCommand("<Do OFD='END'/>")
	if err != nil {
		return fmt.Errorf("ошибка завершения чтения документа ОФД: %w", err)
	}
	return nil
}

// OfdLoadReceipt записывает квитанцию от ОФД в ФН.
// receipt - бинарные данные квитанции
// Команда: <Do OFD='LOAD' LENGTH='размер'>КВИТАНЦИЯ В HEX</OK>
func (d *mitsuDriver) OfdLoadReceipt(receipt []byte) error {
	hexData := strings.ToUpper(hex.EncodeToString(receipt))
	cmd := fmt.Sprintf("<DO OFD='LOAD' LENGTH='%d'>%s</DO>", len(receipt), hexData)

	_, err := d.sendCommand(cmd)
	if err != nil {
		return fmt.Errorf("ошибка записи квитанции ОФД: %w", err)
	}
	return nil
}

// OfdCancelRead отменяет чтение документа.
// Команда: <Do OFD='CANCEL'/>
func (d *mitsuDriver) OfdCancelRead() error {
	_, err := d.sendCommand("<Do OFD='CANCEL'/>")
	if err != nil {
		return fmt.Errorf("ошибка отмены чтения документа ОФД: %w", err)
	}
	return nil
}

// OfdReadFullDocument читает полный документ для отправки в ОФД.
// Возвращает бинарные данные документа (с обёрткой для ОФД).
func (d *mitsuDriver) OfdReadFullDocument() ([]byte, error) {
	// 1. Начинаем чтение, получаем размер
	totalLength, err := d.OfdBeginRead()
	if err != nil {
		return nil, err
	}

	if totalLength == 0 {
		d.OfdEndRead()
		return nil, fmt.Errorf("документ пуст или отсутствует")
	}

	// 2. Читаем блоками
	const blockSize = 1000 // Максимальный размер блока согласно документации
	var fullData []byte
	offset := 0

	for offset < totalLength {
		remaining := totalLength - offset
		chunkSize := blockSize
		if chunkSize > remaining {
			chunkSize = remaining
		}

		data, actualLen, err := d.OfdReadBlock(offset, chunkSize)
		if err != nil {
			d.OfdCancelRead() // Отменяем при ошибке
			return nil, err
		}

		fullData = append(fullData, data...)
		offset += actualLen

		// Защита от бесконечного цикла
		if actualLen == 0 {
			break
		}
	}

	// 3. Завершаем чтение
	if err := d.OfdEndRead(); err != nil {
		return nil, err
	}

	return fullData, nil
}
