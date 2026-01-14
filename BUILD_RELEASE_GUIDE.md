# ChronicleCore Build & Release Guide

This document provides step-by-step instructions for building and releasing ChronicleCore. Follow these steps exactly to ensure consistent builds.

---

## Prerequisites

Before starting, ensure you have:
- Go installed (tested with Go 1.25+)
- Node.js and npm installed
- MinGW GCC for CGO (path: `C:\Users\josh\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin\gcc.exe`)
- Inno Setup 6 installed (path: `%LOCALAPPDATA%\Programs\Inno Setup 6\ISCC.exe`)
- GitHub CLI (`gh`) authenticated

---

## Step 1: Version Consistency Check

**CRITICAL**: Before building, verify ALL version numbers match across these files:

| File | Location | What to check |
|------|----------|---------------|
| `chroniclecore.iss` | Line 5 | `#define MyAppVersion "X.Y.Z"` |
| `apps/chroniclecore-ui/package.json` | Line 3 | `"version": "X.Y.Z"` |
| `apps/chroniclecore-core/cmd/server/main.go` | Line ~27 | `AppVersion = "X.Y.Z"` |
| `VERSION.txt` | Line 1 | `X.Y.Z` |
| `prepare_installer.ps1` | Line 5 | `[string]$Version = "X.Y.Z"` |
| `apps/chroniclecore-ui/src/routes/settings/+page.svelte` | Line 32 | `<div...>vX.Y.Z</div>` |

If any version is wrong, fix it before proceeding.

### Commands to check versions:
```bash
# Check all version files
grep -n "MyAppVersion" chroniclecore.iss
grep -n '"version"' apps/chroniclecore-ui/package.json | head -1
grep -n 'AppVersion' apps/chroniclecore-core/cmd/server/main.go | head -1
cat VERSION.txt
grep -n 'Version = ' prepare_installer.ps1 | head -1
grep -n 'v1\.' apps/chroniclecore-ui/src/routes/settings/+page.svelte
```

---

## Step 2: Build the Go Backend

The backend requires CGO for SQLite support.

```bash
cd apps/chroniclecore-core

# Set environment variables for CGO
set CGO_ENABLED=1
set CC=C:\Users\josh\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin\gcc.exe

# Build
go build -o chroniclecore.exe ./cmd/server
```

**Verify**: The file `chroniclecore.exe` should be ~21MB.

```bash
ls -la chroniclecore.exe
```

---

## Step 3: Build the UI

```bash
cd apps/chroniclecore-ui

# Install dependencies if needed
npm install

# Build for production
npm run build
```

**Verify**: The `build/` directory should contain `index.html` and `_app/` folder.

```bash
ls build/
```

---

## Step 4: Prepare Installer Package

This script downloads Python, installs ML dependencies, and assembles all files.

```bash
cd <project_root>

# Run the preparation script with the version
powershell -ExecutionPolicy Bypass -File prepare_installer.ps1 -Version "X.Y.Z"
```

**This takes 2-3 minutes** (downloads Python + pip packages).

**Verify** the `dist_installer/` directory contains:
- `chroniclecore.exe` (backend)
- `ChronicleCore.bat` (launcher)
- `debug_ml.bat` (ML debug script)
- `python/` (embedded Python + packages)
- `ml/` (ML sidecar code)
- `web/` (UI build output)
- `spec/` (database schema)
- `*.dll` files (MinGW runtime)

```bash
ls dist_installer/
```

### ⚠️ Known Issue: Batch Files

The `prepare_installer.ps1` script may create empty batch files due to PowerShell here-string issues. If `ChronicleCore.bat` or `debug_ml.bat` are empty (0 bytes), create them manually:

**ChronicleCore.bat**:
```batch
@echo off
SETLOCAL EnableExtensions EnableDelayedExpansion
echo [Launcher] Starting ChronicleCore...
set "APP_DIR=%~dp0"
cd /d "%APP_DIR%"
if not exist "%LOCALAPPDATA%\ChronicleCore" (
    mkdir "%LOCALAPPDATA%\ChronicleCore"
)
set "CHRONICLE_DB_PATH=%LOCALAPPDATA%\ChronicleCore\chronicle.db"
taskkill /F /IM chroniclecore.exe >nul 2>&1
echo [Launcher] Starting backend server...
start "" /B "chroniclecore.exe"
echo [Launcher] Waiting for server...
timeout /t 2 /nobreak >nul
echo [Launcher] Opening UI...
start "" "http://localhost:8080/?v=X.Y.Z"
echo [Launcher] running.
exit
```

**debug_ml.bat**:
```batch
@echo off
title ChronicleCore ML Debugger
echo Setting up environment...
set "APP_DIR=%~dp0"
cd /d "%APP_DIR%"
set "PYTHONUNBUFFERED=1"
set "ML_PORT=8081"
set "CC_ML_TOKEN=debug_token"
echo Checking Python...
if not exist "python\python.exe" (
    echo ERROR: python\python.exe not found!
    pause
    exit /b 1
)
"python\python.exe" --version
echo Starting Uvicorn...
cd ml
"..\python\python.exe" -m uvicorn src.main:app --host 127.0.0.1 --port 8081 --log-level debug
pause
```

---

## Step 5: Compile Inno Setup Installer

```bash
"$LOCALAPPDATA/Programs/Inno Setup 6/ISCC.exe" chroniclecore.iss
```

Or on Windows:
```cmd
"%LOCALAPPDATA%\Programs\Inno Setup 6\ISCC.exe" chroniclecore.iss
```

**This takes 1-2 minutes** (compresses all files).

**Verify**: Check `installer_output/ChronicleCore_Setup_vX.Y.Z.exe` exists (~90MB).

```bash
ls -la installer_output/ChronicleCore_Setup_v*.exe
```

---

## Step 6: Commit Changes (If Any)

If you made version fixes:

```bash
git add VERSION.txt prepare_installer.ps1 chroniclecore.iss apps/chroniclecore-ui/package.json apps/chroniclecore-core/cmd/server/main.go

git commit -m "$(cat <<'EOF'
Bump version to vX.Y.Z

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"

git push origin main
```

---

## Step 7: Create GitHub Release

### Option A: New Release
```bash
gh release create vX.Y.Z installer_output/ChronicleCore_Setup_vX.Y.Z.exe \
  --title "ChronicleCore vX.Y.Z" \
  --notes "$(cat <<'EOF'
## What's New
- [List changes here]

## Installation
Download **ChronicleCore_Setup_vX.Y.Z.exe** and run the installer.
EOF
)"
```

### Option B: Update Existing Release
```bash
# Delete old asset
gh release delete-asset vX.Y.Z ChronicleCore_Setup_vX.Y.Z.exe --yes

# Upload new asset
gh release upload vX.Y.Z installer_output/ChronicleCore_Setup_vX.Y.Z.exe --clobber

# Update release notes (optional)
gh release edit vX.Y.Z --notes "..."
```

---

## Verification Checklist

After release, verify:

- [ ] `gh release view vX.Y.Z` shows the asset
- [ ] Download the installer from GitHub
- [ ] Install on a clean Windows machine
- [ ] Launch from Start Menu
- [ ] Web UI opens at `http://localhost:8080`
- [ ] Settings page shows correct version

---

## Troubleshooting

### Build fails with CGO error
Ensure MinGW GCC is installed and path is correct:
```bash
"C:\Users\josh\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin\gcc.exe" --version
```

### Inno Setup not found
Install Inno Setup 6 or check the path:
```bash
ls "$LOCALAPPDATA/Programs/Inno Setup 6/"
```

### Missing files in dist_installer
Re-run `prepare_installer.ps1` from scratch. It cleans the directory first.

### Python packages fail to install
Check internet connection. The script downloads from PyPI.

### Empty batch files
See Step 4 note about manual batch file creation.

---

## Version Bump Process

When creating a new version:

1. Update ALL files listed in Step 1
2. Update `RELEASE_NOTES_X.Y.Z.md` if applicable
3. Run full build process (Steps 2-7)
4. Tag the commit: `git tag vX.Y.Z && git push --tags`

---

*Last updated: January 2026*
