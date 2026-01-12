package registration

import (
	"fmt"
)

// CalculateRNM выполняет расчет РНМ по алгоритму CRC16-CCITT
// Формат входных данных для CRC:
// Pad(Order, 10) + Pad(INN, 12) + Pad(Serial, 20)
// Результат: Pad(Order, 10) + Pad(CRC, 6)
//
// Параметры:
//
//	orderNum - порядковый номер регистрации (обычно "1")
//	inn - ИНН пользователя (10 или 12 цифр)
//	serial - заводской номер ККТ (20 знаков)
//
// Возвращает сгенерированный РНМ или ошибку
func CalculateRNM(orderNum, inn, serial string) (string, error) {
	// 1. Формируем строку для расчета
	paddedOrder := padLeft(orderNum, 10, '0')
	paddedInn := padLeft(inn, 12, '0')
	paddedSerial := padLeft(serial, 20, '0')

	// Строка: 0000000001 + 007804437548 + 00000000000000000156 (пример)
	calcString := paddedOrder + paddedInn + paddedSerial

	// 2. Считаем CRC
	crc := crc16ccitt([]byte(calcString))

	// 3. Формируем хвост (CRC дополненный до 6 цифр нулями)
	// Пример: CRC 33271 -> "033271"
	crcStr := fmt.Sprintf("%d", crc)
	paddedCrc := padLeft(crcStr, 6, '0')

	// 4. Итоговый РНМ
	finalRnm := paddedOrder + paddedCrc

	return finalRnm, nil
}

// crc16ccitt вычисляет CRC-16 (CCITT False)
// Poly: 0x1021, Init: 0xFFFF
func crc16ccitt(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if (crc & 0x8000) != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// padLeft дополняет строку символом padChar слева до длины length
func padLeft(s string, length int, padChar byte) string {
	if len(s) >= length {
		return s // Или обрезать, если требуется строго length
	}
	padding := make([]byte, length-len(s))
	for i := range padding {
		padding[i] = padChar
	}
	return string(padding) + s
}
