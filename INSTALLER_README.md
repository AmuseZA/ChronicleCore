# ChronicleCore Installer - Super Simple

## What You Get

A **single EXE installer** that does everything automatically:
- Embeds Python runtime
- Includes all ML dependencies
- Creates desktop shortcut
- Auto-opens browser to web UI

## How to Build (One Command)

```powershell
.\BUILD_INSTALLER.ps1 -Version "1.0.0"
```

**Time:** 5-10 minutes (mostly downloading Python)

**Output:** `installer_output\ChronicleCore_Setup_v1.0.0.exe` (~70 MB)

## What Happens

### Step 1: Preparation (5 min)
- Downloads Python 3.11.9 embedded (~15 MB)
- Installs ML packages (FastAPI, scikit-learn, etc.)
- Copies backend and ML code
- Creates smart launcher script

### Step 2: Building Installer (1 min)
- Packages everything into single EXE
- Uses Inno Setup (industry standard)
- Compresses to ~70 MB

## User Experience

### For Your Fiancé

1. **Download** `ChronicleCore_Setup_v1.0.0.exe`
2. **Double-click** installer
3. **Follow wizard** (30 seconds, just click Next)
4. **Done!** Desktop shortcut appears

### First Run

1. **Double-click** desktop shortcut "ChronicleCore"
2. **Launcher checks:** Python ✓ Backend ✓ ML deps ✓
3. **Browser opens** to http://localhost:8080
4. **App ready!** Console shows status

### Every Run After

1. Double-click shortcut
2. Browser opens
3. Done!

## What's Inside

- **chroniclecore.exe** (11 MB) - Your Go backend
- **python/** (15 MB) - Embedded Python 3.11.9
- **python packages** (40 MB) - FastAPI, scikit-learn, pandas, etc.
- **ml/** - ML sidecar source code
- **ChronicleCore.bat** - Smart launcher script
- **Desktop shortcut** - One-click startup

## The Launcher Script

When user clicks desktop shortcut, it:

1. ✓ Checks Python exists
2. ✓ Checks backend exists
3. ✓ Verifies ML dependencies
4. ✓ Starts backend server
5. ✓ Opens browser to http://localhost:8080
6. ✓ Shows clear status messages

**If anything is wrong:** Shows clear error message

## Prerequisites

### You Need (One-Time)

1. **Inno Setup** (free)
   - Download: https://jrsoftware.org/isdl.php
   - Install takes 2 minutes

2. **Backend built**
   ```powershell
   cd apps\chroniclecore-core
   .\build.bat
   ```

### She Needs

**Nothing!** Everything is embedded in the installer.

## Full Workflow

### First Deployment

```powershell
# 1. Build backend (if not already)
cd apps\chroniclecore-core
.\build.bat

# 2. Build installer
cd ..\..
.\BUILD_INSTALLER.ps1 -Version "1.0.0"

# 3. Upload and send
# Upload: installer_output\ChronicleCore_Setup_v1.0.0.exe
# Send link to your fiancé
```

### Future Updates

**Option A: Full Installer (New features)**
```powershell
.\BUILD_INSTALLER.ps1 -Version "1.0.1"
```
Send new installer (~70 MB)

**Option B: Binary Only (Bug fixes)**
```powershell
cd apps\chroniclecore-core
.\build.bat
```
Send just `chroniclecore.exe` (~11 MB)
She replaces in installation folder

## Troubleshooting

### "Inno Setup not found"

Install from: https://jrsoftware.org/isdl.php

### "Backend not found"

Build it first:
```powershell
cd apps\chroniclecore-core
.\build.bat
```

### "Python download failed"

Check internet connection. Script downloads from python.org

### "Installer too large"

Normal! Includes:
- Backend: 11 MB
- Python: 15 MB
- ML packages: 40 MB
- Total: ~70 MB

Updates are only 11 MB (just the backend)

## Testing

Before sending to her:

1. Run the installer on your machine
2. Install to a test location
3. Click desktop shortcut
4. Verify:
   - Console window appears
   - Browser opens to localhost:8080
   - Shows `{"status":"ok"}`

## File Structure After Install

```
C:\Program Files\ChronicleCore\
├── ChronicleCore.bat         ← Launcher (what shortcut runs)
├── chroniclecore.exe          ← Backend
├── python\
│   ├── python.exe
│   └── Lib\
│       └── site-packages\     ← ML packages here
├── ml\
│   ├── src\
│   └── requirements.txt
└── spec\
    └── schema.sql

Desktop:
└── ChronicleCore.lnk          ← Shortcut

Data:
%LOCALAPPDATA%\ChronicleCore\
└── chronicle.db               ← User's data
```

## Advantages

| Feature | Old Way | New Way |
|---------|---------|---------|
| **User install** | Manual Python + commands | Double-click installer |
| **Dependencies** | User installs | Embedded automatically |
| **First run** | Complex setup | Click shortcut |
| **Updates** | Full reinstall | Replace 11 MB EXE |
| **Support** | High (many steps) | Low (just works) |

## Summary

**You run once:**
```powershell
.\BUILD_INSTALLER.ps1 -Version "1.0.0"
```

**She runs once:**
- Double-click `ChronicleCore_Setup_v1.0.0.exe`

**She uses:**
- Double-click desktop shortcut "ChronicleCore"

**That's it!** No Python install, no commands, no configuration.

---

## Quick Commands

```powershell
# Build installer
.\BUILD_INSTALLER.ps1 -Version "1.0.0"

# Skip preparation (if already done)
.\BUILD_INSTALLER.ps1 -Version "1.0.0" -SkipPrepare

# Just prepare files (no installer)
.\prepare_installer.ps1 -Version "1.0.0"

# Just build installer (files already prepared)
.\build_inno_installer.ps1 -Version "1.0.0"
```

---

**Ready?** Run `.\BUILD_INSTALLER.ps1 -Version "1.0.0"` and get coffee! ☕
