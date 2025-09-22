APP_NAME=highload-microservice
GO_FILES=$(shell go list ./...)

.PHONY: all build run test cover lint e2e-compose e2e-k8s docker build-docker push-docker fmt vet tidy deps clean

all: build

deps:
	go mod download

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

build:
	CGO_ENABLED=0 go build -o bin/$(APP_NAME) .

run:
	SERVER_PORT=8080 go run main.go

test:
	go test ./...

cover:
	go test -race -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -n 1

lint:
	golangci-lint run --timeout=5m

docker:
	docker build -t $(APP_NAME):latest .

build-docker:
	docker build -t $(APP_NAME):latest .

push-docker:
	docker push $(APP_NAME):latest

e2e-compose:
	docker compose up -d
	sleep 5
	curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8080/health

e2e-k8s:
	@echo "Запустите GitHub Actions workflow 'e2e-k8s' или используйте Kind локально"

clean:
	rm -rf bin coverage.out
# Makefile для высоконагруженного микросервиса

.PHONY: help build run test clean docker-build docker-run docker-stop k8s-deploy k8s-clean

# Переменные
APP_NAME=highload-microservice
DOCKER_IMAGE=$(APP_NAME):latest
K8S_NAMESPACE=highload-microservice

# Цвета для вывода
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Показать справку
	@echo "$(GREEN)Высоконагруженный микросервис на Go$(NC)"
	@echo ""
	@echo "$(YELLOW)Доступные команды:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Разработка
build: ## Собрать приложение
	@echo "$(GREEN)Сборка приложения...$(NC)"
	go build -o bin/$(APP_NAME) .

run: ## Запустить приложение локально
	@echo "$(GREEN)Запуск приложения...$(NC)"
	go run main.go

test: ## Запустить тесты
	@echo "$(GREEN)Запуск тестов...$(NC)"
	go test -v ./...

test-coverage: ## Запустить тесты с покрытием
	@echo "$(GREEN)Запуск тестов с покрытием...$(NC)"
	go test -v -cover ./...

benchmark: ## Запустить бенчмарки
	@echo "$(GREEN)Запуск бенчмарков...$(NC)"
	go test -bench=. ./...

lint: ## Запустить линтер
	@echo "$(GREEN)Запуск линтера...$(NC)"
	golangci-lint run

fmt: ## Форматировать код
	@echo "$(GREEN)Форматирование кода...$(NC)"
	go fmt ./...

mod-tidy: ## Очистить зависимости
	@echo "$(GREEN)Очистка зависимостей...$(NC)"
	go mod tidy

# Docker
docker-build: ## Собрать Docker образ
	@echo "$(GREEN)Сборка Docker образа...$(NC)"
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Запустить с Docker Compose
	@echo "$(GREEN)Запуск с Docker Compose...$(NC)"
	docker-compose up --build -d

docker-stop: ## Остановить Docker Compose
	@echo "$(GREEN)Остановка Docker Compose...$(NC)"
	docker-compose down

docker-logs: ## Показать логи Docker
	@echo "$(GREEN)Логи Docker...$(NC)"
	docker-compose logs -f

docker-clean: ## Очистить Docker ресурсы
	@echo "$(GREEN)Очистка Docker ресурсов...$(NC)"
	docker-compose down -v
	docker system prune -f

# Kubernetes
k8s-deploy: ## Развернуть в Kubernetes
	@echo "$(GREEN)Развертывание в Kubernetes...$(NC)"
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/secret.yaml
	kubectl apply -f k8s/configmap.yaml
	kubectl apply -f k8s/postgres-deployment.yaml
	kubectl apply -f k8s/redis-deployment.yaml
	kubectl apply -f k8s/kafka-deployment.yaml
	kubectl apply -f k8s/app-deployment.yaml
	kubectl apply -f k8s/ingress.yaml
	@echo "$(GREEN)Развертывание завершено!$(NC)"

k8s-status: ## Показать статус Kubernetes
	@echo "$(GREEN)Статус Kubernetes...$(NC)"
	kubectl get pods -n $(K8S_NAMESPACE)
	kubectl get services -n $(K8S_NAMESPACE)
	kubectl get ingress -n $(K8S_NAMESPACE)

k8s-logs: ## Показать логи Kubernetes
	@echo "$(GREEN)Логи Kubernetes...$(NC)"
	kubectl logs -f deployment/$(APP_NAME) -n $(K8S_NAMESPACE)

k8s-scale: ## Масштабировать приложение
	@echo "$(GREEN)Масштабирование приложения...$(NC)"
	kubectl scale deployment $(APP_NAME) --replicas=5 -n $(K8S_NAMESPACE)

k8s-clean: ## Удалить из Kubernetes
	@echo "$(GREEN)Удаление из Kubernetes...$(NC)"
	kubectl delete namespace $(K8S_NAMESPACE)

# Утилиты
clean: ## Очистить артефакты сборки
	@echo "$(GREEN)Очистка артефактов...$(NC)"
	rm -rf bin/
	go clean

install-tools: ## Установить инструменты разработки
	@echo "$(GREEN)Установка инструментов разработки...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/rakyll/hey@latest

# Тестирование производительности
load-test: ## Нагрузочное тестирование
	@echo "$(GREEN)Нагрузочное тестирование...$(NC)"
	hey -n 1000 -c 10 http://localhost:8080/api/v1/users

load-test-create: ## Нагрузочное тестирование создания пользователей
	@echo "$(GREEN)Нагрузочное тестирование создания пользователей...$(NC)"
	hey -n 1000 -c 10 -m POST -H "Content-Type: application/json" -d '{"email":"test@example.com","first_name":"Test","last_name":"User"}' http://localhost:8080/api/v1/users

# Мониторинг
monitor: ## Показать метрики
	@echo "$(GREEN)Метрики приложения...$(NC)"
	curl -s http://localhost:8080/health | jq .

# Полная очистка
clean-all: clean docker-clean ## Полная очистка всех ресурсов
	@echo "$(GREEN)Полная очистка завершена!$(NC)"

# По умолчанию
.DEFAULT_GOAL := help


