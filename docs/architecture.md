# Архитектура AnimoEngine

## Обзор

AnimoEngine построен на модульной архитектуре с четким разделением ответственности между компонентами. Движок использует современные паттерны проектирования и оптимизирован для высокой производительности.

## Основные модули

### 1. Core (pkg/core/)

Ядро движка, содержащее базовые системы:

#### Entity Component System (pkg/core/ecs/)
- **EntityManager** - управляет всеми игровыми сущностями
- **ComponentManager** - хранит и управляет компонентами
- **SystemManager** - координирует работу игровых систем
- **World** - объединяет все ECS компоненты в единый мир
- **ArchetypeManager** - оптимизирует запросы через группировку сущностей

**Особенности реализации:**
- Использование битовых масок для быстрой фильтрации компонентов
- Object pooling через `sync.Pool` для минимизации GC
- Архетипы для эффективной итерации по сущностям
- Thread-safe операции с использованием RWMutex

#### Event System (pkg/core/event/)
- **EventBus** - pub/sub система событий
- Приоритеты для обработчиков
- Возможность отмены событий
- Асинхронная и синхронная обработка
- Одноразовые подписки (SubscribeOnce)

**Паттерн работы:**
```go
// Подписка на событие
eventBus.Subscribe(event.EventPlayerDamage, func(e *event.Event) {
    data := e.Data.(*event.DamageData)
    // Обработка урона
})

// Отправка события
eventBus.Emit(event.NewEvent(event.EventPlayerDamage, damageData))
```

#### Resource Manager (pkg/core/resource/)
- Централизованное управление ресурсами
- Подсчет ссылок (reference counting)
- Асинхронная загрузка через worker pool
- Кеширование с ограничением по памяти
- Автоматическая выгрузка неиспользуемых ресурсов

**Поддерживаемые типы:**
- Текстуры
- Меши
- Шейдеры
- Аудио
- Шрифты
- Сцены

#### Math (pkg/core/math/)
- AABB для collision detection
- Transform для пространственных преобразований
- Утилиты: lerp, clamp, raycasting
- Ray и Plane для физики

### 2. Graphics (pkg/graphics/)

Графическая подсистема с абстракцией API:

#### Shader System (pkg/graphics/shader/)
- Компиляция и линковка шейдеров
- Кеширование uniform locations
- Базовые шейдеры (Basic, Sprite)
- Type-safe uniform setters

#### Rendering Pipeline
```
GraphicsAPI (интерфейс)
    ↓
OpenGLBackend (реализация)
    ↓
Shader → VAO/VBO → Draw Calls
```

**Планируемые оптимизации:**
- Sprite batching для 2D
- Instanced rendering для 3D
- Frustum culling
- LOD система

### 3. Platform (pkg/platform/)

Абстракция платформы для кроссплатформенности:

#### Window System (pkg/platform/window/)
- GLFW обертка
- Управление OpenGL контекстом
- Колбэки событий (resize, close, input)
- Полноэкранный режим
- VSync управление

#### Input System (pkg/platform/input/)
- Обработка клавиатуры и мыши
- Состояния: pressed, just pressed, just released
- Оси ввода (GetAxis)
- Delta движения мыши

### 4. Engine (pkg/core/engine.go)

Главный класс движка, объединяющий все подсистемы:

**Жизненный цикл:**
```
Initialize → Run → [Game Loop] → Shutdown
                    ↓
            PollEvents → Update → Render → SwapBuffers
```

**Игровой цикл:**
1. Обработка событий окна
2. Обновление ввода
3. Событие начала кадра
4. Обновление игровой логики (ECS systems)
5. Рендеринг
6. Событие конца кадра
7. Ограничение FPS

## Потоки выполнения

### Основной поток
- Игровой цикл
- Рендеринг
- Обработка событий

### Worker потоки
- Загрузка ресурсов (ResourceManager)
- Обработка событий (EventBus)

## Управление памятью

### Оптимизации
1. **Object Pooling**
   - Entity в EntityManager
   - События в EventBus

2. **Кеширование**
   - Uniform locations в шейдерах
   - Ресурсы в ResourceManager

3. **Минимизация аллокаций**
   - Переиспользование слайсов
   - sync.Pool для временных объектов

## Расширяемость

### Добавление новых компонентов
```go
type HealthComponent struct {
    Current float32
    Max     float32
}

func (h *HealthComponent) Type() ecs.ComponentType {
    return healthComponentType
}
```

### Создание систем
```go
type HealthSystem struct {
    ecs.BaseSystem
}

func (s *HealthSystem) Update(dt float32, em *ecs.EntityManager) {
    // Логика обработки здоровья
}
```

### Регистрация загрузчиков ресурсов
```go
type TextureLoader struct{}

func (l *TextureLoader) Load(path string) (interface{}, error) {
    // Загрузка текстуры
}

func (l *TextureLoader) GetType() resource.ResourceType {
    return resource.ResourceTypeTexture
}
```

## Паттерны проектирования

1. **Entity Component System** - для гибкой композиции игровых объектов
2. **Observer (Pub/Sub)** - система событий
3. **Object Pool** - переиспользование объектов
4. **Strategy** - абстракция графического API
5. **Facade** - Engine как единая точка входа

## Зависимости между модулями

```
Engine
  ├── Window (platform)
  ├── Input (platform)
  ├── World (ECS)
  ├── EventBus
  └── ResourceManager

World
  ├── EntityManager
  ├── SystemManager
  └── ArchetypeManager

Graphics
  ├── Shader
  └── GraphicsAPI
```

## Производительность

### Целевые метрики
- 60 FPS для сцен с 1000+ объектов
- < 200MB памяти для базовой сцены
- < 100 draw calls для оптимизированной сцены

### Профилирование
```bash
go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=.
go tool pprof cpu.prof
```

## Будущие улучшения

### Фаза 2
- Sprite batching
- Texture atlas система
- Camera система с culling

### Фаза 3
- WebView интеграция для UI
- Go-JS биндинги

### Фаза 4
- 3D model loading (OBJ, GLTF)
- Advanced lighting
- Shadow mapping

### Фаза 5
- RPG системы (инвентарь, квесты, диалоги)
- Сохранение/загрузка состояния
- Networking (опционально)

## Best Practices

1. **Используйте ECS паттерн** - избегайте больших монолитных классов
2. **Минимизируйте аллокации** - используйте пулы объектов
3. **Профилируйте** - оптимизируйте на основе данных
4. **Разделяйте логику и рендеринг** - используйте системы
5. **Обрабатывайте ошибки** - не игнорируйте `error`

## Документация кода

Все публичные функции документированы согласно godoc конвенциям:

```go
// CreateEntity создает новую сущность в мире.
// Возвращает уникальный EntityID для дальнейшей работы с сущностью.
func (w *World) CreateEntity() EntityID
```

## Тестирование

Структура тестов:
```
pkg/
  core/
    ecs/
      entity_test.go
      component_test.go
      system_test.go
```

Запуск тестов:
```bash
go test ./...
go test -v ./pkg/core/ecs/
go test -cover ./...
```
