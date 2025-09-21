# GitHub Packages

Этот проект использует GitHub Packages для хранения Docker образов.

## Доступные образы

### Основной образ
```bash
ghcr.io/oleg2195/highload-microservice:latest
ghcr.io/oleg2195/highload-microservice:v1.0.0
```

### Теги версий
- `latest` - последняя версия из main ветки
- `v1.0.0` - стабильная версия 1.0.0
- `v1.0` - последняя версия 1.x
- `v1` - последняя версия 1.x.x

## Использование

### 1. Локальное использование
```bash
# Скачать образ
docker pull ghcr.io/oleg2195/highload-microservice:latest

# Запустить контейнер
docker run -p 8080:8080 ghcr.io/oleg2195/highload-microservice:latest
```

### 2. Docker Compose
```yaml
version: '3.8'
services:
  app:
    image: ghcr.io/oleg2195/highload-microservice:latest
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - KAFKA_BROKERS=kafka:9092
```

### 3. Kubernetes
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: highload-microservice
spec:
  replicas: 3
  selector:
    matchLabels:
      app: highload-microservice
  template:
    metadata:
      labels:
        app: highload-microservice
    spec:
      containers:
      - name: app
        image: ghcr.io/oleg2195/highload-microservice:latest
        ports:
        - containerPort: 8080
```

## Аутентификация

### Для публичного доступа (только чтение)
```bash
# Никакой аутентификации не требуется для публичных пакетов
docker pull ghcr.io/oleg2195/highload-microservice:latest
```

### Для приватных пакетов
```bash
# Войти в GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Скачать приватный образ
docker pull ghcr.io/oleg2195/highload-microservice:latest
```

## Автоматическое обновление

GitHub Actions автоматически:
1. Собирает Docker образ при каждом push в main
2. Публикует образ в GitHub Packages
3. Создает релиз при создании тега
4. Обновляет теги версий

## Мониторинг

- **Пакеты**: https://github.com/Ol1g2195/highload-microservice/pkgs/container/highload-microservice
- **Релизы**: https://github.com/Ol1g2195/highload-microservice/releases
- **Actions**: https://github.com/Ol1g2195/highload-microservice/actions

## Размеры образов

- **Полный образ**: ~15MB (Alpine Linux + Go binary)
- **Сжатый**: ~5MB (gzip)
- **Поддерживаемые архитектуры**: linux/amd64, linux/arm64
