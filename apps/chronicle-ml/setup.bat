@echo off
REM ChronicleCore ML Sidecar Setup Script
REM Installs Python dependencies for the ML sidecar

echo ============================================
echo ChronicleCore ML Sidecar Setup
echo ============================================
echo.

REM Check if Python is installed
python --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Python not found
    echo Please install Python 3.8+ from https://www.python.org/
    exit /b 1
)

echo Python found:
python --version
echo.

REM Check if pip is available
pip --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: pip not found
    echo Please ensure pip is installed with Python
    exit /b 1
)

echo.
echo Installing ML dependencies...
echo.

REM Install requirements
pip install -r requirements.txt

if %errorlevel% neq 0 (
    echo.
    echo ERROR: Failed to install dependencies
    exit /b 1
)

echo.
echo ============================================
echo Setup complete!
echo ============================================
echo.
echo To run the ML sidecar manually:
echo   python -m uvicorn src.main:app --host 127.0.0.1 --port 8081
echo.
echo To run tests:
echo   pytest tests/
echo.
