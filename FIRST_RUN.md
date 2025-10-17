# Первый запуск AnimoEngine

## Быстрая инструкция

После клонирования или создания проекта выполните следующие команды:

### Шаг 1: Проверка установки Go

```bash
go version
```

Должно показать версию Go 1.21 или выше. Если нет - установите Go с https://golang.org/dl/

### Шаг 2: Инициализация зависимостей

```bash
# В корне проекта AnimoEngine
go mod download
```

Эта команда:
- Скачает все зависимости из go.mod
- Создаст файл go.sum с контрольными суммами
- Кеширует модули в `~/go/pkg/mod/`

### Шаг 3: Проверка сборки

```bash
go build ./...
```

Должно пройти без ошибок. Если возникают ошибки, смотрите раздел "Решение проблем" ниже.

### Шаг 4: Запуск демо

```bash
go run cmd/demo/main.go
```

Должно открыться окно с вращающимся треугольником!

## Альтернативный способ (через Makefile)

Если у вас установлен Make:

```bash
make deps      # Установка зависимостей
make build     # Сборка проекта
make demo      # Запуск демо
```

## Решение проблем

### Windows: "gcc: command not found"

**Проблема**: Нет компилятора C для CGO.

**Решение**:
1. Скачайте TDM-GCC: https://jmeubank.github.io/tdm-gcc/
2. Установите (выберите MinGW-w64/TDM64)
3. Добавьте в PATH: `C:\TDM-GCC-64\bin`
4. Перезапустите терминал
5. Проверьте: `gcc --version`

### Linux: "package gl is not in GOROOT"

**Проблема**: Отсутствуют dev библиотеки OpenGL.

**Решение** (Ubuntu/Debian):
```bash
sudo apt update
sudo apt install libgl1-mesa-dev xorg-dev
```

**Решение** (Fedora):
```bash
sudo dnf install mesa-libGL-devel libXrandr-devel libXcursor-devel libXinerama-devel libXi-devel
```

### macOS: "ld: library not found"

**Проблема**: Отсутствуют Xcode Command Line Tools.

**Решение**:
```bash
xcode-select --install
```

### "malformed go.sum"

**Проблема**: Некорректный файл go.sum.

**Решение**:
```bash
# Удалите go.sum
rm go.sum

# Пересоздайте
go mod download
go mod verify
```

### Долгая первая сборка на Windows

**Причина**: Антивирус сканирует каждый скомпилированный файл.

**Решение**: Добавьте папку проекта в исключения Windows Defender.

## Зависимости, которые будут скачаны

При выполнении `go mod download` будут установлены:

1. **github.com/go-gl/gl** (v0.0.0-20231021071112-07e5d0ea2e71)
   - OpenGL биндинги для Go
   - Размер: ~15 MB

2. **github.com/go-gl/glfw** (v0.0.0-20240506104042-037f3cc74f2a)
   - Оконная система (обертка GLFW)
   - Размер: ~2 MB

3. **github.com/go-gl/mathgl** (v1.1.0)
   - Математическая библиотека для 3D графики
   - Размер: ~1 MB

4. **github.com/webview/webview_go** (v0.0.0-20240831120633-6173450d4dd6)
   - WebView для UI (будет использоваться в Фазе 3)
   - Размер: ~5 MB

5. **golang.org/x/image** (косвенная зависимость)
   - Работа с изображениями
   - Размер: ~2 MB

**Общий размер**: ~25 MB

## Проверка успешной установки

После выполнения всех шагов проверьте:

```bash
# Проверка модулей
go list -m all

# Проверка зависимостей
go mod verify

# Должно вывести: "all modules verified"
```

## Что дальше?

После успешной установки:

1. Изучите [README.md](README.md)
2. Прочитайте [Быстрый старт](docs/quickstart.md)
3. Посмотрите [Примеры кода](docs/ecs-guide.md)
4. Начните создавать свою игру!

## Обновление зависимостей в будущем

```bash
# Обновить все зависимости до последних версий
go get -u ./...
go mod tidy

# Или через Makefile
make update
```

## Очистка

Если нужно очистить кеш и пересобрать:

```bash
go clean -cache
go clean -modcache
go mod download
```

## Дополнительная помощь

Если проблемы остались:
- Создайте issue: https://github.com/Salamander5876/AnimoEngine/issues
- Проверьте [BUILD.md](BUILD.md) для детальных инструкций
- Посмотрите секцию "Частые проблемы" в [docs/quickstart.md](docs/quickstart.md)

---

**Удачного запуска!** 🚀
