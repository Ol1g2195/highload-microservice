# Установка и настройка

## 📋 Требования

### Системные требования
- **Go 1.21+** - основной язык программирования
- **Docker & Docker Compose** - для контейнеризации
- **Git** - для клонирования репозитория
- **Make** (опционально) - для удобства разработки

### Для Kubernetes развертывания
- **kubectl** - для управления кластером
- **Kubernetes кластер** (minikube, kind, или облачный)

## 🚀 Установка

### 1. Установка Go

#### Windows
1. Скачайте Go с https://golang.org/dl/
2. Запустите установщик
3. Добавьте `C:\Program Files\Go\bin` в PATH

#### Linux/macOS
```bash
# Скачивание и установка
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# Добавление в PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Проверка
go version
```

### 2. Установка Docker

#### Windows/macOS
1. Скачайте Docker Desktop с https://www.docker.com/products/docker-desktop
2. Установите и запустите Docker Desktop

#### Linux (Ubuntu/Debian)
```bash
# Обновление пакетов
sudo apt update

# Установка зависимостей
sudo apt install apt-transport-https ca-certificates curl gnupg lsb-release

# Добавление GPG ключа Docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# Добавление репозитория Docker
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Установка Docker
sudo apt update
sudo apt install docker-ce docker-ce-cli containerd.io

# Добавление пользователя в группу docker
sudo usermod -aG docker $USER

# Перезагрузка
sudo reboot
```

### 3. Установка Docker Compose

#### Windows/macOS
Docker Compose входит в состав Docker Desktop

#### Linux
```bash
# Скачивание Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose

# Права на выполнение
sudo chmod +x /usr/local/bin/docker-compose

# Проверка
docker-compose --version
```

### 4. Установка kubectl (для Kubernetes)

#### Windows
1. Скачайте kubectl с https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/
2. Добавьте в PATH

#### Linux/macOS
```bash
# Скачивание kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"

# Установка
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Проверка
kubectl version --client
```

### 5. Установка Kubernetes кластера

#### Minikube (рекомендуется для разработки)
```bash
# Linux
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# macOS
brew install minikube

# Windows
# Скачайте с https://minikube.sigs.k8s.io/docs/start/

# Запуск
minikube start --driver=docker
```

#### Kind (альтернатива)
```bash
# Установка
go install sigs.k8s.io/kind@v0.20.0

# Создание кластера
kind create cluster --name highload
```

## 🔧 Настройка проекта

### 1. Клонирование репозитория
```bash
git clone <repository-url>
cd highload-microservice
```

### 2. Установка зависимостей Go
```bash
# Инициализация модуля (если нужно)
go mod init highload-microservice

# Загрузка зависимостей
go mod download

# Проверка зависимостей
go mod verify
```

### 3. Настройка переменных окружения
```bash
# Копирование файла конфигурации
cp env.example .env

# Редактирование конфигурации (опционально)
nano .env
```

### 4. Установка инструментов разработки (опционально)
```bash
# Линтер
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Нагрузочное тестирование
go install github.com/rakyll/hey@latest

# Утилиты для работы с базой данных
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## 🧪 Проверка установки

### 1. Проверка Go
```bash
go version
go env
```

### 2. Проверка Docker
```bash
docker --version
docker-compose --version
docker run hello-world
```

### 3. Проверка Kubernetes
```bash
kubectl version --client
kubectl cluster-info
```

### 4. Проверка проекта
```bash
# Сборка проекта
go build -o bin/highload-microservice .

# Запуск тестов
go test ./...

# Проверка линтера
golangci-lint run
```

## 🚀 Первый запуск

### 1. Запуск с Docker Compose
```bash
# Запуск всех сервисов
docker-compose up -d

# Проверка статуса
docker-compose ps

# Просмотр логов
docker-compose logs -f
```

### 2. Проверка работы
```bash
# Health check
curl http://localhost:8080/health

# Создание пользователя
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","first_name":"Test","last_name":"User"}'
```

### 3. Запуск в Kubernetes
```bash
# Развертывание
kubectl apply -f k8s/

# Проверка
kubectl get pods -n highload-microservice

# Доступ
kubectl port-forward service/highload-service 8080:80 -n highload-microservice
```

## 🛠 Разработка

### 1. Структура проекта
```
highload-microservice/
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
├── scripts/               # Скрипты запуска
├── examples/              # Примеры API
├── docker-compose.yml     # Docker Compose конфигурация
├── Dockerfile            # Docker образ
├── go.mod                # Go модули
├── Makefile              # Команды для разработки
└── README.md             # Документация
```

### 2. Команды для разработки
```bash
# Сборка
make build

# Запуск
make run

# Тесты
make test

# Линтер
make lint

# Форматирование
make fmt

# Docker
make docker-build
make docker-run
make docker-stop

# Kubernetes
make k8s-deploy
make k8s-status
make k8s-clean
```

### 3. Настройка IDE

#### VS Code
1. Установите расширение Go
2. Настройте `settings.json`:
```json
{
    "go.toolsManagement.checkForUpdates": "local",
    "go.useLanguageServer": true,
    "go.lintTool": "golangci-lint",
    "go.lintFlags": ["--fast"]
}
```

#### GoLand
1. Откройте проект
2. Настройте Go SDK
3. Включите Go modules

## 🔧 Конфигурация

### 1. Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `SERVER_HOST` | Хост HTTP сервера | `0.0.0.0` |
| `SERVER_PORT` | Порт HTTP сервера | `8080` |
| `DB_HOST` | Хост PostgreSQL | `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_USER` | Пользователь PostgreSQL | `postgres` |
| `DB_PASSWORD` | Пароль PostgreSQL | `postgres` |
| `DB_NAME` | Имя базы данных | `highload_db` |
| `REDIS_HOST` | Хост Redis | `localhost` |
| `REDIS_PORT` | Порт Redis | `6379` |
| `KAFKA_BROKERS` | Брокеры Kafka | `localhost:9092` |
| `LOG_LEVEL` | Уровень логирования | `info` |

### 2. Настройка базы данных

#### PostgreSQL
```sql
-- Создание базы данных
CREATE DATABASE highload_db;

-- Создание пользователя
CREATE USER postgres WITH PASSWORD 'postgres';

-- Предоставление прав
GRANT ALL PRIVILEGES ON DATABASE highload_db TO postgres;
```

#### Redis
```bash
# Запуск Redis
redis-server

# Проверка подключения
redis-cli ping
```

#### Kafka
```bash
# Запуск Zookeeper
bin/zookeeper-server-start.sh config/zookeeper.properties

# Запуск Kafka
bin/kafka-server-start.sh config/server.properties

# Создание топика
bin/kafka-topics.sh --create --topic user-events --bootstrap-server localhost:9092
```

## 🚨 Устранение неполадок

### Проблема: Go модули не загружаются
```bash
# Очистка кэша модулей
go clean -modcache

# Перезагрузка модулей
go mod download

# Проверка
go mod verify
```

### Проблема: Docker не запускается
```bash
# Проверка статуса Docker
sudo systemctl status docker

# Запуск Docker
sudo systemctl start docker

# Добавление пользователя в группу docker
sudo usermod -aG docker $USER
```

### Проблема: Kubernetes недоступен
```bash
# Проверка кластера
kubectl cluster-info

# Перезапуск Minikube
minikube stop
minikube start

# Проверка подов
kubectl get pods --all-namespaces
```

### Проблема: Порты заняты
```bash
# Проверка занятых портов
netstat -tulpn | grep :8080
netstat -tulpn | grep :5432
netstat -tulpn | grep :6379
netstat -tulpn | grep :9092

# Освобождение портов
sudo fuser -k 8080/tcp
sudo fuser -k 5432/tcp
sudo fuser -k 6379/tcp
sudo fuser -k 9092/tcp
```

## 📚 Дополнительные ресурсы

- [Go Documentation](https://golang.org/doc/)
- [Docker Documentation](https://docs.docker.com/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Redis Documentation](https://redis.io/documentation)
- [Kafka Documentation](https://kafka.apache.org/documentation/)

## 🤝 Поддержка

Если у вас возникли проблемы с установкой или настройкой:

1. Проверьте раздел "Устранение неполадок"
2. Изучите логи: `docker-compose logs`
3. Проверьте статус сервисов: `docker-compose ps`
4. Создайте issue в репозитории проекта

---

**Готово!** 🎉 Теперь вы можете приступить к разработке и развертыванию высоконагруженного микросервиса!


