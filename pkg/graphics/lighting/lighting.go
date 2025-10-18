package lighting

import (
	"github.com/go-gl/mathgl/mgl32"
)

// LightType тип источника света
type LightType int

const (
	DirectionalLight LightType = iota // Направленный свет (солнце)
	PointLight                        // Точечный свет (лампа)
	SpotLight                         // Прожектор (фонарик)
)

// Light источник света
type Light struct {
	Type      LightType
	Position  mgl32.Vec3 // Для PointLight и SpotLight
	Direction mgl32.Vec3 // Для DirectionalLight и SpotLight
	Color     mgl32.Vec3 // RGB цвет
	Intensity float32    // Интенсивность

	// Для SpotLight
	CutOff      float32 // Угол внутреннего конуса (в градусах)
	OuterCutOff float32 // Угол внешнего конуса (в градусах)

	// Для PointLight
	Constant  float32 // Постоянное затухание
	Linear    float32 // Линейное затухание
	Quadratic float32 // Квадратичное затухание

	// Shadow mapping
	CastShadows    bool
	ShadowMapIndex int // Индекс карты теней
}

// LightingSystem система управления освещением
type LightingSystem struct {
	Lights         []*Light
	AmbientColor   mgl32.Vec3 // Цвет окружающего освещения
	AmbientStrength float32    // Сила окружающего освещения
}

// NewLightingSystem создаёт новую систему освещения
func NewLightingSystem() *LightingSystem {
	return &LightingSystem{
		Lights:          make([]*Light, 0),
		AmbientColor:    mgl32.Vec3{0.2, 0.2, 0.25}, // Слабый синеватый ambient
		AmbientStrength: 0.3,
	}
}

// AddLight добавляет источник света
func (ls *LightingSystem) AddLight(light *Light) {
	ls.Lights = append(ls.Lights, light)
}

// RemoveLight удаляет источник света
func (ls *LightingSystem) RemoveLight(light *Light) {
	for i, l := range ls.Lights {
		if l == light {
			ls.Lights = append(ls.Lights[:i], ls.Lights[i+1:]...)
			break
		}
	}
}

// NewDirectionalLight создаёт направленный свет (солнце)
func NewDirectionalLight(direction mgl32.Vec3, color mgl32.Vec3, intensity float32) *Light {
	return &Light{
		Type:        DirectionalLight,
		Direction:   direction.Normalize(),
		Color:       color,
		Intensity:   intensity,
		CastShadows: true,
	}
}

// NewPointLight создаёт точечный свет (лампа)
func NewPointLight(position mgl32.Vec3, color mgl32.Vec3, intensity float32) *Light {
	return &Light{
		Type:        PointLight,
		Position:    position,
		Color:       color,
		Intensity:   intensity,
		Constant:    1.0,
		Linear:      0.09,    // Для радиуса ~50 единиц
		Quadratic:   0.032,   // Для радиуса ~50 единиц
		CastShadows: true,
	}
}

// NewSpotLight создаёт прожектор (фонарик)
func NewSpotLight(position, direction mgl32.Vec3, color mgl32.Vec3, intensity float32, cutOff, outerCutOff float32) *Light {
	return &Light{
		Type:        SpotLight,
		Position:    position,
		Direction:   direction.Normalize(),
		Color:       color,
		Intensity:   intensity,
		CutOff:      cutOff,
		OuterCutOff: outerCutOff,
		Constant:    1.0,
		Linear:      0.09,
		Quadratic:   0.032,
		CastShadows: true,
	}
}

// CalculateLightSpaceMatrix вычисляет матрицу пространства света для shadow mapping
func (l *Light) CalculateLightSpaceMatrix() mgl32.Mat4 {
	switch l.Type {
	case DirectionalLight:
		// Ортографическая проекция для направленного света
		orthoSize := float32(20.0)
		projection := mgl32.Ortho(-orthoSize, orthoSize, -orthoSize, orthoSize, 0.1, 100.0)

		// Позиция света - в противоположном направлении
		lightPos := l.Direction.Mul(-20)
		view := mgl32.LookAtV(lightPos, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})

		return projection.Mul4(view)

	case PointLight:
		// Для точечного света используем перспективную проекцию
		projection := mgl32.Perspective(mgl32.DegToRad(90.0), 1.0, 0.1, 100.0)
		view := mgl32.LookAtV(l.Position, l.Position.Add(mgl32.Vec3{0, -1, 0}), mgl32.Vec3{0, 0, -1})
		return projection.Mul4(view)

	case SpotLight:
		// Для прожектора - перспективная проекция с узким углом
		fov := l.OuterCutOff * 2.0 // Удваиваем угол для FOV
		projection := mgl32.Perspective(mgl32.DegToRad(fov), 1.0, 0.1, 100.0)
		view := mgl32.LookAtV(l.Position, l.Position.Add(l.Direction), mgl32.Vec3{0, 1, 0})
		return projection.Mul4(view)
	}

	return mgl32.Ident4()
}
