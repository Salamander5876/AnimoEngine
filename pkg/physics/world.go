package physics

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// PhysicsWorld физический мир
type PhysicsWorld struct {
	Gravity       mgl32.Vec3
	Bodies        []*RigidBody
	nextID        int
	GroundPlaneY  float32 // Y координата земли
	EnableDebug   bool
}

// NewPhysicsWorld создает новый физический мир
func NewPhysicsWorld() *PhysicsWorld {
	return &PhysicsWorld{
		Gravity:      mgl32.Vec3{0, -9.81, 0},
		Bodies:       make([]*RigidBody, 0),
		nextID:       0,
		GroundPlaneY: 0.0,
		EnableDebug:  false,
	}
}

// AddBody добавляет тело в мир
func (w *PhysicsWorld) AddBody(body *RigidBody) *RigidBody {
	body.ID = w.nextID
	w.nextID++
	w.Bodies = append(w.Bodies, body)
	return body
}

// RemoveBody удаляет тело из мира
func (w *PhysicsWorld) RemoveBody(body *RigidBody) {
	for i, b := range w.Bodies {
		if b.ID == body.ID {
			w.Bodies = append(w.Bodies[:i], w.Bodies[i+1:]...)
			return
		}
	}
}

// Step делает шаг симуляции
func (w *PhysicsWorld) Step(dt float32) {
	// Интегрируем физику для всех динамических тел
	for _, body := range w.Bodies {
		if body.Type != Dynamic {
			continue
		}

		// Применяем гравитацию
		if body.UseGravity {
			body.ApplyForce(w.Gravity.Mul(body.Mass * dt))
		}

		// Интегрируем скорость
		body.Position = body.Position.Add(body.Velocity.Mul(dt))

		// Применяем трение воздуха
		airResistance := float32(0.99)
		body.Velocity = body.Velocity.Mul(airResistance)

		// Интегрируем угловую скорость
		angle := body.AngularVelocity.Len() * dt
		if angle > 0.0001 {
			axis := body.AngularVelocity.Normalize()
			rotation := mgl32.QuatRotate(angle, axis)
			body.Rotation = rotation.Mul(body.Rotation).Normalize()
		}

		// Проверяем коллизию с землей
		w.checkGroundCollision(body)
	}

	// Проверяем коллизии между телами
	w.checkCollisions()
}

// checkGroundCollision проверяет столкновение с землей
func (w *PhysicsWorld) checkGroundCollision(body *RigidBody) {
	var bottomY float32

	switch body.Shape {
	case BoxShape:
		bottomY = body.Position.Y() - body.Dimensions.Y()*body.Scale.Y()/2
	case SphereShape:
		bottomY = body.Position.Y() - body.Dimensions.X()*body.Scale.X()
	case CapsuleShape:
		bottomY = body.Position.Y() - (body.Dimensions.Y()*body.Scale.Y()/2 + body.Dimensions.X()*body.Scale.X())
	}

	if bottomY <= w.GroundPlaneY {
		// Столкновение с землей
		body.IsGrounded = true

		// Корректируем позицию
		switch body.Shape {
		case BoxShape:
			body.Position[1] = w.GroundPlaneY + body.Dimensions.Y()*body.Scale.Y()/2
		case SphereShape:
			body.Position[1] = w.GroundPlaneY + body.Dimensions.X()*body.Scale.X()
		case CapsuleShape:
			body.Position[1] = w.GroundPlaneY + body.Dimensions.Y()*body.Scale.Y()/2 + body.Dimensions.X()*body.Scale.X()
		}

		// Применяем отскок
		if body.Velocity.Y() < 0 {
			body.Velocity[1] = -body.Velocity.Y() * body.Restitution

			// Если скорость мала, останавливаем
			if math.Abs(float64(body.Velocity.Y())) < 0.1 {
				body.Velocity[1] = 0
			}
		}

		// Применяем трение
		horizontalVel := mgl32.Vec3{body.Velocity.X(), 0, body.Velocity.Z()}
		if horizontalVel.Len() > 0 {
			friction := horizontalVel.Normalize().Mul(-body.Friction * 5.0)
			body.Velocity = body.Velocity.Add(friction.Mul(1.0 / 60.0)) // Предполагаем 60 FPS
		}

		// Замедляем вращение при контакте с землей
		body.AngularVelocity = body.AngularVelocity.Mul(0.95)
	} else {
		body.IsGrounded = false
	}
}

// checkCollisions проверяет столкновения между телами
func (w *PhysicsWorld) checkCollisions() {
	for i := 0; i < len(w.Bodies); i++ {
		for j := i + 1; j < len(w.Bodies); j++ {
			bodyA := w.Bodies[i]
			bodyB := w.Bodies[j]

			// Пропускаем если оба статичные
			if bodyA.Type == Static && bodyB.Type == Static {
				continue
			}

			// Простая AABB проверка
			if w.checkAABBCollision(bodyA, bodyB) {
				w.resolveCollision(bodyA, bodyB)
			}
		}
	}
}

// checkAABBCollision проверяет столкновение AABB
func (w *PhysicsWorld) checkAABBCollision(a, b *RigidBody) bool {
	// Получаем размеры AABB
	aMin, aMax := w.getAABB(a)
	bMin, bMax := w.getAABB(b)

	// Проверяем пересечение по всем осям
	return (aMin.X() <= bMax.X() && aMax.X() >= bMin.X()) &&
		(aMin.Y() <= bMax.Y() && aMax.Y() >= bMin.Y()) &&
		(aMin.Z() <= bMax.Z() && aMax.Z() >= bMin.Z())
}

// getAABB возвращает AABB для тела
func (w *PhysicsWorld) getAABB(body *RigidBody) (mgl32.Vec3, mgl32.Vec3) {
	var halfExtents mgl32.Vec3

	switch body.Shape {
	case BoxShape:
		halfExtents = mgl32.Vec3{
			body.Dimensions.X() * body.Scale.X() / 2,
			body.Dimensions.Y() * body.Scale.Y() / 2,
			body.Dimensions.Z() * body.Scale.Z() / 2,
		}
	case SphereShape:
		r := body.Dimensions.X() * body.Scale.X()
		halfExtents = mgl32.Vec3{r, r, r}
	case CapsuleShape:
		r := body.Dimensions.X() * body.Scale.X()
		h := body.Dimensions.Y() * body.Scale.Y() / 2
		halfExtents = mgl32.Vec3{r, h + r, r}
	}

	min := body.Position.Sub(halfExtents)
	max := body.Position.Add(halfExtents)
	return min, max
}

// resolveCollision разрешает столкновение
func (w *PhysicsWorld) resolveCollision(a, b *RigidBody) {
	// Вычисляем направление столкновения
	direction := b.Position.Sub(a.Position)
	distance := direction.Len()

	if distance < 0.0001 {
		return // Избегаем деления на ноль
	}

	normal := direction.Normalize()

	// Вычисляем глубину проникновения
	aMin, aMax := w.getAABB(a)
	bMin, bMax := w.getAABB(b)

	overlap := mgl32.Vec3{
		float32(math.Min(float64(aMax.X()-bMin.X()), float64(bMax.X()-aMin.X()))),
		float32(math.Min(float64(aMax.Y()-bMin.Y()), float64(bMax.Y()-aMin.Y()))),
		float32(math.Min(float64(aMax.Z()-bMin.Z()), float64(bMax.Z()-aMin.Z()))),
	}

	// Находим минимальную ось проникновения
	penetrationDepth := float32(math.Min(math.Min(float64(overlap.X()), float64(overlap.Y())), float64(overlap.Z())))

	// Разделяем тела
	separation := normal.Mul(penetrationDepth / 2)

	if a.Type == Dynamic {
		a.Position = a.Position.Sub(separation)
	}
	if b.Type == Dynamic {
		b.Position = b.Position.Add(separation)
	}

	// Вычисляем относительную скорость
	relativeVel := b.Velocity.Sub(a.Velocity)
	velAlongNormal := relativeVel.Dot(normal)

	// Не разрешаем столкновение если тела расходятся
	if velAlongNormal > 0 {
		return
	}

	// Вычисляем импульс
	restitution := float32(math.Min(float64(a.Restitution), float64(b.Restitution)))
	invMassA := float32(0.0)
	invMassB := float32(0.0)

	if a.Type == Dynamic {
		invMassA = 1.0 / a.Mass
	}
	if b.Type == Dynamic {
		invMassB = 1.0 / b.Mass
	}

	j := -(1 + restitution) * velAlongNormal
	j /= invMassA + invMassB

	impulse := normal.Mul(j)

	if a.Type == Dynamic {
		a.Velocity = a.Velocity.Sub(impulse.Mul(invMassA))
	}
	if b.Type == Dynamic {
		b.Velocity = b.Velocity.Add(impulse.Mul(invMassB))
	}

	// Применяем небольшое вращение для реалистичности
	if a.Type == Dynamic && distance > 0.1 {
		torque := normal.Cross(mgl32.Vec3{1, 0, 0}).Mul(velAlongNormal * 0.1)
		a.AngularVelocity = a.AngularVelocity.Add(torque)
	}
	if b.Type == Dynamic && distance > 0.1 {
		torque := normal.Cross(mgl32.Vec3{1, 0, 0}).Mul(-velAlongNormal * 0.1)
		b.AngularVelocity = b.AngularVelocity.Add(torque)
	}
}
