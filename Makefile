# AnimoEngine Makefile
# Простые команды для работы с проектом

.PHONY: help build run test clean fmt vet demo

# По умолчанию показываем помощь
help:
	@echo "AnimoEngine - Makefile команды:"
	@echo ""
	@echo "  make build    - Сборка проекта"
	@echo "  make demo     - Запуск демо-приложения"
	@echo "  make run      - То же что и demo"
	@echo "  make test     - Запуск тестов"
	@echo "  make fmt      - Форматирование кода"
	@echo "  make vet      - Проверка кода"
	@echo "  make clean    - Очистка сборки"
	@echo "  make deps     - Установка зависимостей"
	@echo "  make all      - fmt + vet + build + test"

# Установка зависимостей
deps:
	@echo "Установка зависимостей..."
	go mod download
	go mod verify

# Сборка всех пакетов
build: fmt
	@echo "Сборка проекта..."
	go build ./...

# Сборка демо-приложения
build-demo: fmt
	@echo "Сборка демо-приложения..."
	mkdir -p bin
	go build -o bin/demo cmd/demo/main.go

# Запуск демо
demo: build-demo
	@echo "Запуск демо..."
	./bin/demo

run: demo

# Запуск тестов
test:
	@echo "Запуск тестов..."
	go test ./...

# Тесты с покрытием
test-coverage:
	@echo "Тесты с покрытием..."
	go test -cover ./...

# Детальное покрытие
coverage:
	@echo "Генерация отчета покрытия..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Бенчмарки
bench:
	@echo "Запуск бенчмарков..."
	go test -bench=. -benchmem ./...

# Форматирование кода
fmt:
	@echo "Форматирование кода..."
	go fmt ./...

# Проверка кода
vet:
	@echo "Проверка кода..."
	go vet ./...

# Очистка
clean:
	@echo "Очистка..."
	rm -rf bin/
	rm -f coverage.out
	rm -f *.prof
	go clean -cache
	go clean -testcache

# Обновление зависимостей
update:
	@echo "Обновление зависимостей..."
	go get -u ./...
	go mod tidy

# Полная проверка
all: fmt vet build test
	@echo "Все проверки пройдены!"

# Релизная сборка
release:
	@echo "Релизная сборка..."
	mkdir -p bin/release
	CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/release/demo cmd/demo/main.go
	@echo "Релиз создан в bin/release/"

# Докер сборка (если понадобится)
docker-build:
	@echo "Docker сборка..."
	docker build -t animoengine:latest .

# Информация о проекте
info:
	@echo "AnimoEngine - Информация о проекте"
	@echo ""
	@echo "Go версия:"
	@go version
	@echo ""
	@echo "Статистика кода:"
	@find . -name "*.go" | xargs wc -l | tail -1
	@echo ""
	@echo "Количество файлов:"
	@find . -name "*.go" | wc -l
