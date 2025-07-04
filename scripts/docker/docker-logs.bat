@echo off
echo Docker Container Logları
echo =======================
echo.
echo Hangi servisin loglarını gormek istiyorsunuz?
echo 1. API (ecommerce_api)
echo 2. Database (ecommerce_db)
echo 3. Adminer (ecommerce_adminer)
echo 4. Tum servisler
echo.

set /p choice="Seciminizi yapin (1-4): "

if "%choice%"=="1" (
    echo API logları:
    docker logs -f ecommerce_api
) else if "%choice%"=="2" (
    echo Database logları:
    docker logs -f ecommerce_db
) else if "%choice%"=="3" (
    echo Adminer logları:
    docker logs -f ecommerce_adminer
) else if "%choice%"=="4" (
    echo Tum servis logları:
    docker-compose logs -f
) else (
    echo Gecersiz secim!
    pause
)

pause
