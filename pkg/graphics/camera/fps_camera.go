package camera

import (
	"math"

	customMath "github.com/Salamander5876/AnimoEngine/pkg/core/math"
	"github.com/go-gl/mathgl/mgl32"
)

// FPSCamera камера от первого лица
type FPSCamera struct {
	Position mgl32.Vec3
	Front    mgl32.Vec3
	Up       mgl32.Vec3
	Right    mgl32.Vec3
	WorldUp  mgl32.Vec3

	Yaw   float32
	Pitch float32

	MovementSpeed    float32
	MouseSensitivity float32
	Zoom             float32
}

// NewFPSCamera создает новую FPS камеру
func NewFPSCamera(position mgl32.Vec3) *FPSCamera {
	camera := &FPSCamera{
		Position:         position,
		WorldUp:          mgl32.Vec3{0, 1, 0},
		Yaw:              -90.0, // Смотрим вперед
		Pitch:            0.0,
		MovementSpeed:    5.0,
		MouseSensitivity: 0.1,
		Zoom:             45.0,
	}

	camera.updateCameraVectors()
	return camera
}

// GetViewMatrix возвращает матрицу вида
func (c *FPSCamera) GetViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(c.Position, c.Position.Add(c.Front), c.Up)
}

// GetProjectionMatrix возвращает матрицу проекции
func (c *FPSCamera) GetProjectionMatrix(aspectRatio float32) mgl32.Mat4 {
	return mgl32.Perspective(
		customMath.DegToRad(c.Zoom),
		aspectRatio,
		0.1,
		100.0,
	)
}

// ProcessKeyboard обрабатывает движение камеры
func (c *FPSCamera) ProcessKeyboard(forward, backward, left, right bool, deltaTime float32) {
	velocity := c.MovementSpeed * deltaTime

	if forward {
		c.Position = c.Position.Add(c.Front.Mul(velocity))
	}
	if backward {
		c.Position = c.Position.Sub(c.Front.Mul(velocity))
	}
	if left {
		c.Position = c.Position.Sub(c.Right.Mul(velocity))
	}
	if right {
		c.Position = c.Position.Add(c.Right.Mul(velocity))
	}
}

// ProcessMouseMovement обрабатывает движение мыши
func (c *FPSCamera) ProcessMouseMovement(xOffset, yOffset float32, constrainPitch bool) {
	xOffset *= c.MouseSensitivity
	yOffset *= c.MouseSensitivity

	c.Yaw += xOffset
	c.Pitch += yOffset

	// Ограничиваем pitch чтобы экран не переворачивался
	if constrainPitch {
		if c.Pitch > 89.0 {
			c.Pitch = 89.0
		}
		if c.Pitch < -89.0 {
			c.Pitch = -89.0
		}
	}

	c.updateCameraVectors()
}

// ProcessMouseScroll обрабатывает прокрутку мыши (zoom)
func (c *FPSCamera) ProcessMouseScroll(yOffset float32) {
	c.Zoom -= yOffset
	if c.Zoom < 1.0 {
		c.Zoom = 1.0
	}
	if c.Zoom > 45.0 {
		c.Zoom = 45.0
	}
}

// GetRight возвращает правый вектор камеры
func (c *FPSCamera) GetRight() mgl32.Vec3 {
	return c.Right
}

// updateCameraVectors обновляет векторы камеры на основе углов Эйлера
func (c *FPSCamera) updateCameraVectors() {
	// Вычисляем новый Front вектор
	front := mgl32.Vec3{
		float32(math.Cos(float64(customMath.DegToRad(c.Yaw))) * math.Cos(float64(customMath.DegToRad(c.Pitch)))),
		float32(math.Sin(float64(customMath.DegToRad(c.Pitch)))),
		float32(math.Sin(float64(customMath.DegToRad(c.Yaw))) * math.Cos(float64(customMath.DegToRad(c.Pitch)))),
	}
	c.Front = front.Normalize()

	// Пересчитываем Right и Up векторы
	c.Right = c.Front.Cross(c.WorldUp).Normalize()
	c.Up = c.Right.Cross(c.Front).Normalize()
}
