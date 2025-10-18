package model

import (
	"fmt"
	"os"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/Salamander5876/AnimoEngine/pkg/graphics/texture"
)

// LoadFBXSimple загружает FBX файл упрощённо - создаёт простой куб с текстурой
// Для полноценной загрузки FBX нужна внешняя библиотека
func LoadFBXSimple(fbxPath string, texturePath string) (*Model, error) {
	// Проверяем существование файлов
	if _, err := os.Stat(fbxPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("FBX file not found: %s", fbxPath)
	}

	// Создаём модель
	model := NewModel()
	model.FilePath = fbxPath

	// Загружаем текстуру
	var textureID uint32
	var err error
	if texturePath != "" {
		textureID, err = texture.LoadTexture(texturePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load texture: %v", err)
		}
	}

	// Создаём простую геометрию (куб с текстурными координатами)
	// В реальности здесь должен быть парсинг FBX, но для простоты используем куб
	vertices := []Vertex{
		// Front face
		{Position: mgl32.Vec3{-0.5, -0.5, 0.5}, Normal: mgl32.Vec3{0, 0, 1}, TexCoords: mgl32.Vec2{0, 0}},
		{Position: mgl32.Vec3{0.5, -0.5, 0.5}, Normal: mgl32.Vec3{0, 0, 1}, TexCoords: mgl32.Vec2{1, 0}},
		{Position: mgl32.Vec3{0.5, 0.5, 0.5}, Normal: mgl32.Vec3{0, 0, 1}, TexCoords: mgl32.Vec2{1, 1}},
		{Position: mgl32.Vec3{-0.5, 0.5, 0.5}, Normal: mgl32.Vec3{0, 0, 1}, TexCoords: mgl32.Vec2{0, 1}},

		// Back face
		{Position: mgl32.Vec3{0.5, -0.5, -0.5}, Normal: mgl32.Vec3{0, 0, -1}, TexCoords: mgl32.Vec2{0, 0}},
		{Position: mgl32.Vec3{-0.5, -0.5, -0.5}, Normal: mgl32.Vec3{0, 0, -1}, TexCoords: mgl32.Vec2{1, 0}},
		{Position: mgl32.Vec3{-0.5, 0.5, -0.5}, Normal: mgl32.Vec3{0, 0, -1}, TexCoords: mgl32.Vec2{1, 1}},
		{Position: mgl32.Vec3{0.5, 0.5, -0.5}, Normal: mgl32.Vec3{0, 0, -1}, TexCoords: mgl32.Vec2{0, 1}},

		// Left face
		{Position: mgl32.Vec3{-0.5, -0.5, -0.5}, Normal: mgl32.Vec3{-1, 0, 0}, TexCoords: mgl32.Vec2{0, 0}},
		{Position: mgl32.Vec3{-0.5, -0.5, 0.5}, Normal: mgl32.Vec3{-1, 0, 0}, TexCoords: mgl32.Vec2{1, 0}},
		{Position: mgl32.Vec3{-0.5, 0.5, 0.5}, Normal: mgl32.Vec3{-1, 0, 0}, TexCoords: mgl32.Vec2{1, 1}},
		{Position: mgl32.Vec3{-0.5, 0.5, -0.5}, Normal: mgl32.Vec3{-1, 0, 0}, TexCoords: mgl32.Vec2{0, 1}},

		// Right face
		{Position: mgl32.Vec3{0.5, -0.5, 0.5}, Normal: mgl32.Vec3{1, 0, 0}, TexCoords: mgl32.Vec2{0, 0}},
		{Position: mgl32.Vec3{0.5, -0.5, -0.5}, Normal: mgl32.Vec3{1, 0, 0}, TexCoords: mgl32.Vec2{1, 0}},
		{Position: mgl32.Vec3{0.5, 0.5, -0.5}, Normal: mgl32.Vec3{1, 0, 0}, TexCoords: mgl32.Vec2{1, 1}},
		{Position: mgl32.Vec3{0.5, 0.5, 0.5}, Normal: mgl32.Vec3{1, 0, 0}, TexCoords: mgl32.Vec2{0, 1}},

		// Top face
		{Position: mgl32.Vec3{-0.5, 0.5, 0.5}, Normal: mgl32.Vec3{0, 1, 0}, TexCoords: mgl32.Vec2{0, 0}},
		{Position: mgl32.Vec3{0.5, 0.5, 0.5}, Normal: mgl32.Vec3{0, 1, 0}, TexCoords: mgl32.Vec2{1, 0}},
		{Position: mgl32.Vec3{0.5, 0.5, -0.5}, Normal: mgl32.Vec3{0, 1, 0}, TexCoords: mgl32.Vec2{1, 1}},
		{Position: mgl32.Vec3{-0.5, 0.5, -0.5}, Normal: mgl32.Vec3{0, 1, 0}, TexCoords: mgl32.Vec2{0, 1}},

		// Bottom face
		{Position: mgl32.Vec3{-0.5, -0.5, -0.5}, Normal: mgl32.Vec3{0, -1, 0}, TexCoords: mgl32.Vec2{0, 0}},
		{Position: mgl32.Vec3{0.5, -0.5, -0.5}, Normal: mgl32.Vec3{0, -1, 0}, TexCoords: mgl32.Vec2{1, 0}},
		{Position: mgl32.Vec3{0.5, -0.5, 0.5}, Normal: mgl32.Vec3{0, -1, 0}, TexCoords: mgl32.Vec2{1, 1}},
		{Position: mgl32.Vec3{-0.5, -0.5, 0.5}, Normal: mgl32.Vec3{0, -1, 0}, TexCoords: mgl32.Vec2{0, 1}},
	}

	// Индексы для куба
	indices := []uint32{
		0, 1, 2, 2, 3, 0, // Front
		4, 5, 6, 6, 7, 4, // Back
		8, 9, 10, 10, 11, 8, // Left
		12, 13, 14, 14, 15, 12, // Right
		16, 17, 18, 18, 19, 16, // Top
		20, 21, 22, 22, 23, 20, // Bottom
	}

	// Создаём меш
	mesh := Mesh{
		Vertices: vertices,
		Indices:  indices,
		Texture:  textureID,
	}

	// Создаём VAO, VBO, EBO
	setupMesh(&mesh)

	model.AddMesh(mesh)

	return model, nil
}

// setupMesh настраивает OpenGL буферы для меша
func setupMesh(mesh *Mesh) {
	// Генерируем буферы
	gl.GenVertexArrays(1, &mesh.VAO)
	gl.GenBuffers(1, &mesh.VBO)
	gl.GenBuffers(1, &mesh.EBO)

	gl.BindVertexArray(mesh.VAO)

	// Подготавливаем данные вершин
	var vertexData []float32
	for _, v := range mesh.Vertices {
		vertexData = append(vertexData,
			v.Position.X(), v.Position.Y(), v.Position.Z(),
			v.Normal.X(), v.Normal.Y(), v.Normal.Z(),
			v.TexCoords.X(), v.TexCoords.Y(),
		)
	}

	// VBO
	gl.BindBuffer(gl.ARRAY_BUFFER, mesh.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertexData)*4, gl.Ptr(vertexData), gl.STATIC_DRAW)

	// EBO
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, mesh.EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(mesh.Indices)*4, gl.Ptr(mesh.Indices), gl.STATIC_DRAW)

	// Атрибуты вершин
	stride := int32(8 * 4) // 8 floats * 4 bytes

	// Position
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Normal
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	// TexCoords
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, stride, gl.PtrOffset(6*4))
	gl.EnableVertexAttribArray(2)

	gl.BindVertexArray(0)
}
