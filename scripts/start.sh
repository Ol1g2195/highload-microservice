#!/bin/bash

# Скрипт для запуска высоконагруженного микросервиса

set -e

echo "🚀 Запуск высоконагруженного микросервиса..."

# Проверка наличия Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker не найден. Пожалуйста, установите Docker."
    exit 1
fi

# Проверка наличия Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose не найден. Пожалуйста, установите Docker Compose."
    exit 1
fi

# Создание .env файла если не существует
if [ ! -f .env ]; then
    echo "📝 Создание .env файла..."
    cp env.example .env
    echo "✅ .env файл создан. Отредактируйте его при необходимости."
fi

# Остановка существующих контейнеров
echo "🛑 Остановка существующих контейнеров..."
docker-compose down

# Сборка и запуск контейнеров
echo "🔨 Сборка и запуск контейнеров..."
docker-compose up --build -d

# Ожидание готовности сервисов
echo "⏳ Ожидание готовности сервисов..."
sleep 30

# Проверка статуса контейнеров
echo "📊 Статус контейнеров:"
docker-compose ps

# Проверка health check
echo "🏥 Проверка health check..."
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Микросервис запущен и готов к работе!"
    echo "🌐 API доступно по адресу: http://localhost:8080"
    echo "📚 Документация API: http://localhost:8080/health"
else
    echo "❌ Микросервис не отвечает на health check"
    echo "📋 Логи приложения:"
    docker-compose logs app
    exit 1
fi

echo ""
echo "🎉 Готово! Микросервис успешно запущен."
echo ""
echo "📋 Полезные команды:"
echo "  Просмотр логов: docker-compose logs -f"
echo "  Остановка: docker-compose down"
echo "  Перезапуск: docker-compose restart"
echo "  Статус: docker-compose ps"
echo ""
echo "🧪 Тестирование API:"
echo "  Health check: curl http://localhost:8080/health"
echo "  Создание пользователя:"
echo "    curl -X POST http://localhost:8080/api/v1/users \\"
echo "      -H 'Content-Type: application/json' \\"
echo "      -d '{\"email\":\"test@example.com\",\"first_name\":\"Test\",\"last_name\":\"User\"}'"


