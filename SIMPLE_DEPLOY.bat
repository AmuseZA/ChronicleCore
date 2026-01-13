@echo off
REM ChronicleCore Simple Deployment Package Creator
REM This creates a portable ZIP with everything needed

echo ============================================================
echo   ChronicleCore Deployment Package Creator
echo ============================================================
echo.

REM Check if backend exists
if not exist "apps\chroniclecore-core\chroniclecore.exe" (
    echo ERROR: Backend not built!
    echo.
    echo Please build first:
    echo   cd apps\chroniclecore-core
    echo   build.bat
    echo.
    pause
    exit /b 1
)

REM Create deploy directory
if exist "deploy" rmdir /s /q deploy
mkdir deploy
mkdir deploy\ChronicleCore

echo [1/4] Copying backend...
copy "apps\chroniclecore-core\chroniclecore.exe" "deploy\ChronicleCore\"

echo [2/4] Copying ML code...
xcopy "apps\chronicle-ml" "deploy\ChronicleCore\ml\" /E /I /Q

echo [3/4] Copying database schema...
mkdir "deploy\ChronicleCore\spec"
copy "spec\schema.sql" "deploy\ChronicleCore\spec\"

echo [4/4] Creating launcher script...

REM Create a simple launcher
echo @echo off > deploy\ChronicleCore\START_CHRONICLECORE.bat
echo echo Starting ChronicleCore... >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo echo. >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo REM Check Python >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo python --version ^>nul 2^>^&1 >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo if %%errorlevel%% neq 0 ( >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo     echo Python is not installed! >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo     echo. >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo     echo Please install Python from: https://www.python.org/downloads/ >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo     echo Make sure to check "Add Python to PATH" during installation >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo     echo. >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo     pause >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo     exit /b 1 >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo ) >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo. >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo REM Install ML dependencies if needed >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo if not exist "ml\.installed" ( >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo     echo Installing ML dependencies... >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo     cd ml >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo     python -m pip install --upgrade pip ^>nul 2^>^&1 >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo     pip install -r requirements.txt >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo     if %%errorlevel%% equ 0 ( >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo         echo. ^> .installed >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo         echo Dependencies installed! >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo     ) else ( >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo         echo ERROR: Failed to install dependencies >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo         pause >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo         exit /b 1 >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo     ) >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo     cd .. >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo ) >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo. >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo REM Start backend >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo echo. >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo echo ChronicleCore is starting... >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo echo Backend: http://localhost:8080 >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo echo Health check: http://localhost:8080/health >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo echo. >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo echo Press Ctrl+C to stop >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo echo. >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo start http://localhost:8080/health >> deploy\ChronicleCore\START_CHRONICLECORE.bat
echo chroniclecore.exe >> deploy\ChronicleCore\START_CHRONICLECORE.bat

REM Create simple README
echo ChronicleCore > deploy\ChronicleCore\README.txt
echo. >> deploy\ChronicleCore\README.txt
echo INSTALLATION: >> deploy\ChronicleCore\README.txt
echo. >> deploy\ChronicleCore\README.txt
echo 1. Install Python 3.8+ from https://www.python.org/downloads/ >> deploy\ChronicleCore\README.txt
echo    - Check "Add Python to PATH" during installation >> deploy\ChronicleCore\README.txt
echo. >> deploy\ChronicleCore\README.txt
echo 2. Double-click START_CHRONICLECORE.bat >> deploy\ChronicleCore\README.txt
echo. >> deploy\ChronicleCore\README.txt
echo 3. First run will install dependencies automatically (1-2 minutes) >> deploy\ChronicleCore\README.txt
echo. >> deploy\ChronicleCore\README.txt
echo 4. Backend will start at http://localhost:8080 >> deploy\ChronicleCore\README.txt
echo. >> deploy\ChronicleCore\README.txt
echo That's it! >> deploy\ChronicleCore\README.txt

echo.
echo ============================================================
echo   Package Ready!
echo ============================================================
echo.
echo Location: deploy\ChronicleCore\
echo.
echo What to send her:
echo   1. ZIP the deploy\ChronicleCore folder
echo   2. Send her the ZIP
echo   3. She extracts and double-clicks START_CHRONICLECORE.bat
echo.
echo First run:
echo   - Checks for Python (opens download if missing)
echo   - Installs ML dependencies automatically
echo   - Starts backend
echo.
echo Future runs:
echo   - Just double-click START_CHRONICLECORE.bat
echo   - Dependencies already installed, starts immediately
echo.
pause
