package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"

	"github.com/Salamander5876/AnimoEngine/pkg/core"
	"github.com/Salamander5876/AnimoEngine/pkg/graphics/camera"
	"github.com/Salamander5876/AnimoEngine/pkg/graphics/shader"
	"github.com/Salamander5876/AnimoEngine/pkg/physics"
	"github.com/Salamander5876/AnimoEngine/pkg/platform/input"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
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
	fluidSystem  *physics.FluidSystem

	// Рендеринг
	cubeVAO        uint32
	cubeVBO        uint32
	sphereVAO      uint32
	sphereVBO      uint32
	sphereIndexCount int32
	capsuleVAO     uint32
	capsuleVBO     uint32
	planeVAO       uint32
	planeVBO       uint32
	liquidVAO      uint32
	liquidVBO      uint32

	// UI состояние
	selectedShape physics.CollisionShape
	spawnCooldown float32

	// Камера
	firstMouse bool
	lastMouseX float64
	lastMouseY float64

	// Освещение
	flashlightEnabled  bool // Фонарик (клавиша T)
	centerLightEnabled bool // Центральный свет (клавиша Y)
	keyTPrevPressed    bool // Предыдущее состояние клавиши T
	keyYPrevPressed    bool // Предыдущее состояние клавиши Y

	// Тени
	shadowShader *shader.Shader // Шейдер для рендеринга теней
}

func main() {
	app := &PhysicsTest{
		selectedShape:      physics.BoxShape,
		firstMouse:         true,
		flashlightEnabled:  true, // Фонарик включен по умолчанию
		centerLightEnabled: true, // Центральный свет включен по умолчанию
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

	// Захватываем мышь для управления камерой
	engine.GetWindow().SetCursorMode(int(glfw.CursorDisabled))

	// Создаем камеру
	p.camera = camera.NewFPSCamera(mgl32.Vec3{0, 5, 15})

	// Создаем шейдер с освещением
	vertexShader := `
	#version 330 core
	layout (location = 0) in vec3 aPos;
	layout (location = 1) in vec3 aColor;

	out vec3 FragPos;
	out vec3 Color;

	uniform mat4 uModel;
	uniform mat4 uView;
	uniform mat4 uProjection;

	void main() {
		FragPos = vec3(uModel * vec4(aPos, 1.0));
		Color = aColor;
		gl_Position = uProjection * uView * vec4(FragPos, 1.0);
	}
	`

	fragmentShader := `
	#version 330 core
	out vec4 FragColor;

	in vec3 FragPos;
	in vec3 Color;

	uniform bool flashlightEnabled;
	uniform vec3 flashlightPos;
	uniform vec3 flashlightDir;
	uniform vec3 flashlightColor;

	uniform bool centerLightEnabled;
	uniform vec3 centerLightPos;
	uniform vec3 centerLightColor;

	uniform vec3 ambientColor;
	uniform float ambientStrength;

	void main() {
		vec3 ambient = ambientColor * ambientStrength;
		vec3 lighting = ambient;

		// Фонарик
		if (flashlightEnabled) {
			vec3 lightDir = normalize(flashlightPos - FragPos);
			float distance = length(flashlightPos - FragPos);
			float attenuation = 1.0 / (1.0 + 0.09 * distance + 0.032 * (distance * distance));
			float theta = dot(lightDir, normalize(-flashlightDir));
			float cutOff = cos(radians(12.5));
			float outerCutOff = cos(radians(17.5));
			float epsilon = cutOff - outerCutOff;
			float intensity = clamp((theta - outerCutOff) / epsilon, 0.0, 1.0);
			float diffuse = max(1.0 - distance / 20.0, 0.0);
			vec3 flashlight = flashlightColor * diffuse * attenuation * intensity * 2.0;
			lighting += flashlight;
		}

		// Центральный свет
		if (centerLightEnabled) {
			vec3 lightDir = normalize(centerLightPos - FragPos);
			float distance = length(centerLightPos - FragPos);
			float attenuation = 1.0 / (1.0 + 0.09 * distance + 0.032 * (distance * distance));
			float diffuse = max(1.0 - distance / 25.0, 0.0);
			vec3 pointLight = centerLightColor * diffuse * attenuation * 3.0;
			lighting += pointLight;
		}

		vec3 result = Color * lighting;
		FragColor = vec4(result, 1.0);
	}
	`

	var err error
	p.shader, err = shader.NewShader(vertexShader, fragmentShader)
	if err != nil {
		return err
	}

	// Создаем шейдер для теней (planar shadows)
	shadowVertexShader := `
	#version 330 core
	layout (location = 0) in vec3 aPos;

	uniform mat4 uModel;
	uniform mat4 uView;
	uniform mat4 uProjection;
	uniform vec3 uLightPos; // Позиция источника света

	void main() {
		// Проецируем вершину на плоскость Y=0.01 (чуть выше пола)
		vec4 worldPos = uModel * vec4(aPos, 1.0);

		// Вычисляем направление от источника света к вершине
		vec3 lightDir = worldPos.xyz - uLightPos;

		// Проецируем на плоскость пола (Y = 0.01)
		float t = (0.01 - uLightPos.y) / lightDir.y;
		vec3 shadowPos = uLightPos + lightDir * t;

		gl_Position = uProjection * uView * vec4(shadowPos, 1.0);
	}
	`

	shadowFragmentShader := `
	#version 330 core
	out vec4 FragColor;

	void main() {
		// Полупрозрачная чёрная тень
		FragColor = vec4(0.0, 0.0, 0.0, 0.5);
	}
	`

	p.shadowShader, err = shader.NewShader(shadowVertexShader, shadowFragmentShader)
	if err != nil {
		return err
	}

	// Создаем геометрию
	p.createCube()
	p.createSphere()
	p.createCapsule()
	p.createPlane()
	p.createLiquid()

	// Создаем физический мир
	p.physicsWorld = physics.NewPhysicsWorld()
	p.physicsWorld.GroundPlaneY = 0.0

	// Создаем систему жидкости
	p.fluidSystem = physics.NewFluidSystem()
	p.fluidSystem.Bounds = mgl32.Vec3{20, 20, 20}

	// Добавляем статичную плоскость земли
	ground := physics.NewRigidBody(physics.Static, physics.PlaneShape)
	ground.Position = mgl32.Vec3{0, 0, 0}
	ground.Dimensions = mgl32.Vec3{20, 0.1, 20}
	ground.Name = "Ground"
	p.physicsWorld.AddBody(ground)

	fmt.Println("\n=== Управление ===")
	fmt.Println("WASD - Движение камеры")
	fmt.Println("Мышь - Обзор")
	fmt.Println("1 - Выбрать КУБ (красный)")
	fmt.Println("2 - Выбрать СФЕРУ (зелёную)")
	fmt.Println("3 - Выбрать КАПСУЛУ (синюю)")
	fmt.Println("4 - Выбрать ЖИДКОСТЬ (голубую)")
	fmt.Println("ПРОБЕЛ - Создать выбранный объект")
	fmt.Println("R - Удалить все объекты")
	fmt.Println("T - Включить/Выключить ФОНАРИК")
	fmt.Println("Y - Включить/Выключить ЦЕНТРАЛЬНЫЙ СВЕТ")
	fmt.Println("ESC - Выход")
	fmt.Println("\n💡 Текущий объект: КУБ\n")

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

	// Управление мышью (обзор камеры)
	mouseX, mouseY := inputMgr.GetMousePosition()
	if p.firstMouse {
		p.lastMouseX = mouseX
		p.lastMouseY = mouseY
		p.firstMouse = false
		return // Пропускаем первый кадр чтобы избежать рывка
	}

	xOffset := mouseX - p.lastMouseX
	yOffset := p.lastMouseY - mouseY
	p.lastMouseX = mouseX
	p.lastMouseY = mouseY

	// Увеличиваем чувствительность мыши
	sensitivity := float32(0.3)
	p.camera.ProcessMouseMovement(float32(xOffset)*sensitivity, float32(yOffset)*sensitivity, true)

	// Выбор типа объекта (с проверкой чтобы не спамить)
	if inputMgr.IsKeyPressed(input.Key1) && p.selectedShape != physics.BoxShape {
		p.selectedShape = physics.BoxShape
		fmt.Println("✅ Выбран: КУБ (красный)")
	}
	if inputMgr.IsKeyPressed(input.Key2) && p.selectedShape != physics.SphereShape {
		p.selectedShape = physics.SphereShape
		fmt.Println("✅ Выбрана: СФЕРА (зелёная)")
	}
	if inputMgr.IsKeyPressed(input.Key3) && p.selectedShape != physics.CapsuleShape {
		p.selectedShape = physics.CapsuleShape
		fmt.Println("✅ Выбрана: КАПСУЛА (синяя)")
	}
	if inputMgr.IsKeyPressed(input.Key4) && p.selectedShape != physics.LiquidShape {
		p.selectedShape = physics.LiquidShape
		fmt.Println("✅ Выбрана: ЖИДКОСТЬ (голубая, мягкая)")
	}

	// Управление освещением - клавиша T (фонарик)
	keyTPressed := inputMgr.IsKeyPressed(input.KeyT)
	if keyTPressed && !p.keyTPrevPressed {
		p.flashlightEnabled = !p.flashlightEnabled
		if p.flashlightEnabled {
			fmt.Println("💡 Фонарик ВКЛЮЧЕН (клавиша T)")
		} else {
			fmt.Println("🔦 Фонарик ВЫКЛЮЧЕН (клавиша T)")
		}
	}
	p.keyTPrevPressed = keyTPressed

	// Управление освещением - клавиша Y (центральный свет)
	keyYPressed := inputMgr.IsKeyPressed(input.KeyY)
	if keyYPressed && !p.keyYPrevPressed {
		p.centerLightEnabled = !p.centerLightEnabled
		if p.centerLightEnabled {
			fmt.Println("💡 Центральный свет ВКЛЮЧЕН (клавиша Y)")
		} else {
			fmt.Println("🔦 Центральный свет ВЫКЛЮЧЕН (клавиша Y)")
		}
	}
	p.keyYPrevPressed = keyYPressed

	// Создание объекта
	p.spawnCooldown -= dt
	if inputMgr.IsKeyPressed(input.KeySpace) && p.spawnCooldown <= 0 {
		p.spawnObject()
		// Для жидкости - быстрый спавн, для остальных - нормальный
		if p.selectedShape == physics.LiquidShape {
			p.spawnCooldown = 0.05 // 50мс между частицами
		} else {
			p.spawnCooldown = 0.3 // 300мс для других объектов
		}
	}

	// Сброс всех объектов
	if inputMgr.IsKeyJustPressed(input.KeyR) {
		// Считаем сколько объектов удаляем
		count := 0
		newBodies := make([]*physics.RigidBody, 0)
		for _, body := range p.physicsWorld.Bodies {
			if body.Type == physics.Static {
				newBodies = append(newBodies, body)
			} else {
				count++
			}
		}
		p.physicsWorld.Bodies = newBodies
		if count > 0 {
			fmt.Printf("🗑️  Удалено объектов: %d\n", count)
		}
	}

	// Обновляем физику
	p.physicsWorld.Step(dt)

	// Обновляем жидкость
	p.fluidSystem.Update(dt)
}

func (p *PhysicsTest) spawnObject() {
	body := physics.NewRigidBody(physics.Dynamic, p.selectedShape)
	body.Position = p.camera.Position.Add(p.camera.Front.Mul(3))
	body.Velocity = p.camera.Front.Mul(5) // Бросаем вперед
	body.Mass = 1.0
	body.Restitution = 0.4
	body.Friction = 0.6

	var nameRu string
	switch p.selectedShape {
	case physics.BoxShape:
		body.Dimensions = mgl32.Vec3{1, 1, 1}
		body.Name = "Cube"
		nameRu = "КУБ"
	case physics.SphereShape:
		body.Dimensions = mgl32.Vec3{0.5, 0, 0}
		body.Name = "Sphere"
		nameRu = "СФЕРА"
	case physics.CapsuleShape:
		body.Dimensions = mgl32.Vec3{0.3, 1.5, 0} // radius, height, 0
		body.Name = "Capsule"
		nameRu = "КАПСУЛА"
	case physics.LiquidShape:
		// Для жидкости создаём 100 частиц за раз
		nameRu = "ЖИДКОСТЬ"
		spawnPos := p.camera.Position.Add(p.camera.Front.Mul(2))

		// Создаём 10 частиц с небольшим разбросом
		particleCount := 0
		for i := 0; i < 10; i++ {
			// Небольшой случайный разброс для естественности
			randomOffset := mgl32.Vec3{
				(rand.Float32() - 0.5) * 0.2,
				(rand.Float32() - 0.5) * 0.2,
				(rand.Float32() - 0.5) * 0.2,
			}
			particle := p.fluidSystem.AddParticle(spawnPos.Add(randomOffset))
			particle.Velocity = p.camera.Front.Mul(0.0005) // В 100 раз медленнее (почти стоит)
			particleCount++
		}

		fmt.Printf("💧 Создано частиц: %d (всего: %d)\n", particleCount, len(p.fluidSystem.Particles))
		return // Выходим раньше, не добавляем в physicsWorld
	}

	p.physicsWorld.AddBody(body)
	fmt.Printf("➕ Создан объект: %s (всего объектов: %d)\n", nameRu, len(p.physicsWorld.Bodies)-1)
}

func (p *PhysicsTest) onRender(engine *core.Engine) {
	// Получаем актуальный размер окна
	width, height := engine.GetWindow().GetSize()
	widthF, heightF := float32(width), float32(height)

	// Обновляем viewport для поддержки изменения размера окна
	gl.Viewport(0, 0, int32(width), int32(height))

	gl.ClearColor(0.1, 0.1, 0.15, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	p.shader.Use()

	// Настраиваем проекцию и вид
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), widthF/heightF, 0.1, 100.0)
	view := p.camera.GetViewMatrix()

	p.shader.SetMat4("uProjection", projection)
	p.shader.SetMat4("uView", view)

	// Устанавливаем параметры освещения
	// Ambient lighting (базовое окружающее освещение)
	p.shader.SetVec3("ambientColor", mgl32.Vec3{1.0, 1.0, 1.0}) // Белый ambient
	p.shader.SetFloat("ambientStrength", 0.3) // Увеличил для яркости

	// Фонарик (SpotLight от игрока)
	p.shader.SetBool("flashlightEnabled", p.flashlightEnabled)
	if p.flashlightEnabled {
		p.shader.SetVec3("flashlightPos", p.camera.Position)
		p.shader.SetVec3("flashlightDir", p.camera.Front)
		p.shader.SetVec3("flashlightColor", mgl32.Vec3{2.0, 2.0, 1.8}) // Яркий тёплый белый свет
	}

	// Центральный свет (PointLight в центре сцены)
	p.shader.SetBool("centerLightEnabled", p.centerLightEnabled)
	if p.centerLightEnabled {
		p.shader.SetVec3("centerLightPos", mgl32.Vec3{0, 5, 0}) // В центре, на высоте 5
		p.shader.SetVec3("centerLightColor", mgl32.Vec3{2.0, 1.8, 1.4}) // Яркий желтоватый свет
	}

	// Рисуем все физические тела
	for _, body := range p.physicsWorld.Bodies {
		model := body.GetModelMatrix()

		// Применяем размеры в зависимости от типа
		var scale mgl32.Mat4
		switch body.Shape {
		case physics.SphereShape:
			// Для сферы radius хранится в X, применяем его ко всем осям
			radius := body.Dimensions.X() * 2 // Умножаем на 2 для видимости
			scale = mgl32.Scale3D(radius, radius, radius)
		case physics.CapsuleShape:
			// Для капсулы: radius в X, height в Y
			scale = mgl32.Scale3D(body.Dimensions.X()*2, body.Dimensions.Y(), body.Dimensions.X()*2)
		case physics.LiquidShape:
			// Для жидкости - очень маленькие частицы
			radius := body.Dimensions.X() * 0.2 // Сильно уменьшили размер частиц
			scale = mgl32.Scale3D(radius, radius, radius)
		default:
			// Для остальных используем dimensions как есть
			scale = mgl32.Scale3D(body.Dimensions.X(), body.Dimensions.Y(), body.Dimensions.Z())
		}

		model = model.Mul4(scale)
		p.shader.SetMat4("uModel", model)

		switch body.Shape {
		case physics.BoxShape:
			gl.BindVertexArray(p.cubeVAO)
			gl.DrawArrays(gl.TRIANGLES, 0, 36)
		case physics.SphereShape:
			gl.BindVertexArray(p.sphereVAO)
			gl.DrawElements(gl.TRIANGLES, p.sphereIndexCount, gl.UNSIGNED_INT, gl.PtrOffset(0))
		case physics.CapsuleShape:
			gl.BindVertexArray(p.capsuleVAO)
			gl.DrawArrays(gl.TRIANGLES, 0, 36)
		case physics.PlaneShape:
			gl.BindVertexArray(p.planeVAO)
			gl.DrawArrays(gl.TRIANGLES, 0, 6)
		case physics.LiquidShape:
			gl.BindVertexArray(p.liquidVAO)
			gl.DrawArrays(gl.TRIANGLES, 0, 36)
		}
	}

	gl.BindVertexArray(0)

	// Рисуем частицы жидкости
	for _, particle := range p.fluidSystem.Particles {
		model := mgl32.Translate3D(particle.Position.X(), particle.Position.Y(), particle.Position.Z())

		// Очень маленький размер частицы
		particleSize := float32(0.1) // Уменьшил для более плавного вида
		scale := mgl32.Scale3D(particleSize, particleSize, particleSize)
		model = model.Mul4(scale)

		p.shader.SetMat4("uModel", model)

		// Рисуем как голубую сферу
		gl.BindVertexArray(p.liquidVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
	}

	gl.BindVertexArray(0)

	// ===== РЕНДЕРИМ ТЕНИ =====
	// Собираем активные источники света для теней
	var lightSources []mgl32.Vec3

	if p.centerLightEnabled {
		lightSources = append(lightSources, mgl32.Vec3{0, 5, 0}) // Центральный свет
	}
	if p.flashlightEnabled {
		lightSources = append(lightSources, p.camera.Position) // Фонарик от камеры
	}

	// Рисуем тени для каждого активного источника света
	if len(lightSources) > 0 {
		// Включаем blending для полупрозрачности теней
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		// Отключаем запись в depth buffer для теней
		gl.DepthMask(false)

		p.shadowShader.Use()
		p.shadowShader.SetMat4("uProjection", projection)
		p.shadowShader.SetMat4("uView", view)

		// Рендерим тени от каждого источника света
		for _, lightPos := range lightSources {
			p.shadowShader.SetVec3("uLightPos", lightPos)

			// Рисуем тени для всех физических объектов (кроме пола)
			for _, body := range p.physicsWorld.Bodies {
		if body.Type == physics.Static {
			continue // Не рисуем тени для пола
		}

		model := body.GetModelMatrix()

		// Применяем размеры
		var scale mgl32.Mat4
		switch body.Shape {
		case physics.SphereShape:
			radius := body.Dimensions.X() * 2
			scale = mgl32.Scale3D(radius, radius, radius)
		case physics.CapsuleShape:
			scale = mgl32.Scale3D(body.Dimensions.X()*2, body.Dimensions.Y(), body.Dimensions.X()*2)
		case physics.LiquidShape:
			radius := body.Dimensions.X() * 0.2
			scale = mgl32.Scale3D(radius, radius, radius)
		default:
			scale = mgl32.Scale3D(body.Dimensions.X(), body.Dimensions.Y(), body.Dimensions.Z())
		}

		model = model.Mul4(scale)
		p.shadowShader.SetMat4("uModel", model)

		// Рисуем тень
		switch body.Shape {
		case physics.BoxShape:
			gl.BindVertexArray(p.cubeVAO)
			gl.DrawArrays(gl.TRIANGLES, 0, 36)
		case physics.SphereShape:
			gl.BindVertexArray(p.sphereVAO)
			gl.DrawElements(gl.TRIANGLES, p.sphereIndexCount, gl.UNSIGNED_INT, nil)
		case physics.CapsuleShape:
			gl.BindVertexArray(p.capsuleVAO)
			gl.DrawArrays(gl.TRIANGLES, 0, 36)
		case physics.LiquidShape:
			gl.BindVertexArray(p.liquidVAO)
			gl.DrawArrays(gl.TRIANGLES, 0, 36)
		}
			}
		}

		// Восстанавливаем настройки OpenGL
		gl.DepthMask(true)
		gl.Disable(gl.BLEND)
		gl.BindVertexArray(0)
	}
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
	// Создаём настоящую сферу с помощью UV sphere
	var vertices []float32
	stacks := 10  // Вертикальные кольца
	slices := 20  // Горизонтальные сегменты
	radius := float32(0.5)

	// Генерируем вертексы сферы
	for i := 0; i <= stacks; i++ {
		phi := float64(i) * math.Pi / float64(stacks)

		for j := 0; j <= slices; j++ {
			theta := float64(j) * 2.0 * math.Pi / float64(slices)

			x := radius * float32(math.Sin(phi)*math.Cos(theta))
			y := radius * float32(math.Cos(phi))
			z := radius * float32(math.Sin(phi)*math.Sin(theta))

			// Позиция
			vertices = append(vertices, x, y, z)
			// Зелёный цвет (варьируется для эффекта)
			brightness := float32(0.7 + 0.3*math.Abs(math.Cos(phi)))
			vertices = append(vertices, 0.2*brightness, 0.8*brightness, 0.2*brightness)
		}
	}

	// Генерируем индексы для треугольников
	var indices []uint32
	for i := 0; i < stacks; i++ {
		for j := 0; j < slices; j++ {
			first := uint32(i*(slices+1) + j)
			second := first + uint32(slices+1)

			// Первый треугольник
			indices = append(indices, first, second, first+1)
			// Второй треугольник
			indices = append(indices, second, second+1, first+1)
		}
	}

	var ebo uint32
	gl.GenVertexArrays(1, &p.sphereVAO)
	gl.GenBuffers(1, &p.sphereVBO)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(p.sphereVAO)

	gl.BindBuffer(gl.ARRAY_BUFFER, p.sphereVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)

	// Сохраняем количество индексов для рендеринга
	p.sphereIndexCount = int32(len(indices))
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
func (p *PhysicsTest) createLiquid() {
	// Жидкость с голубым цветом (cyan/aqua)
	vertices := []float32{
		// Позиции         // Цвета (голубой)
		-0.5, -0.5, -0.5, 0.0, 0.8, 1.0,
		0.5, -0.5, -0.5, 0.0, 0.8, 1.0,
		0.5, 0.5, -0.5, 0.0, 0.8, 1.0,
		0.5, 0.5, -0.5, 0.0, 0.8, 1.0,
		-0.5, 0.5, -0.5, 0.0, 0.8, 1.0,
		-0.5, -0.5, -0.5, 0.0, 0.8, 1.0,

		-0.5, -0.5, 0.5, 0.1, 0.9, 1.0,
		0.5, -0.5, 0.5, 0.1, 0.9, 1.0,
		0.5, 0.5, 0.5, 0.1, 0.9, 1.0,
		0.5, 0.5, 0.5, 0.1, 0.9, 1.0,
		-0.5, 0.5, 0.5, 0.1, 0.9, 1.0,
		-0.5, -0.5, 0.5, 0.1, 0.9, 1.0,

		-0.5, 0.5, 0.5, 0.0, 0.7, 0.9,
		-0.5, 0.5, -0.5, 0.0, 0.7, 0.9,
		-0.5, -0.5, -0.5, 0.0, 0.7, 0.9,
		-0.5, -0.5, -0.5, 0.0, 0.7, 0.9,
		-0.5, -0.5, 0.5, 0.0, 0.7, 0.9,
		-0.5, 0.5, 0.5, 0.0, 0.7, 0.9,

		0.5, 0.5, 0.5, 0.2, 1.0, 1.0,
		0.5, 0.5, -0.5, 0.2, 1.0, 1.0,
		0.5, -0.5, -0.5, 0.2, 1.0, 1.0,
		0.5, -0.5, -0.5, 0.2, 1.0, 1.0,
		0.5, -0.5, 0.5, 0.2, 1.0, 1.0,
		0.5, 0.5, 0.5, 0.2, 1.0, 1.0,

		-0.5, -0.5, -0.5, 0.0, 0.6, 0.8,
		0.5, -0.5, -0.5, 0.0, 0.6, 0.8,
		0.5, -0.5, 0.5, 0.0, 0.6, 0.8,
		0.5, -0.5, 0.5, 0.0, 0.6, 0.8,
		-0.5, -0.5, 0.5, 0.0, 0.6, 0.8,
		-0.5, -0.5, -0.5, 0.0, 0.6, 0.8,

		-0.5, 0.5, -0.5, 0.3, 1.0, 1.0,
		0.5, 0.5, -0.5, 0.3, 1.0, 1.0,
		0.5, 0.5, 0.5, 0.3, 1.0, 1.0,
		0.5, 0.5, 0.5, 0.3, 1.0, 1.0,
		-0.5, 0.5, 0.5, 0.3, 1.0, 1.0,
		-0.5, 0.5, -0.5, 0.3, 1.0, 1.0,
	}

	gl.GenVertexArrays(1, &p.liquidVAO)
	gl.GenBuffers(1, &p.liquidVBO)

	gl.BindVertexArray(p.liquidVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, p.liquidVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}
