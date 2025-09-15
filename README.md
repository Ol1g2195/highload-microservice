# Высоконагруженный микросервис на Go

Этот проект представляет собой высоконагруженный микросервис, разработанный на языке Go с использованием современных технологий и архитектурных решений.

## 🚀 Особенности

- **Высокая производительность**: Использование горутин и каналов для параллельной обработки
- **Микросервисная архитектура**: Разделение на независимые сервисы
- **Современный стек технологий**:
  - **Go 1.21** - основной язык программирования
  - **PostgreSQL** - основная база данных
  - **Redis** - кэширование и быстрый доступ к данным
  - **Kafka** - брокер сообщений для асинхронной обработки
  - **Docker** - контейнеризация
  - **Kubernetes** - оркестрация контейнеров
- **HTTP API** с RESTful интерфейсом
- **Graceful shutdown** и health checks
- **Автомасштабирование** в Kubernetes

## 📋 Требования

### Для локальной разработки:
- Go 1.21+
- Docker и Docker Compose
- Git

### Для развертывания в Kubernetes:
- Kubernetes кластер (minikube, kind, или облачный)
- kubectl
- Docker registry (опционально)

## 🛠 Установка и запуск

### 1. Клонирование репозитория

```bash
git clone <repository-url>
cd highload-microservice
```

### 2. Настройка окружения

Скопируйте файл с переменными окружения:

```bash
cp env.example .env
```

Отредактируйте `.env` файл при необходимости:

```bash
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=highload_db
DB_SSLMODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Kafka Configuration
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=user-events
KAFKA_GROUP_ID=highload-service

# Logging
LOG_LEVEL=info
```

### 3. Запуск с Docker Compose (рекомендуется)

Самый простой способ запустить весь стек:

```bash
# Запуск всех сервисов
docker-compose up -d

# Просмотр логов
docker-compose logs -f

# Остановка
docker-compose down
```

Это запустит:
- PostgreSQL на порту 5432
- Redis на порту 6379
- Kafka на порту 9092
- Микросервис на порту 8080

### 4. Запуск в режиме разработки

Если вы хотите запустить только зависимости:

```bash
# Запуск только зависимостей
docker-compose up -d postgres redis kafka

# Установка зависимостей Go
go mod download

# Запуск приложения
go run main.go
```

## 🐳 Развертывание в Kubernetes

### 1. Подготовка Docker образа

```bash
# Сборка образа
docker build -t highload-microservice:latest .

# Если используете minikube
eval $(minikube docker-env)
docker build -t highload-microservice:latest .
```

### 2. Применение манифестов Kubernetes

```bash
# Создание namespace
kubectl apply -f k8s/namespace.yaml

# Применение всех манифестов
kubectl apply -f k8s/

# Проверка статуса
kubectl get pods -n highload-microservice
kubectl get services -n highload-microservice
```

### 3. Доступ к приложению

```bash
# Получение внешнего IP
kubectl get service highload-service -n highload-microservice

# Или через port-forward
kubectl port-forward service/highload-service 8080:80 -n highload-microservice
```

## 📚 API Документация

### Базовый URL
- Локально: `http://localhost:8080`
- Kubernetes: `http://<external-ip>`

### Endpoints

#### Health Check
```http
GET /health
```

#### Пользователи

**Создание пользователя:**
```http
POST /api/v1/users
Content-Type: application/json

{
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Получение пользователя:**
```http
GET /api/v1/users/{id}
```

**Обновление пользователя:**
```http
PUT /api/v1/users/{id}
Content-Type: application/json

{
  "email": "newemail@example.com",
  "first_name": "Jane"
}
```

**Удаление пользователя:**
```http
DELETE /api/v1/users/{id}
```

**Список пользователей:**
```http
GET /api/v1/users?page=1&limit=10
```

#### События

**Создание события:**
```http
POST /api/v1/events
Content-Type: application/json

{
  "user_id": "uuid",
  "type": "user_action",
  "data": "{\"action\": \"login\"}"
}
```

**Получение события:**
```http
GET /api/v1/events/{id}
```

**Список событий:**
```http
GET /api/v1/events?page=1&limit=10
```

## 🏗 Архитектура

### Компоненты системы

1. **HTTP API** - RESTful интерфейс для взаимодействия с клиентами
2. **User Service** - управление пользователями с кэшированием
3. **Event Service** - обработка событий и интеграция с Kafka
4. **Worker Pool** - параллельная обработка задач с использованием горутин
5. **Database Layer** - работа с PostgreSQL
6. **Cache Layer** - Redis для быстрого доступа к данным
7. **Message Broker** - Kafka для асинхронной обработки событий

### Поток данных

```
Client → HTTP API → Service Layer → Database/Cache
                    ↓
                Kafka Producer → Kafka → Consumer → Worker Pool
```

### Особенности реализации

- **Горутины и каналы**: Worker pool для параллельной обработки событий
- **Кэширование**: Redis для ускорения доступа к часто запрашиваемым данным
- **Асинхронность**: Kafka для обработки событий без блокировки основного потока
- **Graceful shutdown**: Корректное завершение работы с ожиданием завершения горутин
- **Health checks**: Мониторинг состояния сервиса
- **Автомасштабирование**: HPA в Kubernetes для автоматического масштабирования

## 🔧 Конфигурация

### Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `SERVER_HOST` | Хост для HTTP сервера | `0.0.0.0` |
| `SERVER_PORT` | Порт для HTTP сервера | `8080` |
| `DB_HOST` | Хост PostgreSQL | `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_USER` | Пользователь PostgreSQL | `postgres` |
| `DB_PASSWORD` | Пароль PostgreSQL | `postgres` |
| `DB_NAME` | Имя базы данных | `highload_db` |
| `REDIS_HOST` | Хост Redis | `localhost` |
| `REDIS_PORT` | Порт Redis | `6379` |
| `KAFKA_BROKERS` | Брокеры Kafka | `localhost:9092` |
| `LOG_LEVEL` | Уровень логирования | `info` |

## 📊 Мониторинг и логирование

### Логирование
- Структурированные логи с использованием logrus
- Различные уровни логирования (debug, info, warn, error)
- Логирование всех операций с пользователями и событиями

### Health Checks
- HTTP endpoint `/health` для проверки состояния
- Проверки подключения к базе данных, Redis и Kafka
- Kubernetes liveness и readiness probes

### Метрики
- Количество обработанных запросов
- Время отклика API
- Использование ресурсов (CPU, память)
- Статистика Kafka (сообщения отправлены/получены)

## 🚀 Производительность

### Оптимизации
- **Connection pooling** для PostgreSQL
- **Кэширование** часто запрашиваемых данных в Redis
- **Параллельная обработка** с использованием worker pool
- **Batch операции** для Kafka
- **Индексы** в базе данных для быстрого поиска

### Масштабирование
- **Горизонтальное масштабирование** в Kubernetes
- **Автомасштабирование** на основе CPU и памяти
- **Load balancing** для распределения нагрузки
- **Graceful shutdown** для обновлений без простоя

## 🧪 Тестирование

### Запуск тестов
```bash
# Все тесты
go test ./...

# Тесты с покрытием
go test -cover ./...

# Бенчмарки
go test -bench ./...
```

### Нагрузочное тестирование
```bash
# Установка hey (HTTP load testing tool)
go install github.com/rakyll/hey@latest

# Тест создания пользователей
hey -n 1000 -c 10 -m POST -H "Content-Type: application/json" -d '{"email":"test@example.com","first_name":"Test","last_name":"User"}' http://localhost:8080/api/v1/users

# Тест получения пользователей
hey -n 1000 -c 10 http://localhost:8080/api/v1/users
```

## 🔒 Безопасность

### Реализованные меры
- **Валидация входных данных** на всех уровнях
- **SQL injection protection** через prepared statements
- **Non-root пользователь** в Docker контейнере
- **Secrets management** в Kubernetes
- **Rate limiting** через Ingress

### Рекомендации для продакшена
- Использование HTTPS/TLS
- Аутентификация и авторизация (JWT, OAuth2)
- Валидация и санитизация всех входных данных
- Мониторинг безопасности
- Регулярные обновления зависимостей

## 📈 Мониторинг в продакшене

### Рекомендуемые инструменты
- **Prometheus** + **Grafana** для метрик
- **ELK Stack** (Elasticsearch, Logstash, Kibana) для логов
- **Jaeger** для трейсинга
- **AlertManager** для уведомлений

### Ключевые метрики
- RPS (Requests Per Second)
- Latency (P50, P95, P99)
- Error rate
- Resource utilization (CPU, Memory)
- Database connection pool status
- Kafka lag

## 🤝 Разработка

### Структура проекта
```
.
├── cmd/                    # Точки входа приложения
├── internal/               # Внутренние пакеты
│   ├── config/            # Конфигурация
│   ├── database/          # Работа с БД
│   ├── handlers/          # HTTP обработчики
│   ├── kafka/             # Kafka клиенты
│   ├── models/            # Модели данных
│   ├── redis/             # Redis клиент
│   ├── services/          # Бизнес-логика
│   └── worker/            # Worker pool
├── k8s/                   # Kubernetes манифесты
├── docker-compose.yml     # Docker Compose конфигурация
├── Dockerfile            # Docker образ
└── README.md             # Документация
```

### Соглашения
- **Именование**: camelCase для переменных, PascalCase для типов
- **Обработка ошибок**: всегда проверяем и логируем ошибки
- **Контекст**: используем context.Context для отмены операций
- **Горутины**: всегда используем WaitGroup для синхронизации
- **Ресурсы**: закрываем все ресурсы (defer)

## 📝 Лицензия

MIT License

## 👥 Авторы

- Ваше имя - *Начальная разработка*

## 🙏 Благодарности

- Go community за отличную экосистему
- Docker и Kubernetes за инструменты контейнеризации
- PostgreSQL, Redis и Kafka за надежные хранилища данных


