package ofdclient

// CRC16-CCITT реализация для протокола ОФД
// Полином: 0x1021 (стандарт CCITT)
// Начальное значение: 0xFFFF
// Рефлексия: нет
// XOR на выходе: 0x0000

// crc16Table — предвычисленная таблица для CRC16-CCITT
var crc16Table [256]uint16

func init() {
	poly := uint16(0x1021)
	for i := 0; i < 256; i++ {
		crc := uint16(i) << 8
		for j := 0; j < 8; j++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc <<= 1
			}
		}
		crc16Table[i] = crc
	}
}

// calcCRC16 вычисляет CRC16-CCITT для произвольных данных
func calcCRC16(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc = (crc << 8) ^ crc16Table[(crc>>8)^uint16(b)]
	}
	return crc
}

// CalculateCRC16 — публичная функция для расчета CRC16-CCITT
func CalculateCRC16(data []byte) uint16 {
	return calcCRC16(data)
}
