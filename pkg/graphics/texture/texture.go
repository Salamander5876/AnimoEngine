package texture

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"os"

	"github.com/go-gl/gl/v3.3-core/gl"
)

// LoadTexture загружает текстуру из файла PNG
func LoadTexture(filepath string) (uint32, error) {
	// Открываем файл
	imgFile, err := os.Open(filepath)
	if err != nil {
		return 0, fmt.Errorf("failed to open texture file %s: %v", filepath, err)
	}
	defer imgFile.Close()

	// Декодируем изображение
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return 0, fmt.Errorf("failed to decode texture %s: %v", filepath, err)
	}

	// Конвертируем в RGBA
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	// Создаём текстуру в OpenGL
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	// Настройки текстуры
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// Загружаем данные текстуры
	width := int32(rgba.Bounds().Dx())
	height := int32(rgba.Bounds().Dy())
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		width,
		height,
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix),
	)

	// Генерируем mipmap
	gl.GenerateMipmap(gl.TEXTURE_2D)

	gl.BindTexture(gl.TEXTURE_2D, 0)

	return texture, nil
}

// Cleanup удаляет текстуру
func Cleanup(texture uint32) {
	gl.DeleteTextures(1, &texture)
}