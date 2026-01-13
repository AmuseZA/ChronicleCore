# Simple Deployment Package Creator
# Just creates a folder you can ZIP and send

$ErrorActionPreference = "Stop"

Write-Host "Creating deployment package..." -ForegroundColor Cyan

# Check backend
if (-not (Test-Path "apps\chroniclecore-core\chroniclecore.exe")) {
    Write-Host "ERROR: Backend not found!" -ForegroundColor Red
    Write-Host "Build it first with: cd apps\chroniclecore-core && .\build.bat" -ForegroundColor Yellow
    exit 1
}

# Clean and create
if (Test-Path "deploy") { Remove-Item "deploy" -Recurse -Force }
New-Item -ItemType Directory -Path "deploy\ChronicleCore" -Force | Out-Null

# Copy files
Write-Host "Copying backend..." -ForegroundColor Yellow
Copy-Item "apps\chroniclecore-core\chroniclecore.exe" "deploy\ChronicleCore\"

Write-Host "Copying ML code..." -ForegroundColor Yellow
Copy-Item "apps\chronicle-ml" "deploy\ChronicleCore\ml" -Recurse

Write-Host "Copying schema..." -ForegroundColor Yellow
New-Item -ItemType Directory -Path "deploy\ChronicleCore\spec" -Force | Out-Null
Copy-Item "spec\schema.sql" "deploy\ChronicleCore\spec\"

# Create launcher
Write-Host "Creating launcher..." -ForegroundColor Yellow
$launcher = @'
@echo off
echo ============================================================
echo   ChronicleCore
echo ============================================================
echo.

REM Check Python
python --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Python not installed!
    echo.
    echo Download from: https://www.python.org/downloads/
    echo Make sure to check "Add Python to PATH"
    echo.
    pause
    start https://www.python.org/downloads/
    exit /b 1
)

REM First run: install dependencies
if not exist "ml\.deps_installed" (
    echo First run detected - installing ML dependencies...
    echo This will take 1-2 minutes...
    echo.
    cd ml
    python -m pip install --quiet --upgrade pip
    pip install --quiet -r requirements.txt
    if %errorlevel% equ 0 (
        echo. > .deps_installed
        echo Dependencies installed!
        echo.
    ) else (
        echo ERROR: Failed to install dependencies
        pause
        exit /b 1
    )
    cd ..
)

REM Start backend
echo Starting ChronicleCore...
echo.
echo Backend: http://localhost:8080
echo Health Check: http://localhost:8080/health
echo.
echo Press Ctrl+C to stop
echo.

timeout /t 2 /nobreak >nul
start http://localhost:8080/health
chroniclecore.exe
'@
$launcher | Set-Content "deploy\ChronicleCore\START.bat"

# Create README
Write-Host "Creating README..." -ForegroundColor Yellow
$readme = @"
# ChronicleCore

## Quick Start

1. **Install Python** (if not already installed)
   - Download: https://www.python.org/downloads/
   - During install: CHECK "Add Python to PATH"

2. **Run ChronicleCore**
   - Double-click START.bat
   - First run installs dependencies (1-2 min)
   - Opens browser to http://localhost:8080/health

3. **Done!**
   - Leave the window open while using ChronicleCore
   - To stop: Close the window or press Ctrl+C

## What's Inside

- chroniclecore.exe - Main backend server
- ml/ - Machine learning sidecar
- spec/schema.sql - Database schema

## Troubleshooting

**"Python not installed"**
- Install Python 3.8+ from python.org
- Make sure "Add to PATH" is checked

**"Port already in use"**
- Another program is using port 8080
- Close it or change ChronicleCore's port

**"Dependencies failed"**
- Check internet connection
- Try: cd ml && pip install -r requirements.txt

## Data Location

Your data is stored in:
%LOCALAPPDATA%\ChronicleCore\chronicle.db

## Updates

When you receive an update:
1. Close ChronicleCore
2. Replace chroniclecore.exe with new version
3. Start again with START.bat
"@
$readme | Set-Content "deploy\ChronicleCore\README.md"

Write-Host ""
Write-Host "============================================================" -ForegroundColor Green
Write-Host "  PACKAGE READY!" -ForegroundColor Green
Write-Host "============================================================" -ForegroundColor Green
Write-Host ""
Write-Host "Location:" -ForegroundColor White
Write-Host "  deploy\ChronicleCore\" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor White
Write-Host "  1. ZIP the ChronicleCore folder" -ForegroundColor Yellow
Write-Host "  2. Send ZIP to your fiancé" -ForegroundColor Yellow
Write-Host "  3. She extracts and double-clicks START.bat" -ForegroundColor Yellow
Write-Host ""
Write-Host "What happens:" -ForegroundColor White
Write-Host "  - Checks for Python (prompts to install if missing)" -ForegroundColor Gray
Write-Host "  - First run: Installs ML dependencies (1-2 min)" -ForegroundColor Gray
Write-Host "  - Future runs: Starts immediately" -ForegroundColor Gray
Write-Host ""
