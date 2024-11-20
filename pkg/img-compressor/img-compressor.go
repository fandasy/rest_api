package img_compressor

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/webp"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"net/http"
	"restApi/pkg/e"
	"restApi/pkg/validate"

	"github.com/nfnt/resize"
)

var (
	ErrPageNotFound    = errors.New("page not found")
	ErrIncorrectFormat = errors.New("incorrect format")
)

func Get(ctx context.Context, url string, compressionPercentage float64, maxWidth int, maxHeight int, chars string) (*image.RGBA, error) {
	const op = "img_compressor.Get"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	req.Close = true

	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, e.Wrap(op, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, e.Wrap(op, ErrPageNotFound)
	}

	contentType := resp.Header.Get("Content-Type")

	if !validate.ContentType(contentType, "image/png", "image/jpeg", "image/webp") {
		return nil, fmt.Errorf("%s: %w: %s", op, ErrIncorrectFormat, contentType)
	}

	var img image.Image

	switch contentType {
	case "image/png":
		img, err = png.Decode(resp.Body)
	case "image/jpeg":
		img, err = jpeg.Decode(resp.Body)
	case "image/webp":
		img, err = webp.Decode(resp.Body)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if compressionPercentage < 0 || compressionPercentage > 1 {
		compressionPercentage = 0.0
	}

	if maxWidth <= 0 {
		maxWidth = 5000 // 50000px
	}

	if maxHeight <= 0 {
		maxHeight = 5000 // 50000px
	}

	if chars == "" {
		chars = "@%#*+=:~-. "
	}

	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	if width > maxWidth {
		width = maxWidth
	}

	if height > maxHeight {
		height = maxHeight
	}

	newWidth := uint(float64(width) * (1 - compressionPercentage))
	newHeight := uint(float64(height) * (1 - compressionPercentage))

	img = resize.Resize(newWidth, newHeight, img, resize.Lanczos2)

	return generateASCIIImage(img, chars), nil
}

func generateASCIIImage(img image.Image, chars string) *image.RGBA {
	bounds := img.Bounds()
	asciiWidth := bounds.Max.X
	asciiHeight := bounds.Max.Y

	// Создаем новое изображение
	asciiImg := image.NewRGBA(image.Rect(0, 0, asciiWidth*10, asciiHeight*10)) // Увеличиваем размер для текста

	draw.Draw(asciiImg, asciiImg.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			char := getCharFromBrightness(c, chars)

			// Рисуем символ на изображении
			point := fixed.Point26_6{X: fixed.I(x * 10), Y: fixed.I(y * 10)} // Смещение для текста
			d := &font.Drawer{
				Dst:  asciiImg,
				Src:  image.NewUniform(color.Black),
				Face: basicfont.Face7x13,
				Dot:  point,
			}
			d.DrawString(char)
		}
	}

	return asciiImg
}

func getCharFromBrightness(c color.Color, chars string) string {
	r, g, b, _ := c.RGBA()

	avg := (r + g + b) >> 8 // сдвиг вправо для получения 8-битного значения
	brightness := float64(avg) / 255.0

	idx := int(brightness * float64(len(chars)-1))

	if idx < 0 {
		idx = 0
	} else if idx >= len(chars) {
		idx = len(chars) - 1
	}

	return string(chars[idx])
}
