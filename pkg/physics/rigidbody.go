package physics

import (
	"github.com/go-gl/mathgl/mgl32"
)

// RigidBodyType тип физического тела
type RigidBodyType int

const (
	Static  RigidBodyType = iota // Статичное (не двигается)
	Dynamic                       // Динамическое (подвержено физике)
	Kinematic                     // Кинематическое (управляется кодом)
)

// CollisionShape форма коллайдера
type CollisionShape int

const (
	BoxShape CollisionShape = iota
	SphereShape
	CapsuleShape
	PlaneShape
)

// RigidBody физическое тело
type RigidBody struct {
	// Трансформация
	Position mgl32.Vec3
	Rotation mgl32.Quat
	Scale    mgl32.Vec3

	// Физические свойства
	Velocity        mgl32.Vec3
	AngularVelocity mgl32.Vec3
	Mass            float32
	Restitution     float32 // Коэффициент отскока (0-1)
	Friction        float32 // Коэффициент трения (0-1)

	// Тип и форма
	Type  RigidBodyType
	Shape CollisionShape

	// Размеры (зависит от формы)
	Dimensions mgl32.Vec3 // для Box: width, height, depth; для Sphere: radius, 0, 0

	// Флаги
	UseGravity bool
	IsGrounded bool

	// Для отладки
	ID   int
	Name string
}

// NewRigidBody создает новое физическое тело
func NewRigidBody(bodyType RigidBodyType, shape CollisionShape) *RigidBody {
	return &RigidBody{
		Position:    mgl32.Vec3{0, 0, 0},
		Rotation:    mgl32.QuatIdent(),
		Scale:       mgl32.Vec3{1, 1, 1},
		Velocity:    mgl32.Vec3{0, 0, 0},
		Mass:        1.0,
		Restitution: 0.5,
		Friction:    0.5,
		Type:        bodyType,
		Shape:       shape,
		Dimensions:  mgl32.Vec3{1, 1, 1},
		UseGravity:  true,
		IsGrounded:  false,
	}
}

// ApplyForce применяет силу к телу
func (rb *RigidBody) ApplyForce(force mgl32.Vec3) {
	if rb.Type != Dynamic {
		return
	}
	// F = ma => a = F/m
	acceleration := force.Mul(1.0 / rb.Mass)
	rb.Velocity = rb.Velocity.Add(acceleration)
}

// ApplyImpulse применяет импульс к телу
func (rb *RigidBody) ApplyImpulse(impulse mgl32.Vec3) {
	if rb.Type != Dynamic {
		return
	}
	// p = mv => v = p/m
	rb.Velocity = rb.Velocity.Add(impulse.Mul(1.0 / rb.Mass))
}

// GetModelMatrix возвращает матрицу модели для рендеринга
func (rb *RigidBody) GetModelMatrix() mgl32.Mat4 {
	translation := mgl32.Translate3D(rb.Position.X(), rb.Position.Y(), rb.Position.Z())
	rotation := rb.Rotation.Mat4()
	scale := mgl32.Scale3D(rb.Scale.X(), rb.Scale.Y(), rb.Scale.Z())
	return translation.Mul4(rotation).Mul4(scale)
}
