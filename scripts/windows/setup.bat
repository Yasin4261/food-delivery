@echo off
echo E-Commerce Go Projesi Kurulum ve Calistirma
echo ==========================================

REM Go'nun yuklu olup olmadigini kontrol et
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo HATA: Go yuklu degil!
    echo.
    echo Go'yu yuklemek icin:
    echo 1. https://golang.org/dl/ adresine gidin
    echo 2. Windows icin en son surumu indirin
    echo 3. Yukleyiciyi calistirin ve PATH'e ekleyin
    echo 4. Terminali yeniden baslatip tekrar deneyin
    echo.
    pause
    exit /b 1
)

echo Go versiyonu:
go version
echo.

echo Go modulleri yukleniyor...
go mod tidy

if %ERRORLEVEL% NEQ 0 (
    echo HATA: Go modulleri yuklenemedi!
    pause
    exit /b 1
)

echo.
echo Proje basariyla hazirlandi!
echo.
echo Calistirmak icin: go run cmd/main.go
echo.
pause
