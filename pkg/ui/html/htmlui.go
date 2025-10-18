package html

import (
	"fmt"
	"strings"

	"github.com/Salamander5876/AnimoEngine/pkg/graphics/text"
	"github.com/go-gl/mathgl/mgl32"
)

// HTMLElement элемент HTML
type HTMLElement struct {
	Tag      string
	ID       string
	Class    string
	Content  string
	Style    map[string]string
	Children []*HTMLElement
	X, Y     float32
	Width    float32
	Height   float32
}

// HTMLRenderer рендерер HTML/CSS
type HTMLRenderer struct {
	textRenderer *text.TextRenderer
	elements     []*HTMLElement
	styles       map[string]map[string]string // селектор -> свойства
}

// NewHTMLRenderer создает новый HTML рендерер
func NewHTMLRenderer() (*HTMLRenderer, error) {
	textRenderer, err := text.NewTextRenderer()
	if err != nil {
		return nil, fmt.Errorf("failed to create text renderer: %v", err)
	}

	return &HTMLRenderer{
		textRenderer: textRenderer,
		elements:     make([]*HTMLElement, 0),
		styles:       make(map[string]map[string]string),
	}, nil
}

// LoadHTML загружает HTML строку
func (hr *HTMLRenderer) LoadHTML(html string) error {
	// Упрощенный парсинг HTML
	// В реальности нужен полноценный HTML парсер
	hr.elements = hr.parseHTML(html)
	return nil
}

// LoadCSS загружает CSS стили
func (hr *HTMLRenderer) LoadCSS(css string) error {
	// Упрощенный парсинг CSS
	hr.styles = hr.parseCSS(css)
	return nil
}

// parseHTML простой парсер HTML
func (hr *HTMLRenderer) parseHTML(html string) []*HTMLElement {
	elements := make([]*HTMLElement, 0)

	// Простая реализация для демо
	// Поддерживает только базовые теги: <div>, <button>, <p>, <h1>
	lines := strings.Split(html, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "<div") {
			el := &HTMLElement{
				Tag:      "div",
				Style:    make(map[string]string),
				Children: make([]*HTMLElement, 0),
			}

			// Извлекаем id и class
			if strings.Contains(line, "id=\"") {
				start := strings.Index(line, "id=\"") + 4
				end := strings.Index(line[start:], "\"")
				el.ID = line[start : start+end]
			}

			if strings.Contains(line, "class=\"") {
				start := strings.Index(line, "class=\"") + 7
				end := strings.Index(line[start:], "\"")
				el.Class = line[start : start+end]
			}

			// Извлекаем содержимое
			if strings.Contains(line, ">") && strings.Contains(line, "</") {
				start := strings.Index(line, ">") + 1
				end := strings.Index(line, "</")
				el.Content = line[start:end]
			}

			elements = append(elements, el)
		} else if strings.HasPrefix(line, "<button") {
			el := &HTMLElement{
				Tag:   "button",
				Style: make(map[string]string),
			}

			if strings.Contains(line, "id=\"") {
				start := strings.Index(line, "id=\"") + 4
				end := strings.Index(line[start:], "\"")
				el.ID = line[start : start+end]
			}

			if strings.Contains(line, ">") && strings.Contains(line, "</") {
				start := strings.Index(line, ">") + 1
				end := strings.Index(line, "</")
				el.Content = line[start:end]
			}

			elements = append(elements, el)
		}
	}

	return elements
}

// parseCSS простой парсер CSS
func (hr *HTMLRenderer) parseCSS(css string) map[string]map[string]string {
	styles := make(map[string]map[string]string)

	// Убираем комментарии и лишние пробелы
	css = strings.TrimSpace(css)

	// Простая реализация для демо
	// Ищем селекторы и их свойства
	parts := strings.Split(css, "}")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Разделяем селектор и свойства
		selectorEnd := strings.Index(part, "{")
		if selectorEnd == -1 {
			continue
		}

		selector := strings.TrimSpace(part[:selectorEnd])
		properties := strings.TrimSpace(part[selectorEnd+1:])

		styleMap := make(map[string]string)

		// Парсим свойства
		props := strings.Split(properties, ";")
		for _, prop := range props {
			prop = strings.TrimSpace(prop)
			if prop == "" {
				continue
			}

			parts := strings.Split(prop, ":")
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			styleMap[key] = value
		}

		styles[selector] = styleMap
	}

	return styles
}

// Render рисует HTML элементы
func (hr *HTMLRenderer) Render(width, height float32) {
	projection := mgl32.Ortho(0, width, 0, height, -1, 1)

	// Применяем стили и рисуем элементы
	currentY := height - 50.0

	for _, el := range hr.elements {
		// Применяем стили по селекторам
		hr.applyStyles(el)

		// Позиционируем элемент
		el.X = 20
		el.Y = currentY
		el.Width = 200
		el.Height = 40

		// Рисуем элемент
		hr.renderElement(el, projection)

		currentY -= el.Height + 10
	}
}

// applyStyles применяет CSS стили к элементу
func (hr *HTMLRenderer) applyStyles(el *HTMLElement) {
	// Применяем стили по тегу
	if tagStyles, ok := hr.styles[el.Tag]; ok {
		for key, value := range tagStyles {
			el.Style[key] = value
		}
	}

	// Применяем стили по классу
	if el.Class != "" {
		classSelector := "." + el.Class
		if classStyles, ok := hr.styles[classSelector]; ok {
			for key, value := range classStyles {
				el.Style[key] = value
			}
		}
	}

	// Применяем стили по ID
	if el.ID != "" {
		idSelector := "#" + el.ID
		if idStyles, ok := hr.styles[idSelector]; ok {
			for key, value := range idStyles {
				el.Style[key] = value
			}
		}
	}
}

// renderElement рисует один элемент
func (hr *HTMLRenderer) renderElement(el *HTMLElement, projection mgl32.Mat4) {
	// Получаем цвет из стилей
	color := mgl32.Vec4{1, 1, 1, 1}

	if colorStr, ok := el.Style["color"]; ok {
		// Простой парсинг цвета (только белый/черный для демо)
		if colorStr == "white" || colorStr == "#ffffff" {
			color = mgl32.Vec4{1, 1, 1, 1}
		} else if colorStr == "black" || colorStr == "#000000" {
			color = mgl32.Vec4{0, 0, 0, 1}
		} else if colorStr == "red" || colorStr == "#ff0000" {
			color = mgl32.Vec4{1, 0, 0, 1}
		} else if colorStr == "blue" || colorStr == "#0000ff" {
			color = mgl32.Vec4{0, 0, 1, 1}
		}
	}

	// Получаем размер шрифта
	fontSize := float32(1.5)
	if sizeStr, ok := el.Style["font-size"]; ok {
		// Простой парсинг размера
		if sizeStr == "16px" {
			fontSize = 1.3
		} else if sizeStr == "18px" {
			fontSize = 1.5
		} else if sizeStr == "20px" {
			fontSize = 1.7
		} else if sizeStr == "24px" {
			fontSize = 2.0
		}
	}

	// Рисуем текст элемента
	if el.Content != "" {
		hr.textRenderer.DrawText(el.Content, el.X, el.Y, fontSize, color, projection)
	}
}

// GetElementByID возвращает элемент по ID
func (hr *HTMLRenderer) GetElementByID(id string) *HTMLElement {
	for _, el := range hr.elements {
		if el.ID == id {
			return el
		}
	}
	return nil
}

// IsElementClicked проверяет был ли клик по элементу
func (hr *HTMLRenderer) IsElementClicked(el *HTMLElement, mouseX, mouseY float32, screenHeight float32) bool {
	// Инвертируем Y координату мыши (OpenGL координаты)
	adjustedY := screenHeight - mouseY

	return mouseX >= el.X && mouseX <= el.X+el.Width &&
		adjustedY >= el.Y && adjustedY <= el.Y+el.Height
}
