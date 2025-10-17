package shader

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Ошибки шейдерной системы
var (
	ErrShaderCompilation = errors.New("shader compilation failed")
	ErrShaderLinking     = errors.New("shader program linking failed")
	ErrUniformNotFound   = errors.New("uniform not found")
)

// Shader представляет скомпилированную шейдерную программу
type Shader struct {
	ID              uint32
	uniformCache    map[string]int32
}

// NewShader создает и компилирует шейдерную программу
func NewShader(vertexSource, fragmentSource string) (*Shader, error) {
	// Компилируем вершинный шейдер
	vertexShader, err := compileShader(vertexSource, gl.VERTEX_SHADER)
	if err != nil {
		return nil, fmt.Errorf("%w (vertex): %v", ErrShaderCompilation, err)
	}
	defer gl.DeleteShader(vertexShader)

	// Компилируем фрагментный шейдер
	fragmentShader, err := compileShader(fragmentSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, fmt.Errorf("%w (fragment): %v", ErrShaderCompilation, err)
	}
	defer gl.DeleteShader(fragmentShader)

	// Создаем программу
	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	// Проверяем линковку
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return nil, fmt.Errorf("%w: %s", ErrShaderLinking, log)
	}

	return &Shader{
		ID:           program,
		uniformCache: make(map[string]int32),
	}, nil
}

// compileShader компилирует шейдер
func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("%s", log)
	}

	return shader, nil
}

// Use активирует шейдерную программу
func (s *Shader) Use() {
	gl.UseProgram(s.ID)
}

// Delete удаляет шейдерную программу
func (s *Shader) Delete() {
	gl.DeleteProgram(s.ID)
}

// getUniformLocation получает location uniform переменной (с кешированием)
func (s *Shader) getUniformLocation(name string) int32 {
	if location, exists := s.uniformCache[name]; exists {
		return location
	}

	location := gl.GetUniformLocation(s.ID, gl.Str(name+"\x00"))
	s.uniformCache[name] = location
	return location
}

// SetInt устанавливает int uniform
func (s *Shader) SetInt(name string, value int32) {
	gl.Uniform1i(s.getUniformLocation(name), value)
}

// SetFloat устанавливает float uniform
func (s *Shader) SetFloat(name string, value float32) {
	gl.Uniform1f(s.getUniformLocation(name), value)
}

// SetVec2 устанавливает vec2 uniform
func (s *Shader) SetVec2(name string, value mgl32.Vec2) {
	gl.Uniform2f(s.getUniformLocation(name), value.X(), value.Y())
}

// SetVec3 устанавливает vec3 uniform
func (s *Shader) SetVec3(name string, value mgl32.Vec3) {
	gl.Uniform3f(s.getUniformLocation(name), value.X(), value.Y(), value.Z())
}

// SetVec4 устанавливает vec4 uniform
func (s *Shader) SetVec4(name string, value mgl32.Vec4) {
	gl.Uniform4f(s.getUniformLocation(name), value.X(), value.Y(), value.Z(), value.W())
}

// SetMat4 устанавливает mat4 uniform
func (s *Shader) SetMat4(name string, value mgl32.Mat4) {
	gl.UniformMatrix4fv(s.getUniformLocation(name), 1, false, &value[0])
}

// SetBool устанавливает bool uniform
func (s *Shader) SetBool(name string, value bool) {
	var intValue int32
	if value {
		intValue = 1
	}
	gl.Uniform1i(s.getUniformLocation(name), intValue)
}

// Базовые шейдеры для начала работы

// BasicVertexShader базовый вершинный шейдер
const BasicVertexShader = `
#version 330 core

layout (location = 0) in vec3 aPosition;
layout (location = 1) in vec3 aNormal;
layout (location = 2) in vec2 aTexCoord;
layout (location = 3) in vec4 aColor;

out vec3 FragPos;
out vec3 Normal;
out vec2 TexCoord;
out vec4 Color;

uniform mat4 uModel;
uniform mat4 uView;
uniform mat4 uProjection;

void main() {
    FragPos = vec3(uModel * vec4(aPosition, 1.0));
    Normal = mat3(transpose(inverse(uModel))) * aNormal;
    TexCoord = aTexCoord;
    Color = aColor;

    gl_Position = uProjection * uView * vec4(FragPos, 1.0);
}
`

// BasicFragmentShader базовый фрагментный шейдер
const BasicFragmentShader = `
#version 330 core

in vec3 FragPos;
in vec3 Normal;
in vec2 TexCoord;
in vec4 Color;

out vec4 FragColor;

uniform sampler2D uTexture;
uniform bool uUseTexture;
uniform vec4 uColor;

void main() {
    vec4 texColor = uUseTexture ? texture(uTexture, TexCoord) : vec4(1.0);
    FragColor = texColor * Color * uColor;
}
`

// SpriteVertexShader шейдер для 2D спрайтов
const SpriteVertexShader = `
#version 330 core

layout (location = 0) in vec3 aPosition;
layout (location = 1) in vec2 aTexCoord;
layout (location = 2) in vec4 aColor;

out vec2 TexCoord;
out vec4 Color;

uniform mat4 uProjection;

void main() {
    TexCoord = aTexCoord;
    Color = aColor;
    gl_Position = uProjection * vec4(aPosition, 1.0);
}
`

// SpriteFragmentShader шейдер для 2D спрайтов
const SpriteFragmentShader = `
#version 330 core

in vec2 TexCoord;
in vec4 Color;

out vec4 FragColor;

uniform sampler2D uTexture;

void main() {
    FragColor = texture(uTexture, TexCoord) * Color;
}
`
