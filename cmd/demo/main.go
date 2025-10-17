package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/Salamander5876/AnimoEngine/pkg/core"
	"github.com/Salamander5876/AnimoEngine/pkg/graphics/shader"
	"github.com/Salamander5876/AnimoEngine/pkg/platform/input"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// init вызывается перед main и блокирует главный поток для OpenGL
func init() {
	// КРИТИЧЕСКИ ВАЖНО: OpenGL требует чтобы все вызовы были из одного потока ОС
	runtime.LockOSThread()
}

// DemoApp демонстрационное приложение
type DemoApp struct {
	engine *core.Engine
	shader *shader.Shader
	vao    uint32
	vbo    uint32

	rotation float32
}

func main() {
	app := &DemoApp{}

	// Создаем движок
	config := core.DefaultEngineConfig()
	config.WindowConfig.Title = "AnimoEngine Demo - Rotating Triangle"
	config.WindowConfig.Width = 1280
	config.WindowConfig.Height = 720

	app.engine = core.NewEngineWithConfig(config)

	// Устанавливаем колбэки
	app.engine.SetInitCallback(app.onInit)
	app.engine.SetUpdateCallback(app.onUpdate)
	app.engine.SetRenderCallback(app.onRender)
	app.engine.SetShutdownCallback(app.onShutdown)

	// Запускаем движок
	if err := app.engine.Run(); err != nil {
		log.Fatalf("Engine error: %v", err)
	}
}

func (app *DemoApp) onInit(engine *core.Engine) error {
	fmt.Println("=== AnimoEngine Demo ===")
	fmt.Println("Инициализация движка...")

	// Инициализируем OpenGL
	if err := gl.Init(); err != nil {
		return fmt.Errorf("failed to initialize OpenGL: %w", err)
	}

	fmt.Printf("OpenGL Version: %s\n", gl.GoStr(gl.GetString(gl.VERSION)))
	fmt.Printf("GLSL Version: %s\n", gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION)))

	// Показываем логотип на 3 секунды
	fmt.Println("\nПоказываем логотип AnimoEngine...")
	splash, err := core.NewSplashScreen("logo.png", 3*time.Second)
	if err != nil {
		fmt.Printf("Не удалось загрузить логотип: %v\n", err)
		// Продолжаем без логотипа
	} else {
		splash.Show(engine)
		splash.Cleanup()
		fmt.Println("Логотип показан!")
	}

	// Создаем шейдер для простого треугольника
	simpleVertexShader := `
	#version 330 core

	layout (location = 0) in vec3 aPosition;
	layout (location = 1) in vec4 aColor;

	out vec4 vertexColor;

	uniform mat4 uTransform;

	void main() {
		gl_Position = uTransform * vec4(aPosition, 1.0);
		vertexColor = aColor;
	}
	`

	simpleFragmentShader := `
	#version 330 core

	in vec4 vertexColor;
	out vec4 FragColor;

	void main() {
		FragColor = vertexColor;
	}
	`

	var err error
	app.shader, err = shader.NewShader(simpleVertexShader, simpleFragmentShader)
	if err != nil {
		return fmt.Errorf("failed to create shader: %w", err)
	}

	// Создаем треугольник
	vertices := []float32{
		// Позиции        // Цвета
		0.0,  0.5,  0.0,  1.0, 0.0, 0.0, 1.0, // Верх (красный)
		-0.5, -0.5, 0.0,  0.0, 1.0, 0.0, 1.0, // Левый (зеленый)
		0.5, -0.5, 0.0,  0.0, 0.0, 1.0, 1.0, // Правый (синий)
	}

	// Создаем VAO и VBO
	gl.GenVertexArrays(1, &app.vao)
	gl.GenBuffers(1, &app.vbo)

	gl.BindVertexArray(app.vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, app.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Атрибут позиции
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 7*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Атрибут цвета
	gl.VertexAttribPointer(1, 4, gl.FLOAT, false, 7*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)

	// Настраиваем OpenGL
	gl.ClearColor(0.1, 0.1, 0.1, 1.0)

	fmt.Println("Инициализация завершена!")
	fmt.Println("\nУправление:")
	fmt.Println("  ESC - выход")
	fmt.Println("  SPACE - пауза вращения")
	fmt.Println("  R - сброс вращения")

	return nil
}

func (app *DemoApp) onUpdate(engine *core.Engine, deltaTime float32) {
	inputMgr := engine.GetInputManager()

	// Выход по ESC
	if inputMgr.IsKeyJustPressed(input.KeyEscape) {
		engine.Stop()
	}

	// Пауза вращения по SPACE
	if inputMgr.IsKeyJustPressed(input.KeySpace) {
		if engine.GetWorld().IsPaused() {
			engine.GetWorld().Resume()
			fmt.Println("Вращение возобновлено")
		} else {
			engine.GetWorld().Pause()
			fmt.Println("Вращение приостановлено")
		}
	}

	// Сброс вращения по R
	if inputMgr.IsKeyJustPressed(int('R')) {
		app.rotation = 0
		fmt.Println("Вращение сброшено")
	}

	// Обновляем вращение если не на паузе
	if !engine.GetWorld().IsPaused() {
		app.rotation += deltaTime * 1.0 // 1 радиан в секунду
	}
}

func (app *DemoApp) onRender(engine *core.Engine) {
	// Очищаем экран
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// Используем шейдер
	app.shader.Use()

	// Создаем матрицу трансформации (вращение)
	transform := mgl32.HomogRotate3D(app.rotation, mgl32.Vec3{0, 0, 1})
	app.shader.SetMat4("uTransform", transform)

	// Рендерим треугольник
	gl.BindVertexArray(app.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, 3)
	gl.BindVertexArray(0)

	// Выводим FPS каждые 60 кадров
	if engine.GetFrameCount()%60 == 0 {
		fmt.Printf("\rFPS: %.0f | DeltaTime: %.3fms",
			engine.GetFPS(),
			engine.GetDeltaTime()*1000)
	}
}

func (app *DemoApp) onShutdown(engine *core.Engine) {
	fmt.Println("\n\nЗавершение работы движка...")

	// Очищаем ресурсы OpenGL
	if app.shader != nil {
		app.shader.Delete()
	}

	gl.DeleteVertexArrays(1, &app.vao)
	gl.DeleteBuffers(1, &app.vbo)

	fmt.Println("Движок остановлен. До свидания!")
}
