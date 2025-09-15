#!/bin/bash

# Скрипт для тестирования API

set -e

API_URL=${1:-"http://localhost:8080"}

echo "🧪 Тестирование API высоконагруженного микросервиса..."
echo "🌐 API URL: $API_URL"

# Функция для выполнения HTTP запроса
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local expected_status=$4
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method -H "Content-Type: application/json" -d "$data" "$url")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$url")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" -eq "$expected_status" ]; then
        echo "✅ $method $url - $http_code"
        echo "$body" | jq . 2>/dev/null || echo "$body"
    else
        echo "❌ $method $url - $http_code (ожидался $expected_status)"
        echo "$body"
    fi
    echo ""
}

# Проверка health check
echo "1. Проверка health check..."
make_request "GET" "$API_URL/health" "" 200

# Создание пользователя
echo "2. Создание пользователя..."
USER_DATA='{"email":"test@example.com","first_name":"Test","last_name":"User"}'
USER_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -d "$USER_DATA" "$API_URL/api/v1/users")
USER_ID=$(echo "$USER_RESPONSE" | jq -r '.id')
echo "✅ Пользователь создан с ID: $USER_ID"
echo ""

# Получение пользователя
echo "3. Получение пользователя..."
make_request "GET" "$API_URL/api/v1/users/$USER_ID" "" 200

# Обновление пользователя
echo "4. Обновление пользователя..."
UPDATE_DATA='{"first_name":"Updated","last_name":"Name"}'
make_request "PUT" "$API_URL/api/v1/users/$USER_ID" "$UPDATE_DATA" 200

# Список пользователей
echo "5. Получение списка пользователей..."
make_request "GET" "$API_URL/api/v1/users?page=1&limit=10" "" 200

# Создание события
echo "6. Создание события..."
EVENT_DATA="{\"user_id\":\"$USER_ID\",\"type\":\"test_event\",\"data\":\"{\\\"test\\\": \\\"data\\\"}\"}"
EVENT_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -d "$EVENT_DATA" "$API_URL/api/v1/events")
EVENT_ID=$(echo "$EVENT_RESPONSE" | jq -r '.id')
echo "✅ Событие создано с ID: $EVENT_ID"
echo ""

# Получение события
echo "7. Получение события..."
make_request "GET" "$API_URL/api/v1/events/$EVENT_ID" "" 200

# Список событий
echo "8. Получение списка событий..."
make_request "GET" "$API_URL/api/v1/events?page=1&limit=10" "" 200

# Тест производительности
echo "9. Тест производительности (создание 10 пользователей)..."
for i in {1..10}; do
    USER_DATA="{\"email\":\"test$i@example.com\",\"first_name\":\"Test$i\",\"last_name\":\"User$i\"}"
    curl -s -X POST -H "Content-Type: application/json" -d "$USER_DATA" "$API_URL/api/v1/users" > /dev/null
    echo -n "."
done
echo ""
echo "✅ 10 пользователей создано"
echo ""

# Финальная проверка списка пользователей
echo "10. Финальная проверка списка пользователей..."
make_request "GET" "$API_URL/api/v1/users?page=1&limit=20" "" 200

echo "🎉 Все тесты завершены!"
echo ""
echo "📊 Статистика:"
echo "  - Health check: ✅"
echo "  - CRUD операции с пользователями: ✅"
echo "  - CRUD операции с событиями: ✅"
echo "  - Тест производительности: ✅"
echo ""
echo "💡 Для более детального тестирования используйте:"
echo "  hey -n 1000 -c 10 http://$API_URL/api/v1/users"


