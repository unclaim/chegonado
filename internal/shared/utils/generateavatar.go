package utils

import (
	"crypto/md5"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/rand"
	"os"
)

// hslToRgb конвертирует цвет из HSL в RGB.
// GitHub использует HSL для создания гармоничных цветов.
func hslToRgb(h, s, l float64) (uint8, uint8, uint8) {
	var r, g, b float64

	if s == 0 {
		// Если насыщенность равна 0, цвет серый
		r, g, b = l, l, l
	} else {
		hueToRgb := func(p, q, t float64) float64 {
			if t < 0 {
				t += 1
			}
			if t > 1 {
				t -= 1
			}
			if t < 1.0/6.0 {
				return p + (q-p)*6*t
			}
			if t < 1.0/2.0 {
				return q
			}
			if t < 2.0/3.0 {
				return p + (q-p)*(2.0/3.0-t)*6
			}
			return p
		}

		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hueToRgb(p, q, h+1.0/3.0)
		g = hueToRgb(p, q, h)
		b = hueToRgb(p, q, h-1.0/3.0)
	}

	return uint8(r * 255), uint8(g * 255), uint8(b * 255)
}

// GenerateAvatar создает и сохраняет аватар в стиле GitHub
func GenerateAvatar(seed, filename string) error {
	// 1. Хешируем входную строку
	hash := md5.Sum([]byte(seed))

	// Используем хеш для инициализации генератора случайных чисел
	seedValue := int64(hash[0]) | int64(hash[1])<<8 | int64(hash[2])<<16 | int64(hash[3])<<24
	r := rand.New(rand.NewSource(seedValue))

	// 2. Генерируем цвет в HSL и конвертируем в RGB
	// Используем часть хеша, чтобы получить оттенок (Hue)
	h := float64(hash[7]%255) / 255.0
	// Фиксируем насыщенность (Saturation) и яркость (Lightness)
	s := 0.6
	l := 0.6

	// Конвертируем HSL в RGB
	rgbR, rgbG, rgbB := hslToRgb(h, s, l)
	c := color.RGBA{R: rgbR, G: rgbG, B: rgbB, A: 255}

	// Задаем размер аватара, сетки и отступы
	const size = 300
	const gridSize = 5
	const padding = 30 // Отступ от края изображения

	// Рассчитываем размер области для отрисовки узора
	const drawSize = size - 2*padding
	// Рассчитываем новый размер квадратов с учётом отступов
	const squareSize = drawSize / gridSize

	// Создаем новое изображение
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Закрашиваем фон серым цветом (как у GitHub)
	bgColor := color.RGBA{R: 242, G: 242, B: 242, A: 255}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// 3. Создаем симметричный узор из квадратов
	grid := [gridSize][gridSize]bool{}
	for y := 0; y < gridSize; y++ {
		for x := 0; x < (gridSize+1)/2; x++ {
			if r.Intn(2) == 1 {
				grid[x][y] = true
				// Отзеркаливаем на правую сторону
				grid[gridSize-1-x][y] = true
			}
		}
	}

	// 4. Рисуем узор на изображении с отступами
	for y := 0; y < gridSize; y++ {
		for x := 0; x < gridSize; x++ {
			if grid[x][y] {
				// Закрашиваем квадрат
				for i := 0; i < squareSize; i++ {
					for j := 0; j < squareSize; j++ {
						// Смещаем координаты на padding
						img.Set(padding+x*squareSize+i, padding+y*squareSize+j, c)
					}
				}
			}
		}
	}

	// 5. Сохраняем изображение в файл
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			// Handle the error appropriately, e.g., log it
			log.Printf("Error closing file: %v", err)
		}
	}()

	return png.Encode(file, img)
}
