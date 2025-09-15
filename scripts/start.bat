@echo off
REM Скрипт для запуска высоконагруженного микросервиса на Windows

echo 🚀 Запуск высоконагруженного микросервиса...

REM Проверка наличия Docker
docker --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker не найден. Пожалуйста, установите Docker Desktop.
    pause
    exit /b 1
)

REM Проверка наличия Docker Compose
docker-compose --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker Compose не найден. Пожалуйста, установите Docker Compose.
    pause
    exit /b 1
)

REM Создание .env файла если не существует
if not exist .env (
    echo 📝 Создание .env файла...
    copy env.example .env
    echo ✅ .env файл создан. Отредактируйте его при необходимости.
)

REM Остановка существующих контейнеров
echo 🛑 Остановка существующих контейнеров...
docker-compose down

REM Сборка и запуск контейнеров
echo 🔨 Сборка и запуск контейнеров...
docker-compose up --build -d

REM Ожидание готовности сервисов
echo ⏳ Ожидание готовности сервисов...
timeout /t 30 /nobreak >nul

REM Проверка статуса контейнеров
echo 📊 Статус контейнеров:
docker-compose ps

REM Проверка health check
echo 🏥 Проверка health check...
curl -f http://localhost:8080/health >nul 2>&1
if errorlevel 1 (
    echo ❌ Микросервис не отвечает на health check
    echo 📋 Логи приложения:
    docker-compose logs app
    pause
    exit /b 1
) else (
    echo ✅ Микросервис запущен и готов к работе!
    echo 🌐 API доступно по адресу: http://localhost:8080
    echo 📚 Документация API: http://localhost:8080/health
)

echo.
echo 🎉 Готово! Микросервис успешно запущен.
echo.
echo 📋 Полезные команды:
echo   Просмотр логов: docker-compose logs -f
echo   Остановка: docker-compose down
echo   Перезапуск: docker-compose restart
echo   Статус: docker-compose ps
echo.
echo 🧪 Тестирование API:
echo   Health check: curl http://localhost:8080/health
echo   Создание пользователя:
echo     curl -X POST http://localhost:8080/api/v1/users ^
echo       -H "Content-Type: application/json" ^
echo       -d "{\"email\":\"test@example.com\",\"first_name\":\"Test\",\"last_name\":\"User\"}"

pause


