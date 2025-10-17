package math

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// Константы для математических операций
const (
	Epsilon = 1e-6 // Малое значение для сравнения float32
	Pi      = math.Pi
	Deg2Rad = Pi / 180.0
	Rad2Deg = 180.0 / Pi
)

// Clamp ограничивает значение между min и max
func Clamp(value, min, max float32) float32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Lerp выполняет линейную интерполяцию между a и b
func Lerp(a, b, t float32) float32 {
	return a + (b-a)*t
}

// LerpVec3 выполняет линейную интерполяцию между двумя векторами
func LerpVec3(a, b mgl32.Vec3, t float32) mgl32.Vec3 {
	return mgl32.Vec3{
		Lerp(a.X(), b.X(), t),
		Lerp(a.Y(), b.Y(), t),
		Lerp(a.Z(), b.Z(), t),
	}
}

// SmoothStep выполняет плавную интерполяцию (Hermite interpolation)
func SmoothStep(edge0, edge1, x float32) float32 {
	t := Clamp((x-edge0)/(edge1-edge0), 0.0, 1.0)
	return t * t * (3.0 - 2.0*t)
}

// ApproxEqual проверяет приблизительное равенство двух float32
func ApproxEqual(a, b float32) bool {
	return math.Abs(float64(a-b)) < Epsilon
}

// Sign возвращает знак числа (-1, 0, или 1)
func Sign(x float32) float32 {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

// DegToRad конвертирует градусы в радианы
func DegToRad(degrees float32) float32 {
	return degrees * Deg2Rad
}

// RadToDeg конвертирует радианы в градусы
func RadToDeg(radians float32) float32 {
	return radians * Rad2Deg
}

// Ray представляет луч для raycasting
type Ray struct {
	Origin    mgl32.Vec3 // Начальная точка луча
	Direction mgl32.Vec3 // Направление луча (должно быть нормализовано)
}

// NewRay создает новый луч
func NewRay(origin, direction mgl32.Vec3) Ray {
	return Ray{
		Origin:    origin,
		Direction: direction.Normalize(),
	}
}

// PointAt возвращает точку на луче на заданном расстоянии от начала
func (r Ray) PointAt(distance float32) mgl32.Vec3 {
	return r.Origin.Add(r.Direction.Mul(distance))
}

// IntersectAABB проверяет пересечение луча с AABB
// Возвращает (пересекает, дистанцию до пересечения)
func (r Ray) IntersectAABB(aabb AABB) (bool, float32) {
	// Алгоритм пересечения луча и AABB (slab method)
	invDir := mgl32.Vec3{
		1.0 / r.Direction.X(),
		1.0 / r.Direction.Y(),
		1.0 / r.Direction.Z(),
	}

	t1 := (aabb.Min.X() - r.Origin.X()) * invDir.X()
	t2 := (aabb.Max.X() - r.Origin.X()) * invDir.X()
	t3 := (aabb.Min.Y() - r.Origin.Y()) * invDir.Y()
	t4 := (aabb.Max.Y() - r.Origin.Y()) * invDir.Y()
	t5 := (aabb.Min.Z() - r.Origin.Z()) * invDir.Z()
	t6 := (aabb.Max.Z() - r.Origin.Z()) * invDir.Z()

	tmin := max(max(min(t1, t2), min(t3, t4)), min(t5, t6))
	tmax := min(min(max(t1, t2), max(t3, t4)), max(t5, t6))

	// Луч не пересекает AABB
	if tmax < 0 || tmin > tmax {
		return false, 0
	}

	// Если tmin отрицателен, начало луча внутри AABB
	if tmin < 0 {
		return true, tmax
	}

	return true, tmin
}

// Plane представляет плоскость в 3D пространстве
type Plane struct {
	Normal   mgl32.Vec3 // Нормаль плоскости
	Distance float32    // Расстояние от начала координат
}

// NewPlane создает плоскость из нормали и точки на плоскости
func NewPlane(normal mgl32.Vec3, point mgl32.Vec3) Plane {
	n := normal.Normalize()
	return Plane{
		Normal:   n,
		Distance: n.Dot(point),
	}
}

// DistanceToPoint возвращает знаковое расстояние от плоскости до точки
func (p Plane) DistanceToPoint(point mgl32.Vec3) float32 {
	return p.Normal.Dot(point) - p.Distance
}

// ProjectPoint проецирует точку на плоскость
func (p Plane) ProjectPoint(point mgl32.Vec3) mgl32.Vec3 {
	distance := p.DistanceToPoint(point)
	return point.Sub(p.Normal.Mul(distance))
}

// IntersectRay возвращает точку пересечения луча с плоскостью
// Возвращает (пересекает, дистанцию)
func (p Plane) IntersectRay(ray Ray) (bool, float32) {
	denom := p.Normal.Dot(ray.Direction)

	// Луч параллелен плоскости
	if math.Abs(float64(denom)) < Epsilon {
		return false, 0
	}

	t := (p.Distance - p.Normal.Dot(ray.Origin)) / denom

	// Пересечение за началом луча
	if t < 0 {
		return false, 0
	}

	return true, t
}
