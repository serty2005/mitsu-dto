package ofdclient

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// MessageHeaderSize — размер заголовка сообщения в байтах
const MessageHeaderSize = 30

// SerializeMessage сериализует сообщение в байтовый поток согласно спецификации ФНС
// Порядок байт:
//   - Signature, SProtoVersion, PProtoVersion, FnNumber — порядок нумерации
//   - BodySize, Flags, CRC — Little Endian
//
// [0:4]   = 2A 08 41 0A     (сигнатура — порядок нумерации)
// [4:6]   = 81 A2           (S-proto — порядок нумерации)
// [6:8]   = 01 20           (FFD 1.2 — порядок нумерации)
// [24:26] = Little Endian   (BodySize)
// [26:28] = Little Endian   (Flags)
// [28:30] = Little Endian   (CRC)
func SerializeMessage(header *MessageHeader, body []byte) ([]byte, error) {
	if header == nil {
		return nil, errors.New("header cannot be nil")
	}

	buf := make([]byte, MessageHeaderSize+len(body))

	// Сигнатура — массив байт, порядок нумерации (копируем как есть)
	copy(buf[0:4], header.Signature[:])

	// Версия S-протокола — массив байт, порядок нумерации
	copy(buf[4:6], header.SProtoVersion[:])

	// Версия P-протокола — массив байт, порядок нумерации
	copy(buf[6:8], header.PProtoVersion[:])

	// Номер ФН — 16 байт ASCII
	copy(buf[8:24], header.FnNumber[:])

	// Размер тела — Little Endian!
	binary.LittleEndian.PutUint16(buf[24:26], header.BodySize)

	// Флаги — Little Endian!
	binary.LittleEndian.PutUint16(buf[26:28], uint16(header.Flags))

	// Тело сообщения
	if len(body) > 0 {
		copy(buf[MessageHeaderSize:], body)
	}

	// Вычисляем CRC в зависимости от флагов
	var crc uint16
	if header.Flags&FlagCRCFull != 0 {
		// CRC по заголовку (без поля CRC) + тело
		crc = calculateMessageCRC(buf[:28], body)
	} else if header.Flags&FlagCRCHeader != 0 {
		// CRC только по заголовку (без поля CRC)
		crc = calculateMessageCRC(buf[:28], nil)
	}

	// Записываем CRC — Little Endian!
	binary.LittleEndian.PutUint16(buf[28:30], crc)

	return buf, nil
}

// DeserializeMessage десериализует байтовый поток в сообщение
func DeserializeMessage(data []byte) (*MessageHeader, []byte, error) {
	if len(data) < MessageHeaderSize {
		return nil, nil, fmt.Errorf("message too short: %d < %d bytes", len(data), MessageHeaderSize)
	}

	header := &MessageHeader{}

	// Сигнатура — массив байт
	copy(header.Signature[:], data[0:4])

	// Версии — массивы байт
	copy(header.SProtoVersion[:], data[4:6])
	copy(header.PProtoVersion[:], data[6:8])

	// Номер ФН
	copy(header.FnNumber[:], data[8:24])

	// Размер тела — Little Endian!
	header.BodySize = binary.LittleEndian.Uint16(data[24:26])

	// Флаги — Little Endian!
	header.Flags = MessageFlags(binary.LittleEndian.Uint16(data[26:28]))

	// CRC — Little Endian!
	header.CRC = binary.LittleEndian.Uint16(data[28:30])

	// Тело
	var body []byte
	if header.BodySize > 0 && len(data) > MessageHeaderSize {
		bodyEnd := MessageHeaderSize + int(header.BodySize)
		if bodyEnd > len(data) {
			bodyEnd = len(data)
		}
		body = data[MessageHeaderSize:bodyEnd]
	}

	// Проверка CRC
	if header.Flags&FlagCRCFull != 0 {
		expectedCRC := calculateMessageCRC(data[:28], body)
		if header.CRC != expectedCRC {
			return nil, nil, fmt.Errorf("%w: got %04X, expected %04X", ErrCRCMismatch, header.CRC, expectedCRC)
		}
	} else if header.Flags&FlagCRCHeader != 0 {
		expectedCRC := calculateMessageCRC(data[:28], nil)
		if header.CRC != expectedCRC {
			return nil, nil, fmt.Errorf("%w: got %04X, expected %04X", ErrCRCMismatch, header.CRC, expectedCRC)
		}
	}

	return header, body, nil
}

// CreateMessageHeader создает заголовок сообщения с заданными параметрами
func CreateMessageHeader(fnNumber string, ffdVersion string, flags MessageFlags, bodySize uint16) (*MessageHeader, error) {
	if len(fnNumber) != 16 {
		return nil, fmt.Errorf("%w: got %d chars", ErrInvalidFnNumber, len(fnNumber))
	}

	pProtoBytes, err := FFDVersionToBytes(ffdVersion)
	if err != nil {
		return nil, err
	}

	header := &MessageHeader{
		Signature:     SignatureOFDBytes,
		SProtoVersion: SProtoVersionOFDBytes,
		PProtoVersion: pProtoBytes,
		BodySize:      bodySize,
		Flags:         flags,
	}

	copy(header.FnNumber[:], []byte(fnNumber))

	return header, nil
}

// SerializeContainer сериализует контейнер в байтовый поток
// Все поля в Little Endian
func SerializeContainer(header *ContainerHeader, data []byte) ([]byte, error) {
	if header == nil {
		return nil, errors.New("container header cannot be nil")
	}

	// Длина контейнера = заголовок (7 байт) + данные
	header.Length = uint16(7 + len(data))

	buf := make([]byte, 7+len(data))

	// Length — Little Endian
	binary.LittleEndian.PutUint16(buf[0:2], header.Length)

	// CRC вычисляем позже, пока пропускаем (байты 2-3)

	buf[4] = header.ContainerType
	buf[5] = header.DataType
	buf[6] = header.FormatVersion

	// Данные
	if len(data) > 0 {
		copy(buf[7:], data)
	}

	// Вычисляем CRC контейнера (по Length + Type + DataType + FormatVersion + Data, без CRC)
	crcData := make([]byte, 2+3+len(data))
	copy(crcData[0:2], buf[0:2]) // Length
	copy(crcData[2:5], buf[4:7]) // Type, DataType, FormatVersion
	copy(crcData[5:], buf[7:])   // Data
	crc := calcCRC16(crcData)

	// CRC — Little Endian
	binary.LittleEndian.PutUint16(buf[2:4], crc)

	return buf, nil
}

// DeserializeContainer десериализует байтовый поток в контейнер
func DeserializeContainer(data []byte) (*ContainerHeader, []byte, error) {
	if len(data) < 7 {
		return nil, nil, errors.New("container too short, must be at least 7 bytes")
	}

	header := &ContainerHeader{}

	// Little Endian
	header.Length = binary.LittleEndian.Uint16(data[0:2])
	header.CRC = binary.LittleEndian.Uint16(data[2:4])
	header.ContainerType = data[4]
	header.DataType = data[5]
	header.FormatVersion = data[6]

	// Данные
	var containerData []byte
	if len(data) > 7 {
		containerData = data[7:]
	}

	// Проверяем CRC
	crcData := make([]byte, 2+3+len(containerData))
	copy(crcData[0:2], data[0:2])
	copy(crcData[2:5], data[4:7])
	copy(crcData[5:], containerData)
	expectedCRC := calcCRC16(crcData)

	if header.CRC != expectedCRC {
		return nil, nil, fmt.Errorf("container CRC mismatch: got %04X, expected %04X", header.CRC, expectedCRC)
	}

	return header, containerData, nil
}

// CreateContainerHeader создает заголовок контейнера
func CreateContainerHeader(containerType byte, dataType byte, formatVersion byte) *ContainerHeader {
	return &ContainerHeader{
		ContainerType: containerType,
		DataType:      dataType,
		FormatVersion: formatVersion,
	}
}

// calculateMessageCRC вычисляет CRC для заголовка сообщения (и опционально тела)
// headerWithoutCRC должен быть 28 байт (заголовок без поля CRC)
func calculateMessageCRC(headerWithoutCRC []byte, body []byte) uint16 {
	if body == nil {
		return calcCRC16(headerWithoutCRC)
	}
	data := make([]byte, len(headerWithoutCRC)+len(body))
	copy(data, headerWithoutCRC)
	copy(data[len(headerWithoutCRC):], body)
	return calcCRC16(data)
}
