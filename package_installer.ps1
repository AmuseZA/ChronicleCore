<#
.SYNOPSIS
    Builds and packages ChronicleCore for Windows distribution.

.DESCRIPTION
    This script automates the build process:
    1. Builds the SvelteKit frontend (SPA mode).
    2. Builds the Go backend with embedded static serving.
    3. Assembles a 'dist' directory with all required assets.
    4. Bundles a self-contained Python 3.11 environment.
    5. Caches the Python bundle to speed up future builds.
    6. Creates a final portable ZIP archive.

.EXAMPLE
    .\package_installer.ps1 -Quick
    Running with -Quick skips the final ZIP creation and assumes dependencies are fresh.
#>

param(
    [switch]$Quick
)

$ErrorActionPreference = "Stop"

# Configuration
$ProjectRoot = $PSScriptRoot
$FrontendDir = Join-Path $ProjectRoot "apps\chroniclecore-ui"
$BackendDir = Join-Path $ProjectRoot "apps\chroniclecore-core"
$MLDir = Join-Path $ProjectRoot "apps\chronicle-ml"
$DistDir = Join-Path $ProjectRoot "dist_release"
$PythonCache = Join-Path $ProjectRoot ".python_cache"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "   ChronicleCore Packaging Builder" -ForegroundColor Cyan
if ($Quick) {
    Write-Host "   (QUICK MODE: Skipping Zip)" -ForegroundColor Magenta
}
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# 0. Kill Stale Processes to unlock files
Write-Host "[0/6] Checking for stale processes..." -ForegroundColor Yellow
Get-Process "chroniclecore", "python" -ErrorAction SilentlyContinue | Foreach-Object {
    Write-Warning "Killing running process: $($_.Name) ($($_.Id))"
    Stop-Process -Id $_.Id -Force
    Start-Sleep -Seconds 1
}

# 1. Cleanup & Cache
Write-Host "[1/6] Preparing workspace..." -ForegroundColor Yellow

# Preserve Python environment if it exists (saves 5+ mins)
if (Test-Path "$DistDir\python") {
    if (-not (Test-Path $PythonCache)) {
        Write-Host "   Caching existing Python environment..."
        Copy-Item "$DistDir\python" $PythonCache -Recurse
    }
}

if (-not $Quick) {
    # Full Clean
    if (Test-Path $DistDir) {
        Remove-Item -Path $DistDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $DistDir | Out-Null
    Write-Host "   Cleaned 'dist' directory." -ForegroundColor Green
}
else {
    # Quick Clean - Create if missing, but don't wipe
    if (-not (Test-Path $DistDir)) {
        New-Item -ItemType Directory -Path $DistDir | Out-Null
    }
    Write-Host "   Skipped full clean (Quick Mode)." -ForegroundColor Gray
}

# 2. Build Frontend
Write-Host "[2/6] Building Frontend (SvelteKit)..." -ForegroundColor Yellow
Push-Location $FrontendDir
try {
    # Install dependencies if needed
    if (-not (Test-Path "node_modules")) {
        Write-Host "   Installing npm dependencies..."
        cmd /c "npm install"
    }
    
    # Build
    Write-Host "   Running build..."
    cmd /c "npm run build"
    
    if (-not (Test-Path "build")) {
        Throw "Frontend build failed: 'build' directory not created."
    }
}
finally {
    Pop-Location
}
Write-Host "   Frontend built successfully." -ForegroundColor Green

# 3. Build Backend
Write-Host "[3/6] Building Backend (Go)..." -ForegroundColor Yellow
Push-Location $BackendDir
try {
    $GoExe = "chroniclecore.exe"
    $Env:CGO_ENABLED = "1"
    
    # Configure Paths (Fallback to known locations if not in PATH)
    $GCCPath = Get-Command gcc -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source
    if (-not $GCCPath) {
        $GCCPath = "C:\Users\josh\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin\gcc.exe"
        if (Test-Path $GCCPath) {
            Write-Host "   Using GCC: $GCCPath"
            $Env:CC = $GCCPath
            $GCCDir = Split-Path $GCCPath -Parent
            $Env:PATH = "$GCCDir;$Env:PATH"
        }
        else {
            Write-Warning "GCC not found. CGO build might fail."
        }
    }

    $GoCmd = "go"
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        $KnownGoPath = "C:\Program Files\Go\bin\go.exe"
        if (Test-Path $KnownGoPath) {
            Write-Host "   Using Go: $KnownGoPath"
            $GoCmd = "& `"$KnownGoPath`""
        }
        else {
            Throw "Go not found in PATH or standard location."
        }
    }

    # Execute Build
    Invoke-Expression "$GoCmd build -o $GoExe ./cmd/server"
    
    if (-not (Test-Path $GoExe)) {
        Throw "Backend build failed: $GoExe not created."
    }
    
    move-Item -Path $GoExe -Destination $DistDir -Force

    # Copy MinGW Runtime DLLs (Required for CGO/SQLite on clean systems)
    if ($Env:CC) {
        $GCCBin = Split-Path $Env:CC -Parent
        $Dlls = @("libwinpthread-1.dll", "libgcc_s_seh-1.dll", "libstdc++-6.dll")
        foreach ($Dll in $Dlls) {
            $SrcInfo = Join-Path $GCCBin $Dll
            if (Test-Path $SrcInfo) {
                Write-Host "   Bundling $Dll..."
                Copy-Item $SrcInfo $DistDir -Force
            }
            else {
                Write-Warning "Could not find $Dll in GCC bin. Binary might fail on clean machines."
            }
        }
    }
}
finally {
    Pop-Location
}
Write-Host "   Backend built successfully." -ForegroundColor Green

# 4. Assemble Distribution
Write-Host "[4/6] Assembling Distribution..." -ForegroundColor Yellow

# Copy Web Assets
$DistWeb = Join-Path $DistDir "web"
Copy-Item -Path (Join-Path $FrontendDir "build") -Destination $DistWeb -Recurse -Force
Write-Host "   Web assets copied."

# Copy ML Assets
$DistML = Join-Path $DistDir "ml"
New-Item -ItemType Directory -Path $DistML -Force | Out-Null
Copy-Item -Path (Join-Path $MLDir "src") -Destination $DistML -Recurse -Force
Copy-Item -Path (Join-Path $MLDir "requirements.txt") -Destination $DistML -Force
Write-Host "   ML assets copied."

# Copy Database Schema (Required for DB initialization)
$SchemaSrc = Join-Path $ProjectRoot "spec\schema.sql"
if (Test-Path $SchemaSrc) {
    Copy-Item -Path $SchemaSrc -Destination $DistDir -Force
    Write-Host "   Database schema copied."
}
else {
    Write-Warning "Schema file not found at $SchemaSrc"
}

# Create Standard Launcher (Hidden Console)
$RunnerPath = Join-Path $DistDir "ChronicleCore_Launcher.bat"
$RunnerContent = @"
@echo off
cd /d "%~dp0"
start "" "chroniclecore.exe"
timeout /t 5 >nul
start http://localhost:8080
"@
Set-Content -Path $RunnerPath -Value $RunnerContent

# Create Debug Launcher (Visible Console + Pause)
$DebugRunnerPath = Join-Path $DistDir "DEBUG_LAUNCHER.bat"
$DebugRunnerContent = @"
@echo off
cd /d "%~dp0"
echo Starting ChronicleCore in DEBUG mode...
echo Do not close this window.
echo.
"chroniclecore.exe"
pause
"@
Set-Content -Path $DebugRunnerPath -Value $DebugRunnerContent

Write-Host "   Launchers created." -ForegroundColor Green

# 5. Bundle Python
Write-Host "[5/6] Bundling Python Environment..." -ForegroundColor Yellow
$DistPython = Join-Path $DistDir "python"
$PythonUrl = "https://www.python.org/ftp/python/3.11.7/python-3.11.7-embed-amd64.zip"
$PythonZip = Join-Path $DistDir "python.zip"

# Restore from cache if available
if (Test-Path $PythonCache) {
    Write-Host "   Restoring Python from cache..."
    if (Test-Path $DistPython) { Remove-Item $DistPython -Recurse -Force }
    Move-Item $PythonCache $DistPython
}

if (-not (Test-Path $DistPython)) {
    # 1. Download Python
    Write-Host "   Downloading Python 3.11 Embeddable..."
    Invoke-WebRequest -Uri $PythonUrl -OutFile $PythonZip
    
    # 2. Extract
    Write-Host "   Extracting Python..."
    Expand-Archive -Path $PythonZip -DestinationPath $DistPython -Force
    Remove-Item $PythonZip

    # 3. Configure ._pth to allow site-packages (uncomment 'import site')
    Write-Host "   Configuring python311._pth..."
    $PthFile = Join-Path $DistPython "python311._pth"
    $PthContent = Get-Content $PthFile
    $PthContent = $PthContent -replace "#import site", "import site"
    Set-Content -Path $PthFile -Value $PthContent

    # 4. Install pip
    Write-Host "   Installing pip..."
    $GetPipUrl = "https://bootstrap.pypa.io/get-pip.py"
    $GetPipFile = Join-Path $DistPython "get-pip.py"
    Invoke-WebRequest -Uri $GetPipUrl -OutFile $GetPipFile
    
    $PythonExe = Join-Path $DistPython "python.exe"
    & $PythonExe $GetPipFile --no-warn-script-location
    Remove-Item $GetPipFile

    # 5. Install Requirements
    Write-Host "   Installing requirements (this may take a while)..."
    $ReqFile = Join-Path $MLDir "requirements.txt"
    & $PythonExe -m pip install -r $ReqFile --no-warn-script-location
    
    if ($LASTEXITCODE -ne 0) {
        Throw "Failed to install Python requirements. Check internet connection and PyPI access."
    }


    Write-Host "   Dependencies installed successfully." -ForegroundColor Green
}
else {
    Write-Host "   Python environment restored."
}

# 6. Verify Installation (Always run)
Write-Host "   Verifying ML dependencies..."
$PythonExe = Join-Path $DistPython "python.exe"
$VerifyCmd = "import fastapi; import uvicorn; import sklearn; import pandas; import numpy; import pydantic; import httpx; print('Verification successful')"
& $PythonExe -c $VerifyCmd

if ($LASTEXITCODE -ne 0) {
    Throw "Python dependency verification failed. Some packages are missing or broken."
}

Write-Host "   Dependencies verified successfully." -ForegroundColor Green

# 6. Create Zip Archive
if (-not $Quick) {
    Write-Host "[6/6] Creating Final Zip Archive..." -ForegroundColor Yellow
    $ReleaseZip = Join-Path $ProjectRoot "ChronicleCore_Portable.zip"
    if (Test-Path $ReleaseZip) { Remove-Item $ReleaseZip -Force }

    Write-Host "   Compressing '$DistDir' to '$ReleaseZip'..."
    Compress-Archive -Path "$DistDir\*" -DestinationPath $ReleaseZip -Force
}
else {
    Write-Host "[6/6] Skipping Zip Archive (Quick Mode)." -ForegroundColor Magenta
}

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "   PACKAGING COMPLETE" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "1. Folder:  $DistDir"
if (-not $Quick) {
    Write-Host "2. ARCHIVE: $ReleaseZip"
    Write-Host ""
    Write-Host "-> Ready to send! Just copy 'ChronicleCore_Portable.zip'." -ForegroundColor Green
}
else {
    Write-Host "2. ARCHIVE: (Skipped)"
    Write-Host ""
    Write-Host "-> Quick build done. Run from the dist folder." -ForegroundColor Green
}
Write-Host ""
