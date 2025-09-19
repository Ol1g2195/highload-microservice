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

### 2. Практическое развертывание (Docker Desktop Kubernetes)

Ниже — проверенная последовательность для Docker Desktop Kubernetes (узел `desktop-control-plane`).

1) Namespace и базовые манифесты:
```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/secret.yaml -n highload-microservice
kubectl apply -f k8s/configmap.yaml -n highload-microservice
```

2) Поднять зависимости (Postgres, Redis, ZooKeeper, Kafka):
```bash
kubectl apply -f k8s/postgres-deployment.yaml -n highload-microservice
kubectl apply -f k8s/redis-deployment.yaml -n highload-microservice
kubectl apply -f k8s/kafka-deployment.yaml -n highload-microservice
```

3) Дождаться готовности зависимостей:
```bash
kubectl get pods -n highload-microservice
kubectl wait --for=condition=ready pod -l app=postgres  -n highload-microservice --timeout=600s
kubectl wait --for=condition=ready pod -l app=redis     -n highload-microservice --timeout=600s
kubectl wait --for=condition=ready pod -l app=zookeeper -n highload-microservice --timeout=900s
kubectl wait --for=condition=ready pod -l app=kafka     -n highload-microservice --timeout=900s
```

4) Сборка и публикация образа приложения в Docker Hub (неймспейс замените на свой, пример: `docker.io/oleg2195`):
```bash
docker build -t docker.io/<username>/highload-microservice:latest .
docker login -u <username>   # рекомендуется вход по Personal Access Token
docker push docker.io/<username>/highload-microservice:latest
```

Важно: Если в Docker Desktop был включён `registry-mirrors`, удалите его в Settings → Docker Engine ("registry-mirrors": []) и перезапустите Docker Desktop.

5) Указать образ приложения в деплойменте `k8s/app-deployment.yaml` (поле `image`):
```yaml
image: docker.io/<username>/highload-microservice:latest
imagePullPolicy: IfNotPresent
```

6) Применить приложение и дождаться готовности:
```bash
kubectl apply -f k8s/app-deployment.yaml -n highload-microservice
kubectl rollout restart deployment/highload-microservice -n highload-microservice
kubectl rollout status  deployment/highload-microservice -n highload-microservice
kubectl get pods -n highload-microservice
```

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

### 4. Настройки Kafka, которые мы добавили

В `k8s/kafka-deployment.yaml` настроены корректные слушатели и пробы:
```yaml
- name: KAFKA_LISTENERS
  value: "PLAINTEXT://0.0.0.0:9092"
- name: KAFKA_ADVERTISED_LISTENERS
  value: "PLAINTEXT://kafka-service:9092"
- name: KAFKA_LISTENER_SECURITY_PROTOCOL_MAP
  value: "PLAINTEXT:PLAINTEXT"
```
Пробы переведены на TCP 9092, что делает readiness/liveness стабильными.

Если образы тянутся долго или появляются ImagePullBackOff — проверьте отсутствие registry mirror и авторизацию в Docker Hub; при приватном репозитории добавьте `imagePullSecrets` в pod spec.

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

## 🧪 E2E запуск (smoke)

### Локально: Docker Compose

```bash
# 1) Запустить весь стек
docker-compose up -d

# 2) Проверить здоровье сервиса
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8080/health

# 3) Запустить smoke-тесты
# Windows PowerShell
./scripts/smoke.ps1
# Linux/macOS
bash scripts/smoke.sh

# 4) Остановить окружение
docker-compose down -v
```

### Kubernetes

```bash
# 1) Применить манифесты (если ещё не применены)
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/secret.yaml -n highload-microservice
kubectl apply -f k8s/configmap.yaml -n highload-microservice
kubectl apply -f k8s/postgres-deployment.yaml -n highload-microservice
kubectl apply -f k8s/redis-deployment.yaml -n highload-microservice
kubectl apply -f k8s/kafka-deployment.yaml -n highload-microservice
kubectl apply -f k8s/app-deployment.yaml -n highload-microservice

# 2) Дождаться готовности
kubectl wait --for=condition=ready pod -l app=highload-microservice -n highload-microservice --timeout=600s

# 3) Port-forward и запустить smoke
kubectl port-forward service/highload-service 8080:80 -n highload-microservice &
# Windows PowerShell
./scripts/smoke.ps1
# Linux/macOS
bash scripts/smoke.sh
```

### GitHub Actions

- Compose smoke: `.github/workflows/e2e-compose.yml` (сборка образа, docker-compose up, smoke)
- K8s smoke (Kind): `.github/workflows/e2e-k8s.yml` (Kind cluster, загрузка образа, манифесты, smoke)

Запуск:
- Автоматически на push/pull_request в ветку `main`
- Вручную: Actions → выбрать workflow → Run workflow

### Команды для пуша в ветку main

```bash
# Создать/переключиться на main локально
git checkout -B main

# Добавить изменения и закоммитить
git add .
git commit -m "docs: add e2e (compose+k8s) run instructions"

# Отправить в origin/main
git push -u origin main
```

## ✅ Полная проверка работоспособности (чек‑лист)

1) Поды и сервисы:
```bash
kubectl get pods -n highload-microservice
kubectl get svc  -n highload-microservice
```

2) API:
```bash
kubectl port-forward service/highload-service 8080:80 -n highload-microservice
curl http://localhost:8080/health
```

3) PostgreSQL:
```bash
kubectl exec -n highload-microservice deploy/postgres -- pg_isready -U postgres
kubectl exec -n highload-microservice deploy/postgres -- psql -U postgres -d highload_db -c "SELECT 1;"
```

4) Redis:
```bash
kubectl exec -n highload-microservice deploy/redis -- redis-cli ping
```

5) Kafka:
```bash
kubectl exec -n highload-microservice deploy/kafka -- kafka-topics --bootstrap-server kafka-service:9092 --list
```

6) Функционал:
- POST /api/v1/users — создать пользователя и убедиться, что он в БД.
- GET того же пользователя дважды — второй раз быстрее (кэш Redis).
- POST /api/v1/events — создать событие, в логах сервиса появилась обработка (worker pool + Kafka consumer).

## 🧯 Траблшутинг

- `ImagePullBackOff` у приложения — опубликуйте образ в `docker.io/<username>/highload-microservice:latest`, обновите поле `image` и перезапустите rollout.
- Ошибки Kafka при pull — проверьте, что нет `registry-mirrors` и выполнен `docker login`. При необходимости используйте облегчённую Kafka KRaft (см. комментарии в `k8s/kafka-deployment.yaml`).
- Таймауты `context deadline exceeded` у консюмера — нормальны при отсутствии новых сообщений в топике.

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



