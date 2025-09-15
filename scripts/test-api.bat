@echo off
REM Скрипт для тестирования API на Windows

set API_URL=%1
if "%API_URL%"=="" set API_URL=http://localhost:8080

echo 🧪 Тестирование API высоконагруженного микросервиса...
echo 🌐 API URL: %API_URL%

REM Проверка health check
echo 1. Проверка health check...
curl -s http://%API_URL%/health
echo.
echo.

REM Создание пользователя
echo 2. Создание пользователя...
curl -s -X POST -H "Content-Type: application/json" -d "{\"email\":\"test@example.com\",\"first_name\":\"Test\",\"last_name\":\"User\"}" http://%API_URL%/api/v1/users
echo.
echo.

REM Получение списка пользователей
echo 3. Получение списка пользователей...
curl -s http://%API_URL%/api/v1/users?page=1^&limit=10
echo.
echo.

REM Создание события
echo 4. Создание события...
curl -s -X POST -H "Content-Type: application/json" -d "{\"user_id\":\"00000000-0000-0000-0000-000000000000\",\"type\":\"test_event\",\"data\":\"{\\\"test\\\": \\\"data\\\"}\"}" http://%API_URL%/api/v1/events
echo.
echo.

REM Получение списка событий
echo 5. Получение списка событий...
curl -s http://%API_URL%/api/v1/events?page=1^&limit=10
echo.
echo.

echo 🎉 Тестирование завершено!
echo.
echo 💡 Для более детального тестирования используйте PowerShell или установите curl для Windows.

pause


