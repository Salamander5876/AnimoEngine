package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/Salamander5876/AnimoEngine/pkg/core"
	"github.com/Salamander5876/AnimoEngine/pkg/graphics/camera"
	"github.com/Salamander5876/AnimoEngine/pkg/graphics/shader"
	"github.com/Salamander5876/AnimoEngine/pkg/physics"
	"github.com/Salamander5876/AnimoEngine/pkg/platform/input"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

func init() {
	runtime.LockOSThread()
}

// PhysicsTest тест физики
type PhysicsTest struct {
	engine *core.Engine
	camera *camera.FPSCamera
	shader *shader.Shader

	// Физика
	physicsWorld *physics.PhysicsWorld

	// Рендеринг
	cubeVAO    uint32
	cubeVBO    uint32
	sphereVAO  uint32
	sphereVBO  uint32
	capsuleVAO uint32
	capsuleVBO uint32
	planeVAO   uint32
	planeVBO   uint32

	// UI состояние
	selectedShape physics.CollisionShape
	spawnCooldown float32

	// Камера
	firstMouse bool
	lastMouseX float64
	lastMouseY float64
}

func main() {
	app := &PhysicsTest{
		selectedShape: physics.BoxShape,
		firstMouse:    true,
	}

	engineCfg := core.DefaultEngineConfig()
	engineCfg.WindowConfig.Title = "Physics Test - AnimoEngine"
	engineCfg.WindowConfig.Width = 1280
	engineCfg.WindowConfig.Height = 720

	engine := core.NewEngine()

	app.engine = engine

	engine.SetInitCallback(app.onInit)
	engine.SetUpdateCallback(app.onUpdate)
	engine.SetRenderCallback(app.onRender)

	if err := engine.Run(); err != nil {
		log.Fatalf("Engine error: %v", err)
	}
}

func (p *PhysicsTest) onInit(engine *core.Engine) error {
	fmt.Println("=== Physics Test ===")

	// Инициализируем OpenGL
	if err := gl.Init(); err != nil {
		return err
	}

	gl.Enable(gl.DEPTH_TEST)

	// Создаем камеру
	p.camera = camera.NewFPSCamera(mgl32.Vec3{0, 5, 15})

	// Создаем шейдер
	vertexShader := `
	#version 330 core
	layout (location = 0) in vec3 aPos;
	layout (location = 1) in vec3 aColor;

	uniform mat4 uModel;
	uniform mat4 uView;
	uniform mat4 uProjection;

	out vec3 FragColor;

	void main() {
		gl_Position = uProjection * uView * uModel * vec4(aPos, 1.0);
		FragColor = aColor;
	}
	`

	fragmentShader := `
	#version 330 core
	in vec3 FragColor;
	out vec4 color;

	void main() {
		color = vec4(FragColor, 1.0);
	}
	`

	var err error
	p.shader, err = shader.NewShader(vertexShader, fragmentShader)
	if err != nil {
		return err
	}

	// Создаем геометрию
	p.createCube()
	p.createSphere()
	p.createCapsule()
	p.createPlane()

	// Создаем физический мир
	p.physicsWorld = physics.NewPhysicsWorld()
	p.physicsWorld.GroundPlaneY = 0.0

	// Добавляем статичную плоскость земли
	ground := physics.NewRigidBody(physics.Static, physics.PlaneShape)
	ground.Position = mgl32.Vec3{0, 0, 0}
	ground.Dimensions = mgl32.Vec3{20, 0.1, 20}
	ground.Name = "Ground"
	p.physicsWorld.AddBody(ground)

	fmt.Println("\n=== Управление ===")
	fmt.Println("WASD - Движение камеры")
	fmt.Println("Мышь - Обзор")
	fmt.Println("1 - Куб")
	fmt.Println("2 - Сфера")
	fmt.Println("3 - Капсула")
	fmt.Println("SPACE - Создать объект")
	fmt.Println("R - Сбросить все объекты")
	fmt.Println("ESC - Выход\n")

	return nil
}

func (p *PhysicsTest) onUpdate(engine *core.Engine, dt float32) {
	inputMgr := engine.GetInputManager()

	// Выход
	if inputMgr.IsKeyPressed(input.KeyEscape) {
		engine.Stop()
		return
	}

	// Управление камерой
	moveSpeed := float32(5.0)
	forward := inputMgr.IsKeyPressed(input.KeyW)
	backward := inputMgr.IsKeyPressed(input.KeyS)
	left := inputMgr.IsKeyPressed(input.KeyA)
	right := inputMgr.IsKeyPressed(input.KeyD)

	p.camera.ProcessKeyboard(forward, backward, left, right, dt*moveSpeed)

	// Мышь
	mouseX, mouseY := inputMgr.GetMousePosition()
	if p.firstMouse {
		p.lastMouseX = mouseX
		p.lastMouseY = mouseY
		p.firstMouse = false
	}

	xOffset := mouseX - p.lastMouseX
	yOffset := p.lastMouseY - mouseY
	p.lastMouseX = mouseX
	p.lastMouseY = mouseY

	p.camera.ProcessMouseMovement(float32(xOffset), float32(yOffset), true)

	// Выбор типа объекта
	if inputMgr.IsKeyPressed(input.Key1) {
		p.selectedShape = physics.BoxShape
		fmt.Println("✓ Выбран: Куб")
	}
	if inputMgr.IsKeyPressed(input.Key2) {
		p.selectedShape = physics.SphereShape
		fmt.Println("✓ Выбран: Сфера")
	}
	if inputMgr.IsKeyPressed(input.Key3) {
		p.selectedShape = physics.CapsuleShape
		fmt.Println("✓ Выбран: Капсула")
	}

	// Создание объекта
	p.spawnCooldown -= dt
	if inputMgr.IsKeyPressed(input.KeySpace) && p.spawnCooldown <= 0 {
		p.spawnObject()
		p.spawnCooldown = 0.3 // Кулдаун 300мс
	}

	// Сброс всех объектов
	if inputMgr.IsKeyPressed(input.KeyR) {
		// Удаляем все динамические тела
		newBodies := make([]*physics.RigidBody, 0)
		for _, body := range p.physicsWorld.Bodies {
			if body.Type == physics.Static {
				newBodies = append(newBodies, body)
			}
		}
		p.physicsWorld.Bodies = newBodies
		fmt.Println("🔄 Все объекты удалены")
	}

	// Обновляем физику
	p.physicsWorld.Step(dt)
}

func (p *PhysicsTest) spawnObject() {
	body := physics.NewRigidBody(physics.Dynamic, p.selectedShape)
	body.Position = p.camera.Position.Add(p.camera.Front.Mul(3))
	body.Velocity = p.camera.Front.Mul(5) // Бросаем вперед
	body.Mass = 1.0
	body.Restitution = 0.4
	body.Friction = 0.6

	switch p.selectedShape {
	case physics.BoxShape:
		body.Dimensions = mgl32.Vec3{1, 1, 1}
		body.Name = "Cube"
	case physics.SphereShape:
		body.Dimensions = mgl32.Vec3{0.5, 0, 0}
		body.Name = "Sphere"
	case physics.CapsuleShape:
		body.Dimensions = mgl32.Vec3{0.3, 1.5, 0} // radius, height, 0
		body.Name = "Capsule"
	}

	p.physicsWorld.AddBody(body)
	fmt.Printf("➕ Создан: %s (ID: %d)\n", body.Name, body.ID)
}

func (p *PhysicsTest) onRender(engine *core.Engine) {
	gl.ClearColor(0.1, 0.1, 0.15, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	p.shader.Use()

	// Настраиваем проекцию и вид
	width, height := engine.GetWindow().GetSize()
	widthF, heightF := float32(width), float32(height)
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), widthF/heightF, 0.1, 100.0)
	view := p.camera.GetViewMatrix()

	p.shader.SetMat4("uProjection", projection)
	p.shader.SetMat4("uView", view)

	// Рисуем все физические тела
	for _, body := range p.physicsWorld.Bodies {
		model := body.GetModelMatrix()

		// Применяем размеры
		scale := mgl32.Scale3D(body.Dimensions.X(), body.Dimensions.Y(), body.Dimensions.Z())
		model = model.Mul4(scale)

		p.shader.SetMat4("uModel", model)

		switch body.Shape {
		case physics.BoxShape:
			gl.BindVertexArray(p.cubeVAO)
			gl.DrawArrays(gl.TRIANGLES, 0, 36)
		case physics.SphereShape:
			gl.BindVertexArray(p.sphereVAO)
			gl.DrawArrays(gl.TRIANGLES, 0, 36) // Простая сфера из куба
		case physics.CapsuleShape:
			gl.BindVertexArray(p.capsuleVAO)
			gl.DrawArrays(gl.TRIANGLES, 0, 36)
		case physics.PlaneShape:
			gl.BindVertexArray(p.planeVAO)
			gl.DrawArrays(gl.TRIANGLES, 0, 6)
		}
	}

	gl.BindVertexArray(0)
}

func (p *PhysicsTest) createCube() {
	// Куб с красным цветом
	vertices := []float32{
		// Позиции         // Цвета
		-0.5, -0.5, -0.5, 0.8, 0.2, 0.2,
		0.5, -0.5, -0.5, 0.8, 0.2, 0.2,
		0.5, 0.5, -0.5, 0.8, 0.2, 0.2,
		0.5, 0.5, -0.5, 0.8, 0.2, 0.2,
		-0.5, 0.5, -0.5, 0.8, 0.2, 0.2,
		-0.5, -0.5, -0.5, 0.8, 0.2, 0.2,

		-0.5, -0.5, 0.5, 0.9, 0.3, 0.3,
		0.5, -0.5, 0.5, 0.9, 0.3, 0.3,
		0.5, 0.5, 0.5, 0.9, 0.3, 0.3,
		0.5, 0.5, 0.5, 0.9, 0.3, 0.3,
		-0.5, 0.5, 0.5, 0.9, 0.3, 0.3,
		-0.5, -0.5, 0.5, 0.9, 0.3, 0.3,

		-0.5, 0.5, 0.5, 0.7, 0.2, 0.2,
		-0.5, 0.5, -0.5, 0.7, 0.2, 0.2,
		-0.5, -0.5, -0.5, 0.7, 0.2, 0.2,
		-0.5, -0.5, -0.5, 0.7, 0.2, 0.2,
		-0.5, -0.5, 0.5, 0.7, 0.2, 0.2,
		-0.5, 0.5, 0.5, 0.7, 0.2, 0.2,

		0.5, 0.5, 0.5, 1.0, 0.4, 0.4,
		0.5, 0.5, -0.5, 1.0, 0.4, 0.4,
		0.5, -0.5, -0.5, 1.0, 0.4, 0.4,
		0.5, -0.5, -0.5, 1.0, 0.4, 0.4,
		0.5, -0.5, 0.5, 1.0, 0.4, 0.4,
		0.5, 0.5, 0.5, 1.0, 0.4, 0.4,

		-0.5, -0.5, -0.5, 0.6, 0.15, 0.15,
		0.5, -0.5, -0.5, 0.6, 0.15, 0.15,
		0.5, -0.5, 0.5, 0.6, 0.15, 0.15,
		0.5, -0.5, 0.5, 0.6, 0.15, 0.15,
		-0.5, -0.5, 0.5, 0.6, 0.15, 0.15,
		-0.5, -0.5, -0.5, 0.6, 0.15, 0.15,

		-0.5, 0.5, -0.5, 1.0, 0.5, 0.5,
		0.5, 0.5, -0.5, 1.0, 0.5, 0.5,
		0.5, 0.5, 0.5, 1.0, 0.5, 0.5,
		0.5, 0.5, 0.5, 1.0, 0.5, 0.5,
		-0.5, 0.5, 0.5, 1.0, 0.5, 0.5,
		-0.5, 0.5, -0.5, 1.0, 0.5, 0.5,
	}

	gl.GenVertexArrays(1, &p.cubeVAO)
	gl.GenBuffers(1, &p.cubeVBO)

	gl.BindVertexArray(p.cubeVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, p.cubeVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}

func (p *PhysicsTest) createSphere() {
	// Сфера (аппроксимация кубом) с зеленым цветом
	vertices := []float32{
		-0.5, -0.5, -0.5, 0.2, 0.8, 0.2,
		0.5, -0.5, -0.5, 0.2, 0.8, 0.2,
		0.5, 0.5, -0.5, 0.2, 0.8, 0.2,
		0.5, 0.5, -0.5, 0.2, 0.8, 0.2,
		-0.5, 0.5, -0.5, 0.2, 0.8, 0.2,
		-0.5, -0.5, -0.5, 0.2, 0.8, 0.2,

		-0.5, -0.5, 0.5, 0.3, 0.9, 0.3,
		0.5, -0.5, 0.5, 0.3, 0.9, 0.3,
		0.5, 0.5, 0.5, 0.3, 0.9, 0.3,
		0.5, 0.5, 0.5, 0.3, 0.9, 0.3,
		-0.5, 0.5, 0.5, 0.3, 0.9, 0.3,
		-0.5, -0.5, 0.5, 0.3, 0.9, 0.3,

		-0.5, 0.5, 0.5, 0.2, 0.7, 0.2,
		-0.5, 0.5, -0.5, 0.2, 0.7, 0.2,
		-0.5, -0.5, -0.5, 0.2, 0.7, 0.2,
		-0.5, -0.5, -0.5, 0.2, 0.7, 0.2,
		-0.5, -0.5, 0.5, 0.2, 0.7, 0.2,
		-0.5, 0.5, 0.5, 0.2, 0.7, 0.2,

		0.5, 0.5, 0.5, 0.4, 1.0, 0.4,
		0.5, 0.5, -0.5, 0.4, 1.0, 0.4,
		0.5, -0.5, -0.5, 0.4, 1.0, 0.4,
		0.5, -0.5, -0.5, 0.4, 1.0, 0.4,
		0.5, -0.5, 0.5, 0.4, 1.0, 0.4,
		0.5, 0.5, 0.5, 0.4, 1.0, 0.4,

		-0.5, -0.5, -0.5, 0.15, 0.6, 0.15,
		0.5, -0.5, -0.5, 0.15, 0.6, 0.15,
		0.5, -0.5, 0.5, 0.15, 0.6, 0.15,
		0.5, -0.5, 0.5, 0.15, 0.6, 0.15,
		-0.5, -0.5, 0.5, 0.15, 0.6, 0.15,
		-0.5, -0.5, -0.5, 0.15, 0.6, 0.15,

		-0.5, 0.5, -0.5, 0.5, 1.0, 0.5,
		0.5, 0.5, -0.5, 0.5, 1.0, 0.5,
		0.5, 0.5, 0.5, 0.5, 1.0, 0.5,
		0.5, 0.5, 0.5, 0.5, 1.0, 0.5,
		-0.5, 0.5, 0.5, 0.5, 1.0, 0.5,
		-0.5, 0.5, -0.5, 0.5, 1.0, 0.5,
	}

	gl.GenVertexArrays(1, &p.sphereVAO)
	gl.GenBuffers(1, &p.sphereVBO)

	gl.BindVertexArray(p.sphereVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, p.sphereVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}

func (p *PhysicsTest) createCapsule() {
	// Капсула (аппроксимация кубом) с синим цветом
	vertices := []float32{
		-0.5, -0.5, -0.5, 0.2, 0.2, 0.8,
		0.5, -0.5, -0.5, 0.2, 0.2, 0.8,
		0.5, 0.5, -0.5, 0.2, 0.2, 0.8,
		0.5, 0.5, -0.5, 0.2, 0.2, 0.8,
		-0.5, 0.5, -0.5, 0.2, 0.2, 0.8,
		-0.5, -0.5, -0.5, 0.2, 0.2, 0.8,

		-0.5, -0.5, 0.5, 0.3, 0.3, 0.9,
		0.5, -0.5, 0.5, 0.3, 0.3, 0.9,
		0.5, 0.5, 0.5, 0.3, 0.3, 0.9,
		0.5, 0.5, 0.5, 0.3, 0.3, 0.9,
		-0.5, 0.5, 0.5, 0.3, 0.3, 0.9,
		-0.5, -0.5, 0.5, 0.3, 0.3, 0.9,

		-0.5, 0.5, 0.5, 0.2, 0.2, 0.7,
		-0.5, 0.5, -0.5, 0.2, 0.2, 0.7,
		-0.5, -0.5, -0.5, 0.2, 0.2, 0.7,
		-0.5, -0.5, -0.5, 0.2, 0.2, 0.7,
		-0.5, -0.5, 0.5, 0.2, 0.2, 0.7,
		-0.5, 0.5, 0.5, 0.2, 0.2, 0.7,

		0.5, 0.5, 0.5, 0.4, 0.4, 1.0,
		0.5, 0.5, -0.5, 0.4, 0.4, 1.0,
		0.5, -0.5, -0.5, 0.4, 0.4, 1.0,
		0.5, -0.5, -0.5, 0.4, 0.4, 1.0,
		0.5, -0.5, 0.5, 0.4, 0.4, 1.0,
		0.5, 0.5, 0.5, 0.4, 0.4, 1.0,

		-0.5, -0.5, -0.5, 0.15, 0.15, 0.6,
		0.5, -0.5, -0.5, 0.15, 0.15, 0.6,
		0.5, -0.5, 0.5, 0.15, 0.15, 0.6,
		0.5, -0.5, 0.5, 0.15, 0.15, 0.6,
		-0.5, -0.5, 0.5, 0.15, 0.15, 0.6,
		-0.5, -0.5, -0.5, 0.15, 0.15, 0.6,

		-0.5, 0.5, -0.5, 0.5, 0.5, 1.0,
		0.5, 0.5, -0.5, 0.5, 0.5, 1.0,
		0.5, 0.5, 0.5, 0.5, 0.5, 1.0,
		0.5, 0.5, 0.5, 0.5, 0.5, 1.0,
		-0.5, 0.5, 0.5, 0.5, 0.5, 1.0,
		-0.5, 0.5, -0.5, 0.5, 0.5, 1.0,
	}

	gl.GenVertexArrays(1, &p.capsuleVAO)
	gl.GenBuffers(1, &p.capsuleVBO)

	gl.BindVertexArray(p.capsuleVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, p.capsuleVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}

func (p *PhysicsTest) createPlane() {
	// Плоскость с серым цветом
	vertices := []float32{
		// Позиции         // Цвета
		-0.5, 0, -0.5, 0.3, 0.3, 0.3,
		0.5, 0, -0.5, 0.3, 0.3, 0.3,
		0.5, 0, 0.5, 0.3, 0.3, 0.3,

		-0.5, 0, -0.5, 0.3, 0.3, 0.3,
		0.5, 0, 0.5, 0.3, 0.3, 0.3,
		-0.5, 0, 0.5, 0.3, 0.3, 0.3,
	}

	gl.GenVertexArrays(1, &p.planeVAO)
	gl.GenBuffers(1, &p.planeVBO)

	gl.BindVertexArray(p.planeVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, p.planeVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}
