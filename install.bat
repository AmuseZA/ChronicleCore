@echo off
REM ChronicleCore One-Click Installer
REM Version 1.0.0

SETLOCAL EnableDelayedExpansion

echo ============================================================
echo    ChronicleCore Installer
echo    Version 1.0.0
echo ============================================================
echo.

REM Check for admin rights (optional, but helpful)
NET SESSION >nul 2>&1
if %errorLevel% neq 0 (
    echo Note: Not running as administrator
    echo This is fine for local installation
    echo.
)

REM Step 1: Check Python
echo [1/5] Checking Python installation...
python --version >nul 2>&1
if %errorlevel% neq 0 (
    echo.
    echo Python is not installed.
    echo.
    echo Opening Python download page...
    echo Please install Python 3.8 or higher
    echo IMPORTANT: Check "Add Python to PATH" during installation
    echo.
    start https://www.python.org/downloads/
    echo.
    echo After installing Python, run this installer again.
    pause
    exit /b 1
)

echo    Python found!
python --version
echo.

REM Step 2: Install ML dependencies
echo [2/5] Installing ML dependencies...
cd apps\chronicle-ml
if not exist "requirements.txt" (
    echo ERROR: requirements.txt not found
    echo Please run this installer from the ChronicleCore root directory
    pause
    exit /b 1
)

python -m pip install --upgrade pip >nul 2>&1
pip install -r requirements.txt

if %errorlevel% neq 0 (
    echo.
    echo ERROR: Failed to install dependencies
    pause
    exit /b 1
)

REM Verify imports
python -c "import fastapi; import uvicorn; import sklearn; import pandas; import numpy; import pydantic; import httpx; print('Verification successful')"
if %errorlevel% neq 0 (
    echo.
    echo ERROR: Dependency verification failed!
    echo Some packages are missing or broken.
    pause
    exit /b 1
)

echo    Dependencies installed and verified successfully!

cd ..\..
echo.

REM Step 3: Check backend binary
echo [3/5] Checking backend binary...
if not exist "apps\chroniclecore-core\chroniclecore.exe" (
    echo ERROR: Backend binary not found at apps\chroniclecore-core\chroniclecore.exe
    echo Please ensure the binary is built and in the correct location
    pause
    exit /b 1
)

echo    Backend binary found!
echo.

REM Step 4: Create shortcuts
echo [4/5] Creating shortcuts...

REM Create data directory
set DATA_DIR=%LOCALAPPDATA%\ChronicleCore
if not exist "%DATA_DIR%" (
    mkdir "%DATA_DIR%"
    echo    Created data directory: %DATA_DIR%
)

REM Create desktop shortcut using PowerShell
set SCRIPT_DIR=%~dp0
set BACKEND_PATH=%SCRIPT_DIR%apps\chroniclecore-core\chroniclecore.exe

echo Set oWS = WScript.CreateObject("WScript.Shell") > CreateShortcut.vbs
echo sLinkFile = "%USERPROFILE%\Desktop\ChronicleCore.lnk" >> CreateShortcut.vbs
echo Set oLink = oWS.CreateShortcut(sLinkFile) >> CreateShortcut.vbs
echo oLink.TargetPath = "%BACKEND_PATH%" >> CreateShortcut.vbs
echo oLink.WorkingDirectory = "%SCRIPT_DIR%apps\chroniclecore-core" >> CreateShortcut.vbs
echo oLink.Description = "ChronicleCore Time Tracker" >> CreateShortcut.vbs
echo oLink.Save >> CreateShortcut.vbs

cscript CreateShortcut.vbs >nul
del CreateShortcut.vbs

if exist "%USERPROFILE%\Desktop\ChronicleCore.lnk" (
    echo    Desktop shortcut created!
) else (
    echo    WARNING: Could not create desktop shortcut
)

REM Create startup script
set STARTUP_SCRIPT=%DATA_DIR%\start_chroniclecore.bat
echo @echo off > "%STARTUP_SCRIPT%"
echo cd /d "%SCRIPT_DIR%apps\chroniclecore-core" >> "%STARTUP_SCRIPT%"
echo start "ChronicleCore Backend" chroniclecore.exe >> "%STARTUP_SCRIPT%"
echo echo ChronicleCore is starting... >> "%STARTUP_SCRIPT%"
echo timeout /t 2 ^>nul >> "%STARTUP_SCRIPT%"
echo start http://localhost:8080/health >> "%STARTUP_SCRIPT%"

echo    Startup script created!
echo.

REM Step 5: Installation complete
echo [5/5] Installation complete!
echo.
echo ============================================================
echo    Installation Summary
echo ============================================================
echo.
echo  Status: READY
echo  Backend: %BACKEND_PATH%
echo  Data:    %DATA_DIR%
echo.
echo ============================================================
echo    What's Next?
echo ============================================================
echo.
echo  1. Double-click the desktop shortcut "ChronicleCore"
echo     OR
echo     Run: %STARTUP_SCRIPT%
echo.
echo  2. Backend will start on http://localhost:8080
echo.
echo  3. Check it's running by visiting:
echo     http://localhost:8080/health
echo.
echo  4. You should see: {"status":"ok"}
echo.
echo ============================================================
echo    First-Time Setup (in browser/UI)
echo ============================================================
echo.
echo  1. Create your first client (e.g., "Acme Corp")
echo  2. Create a service (e.g., "Bookkeeping")
echo  3. Set your hourly rate (e.g., R 150)
echo  4. The system will start tracking automatically
echo.
echo ============================================================
echo    ML Auto-Assignment
echo ============================================================
echo.
echo  - For the first 1-2 weeks: Manually assign blocks
echo  - After 50+ assignments: ML model trains automatically
echo  - Week 3+: Auto-suggestions with 80-90%% accuracy
echo.
echo ============================================================
echo    Documentation
echo ============================================================
echo.
echo  Quick Start:  QUICK_START.md
echo  ML Guide:     workflow\ml_user_guide.md
echo  Deployment:   DEPLOYMENT_CHECKLIST.md
echo.
echo ============================================================
echo.
echo Would you like to start ChronicleCore now?
echo.
set /p START_NOW="Start now? (Y/N): "

if /i "%START_NOW%"=="Y" (
    echo.
    echo Starting ChronicleCore...
    call "%STARTUP_SCRIPT%"
    echo.
    echo Backend is starting in a new window...
    echo Check http://localhost:8080/health to verify
    echo.
) else (
    echo.
    echo To start later, use the desktop shortcut or run:
    echo %STARTUP_SCRIPT%
)

echo.
echo Installation complete! Press any key to exit.
pause >nul
