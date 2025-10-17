package main

import (
	"fmt"
	"log"
	"math"
	"runtime"
	"time"

	"github.com/Salamander5876/AnimoEngine/pkg/core"
	customMath "github.com/Salamander5876/AnimoEngine/pkg/core/math"
	"github.com/Salamander5876/AnimoEngine/pkg/graphics/camera"
	"github.com/Salamander5876/AnimoEngine/pkg/graphics/shader"
	"github.com/Salamander5876/AnimoEngine/pkg/graphics/ui"
	"github.com/Salamander5876/AnimoEngine/pkg/platform/input"
	"github.com/Salamander5876/AnimoEngine/pkg/platform/window"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// init блокирует главный поток для OpenGL
func init() {
	runtime.LockOSThread()
}

// BulletTracer трассер пули для визуализации выстрела
type BulletTracer struct {
	start    mgl32.Vec3
	end      mgl32.Vec3
	lifetime float32
	maxLife  float32
}

// DestructibleObject разрушаемый объект
type DestructibleObject struct {
	position mgl32.Vec3
	health   int
	maxHP    int
	size     mgl32.Vec3
}

// Debris осколки от разрушенного объекта
type Debris struct {
	position mgl32.Vec3
	velocity mgl32.Vec3
	rotation float32
	lifetime float32
	size     float32
}

// DoomGame игра в стиле Doom
type DoomGame struct {
	engine *core.Engine
	camera *camera.FPSCamera
	shader *shader.Shader

	// Геометрия уровня
	wallVAO     uint32
	wallVBO     uint32
	floorVAO    uint32
	floorVBO    uint32
	enemyVAO    uint32
	enemyVBO    uint32

	// Позиции врагов (красные кубы)
	enemyPositions []mgl32.Vec3
	enemiesKilled  int

	// Состояние мыши
	firstMouse bool
	lastMouseX float64
	lastMouseY float64

	// Стрельба
	canShoot      bool
	shootCooldown float32
	bulletTracers []BulletTracer // Активные трассеры пуль

	// Патроны
	currentAmmo int
	maxAmmo     int
	clipSize    int
	isReloading bool
	reloadTime  float32

	// Физика игрока
	playerVelocityY float32 // Вертикальная скорость
	isGrounded      bool     // На земле ли игрок
	playerHeight    float32  // Высота камеры над землей

	// Здоровье игрока
	playerHealth    int
	maxHealth       int
	damageCooldown  float32 // Кулдаун получения урона
	canTakeDamage   bool
	isDead          bool

	// UI
	uiRenderer *ui.UIRenderer
	gunRecoil  float32 // Анимация отдачи пистолета

	// Геометрия для трассеров
	lineVAO uint32
	lineVBO uint32

	// Разрушаемые объекты
	destructibleObjects []DestructibleObject
	debris              []Debris
	boxVAO              uint32
	boxVBO              uint32
}

func main() {
	game := &DoomGame{
		firstMouse:      true,
		canShoot:        true,
		playerHeight:    1.6,
		isGrounded:      true,
		playerHealth:    100,
		maxHealth:       100,
		canTakeDamage:   true,
		isDead:          false,
		currentAmmo:     12,
		maxAmmo:         60,
		clipSize:        12,
		isReloading:     false,
		enemyPositions: []mgl32.Vec3{
			{5, 0.5, -5},
			{-5, 0.5, -5},
			{5, 0.5, 5},
			{-5, 0.5, 5},
			{0, 0.5, -8},
			{8, 0.5, 0},
			{-8, 0.5, 0},
		},
		destructibleObjects: []DestructibleObject{
			{position: mgl32.Vec3{3, 0.5, 0}, health: 3, maxHP: 3, size: mgl32.Vec3{1, 1, 1}},
			{position: mgl32.Vec3{-3, 0.5, 0}, health: 3, maxHP: 3, size: mgl32.Vec3{1, 1, 1}},
			{position: mgl32.Vec3{0, 0.5, 3}, health: 3, maxHP: 3, size: mgl32.Vec3{1, 1, 1}},
			{position: mgl32.Vec3{0, 0.5, -3}, health: 3, maxHP: 3, size: mgl32.Vec3{1, 1, 1}},
			{position: mgl32.Vec3{6, 0.5, 6}, health: 3, maxHP: 3, size: mgl32.Vec3{1, 1, 1}},
		},
	}

	config := core.DefaultEngineConfig()
	config.WindowConfig.Title = "Doom-like Game - AnimoEngine"
	config.WindowConfig.Width = 1280
	config.WindowConfig.Height = 720
	game.engine = core.NewEngineWithConfig(config)

	game.engine.SetInitCallback(game.onInit)
	game.engine.SetUpdateCallback(game.onUpdate)
	game.engine.SetRenderCallback(game.onRender)

	if err := game.engine.Run(); err != nil {
		log.Fatalf("Engine error: %v", err)
	}
}

func (g *DoomGame) onInit(engine *core.Engine) error {
	fmt.Println("=== Doom-like Game ===")

	// Инициализируем OpenGL
	if err := gl.Init(); err != nil {
		return err
	}

	// Показываем логотип
	splash, err := core.NewSplashScreen("logo.png", 2*time.Second)
	if err != nil {
		fmt.Printf("Не удалось загрузить логотип: %v\n", err)
	} else {
		splash.Show(engine)
		splash.Cleanup()
	}

	// Создаем камеру
	g.camera = camera.NewFPSCamera(mgl32.Vec3{0, 1.6, 3})

	// Создаем шейдер
	vertexShader := `
	#version 330 core
	layout (location = 0) in vec3 aPosition;
	layout (location = 1) in vec3 aColor;

	out vec3 FragColor;

	uniform mat4 uModel;
	uniform mat4 uView;
	uniform mat4 uProjection;

	void main() {
		FragColor = aColor;
		gl_Position = uProjection * uView * uModel * vec4(aPosition, 1.0);
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

	g.shader, err = shader.NewShader(vertexShader, fragmentShader)
	if err != nil {
		return err
	}

	// Создаем геометрию
	g.createWalls()
	g.createFloor()
	g.createEnemyCube()
	g.createLineVAO()
	g.createBox()

	// Создаем UI рендерер
	g.uiRenderer, err = ui.NewUIRenderer()
	if err != nil {
		return err
	}
	width, height := engine.GetWindow().GetSize()
	g.uiRenderer.SetProjection(float32(width), float32(height))

	// Настройки OpenGL
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(0.1, 0.1, 0.15, 1.0)

	// Захватываем курсор для FPS
	engine.GetWindow().SetCursorMode(window.CursorDisabled)

	fmt.Println("\n=== Управление ===")
	fmt.Println("WASD - Движение")
	fmt.Println("Пробел - Прыжок")
	fmt.Println("Мышь - Обзор")
	fmt.Println("ЛКМ - Стрельба")
	fmt.Println("R - Перезарядка")
	fmt.Println("ESC - Выход")
	fmt.Printf("\nЗдоровье: %d/%d\n", g.playerHealth, g.maxHealth)
	fmt.Printf("Патроны: %d/%d\n", g.currentAmmo, g.maxAmmo)
	fmt.Printf("Убей всех врагов! Осталось: %d\n", len(g.enemyPositions))

	return nil
}

func (g *DoomGame) createWalls() {
	// Создаем куб для стен (серый цвет)
	vertices := []float32{
		// Позиции         // Цвета (серый)
		-0.5, -0.5, -0.5,  0.5, 0.5, 0.5,
		0.5, -0.5, -0.5,   0.5, 0.5, 0.5,
		0.5, 0.5, -0.5,    0.5, 0.5, 0.5,
		0.5, 0.5, -0.5,    0.5, 0.5, 0.5,
		-0.5, 0.5, -0.5,   0.5, 0.5, 0.5,
		-0.5, -0.5, -0.5,  0.5, 0.5, 0.5,

		-0.5, -0.5, 0.5,   0.5, 0.5, 0.5,
		0.5, -0.5, 0.5,    0.5, 0.5, 0.5,
		0.5, 0.5, 0.5,     0.5, 0.5, 0.5,
		0.5, 0.5, 0.5,     0.5, 0.5, 0.5,
		-0.5, 0.5, 0.5,    0.5, 0.5, 0.5,
		-0.5, -0.5, 0.5,   0.5, 0.5, 0.5,

		-0.5, 0.5, 0.5,    0.5, 0.5, 0.5,
		-0.5, 0.5, -0.5,   0.5, 0.5, 0.5,
		-0.5, -0.5, -0.5,  0.5, 0.5, 0.5,
		-0.5, -0.5, -0.5,  0.5, 0.5, 0.5,
		-0.5, -0.5, 0.5,   0.5, 0.5, 0.5,
		-0.5, 0.5, 0.5,    0.5, 0.5, 0.5,

		0.5, 0.5, 0.5,     0.5, 0.5, 0.5,
		0.5, 0.5, -0.5,    0.5, 0.5, 0.5,
		0.5, -0.5, -0.5,   0.5, 0.5, 0.5,
		0.5, -0.5, -0.5,   0.5, 0.5, 0.5,
		0.5, -0.5, 0.5,    0.5, 0.5, 0.5,
		0.5, 0.5, 0.5,     0.5, 0.5, 0.5,

		-0.5, -0.5, -0.5,  0.5, 0.5, 0.5,
		0.5, -0.5, -0.5,   0.5, 0.5, 0.5,
		0.5, -0.5, 0.5,    0.5, 0.5, 0.5,
		0.5, -0.5, 0.5,    0.5, 0.5, 0.5,
		-0.5, -0.5, 0.5,   0.5, 0.5, 0.5,
		-0.5, -0.5, -0.5,  0.5, 0.5, 0.5,

		-0.5, 0.5, -0.5,   0.5, 0.5, 0.5,
		0.5, 0.5, -0.5,    0.5, 0.5, 0.5,
		0.5, 0.5, 0.5,     0.5, 0.5, 0.5,
		0.5, 0.5, 0.5,     0.5, 0.5, 0.5,
		-0.5, 0.5, 0.5,    0.5, 0.5, 0.5,
		-0.5, 0.5, -0.5,   0.5, 0.5, 0.5,
	}

	gl.GenVertexArrays(1, &g.wallVAO)
	gl.GenBuffers(1, &g.wallVBO)

	gl.BindVertexArray(g.wallVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, g.wallVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}

func (g *DoomGame) createFloor() {
	// Пол (темно-зеленый)
	vertices := []float32{
		// Позиции         // Цвета
		-20, 0, -20,  0.1, 0.3, 0.1,
		20, 0, -20,   0.1, 0.3, 0.1,
		20, 0, 20,    0.1, 0.3, 0.1,

		20, 0, 20,    0.1, 0.3, 0.1,
		-20, 0, 20,   0.1, 0.3, 0.1,
		-20, 0, -20,  0.1, 0.3, 0.1,
	}

	gl.GenVertexArrays(1, &g.floorVAO)
	gl.GenBuffers(1, &g.floorVBO)

	gl.BindVertexArray(g.floorVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, g.floorVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}

func (g *DoomGame) createEnemyCube() {
	// Враг (красный куб)
	vertices := []float32{
		// Позиции         // Цвета (красный)
		-0.5, -0.5, -0.5,  1.0, 0.0, 0.0,
		0.5, -0.5, -0.5,   1.0, 0.0, 0.0,
		0.5, 0.5, -0.5,    1.0, 0.0, 0.0,
		0.5, 0.5, -0.5,    1.0, 0.0, 0.0,
		-0.5, 0.5, -0.5,   1.0, 0.0, 0.0,
		-0.5, -0.5, -0.5,  1.0, 0.0, 0.0,

		-0.5, -0.5, 0.5,   0.8, 0.0, 0.0,
		0.5, -0.5, 0.5,    0.8, 0.0, 0.0,
		0.5, 0.5, 0.5,     0.8, 0.0, 0.0,
		0.5, 0.5, 0.5,     0.8, 0.0, 0.0,
		-0.5, 0.5, 0.5,    0.8, 0.0, 0.0,
		-0.5, -0.5, 0.5,   0.8, 0.0, 0.0,

		-0.5, 0.5, 0.5,    0.9, 0.0, 0.0,
		-0.5, 0.5, -0.5,   0.9, 0.0, 0.0,
		-0.5, -0.5, -0.5,  0.9, 0.0, 0.0,
		-0.5, -0.5, -0.5,  0.9, 0.0, 0.0,
		-0.5, -0.5, 0.5,   0.9, 0.0, 0.0,
		-0.5, 0.5, 0.5,    0.9, 0.0, 0.0,

		0.5, 0.5, 0.5,     0.9, 0.0, 0.0,
		0.5, 0.5, -0.5,    0.9, 0.0, 0.0,
		0.5, -0.5, -0.5,   0.9, 0.0, 0.0,
		0.5, -0.5, -0.5,   0.9, 0.0, 0.0,
		0.5, -0.5, 0.5,    0.9, 0.0, 0.0,
		0.5, 0.5, 0.5,     0.9, 0.0, 0.0,

		-0.5, -0.5, -0.5,  0.7, 0.0, 0.0,
		0.5, -0.5, -0.5,   0.7, 0.0, 0.0,
		0.5, -0.5, 0.5,    0.7, 0.0, 0.0,
		0.5, -0.5, 0.5,    0.7, 0.0, 0.0,
		-0.5, -0.5, 0.5,   0.7, 0.0, 0.0,
		-0.5, -0.5, -0.5,  0.7, 0.0, 0.0,

		-0.5, 0.5, -0.5,   1.0, 0.1, 0.1,
		0.5, 0.5, -0.5,    1.0, 0.1, 0.1,
		0.5, 0.5, 0.5,     1.0, 0.1, 0.1,
		0.5, 0.5, 0.5,     1.0, 0.1, 0.1,
		-0.5, 0.5, 0.5,    1.0, 0.1, 0.1,
		-0.5, 0.5, -0.5,   1.0, 0.1, 0.1,
	}

	gl.GenVertexArrays(1, &g.enemyVAO)
	gl.GenBuffers(1, &g.enemyVBO)

	gl.BindVertexArray(g.enemyVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, g.enemyVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}

func (g *DoomGame) createLineVAO() {
	// Создаем VAO и VBO для динамической отрисовки линий
	gl.GenVertexArrays(1, &g.lineVAO)
	gl.GenBuffers(1, &g.lineVBO)

	gl.BindVertexArray(g.lineVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, g.lineVBO)

	// Позиция (3 float) + Цвет (3 float)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}

func (g *DoomGame) createBox() {
	// Создаем ящик (коричневый цвет)
	vertices := []float32{
		// Позиции         // Цвета (коричневый)
		-0.5, -0.5, -0.5, 0.6, 0.4, 0.2,
		0.5, -0.5, -0.5, 0.6, 0.4, 0.2,
		0.5, 0.5, -0.5, 0.6, 0.4, 0.2,
		0.5, 0.5, -0.5, 0.6, 0.4, 0.2,
		-0.5, 0.5, -0.5, 0.6, 0.4, 0.2,
		-0.5, -0.5, -0.5, 0.6, 0.4, 0.2,

		-0.5, -0.5, 0.5, 0.6, 0.4, 0.2,
		0.5, -0.5, 0.5, 0.6, 0.4, 0.2,
		0.5, 0.5, 0.5, 0.6, 0.4, 0.2,
		0.5, 0.5, 0.5, 0.6, 0.4, 0.2,
		-0.5, 0.5, 0.5, 0.6, 0.4, 0.2,
		-0.5, -0.5, 0.5, 0.6, 0.4, 0.2,

		-0.5, 0.5, 0.5, 0.6, 0.4, 0.2,
		-0.5, 0.5, -0.5, 0.6, 0.4, 0.2,
		-0.5, -0.5, -0.5, 0.6, 0.4, 0.2,
		-0.5, -0.5, -0.5, 0.6, 0.4, 0.2,
		-0.5, -0.5, 0.5, 0.6, 0.4, 0.2,
		-0.5, 0.5, 0.5, 0.6, 0.4, 0.2,

		0.5, 0.5, 0.5, 0.6, 0.4, 0.2,
		0.5, 0.5, -0.5, 0.6, 0.4, 0.2,
		0.5, -0.5, -0.5, 0.6, 0.4, 0.2,
		0.5, -0.5, -0.5, 0.6, 0.4, 0.2,
		0.5, -0.5, 0.5, 0.6, 0.4, 0.2,
		0.5, 0.5, 0.5, 0.6, 0.4, 0.2,

		-0.5, -0.5, -0.5, 0.6, 0.4, 0.2,
		0.5, -0.5, -0.5, 0.6, 0.4, 0.2,
		0.5, -0.5, 0.5, 0.6, 0.4, 0.2,
		0.5, -0.5, 0.5, 0.6, 0.4, 0.2,
		-0.5, -0.5, 0.5, 0.6, 0.4, 0.2,
		-0.5, -0.5, -0.5, 0.6, 0.4, 0.2,

		-0.5, 0.5, -0.5, 0.6, 0.4, 0.2,
		0.5, 0.5, -0.5, 0.6, 0.4, 0.2,
		0.5, 0.5, 0.5, 0.6, 0.4, 0.2,
		0.5, 0.5, 0.5, 0.6, 0.4, 0.2,
		-0.5, 0.5, 0.5, 0.6, 0.4, 0.2,
		-0.5, 0.5, -0.5, 0.6, 0.4, 0.2,
	}

	gl.GenVertexArrays(1, &g.boxVAO)
	gl.GenBuffers(1, &g.boxVBO)

	gl.BindVertexArray(g.boxVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, g.boxVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}

func (g *DoomGame) onUpdate(engine *core.Engine, dt float32) {
	if g.isDead {
		return
	}

	inputMgr := engine.GetInputManager()

	// Выход (используем IsKeyPressed вместо IsKeyJustPressed)
	if inputMgr.IsKeyPressed(input.KeyEscape) {
		engine.Stop()
		return
	}

	// === ФИЗИКА ГРАВИТАЦИИ ===
	const gravity = -15.0 // Ускорение гравитации
	const groundLevel = 1.6 // Высота камеры над землей

	// Применяем гравитацию если не на земле
	if !g.isGrounded {
		g.playerVelocityY += gravity * dt
		g.camera.Position = g.camera.Position.Add(mgl32.Vec3{0, g.playerVelocityY * dt, 0})
	}

	// Проверка на касание земли
	if g.camera.Position.Y() <= groundLevel {
		g.camera.Position[1] = groundLevel
		g.playerVelocityY = 0
		g.isGrounded = true
	} else {
		g.isGrounded = false
	}

	// === ПРЫЖОК ===
	if inputMgr.IsKeyPressed(input.KeySpace) && g.isGrounded {
		g.playerVelocityY = 7.0 // Скорость прыжка
		g.isGrounded = false
	}

	// === ДВИЖЕНИЕ С КОЛЛИЗИЯМИ ===
	forward := inputMgr.IsKeyPressed(input.KeyW)
	backward := inputMgr.IsKeyPressed(input.KeyS)
	left := inputMgr.IsKeyPressed(input.KeyA)
	right := inputMgr.IsKeyPressed(input.KeyD)

	// Пробуем переместиться
	g.camera.ProcessKeyboard(forward, backward, left, right, dt)

	// Проверяем коллизии со стенами (периметр арены)
	arenaSize := float32(10.0)
	playerRadius := float32(0.5)

	if g.camera.Position.X() > arenaSize-playerRadius {
		g.camera.Position[0] = arenaSize - playerRadius
	}
	if g.camera.Position.X() < -arenaSize+playerRadius {
		g.camera.Position[0] = -arenaSize + playerRadius
	}
	if g.camera.Position.Z() > arenaSize-playerRadius {
		g.camera.Position[2] = arenaSize - playerRadius
	}
	if g.camera.Position.Z() < -arenaSize+playerRadius {
		g.camera.Position[2] = -arenaSize + playerRadius
	}

	// === КОЛЛИЗИИ С ЯЩИКАМИ ===
	for _, box := range g.destructibleObjects {
		// AABB коллизия игрока с ящиком
		boxMin := box.position.Sub(box.size.Mul(0.5))
		boxMax := box.position.Add(box.size.Mul(0.5))

		playerMin := g.camera.Position.Sub(mgl32.Vec3{playerRadius, 0, playerRadius})
		playerMax := g.camera.Position.Add(mgl32.Vec3{playerRadius, playerRadius * 2, playerRadius})

		// Проверка пересечения AABB
		if playerMax.X() > boxMin.X() && playerMin.X() < boxMax.X() &&
			playerMax.Y() > boxMin.Y() && playerMin.Y() < boxMax.Y() &&
			playerMax.Z() > boxMin.Z() && playerMin.Z() < boxMax.Z() {

			// Вычисляем направление выталкивания (по наименьшей проникающей оси)
			overlapX := mgl32.Min(playerMax.X()-boxMin.X(), boxMax.X()-playerMin.X())
			overlapZ := mgl32.Min(playerMax.Z()-boxMin.Z(), boxMax.Z()-playerMin.Z())

			if overlapX < overlapZ {
				// Выталкиваем по X
				if g.camera.Position.X() < box.position.X() {
					g.camera.Position[0] -= overlapX
				} else {
					g.camera.Position[0] += overlapX
				}
			} else {
				// Выталкиваем по Z
				if g.camera.Position.Z() < box.position.Z() {
					g.camera.Position[2] -= overlapZ
				} else {
					g.camera.Position[2] += overlapZ
				}
			}
		}
	}

	// === ОБНОВЛЕНИЕ ОСКОЛКОВ ===
	for i := len(g.debris) - 1; i >= 0; i-- {
		g.debris[i].lifetime -= dt
		if g.debris[i].lifetime <= 0 {
			g.debris = append(g.debris[:i], g.debris[i+1:]...)
			continue
		}

		// Физика осколков (гравитация + движение)
		g.debris[i].velocity[1] += -9.8 * dt
		g.debris[i].position = g.debris[i].position.Add(g.debris[i].velocity.Mul(dt))
		g.debris[i].rotation += dt * 5

		// Удаляем если упали через пол
		if g.debris[i].position.Y() < -2 {
			g.debris = append(g.debris[:i], g.debris[i+1:]...)
		}
	}

	// === ОБРАБОТКА МЫШИ ===
	mouseX, mouseY := inputMgr.GetMousePosition()
	if g.firstMouse {
		g.lastMouseX = mouseX
		g.lastMouseY = mouseY
		g.firstMouse = false
	}

	xOffset := mouseX - g.lastMouseX
	yOffset := g.lastMouseY - mouseY
	g.lastMouseX = mouseX
	g.lastMouseY = mouseY

	g.camera.ProcessMouseMovement(float32(xOffset), float32(yOffset), true)

	// === ПЕРЕЗАРЯДКА ===
	if inputMgr.IsKeyPressed(input.KeyR) && !g.isReloading && g.currentAmmo < g.clipSize && g.maxAmmo > 0 {
		g.isReloading = true
		g.reloadTime = 2.0 // 2 секунды на перезарядку
		fmt.Println("🔄 Перезарядка...")
	}

	if g.isReloading {
		g.reloadTime -= dt
		if g.reloadTime <= 0 {
			// Перезарядка завершена
			ammoNeeded := g.clipSize - g.currentAmmo
			if ammoNeeded > g.maxAmmo {
				ammoNeeded = g.maxAmmo
			}
			g.currentAmmo += ammoNeeded
			g.maxAmmo -= ammoNeeded
			g.isReloading = false
			fmt.Printf("✅ Перезарядка завершена! Патроны: %d/%d\n", g.currentAmmo, g.maxAmmo)
		}
	}

	// === КУЛДАУН СТРЕЛЬБЫ ===
	if !g.canShoot {
		g.shootCooldown -= dt
		if g.shootCooldown <= 0 {
			g.canShoot = true
		}
	}

	// === СТРЕЛЬБА ===
	if inputMgr.IsMouseButtonPressed(input.MouseButtonLeft) && g.canShoot && !g.isReloading {
		if g.currentAmmo > 0 {
			g.shoot()
			g.currentAmmo--
			g.canShoot = false
			g.shootCooldown = 0.2 // Быстрее стрельба

			// Автоматическая перезарядка если закончились патроны
			if g.currentAmmo == 0 && g.maxAmmo > 0 {
				fmt.Println("⚠️ Магазин пуст!")
			}
		} else {
			// Щелчок пустого магазина
			fmt.Println("*клик* - Нет патронов! Нажми R для перезарядки")
			g.canShoot = false
			g.shootCooldown = 0.3
		}
	}

	// === AI ВРАГОВ ===
	const enemySpeed = 2.0
	const enemyDamage = 10
	const damageRange = 1.5

	for i := range g.enemyPositions {
		// Враги движутся к игроку
		toPlayer := g.camera.Position.Sub(g.enemyPositions[i])
		toPlayer[1] = 0 // Не учитываем высоту
		distance := toPlayer.Len()

		if distance > 0.1 {
			direction := toPlayer.Normalize()
			g.enemyPositions[i] = g.enemyPositions[i].Add(direction.Mul(enemySpeed * dt))
		}

		// Проверка столкновения с игроком
		if distance < damageRange && g.canTakeDamage {
			g.playerHealth -= enemyDamage
			g.canTakeDamage = false
			g.damageCooldown = 1.0 // Урон раз в секунду
			fmt.Printf("💔 Получен урон! Здоровье: %d/%d\n", g.playerHealth, g.maxHealth)

			if g.playerHealth <= 0 {
				g.isDead = true
				fmt.Println("\n💀 GAME OVER! Вы мертвы!")
				fmt.Println("Нажмите ESC для выхода")
				return
			}
		}
	}

	// === КУЛДАУН УРОНА ===
	if !g.canTakeDamage {
		g.damageCooldown -= dt
		if g.damageCooldown <= 0 {
			g.canTakeDamage = true
		}
	}

	// === АНИМАЦИЯ ОТДАЧИ ПИСТОЛЕТА ===
	if g.gunRecoil > 0 {
		g.gunRecoil -= dt * 10.0 // Быстро возвращаем пистолет на место
		if g.gunRecoil < 0 {
			g.gunRecoil = 0
		}
	}

	// === ОБНОВЛЕНИЕ ТРАССЕРОВ ПУЛЬ ===
	for i := len(g.bulletTracers) - 1; i >= 0; i-- {
		g.bulletTracers[i].lifetime -= dt
		if g.bulletTracers[i].lifetime <= 0 {
			// Удаляем истекший трассер
			g.bulletTracers = append(g.bulletTracers[:i], g.bulletTracers[i+1:]...)
		}
	}
}

func (g *DoomGame) shoot() {
	// Анимация отдачи
	g.gunRecoil = 0.2

	// Простой рейкаст от камеры вперед
	ray := customMath.NewRay(g.camera.Position, g.camera.Front)

	// Конечная точка трассера (по умолчанию - промах)
	tracerEnd := g.camera.Position.Add(g.camera.Front.Mul(50.0))
	closestDist := float32(50.0)
	hitSomething := false

	// Проверяем попадание по ящикам (сначала, чтобы они блокировали выстрелы)
	for i := len(g.destructibleObjects) - 1; i >= 0; i-- {
		box := &g.destructibleObjects[i]
		boxAABB := customMath.NewAABBFromCenter(box.position, box.size)

		if hit, distance := ray.IntersectAABB(boxAABB); hit && distance < closestDist {
			// Попали в ящик!
			tracerEnd = g.camera.Position.Add(g.camera.Front.Mul(distance))
			closestDist = distance
			hitSomething = true

			// Наносим урон ящику
			box.health--
			fmt.Printf("📦 Попадание по ящику! HP: %d/%d\n", box.health, box.maxHP)

			if box.health <= 0 {
				// Ящик разрушен! Создаем осколки
				fmt.Println("💥 Ящик разрушен!")
				g.createDebris(box.position, 8)

				// Удаляем ящик
				g.destructibleObjects = append(g.destructibleObjects[:i], g.destructibleObjects[i+1:]...)
			}
			break
		}
	}

	// Проверяем попадание по врагам (только если не попали в ящик)
	if !hitSomething {
		for i := len(g.enemyPositions) - 1; i >= 0; i-- {
			enemyPos := g.enemyPositions[i]

			// Создаем AABB для врага
			enemyAABB := customMath.NewAABBFromCenter(enemyPos, mgl32.Vec3{0.5, 0.5, 0.5})

			// Проверяем пересечение
			if hit, distance := ray.IntersectAABB(enemyAABB); hit && distance < closestDist {
				// Попали! Трассер идет до врага
				tracerEnd = g.camera.Position.Add(g.camera.Front.Mul(distance))

				// Убили врага!
				g.enemyPositions = append(g.enemyPositions[:i], g.enemyPositions[i+1:]...)
				g.enemiesKilled++

				fmt.Printf("💀 Враг убит! Осталось: %d\n", len(g.enemyPositions))

				if len(g.enemyPositions) == 0 {
					fmt.Println("\n🎉 Победа! Все враги уничтожены!")
					fmt.Printf("Нажмите ESC для выхода\n")
				}
				break
			}
		}
	}

	// Создаем трассер пули
	tracer := BulletTracer{
		start:    g.camera.Position,
		end:      tracerEnd,
		lifetime: 0.1, // Трассер видим 0.1 секунды
		maxLife:  0.1,
	}
	g.bulletTracers = append(g.bulletTracers, tracer)
}

// createDebris создает осколки при разрушении объекта
func (g *DoomGame) createDebris(position mgl32.Vec3, count int) {
	for i := 0; i < count; i++ {
		// Случайная скорость во все стороны
		angle := float32(i) * (2.0 * math.Pi / float32(count))
		speed := float32(3.0 + float64(i)*0.5)

		velocity := mgl32.Vec3{
			float32(math.Cos(float64(angle))) * speed,
			float32(2.0 + float64(i)*0.3), // Вверх
			float32(math.Sin(float64(angle))) * speed,
		}

		debris := Debris{
			position: position,
			velocity: velocity,
			rotation: float32(i) * 0.5,
			lifetime: 2.0, // Осколки живут 2 секунды
			size:     0.2,
		}
		g.debris = append(g.debris, debris)
	}
}

func (g *DoomGame) onRender(engine *core.Engine) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	width, height := engine.GetWindow().GetSize()
	widthF := float32(width)
	heightF := float32(height)

	// === РИСУЕМ 3D СЦЕНУ ===
	gl.Enable(gl.DEPTH_TEST)
	g.shader.Use()

	// Получаем матрицы
	aspectRatio := widthF / heightF
	projection := g.camera.GetProjectionMatrix(aspectRatio)
	view := g.camera.GetViewMatrix()

	g.shader.SetMat4("uProjection", projection)
	g.shader.SetMat4("uView", view)

	// Рисуем пол
	model := mgl32.Ident4()
	g.shader.SetMat4("uModel", model)
	gl.BindVertexArray(g.floorVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)

	// Рисуем стены (периметр арены)
	wallPositions := []mgl32.Vec3{
		{0, 1.5, -10}, {10, 1.5, 0}, {-10, 1.5, 0}, {0, 1.5, 10},
		{5, 1.5, -10}, {-5, 1.5, -10}, {10, 1.5, 5}, {10, 1.5, -5},
		{-10, 1.5, 5}, {-10, 1.5, -5}, {5, 1.5, 10}, {-5, 1.5, 10},
	}

	gl.BindVertexArray(g.wallVAO)
	for _, pos := range wallPositions {
		model = mgl32.Translate3D(pos.X(), pos.Y(), pos.Z())
		model = model.Mul4(mgl32.Scale3D(1, 3, 1))
		g.shader.SetMat4("uModel", model)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
	}

	// Рисуем врагов
	gl.BindVertexArray(g.enemyVAO)
	currentTime := float32(time.Now().UnixNano()) / 1e9
	for _, pos := range g.enemyPositions {
		// Анимация: враги немного "дышат" (пульсируют)
		scale := 1.0 + float32(math.Sin(float64(currentTime*2)))*0.1

		model = mgl32.Translate3D(pos.X(), pos.Y(), pos.Z())
		model = model.Mul4(mgl32.Scale3D(scale, scale, scale))
		g.shader.SetMat4("uModel", model)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
	}

	// === РИСУЕМ ТРАССЕРЫ ПУЛЬ (3D линии) ===
	if len(g.bulletTracers) > 0 {
		gl.Disable(gl.DEPTH_TEST)
		gl.LineWidth(3.0)

		for _, tracer := range g.bulletTracers {
			// Альфа на основе времени жизни
			alpha := tracer.lifetime / tracer.maxLife

			vertices := []float32{
				// Начало линии (желтый)
				tracer.start.X(), tracer.start.Y(), tracer.start.Z(), 1.0, 1.0, 0.0,
				// Конец линии (оранжевый с альфой)
				tracer.end.X(), tracer.end.Y(), tracer.end.Z(), 1.0 * alpha, 0.5 * alpha, 0.0,
			}

			gl.BindVertexArray(g.lineVAO)
			gl.BindBuffer(gl.ARRAY_BUFFER, g.lineVBO)
			gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.DYNAMIC_DRAW)

			g.shader.SetMat4("uModel", mgl32.Ident4())
			gl.DrawArrays(gl.LINES, 0, 2)
		}

		gl.LineWidth(1.0)
		gl.Enable(gl.DEPTH_TEST)
	}

	gl.BindVertexArray(0)

	// === РИСУЕМ UI (2D поверх всего) ===
	gl.Disable(gl.DEPTH_TEST)

	// Прицел (крестик в центре экрана)
	centerX := widthF / 2
	centerY := heightF / 2
	crosshairSize := float32(10)
	crosshairThickness := float32(2)
	crosshairColor := mgl32.Vec4{0, 1, 0, 0.7} // Зеленый полупрозрачный

	g.uiRenderer.DrawLine(centerX-crosshairSize, centerY, centerX+crosshairSize, centerY, crosshairThickness, crosshairColor)
	g.uiRenderer.DrawLine(centerX, centerY-crosshairSize, centerX, centerY+crosshairSize, crosshairThickness, crosshairColor)

	// Полоска здоровья (красная)
	healthBarX := float32(20)
	healthBarY := heightF - 40
	healthBarWidth := float32(200)
	healthBarHeight := float32(20)

	// Фон полоски здоровья (темный)
	g.uiRenderer.DrawRect(healthBarX, healthBarY, healthBarWidth, healthBarHeight, mgl32.Vec4{0.2, 0.2, 0.2, 0.8})

	// Актуальное здоровье (красное)
	healthPercent := float32(g.playerHealth) / float32(g.maxHealth)
	healthColor := mgl32.Vec4{1, 0, 0, 0.9}
	if healthPercent < 0.3 {
		// Мигающее здоровье когда мало HP
		pulse := float32(math.Sin(float64(currentTime * 5)))
		healthColor = mgl32.Vec4{1, pulse*0.3 + 0.4, 0, 0.9}
	}
	g.uiRenderer.DrawRect(healthBarX+2, healthBarY+2, (healthBarWidth-4)*healthPercent, healthBarHeight-4, healthColor)

	// Счетчик врагов
	enemyCountY := healthBarY + healthBarHeight + 10
	enemyBarWidth := float32(150)
	g.uiRenderer.DrawRect(healthBarX, enemyCountY, enemyBarWidth, 20, mgl32.Vec4{0.2, 0.1, 0.1, 0.8})

	// Показываем количество оставшихся врагов красными квадратиками
	for i := 0; i < len(g.enemyPositions); i++ {
		squareSize := float32(12)
		squareX := healthBarX + 5 + float32(i)*(squareSize+3)
		squareY := enemyCountY + 4
		g.uiRenderer.DrawRect(squareX, squareY, squareSize, squareSize, mgl32.Vec4{1, 0, 0, 0.9})
	}

	// Счетчик патронов (справа внизу)
	ammoX := widthF - 220
	ammoY := heightF - 60
	ammoWidth := float32(200)
	ammoHeight := float32(40)

	// Фон счетчика патронов
	g.uiRenderer.DrawRect(ammoX, ammoY, ammoWidth, ammoHeight, mgl32.Vec4{0.1, 0.1, 0.1, 0.8})

	// Индикатор текущих патронов (желтые полоски)
	for i := 0; i < g.currentAmmo; i++ {
		bulletWidth := float32(12)
		bulletHeight := float32(25)
		bulletX := ammoX + 10 + float32(i)*(bulletWidth+2)
		bulletY := ammoY + 7
		bulletColor := mgl32.Vec4{1, 0.8, 0, 0.9}
		if g.isReloading {
			// Мигание при перезарядке
			pulse := float32(math.Sin(float64(currentTime * 8)))
			bulletColor = mgl32.Vec4{0.5 + pulse*0.5, 0.4, 0, 0.9}
		}
		g.uiRenderer.DrawRect(bulletX, bulletY, bulletWidth, bulletHeight, bulletColor)
	}

	// Текст "RELOAD" при перезарядке (большими прямоугольниками)
	if g.isReloading {
		reloadX := widthF/2 - 100
		reloadY := heightF - 150
		pulse := float32(math.Sin(float64(currentTime * 4)))
		reloadAlpha := 0.5 + pulse*0.3
		g.uiRenderer.DrawRect(reloadX, reloadY, 200, 40, mgl32.Vec4{1, 1, 0, reloadAlpha})
	}

	// === РИСУЕМ ПИСТОЛЕТ (2D спрайт в правом нижнем углу) ===
	gunX := widthF - 250
	gunY := heightF - 200

	// Отдача - двигаем пистолет вверх
	if g.gunRecoil > 0 {
		gunY -= g.gunRecoil * 100
	}

	// Ствол пистолета
	barrelColor := mgl32.Vec4{0.2, 0.2, 0.2, 1.0}
	g.uiRenderer.DrawRect(gunX+40, gunY+20, 100, 30, barrelColor)

	// Прицельная планка
	g.uiRenderer.DrawRect(gunX+130, gunY+15, 8, 10, mgl32.Vec4{0.8, 0.8, 0.8, 1.0})

	// Рукоятка
	gripColor := mgl32.Vec4{0.15, 0.1, 0.05, 1.0}
	g.uiRenderer.DrawRect(gunX+50, gunY+50, 40, 80, gripColor)

	// Затвор
	slideColor := mgl32.Vec4{0.3, 0.3, 0.3, 1.0}
	g.uiRenderer.DrawRect(gunX+45, gunY+10, 90, 25, slideColor)

	// Спусковой крючок
	g.uiRenderer.DrawRect(gunX+60, gunY+60, 15, 25, mgl32.Vec4{0.1, 0.1, 0.1, 1.0})

	gl.Enable(gl.DEPTH_TEST)
}

func (g *DoomGame) onShutdown(engine *core.Engine) {
	fmt.Println("\n=== Статистика ===")
	fmt.Printf("Убито врагов: %d\n", g.enemiesKilled)

	if g.shader != nil {
		g.shader.Delete()
	}
	if g.uiRenderer != nil {
		g.uiRenderer.Cleanup()
	}
	gl.DeleteVertexArrays(1, &g.wallVAO)
	gl.DeleteBuffers(1, &g.wallVBO)
	gl.DeleteVertexArrays(1, &g.floorVAO)
	gl.DeleteBuffers(1, &g.floorVBO)
	gl.DeleteVertexArrays(1, &g.enemyVAO)
	gl.DeleteBuffers(1, &g.enemyVBO)
	gl.DeleteVertexArrays(1, &g.lineVAO)
	gl.DeleteBuffers(1, &g.lineVBO)
}
