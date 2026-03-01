# Проверяем существование .env файла и создаём из шаблона, если его нет
ifneq ("$(wildcard ./internal/config/.env)","")
include ./internal/config/.env
export
else
$(info Файл .env не найден. Создаём из шаблона...)
$(shell cp ./internal/config/.env.example ./internal/config/.env)
include ./internal/config/.env
export
$(info Создан файл .env из шаблона .env.example)
endif


.PHONY: all lint build docker-compose docker-down clean 

all: down lint build docker-compose
	@echo "Все шаги выполнены успешно!"

down: docker-compose down -v

lint:
	@echo "Запуск линтера"
	golangci-lint run

build: lint
	@echo "Сборка приложения"
	go build -o main ./cmd/main.go

docker-compose: build
	@echo "Запуск Docker Compose"
	docker-compose --env-file ./internal/config/.env up -d --build

docker-down:
	@echo "Остановка контейнеров"
	docker-compose --env-file ./internal/config/.env down

clean:
	@echo "Очистка собранных файлов"
	rm -f main

