# AnimoEngine

**AnimoEngine** — высокопроизводительный игровой движок на языке Go, ориентированный на разработку RPG игр с пиксельной и lowpoly графикой.

## Особенности

- **Entity Component System (ECS)** — гибкая архитектура для управления игровыми объектами
- **OpenGL рендеринг** — поддержка 2D и 3D графики с оптимизациями (sprite batching, instanced rendering, frustum culling)
- **HTML/CSS/JS интерфейсы** — современные UI через встроенный webview
- **Кроссплатформенность** — поддержка Linux, Windows, macOS (с планами на Web)
- **RPG системы** — встроенные компоненты для создания RPG игр (инвентарь, диалоги, квесты, боевая система)
- **Высокая производительность** — оптимизации через object pooling, минимизация аллокаций, параллельная обработка

## Системные требования

- Go 1.21 или выше
- OpenGL 3.3+
- Для сборки на Windows: MinGW-w64 или TDM-GCC
- Для сборки на Linux: GCC и необходимые библиотеки разработки

## Быстрый старт

### Установка зависимостей

```bash
go mod download
```

### Запуск демо-приложения

```bash
go run cmd/demo/main.go
```

## Структура проекта

```
AnimoEngine/
├── cmd/
│   └── demo/              # Демонстрационное приложение
├── pkg/
│   ├── core/              # Ядро движка
│   │   ├── ecs/          # Entity Component System
│   │   ├── math/         # Математические утилиты
│   │   ├── event/        # Система событий
│   │   └── resource/     # Управление ресурсами
│   ├── graphics/          # Графическая подсистема
│   │   ├── opengl/       # OpenGL реализация
│   │   ├── batch/        # Sprite batching
│   │   ├── camera/       # Система камер
│   │   └── shader/       # Управление шейдерами
│   ├── platform/          # Платформенная абстракция
│   │   ├── window/       # Оконная система (GLFW)
│   │   ├── input/        # Система ввода
│   │   └── filesystem/   # Файловая система
│   ├── ui/                # Пользовательский интерфейс
│   │   ├── webview/      # Интеграция webview
│   │   └── bindings/     # Go-JS биндинги
│   ├── physics/           # Физическая система
│   │   └── collision/    # Обнаружение коллизий
│   └── game/              # Игровые системы
│       └── rpg/          # RPG-специфичные компоненты
├── assets/                # Игровые ресурсы
│   ├── shaders/          # GLSL шейдеры
│   └── ui/              # HTML/CSS/JS файлы
├── examples/              # Примеры использования
│   └── rpg_demo/        # Демо RPG игры
└── docs/                 # Документация
```

## Пример использования

```go
package main

import (
    "github.com/Salamander5876/AnimoEngine/pkg/core"
    "github.com/Salamander5876/AnimoEngine/pkg/graphics"
)

func main() {
    // Инициализация движка
    engine := core.NewEngine()

    // Создание сцены
    scene := engine.CreateScene("main")

    // Добавление спрайта
    sprite := scene.CreateEntity()
    sprite.AddComponent(&graphics.SpriteComponent{
        Texture: "assets/player.png",
        Width:   32,
        Height:  32,
    })

    // Запуск игрового цикла
    engine.Run()
}
```

## Документация

Полная документация доступна в каталоге [docs/](docs/).

- [Архитектура движка](docs/architecture.md)
- [Руководство по ECS](docs/ecs-guide.md)
- [Графическая подсистема](docs/graphics.md)
- [Создание UI](docs/ui-guide.md)
- [RPG системы](docs/rpg-systems.md)
- [Оптимизация и производительность](docs/performance.md)

## Производительность

- Целевой FPS: 60+ для сцен с 1000+ lowpoly объектов
- Использование памяти: < 200MB для базовой RPG локации
- Время загрузки: < 3 секунд для типичной локации

## Зависимости

- [go-gl/gl](https://github.com/go-gl/gl) — OpenGL биндинги
- [go-gl/glfw](https://github.com/go-gl/glfw) — Оконная система
- [go-gl/mathgl](https://github.com/go-gl/mathgl) — Математическая библиотека
- [webview/webview](https://github.com/webview/webview) — UI через HTML/CSS/JS

## Разработка

### Запуск тестов

```bash
go test ./...
```

### Бенчмарки

```bash
go test -bench=. ./...
```

### Профилирование

```bash
go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=.
go tool pprof cpu.prof
```

## Лицензия

MIT License

## Участие в разработке

Мы приветствуем вклад в развитие проекта! Пожалуйста, ознакомьтесь с [руководством по участию](CONTRIBUTING.md).

## Контакты

- GitHub: [Salamander5876/AnimoEngine](https://github.com/Salamander5876/AnimoEngine)
- Вопросы и предложения: [Issues](https://github.com/Salamander5876/AnimoEngine/issues)

---

**AnimoEngine** разработан с любовью к Go и играм 🎮
