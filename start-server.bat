@echo off
chcp 65001 >nul
title Repository UN - Server

echo ========================================
echo   Repository UN - Build ^& Start
echo ========================================
echo.

:: Kill proses lama yang masih jalan di port 8080
echo [0/3] Menghentikan server lama...
taskkill /F /IM main.exe >nul 2>&1
timeout /t 1 /nobreak >nul
echo.

:: Step 1: Build Frontend
echo [1/3] Building frontend...
cd /d "%~dp0frontend"
call npm run build
if %ERRORLEVEL% neq 0 (
    echo.
    echo ❌ Frontend build gagal!
    pause
    exit /b 1
)
echo ✅ Frontend build selesai.
echo.

:: Step 2: Build Backend
echo [2/3] Building backend...
cd /d "%~dp0backend"
go build -o main.exe ./cmd/server
if %ERRORLEVEL% neq 0 (
    echo.
    echo ❌ Backend build gagal!
    pause
    exit /b 1
)
echo ✅ Backend build selesai.
echo.

:: Step 3: Start Server
echo [3/3] Starting server...
echo.
echo ========================================
echo   Tekan Ctrl+C untuk menghentikan
echo ========================================
echo.
main.exe
pause
