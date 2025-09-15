# 🚀 Быстрый старт - Высоконагруженный микросервис

## Что вы получите

Высоконагруженный микросервис на Go с полным стеком современных технологий:

- ✅ **Go 1.21** + **Gin** для HTTP API
- ✅ **PostgreSQL** для хранения данных  
- ✅ **Redis** для кэширования
- ✅ **Kafka** для асинхронной обработки
- ✅ **Docker** + **Kubernetes** для развертывания
- ✅ **Горутины и каналы** для параллельной обработки
- ✅ **Worker Pool** для управления горутинами
- ✅ **Health checks** и мониторинг
- ✅ **Graceful shutdown** и обработка ошибок

## ⚡ Запуск за 3 минуты

### 1. Клонирование
```bash
git clone <repository-url>
cd highload-microservice
```

### 2. Запуск с Docker Compose
```bash
# Linux/Mac
./scripts/start.sh

# Windows
scripts\start.bat

# Или вручную
docker-compose up -d
```

### 3. Проверка работы
```bash
# Health check
curl http://localhost:8080/health

# Создание пользователя
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","first_name":"Test","last_name":"User"}'

# Список пользователей
curl http://localhost:8080/api/v1/users
```

## 🎯 Что происходит под капотом

### Архитектура
```
HTTP Client → Gin Router → Services → Database/Cache
                    ↓
                Kafka Producer → Kafka → Consumer → Worker Pool
```

### Компоненты
1. **HTTP API** - RESTful интерфейс
2. **User Service** - управление пользователями с кэшированием
3. **Event Service** - обработка событий через Kafka
4. **Worker Pool** - параллельная обработка с горутинами
5. **Database Layer** - PostgreSQL с connection pooling
6. **Cache Layer** - Redis для быстрого доступа
7. **Message Broker** - Kafka для асинхронности

## 📚 API Endpoints

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/health` | Проверка состояния |
| POST | `/api/v1/users` | Создание пользователя |
| GET | `/api/v1/users/{id}` | Получение пользователя |
| PUT | `/api/v1/users/{id}` | Обновление пользователя |
| DELETE | `/api/v1/users/{id}` | Удаление пользователя |
| GET | `/api/v1/users` | Список пользователей |
| POST | `/api/v1/events` | Создание события |
| GET | `/api/v1/events/{id}` | Получение события |
| GET | `/api/v1/events` | Список событий |

## 🧪 Тестирование

### Автоматическое тестирование
```bash
# Полный тест API
./scripts/test-api.sh

# Нагрузочное тестирование
hey -n 1000 -c 10 http://localhost:8080/api/v1/users
```

### Ручное тестирование
```bash
# 1. Создание пользователя
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","first_name":"John","last_name":"Doe"}'

# 2. Получение пользователя (замените ID)
curl http://localhost:8080/api/v1/users/123e4567-e89b-12d3-a456-426614174000

# 3. Создание события
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{"user_id":"123e4567-e89b-12d3-a456-426614174000","type":"login","data":"{\"ip\":\"192.168.1.1\"}"}'

# 4. Список пользователей
curl http://localhost:8080/api/v1/users?page=1&limit=10
```

## 🐳 Docker команды

```bash
# Запуск
docker-compose up -d

# Просмотр логов
docker-compose logs -f

# Остановка
docker-compose down

# Перезапуск
docker-compose restart

# Статус
docker-compose ps

# Очистка
docker-compose down -v
```

## ☸️ Kubernetes развертывание

### Локальный кластер (Minikube)
```bash
# Запуск Minikube
minikube start

# Развертывание
./scripts/k8s-deploy.sh

# Проверка
kubectl get pods -n highload-microservice

# Доступ
kubectl port-forward service/highload-service 8080:80 -n highload-microservice
```

### Облачный кластер
```bash
# Развертывание
kubectl apply -f k8s/

# Проверка
kubectl get all -n highload-microservice

# Получение внешнего IP
kubectl get service highload-service -n highload-microservice
```

## 🔧 Разработка

### Структура проекта
```
highload-microservice/
├── internal/
│   ├── config/          # Конфигурация
│   ├── database/        # PostgreSQL
│   ├── handlers/        # HTTP обработчики
│   ├── kafka/           # Kafka клиенты
│   ├── models/          # Модели данных
│   ├── redis/           # Redis клиент
│   ├── services/        # Бизнес-логика
│   └── worker/          # Worker pool
├── k8s/                 # Kubernetes манифесты
├── scripts/             # Скрипты запуска
└── examples/            # Примеры API
```

### Команды разработки
```bash
# Сборка
go build -o bin/highload-microservice .

# Запуск локально
go run main.go

# Тесты
go test ./...

# Линтер
golangci-lint run

# Форматирование
go fmt ./...
```

## 📊 Мониторинг

### Логи
```bash
# Все сервисы
docker-compose logs -f

# Только приложение
docker-compose logs -f app

# Kubernetes
kubectl logs -f deployment/highload-microservice -n highload-microservice
```

### Метрики
```bash
# Health check
curl http://localhost:8080/health

# Статус контейнеров
docker-compose ps

# Статус Kubernetes
kubectl get pods -n highload-microservice
```

## 🚨 Устранение неполадок

### Проблема: Сервис не запускается
```bash
# Проверка логов
docker-compose logs app

# Проверка статуса
docker-compose ps

# Перезапуск
docker-compose restart app
```

### Проблема: База данных недоступна
```bash
# Проверка PostgreSQL
docker-compose logs postgres

# Подключение к БД
docker-compose exec postgres psql -U postgres -d highload_db
```

### Проблема: Redis недоступен
```bash
# Проверка Redis
docker-compose logs redis

# Подключение к Redis
docker-compose exec redis redis-cli ping
```

### Проблема: Kafka недоступен
```bash
# Проверка Kafka
docker-compose logs kafka

# Список топиков
docker-compose exec kafka kafka-topics --bootstrap-server localhost:9092 --list
```

## 🎓 Что изучить дальше

### Основы
1. **Go** - язык программирования
2. **Gin** - веб-фреймворк
3. **PostgreSQL** - база данных
4. **Redis** - кэш
5. **Kafka** - брокер сообщений

### Продвинутые темы
1. **Микросервисы** - архитектурный подход
2. **Docker** - контейнеризация
3. **Kubernetes** - оркестрация
4. **Мониторинг** - Prometheus, Grafana
5. **Безопасность** - аутентификация, авторизация

### Практические навыки
1. **API дизайн** - RESTful интерфейсы
2. **Тестирование** - unit и integration тесты
3. **DevOps** - CI/CD и автоматизация
4. **Производительность** - профилирование и оптимизация

## 📚 Дополнительные ресурсы

- **Документация**: `README.md`
- **Примеры API**: `examples/api-examples.md`
- **Развертывание**: `DEPLOYMENT.md`
- **Установка**: `INSTALL.md`
- **Обзор проекта**: `PROJECT_OVERVIEW.md`

## 🤝 Поддержка

Если у вас возникли вопросы:

1. Проверьте раздел "Устранение неполадок"
2. Изучите логи: `docker-compose logs`
3. Проверьте статус: `docker-compose ps`
4. Создайте issue в репозитории

---

**Готово!** 🎉 Теперь у вас есть полнофункциональный высоконагруженный микросервис!

**Следующий шаг**: Изучите код в `internal/` директории и начните экспериментировать с API!


