@echo off
echo Docker Container'lari Durduruluyor...
echo ====================================

docker-compose down

if %ERRORLEVEL% EQU 0 (
    echo ✓ Tum container'lar basariyla durduruldu.
) else (
    echo HATA: Container'lar durdurulamadi!
)

echo.
pause
