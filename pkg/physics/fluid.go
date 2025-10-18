package physics

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// FluidParticle частица жидкости
type FluidParticle struct {
	Position mgl32.Vec3
	Velocity mgl32.Vec3
	Force    mgl32.Vec3
	Density  float32
	Pressure float32
}

// FluidSystem система симуляции жидкости (SPH - Smoothed Particle Hydrodynamics)
type FluidSystem struct {
	Particles []*FluidParticle

	// Параметры SPH
	SmoothingRadius float32 // Радиус влияния частиц
	RestDensity     float32 // Плотность покоя
	GasConstant     float32 // Константа газа
	Viscosity       float32 // Вязкость
	Mass            float32 // Масса частицы

	// Параметры симуляции
	Gravity      mgl32.Vec3
	Bounds       mgl32.Vec3 // Границы контейнера
	Damping      float32    // Затухание при столкновении
	TimeStep     float32    // Шаг времени
}

// NewFluidSystem создает новую систему жидкости
func NewFluidSystem() *FluidSystem {
	return &FluidSystem{
		Particles:       make([]*FluidParticle, 0),
		SmoothingRadius: 0.4,     // Меньший радиус для более плотной жидкости
		RestDensity:     998.0,   // Плотность воды
		GasConstant:     0.5,     // Минимальное давление - частицы не отталкиваются
		Viscosity:       15.0,    // Очень высокая вязкость - очень медленное течение
		Mass:            0.02,
		Gravity:         mgl32.Vec3{0, -2.0, 0}, // Слабая гравитация - медленное падение
		Bounds:          mgl32.Vec3{10, 10, 10},
		Damping:         0.01,    // Почти нет отскока - частицы прилипают
		TimeStep:        0.016,
	}
}

// AddParticle добавляет частицу в систему
func (fs *FluidSystem) AddParticle(position mgl32.Vec3) *FluidParticle {
	particle := &FluidParticle{
		Position: position,
		Velocity: mgl32.Vec3{0, 0, 0},
		Force:    mgl32.Vec3{0, 0, 0},
		Density:  fs.RestDensity,
		Pressure: 0,
	}
	fs.Particles = append(fs.Particles, particle)
	return particle
}

// Update обновляет симуляцию жидкости
func (fs *FluidSystem) Update(dt float32) {
	// Вычисляем плотность и давление
	fs.computeDensityPressure()

	// Вычисляем силы
	fs.computeForces()

	// Интегрируем
	fs.integrate(dt)
}

// computeDensityPressure вычисляет плотность и давление для каждой частицы
func (fs *FluidSystem) computeDensityPressure() {
	h2 := fs.SmoothingRadius * fs.SmoothingRadius

	for _, pi := range fs.Particles {
		pi.Density = 0

		// Суммируем вклад всех соседних частиц
		for _, pj := range fs.Particles {
			diff := pj.Position.Sub(pi.Position)
			r2 := diff.Len() * diff.Len()

			if r2 < h2 {
				// Poly6 kernel
				pi.Density += fs.Mass * fs.poly6Kernel(r2)
			}
		}

		// Вычисляем давление из плотности
		pi.Pressure = fs.GasConstant * (pi.Density - fs.RestDensity)
	}
}

// computeForces вычисляет силы для каждой частицы
func (fs *FluidSystem) computeForces() {
	h := fs.SmoothingRadius

	for _, pi := range fs.Particles {
		pressureForce := mgl32.Vec3{0, 0, 0}
		viscosityForce := mgl32.Vec3{0, 0, 0}

		for _, pj := range fs.Particles {
			if pi == pj {
				continue
			}

			diff := pj.Position.Sub(pi.Position)
			r := diff.Len()

			if r < h && r > 0.0001 {
				// Сила давления (Spiky kernel gradient)
				pressureForce = pressureForce.Add(
					diff.Normalize().Mul(-fs.Mass * (pi.Pressure + pj.Pressure) / (2.0 * pj.Density) * fs.spikyGradient(r)),
				)

				// Сила вязкости (Viscosity kernel laplacian)
				viscosityForce = viscosityForce.Add(
					pj.Velocity.Sub(pi.Velocity).Mul(fs.Mass * fs.Viscosity / pj.Density * fs.viscosityLaplacian(r)),
				)
			}
		}

		// Гравитация
		gravityForce := fs.Gravity.Mul(pi.Density)

		// Суммируем все силы
		pi.Force = pressureForce.Add(viscosityForce).Add(gravityForce)
	}
}

// integrate интегрирует частицы
func (fs *FluidSystem) integrate(dt float32) {
	for _, p := range fs.Particles {
		// Обновляем скорость
		p.Velocity = p.Velocity.Add(p.Force.Mul(dt / p.Density))

		// Обновляем позицию
		p.Position = p.Position.Add(p.Velocity.Mul(dt))

		// Обрабатываем столкновения с границами
		fs.handleBoundaryCollision(p)
	}
}

// handleBoundaryCollision обрабатывает столкновение с границами
func (fs *FluidSystem) handleBoundaryCollision(p *FluidParticle) {
	epsilon := float32(0.01) // Небольшой отступ от стен

	// X границы
	if p.Position.X() < -fs.Bounds.X()/2 {
		p.Position[0] = -fs.Bounds.X()/2 + epsilon
		if p.Velocity.X() < 0 {
			p.Velocity[0] *= -fs.Damping
		}
	}
	if p.Position.X() > fs.Bounds.X()/2 {
		p.Position[0] = fs.Bounds.X()/2 - epsilon
		if p.Velocity.X() > 0 {
			p.Velocity[0] *= -fs.Damping
		}
	}

	// Y границы (пол и потолок)
	if p.Position.Y() < 0 {
		p.Position[1] = epsilon
		if p.Velocity.Y() < 0 {
			p.Velocity[1] *= -fs.Damping
		}
	}
	if p.Position.Y() > fs.Bounds.Y() {
		p.Position[1] = fs.Bounds.Y() - epsilon
		if p.Velocity.Y() > 0 {
			p.Velocity[1] *= -fs.Damping
		}
	}

	// Z границы
	if p.Position.Z() < -fs.Bounds.Z()/2 {
		p.Position[2] = -fs.Bounds.Z()/2 + epsilon
		if p.Velocity.Z() < 0 {
			p.Velocity[2] *= -fs.Damping
		}
	}
	if p.Position.Z() > fs.Bounds.Z()/2 {
		p.Position[2] = fs.Bounds.Z()/2 - epsilon
		if p.Velocity.Z() > 0 {
			p.Velocity[2] *= -fs.Damping
		}
	}
}

// poly6Kernel ядро Poly6 для вычисления плотности
func (fs *FluidSystem) poly6Kernel(r2 float32) float32 {
	h := fs.SmoothingRadius
	h2 := h * h
	h9 := h2 * h2 * h2 * h2 * h

	if r2 >= h2 {
		return 0
	}

	coeff := 315.0 / (64.0 * math.Pi * h9)
	return float32(coeff) * (h2 - r2) * (h2 - r2) * (h2 - r2)
}

// spikyGradient градиент ядра Spiky для силы давления
func (fs *FluidSystem) spikyGradient(r float32) float32 {
	h := fs.SmoothingRadius
	h6 := h * h * h * h * h * h

	if r >= h {
		return 0
	}

	coeff := -45.0 / (math.Pi * h6)
	return float32(coeff) * (h - r) * (h - r)
}

// viscosityLaplacian лапласиан ядра вязкости
func (fs *FluidSystem) viscosityLaplacian(r float32) float32 {
	h := fs.SmoothingRadius
	h6 := h * h * h * h * h * h

	if r >= h {
		return 0
	}

	coeff := 45.0 / (math.Pi * h6)
	return float32(coeff) * (h - r)
}
