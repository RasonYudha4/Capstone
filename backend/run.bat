@echo off
setlocal EnableDelayedExpansion

:: ============================================================
::  run.bat — Capstone dev launcher
::  - Starts all Docker services (excluding agent-service)
::  - Opens a second window tailing docker compose logs
::  - Creates/reuses local venv for agent-service
::  - Runs agent-service in foreground (uvicorn)
::  - On Ctrl+C: tears down compose
:: ============================================================

set ROOT=%~dp0
set AGENT_DIR=%ROOT%services\agent-service
set VENV_DIR=%AGENT_DIR%\venv
set ACTIVATE=%VENV_DIR%\Scripts\activate.bat
set REQUIREMENTS=%AGENT_DIR%\requirements.txt

echo.
echo =====================================================
echo   Capstone Dev Launcher
echo =====================================================
echo.

:: ------------------------------------------------------------
:: 1. Start Docker Compose (agent-service excluded via profile)
:: ------------------------------------------------------------
echo [1/4] Starting Docker services...
docker compose -f "%ROOT%docker-compose.yml" up -d
if errorlevel 1 (
    echo [ERROR] docker compose up failed. Aborting.
    pause
    exit /b 1
)
echo [OK] Docker services started.

:: Spawn a second window that tails docker compose logs
start "Docker Logs" cmd /k "docker compose -f "%ROOT%docker-compose.yml" logs -f"
echo [OK] Docker log window opened.
echo.

:: ------------------------------------------------------------
:: 2. Create venv if missing or incomplete
:: ------------------------------------------------------------
echo [2/4] Checking Python virtual environment...
if exist "%ACTIVATE%" (
    echo [OK] venv already exists, skipping creation.
) else (
    echo [INFO] venv not found or incomplete. Creating fresh venv...
    if exist "%VENV_DIR%" rmdir /s /q "%VENV_DIR%"
    python -m venv "%VENV_DIR%"
    if errorlevel 1 (
        echo [ERROR] Failed to create venv. Is Python 3.11 installed and on PATH?
        goto :teardown
    )
    echo [OK] venv created.
)
echo.

:: ------------------------------------------------------------
:: 3. Install / sync dependencies
:: ------------------------------------------------------------
echo [3/4] Installing dependencies...
call "%ACTIVATE%"
pip install --no-cache-dir --upgrade pip >nul 2>&1
pip install --no-cache-dir -r "%REQUIREMENTS%"
if errorlevel 1 (
    echo [ERROR] pip install failed.
    goto :teardown
)
echo [OK] Dependencies ready.
echo.

:: ------------------------------------------------------------
:: 4. Run agent-service in foreground
:: ------------------------------------------------------------
echo [4/4] Starting agent-service on http://0.0.0.0:8001
echo        Press Ctrl+C to stop.
echo =====================================================
echo.
cd /d "%AGENT_DIR%"
python main.py

:teardown
echo.
echo [INFO] Stopping Docker services...
docker compose -f "%ROOT%docker-compose.yml" down
echo [OK] All services stopped.
echo.
pause
exit /b 0