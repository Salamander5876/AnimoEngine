package graphics

import (
	"github.com/go-gl/mathgl/mgl32"
)

// GraphicsAPI интерфейс для абстракции графического API
type GraphicsAPI interface {
	// Initialize инициализирует графический API
	Initialize() error

	// Shutdown завершает работу с графическим API
	Shutdown()

	// Clear очищает экран
	Clear(r, g, b, a float32)

	// SetViewport устанавливает область отрисовки
	SetViewport(x, y, width, height int)

	// Present отображает кадр
	Present()
}

// TextureID представляет идентификатор текстуры
type TextureID uint32

// MeshID представляет идентификатор меша
type MeshID uint32

// ShaderID представляет идентификатор шейдера
type ShaderID uint32

// TextureFilter режим фильтрации текстур
type TextureFilter int

const (
	TextureFilterNearest TextureFilter = iota
	TextureFilterLinear
	TextureFilterMipmap
)

// TextureWrap режим повторения текстур
type TextureWrap int

const (
	TextureWrapRepeat TextureWrap = iota
	TextureWrapClampToEdge
	TextureWrapMirroredRepeat
)

// TextureConfig конфигурация текстуры
type TextureConfig struct {
	MinFilter TextureFilter
	MagFilter TextureFilter
	WrapS     TextureWrap
	WrapT     TextureWrap
	GenerateMipmaps bool
}

// DefaultTextureConfig возвращает конфигурацию текстуры по умолчанию
func DefaultTextureConfig() TextureConfig {
	return TextureConfig{
		MinFilter:       TextureFilterLinear,
		MagFilter:       TextureFilterLinear,
		WrapS:           TextureWrapRepeat,
		WrapT:           TextureWrapRepeat,
		GenerateMipmaps: false,
	}
}

// VertexAttribute описывает атрибут вершины
type VertexAttribute struct {
	Name       string
	Size       int32  // Количество компонентов (1, 2, 3, или 4)
	Type       uint32 // Тип данных (GL_FLOAT и т.д.)
	Normalized bool
	Stride     int32
	Offset     int
}

// Vertex представляет вершину меша
type Vertex struct {
	Position mgl32.Vec3
	Normal   mgl32.Vec3
	TexCoord mgl32.Vec2
	Color    mgl32.Vec4
}

// Mesh представляет 3D меш
type Mesh struct {
	Vertices []Vertex
	Indices  []uint32
}

// SpriteVertex вершина спрайта для батчинга
type SpriteVertex struct {
	Position mgl32.Vec3
	TexCoord mgl32.Vec2
	Color    mgl32.Vec4
}

// Material описывает материал для рендеринга
type Material struct {
	DiffuseTexture  TextureID
	SpecularTexture TextureID
	NormalTexture   TextureID
	Shininess       float32
	Color           mgl32.Vec4
}

// RenderCommand команда рендеринга
type RenderCommand struct {
	Mesh      MeshID
	Shader    ShaderID
	Texture   TextureID
	Transform mgl32.Mat4
	Material  *Material
}

// BlendMode режим смешивания
type BlendMode int

const (
	BlendModeNone BlendMode = iota
	BlendModeAlpha
	BlendModeAdditive
	BlendModeMultiply
)

// CullMode режим отсечения граней
type CullMode int

const (
	CullModeNone CullMode = iota
	CullModeBack
	CullModeFront
	CullModeFrontAndBack
)

// DepthTestMode режим теста глубины
type DepthTestMode int

const (
	DepthTestNone DepthTestMode = iota
	DepthTestLess
	DepthTestLessOrEqual
	DepthTestGreater
	DepthTestGreaterOrEqual
	DepthTestEqual
	DepthTestNotEqual
	DepthTestAlways
)

// RenderState состояние рендеринга
type RenderState struct {
	BlendMode     BlendMode
	CullMode      CullMode
	DepthTest     DepthTestMode
	DepthWrite    bool
	Wireframe     bool
	ScissorTest   bool
	ScissorX      int
	ScissorY      int
	ScissorWidth  int
	ScissorHeight int
}

// DefaultRenderState возвращает состояние рендеринга по умолчанию
func DefaultRenderState() RenderState {
	return RenderState{
		BlendMode:  BlendModeAlpha,
		CullMode:   CullModeBack,
		DepthTest:  DepthTestLess,
		DepthWrite: true,
		Wireframe:  false,
		ScissorTest: false,
	}
}

// Color представляет цвет в формате RGBA
type Color struct {
	R, G, B, A float32
}

// Предопределенные цвета
var (
	ColorWhite      = Color{1, 1, 1, 1}
	ColorBlack      = Color{0, 0, 0, 1}
	ColorRed        = Color{1, 0, 0, 1}
	ColorGreen      = Color{0, 1, 0, 1}
	ColorBlue       = Color{0, 0, 1, 1}
	ColorYellow     = Color{1, 1, 0, 1}
	ColorCyan       = Color{0, 1, 1, 1}
	ColorMagenta    = Color{1, 0, 1, 1}
	ColorTransparent = Color{0, 0, 0, 0}
)

// ToVec4 конвертирует цвет в Vec4
func (c Color) ToVec4() mgl32.Vec4 {
	return mgl32.Vec4{c.R, c.G, c.B, c.A}
}

// NewColor создает цвет из значений 0-255
func NewColor(r, g, b, a uint8) Color {
	return Color{
		R: float32(r) / 255.0,
		G: float32(g) / 255.0,
		B: float32(b) / 255.0,
		A: float32(a) / 255.0,
	}
}

// NewColorF создает цвет из значений 0-1
func NewColorF(r, g, b, a float32) Color {
	return Color{R: r, G: g, B: b, A: a}
}
