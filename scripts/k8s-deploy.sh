#!/bin/bash

# Скрипт для развертывания в Kubernetes

set -e

echo "🚀 Развертывание высоконагруженного микросервиса в Kubernetes..."

# Проверка наличия kubectl
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl не найден. Пожалуйста, установите kubectl."
    exit 1
fi

# Проверка подключения к кластеру
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Не удается подключиться к Kubernetes кластеру."
    echo "Убедитесь, что кластер запущен и kubectl настроен правильно."
    exit 1
fi

# Создание namespace
echo "📦 Создание namespace..."
kubectl apply -f k8s/namespace.yaml

# Применение секретов
echo "🔐 Применение секретов..."
kubectl apply -f k8s/secret.yaml

# Применение конфигурации
echo "⚙️ Применение конфигурации..."
kubectl apply -f k8s/configmap.yaml

# Развертывание PostgreSQL
echo "🐘 Развертывание PostgreSQL..."
kubectl apply -f k8s/postgres-deployment.yaml

# Ожидание готовности PostgreSQL
echo "⏳ Ожидание готовности PostgreSQL..."
kubectl wait --for=condition=ready pod -l app=postgres -n highload-microservice --timeout=300s

# Развертывание Redis
echo "🔴 Развертывание Redis..."
kubectl apply -f k8s/redis-deployment.yaml

# Ожидание готовности Redis
echo "⏳ Ожидание готовности Redis..."
kubectl wait --for=condition=ready pod -l app=redis -n highload-microservice --timeout=300s

# Развертывание Kafka
echo "📨 Развертывание Kafka..."
kubectl apply -f k8s/kafka-deployment.yaml

# Ожидание готовности Kafka
echo "⏳ Ожидание готовности Kafka..."
kubectl wait --for=condition=ready pod -l app=kafka -n highload-microservice --timeout=300s

# Сборка Docker образа (если используется minikube)
if command -v minikube &> /dev/null && minikube status &> /dev/null; then
    echo "🔨 Сборка Docker образа для minikube..."
    eval $(minikube docker-env)
    docker build -t highload-microservice:latest .
fi

# Развертывание приложения
echo "🚀 Развертывание приложения..."
kubectl apply -f k8s/app-deployment.yaml

# Ожидание готовности приложения
echo "⏳ Ожидание готовности приложения..."
kubectl wait --for=condition=ready pod -l app=highload-microservice -n highload-microservice --timeout=300s

# Применение Ingress
echo "🌐 Применение Ingress..."
kubectl apply -f k8s/ingress.yaml

# Получение информации о сервисах
echo "📊 Информация о развертывании:"
echo ""
echo "🔍 Статус подов:"
kubectl get pods -n highload-microservice

echo ""
echo "🌐 Сервисы:"
kubectl get services -n highload-microservice

echo ""
echo "📈 HPA:"
kubectl get hpa -n highload-microservice

# Получение внешнего IP
echo ""
echo "🌍 Доступ к приложению:"
EXTERNAL_IP=$(kubectl get service highload-service -n highload-microservice -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
if [ -n "$EXTERNAL_IP" ]; then
    echo "  Внешний IP: http://$EXTERNAL_IP"
else
    echo "  Используйте port-forward для доступа:"
    echo "  kubectl port-forward service/highload-service 8080:80 -n highload-microservice"
    echo "  Затем откройте: http://localhost:8080"
fi

echo ""
echo "🎉 Развертывание завершено!"
echo ""
echo "📋 Полезные команды:"
echo "  Просмотр логов: kubectl logs -f deployment/highload-microservice -n highload-microservice"
echo "  Масштабирование: kubectl scale deployment highload-microservice --replicas=5 -n highload-microservice"
echo "  Удаление: kubectl delete namespace highload-microservice"
echo ""
echo "🧪 Тестирование:"
echo "  Health check: curl http://$EXTERNAL_IP/health"
echo "  Или через port-forward: curl http://localhost:8080/health"


