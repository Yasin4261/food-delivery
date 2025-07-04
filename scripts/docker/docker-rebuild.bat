@echo off
echo Docker Container'lari Yeniden Olusturuyor...
echo ============================================

echo Mevcut container'lari durduruyor...
docker-compose down

echo Container'lari ve image'lari yeniden olusturuyor...
docker-compose up -d --build --force-recreate

if %ERRORLEVEL% EQU 0 (
    echo ✓ Container'lar basariyla yeniden olusturuldu!
    echo.
    echo Servisler:
    echo - API: http://localhost:8080
    echo - Adminer: http://localhost:8081
    echo - PostgreSQL: localhost:5432
) else (
    echo HATA: Container'lar yeniden olusturulamadi!
)

echo.
pause
