package window

import (
	"errors"
	"fmt"

	"github.com/go-gl/glfw/v3.3/glfw"
)

// Ошибки оконной системы
var (
	ErrWindowNotInitialized = errors.New("window not initialized")
	ErrGLFWInit             = errors.New("failed to initialize GLFW")
	ErrWindowCreation       = errors.New("failed to create window")
)

// WindowConfig конфигурация окна
type WindowConfig struct {
	Title      string
	Width      int
	Height     int
	Fullscreen bool
	VSync      bool
	Resizable  bool
	MSAA       int // Количество сэмплов для MSAA (0 = выключено)
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() WindowConfig {
	return WindowConfig{
		Title:      "AnimoEngine",
		Width:      1280,
		Height:     720,
		Fullscreen: false,
		VSync:      true,
		Resizable:  true,
		MSAA:       4,
	}
}

// Window представляет игровое окно
type Window struct {
	handle *glfw.Window
	config WindowConfig

	// Колбэки
	resizeCallback    func(width, height int)
	closeCallback     func()
	keyCallback       func(key, scancode, action, mods int)
	mouseButtonCallback func(button, action, mods int)
	mouseMoveCallback func(x, y float64)
	mouseScrollCallback func(xOffset, yOffset float64)
}

// NewWindow создает новое окно
func NewWindow(config WindowConfig) (*Window, error) {
	// Инициализируем GLFW
	if err := glfw.Init(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGLFWInit, err)
	}

	// Настраиваем OpenGL контекст
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Настройки окна
	if config.Resizable {
		glfw.WindowHint(glfw.Resizable, glfw.True)
	} else {
		glfw.WindowHint(glfw.Resizable, glfw.False)
	}

	if config.MSAA > 0 {
		glfw.WindowHint(glfw.Samples, config.MSAA)
	}

	// Создаем окно
	var monitor *glfw.Monitor
	if config.Fullscreen {
		monitor = glfw.GetPrimaryMonitor()
	}

	window, err := glfw.CreateWindow(config.Width, config.Height, config.Title, monitor, nil)
	if err != nil {
		glfw.Terminate()
		return nil, fmt.Errorf("%w: %v", ErrWindowCreation, err)
	}

	window.MakeContextCurrent()

	// VSync
	if config.VSync {
		glfw.SwapInterval(1)
	} else {
		glfw.SwapInterval(0)
	}

	w := &Window{
		handle: window,
		config: config,
	}

	// Устанавливаем колбэки
	w.setupCallbacks()

	return w, nil
}

// setupCallbacks устанавливает GLFW колбэки
func (w *Window) setupCallbacks() {
	// Изменение размера окна
	w.handle.SetSizeCallback(func(window *glfw.Window, width, height int) {
		if w.resizeCallback != nil {
			w.resizeCallback(width, height)
		}
	})

	// Закрытие окна
	w.handle.SetCloseCallback(func(window *glfw.Window) {
		if w.closeCallback != nil {
			w.closeCallback()
		}
	})

	// Клавиатура
	w.handle.SetKeyCallback(func(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if w.keyCallback != nil {
			w.keyCallback(int(key), scancode, int(action), int(mods))
		}
	})

	// Кнопки мыши
	w.handle.SetMouseButtonCallback(func(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		if w.mouseButtonCallback != nil {
			w.mouseButtonCallback(int(button), int(action), int(mods))
		}
	})

	// Движение мыши
	w.handle.SetCursorPosCallback(func(window *glfw.Window, xpos, ypos float64) {
		if w.mouseMoveCallback != nil {
			w.mouseMoveCallback(xpos, ypos)
		}
	})

	// Прокрутка мыши
	w.handle.SetScrollCallback(func(window *glfw.Window, xoff, yoff float64) {
		if w.mouseScrollCallback != nil {
			w.mouseScrollCallback(xoff, yoff)
		}
	})
}

// PollEvents обрабатывает события окна
func (w *Window) PollEvents() {
	glfw.PollEvents()
}

// SwapBuffers меняет буферы отрисовки
func (w *Window) SwapBuffers() {
	w.handle.SwapBuffers()
}

// ShouldClose возвращает true, если окно должно закрыться
func (w *Window) ShouldClose() bool {
	return w.handle.ShouldClose()
}

// SetShouldClose устанавливает флаг закрытия окна
func (w *Window) SetShouldClose(value bool) {
	w.handle.SetShouldClose(value)
}

// Close закрывает окно
func (w *Window) Close() {
	if w.handle != nil {
		w.handle.Destroy()
		w.handle = nil
	}
	glfw.Terminate()
}

// GetSize возвращает размер окна
func (w *Window) GetSize() (width, height int) {
	return w.handle.GetSize()
}

// GetFramebufferSize возвращает размер framebuffer (может отличаться на HiDPI дисплеях)
func (w *Window) GetFramebufferSize() (width, height int) {
	return w.handle.GetFramebufferSize()
}

// SetTitle устанавливает заголовок окна
func (w *Window) SetTitle(title string) {
	w.config.Title = title
	w.handle.SetTitle(title)
}

// GetTitle возвращает заголовок окна
func (w *Window) GetTitle() string {
	return w.config.Title
}

// SetSize устанавливает размер окна
func (w *Window) SetSize(width, height int) {
	w.config.Width = width
	w.config.Height = height
	w.handle.SetSize(width, height)
}

// GetHandle возвращает нативный GLFW handle
func (w *Window) GetHandle() *glfw.Window {
	return w.handle
}

// SetVSync включает/выключает VSync
func (w *Window) SetVSync(enabled bool) {
	w.config.VSync = enabled
	if enabled {
		glfw.SwapInterval(1)
	} else {
		glfw.SwapInterval(0)
	}
}

// GetVSync возвращает состояние VSync
func (w *Window) GetVSync() bool {
	return w.config.VSync
}

// SetFullscreen переключает полноэкранный режим
func (w *Window) SetFullscreen(fullscreen bool) {
	if w.config.Fullscreen == fullscreen {
		return
	}

	w.config.Fullscreen = fullscreen

	var monitor *glfw.Monitor
	if fullscreen {
		monitor = glfw.GetPrimaryMonitor()
		mode := monitor.GetVideoMode()
		w.handle.SetMonitor(monitor, 0, 0, mode.Width, mode.Height, mode.RefreshRate)
	} else {
		w.handle.SetMonitor(nil, 100, 100, w.config.Width, w.config.Height, 0)
	}
}

// IsFullscreen возвращает true, если окно в полноэкранном режиме
func (w *Window) IsFullscreen() bool {
	return w.config.Fullscreen
}

// SetResizeCallback устанавливает колбэк изменения размера
func (w *Window) SetResizeCallback(callback func(width, height int)) {
	w.resizeCallback = callback
}

// SetCloseCallback устанавливает колбэк закрытия окна
func (w *Window) SetCloseCallback(callback func()) {
	w.closeCallback = callback
}

// SetKeyCallback устанавливает колбэк клавиатуры
func (w *Window) SetKeyCallback(callback func(key, scancode, action, mods int)) {
	w.keyCallback = callback
}

// SetMouseButtonCallback устанавливает колбэк кнопок мыши
func (w *Window) SetMouseButtonCallback(callback func(button, action, mods int)) {
	w.mouseButtonCallback = callback
}

// SetMouseMoveCallback устанавливает колбэк движения мыши
func (w *Window) SetMouseMoveCallback(callback func(x, y float64)) {
	w.mouseMoveCallback = callback
}

// SetMouseScrollCallback устанавливает колбэк прокрутки мыши
func (w *Window) SetMouseScrollCallback(callback func(xOffset, yOffset float64)) {
	w.mouseScrollCallback = callback
}

// GetTime возвращает время с момента инициализации GLFW
func (w *Window) GetTime() float64 {
	return glfw.GetTime()
}

// SetTime устанавливает время GLFW
func (w *Window) SetTime(time float64) {
	glfw.SetTime(time)
}

// GetCursorPos возвращает позицию курсора
func (w *Window) GetCursorPos() (x, y float64) {
	return w.handle.GetCursorPos()
}

// SetCursorPos устанавливает позицию курсора
func (w *Window) SetCursorPos(x, y float64) {
	w.handle.SetCursorPos(x, y)
}

// SetCursorMode устанавливает режим курсора
func (w *Window) SetCursorMode(mode int) {
	w.handle.SetInputMode(glfw.CursorMode, mode)
}

// GetKey возвращает состояние клавиши
func (w *Window) GetKey(key int) int {
	return int(w.handle.GetKey(glfw.Key(key)))
}

// GetMouseButton возвращает состояние кнопки мыши
func (w *Window) GetMouseButton(button int) int {
	return int(w.handle.GetMouseButton(glfw.MouseButton(button)))
}

// MakeContextCurrent делает OpenGL контекст текущим
func (w *Window) MakeContextCurrent() {
	w.handle.MakeContextCurrent()
}
