# 🚀 Быстрый старт - Высоконагруженный микросервис

## Что это?

Высоконагруженный микросервис на Go с использованием современных технологий:
- **Go 1.21** + **Gin** для HTTP API
- **PostgreSQL** для хранения данных
- **Redis** для кэширования
- **Kafka** для асинхронной обработки событий
- **Docker** + **Kubernetes** для развертывания
- **Горутины и каналы** для параллельной обработки

## ⚡ Запуск за 5 минут

### 1. Клонирование и настройка
```bash
git clone <repository-url>
cd highload-microservice
cp env.example .env
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
```

## ☸️ Kubernetes развертывание

```bash
# Развертывание
./scripts/k8s-deploy.sh

# Или вручную
kubectl apply -f k8s/

# Проверка
kubectl get pods -n highload-microservice

# Доступ
kubectl port-forward service/highload-service 8080:80 -n highload-microservice
```

## 🧪 Тестирование

```bash
# Автоматическое тестирование
./scripts/test-api.sh

# Нагрузочное тестирование
hey -n 1000 -c 10 http://localhost:8080/api/v1/users

# Создание пользователей
hey -n 100 -c 5 -m POST -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","first_name":"Test","last_name":"User"}' \
  http://localhost:8080/api/v1/users
```

## 📚 API Endpoints

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/health` | Health check |
| POST | `/api/v1/users` | Создание пользователя |
| GET | `/api/v1/users/{id}` | Получение пользователя |
| PUT | `/api/v1/users/{id}` | Обновление пользователя |
| DELETE | `/api/v1/users/{id}` | Удаление пользователя |
| GET | `/api/v1/users` | Список пользователей |
| POST | `/api/v1/events` | Создание события |
| GET | `/api/v1/events/{id}` | Получение события |
| GET | `/api/v1/events` | Список событий |

## 🔧 Полезные команды

```bash
# Сборка приложения
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

```bash
# Логи приложения
docker-compose logs -f app

# Логи Kubernetes
kubectl logs -f deployment/highload-microservice -n highload-microservice

# Метрики ресурсов
kubectl top pods -n highload-microservice
```

## 🛠 Разработка

### Структура проекта
```
├── cmd/                    # Точки входа
├── internal/               # Внутренние пакеты
│   ├── config/            # Конфигурация
│   ├── database/          # PostgreSQL
│   ├── handlers/          # HTTP обработчики
│   ├── kafka/             # Kafka клиенты
│   ├── models/            # Модели данных
│   ├── redis/             # Redis клиент
│   ├── services/          # Бизнес-логика
│   └── worker/            # Worker pool
├── k8s/                   # Kubernetes манифесты
├── scripts/               # Скрипты запуска
├── examples/              # Примеры API
└── docs/                  # Документация
```

### Основные компоненты

1. **HTTP API** - RESTful интерфейс
2. **User Service** - управление пользователями
3. **Event Service** - обработка событий
4. **Worker Pool** - параллельная обработка
5. **Database Layer** - PostgreSQL
6. **Cache Layer** - Redis
7. **Message Broker** - Kafka

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

## 📈 Производительность

### Оптимизации
- **Connection pooling** для PostgreSQL
- **Кэширование** в Redis
- **Параллельная обработка** с worker pool
- **Batch операции** для Kafka
- **Индексы** в базе данных

### Масштабирование
- **Горизонтальное** в Kubernetes
- **Автомасштабирование** (HPA)
- **Load balancing**
- **Graceful shutdown**

## 🔒 Безопасность

- Валидация входных данных
- SQL injection protection
- Non-root пользователь в Docker
- Secrets management в Kubernetes
- Rate limiting через Ingress

## 📞 Поддержка

- **Документация**: `README.md`
- **Примеры API**: `examples/api-examples.md`
- **Развертывание**: `DEPLOYMENT.md`
- **Исходный код**: `internal/`

## 🎯 Следующие шаги

1. **Изучите код** в `internal/` директории
2. **Протестируйте API** с помощью примеров
3. **Настройте мониторинг** для продакшена
4. **Добавьте аутентификацию** при необходимости
5. **Настройте CI/CD** для автоматического развертывания

---

**Готово!** 🎉 Ваш высоконагруженный микросервис запущен и готов к работе!


