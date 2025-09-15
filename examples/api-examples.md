# Примеры использования API

## Базовый URL
- Локально: `http://localhost:8080`
- Kubernetes: `http://<external-ip>`

## 1. Health Check

### Проверка состояния сервиса
```bash
curl http://localhost:8080/health
```

**Ответ:**
```json
{
  "status": "healthy",
  "timestamp": 1703123456
}
```

## 2. Управление пользователями

### Создание пользователя
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

**Ответ:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "john.doe@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "created_at": "2023-12-21T10:30:00Z",
  "updated_at": "2023-12-21T10:30:00Z"
}
```

### Получение пользователя по ID
```bash
curl http://localhost:8080/api/v1/users/123e4567-e89b-12d3-a456-426614174000
```

### Обновление пользователя
```bash
curl -X PUT http://localhost:8080/api/v1/users/123e4567-e89b-12d3-a456-426614174000 \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Jane",
    "last_name": "Smith"
  }'
```

### Удаление пользователя
```bash
curl -X DELETE http://localhost:8080/api/v1/users/123e4567-e89b-12d3-a456-426614174000
```

### Получение списка пользователей
```bash
# Базовый запрос
curl http://localhost:8080/api/v1/users

# С пагинацией
curl "http://localhost:8080/api/v1/users?page=1&limit=10"

# С фильтрацией (если реализовано)
curl "http://localhost:8080/api/v1/users?search=john"
```

**Ответ:**
```json
{
  "users": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "john.doe@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "created_at": "2023-12-21T10:30:00Z",
      "updated_at": "2023-12-21T10:30:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10
}
```

## 3. Управление событиями

### Создание события
```bash
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "type": "user_login",
    "data": "{\"ip\": \"192.168.1.1\", \"user_agent\": \"Mozilla/5.0\"}"
  }'
```

**Ответ:**
```json
{
  "id": "456e7890-e89b-12d3-a456-426614174001",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "type": "user_login",
  "data": "{\"ip\": \"192.168.1.1\", \"user_agent\": \"Mozilla/5.0\"}",
  "created_at": "2023-12-21T10:35:00Z"
}
```

### Получение события по ID
```bash
curl http://localhost:8080/api/v1/events/456e7890-e89b-12d3-a456-426614174001
```

### Получение списка событий
```bash
# Базовый запрос
curl http://localhost:8080/api/v1/events

# С пагинацией
curl "http://localhost:8080/api/v1/events?page=1&limit=20"

# Фильтрация по пользователю (если реализовано)
curl "http://localhost:8080/api/v1/events?user_id=123e4567-e89b-12d3-a456-426614174000"
```

**Ответ:**
```json
{
  "events": [
    {
      "id": "456e7890-e89b-12d3-a456-426614174001",
      "user_id": "123e4567-e89b-12d3-a456-426614174000",
      "type": "user_login",
      "data": "{\"ip\": \"192.168.1.1\", \"user_agent\": \"Mozilla/5.0\"}",
      "created_at": "2023-12-21T10:35:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20
}
```

## 4. Примеры с использованием различных инструментов

### PowerShell (Windows)
```powershell
# Создание пользователя
$body = @{
    email = "test@example.com"
    first_name = "Test"
    last_name = "User"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/users" -Method POST -Body $body -ContentType "application/json"

# Получение списка пользователей
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/users" -Method GET
```

### Python
```python
import requests
import json

# Создание пользователя
user_data = {
    "email": "python@example.com",
    "first_name": "Python",
    "last_name": "User"
}

response = requests.post(
    "http://localhost:8080/api/v1/users",
    json=user_data,
    headers={"Content-Type": "application/json"}
)

print(response.json())

# Получение списка пользователей
response = requests.get("http://localhost:8080/api/v1/users")
print(response.json())
```

### JavaScript (Node.js)
```javascript
const axios = require('axios');

// Создание пользователя
const userData = {
    email: "nodejs@example.com",
    first_name: "Node",
    last_name: "User"
};

axios.post('http://localhost:8080/api/v1/users', userData)
    .then(response => console.log(response.data))
    .catch(error => console.error(error));

// Получение списка пользователей
axios.get('http://localhost:8080/api/v1/users')
    .then(response => console.log(response.data))
    .catch(error => console.error(error));
```

## 5. Нагрузочное тестирование

### Использование hey
```bash
# Установка hey
go install github.com/rakyll/hey@latest

# Тест получения пользователей
hey -n 1000 -c 10 http://localhost:8080/api/v1/users

# Тест создания пользователей
hey -n 100 -c 5 -m POST -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","first_name":"Test","last_name":"User"}' \
  http://localhost:8080/api/v1/users
```

### Использование Apache Bench (ab)
```bash
# Тест получения пользователей
ab -n 1000 -c 10 http://localhost:8080/api/v1/users

# Тест с POST запросом
ab -n 100 -c 5 -p user.json -T application/json http://localhost:8080/api/v1/users
```

## 6. Мониторинг и отладка

### Проверка логов
```bash
# Docker Compose
docker-compose logs -f app

# Kubernetes
kubectl logs -f deployment/highload-microservice -n highload-microservice
```

### Проверка метрик
```bash
# Health check
curl http://localhost:8080/health

# Проверка подключения к базе данных
curl http://localhost:8080/health | jq '.database'

# Проверка подключения к Redis
curl http://localhost:8080/health | jq '.redis'
```

## 7. Обработка ошибок

### Примеры ошибок и их коды

**400 Bad Request** - Неверный формат запроса:
```json
{
  "error": "Invalid request body",
  "details": "Key: 'CreateUserRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"
}
```

**404 Not Found** - Ресурс не найден:
```json
{
  "error": "User not found"
}
```

**500 Internal Server Error** - Внутренняя ошибка сервера:
```json
{
  "error": "Failed to create user"
}
```

### Обработка ошибок в коде
```python
import requests

try:
    response = requests.post("http://localhost:8080/api/v1/users", json=user_data)
    response.raise_for_status()  # Вызовет исключение для HTTP ошибок
    print(response.json())
except requests.exceptions.HTTPError as e:
    print(f"HTTP ошибка: {e}")
    print(f"Ответ сервера: {e.response.text}")
except requests.exceptions.RequestException as e:
    print(f"Ошибка запроса: {e}")
```


