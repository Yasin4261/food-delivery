@echo off
echo E-Commerce API Baslatiliyor...
echo =============================

REM Go'nun yuklu olup olmadigini kontrol et
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo HATA: Go yuklu degil! Once setup.bat dosyasini calistirin.
    pause
    exit /b 1
)

echo API serveri 8080 portunda baslatiliyor...
echo Durdurmak icin Ctrl+C basin
echo.

go run cmd/main.go

pause
