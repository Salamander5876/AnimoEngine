# Быстрый старт с AnimoEngine

## Установка

### Требования

- Go 1.21 или выше
- Компилятор C (GCC на Linux, MinGW на Windows, Clang на macOS)
- OpenGL 3.3+ поддержка

### Windows

1. Установите [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) или MinGW-w64
2. Клонируйте репозиторий:
```bash
git clone https://github.com/Salamander5876/AnimoEngine.git
cd AnimoEngine
```

3. Установите зависимости:
```bash
go mod download
```

### Linux

1. Установите зависимости:
```bash
# Ubuntu/Debian
sudo apt-get install libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install mesa-libGL-devel libXrandr-devel libXcursor-devel libXinerama-devel libXi-devel
```

2. Клонируйте и установите:
```bash
git clone https://github.com/Salamander5876/AnimoEngine.git
cd AnimoEngine
go mod download
```

### macOS

```bash
git clone https://github.com/Salamander5876/AnimoEngine.git
cd AnimoEngine
go mod download
```

## Первая программа

### Создание простого окна

```go
package main

import (
    "log"

    "github.com/Salamander5876/AnimoEngine/pkg/core"
)

func main() {
    // Создаем движок с настройками по умолчанию
    engine := core.NewEngine()

    // Запускаем
    if err := engine.Run(); err != nil {
        log.Fatal(err)
    }
}
```

### Запуск демо

```bash
go run cmd/demo/main.go
```

Вы должны увидеть окно с вращающимся разноцветным треугольником.

**Управление в демо:**
- `ESC` - выход
- `SPACE` - пауза вращения
- `R` - сброс вращения

## Создание игры

### Шаг 1: Настройка движка

```go
package main

import (
    "github.com/Salamander5876/AnimoEngine/pkg/core"
    "github.com/Salamander5876/AnimoEngine/pkg/platform/window"
)

func main() {
    // Создаем конфигурацию
    config := core.DefaultEngineConfig()
    config.WindowConfig.Title = "Моя первая игра"
    config.WindowConfig.Width = 1920
    config.WindowConfig.Height = 1080
    config.TargetFPS = 60

    engine := core.NewEngineWithConfig(config)
    engine.Run()
}
```

### Шаг 2: Инициализация

```go
func main() {
    engine := core.NewEngine()

    // Устанавливаем колбэк инициализации
    engine.SetInitCallback(func(e *core.Engine) error {
        // Инициализация OpenGL
        gl.Init()

        // Загрузка ресурсов
        // Создание сущностей

        return nil
    })

    engine.Run()
}
```

### Шаг 3: Создание сущности

```go
func createPlayer(engine *core.Engine) ecs.EntityID {
    world := engine.GetWorld()

    // Создаем сущность игрока
    player := world.CreateEntity()

    // Добавляем компоненты
    world.AddComponent(player, &TransformComponent{
        Position: mgl32.Vec3{0, 0, 0},
        Rotation: mgl32.QuatIdent(),
        Scale:    mgl32.Vec3{1, 1, 1},
    })

    world.AddComponent(player, &SpriteComponent{
        TexturePath: "assets/player.png",
        Width:       32,
        Height:      32,
    })

    return player
}
```

### Шаг 4: Создание системы

```go
type MovementSystem struct {
    ecs.BaseSystem
}

func NewMovementSystem() *MovementSystem {
    return &MovementSystem{
        BaseSystem: ecs.NewBaseSystem(0), // Приоритет 0
    }
}

func (s *MovementSystem) Update(dt float32, em *ecs.EntityManager) {
    // Получаем все сущности с Transform компонентом
    entities := em.GetEntitiesWithComponents(transformMask)

    for _, entityID := range entities {
        transform, _ := em.GetComponent(entityID, TransformType)
        t := transform.(*TransformComponent)

        // Обновляем позицию
        t.Position = t.Position.Add(t.Velocity.Mul(dt))
    }
}

// Регистрация системы
world.AddSystem(NewMovementSystem())
```

### Шаг 5: Обработка ввода

```go
engine.SetUpdateCallback(func(e *core.Engine, dt float32) {
    input := e.GetInputManager()

    // Проверка нажатия клавиш
    if input.IsKeyPressed(input.KeyW) {
        // Движение вверх
        playerVelocity.Y += speed * dt
    }

    if input.IsKeyJustPressed(input.KeySpace) {
        // Прыжок (одно нажатие)
        jump()
    }

    // Получение оси (для WASD управления)
    horizontal := input.GetAxis(input.KeyA, input.KeyD)
    vertical := input.GetAxis(input.KeyS, input.KeyW)

    // Движение по осям
    playerVelocity.X = horizontal * speed
    playerVelocity.Y = vertical * speed
})
```

### Шаг 6: Рендеринг

```go
engine.SetRenderCallback(func(e *core.Engine) {
    // Очищаем экран
    gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

    // Используем шейдер
    shader.Use()

    // Устанавливаем uniform переменные
    shader.SetMat4("uProjection", projectionMatrix)
    shader.SetMat4("uView", viewMatrix)

    // Рендерим сущности
    renderEntities(e.GetWorld())
})
```

## Работа с событиями

### Подписка на события

```go
eventBus := engine.GetEventBus()

// Подписываемся на событие урона
eventBus.Subscribe(event.EventPlayerDamage, func(e *event.Event) {
    data := e.Data.(*event.DamageData)
    fmt.Printf("Игрок получил %f урона!\n", data.Amount)
})

// Одноразовая подписка
eventBus.SubscribeOnce(event.EventPlayerLevelUp, func(e *event.Event) {
    fmt.Println("Поздравляем с повышением уровня!")
})
```

### Отправка событий

```go
// Синхронная отправка
eventBus.EmitSync(event.NewEvent(event.EventPlayerDamage, &event.DamageData{
    EntityID: playerID,
    Amount:   25.0,
    DamageType: "fire",
}))

// Асинхронная отправка (через очередь)
eventBus.Emit(event.NewEvent(event.EventItemPickup, itemData))
```

## Загрузка ресурсов

### Синхронная загрузка

```go
rm := engine.GetResourceManager()

// Регистрируем загрузчик
rm.RegisterLoader(&TextureLoader{})

// Загружаем текстуру
textureID, err := rm.LoadSync("assets/player.png", resource.ResourceTypeTexture)
if err != nil {
    log.Fatal(err)
}

// Получаем ресурс
res, _ := rm.Get(textureID)
texture := res.Data.(TextureData)
```

### Асинхронная загрузка

```go
rm.LoadAsync("assets/level1.dat", resource.ResourceTypeScene, func(id resource.ResourceID, err error) {
    if err != nil {
        log.Printf("Ошибка загрузки: %v", err)
        return
    }

    fmt.Println("Уровень загружен!")
    res, _ := rm.Get(id)
    // Используем ресурс
})
```

## Полный пример игры

```go
package main

import (
    "github.com/Salamander5876/AnimoEngine/pkg/core"
    "github.com/Salamander5876/AnimoEngine/pkg/core/ecs"
    "github.com/go-gl/gl/v3.3-core/gl"
)

type Game struct {
    engine   *core.Engine
    player   ecs.EntityID
    shader   *shader.Shader
}

func main() {
    game := &Game{}
    game.engine = core.NewEngine()

    game.engine.SetInitCallback(game.init)
    game.engine.SetUpdateCallback(game.update)
    game.engine.SetRenderCallback(game.render)

    game.engine.Run()
}

func (g *Game) init(e *core.Engine) error {
    gl.Init()

    // Загружаем ресурсы
    // Создаем игрока
    g.player = e.GetWorld().CreateEntity()

    // Добавляем системы
    e.GetWorld().AddSystem(NewMovementSystem())

    return nil
}

func (g *Game) update(e *core.Engine, dt float32) {
    input := e.GetInputManager()

    if input.IsKeyPressed(input.KeyEscape) {
        e.Stop()
    }

    // Логика игры
}

func (g *Game) render(e *core.Engine) {
    gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

    // Рендеринг
}
```

## Отладка

### Вывод FPS

```go
if engine.GetFrameCount() % 60 == 0 {
    fmt.Printf("FPS: %.0f\n", engine.GetFPS())
}
```

### Профилирование

```go
import _ "net/http/pprof"
import "net/http"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

Затем:
```bash
go tool pprof http://localhost:6060/debug/pprof/profile
```

## Следующие шаги

1. Изучите [архитектуру движка](architecture.md)
2. Ознакомьтесь с [ECS системой](ecs-guide.md)
3. Посмотрите [примеры](../examples/)
4. Присоединяйтесь к разработке на [GitHub](https://github.com/Salamander5876/AnimoEngine)

## Частые проблемы

### Windows: "undefined reference to..."

Установите TDM-GCC или MinGW-w64 и убедитесь что компилятор в PATH.

### Linux: "package gl is not in GOROOT"

Установите необходимые dev пакеты:
```bash
sudo apt-get install libgl1-mesa-dev xorg-dev
```

### macOS: "ld: library not found"

Убедитесь что у вас установлены Xcode Command Line Tools:
```bash
xcode-select --install
```

## Поддержка

- [GitHub Issues](https://github.com/Salamander5876/AnimoEngine/issues)
- [Документация](../docs/)
- [Примеры кода](../examples/)
