package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"strings"

	"github.com/Salamander5876/AnimoEngine/pkg/core"
	"github.com/Salamander5876/AnimoEngine/pkg/graphics"
	"github.com/Salamander5876/AnimoEngine/pkg/graphics/shader"
	"github.com/Salamander5876/AnimoEngine/pkg/graphics/ui"
	"github.com/Salamander5876/AnimoEngine/pkg/platform/input"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

func init() {
	runtime.LockOSThread()
}

// Константы физики
const (
	MaxSpeed        = 12.0
	Acceleration    = 0.15
	Deceleration    = 0.015
	RotationSpeed   = 2.5
	ReverseSpeedMul = 0.5
	GrassMaxSpeed   = 5.0
	GrassDecel      = 3.0
	CollisionTransfer = 0.7
)

// Тип тайла
type TileType int

const (
	TileAsphalt TileType = iota
	TileWall
	TileGrass
	TileSpawn
	TileFinish
)

// Тип управления
type ControlType int

const (
	ControlWASD ControlType = iota
	ControlArrows
	ControlGamepad
)

// Car машина игрока
type Car struct {
	x, y          float32
	angle         float32
	speed         float32
	maxSpeed      float32
	texture       *graphics.Texture
	laps          int
	lastLapTime   float64
	controlType   ControlType
	playerID      int
	collisionBox  [4]mgl32.Vec2 // 4 точки для коллизии
}

// Map игровая карта
type Map struct {
	width, height int
	tiles         [][]TileType
	tileSize      float32
	textures      map[TileType]*graphics.Texture
}

// GameState состояние игры
type GameState int

const (
	StateMenu GameState = iota
	StateGame
	StateVictory
)

// RacingGame главная структура игры
type RacingGame struct {
	engine      *core.Engine
	shader      *shader.Shader
	uiRenderer  *ui.UIRenderer

	// Состояние
	state       GameState
	winner      int

	// Игроки
	cars        []*Car
	numPlayers  int
	lapsToWin   int

	// Карта
	gameMap     *Map

	// Геометрия
	quadVAO     uint32
	quadVBO     uint32

	// Камера
	cameraX     float32
	cameraY     float32
	zoom        float32

	// Время
	gameTime    float64
}

func main() {
	game := &RacingGame{
		state:      StateMenu,
		numPlayers: 1,
		lapsToWin:  3,
		zoom:       1.0,
	}

	config := core.DefaultEngineConfig()
	config.WindowConfig.Title = "Racing Game - AnimoEngine"
	config.WindowConfig.Width = 1800
	config.WindowConfig.Height = 1000

	engine := core.NewEngineWithConfig(config)
	game.engine = engine
	engine.SetInitCallback(game.onInit)
	engine.SetUpdateCallback(game.onUpdate)
	engine.SetRenderCallback(game.onRender)
	engine.SetShutdownCallback(game.onShutdown)

	if err := engine.Run(); err != nil {
		log.Fatal(err)
	}
}

func (g *RacingGame) onInit(engine *core.Engine) error {
	// Инициализируем OpenGL
	if err := gl.Init(); err != nil {
		return fmt.Errorf("failed to initialize OpenGL: %v", err)
	}

	// Создаем шейдер для 2D рендеринга
	vertexShader := `
#version 330 core
layout (location = 0) in vec2 aPos;
layout (location = 1) in vec2 aTexCoord;

out vec2 TexCoord;

uniform mat4 uProjection;
uniform mat4 uView;
uniform mat4 uModel;

void main() {
    gl_Position = uProjection * uView * uModel * vec4(aPos, 0.0, 1.0);
    TexCoord = aTexCoord;
}
`

	fragmentShader := `
#version 330 core
in vec2 TexCoord;
out vec4 FragColor;

uniform sampler2D texture1;

void main() {
    FragColor = texture(texture1, TexCoord);
}
`

	var err error
	g.shader, err = shader.NewShader(vertexShader, fragmentShader)
	if err != nil {
		return err
	}

	// Создаем UI рендерер
	g.uiRenderer, err = ui.NewUIRenderer()
	if err != nil {
		return err
	}
	width, height := engine.GetWindow().GetSize()
	g.uiRenderer.SetProjection(float32(width), float32(height))

	// Создаем quad для отрисовки спрайтов
	g.createQuad()

	// Загружаем карту
	err = g.loadMap("otherGame/race/src/maps/map1.txt")
	if err != nil {
		return fmt.Errorf("failed to load map: %v", err)
	}

	// Настройки OpenGL
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(0.1, 0.1, 0.1, 1.0)

	fmt.Println("\n=== Racing Game ===")
	fmt.Println("Press ENTER to start!")

	return nil
}

func (g *RacingGame) createQuad() {
	vertices := []float32{
		// Позиции   // TexCoords
		-0.5, -0.5,  0.0, 1.0,
		0.5, -0.5,   1.0, 1.0,
		0.5, 0.5,    1.0, 0.0,

		-0.5, -0.5,  0.0, 1.0,
		0.5, 0.5,    1.0, 0.0,
		-0.5, 0.5,   0.0, 0.0,
	}

	gl.GenVertexArrays(1, &g.quadVAO)
	gl.GenBuffers(1, &g.quadVBO)

	gl.BindVertexArray(g.quadVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, g.quadVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Позиция
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	// TexCoord
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}

func (g *RacingGame) loadMap(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	g.gameMap = &Map{
		textures: make(map[TileType]*graphics.Texture),
	}

	// Загружаем текстуры тайлов
	asphaltTex, err := graphics.LoadTexture("otherGame/race/src/maps/asphalt.png")
	if err != nil {
		return err
	}
	g.gameMap.textures[TileAsphalt] = asphaltTex
	g.gameMap.textures[TileSpawn] = asphaltTex // Spawn использует асфальт

	grassTex, err := graphics.LoadTexture("otherGame/race/src/maps/grass.png")
	if err != nil {
		return err
	}
	g.gameMap.textures[TileGrass] = grassTex

	wallTex, err := graphics.LoadTexture("otherGame/race/src/maps/wall.png")
	if err != nil {
		return err
	}
	g.gameMap.textures[TileWall] = wallTex

	finishTex, err := graphics.LoadTexture("otherGame/race/src/maps/finish.png")
	if err != nil {
		return err
	}
	g.gameMap.textures[TileFinish] = finishTex

	// Читаем карту
	scanner := bufio.NewScanner(file)
	var tiles [][]TileType

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		chars := strings.Split(line, "")
		row := make([]TileType, len(chars))

		for i, char := range chars {
			switch char {
			case "0":
				row[i] = TileAsphalt
			case "1":
				row[i] = TileWall
			case "2":
				row[i] = TileGrass
			case "8":
				row[i] = TileSpawn
			case "9":
				row[i] = TileFinish
			default:
				row[i] = TileAsphalt
			}
		}
		tiles = append(tiles, row)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	g.gameMap.tiles = tiles
	g.gameMap.height = len(tiles)
	if g.gameMap.height > 0 {
		g.gameMap.width = len(tiles[0])
	}

	// Вычисляем размер тайла
	width, height := g.engine.GetWindow().GetSize()
	tileW := float32(width) / float32(g.gameMap.width)
	tileH := float32(height) / float32(g.gameMap.height)
	g.gameMap.tileSize = float32(math.Min(float64(tileW), float64(tileH)))

	fmt.Printf("Map loaded: %dx%d, tile size: %.1f\n", g.gameMap.width, g.gameMap.height, g.gameMap.tileSize)

	return nil
}

func (g *RacingGame) startGame() {
	g.state = StateGame
	g.gameTime = 0
	g.cars = make([]*Car, 0)

	// Находим spawn точки
	spawnPoints := make([]mgl32.Vec2, 0)
	for y := 0; y < g.gameMap.height; y++ {
		for x := 0; x < g.gameMap.width; x++ {
			if g.gameMap.tiles[y][x] == TileSpawn {
				spawnPoints = append(spawnPoints, mgl32.Vec2{float32(x), float32(y)})
			}
		}
	}

	// Создаем машины для игроков
	carTextures := []string{
		"otherGame/race/src/cars/porshe.png",
		"otherGame/race/src/cars/green.png",
		"otherGame/race/src/cars/Huracan.png",
	}

	for i := 0; i < g.numPlayers && i < len(spawnPoints); i++ {
		texture, err := graphics.LoadTexture(carTextures[i%len(carTextures)])
		if err != nil {
			log.Printf("Failed to load car texture: %v", err)
			continue
		}

		spawn := spawnPoints[i]
		car := &Car{
			x:           (spawn.X() + 0.5) * g.gameMap.tileSize,
			y:           (spawn.Y() + 0.5) * g.gameMap.tileSize,
			angle:       0,
			speed:       0,
			maxSpeed:    MaxSpeed,
			texture:     texture,
			laps:        0,
			controlType: ControlType(i),
			playerID:    i + 1,
		}
		g.cars = append(g.cars, car)
	}

	fmt.Printf("Game started with %d players, racing to %d laps!\n", g.numPlayers, g.lapsToWin)
}

func (g *RacingGame) onUpdate(engine *core.Engine, dt float32) {
	inputMgr := engine.GetInputManager()
	g.gameTime += float64(dt)

	switch g.state {
	case StateMenu:
		// Меню: нажми Enter для старта
		if inputMgr.IsKeyPressed(input.KeyEnter) {
			g.startGame()
		}
		// Изменение количества игроков
		if inputMgr.IsKeyJustPressed(input.KeyUp) && g.numPlayers < 3 {
			g.numPlayers++
		}
		if inputMgr.IsKeyJustPressed(input.KeyDown) && g.numPlayers > 1 {
			g.numPlayers--
		}

	case StateGame:
		// Обновляем все машины
		for _, car := range g.cars {
			g.updateCar(car, dt, inputMgr)
		}

		// Проверка коллизий между машинами
		for i := 0; i < len(g.cars); i++ {
			for j := i + 1; j < len(g.cars); j++ {
				g.checkCarCollision(g.cars[i], g.cars[j])
			}
		}

		// Проверка победы
		for _, car := range g.cars {
			if car.laps >= g.lapsToWin {
				g.state = StateVictory
				g.winner = car.playerID
				fmt.Printf("\n🏁 Player %d wins!\n", g.winner)
			}
		}

		// Обновляем камеру (следим за первым игроком)
		if len(g.cars) > 0 {
			g.cameraX = g.cars[0].x
			g.cameraY = g.cars[0].y
		}

	case StateVictory:
		if inputMgr.IsKeyPressed(input.KeyEnter) {
			g.state = StateMenu
		}
	}

	// ESC для выхода
	if inputMgr.IsKeyPressed(input.KeyEscape) {
		engine.Stop()
	}
}

func (g *RacingGame) updateCar(car *Car, dt float32, inputMgr *input.InputManager) {
	// Получаем input в зависимости от типа управления
	forward, backward, left, right, reset := g.getInput(car.controlType, inputMgr)

	// Сброс позиции
	if reset {
		// Найти ближайший spawn
		// TODO: implement
	}

	// Ускорение/торможение
	if forward {
		car.speed += Acceleration
		if car.speed > car.maxSpeed {
			car.speed = car.maxSpeed
		}
	} else if backward {
		car.speed -= Acceleration
		if car.speed < -car.maxSpeed*ReverseSpeedMul {
			car.speed = -car.maxSpeed * ReverseSpeedMul
		}
	} else {
		// Естественное замедление
		if car.speed > 0 {
			car.speed -= Deceleration
			if car.speed < 0 {
				car.speed = 0
			}
		} else if car.speed < 0 {
			car.speed += Deceleration
			if car.speed > 0 {
				car.speed = 0
			}
		}
	}

	// Поворот
	if left && car.speed != 0 {
		car.angle -= RotationSpeed * float32(math.Abs(float64(car.speed))/MaxSpeed)
	}
	if right && car.speed != 0 {
		car.angle += RotationSpeed * float32(math.Abs(float64(car.speed))/MaxSpeed)
	}

	// Движение
	angleRad := car.angle * math.Pi / 180.0
	car.x += float32(math.Cos(float64(angleRad))) * car.speed * dt * 60
	car.y += float32(math.Sin(float64(angleRad))) * car.speed * dt * 60

	// Проверка коллизии с картой
	g.checkMapCollision(car)

	// Обновление collision box
	g.updateCollisionBox(car)
}

func (g *RacingGame) getInput(controlType ControlType, inputMgr *input.InputManager) (forward, backward, left, right, reset bool) {
	switch controlType {
	case ControlWASD:
		return inputMgr.IsKeyPressed(input.KeyW),
			inputMgr.IsKeyPressed(input.KeyS),
			inputMgr.IsKeyPressed(input.KeyA),
			inputMgr.IsKeyPressed(input.KeyD),
			inputMgr.IsKeyPressed(input.KeyR)
	case ControlArrows:
		return inputMgr.IsKeyPressed(input.KeyUp),
			inputMgr.IsKeyPressed(input.KeyDown),
			inputMgr.IsKeyPressed(input.KeyLeft),
			inputMgr.IsKeyPressed(input.KeyRight),
			inputMgr.IsKeyPressed(input.KeyLeftShift)
	}
	return false, false, false, false, false
}

func (g *RacingGame) checkMapCollision(car *Car) {
	// Получаем тайл под машиной
	tileX := int(car.x / g.gameMap.tileSize)
	tileY := int(car.y / g.gameMap.tileSize)

	if tileX < 0 || tileX >= g.gameMap.width || tileY < 0 || tileY >= g.gameMap.height {
		// За границами карты - отталкиваем назад
		car.speed *= -0.5
		return
	}

	tile := g.gameMap.tiles[tileY][tileX]

	switch tile {
	case TileWall:
		// Стена - отскок
		car.speed *= -0.5
		angleRad := car.angle * math.Pi / 180.0
		car.x -= float32(math.Cos(float64(angleRad))) * 5
		car.y -= float32(math.Sin(float64(angleRad))) * 5

	case TileGrass:
		// Трава - замедление
		car.maxSpeed = GrassMaxSpeed
		if car.speed > 0 {
			car.speed -= Deceleration * GrassDecel
		}

	case TileAsphalt, TileSpawn:
		// Асфальт - нормальная скорость
		car.maxSpeed = MaxSpeed

	case TileFinish:
		// Финишная линия - засчитываем круг
		if g.gameTime-car.lastLapTime > 3.0 { // 3 секунды кулдаун
			car.laps++
			car.lastLapTime = g.gameTime
			fmt.Printf("Player %d completed lap %d/%d\n", car.playerID, car.laps, g.lapsToWin)
		}
		car.maxSpeed = MaxSpeed
	}
}

func (g *RacingGame) updateCollisionBox(car *Car) {
	// 4 точки вокруг машины (20x40 пикселей)
	w := float32(10.0)
	h := float32(20.0)

	angleRad := float64(car.angle * math.Pi / 180.0)
	cos := float32(math.Cos(angleRad))
	sin := float32(math.Sin(angleRad))

	// Поворачиваем точки
	points := []mgl32.Vec2{
		{-w, -h}, {w, -h}, {w, h}, {-w, h},
	}

	for i, p := range points {
		rx := p.X()*cos - p.Y()*sin
		ry := p.X()*sin + p.Y()*cos
		car.collisionBox[i] = mgl32.Vec2{car.x + rx, car.y + ry}
	}
}

func (g *RacingGame) checkCarCollision(car1, car2 *Car) {
	// Простая дистанционная коллизия
	dx := car1.x - car2.x
	dy := car1.y - car2.y
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	if dist < 25.0 { // Радиус коллизии
		// Отталкивание
		angle := float32(math.Atan2(float64(dy), float64(dx)))
		overlap := 25.0 - dist

		car1.x += float32(math.Cos(float64(angle))) * overlap * 0.5
		car1.y += float32(math.Sin(float64(angle))) * overlap * 0.5
		car2.x -= float32(math.Cos(float64(angle))) * overlap * 0.5
		car2.y -= float32(math.Sin(float64(angle))) * overlap * 0.5

		// Передача скорости
		speedDiff := car1.speed - car2.speed
		car1.speed -= speedDiff * CollisionTransfer
		car2.speed += speedDiff * CollisionTransfer
	}
}

func (g *RacingGame) onRender(engine *core.Engine) {
	gl.Clear(gl.COLOR_BUFFER_BIT)

	width, height := engine.GetWindow().GetSize()
	widthF := float32(width)
	heightF := float32(height)

	g.shader.Use()

	// Ортографическая проекция для 2D
	projection := mgl32.Ortho(0, widthF, heightF, 0, -1, 1)

	// View матрица (камера следит за игроком)
	view := mgl32.Ident4()
	if g.state == StateGame && len(g.cars) > 0 {
		// Центрируем камеру на первом игроке
		view = mgl32.Translate3D(-g.cameraX+widthF/2, -g.cameraY+heightF/2, 0)
	}

	g.shader.SetMat4("uProjection", projection)
	g.shader.SetMat4("uView", view)

	switch g.state {
	case StateMenu:
		g.renderMenu(widthF, heightF)
	case StateGame:
		g.renderGame()
	case StateVictory:
		g.renderVictory(widthF, heightF)
	}
}

func (g *RacingGame) renderMenu(width, height float32) {
	// Используем UI renderer для меню
	g.shader.SetMat4("uView", mgl32.Ident4())

	// Заголовок
	titleY := height * 0.2
	g.uiRenderer.DrawRect(width/2-200, titleY, 400, 80, mgl32.Vec4{0.2, 0.2, 0.2, 0.9})

	// Настройки
	optionsY := height * 0.5
	g.uiRenderer.DrawRect(width/2-150, optionsY, 300, 200, mgl32.Vec4{0.15, 0.15, 0.15, 0.9})

	// Кнопка старта
	startY := height * 0.8
	g.uiRenderer.DrawRect(width/2-100, startY, 200, 60, mgl32.Vec4{0, 0.6, 0, 0.9})
}

func (g *RacingGame) renderGame() {
	gl.BindVertexArray(g.quadVAO)

	// Рисуем карту
	for y := 0; y < g.gameMap.height; y++ {
		for x := 0; x < g.gameMap.width; x++ {
			tile := g.gameMap.tiles[y][x]
			texture := g.gameMap.textures[tile]

			if texture != nil {
				gl.ActiveTexture(gl.TEXTURE0)
				gl.BindTexture(gl.TEXTURE_2D, texture.ID)

				model := mgl32.Translate3D(
					float32(x)*g.gameMap.tileSize+g.gameMap.tileSize/2,
					float32(y)*g.gameMap.tileSize+g.gameMap.tileSize/2,
					0,
				)
				model = model.Mul4(mgl32.Scale3D(g.gameMap.tileSize, g.gameMap.tileSize, 1))

				g.shader.SetMat4("uModel", model)
				gl.DrawArrays(gl.TRIANGLES, 0, 6)
			}
		}
	}

	// Рисуем машины
	for _, car := range g.cars {
		if car.texture != nil {
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, car.texture.ID)

			model := mgl32.Translate3D(car.x, car.y, 0)
			model = model.Mul4(mgl32.HomogRotate3DZ(car.angle * math.Pi / 180.0))
			model = model.Mul4(mgl32.Scale3D(20, 40, 1))

			g.shader.SetMat4("uModel", model)
			gl.DrawArrays(gl.TRIANGLES, 0, 6)
		}
	}

	gl.BindVertexArray(0)

	// HUD поверх игры
	g.renderHUD()
}

func (g *RacingGame) renderHUD() {
	g.shader.SetMat4("uView", mgl32.Ident4())

	// Информация о каждом игроке
	for i, car := range g.cars {
		y := float32(20 + i*80)

		// Фон
		g.uiRenderer.DrawRect(10, y, 200, 70, mgl32.Vec4{0, 0, 0, 0.6})

		// Полоска скорости
		speedPercent := float32(math.Abs(float64(car.speed)) / MaxSpeed)
		g.uiRenderer.DrawRect(15, y+50, 190*speedPercent, 15, mgl32.Vec4{0, 1, 0, 0.8})
	}
}

func (g *RacingGame) renderVictory(width, height float32) {
	g.shader.SetMat4("uView", mgl32.Ident4())

	// Экран победы
	g.uiRenderer.DrawRect(width/2-250, height/2-150, 500, 300, mgl32.Vec4{0.1, 0.1, 0.1, 0.95})
	g.uiRenderer.DrawRect(width/2-200, height/2-100, 400, 80, mgl32.Vec4{1, 0.8, 0, 0.9})
}

func (g *RacingGame) onShutdown(engine *core.Engine) {
	fmt.Println("\n=== Game Over ===")

	if g.shader != nil {
		g.shader.Delete()
	}
	if g.uiRenderer != nil {
		g.uiRenderer.Cleanup()
	}
	gl.DeleteVertexArrays(1, &g.quadVAO)
	gl.DeleteBuffers(1, &g.quadVBO)
}
