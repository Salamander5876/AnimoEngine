package model

import (
	"github.com/go-gl/mathgl/mgl32"
)

// Vertex представляет вершину модели
type Vertex struct {
	Position  mgl32.Vec3
	Normal    mgl32.Vec3
	TexCoords mgl32.Vec2
}

// Mesh представляет меш (часть модели)
type Mesh struct {
	Vertices []Vertex
	Indices  []uint32
	VAO      uint32
	VBO      uint32
	EBO      uint32
	Texture  uint32 // ID текстуры OpenGL
}

// Model представляет 3D модель
type Model struct {
	Meshes   []Mesh
	FilePath string
}

// NewModel создаёт новую модель
func NewModel() *Model {
	return &Model{
		Meshes: make([]Mesh, 0),
	}
}

// AddMesh добавляет меш к модели
func (m *Model) AddMesh(mesh Mesh) {
	m.Meshes = append(m.Meshes, mesh)
}