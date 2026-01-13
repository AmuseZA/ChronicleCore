@echo off
REM ChronicleCore Update Script
REM Updates the backend binary and dependencies

SETLOCAL EnableDelayedExpansion

echo ============================================================
echo    ChronicleCore Update
echo ============================================================
echo.

REM Check if ChronicleCore is running
tasklist /FI "IMAGENAME eq chroniclecore.exe" 2>NUL | find /I /N "chroniclecore.exe">NUL
if "%ERRORLEVEL%"=="0" (
    echo ChronicleCore is currently running.
    echo.
    set /p STOP_NOW="Stop it now to update? (Y/N): "

    if /i "!STOP_NOW!"=="Y" (
        echo Stopping ChronicleCore...
        taskkill /IM chroniclecore.exe /F >nul 2>&1
        timeout /t 2 >nul
        echo.
    ) else (
        echo Please stop ChronicleCore manually before updating
        pause
        exit /b 1
    )
)

echo [1/4] Backing up current version...
set BACKUP_DIR=%LOCALAPPDATA%\ChronicleCore\backups
if not exist "%BACKUP_DIR%" mkdir "%BACKUP_DIR%"

set TIMESTAMP=%date:~-4%%date:~3,2%%date:~0,2%_%time:~0,2%%time:~3,2%%time:~6,2%
set TIMESTAMP=%TIMESTAMP: =0%

if exist "apps\chroniclecore-core\chroniclecore.exe" (
    copy "apps\chroniclecore-core\chroniclecore.exe" "%BACKUP_DIR%\chroniclecore_%TIMESTAMP%.exe" >nul
    echo    Backup created: chroniclecore_%TIMESTAMP%.exe
) else (
    echo    No existing binary to backup
)
echo.

echo [2/4] Updating ML dependencies...
cd apps\chronicle-ml
pip install --upgrade -r requirements.txt >nul 2>&1
if %errorlevel% equ 0 (
    echo    Dependencies updated successfully!
) else (
    echo    WARNING: Some dependencies may not have updated
)
cd ..\..
echo.

echo [3/4] Checking new backend binary...
if exist "apps\chroniclecore-core\chroniclecore.exe" (
    echo    New binary found: apps\chroniclecore-core\chroniclecore.exe

    REM Get file size for verification
    for %%A in ("apps\chroniclecore-core\chroniclecore.exe") do set SIZE=%%~zA
    echo    Size: !SIZE! bytes
) else (
    echo    ERROR: No new binary found at apps\chroniclecore-core\chroniclecore.exe
    echo    Please ensure you've copied the new version
    pause
    exit /b 1
)
echo.

echo [4/4] Verifying database...
set DATA_DIR=%LOCALAPPDATA%\ChronicleCore
if exist "%DATA_DIR%\chronicle.db" (
    echo    Existing database found
    echo    Database location: %DATA_DIR%\chronicle.db
    echo.
    echo    NOTE: Database schema may need migration
    echo    See: spec\migrations\ for migration scripts
) else (
    echo    No existing database (fresh install)
)
echo.

echo ============================================================
echo    Update Complete!
echo ============================================================
echo.
echo  New version:  apps\chroniclecore-core\chroniclecore.exe
echo  Backup:       %BACKUP_DIR%\
echo  Database:     %DATA_DIR%\chronicle.db
echo.
echo ============================================================
echo    What's New in This Version
echo ============================================================
echo.
echo  - Multi-currency support (auto-detection)
echo  - ML auto-assignment improvements
echo  - System locale detection
echo  - Performance optimizations
echo.
echo  See CHANGELOG.md for full details
echo.
echo ============================================================
echo    Next Steps
echo ============================================================
echo.
echo  1. Start ChronicleCore (desktop shortcut or startup script)
echo  2. Verify it starts successfully
echo  3. Check http://localhost:8080/health
echo  4. If issues, restore backup:
echo     copy "%BACKUP_DIR%\chroniclecore_%TIMESTAMP%.exe" ^
echo          "apps\chroniclecore-core\chroniclecore.exe"
echo.
echo ============================================================
echo.

set /p START_NOW="Start ChronicleCore now? (Y/N): "

if /i "%START_NOW%"=="Y" (
    echo.
    echo Starting ChronicleCore...
    cd apps\chroniclecore-core
    start "ChronicleCore Backend" chroniclecore.exe
    timeout /t 2 >nul
    start http://localhost:8080/health
    echo.
    echo ChronicleCore is starting...
    echo Check the browser for health status
) else (
    echo.
    echo To start later, use the desktop shortcut
)

echo.
echo Press any key to exit.
pause >nul
