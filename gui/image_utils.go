package gui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg" // Поддержка декодирования JPG
	_ "image/png"  // Поддержка декодирования PNG
	"os"

	"golang.org/x/image/bmp"
	"golang.org/x/image/draw"
)

// PrepareImageForKKT загружает изображение, изменяет размер и конвертирует в монохромный BMP.
// maxWidth: максимальная ширина (обычно 384 для 57мм или 576 для 80мм).
func PrepareImageForKKT(path string, maxWidth int) ([]byte, error) {
	// 1. Открываем файл
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования изображения: %w", err)
	}

	// 2. Ресайз (сохраняя пропорции)
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var newWidth, newHeight int
	if width > maxWidth {
		ratio := float64(maxWidth) / float64(width)
		newWidth = maxWidth
		newHeight = int(float64(height) * ratio)
	} else {
		newWidth = width
		newHeight = height
	}

	// Создаем rect для ресайза
	// Используем Gray, чтобы сразу получить ч/б оттенки
	grayImg := image.NewGray(image.Rect(0, 0, newWidth, newHeight))

	// Используем бикубическую интерполяцию для качественного ресайза
	draw.CatmullRom.Scale(grayImg, grayImg.Bounds(), img, bounds, draw.Over, nil)

	// 3. Бинаризация (Thresholding) -> превращаем в 1-битный Paletted
	// Создаем палитру: Индекс 0 - Черный, 1 - Белый (стандарт для термопечати)
	palette := color.Palette{
		color.Black,
		color.White,
	}

	palettedImg := image.NewPaletted(grayImg.Bounds(), palette)

	// Порог бинаризации (128 из 255)
	threshold := 128

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			grayColor := grayImg.GrayAt(x, y)
			if int(grayColor.Y) > threshold {
				palettedImg.SetColorIndex(x, y, 1) // Белый
			} else {
				palettedImg.SetColorIndex(x, y, 0) // Черный
			}
		}
	}

	// 4. Кодирование в BMP
	var buf bytes.Buffer
	// bmp.Encode поддерживает сохранение Paletted изображений с глубиной 1 бит,
	// если в палитре 2 цвета.
	if err := bmp.Encode(&buf, palettedImg); err != nil {
		return nil, fmt.Errorf("ошибка кодирования в BMP: %w", err)
	}

	return buf.Bytes(), nil
}
