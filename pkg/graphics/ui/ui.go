package ui

import (
	"github.com/Salamander5876/AnimoEngine/pkg/graphics/shader"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// UIRenderer рендерер для 2D UI элементов
type UIRenderer struct {
	shader *shader.Shader
	vao    uint32
	vbo    uint32
}

// NewUIRenderer создает новый UI рендерер
func NewUIRenderer() (*UIRenderer, error) {
	// Шейдер для UI (2D, без трансформаций камеры)
	vertexShader := `
#version 330 core
layout (location = 0) in vec2 aPos;
layout (location = 1) in vec2 aTexCoord;
layout (location = 2) in vec4 aColor;

out vec2 TexCoord;
out vec4 Color;

uniform mat4 projection;

void main() {
    gl_Position = projection * vec4(aPos, 0.0, 1.0);
    TexCoord = aTexCoord;
    Color = aColor;
}
`

	fragmentShader := `
#version 330 core
in vec2 TexCoord;
in vec4 Color;
out vec4 FragColor;

uniform sampler2D texture1;
uniform bool useTexture;

void main() {
    if (useTexture) {
        FragColor = texture(texture1, TexCoord) * Color;
    } else {
        FragColor = Color;
    }
}
`

	shaderProgram, err := shader.NewShader(vertexShader, fragmentShader)
	if err != nil {
		return nil, err
	}

	// Создаем VAO и VBO для динамической геометрии
	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)

	// Позиция (2 float)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Текстурные координаты (2 float)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)

	// Цвет (4 float)
	gl.VertexAttribPointer(2, 4, gl.FLOAT, false, 8*4, gl.PtrOffset(4*4))
	gl.EnableVertexAttribArray(2)

	gl.BindVertexArray(0)

	return &UIRenderer{
		shader: shaderProgram,
		vao:    vao,
		vbo:    vbo,
	}, nil
}

// SetProjection устанавливает ортографическую проекцию
func (r *UIRenderer) SetProjection(width, height float32) {
	projection := mgl32.Ortho(0, width, height, 0, -1, 1)
	r.shader.Use()
	r.shader.SetMat4("projection", projection)
}

// DrawRect рисует прямоугольник
func (r *UIRenderer) DrawRect(x, y, width, height float32, color mgl32.Vec4) {
	vertices := []float32{
		// Позиции      // TexCoords  // Цвет
		x, y,           0, 0,         color[0], color[1], color[2], color[3],
		x + width, y,   1, 0,         color[0], color[1], color[2], color[3],
		x + width, y + height, 1, 1,  color[0], color[1], color[2], color[3],

		x, y,           0, 0,         color[0], color[1], color[2], color[3],
		x + width, y + height, 1, 1,  color[0], color[1], color[2], color[3],
		x, y + height,  0, 1,         color[0], color[1], color[2], color[3],
	}

	r.shader.Use()
	r.shader.SetBool("useTexture", false)

	gl.BindVertexArray(r.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.DYNAMIC_DRAW)

	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
}

// DrawLine рисует линию
func (r *UIRenderer) DrawLine(x1, y1, x2, y2, thickness float32, color mgl32.Vec4) {
	// Вычисляем перпендикуляр для толщины
	dx := x2 - x1
	dy := y2 - y1
	length := float32(mgl32.Vec2{dx, dy}.Len())
	if length == 0 {
		return
	}

	perpX := -dy / length * thickness * 0.5
	perpY := dx / length * thickness * 0.5

	vertices := []float32{
		// Позиции                    // TexCoords  // Цвет
		x1 - perpX, y1 - perpY,       0, 0,         color[0], color[1], color[2], color[3],
		x1 + perpX, y1 + perpY,       1, 0,         color[0], color[1], color[2], color[3],
		x2 + perpX, y2 + perpY,       1, 1,         color[0], color[1], color[2], color[3],

		x1 - perpX, y1 - perpY,       0, 0,         color[0], color[1], color[2], color[3],
		x2 + perpX, y2 + perpY,       1, 1,         color[0], color[1], color[2], color[3],
		x2 - perpX, y2 - perpY,       0, 1,         color[0], color[1], color[2], color[3],
	}

	r.shader.Use()
	r.shader.SetBool("useTexture", false)

	gl.BindVertexArray(r.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.DYNAMIC_DRAW)

	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
}

// Cleanup освобождает ресурсы
func (r *UIRenderer) Cleanup() {
	gl.DeleteVertexArrays(1, &r.vao)
	gl.DeleteBuffers(1, &r.vbo)
}
