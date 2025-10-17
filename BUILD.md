# Инструкции по сборке AnimoEngine

## Требования

### Общие требования
- Go 1.21 или выше
- Git
- OpenGL 3.3+ драйверы

### Windows
- [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) или MinGW-w64
- Windows 7 или выше

### Linux
- GCC
- Mesa OpenGL development files
- X11 development files

### macOS
- Xcode Command Line Tools
- macOS 10.12 или выше

## Установка зависимостей

### Windows

1. Установите Go с [официального сайта](https://golang.org/dl/)

2. Установите TDM-GCC:
   - Скачайте с https://jmeubank.github.io/tdm-gcc/
   - Запустите установщик
   - Выберите "MinGW-w64/TDM64"
   - Убедитесь что добавлен в PATH

3. Проверьте установку:
   ```cmd
   go version
   gcc --version
   ```

### Linux (Ubuntu/Debian)

```bash
# Установка Go
sudo apt update
sudo apt install golang-go

# Установка зависимостей для OpenGL и GLFW
sudo apt install libgl1-mesa-dev xorg-dev

# Проверка
go version
```

### Linux (Fedora)

```bash
# Установка Go
sudo dnf install golang

# Установка зависимостей
sudo dnf install mesa-libGL-devel libXrandr-devel libXcursor-devel libXinerama-devel libXi-devel

# Проверка
go version
```

### macOS

```bash
# Установка Go (через Homebrew)
brew install go

# Установка Xcode Command Line Tools
xcode-select --install

# Проверка
go version
```

## Клонирование репозитория

```bash
git clone https://github.com/Salamander5876/AnimoEngine.git
cd AnimoEngine
```

## Установка Go зависимостей

```bash
go mod download
```

Это скачает все необходимые Go пакеты:
- github.com/go-gl/gl
- github.com/go-gl/glfw
- github.com/go-gl/mathgl
- github.com/webview/webview_go (для будущей UI интеграции)

## Сборка

### Сборка демо-приложения

```bash
# Сборка
go build -o bin/demo cmd/demo/main.go

# Запуск
./bin/demo                  # Linux/macOS
bin\demo.exe               # Windows
```

### Сборка с оптимизациями

```bash
# Релизная сборка с оптимизациями
go build -ldflags="-s -w" -o bin/demo cmd/demo/main.go
```

Флаги:
- `-s` - отключить таблицу символов
- `-w` - отключить DWARF отладочную информацию
- Результат: меньший размер исполняемого файла

### Кросс-компиляция

```bash
# Windows из Linux/macOS
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o bin/demo.exe cmd/demo/main.go

# Linux из Windows (требует WSL)
GOOS=linux GOARCH=amd64 go build -o bin/demo cmd/demo/main.go
```

## Запуск тестов

```bash
# Все тесты
go test ./...

# С подробным выводом
go test -v ./...

# Конкретный пакет
go test -v ./pkg/core/ecs/

# С покрытием
go test -cover ./...

# Генерация отчета покрытия
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Бенчмарки

```bash
# Запуск всех бенчмарков
go test -bench=. ./...

# Конкретный пакет
go test -bench=. ./pkg/core/ecs/

# С профилированием памяти
go test -bench=. -benchmem ./...

# CPU профиль
go test -bench=. -cpuprofile=cpu.prof ./pkg/core/ecs/
go tool pprof cpu.prof
```

## Профилирование

### CPU профилирование

```bash
# Создание профиля
go test -cpuprofile=cpu.prof -bench=. ./pkg/core/ecs/

# Анализ
go tool pprof cpu.prof

# Команды в pprof:
# top - топ функций по времени
# list <function> - просмотр кода функции
# web - визуализация (требует graphviz)
```

### Память профилирование

```bash
# Создание профиля
go test -memprofile=mem.prof -bench=. ./pkg/core/ecs/

# Анализ
go tool pprof mem.prof
```

### Runtime профилирование

Добавьте в код:

```go
import _ "net/http/pprof"
import "net/http"

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    // Ваш код
}
```

Затем:

```bash
# CPU профиль
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Heap профиль
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutines
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

## Структура сборки

```
AnimoEngine/
├── bin/              # Скомпилированные исполняемые файлы
├── cmd/              # Точки входа приложений
│   └── demo/        # Демо-приложение
├── pkg/             # Библиотечный код
└── assets/          # Игровые ресурсы (создайте при необходимости)
```

## Частые проблемы

### Windows: "gcc: command not found"

**Решение**: Установите TDM-GCC и добавьте в PATH:
```cmd
set PATH=%PATH%;C:\TDM-GCC-64\bin
```

### Linux: "package gl is not in GOROOT"

**Решение**: Установите dev пакеты:
```bash
sudo apt-get install libgl1-mesa-dev xorg-dev
```

### macOS: "ld: library not found"

**Решение**: Установите Xcode Command Line Tools:
```bash
xcode-select --install
```

### "undefined reference to `glfwCreateWindow`"

**Решение**: Убедитесь что CGO включен:
```bash
export CGO_ENABLED=1
```

### Windows: Долгая сборка

**Причина**: Антивирус сканирует компилируемые файлы.

**Решение**: Добавьте папку проекта в исключения антивируса.

## Оптимизация сборки

### Кеширование модулей

Go автоматически кеширует модули в:
- Linux/macOS: `~/go/pkg/mod`
- Windows: `%USERPROFILE%\go\pkg\mod`

Очистка кеша:
```bash
go clean -modcache
```

### Параллельная сборка

```bash
# Установка количества параллельных задач
go build -p 4 cmd/demo/main.go
```

### Использование сборочного кеша

```bash
# Проверка кеша
go env GOCACHE

# Очистка кеша
go clean -cache
```

## Continuous Integration

### GitHub Actions

Пример `.github/workflows/build.yml`:

```yaml
name: Build

on: [push, pull_request]

jobs:
  build:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        go: ['1.21']

    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go }}

    - name: Install dependencies (Linux)
      if: matrix.os == 'ubuntu-latest'
      run: |
        sudo apt-get update
        sudo apt-get install -y libgl1-mesa-dev xorg-dev

    - name: Build
      run: go build -v ./...

    - name: Test
      run: go test -v ./...
```

## Развертывание

### Создание релиза

```bash
# Сборка для всех платформ
./scripts/build-all.sh  # Создайте этот скрипт

# Структура:
# releases/
# ├── AnimoEngine-v0.1.0-windows-amd64.zip
# ├── AnimoEngine-v0.1.0-linux-amd64.tar.gz
# └── AnimoEngine-v0.1.0-darwin-amd64.tar.gz
```

### Docker сборка (опционально)

```dockerfile
FROM golang:1.21 as builder

WORKDIR /app
COPY . .

RUN apt-get update && apt-get install -y libgl1-mesa-dev xorg-dev
RUN go mod download
RUN CGO_ENABLED=1 go build -o /demo cmd/demo/main.go

FROM ubuntu:22.04
RUN apt-get update && apt-get install -y libgl1
COPY --from=builder /demo /demo
CMD ["/demo"]
```

## Дополнительные команды

```bash
# Форматирование кода
go fmt ./...

# Проверка кода
go vet ./...

# Статический анализ (требует golangci-lint)
golangci-lint run

# Обновление зависимостей
go get -u ./...
go mod tidy

# Проверка модулей
go mod verify

# Документация
godoc -http=:6060
# Откройте http://localhost:6060
```

## Следующие шаги

После успешной сборки:

1. Изучите [README.md](README.md)
2. Запустите демо: `./bin/demo`
3. Прочитайте [Быстрый старт](docs/quickstart.md)
4. Посмотрите [примеры](examples/)

---

**Успешной разработки!** 🚀
