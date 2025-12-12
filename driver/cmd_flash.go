package driver

import (
	"encoding/hex"
	"fmt"
)

// UploadImage реализует загрузку изображения с разбивкой на пакеты (Chunking).
// Это необходимо, так как буфер COM-порта ограничен 2040 байтами.
func (d *mitsuDriver) UploadImage(index int, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("файл изображения пуст")
	}

	// Максимальный размер полезной нагрузки (бинарной) в одном пакете.
	// 512 байт bin -> 1024 байт hex.
	// 1024 + ~60 байт XML-обвязки < 2040 байт (лимит протокола).
	const maxChunkSize = 512

	// OFFSET в данном контексте - это ID слота (100 + номер).
	// Мы предполагаем, что последовательные команды MODE='1' в один и тот же слот
	// дописывают данные в буфер устройства.
	offset := 100 + index

	totalLen := len(data)
	sent := 0

	for sent < totalLen {
		// 1. Определяем размер текущего куска
		chunkSize := maxChunkSize
		if sent+chunkSize > totalLen {
			chunkSize = totalLen - sent
		}

		// 2. Выделяем данные и кодируем в HEX
		chunk := data[sent : sent+chunkSize]
		hexData := hex.EncodeToString(chunk)

		// 3. Формируем команду записи части
		// LENGTH - указывает размер текущей порции данных
		cmdWrite := fmt.Sprintf("<FLASH MODE='1' LENGTH='%d' OFFSET='%d'>%s</FLASH>", len(chunk), offset, hexData)

		if d.config.Logger != nil {
			d.config.Logger(fmt.Sprintf("Загрузка пакета: %d/%d байт...", sent+chunkSize, totalLen))
		}

		// 4. Отправляем
		if _, err := d.sendCommand(cmdWrite); err != nil {
			return fmt.Errorf("ошибка записи блока (offset %d): %w", sent, err)
		}

		sent += chunkSize
	}

	// 5. Фиксация (Commit)
	// MODE='3' завершает загрузку и сохраняет буфер во флеш-память.
	cmdCommit := "<FLASH MODE='3' LENGTH='0' OFFSET='0'/>"
	if _, err := d.sendCommand(cmdCommit); err != nil {
		return fmt.Errorf("ошибка фиксации изображения (MODE=3): %w", err)
	}

	return nil
}
