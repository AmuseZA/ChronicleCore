# Prepare everything for Inno Setup installer
# This downloads embedded Python and prepares all files

param(
    [string]$Version = "2.1.0"
)

$ErrorActionPreference = "Stop"

Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "  Preparing ChronicleCore Installer" -ForegroundColor Cyan
Write-Host "  Version: $Version" -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host ""

# Paths
$RootDir = $PSScriptRoot
$DistDir = Join-Path $RootDir "dist_installer"
$PythonDir = Join-Path $DistDir "python"
$BackendExe = Join-Path $RootDir "apps\chroniclecore-core\chroniclecore.exe"

# Clean previous build
if (Test-Path $DistDir) {
    Write-Host "Cleaning previous build..." -ForegroundColor Yellow
    Remove-Item $DistDir -Recurse -Force
}
New-Item -ItemType Directory -Path $DistDir | Out-Null
Write-Host ""

# Step 1: Check backend exists
Write-Host "[1/6] Checking backend binary..." -ForegroundColor Yellow
if (-not (Test-Path $BackendExe)) {
    Write-Host "   ERROR: Backend not found at $BackendExe" -ForegroundColor Red
    Write-Host "   Build it first: cd apps\chroniclecore-core && .\build.bat" -ForegroundColor Yellow
    exit 1
}
$backendSize = (Get-Item $BackendExe).Length / 1MB
Write-Host "   Found: $([math]::Round($backendSize, 2)) MB" -ForegroundColor Green
Write-Host ""

# Step 2: Copy backend
Write-Host "[2/6] Copying backend..." -ForegroundColor Yellow
Copy-Item $BackendExe (Join-Path $DistDir "chroniclecore.exe")
Write-Host "   Copied!" -ForegroundColor Green
Write-Host ""

# Step 2a: Copy Frontend Assets
Write-Host "[2a/6] Copying Frontend Assets..." -ForegroundColor Yellow
$FrontendBuildDir = Join-Path $RootDir "apps\chroniclecore-ui\build"
if (-not (Test-Path $FrontendBuildDir)) {
    Write-Host "   ERROR: Frontend build not found at $FrontendBuildDir" -ForegroundColor Red
    Write-Host "   Build it first: cd apps\chroniclecore-ui && npm run build" -ForegroundColor Yellow
    exit 1
}
Copy-Item $FrontendBuildDir (Join-Path $DistDir "web") -Recurse -Force
Write-Host "   Copied web assets!" -ForegroundColor Green
Write-Host ""

# Step 2b: Copy Runtime DLLs (MinGW/SQLite)
Write-Host "[2b/6] Copying Runtime DLLs..." -ForegroundColor Yellow
$GCCPath = "C:\Users\josh\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin"
$Dlls = @("libwinpthread-1.dll", "libgcc_s_seh-1.dll", "libstdc++-6.dll")
foreach ($Dll in $Dlls) {
    $SrcDll = Join-Path $GCCPath $Dll
    if (Test-Path $SrcDll) {
        Copy-Item $SrcDll $DistDir -Force
        Write-Host "   Bundled $Dll"
    }
    else {
        Write-Warning "   Missing $Dll - Application may fail on clean systems!"
    }
}
Write-Host ""

# Step 3: Download embedded Python
Write-Host "[3/6] Downloading embedded Python 3.11.9..." -ForegroundColor Yellow
$pythonVersion = "3.11.9"
$pythonUrl = "https://www.python.org/ftp/python/$pythonVersion/python-$pythonVersion-embed-amd64.zip"
$pythonZip = Join-Path $DistDir "python_embed.zip"

Write-Host "   Downloading from python.org..." -ForegroundColor Gray
try {
    Invoke-WebRequest -Uri $pythonUrl -OutFile $pythonZip -UseBasicParsing
    Write-Host "   Downloaded!" -ForegroundColor Green
}
catch {
    Write-Host "   ERROR: Failed to download Python: $_" -ForegroundColor Red
    exit 1
}

Write-Host "   Extracting..." -ForegroundColor Gray
Expand-Archive -Path $pythonZip -DestinationPath $PythonDir -Force
Remove-Item $pythonZip

# Enable site-packages
$pthFile = Get-ChildItem -Path $PythonDir -Filter "python*._pth" | Select-Object -First 1
if ($pthFile) {
    $content = Get-Content $pthFile.FullName
    $content = $content -replace "#import site", "import site"
    $content | Set-Content $pthFile.FullName
    Write-Host "   Enabled site-packages" -ForegroundColor Green
}
Write-Host ""

# Step 4: Install pip
Write-Host "[4/6] Installing pip..." -ForegroundColor Yellow
$getPipUrl = "https://bootstrap.pypa.io/get-pip.py"
$getPipPath = Join-Path $PythonDir "get-pip.py"
Invoke-WebRequest -Uri $getPipUrl -OutFile $getPipPath -UseBasicParsing
$pythonExe = Join-Path $PythonDir "python.exe"
& $pythonExe $getPipPath --no-warn-script-location 2>&1 | Out-Null
Write-Host "   Pip installed!" -ForegroundColor Green
Write-Host ""

# Step 5: Install ML dependencies
Write-Host "[5/6] Installing ML dependencies..." -ForegroundColor Yellow
$requirementsFile = Join-Path $RootDir "apps\chronicle-ml\requirements.txt"
Write-Host "   Installing packages (this may take 2-3 minutes)..." -ForegroundColor Gray
& $pythonExe -m pip install -r $requirementsFile
if ($LASTEXITCODE -eq 0) {
    Write-Host "   Dependencies installed!" -ForegroundColor Green
}
else {
    Write-Host "   ERROR: Failed to install dependencies" -ForegroundColor Red
    exit 1
}

# Verify
Write-Host "   Verifying installations..." -ForegroundColor Gray
$verifyScript = @"
try:
    import fastapi
    import uvicorn
    import sklearn
    import pandas
    import numpy
    import pydantic
    import httpx
    # Deep import check to verify DLLs
    from sklearn.linear_model import LogisticRegression
    from sklearn.feature_extraction.text import TfidfVectorizer
    print('OK')
except ImportError as e:
    print(f'ERROR: {e}')
    exit(1)
except Exception as e:
    print(f'ERROR: {e}')
    exit(1)
"@
$verifyPath = Join-Path $DistDir "verify.py"
$verifyScript | Set-Content $verifyPath
$result = & $pythonExe $verifyPath
if ($result -eq "OK") {
    Write-Host "   All packages verified!" -ForegroundColor Green
}
else {
    Write-Host "   ERROR: Verification failed: $result" -ForegroundColor Red
    exit 1
}
Remove-Item $verifyPath
Write-Host ""

# Step 6: Copy ML code and other files
Write-Host "[6/6] Copying additional files..." -ForegroundColor Yellow
Copy-Item (Join-Path $RootDir "apps\chronicle-ml") (Join-Path $DistDir "ml") -Recurse
New-Item -ItemType Directory -Path (Join-Path $DistDir "spec") | Out-Null
Copy-Item (Join-Path $RootDir "spec\schema.sql") (Join-Path $DistDir "spec\")

# Copy docs
Copy-Item (Join-Path $RootDir "QUICK_START.md") $DistDir -ErrorAction SilentlyContinue
Copy-Item (Join-Path $RootDir "README_USER.md") $DistDir -ErrorAction SilentlyContinue
Write-Host "   Copied!" -ForegroundColor Green
Write-Host ""

# Create launcher script
Write-Host "Creating launcher script..." -ForegroundColor Yellow
$launcherScript = @'
@echo off
REM ChronicleCore Launcher
title ChronicleCore

echo ============================================================
echo   ChronicleCore
echo ============================================================
echo.
echo Starting services...
echo.

REM Get the directory where this script is located
set "APP_DIR=%~dp0"
cd /d "%APP_DIR%"

REM Check Python
if not exist "python\python.exe" (
    echo ERROR: Python runtime not found!
    echo Please reinstall ChronicleCore.
    pause
    exit /b 1
)

REM Check backend
if not exist "chroniclecore.exe" (
    echo ERROR: Backend binary not found!
    echo Please reinstall ChronicleCore.
    pause
    exit /b 1
)

REM Verify ML dependencies (quick check)
python\python.exe -c "import fastapi; import sklearn" >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: ML dependencies missing or corrupted!
    echo Please reinstall ChronicleCore.
    pause
    exit /b 1
)

echo [1/3] Backend: Starting...
echo [2/3] ML Sidecar: Starting...
echo [3/3] Web UI: Opening browser...
echo.

REM Wait a moment for services to start
timeout /t 3 /nobreak >nul

REM Open browser to web UI (Force cache bust)
start http://localhost:8080/?v=$Version

REM Start backend (this blocks and shows logs)
echo.
echo ============================================================
echo   ChronicleCore is running!
echo ============================================================
echo.
echo   Web UI: http://localhost:8080
echo   API: http://localhost:8080/health
echo.
echo   Keep this window open while using ChronicleCore
echo   Press Ctrl+C to stop
echo.
echo ============================================================
echo.

chroniclecore.exe
"@
$launcherScript | Set-Content (Join-Path $DistDir "ChronicleCore.bat")
Write-Host "   Launcher created!" -ForegroundColor Green
Write-Host ""

# Create Debug ML script
Write-Host "Creating debug_ml script..." -ForegroundColor Yellow
$debugMlScript = @'
@echo off
title ChronicleCore ML Debugger
echo ============================================================
echo   ChronicleCore ML Sidecar Debugger
echo ============================================================
echo.
echo Setting up environment...
set "APP_DIR=%~dp0"
cd /d "%APP_DIR%"

set "PYTHONUNBUFFERED=1"
set "ML_PORT=8081"
set "CC_ML_TOKEN=debug_token"

echo.
echo Checking Python...
if not exist "python\python.exe" (
    echo ERROR: python\python.exe not found!
    pause
    exit /b 1
)
"python\python.exe" --version

echo.
echo Checking ML directory...
if not exist "ml\src\main.py" (
    echo ERROR: ml\src\main.py not found!
    pause
    exit /b 1
)

echo.
echo Starting Uvicorn...
echo Command: python\python.exe -m uvicorn src.main:app --host 127.0.0.1 --port 8081 --log-level debug --app-dir ml
echo.
echo ------------------------------------------------------------
cd ml
"..\python\python.exe" -m uvicorn src.main:app --host 127.0.0.1 --port 8081 --log-level debug
echo ------------------------------------------------------------
echo.
echo Sidecar exited with code %errorlevel%
pause
'@
$debugMlScript | Set-Content (Join-Path $DistDir "debug_ml.bat")
Write-Host "   Debug script created!" -ForegroundColor Green
Write-Host ""

# Update VERSION.txt
$Version | Set-Content (Join-Path $RootDir "VERSION.txt")

# Success
Write-Host "============================================================" -ForegroundColor Green
Write-Host "  PREPARATION COMPLETE!" -ForegroundColor Green
Write-Host "============================================================" -ForegroundColor Green
Write-Host ""
Write-Host "Files prepared in: $DistDir" -ForegroundColor Cyan
Write-Host ""
Write-Host "Contents:" -ForegroundColor White
Write-Host "  - chroniclecore.exe ($([math]::Round($backendSize, 2)) MB)" -ForegroundColor Gray
Write-Host "  - python/ (embedded runtime, ~15 MB)" -ForegroundColor Gray
Write-Host "  - ml/ (sidecar code)" -ForegroundColor Gray
Write-Host "  - ChronicleCore.bat (launcher)" -ForegroundColor Gray
Write-Host ""
Write-Host "Next step:" -ForegroundColor White
Write-Host "  Run: .\build_inno_installer.ps1 -Version $Version" -ForegroundColor Yellow
Write-Host ""
