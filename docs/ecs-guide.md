# Руководство по Entity Component System

## Введение

Entity Component System (ECS) - это архитектурный паттерн, используемый в AnimoEngine для организации игровой логики. ECS разделяет данные (компоненты) и поведение (системы), обеспечивая высокую производительность и гибкость.

## Основные концепции

### Entity (Сущность)

Сущность - это уникальный идентификатор игрового объекта. Сама по себе сущность не содержит данных или логики.

```go
// Создание сущности
entityID := world.CreateEntity()

// Удаление сущности
world.DestroyEntity(entityID)
```

### Component (Компонент)

Компонент - это структура данных, которая добавляется к сущности. Компоненты содержат только данные, без логики.

```go
// Определение компонента
type TransformComponent struct {
    Position mgl32.Vec3
    Rotation mgl32.Quat
    Scale    mgl32.Vec3
}

func (t *TransformComponent) Type() ecs.ComponentType {
    return TransformComponentType
}

// Добавление компонента к сущности
world.AddComponent(entityID, &TransformComponent{
    Position: mgl32.Vec3{0, 0, 0},
    Rotation: mgl32.QuatIdent(),
    Scale:    mgl32.Vec3{1, 1, 1},
})

// Получение компонента
comp, err := world.GetComponent(entityID, TransformComponentType)
if err == nil {
    transform := comp.(*TransformComponent)
    // Использование компонента
}

// Проверка наличия компонента
if world.HasComponent(entityID, TransformComponentType) {
    // Компонент существует
}

// Удаление компонента
world.RemoveComponent(entityID, TransformComponentType)
```

### System (Система)

Система содержит логику, которая работает с компонентами. Системы обновляются каждый кадр.

```go
// Определение системы
type MovementSystem struct {
    ecs.BaseSystem
}

func NewMovementSystem() *MovementSystem {
    return &MovementSystem{
        BaseSystem: ecs.NewBaseSystem(0), // Приоритет 0
    }
}

func (s *MovementSystem) Update(deltaTime float32, em *ecs.EntityManager) {
    // Получаем все сущности с Transform компонентом
    entities := em.GetEntitiesWithComponents(transformMask)

    for _, entityID := range entities {
        comp, _ := em.GetComponent(entityID, TransformComponentType)
        transform := comp.(*TransformComponent)

        // Обновляем позицию
        // ...
    }
}

// Регистрация системы
world.AddSystem(NewMovementSystem())
```

## Примеры использования

### Создание игрока

```go
func CreatePlayer(world *ecs.World) ecs.EntityID {
    player := world.CreateEntity()

    // Transform
    world.AddComponent(player, &TransformComponent{
        Position: mgl32.Vec3{0, 0, 0},
        Rotation: mgl32.QuatIdent(),
        Scale:    mgl32.Vec3{1, 1, 1},
    })

    // Velocity для движения
    world.AddComponent(player, &VelocityComponent{
        Linear:  mgl32.Vec3{0, 0, 0},
        Angular: mgl32.Vec3{0, 0, 0},
    })

    // Sprite для отображения
    world.AddComponent(player, &SpriteComponent{
        TexturePath: "assets/player.png",
        Width:       32,
        Height:      32,
    })

    // RPG компоненты
    world.AddComponent(player, &rpg.HealthComponent{
        Current: 100,
        Max:     100,
        Regeneration: 5.0,
    })

    world.AddComponent(player, &rpg.StatsComponent{
        Level:        1,
        Strength:     10,
        Agility:      10,
        Intelligence: 10,
        Vitality:     10,
        Luck:         5,
    })

    return player
}
```

### Система физики

```go
type PhysicsSystem struct {
    ecs.BaseSystem
}

func NewPhysicsSystem() *PhysicsSystem {
    return &PhysicsSystem{
        BaseSystem: ecs.NewBaseSystem(1), // Высокий приоритет
    }
}

func (s *PhysicsSystem) Update(deltaTime float32, em *ecs.EntityManager) {
    // Получаем все сущности с Transform и Velocity
    mask := (1 << TransformComponentType) | (1 << VelocityComponentType)
    entities := em.GetEntitiesWithComponents(mask)

    for _, entityID := range entities {
        transformComp, _ := em.GetComponent(entityID, TransformComponentType)
        transform := transformComp.(*TransformComponent)

        velocityComp, _ := em.GetComponent(entityID, VelocityComponentType)
        velocity := velocityComp.(*VelocityComponent)

        // Обновляем позицию на основе скорости
        transform.Position = transform.Position.Add(velocity.Linear.Mul(deltaTime))

        // Применяем гравитацию если есть компонент
        if em.HasComponent(entityID, GravityComponentType) {
            velocity.Linear = velocity.Linear.Add(mgl32.Vec3{0, -9.8, 0}.Mul(deltaTime))
        }
    }
}
```

### Система рендеринга

```go
type RenderSystem struct {
    ecs.BaseSystem
    shader *shader.Shader
    camera *Camera
}

func (s *RenderSystem) Update(deltaTime float32, em *ecs.EntityManager) {
    // Получаем все сущности со спрайтами
    mask := (1 << TransformComponentType) | (1 << SpriteComponentType)
    entities := em.GetEntitiesWithComponents(mask)

    s.shader.Use()
    s.shader.SetMat4("uProjection", s.camera.GetProjectionMatrix())
    s.shader.SetMat4("uView", s.camera.GetViewMatrix())

    for _, entityID := range entities {
        transformComp, _ := em.GetComponent(entityID, TransformComponentType)
        transform := transformComp.(*TransformComponent)

        spriteComp, _ := em.GetComponent(entityID, SpriteComponentType)
        sprite := spriteComp.(*SpriteComponent)

        // Рендерим спрайт
        s.shader.SetMat4("uModel", transform.Matrix())
        s.renderSprite(sprite)
    }
}
```

## Работа с World

### Создание и управление миром

```go
// Создание мира
world := ecs.NewWorld()

// Добавление систем
world.AddSystem(NewPhysicsSystem())
world.AddSystem(NewRenderSystem())
world.AddSystem(NewAISystem())

// Запуск мира
world.Start()

// Обновление (в игровом цикле)
world.Update(deltaTime)

// Пауза
world.Pause()
world.Resume()

// Остановка
world.Stop()

// Очистка
world.Clear() // Удаляет все сущности
world.Destroy() // Полное уничтожение мира
```

### Запросы сущностей

```go
// Получить все сущности
entities := world.GetAllEntities()

// Получить сущности с определенными компонентами
mask := (1 << TransformComponentType) | (1 << HealthComponentType)
entities := world.GetEntitiesWithComponents(mask)

// Использование Query API
query := world.NewQuery()
entities := query.
    With(TransformComponentType).
    With(HealthComponentType).
    Execute()
```

## Приоритеты систем

Системы выполняются в порядке приоритета (меньше число = раньше выполняется):

```go
// Приоритет 0 - выполняется первой
physicsSystem := NewPhysicsSystem()
physicsSystem.BaseSystem = ecs.NewBaseSystem(0)

// Приоритет 5 - выполняется после физики
aiSystem := NewAISystem()
aiSystem.BaseSystem = ecs.NewBaseSystem(5)

// Приоритет 10 - выполняется последней
renderSystem := NewRenderSystem()
renderSystem.BaseSystem = ecs.NewBaseSystem(10)
```

## Архетипы

Архетипы - это оптимизация для группировки сущностей с одинаковым набором компонентов.

```go
// Получение архетипа
mask := (1 << TransformComponentType) | (1 << SpriteComponentType)
archetype := world.GetArchetypeManager().GetOrCreateArchetype(mask)

// Получение сущностей архетипа
entities := archetype.GetEntities()
```

## Best Practices

### 1. Разделяйте данные и логику

**Плохо:**
```go
type PlayerComponent struct {
    Health int
    Speed  float32
}

func (p *PlayerComponent) Update(dt float32) {
    // Логика в компоненте - плохая практика
}
```

**Хорошо:**
```go
type PlayerComponent struct {
    Health int
    Speed  float32
}

type PlayerSystem struct {
    ecs.BaseSystem
}

func (s *PlayerSystem) Update(dt float32, em *ecs.EntityManager) {
    // Логика в системе
}
```

### 2. Используйте маленькие компоненты

**Плохо:**
```go
type CharacterComponent struct {
    Position     mgl32.Vec3
    Rotation     mgl32.Quat
    Health       int
    Mana         int
    Inventory    []Item
    Quests       []Quest
    // Слишком много данных в одном компоненте
}
```

**Хорошо:**
```go
type TransformComponent struct {
    Position mgl32.Vec3
    Rotation mgl32.Quat
}

type HealthComponent struct {
    Current int
    Max     int
}

type InventoryComponent struct {
    Items []Item
}
// Каждый компонент отвечает за одну вещь
```

### 3. Избегайте взаимозависимостей систем

Используйте события для коммуникации между системами:

```go
// Плохо: прямая зависимость
type CombatSystem struct {
    healthSystem *HealthSystem
}

// Хорошо: через события
type CombatSystem struct {
    eventBus *event.EventBus
}

func (s *CombatSystem) DealDamage(target ecs.EntityID, damage float32) {
    s.eventBus.Emit(event.NewEvent(event.EventPlayerDamage, &event.DamageData{
        EntityID: target,
        Amount:   damage,
    }))
}
```

### 4. Используйте object pooling

```go
type BulletPool struct {
    bullets []ecs.EntityID
    world   *ecs.World
}

func (p *BulletPool) Get() ecs.EntityID {
    if len(p.bullets) > 0 {
        bullet := p.bullets[len(p.bullets)-1]
        p.bullets = p.bullets[:len(p.bullets)-1]
        return bullet
    }

    return p.createBullet()
}

func (p *BulletPool) Return(bullet ecs.EntityID) {
    p.bullets = append(p.bullets, bullet)
}
```

## Производительность

### Оптимизация запросов

```go
// Кешируйте маски компонентов
const (
    RenderableMask = (1 << TransformComponentType) | (1 << SpriteComponentType)
    MovableMask    = (1 << TransformComponentType) | (1 << VelocityComponentType)
)

func (s *RenderSystem) Update(dt float32, em *ecs.EntityManager) {
    entities := em.GetEntitiesWithComponents(RenderableMask)
    // ...
}
```

### Batch обработка

```go
func (s *PhysicsSystem) Update(dt float32, em *ecs.EntityManager) {
    entities := em.GetEntitiesWithComponents(PhysicsMask)

    // Обрабатываем батчами для лучшей производительности кеша
    const batchSize = 64
    for i := 0; i < len(entities); i += batchSize {
        end := i + batchSize
        if end > len(entities) {
            end = len(entities)
        }

        s.processBatch(entities[i:end], dt, em)
    }
}
```

## Отладка

### Вывод информации о мире

```go
func PrintWorldInfo(world *ecs.World) {
    fmt.Printf("Entities: %d\n", world.EntityCount())

    entities := world.GetAllEntities()
    for _, entityID := range entities {
        fmt.Printf("Entity %d:\n", entityID)

        components := world.GetEntityManager().GetComponentManager().GetAllComponents(entityID)
        for _, comp := range components {
            fmt.Printf("  - %T\n", comp)
        }
    }
}
```

## Заключение

ECS в AnimoEngine предоставляет мощный и гибкий способ организации игровой логики. Следуя этим рекомендациям, вы сможете создавать эффективные и масштабируемые игры.

### Дополнительные ресурсы

- [Архитектура движка](architecture.md)
- [Быстрый старт](quickstart.md)
- [Примеры кода](../examples/)
