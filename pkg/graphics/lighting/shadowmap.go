package lighting

import (
	"github.com/go-gl/gl/v3.3-core/gl"
)

// ShadowMap карта теней
type ShadowMap struct {
	FBO       uint32 // Framebuffer Object
	Texture   uint32 // Текстура depth map
	Width     int32
	Height    int32
	DepthOnly bool // Только depth, без color attachment
}

// NewShadowMap создаёт новую карту теней
func NewShadowMap(width, height int32) *ShadowMap {
	sm := &ShadowMap{
		Width:     width,
		Height:    height,
		DepthOnly: true,
	}

	// Создаём framebuffer
	gl.GenFramebuffers(1, &sm.FBO)

	// Создаём текстуру для depth map
	gl.GenTextures(1, &sm.Texture)
	gl.BindTexture(gl.TEXTURE_2D, sm.Texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT, width, height, 0, gl.DEPTH_COMPONENT, gl.FLOAT, nil)

	// Настройки текстуры
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_BORDER)

	// Устанавливаем border color - за пределами shadow map всё освещено
	borderColor := []float32{1.0, 1.0, 1.0, 1.0}
	gl.TexParameterfv(gl.TEXTURE_2D, gl.TEXTURE_BORDER_COLOR, &borderColor[0])

	// Привязываем текстуру к framebuffer
	gl.BindFramebuffer(gl.FRAMEBUFFER, sm.FBO)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, sm.Texture, 0)

	// Явно говорим, что не используем color buffer
	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)

	// Проверяем framebuffer
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("Shadow map framebuffer is not complete!")
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	return sm
}

// Bind активирует framebuffer для рендеринга в карту теней
func (sm *ShadowMap) Bind() {
	gl.Viewport(0, 0, sm.Width, sm.Height)
	gl.BindFramebuffer(gl.FRAMEBUFFER, sm.FBO)
	gl.Clear(gl.DEPTH_BUFFER_BIT)
}

// Unbind деактивирует framebuffer
func (sm *ShadowMap) Unbind(viewportWidth, viewportHeight int32) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(0, 0, viewportWidth, viewportHeight)
}

// BindTexture привязывает текстуру shadow map для чтения
func (sm *ShadowMap) BindTexture(textureUnit uint32) {
	gl.ActiveTexture(gl.TEXTURE0 + textureUnit)
	gl.BindTexture(gl.TEXTURE_2D, sm.Texture)
}

// Cleanup освобождает ресурсы
func (sm *ShadowMap) Cleanup() {
	gl.DeleteTextures(1, &sm.Texture)
	gl.DeleteFramebuffers(1, &sm.FBO)
}
