package math

import (
	"github.com/go-gl/mathgl/mgl32"
)

// Transform представляет пространственную трансформацию объекта
type Transform struct {
	Position mgl32.Vec3 // Позиция в мировых координатах
	Rotation mgl32.Quat // Вращение как кватернион
	Scale    mgl32.Vec3 // Масштаб по каждой оси
}

// NewTransform создает новую трансформацию с дефолтными значениями
func NewTransform() Transform {
	return Transform{
		Position: mgl32.Vec3{0, 0, 0},
		Rotation: mgl32.QuatIdent(),
		Scale:    mgl32.Vec3{1, 1, 1},
	}
}

// NewTransformWithPosition создает трансформацию с заданной позицией
func NewTransformWithPosition(position mgl32.Vec3) Transform {
	return Transform{
		Position: position,
		Rotation: mgl32.QuatIdent(),
		Scale:    mgl32.Vec3{1, 1, 1},
	}
}

// Matrix возвращает матрицу трансформации 4x4
func (t *Transform) Matrix() mgl32.Mat4 {
	// Создаем матрицу трансформации: Translation * Rotation * Scale
	translation := mgl32.Translate3D(t.Position.X(), t.Position.Y(), t.Position.Z())
	rotation := t.Rotation.Mat4()
	scale := mgl32.Scale3D(t.Scale.X(), t.Scale.Y(), t.Scale.Z())

	return translation.Mul4(rotation).Mul4(scale)
}

// Translate перемещает объект на заданный вектор
func (t *Transform) Translate(delta mgl32.Vec3) {
	t.Position = t.Position.Add(delta)
}

// Rotate вращает объект на заданный угол вокруг оси
func (t *Transform) Rotate(angle float32, axis mgl32.Vec3) {
	rotation := mgl32.QuatRotate(angle, axis.Normalize())
	t.Rotation = rotation.Mul(t.Rotation).Normalize()
}

// RotateEuler вращает объект используя углы Эйлера (в радианах)
func (t *Transform) RotateEuler(pitch, yaw, roll float32) {
	qPitch := mgl32.QuatRotate(pitch, mgl32.Vec3{1, 0, 0})
	qYaw := mgl32.QuatRotate(yaw, mgl32.Vec3{0, 1, 0})
	qRoll := mgl32.QuatRotate(roll, mgl32.Vec3{0, 0, 1})

	t.Rotation = qYaw.Mul(qPitch).Mul(qRoll).Mul(t.Rotation).Normalize()
}

// SetRotationEuler устанавливает вращение используя углы Эйлера (в радианах)
func (t *Transform) SetRotationEuler(pitch, yaw, roll float32) {
	qPitch := mgl32.QuatRotate(pitch, mgl32.Vec3{1, 0, 0})
	qYaw := mgl32.QuatRotate(yaw, mgl32.Vec3{0, 1, 0})
	qRoll := mgl32.QuatRotate(roll, mgl32.Vec3{0, 0, 1})

	t.Rotation = qYaw.Mul(qPitch).Mul(qRoll).Normalize()
}

// LookAt поворачивает объект в сторону целевой точки
func (t *Transform) LookAt(target mgl32.Vec3, up mgl32.Vec3) {
	direction := target.Sub(t.Position).Normalize()

	// Избегаем вырожденного случая когда direction совпадает с up
	if direction.ApproxEqual(up) || direction.ApproxEqual(up.Mul(-1)) {
		return
	}

	right := direction.Cross(up).Normalize()
	newUp := right.Cross(direction).Normalize()

	// Создаем матрицу из базисных векторов
	mat := mgl32.Mat4{
		right.X(), right.Y(), right.Z(), 0,
		newUp.X(), newUp.Y(), newUp.Z(), 0,
		-direction.X(), -direction.Y(), -direction.Z(), 0,
		0, 0, 0, 1,
	}

	t.Rotation = mat.Mat3().Quat().Normalize()
}

// Forward возвращает вектор направления "вперед" в локальных координатах
func (t *Transform) Forward() mgl32.Vec3 {
	return t.Rotation.Rotate(mgl32.Vec3{0, 0, -1})
}

// Right возвращает вектор направления "вправо" в локальных координатах
func (t *Transform) Right() mgl32.Vec3 {
	return t.Rotation.Rotate(mgl32.Vec3{1, 0, 0})
}

// Up возвращает вектор направления "вверх" в локальных координатах
func (t *Transform) Up() mgl32.Vec3 {
	return t.Rotation.Rotate(mgl32.Vec3{0, 1, 0})
}

// Copy создает копию трансформации
func (t *Transform) Copy() Transform {
	return Transform{
		Position: t.Position,
		Rotation: t.Rotation,
		Scale:    t.Scale,
	}
}
