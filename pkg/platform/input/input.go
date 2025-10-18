package input

import (
	"sync"

	"github.com/go-gl/glfw/v3.3/glfw"
)

// Константы для клавиш (переэкспорт из GLFW)
const (
	KeyUnknown      = int(glfw.KeyUnknown)
	KeySpace        = int(glfw.KeySpace)
	KeyEscape       = int(glfw.KeyEscape)
	KeyEnter        = int(glfw.KeyEnter)
	KeyTab          = int(glfw.KeyTab)
	KeyBackspace    = int(glfw.KeyBackspace)
	KeyUp           = int(glfw.KeyUp)
	KeyDown         = int(glfw.KeyDown)
	KeyLeft         = int(glfw.KeyLeft)
	KeyRight        = int(glfw.KeyRight)
	KeyA            = int(glfw.KeyA)
	KeyD            = int(glfw.KeyD)
	KeyS            = int(glfw.KeyS)
	KeyW            = int(glfw.KeyW)
	KeyR            = int(glfw.KeyR)
	KeyF            = int(glfw.KeyF)
	KeyT            = int(glfw.KeyT)
	KeyY            = int(glfw.KeyY)
	Key1            = int(glfw.Key1)
	Key2            = int(glfw.Key2)
	Key3            = int(glfw.Key3)
	Key4            = int(glfw.Key4)
	KeyLeftShift    = int(glfw.KeyLeftShift)
	KeyLeftControl  = int(glfw.KeyLeftControl)
	KeyLeftAlt      = int(glfw.KeyLeftAlt)

	MouseButton1    = int(glfw.MouseButton1)
	MouseButton2    = int(glfw.MouseButton2)
	MouseButton3    = int(glfw.MouseButton3)
	MouseButtonLeft = MouseButton1
	MouseButtonRight = MouseButton2
	MouseButtonMiddle = MouseButton3

	Press   = int(glfw.Press)
	Release = int(glfw.Release)
	Repeat  = int(glfw.Repeat)
)

// InputManager управляет вводом с клавиатуры и мыши
type InputManager struct {
	// Состояние клавиш
	keys         map[int]bool
	prevKeys     map[int]bool

	// Состояние кнопок мыши
	mouseButtons map[int]bool
	prevMouseButtons map[int]bool

	// Позиция и движение мыши
	mouseX      float64
	mouseY      float64
	prevMouseX  float64
	prevMouseY  float64
	mouseDeltaX float64
	mouseDeltaY float64

	// Прокрутка мыши
	scrollX float64
	scrollY float64

	mu sync.RWMutex
}

// NewInputManager создает новый менеджер ввода
func NewInputManager() *InputManager {
	return &InputManager{
		keys:             make(map[int]bool),
		prevKeys:         make(map[int]bool),
		mouseButtons:     make(map[int]bool),
		prevMouseButtons: make(map[int]bool),
	}
}

// Update обновляет состояние ввода (вызывается каждый кадр)
func (im *InputManager) Update() {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Сохраняем предыдущее состояние клавиш
	im.prevKeys = make(map[int]bool)
	for k, v := range im.keys {
		im.prevKeys[k] = v
	}

	// Сохраняем предыдущее состояние кнопок мыши
	im.prevMouseButtons = make(map[int]bool)
	for k, v := range im.mouseButtons {
		im.prevMouseButtons[k] = v
	}

	// Обновляем дельту мыши
	im.mouseDeltaX = im.mouseX - im.prevMouseX
	im.mouseDeltaY = im.mouseY - im.prevMouseY
	im.prevMouseX = im.mouseX
	im.prevMouseY = im.mouseY

	// Сбрасываем прокрутку
	im.scrollX = 0
	im.scrollY = 0
}

// OnKey обработчик события клавиатуры
func (im *InputManager) OnKey(key, scancode, action, mods int) {
	im.mu.Lock()
	defer im.mu.Unlock()

	if action == Press || action == Repeat {
		im.keys[key] = true
	} else if action == Release {
		im.keys[key] = false
	}
}

// OnMouseButton обработчик события кнопок мыши
func (im *InputManager) OnMouseButton(button, action, mods int) {
	im.mu.Lock()
	defer im.mu.Unlock()

	if action == Press {
		im.mouseButtons[button] = true
	} else if action == Release {
		im.mouseButtons[button] = false
	}
}

// OnMouseMove обработчик движения мыши
func (im *InputManager) OnMouseMove(x, y float64) {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.mouseX = x
	im.mouseY = y
}

// OnMouseScroll обработчик прокрутки мыши
func (im *InputManager) OnMouseScroll(xOffset, yOffset float64) {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.scrollX = xOffset
	im.scrollY = yOffset
}

// IsKeyPressed возвращает true, если клавиша нажата
func (im *InputManager) IsKeyPressed(key int) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return im.keys[key]
}

// IsKeyJustPressed возвращает true, если клавиша была нажата в этом кадре
func (im *InputManager) IsKeyJustPressed(key int) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return im.keys[key] && !im.prevKeys[key]
}

// IsKeyJustReleased возвращает true, если клавиша была отпущена в этом кадре
func (im *InputManager) IsKeyJustReleased(key int) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return !im.keys[key] && im.prevKeys[key]
}

// IsMouseButtonPressed возвращает true, если кнопка мыши нажата
func (im *InputManager) IsMouseButtonPressed(button int) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return im.mouseButtons[button]
}

// IsMouseButtonJustPressed возвращает true, если кнопка мыши была нажата в этом кадре
func (im *InputManager) IsMouseButtonJustPressed(button int) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return im.mouseButtons[button] && !im.prevMouseButtons[button]
}

// IsMouseButtonJustReleased возвращает true, если кнопка мыши была отпущена в этом кадре
func (im *InputManager) IsMouseButtonJustReleased(button int) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return !im.mouseButtons[button] && im.prevMouseButtons[button]
}

// GetMousePosition возвращает позицию мыши
func (im *InputManager) GetMousePosition() (x, y float64) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return im.mouseX, im.mouseY
}

// GetMouseDelta возвращает смещение мыши с прошлого кадра
func (im *InputManager) GetMouseDelta() (dx, dy float64) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return im.mouseDeltaX, im.mouseDeltaY
}

// GetScroll возвращает значение прокрутки мыши
func (im *InputManager) GetScroll() (x, y float64) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return im.scrollX, im.scrollY
}

// GetAxis возвращает значение оси (-1, 0, или 1)
// Например, для WASD: GetAxis(KeyA, KeyD) вернет -1 если A нажата, 1 если D, 0 если обе или ни одной
func (im *InputManager) GetAxis(negative, positive int) float32 {
	im.mu.RLock()
	defer im.mu.RUnlock()

	value := float32(0)
	if im.keys[negative] {
		value -= 1
	}
	if im.keys[positive] {
		value += 1
	}
	return value
}

// Clear очищает все состояния ввода
func (im *InputManager) Clear() {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.keys = make(map[int]bool)
	im.prevKeys = make(map[int]bool)
	im.mouseButtons = make(map[int]bool)
	im.prevMouseButtons = make(map[int]bool)
	im.mouseDeltaX = 0
	im.mouseDeltaY = 0
	im.scrollX = 0
	im.scrollY = 0
}
