# Участие в разработке AnimoEngine

Спасибо за интерес к AnimoEngine! Мы рады любому вкладу в развитие проекта.

## С чего начать

1. Изучите [README.md](README.md) и [документацию](docs/)
2. Запустите [демо-приложение](cmd/demo/main.go)
3. Ознакомьтесь с [архитектурой](docs/architecture.md)
4. Посмотрите [открытые issues](https://github.com/Salamander5876/AnimoEngine/issues)

## Как внести вклад

### Сообщения об ошибках

Если вы нашли ошибку:

1. Проверьте, не была ли она уже зарегистрирована в [Issues](https://github.com/Salamander5876/AnimoEngine/issues)
2. Создайте новый issue с описанием:
   - Версия Go
   - Операционная система
   - Шаги для воспроизведения
   - Ожидаемое и фактическое поведение
   - Скриншоты (если применимо)

### Предложения новых функций

1. Создайте issue с тегом `enhancement`
2. Опишите желаемую функциональность
3. Объясните, зачем она нужна
4. Предложите возможную реализацию (опционально)

### Pull Requests

#### Процесс

1. Форкните репозиторий
2. Создайте новую ветку:
   ```bash
   git checkout -b feature/my-feature
   ```
3. Внесите изменения
4. Напишите тесты
5. Убедитесь что все тесты проходят:
   ```bash
   go test ./...
   ```
6. Проверьте форматирование:
   ```bash
   go fmt ./...
   go vet ./...
   ```
7. Закоммитьте изменения:
   ```bash
   git commit -m "Add feature: my feature"
   ```
8. Запушьте в свой форк:
   ```bash
   git push origin feature/my-feature
   ```
9. Создайте Pull Request

#### Требования к PR

- Код должен соответствовать [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Все публичные функции должны быть документированы
- Добавьте тесты для новой функциональности
- Обновите документацию при необходимости
- PR должен решать одну задачу
- Коммиты должны быть атомарными и иметь понятные сообщения

## Стандарты кодирования

### Go стиль

Следуйте официальным рекомендациям Go:

```go
// Хорошо: краткие, понятные имена
func NewEngine() *Engine

// Плохо: verbose, не идиоматично
func CreateNewEngineInstance() *Engine
```

### Комментарии

Все публичные функции и типы должны быть документированы:

```go
// Engine представляет главный класс игрового движка.
// Он объединяет все подсистемы и управляет игровым циклом.
type Engine struct {
    // ...
}

// NewEngine создает новый экземпляр движка с настройками по умолчанию.
// Возвращает готовый к использованию Engine.
func NewEngine() *Engine {
    // ...
}
```

### Обработка ошибок

Всегда обрабатывайте ошибки:

```go
// Хорошо
if err := window.Initialize(); err != nil {
    return fmt.Errorf("failed to initialize window: %w", err)
}

// Плохо
window.Initialize() // Игнорирование ошибки
```

### Тестирование

Пишите unit-тесты для всех функций:

```go
func TestEntityManager_CreateEntity(t *testing.T) {
    em := NewEntityManager()

    entity1 := em.CreateEntity()
    entity2 := em.CreateEntity()

    if entity1 == entity2 {
        t.Error("CreateEntity returned duplicate IDs")
    }

    if !em.Exists(entity1) {
        t.Error("Created entity does not exist")
    }
}
```

### Бенчмарки

Для критичных по производительности функций добавляйте бенчмарки:

```go
func BenchmarkEntityManager_CreateEntity(b *testing.B) {
    em := NewEntityManager()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        em.CreateEntity()
    }
}
```

## Структура проекта

При добавлении новых компонентов следуйте существующей структуре:

```
pkg/
  core/           # Ядро движка
  graphics/       # Графическая подсистема
  platform/       # Платформенная абстракция
  game/           # Игровые системы
    rpg/         # RPG компоненты

cmd/             # Исполняемые файлы
  demo/         # Демо-приложения

examples/        # Примеры использования

docs/           # Документация
```

## Приоритеты разработки

### Высокий приоритет

- Исправление критических багов
- Улучшение производительности
- Документация
- Базовые функции движка

### Средний приоритет

- Новые функции
- Оптимизации
- Улучшение API
- Примеры кода

### Низкий приоритет

- Рефакторинг (без изменения функциональности)
- Косметические изменения
- Экспериментальные функции

## Roadmap

### Фаза 2 (В разработке)
- [ ] Sprite batching
- [ ] Camera система с frustum culling
- [ ] Texture atlas поддержка
- [ ] 2D физика (коллизии)

### Фаза 3 (Планируется)
- [ ] WebView интеграция
- [ ] Go-JS биндинги
- [ ] UI система на HTML/CSS

### Фаза 4 (Будущее)
- [ ] 3D model loading (OBJ, GLTF)
- [ ] Advanced lighting
- [ ] Shadow mapping
- [ ] Particle systems

### Фаза 5 (Долгосрочно)
- [ ] Полная RPG система
- [ ] Редактор уровней
- [ ] Networking (опционально)
- [ ] Web поддержка (WASM)

## Вопросы и поддержка

- **GitHub Issues**: [создайте issue](https://github.com/Salamander5876/AnimoEngine/issues)
- **Документация**: [docs/](docs/)
- **Примеры**: [examples/](examples/)

## Лицензия

Внося вклад в AnimoEngine, вы соглашаетесь что ваш код будет лицензирован под MIT License.

## Благодарности

Мы ценим каждый вклад в проект! Все контрибьюторы будут упомянуты в README.

---

**Спасибо за участие в разработке AnimoEngine!** 🎮
