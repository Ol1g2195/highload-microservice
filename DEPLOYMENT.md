# Инструкции по развертыванию

## 🚀 Быстрый старт

### 1. Локальная разработка с Docker Compose

```bash
# Клонирование репозитория
git clone <repository-url>
cd highload-microservice

# Запуск всех сервисов
./scripts/start.sh  # Linux/Mac
# или
scripts\start.bat   # Windows

# Проверка статуса
docker-compose ps

# Просмотр логов
docker-compose logs -f
```

### 2. Развертывание в Kubernetes

```bash
# Развертывание в Kubernetes
./scripts/k8s-deploy.sh  # Linux/Mac
# или
kubectl apply -f k8s/    # Ручное развертывание

# Проверка статуса
kubectl get pods -n highload-microservice
```

## 📋 Подробные инструкции

### Требования к системе

#### Минимальные требования:
- **CPU**: 2 ядра
- **RAM**: 4 GB
- **Диск**: 10 GB свободного места
- **Сеть**: Доступ к интернету для загрузки образов

#### Рекомендуемые требования:
- **CPU**: 4+ ядер
- **RAM**: 8+ GB
- **Диск**: 50+ GB SSD
- **Сеть**: Стабильное подключение

### Установка зависимостей

#### 1. Docker и Docker Compose

**Ubuntu/Debian:**
```bash
# Обновление пакетов
sudo apt update

# Установка Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Добавление пользователя в группу docker
sudo usermod -aG docker $USER

# Установка Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Перезагрузка для применения изменений
sudo reboot
```

**Windows:**
1. Скачайте Docker Desktop с официального сайта
2. Установите и запустите Docker Desktop
3. Включите WSL 2 backend (рекомендуется)

**macOS:**
1. Скачайте Docker Desktop с официального сайта
2. Установите и запустите Docker Desktop

#### 2. Go (для разработки)

```bash
# Установка Go 1.21+
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# Добавление в PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Проверка установки
go version
```

#### 3. Kubernetes (для развертывания)

**Minikube (локальная разработка):**
```bash
# Установка Minikube
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# Запуск Minikube
minikube start --driver=docker

# Проверка статуса
kubectl get nodes
```

**Kind (альтернатива Minikube):**
```bash
# Установка Kind
go install sigs.k8s.io/kind@v0.20.0

# Создание кластера
kind create cluster --name highload

# Проверка статуса
kubectl get nodes
```

### Конфигурация

#### 1. Переменные окружения

Скопируйте файл конфигурации:
```bash
cp env.example .env
```

Отредактируйте `.env` файл:
```bash
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=highload_db
DB_SSLMODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0

# Kafka Configuration
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=user-events
KAFKA_GROUP_ID=highload-service

# Logging
LOG_LEVEL=info
```

#### 2. Настройка Kubernetes

Для продакшена создайте отдельные секреты:
```bash
# Создание секретов
kubectl create secret generic highload-secret \
  --from-literal=DB_PASSWORD=your_secure_password \
  --from-literal=REDIS_PASSWORD=your_redis_password \
  -n highload-microservice
```

### Развертывание

#### 1. Docker Compose (рекомендуется для разработки)

```bash
# Запуск всех сервисов
docker-compose up -d

# Проверка статуса
docker-compose ps

# Просмотр логов
docker-compose logs -f app

# Остановка
docker-compose down
```

#### 2. Kubernetes (продакшен)

```bash
# Создание namespace
kubectl create namespace highload-microservice

# Применение манифестов
kubectl apply -f k8s/

# Проверка развертывания
kubectl get all -n highload-microservice

# Получение внешнего IP
kubectl get service highload-service -n highload-microservice
```

#### 3. Helm (альтернативный способ)

Создайте `values.yaml`:
```yaml
replicaCount: 3

image:
  repository: highload-microservice
  tag: latest
  pullPolicy: IfNotPresent

service:
  type: LoadBalancer
  port: 80

ingress:
  enabled: true
  host: highload-microservice.local

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
```

Развертывание:
```bash
# Установка Helm chart
helm install highload-microservice ./helm-chart -f values.yaml

# Обновление
helm upgrade highload-microservice ./helm-chart -f values.yaml

# Удаление
helm uninstall highload-microservice
```

### Мониторинг и логирование

#### 1. Prometheus + Grafana

```bash
# Установка Prometheus
kubectl apply -f monitoring/prometheus.yaml

# Установка Grafana
kubectl apply -f monitoring/grafana.yaml

# Доступ к Grafana
kubectl port-forward service/grafana 3000:80 -n monitoring
```

#### 2. ELK Stack

```bash
# Установка Elasticsearch
kubectl apply -f logging/elasticsearch.yaml

# Установка Logstash
kubectl apply -f logging/logstash.yaml

# Установка Kibana
kubectl apply -f logging/kibana.yaml
```

### Безопасность

#### 1. TLS/SSL сертификаты

```bash
# Создание сертификата
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout tls.key -out tls.crt \
  -subj "/CN=highload-microservice.local"

# Создание Kubernetes secret
kubectl create secret tls highload-tls \
  --key tls.key --cert tls.crt \
  -n highload-microservice
```

#### 2. Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: highload-network-policy
  namespace: highload-microservice
spec:
  podSelector:
    matchLabels:
      app: highload-microservice
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: highload-microservice
    ports:
    - protocol: TCP
      port: 5432  # PostgreSQL
    - protocol: TCP
      port: 6379  # Redis
    - protocol: TCP
      port: 9092  # Kafka
```

### Масштабирование

#### 1. Горизонтальное масштабирование

```bash
# Ручное масштабирование
kubectl scale deployment highload-microservice --replicas=5 -n highload-microservice

# Автоматическое масштабирование (HPA)
kubectl apply -f k8s/hpa.yaml
```

#### 2. Вертикальное масштабирование

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: highload-microservice
spec:
  template:
    spec:
      containers:
      - name: highload-microservice
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
```

### Резервное копирование

#### 1. База данных

```bash
# Создание бэкапа PostgreSQL
kubectl exec -it postgres-0 -n highload-microservice -- \
  pg_dump -U postgres highload_db > backup.sql

# Восстановление из бэкапа
kubectl exec -i postgres-0 -n highload-microservice -- \
  psql -U postgres highload_db < backup.sql
```

#### 2. Конфигурация

```bash
# Экспорт конфигурации
kubectl get configmap highload-config -n highload-microservice -o yaml > config-backup.yaml

# Импорт конфигурации
kubectl apply -f config-backup.yaml
```

### Устранение неполадок

#### 1. Проверка статуса

```bash
# Статус подов
kubectl get pods -n highload-microservice

# Логи приложения
kubectl logs -f deployment/highload-microservice -n highload-microservice

# Описание пода
kubectl describe pod <pod-name> -n highload-microservice
```

#### 2. Частые проблемы

**Проблема**: Поды не запускаются
```bash
# Проверка событий
kubectl get events -n highload-microservice

# Проверка ресурсов
kubectl top nodes
kubectl top pods -n highload-microservice
```

**Проблема**: База данных недоступна
```bash
# Проверка подключения
kubectl exec -it postgres-0 -n highload-microservice -- psql -U postgres -c "SELECT 1"

# Проверка логов PostgreSQL
kubectl logs postgres-0 -n highload-microservice
```

**Проблема**: Redis недоступен
```bash
# Проверка подключения
kubectl exec -it redis-0 -n highload-microservice -- redis-cli ping

# Проверка логов Redis
kubectl logs redis-0 -n highload-microservice
```

### Обновление

#### 1. Rolling Update

```bash
# Обновление образа
kubectl set image deployment/highload-microservice \
  highload-microservice=highload-microservice:v2.0.0 \
  -n highload-microservice

# Проверка статуса обновления
kubectl rollout status deployment/highload-microservice -n highload-microservice

# Откат к предыдущей версии
kubectl rollout undo deployment/highload-microservice -n highload-microservice
```

#### 2. Blue-Green Deployment

```bash
# Развертывание новой версии
kubectl apply -f k8s/app-deployment-v2.yaml

# Переключение трафика
kubectl patch service highload-service -n highload-microservice \
  -p '{"spec":{"selector":{"version":"v2"}}}'

# Удаление старой версии
kubectl delete deployment highload-microservice-v1 -n highload-microservice
```

### Мониторинг производительности

#### 1. Метрики приложения

```bash
# CPU и память
kubectl top pods -n highload-microservice

# Сетевой трафик
kubectl get --raw /apis/metrics.k8s.io/v1beta1/pods | jq '.'
```

#### 2. Логирование

```bash
# Централизованные логи
kubectl logs -f deployment/highload-microservice -n highload-microservice | \
  grep -E "(ERROR|WARN|INFO)"

# Логи по компонентам
kubectl logs -f deployment/highload-microservice -n highload-microservice | \
  grep "user-service"
```

### Поддержка и обслуживание

#### 1. Регулярные задачи

- **Еженедельно**: Проверка логов на ошибки
- **Ежемесячно**: Обновление зависимостей
- **Ежеквартально**: Аудит безопасности

#### 2. Мониторинг

- **Uptime**: 99.9%+
- **Response time**: < 100ms (P95)
- **Error rate**: < 0.1%
- **CPU usage**: < 70%
- **Memory usage**: < 80%

#### 3. Алерты

Настройте алерты для:
- Высокого использования CPU/памяти
- Ошибок в логах
- Недоступности сервиса
- Медленного отклика API


