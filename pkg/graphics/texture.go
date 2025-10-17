package graphics

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/go-gl/gl/v3.3-core/gl"
)

// Texture представляет OpenGL текстуру
type Texture struct {
	ID     uint32
	Width  int
	Height int
	Path   string
}

// LoadTexture загружает текстуру из файла
func LoadTexture(path string) (*Texture, error) {
	// Открываем файл
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open texture file: %w", err)
	}
	defer file.Close()

	// Декодируем изображение
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Конвертируем в RGBA
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	// Создаем OpenGL текстуру
	var textureID uint32
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_2D, textureID)

	// Загружаем данные
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix),
	)

	// Настраиваем параметры текстуры
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.BindTexture(gl.TEXTURE_2D, 0)

	return &Texture{
		ID:     textureID,
		Width:  rgba.Rect.Size().X,
		Height: rgba.Rect.Size().Y,
		Path:   path,
	}, nil
}

// Bind привязывает текстуру для использования
func (t *Texture) Bind() {
	gl.BindTexture(gl.TEXTURE_2D, t.ID)
}

// Unbind отвязывает текстуру
func (t *Texture) Unbind() {
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

// Delete удаляет текстуру
func (t *Texture) Delete() {
	gl.DeleteTextures(1, &t.ID)
}
