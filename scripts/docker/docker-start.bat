@echo off
echo Docker E-Commerce Projesi Kurulum
echo ==================================

REM Docker ve Docker Compose'un yuklu olup olmadigini kontrol et
where docker >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo HATA: Docker yuklu degil!
    echo.
    echo Docker Desktop'i yuklemek icin:
    echo 1. https://www.docker.com/products/docker-desktop adresine gidin
    echo 2. Windows icin Docker Desktop'i indirin ve kurun
    echo 3. Docker Desktop'i baslatip whale ikonunun sistem tepsisinde gorunmesini bekleyin
    echo 4. Bu scripti tekrar calistirin
    echo.
    pause
    exit /b 1
)

where docker-compose >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo HATA: Docker Compose yuklu degil!
    echo Docker Desktop ile birlikte kurulmus olmali.
    pause
    exit /b 1
)

echo Docker versiyonu:
docker --version
echo.

echo Docker Compose versiyonu:
docker-compose --version
echo.

echo Container'lari olusturuyor ve baslatiliyor...
docker-compose up -d --build

if %ERRORLEVEL% NEQ 0 (
    echo HATA: Container'lar baslatiliamadi!
    pause
    exit /b 1
)

echo.
echo ✓ Proje basariyla Docker'da baslatildi!
echo.
echo Servisler:
echo - API: http://localhost:8080
echo - Adminer (DB Yonetimi): http://localhost:8081
echo - PostgreSQL: localhost:5432
echo.
echo Durdurmak icin: docker-stop.bat
echo Logları gormek icin: docker-logs.bat
echo.
pause
