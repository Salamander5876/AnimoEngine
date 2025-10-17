package core

import (
	"time"

	"github.com/Salamander5876/AnimoEngine/pkg/graphics"
	"github.com/Salamander5876/AnimoEngine/pkg/graphics/shader"
	"github.com/go-gl/gl/v3.3-core/gl"
)

// SplashScreen показывает логотип при запуске
type SplashScreen struct {
	texture       *graphics.Texture
	shader        *shader.Shader
	vao           uint32
	vbo           uint32
	duration      time.Duration
	fadeInTime    time.Duration
	fadeOutTime   time.Duration
	startTime     time.Time
	alpha         float32
}

// NewSplashScreen создает новый splash screen
func NewSplashScreen(logoPath string, duration time.Duration) (*SplashScreen, error) {
	// Загружаем текстуру логотипа
	texture, err := graphics.LoadTexture(logoPath)
	if err != nil {
		return nil, err
	}

	// Создаем шейдер для отображения логотипа
	vertexShader := `
	#version 330 core

	layout (location = 0) in vec2 aPosition;
	layout (location = 1) in vec2 aTexCoord;

	out vec2 TexCoord;

	void main() {
		TexCoord = aTexCoord;
		gl_Position = vec4(aPosition, 0.0, 1.0);
	}
	`

	fragmentShader := `
	#version 330 core

	in vec2 TexCoord;
	out vec4 FragColor;

	uniform sampler2D uTexture;
	uniform float uAlpha;

	void main() {
		vec4 texColor = texture(uTexture, TexCoord);
		FragColor = vec4(texColor.rgb, texColor.a * uAlpha);
	}
	`

	splashShader, err := shader.NewShader(vertexShader, fragmentShader)
	if err != nil {
		texture.Delete()
		return nil, err
	}

	// Создаем quad для отображения логотипа (центрированный)
	// Вычисляем aspect ratio для правильного отображения
	aspectRatio := float32(texture.Width) / float32(texture.Height)
	width := float32(0.5)
	height := width / aspectRatio

	vertices := []float32{
		// Позиции        // Текстурные координаты
		-width, -height,  0.0, 1.0, // Левый нижний
		width, -height,   1.0, 1.0, // Правый нижний
		width, height,    1.0, 0.0, // Правый верхний
		-width, height,   0.0, 0.0, // Левый верхний
	}

	indices := []uint32{
		0, 1, 2,
		2, 3, 0,
	}

	// Создаем VAO и VBO
	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// Позиция
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Текстурные координаты
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)

	return &SplashScreen{
		texture:     texture,
		shader:      splashShader,
		vao:         vao,
		vbo:         vbo,
		duration:    duration,
		fadeInTime:  500 * time.Millisecond,
		fadeOutTime: 500 * time.Millisecond,
		alpha:       0.0,
	}, nil
}

// Show показывает splash screen на указанное время
func (s *SplashScreen) Show(engine *Engine) {
	s.startTime = time.Now()

	// Включаем blend для прозрачности
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	for time.Since(s.startTime) < s.duration {
		elapsed := time.Since(s.startTime)

		// Вычисляем alpha для fade in/out
		if elapsed < s.fadeInTime {
			s.alpha = float32(elapsed) / float32(s.fadeInTime)
		} else if elapsed > s.duration-s.fadeOutTime {
			remaining := s.duration - elapsed
			s.alpha = float32(remaining) / float32(s.fadeOutTime)
		} else {
			s.alpha = 1.0
		}

		// Очищаем экран
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Рендерим логотип
		s.shader.Use()
		s.shader.SetFloat("uAlpha", s.alpha)
		s.shader.SetInt("uTexture", 0)

		gl.ActiveTexture(gl.TEXTURE0)
		s.texture.Bind()

		gl.BindVertexArray(s.vao)
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)
		gl.BindVertexArray(0)

		s.texture.Unbind()

		// Обновляем окно
		engine.window.SwapBuffers()
		engine.window.PollEvents()

		// Небольшая задержка для плавности
		time.Sleep(16 * time.Millisecond) // ~60 FPS
	}

	gl.Disable(gl.BLEND)
}

// Cleanup освобождает ресурсы
func (s *SplashScreen) Cleanup() {
	if s.texture != nil {
		s.texture.Delete()
	}
	if s.shader != nil {
		s.shader.Delete()
	}
	gl.DeleteVertexArrays(1, &s.vao)
	gl.DeleteBuffers(1, &s.vbo)
}
