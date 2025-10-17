package math

import "github.com/go-gl/mathgl/mgl32"

// AABB представляет axis-aligned bounding box для collision detection
type AABB struct {
	Min mgl32.Vec3 // Минимальная точка (левый нижний угол)
	Max mgl32.Vec3 // Максимальная точка (правый верхний угол)
}

// NewAABB создает новый AABB с заданными минимальной и максимальной точками
func NewAABB(min, max mgl32.Vec3) AABB {
	return AABB{
		Min: min,
		Max: max,
	}
}

// NewAABBFromCenter создает AABB из центральной точки и размеров
func NewAABBFromCenter(center mgl32.Vec3, halfExtents mgl32.Vec3) AABB {
	return AABB{
		Min: center.Sub(halfExtents),
		Max: center.Add(halfExtents),
	}
}

// Intersects проверяет пересечение с другим AABB
func (a AABB) Intersects(other AABB) bool {
	return a.Min.X() <= other.Max.X() && a.Max.X() >= other.Min.X() &&
		a.Min.Y() <= other.Max.Y() && a.Max.Y() >= other.Min.Y() &&
		a.Min.Z() <= other.Max.Z() && a.Max.Z() >= other.Min.Z()
}

// Contains проверяет, содержит ли AABB точку
func (a AABB) Contains(point mgl32.Vec3) bool {
	return point.X() >= a.Min.X() && point.X() <= a.Max.X() &&
		point.Y() >= a.Min.Y() && point.Y() <= a.Max.Y() &&
		point.Z() >= a.Min.Z() && point.Z() <= a.Max.Z()
}

// Center возвращает центр AABB
func (a AABB) Center() mgl32.Vec3 {
	return a.Min.Add(a.Max).Mul(0.5)
}

// Size возвращает размеры AABB
func (a AABB) Size() mgl32.Vec3 {
	return a.Max.Sub(a.Min)
}

// HalfExtents возвращает половинные размеры AABB
func (a AABB) HalfExtents() mgl32.Vec3 {
	return a.Size().Mul(0.5)
}

// Expand расширяет AABB на заданное значение во все стороны
func (a AABB) Expand(amount float32) AABB {
	expansion := mgl32.Vec3{amount, amount, amount}
	return AABB{
		Min: a.Min.Sub(expansion),
		Max: a.Max.Add(expansion),
	}
}

// Merge объединяет два AABB в один, содержащий оба
func (a AABB) Merge(other AABB) AABB {
	return AABB{
		Min: mgl32.Vec3{
			min(a.Min.X(), other.Min.X()),
			min(a.Min.Y(), other.Min.Y()),
			min(a.Min.Z(), other.Min.Z()),
		},
		Max: mgl32.Vec3{
			max(a.Max.X(), other.Max.X()),
			max(a.Max.Y(), other.Max.Y()),
			max(a.Max.Z(), other.Max.Z()),
		},
	}
}

// Transform применяет трансформацию к AABB
func (a AABB) Transform(matrix mgl32.Mat4) AABB {
	// Трансформируем все 8 углов и находим новые min/max
	corners := [8]mgl32.Vec3{
		a.Min,
		{a.Max.X(), a.Min.Y(), a.Min.Z()},
		{a.Min.X(), a.Max.Y(), a.Min.Z()},
		{a.Max.X(), a.Max.Y(), a.Min.Z()},
		{a.Min.X(), a.Min.Y(), a.Max.Z()},
		{a.Max.X(), a.Min.Y(), a.Max.Z()},
		{a.Min.X(), a.Max.Y(), a.Max.Z()},
		a.Max,
	}

	transformed := matrix.Mul4x1(corners[0].Vec4(1)).Vec3()
	newMin := transformed
	newMax := transformed

	for i := 1; i < 8; i++ {
		transformed = matrix.Mul4x1(corners[i].Vec4(1)).Vec3()
		newMin = mgl32.Vec3{
			min(newMin.X(), transformed.X()),
			min(newMin.Y(), transformed.Y()),
			min(newMin.Z(), transformed.Z()),
		}
		newMax = mgl32.Vec3{
			max(newMax.X(), transformed.X()),
			max(newMax.Y(), transformed.Y()),
			max(newMax.Z(), transformed.Z()),
		}
	}

	return AABB{Min: newMin, Max: newMax}
}

// min возвращает минимальное из двух float32
func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

// max возвращает максимальное из двух float32
func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
