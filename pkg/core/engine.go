package core

import (
	"fmt"
	"time"

	"github.com/Salamander5876/AnimoEngine/pkg/core/ecs"
	"github.com/Salamander5876/AnimoEngine/pkg/core/event"
	"github.com/Salamander5876/AnimoEngine/pkg/core/resource"
	"github.com/Salamander5876/AnimoEngine/pkg/platform/input"
	"github.com/Salamander5876/AnimoEngine/pkg/platform/window"
)

// EngineConfig конфигурация движка
type EngineConfig struct {
	WindowConfig        window.WindowConfig
	TargetFPS           int
	MaxResourceCacheSize int64
	LoadWorkers         int
}

// DefaultEngineConfig возвращает конфигурацию движка по умолчанию
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		WindowConfig:        window.DefaultConfig(),
		TargetFPS:           60,
		MaxResourceCacheSize: 512 * 1024 * 1024, // 512MB
		LoadWorkers:         4,
	}
}

// Engine главный класс игрового движка
type Engine struct {
	config EngineConfig

	// Подсистемы
	window          *window.Window
	world           *ecs.World
	eventBus        *event.EventBus
	resourceManager *resource.ResourceManager
	inputManager    *input.InputManager

	// Состояние
	running    bool
	targetFPS  int
	frameTime  time.Duration
	deltaTime  float32
	fps        float64
	frameCount uint64

	// Колбэки
	initCallback   func(*Engine) error
	updateCallback func(*Engine, float32)
	renderCallback func(*Engine)
	shutdownCallback func(*Engine)
}

// NewEngine создает новый экземпляр движка
func NewEngine() *Engine {
	return NewEngineWithConfig(DefaultEngineConfig())
}

// NewEngineWithConfig создает движок с заданной конфигурацией
func NewEngineWithConfig(config EngineConfig) *Engine {
	return &Engine{
		config:          config,
		targetFPS:       config.TargetFPS,
		frameTime:       time.Second / time.Duration(config.TargetFPS),
		world:           ecs.NewWorld(),
		eventBus:        event.NewEventBus(1000, 4),
		resourceManager: resource.NewResourceManager(config.LoadWorkers, config.MaxResourceCacheSize),
		inputManager:    input.NewInputManager(),
	}
}

// Initialize инициализирует движок
func (e *Engine) Initialize() error {
	// Создаем окно
	var err error
	e.window, err = window.NewWindow(e.config.WindowConfig)
	if err != nil {
		return fmt.Errorf("failed to create window: %w", err)
	}

	// Настраиваем колбэки ввода
	e.setupInputCallbacks()

	// Запускаем подсистемы
	e.eventBus.Start()
	e.resourceManager.Start()
	e.world.Start()

	// Вызываем пользовательский колбэк инициализации
	if e.initCallback != nil {
		if err := e.initCallback(e); err != nil {
			return fmt.Errorf("init callback failed: %w", err)
		}
	}

	// Отправляем событие инициализации
	e.eventBus.EmitSync(event.NewEvent(event.EventAppInit, nil))

	return nil
}

// setupInputCallbacks настраивает колбэки ввода
func (e *Engine) setupInputCallbacks() {
	e.window.SetKeyCallback(func(key, scancode, action, mods int) {
		e.inputManager.OnKey(key, scancode, action, mods)
		e.eventBus.Emit(event.NewEvent(event.EventKeyPress, &event.KeyEventData{
			Key:      key,
			Scancode: scancode,
			Action:   action,
			Mods:     mods,
		}))
	})

	e.window.SetMouseButtonCallback(func(button, action, mods int) {
		e.inputManager.OnMouseButton(button, action, mods)
		x, y := e.window.GetCursorPos()
		e.eventBus.Emit(event.NewEvent(event.EventMouseButtonPress, &event.MouseButtonData{
			Button: button,
			Action: action,
			Mods:   mods,
			X:      x,
			Y:      y,
		}))
	})

	e.window.SetMouseMoveCallback(func(x, y float64) {
		e.inputManager.OnMouseMove(x, y)
	})

	e.window.SetMouseScrollCallback(func(xOffset, yOffset float64) {
		e.inputManager.OnMouseScroll(xOffset, yOffset)
	})

	e.window.SetResizeCallback(func(width, height int) {
		e.eventBus.Emit(event.NewEvent(event.EventWindowResize, &event.WindowResizeData{
			Width:  width,
			Height: height,
		}))
	})

	e.window.SetCloseCallback(func() {
		e.Stop()
	})
}

// Run запускает главный игровой цикл
func (e *Engine) Run() error {
	if err := e.Initialize(); err != nil {
		return err
	}

	e.running = true
	e.eventBus.EmitSync(event.NewEvent(event.EventAppStart, nil))

	lastTime := time.Now()
	fpsTimer := time.Now()
	fpsCounter := 0

	// Главный игровой цикл
	for e.running && !e.window.ShouldClose() {
		frameStart := time.Now()

		// Вычисляем delta time
		currentTime := time.Now()
		e.deltaTime = float32(currentTime.Sub(lastTime).Seconds())
		lastTime = currentTime

		// Обрабатываем события окна
		e.window.PollEvents()

		// Обновляем ввод
		e.inputManager.Update()

		// Событие начала кадра
		e.eventBus.EmitSync(event.NewEvent(event.EventFrameBegin, nil))

		// Обновляем игровую логику
		e.Update(e.deltaTime)

		// Рендерим
		e.Render()

		// Событие конца кадра
		e.eventBus.EmitSync(event.NewEvent(event.EventFrameEnd, nil))

		// Меняем буферы
		e.window.SwapBuffers()

		// Подсчет FPS
		fpsCounter++
		if time.Since(fpsTimer) >= time.Second {
			e.fps = float64(fpsCounter)
			fpsCounter = 0
			fpsTimer = time.Now()
		}

		e.frameCount++

		// Ограничиваем FPS
		elapsed := time.Since(frameStart)
		if elapsed < e.frameTime {
			time.Sleep(e.frameTime - elapsed)
		}
	}

	e.Shutdown()
	return nil
}

// Update обновляет логику игры
func (e *Engine) Update(deltaTime float32) {
	// Обновляем мир (все системы)
	e.world.Update(deltaTime)

	// Пользовательский колбэк обновления
	if e.updateCallback != nil {
		e.updateCallback(e, deltaTime)
	}
}

// Render рендерит кадр
func (e *Engine) Render() {
	e.eventBus.EmitSync(event.NewEvent(event.EventRenderBegin, nil))

	// Пользовательский колбэк рендеринга
	if e.renderCallback != nil {
		e.renderCallback(e)
	}

	e.eventBus.EmitSync(event.NewEvent(event.EventRenderEnd, nil))
}

// Shutdown завершает работу движка
func (e *Engine) Shutdown() {
	e.eventBus.EmitSync(event.NewEvent(event.EventAppShutdown, nil))

	// Пользовательский колбэк завершения
	if e.shutdownCallback != nil {
		e.shutdownCallback(e)
	}

	// Останавливаем подсистемы
	e.world.Destroy()
	e.resourceManager.Stop()
	e.eventBus.Stop()

	// Закрываем окно
	if e.window != nil {
		e.window.Close()
	}
}

// Stop останавливает игровой цикл
func (e *Engine) Stop() {
	e.running = false
}

// SetInitCallback устанавливает колбэк инициализации
func (e *Engine) SetInitCallback(callback func(*Engine) error) {
	e.initCallback = callback
}

// SetUpdateCallback устанавливает колбэк обновления
func (e *Engine) SetUpdateCallback(callback func(*Engine, float32)) {
	e.updateCallback = callback
}

// SetRenderCallback устанавливает колбэк рендеринга
func (e *Engine) SetRenderCallback(callback func(*Engine)) {
	e.renderCallback = callback
}

// SetShutdownCallback устанавливает колбэк завершения
func (e *Engine) SetShutdownCallback(callback func(*Engine)) {
	e.shutdownCallback = callback
}

// GetWindow возвращает окно
func (e *Engine) GetWindow() *window.Window {
	return e.window
}

// GetWorld возвращает мир
func (e *Engine) GetWorld() *ecs.World {
	return e.world
}

// GetEventBus возвращает шину событий
func (e *Engine) GetEventBus() *event.EventBus {
	return e.eventBus
}

// GetResourceManager возвращает менеджер ресурсов
func (e *Engine) GetResourceManager() *resource.ResourceManager {
	return e.resourceManager
}

// GetInputManager возвращает менеджер ввода
func (e *Engine) GetInputManager() *input.InputManager {
	return e.inputManager
}

// GetFPS возвращает текущий FPS
func (e *Engine) GetFPS() float64 {
	return e.fps
}

// GetDeltaTime возвращает время последнего кадра
func (e *Engine) GetDeltaTime() float32 {
	return e.deltaTime
}

// GetFrameCount возвращает количество отрисованных кадров
func (e *Engine) GetFrameCount() uint64 {
	return e.frameCount
}

// SetTargetFPS устанавливает целевой FPS
func (e *Engine) SetTargetFPS(fps int) {
	e.targetFPS = fps
	e.frameTime = time.Second / time.Duration(fps)
}

// GetTargetFPS возвращает целевой FPS
func (e *Engine) GetTargetFPS() int {
	return e.targetFPS
}
