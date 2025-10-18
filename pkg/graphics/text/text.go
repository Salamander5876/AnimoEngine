package text

import (
	"image"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// TextRenderer рендерер текста
type TextRenderer struct {
	shader    uint32
	vao       uint32
	vbo       uint32
	texture   uint32
	charWidth int
	charHeight int
}

// NewTextRenderer создает новый текстовый рендерер
func NewTextRenderer() (*TextRenderer, error) {
	tr := &TextRenderer{
		charWidth:  8,
		charHeight: 16,
	}

	// Создаем текстуру с символами ASCII
	if err := tr.createFontTexture(); err != nil {
		return nil, err
	}

	// Создаем VAO и VBO
	tr.createGeometry()

	// Создаем шейдер
	if err := tr.createShader(); err != nil {
		return nil, err
	}

	return tr, nil
}

func (tr *TextRenderer) createFontTexture() error {
	// Создаем изображение для всех символов ASCII (32-127)
	const chars = 96 // символов
	const cols = 16  // колонок
	rows := chars / cols

	img := image.NewRGBA(image.Rect(0, 0, cols*tr.charWidth, rows*tr.charHeight))

	// Рисуем каждый символ
	d := &font.Drawer{
		Dst:  img,
		Src:  image.White,
		Face: basicfont.Face7x13,
	}

	for i := 0; i < chars; i++ {
		ch := rune(32 + i) // ASCII начинается с 32 (пробел)
		x := (i % cols) * tr.charWidth
		y := (i/cols)*tr.charHeight + 12 // Смещение для базовой линии

		d.Dot = fixed.P(x, y)
		d.DrawString(string(ch))
	}

	// Создаем OpenGL текстуру
	gl.GenTextures(1, &tr.texture)
	gl.BindTexture(gl.TEXTURE_2D, tr.texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(img.Bounds().Dx()),
		int32(img.Bounds().Dy()),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(img.Pix),
	)

	return nil
}

func (tr *TextRenderer) createGeometry() {
	vertices := []float32{
		// Позиции    // TexCoords
		0, 0,         0, 0,
		1, 0,         1, 0,
		1, 1,         1, 1,

		0, 0,         0, 0,
		1, 1,         1, 1,
		0, 1,         0, 1,
	}

	gl.GenVertexArrays(1, &tr.vao)
	gl.GenBuffers(1, &tr.vbo)

	gl.BindVertexArray(tr.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, tr.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Позиция
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	// TexCoord
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}

func (tr *TextRenderer) createShader() error {
	vertexShader := `
#version 330 core
layout (location = 0) in vec2 aPos;
layout (location = 1) in vec2 aTexCoord;

out vec2 TexCoord;

uniform mat4 projection;
uniform mat4 model;

void main() {
    gl_Position = projection * model * vec4(aPos, 0.0, 1.0);
    TexCoord = aTexCoord;
}
`

	fragmentShader := `
#version 330 core
in vec2 TexCoord;
out vec4 FragColor;

uniform sampler2D text;
uniform vec4 textColor;

void main() {
    vec4 sampled = texture(text, TexCoord);
    FragColor = vec4(textColor.rgb, sampled.a);
}
`

	// Компилируем шейдеры
	vs := gl.CreateShader(gl.VERTEX_SHADER)
	csources, free := gl.Strs(vertexShader + "\x00")
	gl.ShaderSource(vs, 1, csources, nil)
	free()
	gl.CompileShader(vs)

	fs := gl.CreateShader(gl.FRAGMENT_SHADER)
	csources, free = gl.Strs(fragmentShader + "\x00")
	gl.ShaderSource(fs, 1, csources, nil)
	free()
	gl.CompileShader(fs)

	// Создаем программу
	tr.shader = gl.CreateProgram()
	gl.AttachShader(tr.shader, vs)
	gl.AttachShader(tr.shader, fs)
	gl.LinkProgram(tr.shader)

	gl.DeleteShader(vs)
	gl.DeleteShader(fs)

	return nil
}

// DrawText рисует текст на экране
func (tr *TextRenderer) DrawText(text string, x, y, scale float32, color mgl32.Vec4, projection mgl32.Mat4) {
	gl.UseProgram(tr.shader)

	// Устанавливаем uniform'ы
	projLoc := gl.GetUniformLocation(tr.shader, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])

	colorLoc := gl.GetUniformLocation(tr.shader, gl.Str("textColor\x00"))
	gl.Uniform4f(colorLoc, color[0], color[1], color[2], color[3])

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, tr.texture)

	gl.BindVertexArray(tr.vao)

	currentX := x

	for _, ch := range text {
		if ch < 32 || ch > 127 {
			ch = '?' // Замена для неподдерживаемых символов
		}

		charIndex := int(ch - 32)
		cols := 16
		col := charIndex % cols
		row := charIndex / cols

		// Вычисляем UV координаты для символа
		uvX := float32(col) / float32(cols)
		uvY := float32(row) / 6.0 // 6 рядов
		uvW := 1.0 / float32(cols)
		uvH := float32(1.0 / 6.0)

		// Обновляем VBO с правильными UV координатами (инвертируем Y)
		vertices := []float32{
			0, 0, uvX, uvY + uvH,
			1, 0, uvX + uvW, uvY + uvH,
			1, 1, uvX + uvW, uvY,

			0, 0, uvX, uvY + uvH,
			1, 1, uvX + uvW, uvY,
			0, 1, uvX, uvY,
		}

		gl.BindBuffer(gl.ARRAY_BUFFER, tr.vbo)
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(vertices)*4, gl.Ptr(vertices))

		// Матрица модели для этого символа
		model := mgl32.Translate3D(currentX, y, 0)
		model = model.Mul4(mgl32.Scale3D(float32(tr.charWidth)*scale, float32(tr.charHeight)*scale, 1))

		modelLoc := gl.GetUniformLocation(tr.shader, gl.Str("model\x00"))
		gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])

		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		currentX += float32(tr.charWidth) * scale
	}

	gl.BindVertexArray(0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

// Cleanup освобождает ресурсы
func (tr *TextRenderer) Cleanup() {
	gl.DeleteVertexArrays(1, &tr.vao)
	gl.DeleteBuffers(1, &tr.vbo)
	gl.DeleteTextures(1, &tr.texture)
	gl.DeleteProgram(tr.shader)
}
