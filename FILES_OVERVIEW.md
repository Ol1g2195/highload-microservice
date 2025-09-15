# Обзор файлов проекта

## 📁 Структура проекта

```
highload-microservice/
├── 📄 main.go                           # Точка входа приложения
├── 📄 go.mod                            # Go модули и зависимости
├── 📄 go.sum                            # Хеши зависимостей
├── 📄 Dockerfile                        # Docker образ
├── 📄 docker-compose.yml                # Docker Compose конфигурация
├── 📄 .dockerignore                     # Исключения для Docker
├── 📄 env.example                       # Пример переменных окружения
├── 📄 Makefile                          # Команды для разработки
├── 📁 internal/                         # Внутренние пакеты
│   ├── 📁 config/                       # Конфигурация
│   │   └── 📄 config.go                 # Загрузка конфигурации
│   ├── 📁 database/                     # Работа с базой данных
│   │   ├── 📄 connection.go             # Подключение к PostgreSQL
│   │   └── 📄 migrations.sql            # SQL миграции
│   ├── 📁 handlers/                     # HTTP обработчики
│   │   ├── 📄 user_handler.go           # Обработчики пользователей
│   │   └── 📄 event_handler.go          # Обработчики событий
│   ├── 📁 kafka/                        # Kafka клиенты
│   │   ├── 📄 producer.go               # Kafka producer
│   │   └── 📄 consumer.go               # Kafka consumer
│   ├── 📁 models/                       # Модели данных
│   │   ├── 📄 user.go                   # Модели пользователей
│   │   └── 📄 event.go                  # Модели событий
│   ├── 📁 redis/                        # Redis клиент
│   │   └── 📄 client.go                 # Redis клиент
│   ├── 📁 services/                     # Бизнес-логика
│   │   ├── 📄 user_service.go           # Сервис пользователей
│   │   └── 📄 event_service.go          # Сервис событий
│   └── 📁 worker/                       # Worker pool
│       └── 📄 pool.go                   # Пул горутин
├── 📁 k8s/                             # Kubernetes манифесты
│   ├── 📄 namespace.yaml                # Namespace
│   ├── 📄 configmap.yaml                # ConfigMap
│   ├── 📄 secret.yaml                   # Secret
│   ├── 📄 postgres-deployment.yaml      # PostgreSQL
│   ├── 📄 redis-deployment.yaml         # Redis
│   ├── 📄 kafka-deployment.yaml         # Kafka
│   ├── 📄 app-deployment.yaml           # Приложение
│   └── 📄 ingress.yaml                  # Ingress
├── 📁 scripts/                         # Скрипты запуска
│   ├── 📄 start.sh                      # Запуск (Linux/Mac)
│   ├── 📄 start.bat                     # Запуск (Windows)
│   ├── 📄 k8s-deploy.sh                 # Развертывание K8s
│   ├── 📄 test-api.sh                   # Тестирование API
│   └── 📄 test-api.bat                  # Тестирование API (Windows)
├── 📁 examples/                        # Примеры использования
│   └── 📄 api-examples.md               # Примеры API запросов
└── 📁 docs/                            # Документация
    ├── 📄 README.md                     # Основная документация
    ├── 📄 QUICKSTART.md                 # Быстрый старт
    ├── 📄 GETTING_STARTED.md            # Пошаговое руководство
    ├── 📄 DEPLOYMENT.md                 # Инструкции по развертыванию
    ├── 📄 INSTALL.md                    # Установка и настройка
    ├── 📄 PROJECT_OVERVIEW.md           # Обзор проекта
    └── 📄 FILES_OVERVIEW.md             # Этот файл
```

## 📄 Описание файлов

### Основные файлы

#### `main.go`
- **Назначение**: Точка входа приложения
- **Содержит**: Инициализацию всех компонентов, настройку HTTP сервера, graceful shutdown
- **Ключевые функции**: `main()`, инициализация сервисов, настройка роутинга

#### `go.mod`
- **Назначение**: Управление зависимостями Go
- **Содержит**: Список всех внешних пакетов и их версий
- **Зависимости**: Gin, PostgreSQL, Redis, Kafka, Logrus, UUID

#### `Dockerfile`
- **Назначение**: Создание Docker образа
- **Содержит**: Многоэтапная сборка, настройка безопасности, health check
- **Особенности**: Non-root пользователь, оптимизированный размер

#### `docker-compose.yml`
- **Назначение**: Локальная разработка с Docker
- **Содержит**: Все сервисы (PostgreSQL, Redis, Kafka, приложение)
- **Особенности**: Health checks, volumes, networks

### Внутренние пакеты (`internal/`)

#### `config/config.go`
- **Назначение**: Загрузка и управление конфигурацией
- **Содержит**: Структуры конфигурации, загрузка из .env файла
- **Функции**: `Load()`, `getEnv()`, `getEnvAsInt()`

#### `database/connection.go`
- **Назначение**: Подключение к PostgreSQL
- **Содержит**: Настройка connection pool, проверка подключения
- **Функции**: `NewConnection()`

#### `database/migrations.sql`
- **Назначение**: SQL миграции для базы данных
- **Содержит**: Создание таблиц, индексов, триггеров
- **Таблицы**: `users`, `events`

#### `handlers/user_handler.go`
- **Назначение**: HTTP обработчики для пользователей
- **Содержит**: CRUD операции, валидация, обработка ошибок
- **Endpoints**: POST, GET, PUT, DELETE для пользователей

#### `handlers/event_handler.go`
- **Назначение**: HTTP обработчики для событий
- **Содержит**: Создание и получение событий
- **Endpoints**: POST, GET для событий

#### `kafka/producer.go`
- **Назначение**: Отправка сообщений в Kafka
- **Содержит**: Настройка producer, отправка событий
- **Функции**: `NewProducer()`, `SendEvent()`

#### `kafka/consumer.go`
- **Назначение**: Получение сообщений из Kafka
- **Содержит**: Настройка consumer, чтение сообщений
- **Функции**: `NewConsumer()`, `ReadMessage()`

#### `models/user.go`
- **Назначение**: Модели данных для пользователей
- **Содержит**: Структуры User, CreateUserRequest, UpdateUserRequest
- **Особенности**: JSON теги, валидация

#### `models/event.go`
- **Назначение**: Модели данных для событий
- **Содержит**: Структуры Event, CreateEventRequest, KafkaEvent
- **Особенности**: UUID, временные метки

#### `redis/client.go`
- **Назначение**: Клиент для Redis
- **Содержит**: Операции с кэшем, подключение
- **Функции**: `Set()`, `Get()`, `Del()`, `Exists()`

#### `services/user_service.go`
- **Назначение**: Бизнес-логика для пользователей
- **Содержит**: CRUD операции, кэширование, отправка событий
- **Функции**: `CreateUser()`, `GetUser()`, `UpdateUser()`, `DeleteUser()`

#### `services/event_service.go`
- **Назначение**: Бизнес-логика для событий
- **Содержит**: Создание событий, обработка через Kafka
- **Функции**: `CreateEvent()`, `ProcessEvents()`

#### `worker/pool.go`
- **Назначение**: Управление горутинами
- **Содержит**: Worker pool, каналы, синхронизация
- **Функции**: `Start()`, `Stop()`, `AddJob()`

### Kubernetes манифесты (`k8s/`)

#### `namespace.yaml`
- **Назначение**: Создание namespace для приложения
- **Содержит**: Определение namespace `highload-microservice`

#### `configmap.yaml`
- **Назначение**: Конфигурация приложения
- **Содержит**: Переменные окружения для всех сервисов

#### `secret.yaml`
- **Назначение**: Секретные данные
- **Содержит**: Пароли для базы данных и Redis

#### `postgres-deployment.yaml`
- **Назначение**: Развертывание PostgreSQL
- **Содержит**: Deployment, Service, PVC, ConfigMap с миграциями

#### `redis-deployment.yaml`
- **Назначение**: Развертывание Redis
- **Содержит**: Deployment, Service, PVC

#### `kafka-deployment.yaml`
- **Назначение**: Развертывание Kafka и Zookeeper
- **Содержит**: Deployments и Services для Kafka и Zookeeper

#### `app-deployment.yaml`
- **Назначение**: Развертывание приложения
- **Содержит**: Deployment, Service, HPA для автомасштабирования

#### `ingress.yaml`
- **Назначение**: Внешний доступ к приложению
- **Содержит**: Ingress с rate limiting

### Скрипты (`scripts/`)

#### `start.sh` / `start.bat`
- **Назначение**: Запуск приложения локально
- **Содержит**: Проверка зависимостей, запуск Docker Compose, проверка health

#### `k8s-deploy.sh`
- **Назначение**: Развертывание в Kubernetes
- **Содержит**: Применение манифестов, проверка статуса, получение доступа

#### `test-api.sh` / `test-api.bat`
- **Назначение**: Тестирование API
- **Содержит**: Автоматические тесты всех endpoints

### Документация

#### `README.md`
- **Назначение**: Основная документация проекта
- **Содержит**: Описание, установка, использование, архитектура

#### `QUICKSTART.md`
- **Назначение**: Быстрый старт за 5 минут
- **Содержит**: Краткие инструкции по запуску

#### `GETTING_STARTED.md`
- **Назначение**: Пошаговое руководство
- **Содержит**: Детальные инструкции для новичков

#### `DEPLOYMENT.md`
- **Назначение**: Инструкции по развертыванию
- **Содержит**: Docker, Kubernetes, мониторинг, безопасность

#### `INSTALL.md`
- **Назначение**: Установка и настройка
- **Содержит**: Установка всех зависимостей, настройка окружения

#### `PROJECT_OVERVIEW.md`
- **Назначение**: Обзор проекта
- **Содержит**: Архитектура, технологии, возможности

#### `examples/api-examples.md`
- **Назначение**: Примеры использования API
- **Содержит**: Примеры запросов на разных языках

## 🔧 Как использовать файлы

### Для разработки
1. **Изучите** `main.go` для понимания архитектуры
2. **Настройте** `.env` на основе `env.example`
3. **Запустите** через `scripts/start.sh` или `docker-compose up`
4. **Тестируйте** через `scripts/test-api.sh`

### Для развертывания
1. **Настройте** Kubernetes кластер
2. **Примените** манифесты из `k8s/`
3. **Проверьте** статус через `kubectl get pods`

### Для изучения
1. **Начните** с `README.md`
2. **Изучите** `PROJECT_OVERVIEW.md`
3. **Попробуйте** примеры из `examples/api-examples.md`

## 📝 Примечания

- Все файлы содержат подробные комментарии
- Код следует Go conventions и best practices
- Документация написана на русском языке
- Скрипты работают на Linux, macOS и Windows
- Kubernetes манифесты совместимы с различными кластерами

---

**Этот обзор поможет вам быстро понять структуру проекта и найти нужные файлы для изучения или модификации.**


