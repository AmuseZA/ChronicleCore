# ChronicleCore Installer - Quick Start

## 🎯 Goal

Create a **single EXE installer** with all dependencies that your fiancé can double-click and use immediately.

---

## 📋 Prerequisites (One-Time Setup)

### 1. Install Inno Setup (Free)

Download and install from: https://jrsoftware.org/isdl.php

**Why?** Creates professional Windows installers (used by VS Code, Firefox, etc.)

### 2. Verify You Have

- [x] Go 1.25 installed
- [x] GCC (MinGW) installed
- [x] Backend builds successfully (`build.bat`)

---

## 🚀 Creating First Installer

### Step 1: Build Everything

```powershell
# Run from ChronicleCore root directory
.\build_installer.ps1 -Version "1.0.0"
```

**What it does:**
1. Builds backend binary
2. Downloads Python 3.11.9 (embedded, ~15 MB)
3. Installs ML packages (FastAPI, scikit-learn, etc.)
4. Packages everything into single EXE installer

**Time:** 5-10 minutes first run, 2 minutes subsequent runs

**Output:** `installer_output/ChronicleCore_Setup_v1.0.0.exe` (~80 MB)

### Step 2: Test Locally

```powershell
# Run the installer on your machine
.\installer_output\ChronicleCore_Setup_v1.0.0.exe
```

Verify:
- [x] Installs to `C:\Users\YOU\AppData\Local\Programs\ChronicleCore`
- [x] Creates desktop shortcut
- [x] Backend starts on double-click
- [x] http://localhost:8080/health returns `{"status":"ok"}`

### Step 3: Send to Her

1. Upload `ChronicleCore_Setup_v1.0.0.exe` to Dropbox/OneDrive
2. Send her the link
3. Tell her: "Download and double-click!"

**That's it!** No manual steps, no Python installation, no commands.

---

## 🔄 Pushing Updates (After Initial Install)

### When You Make Changes

```powershell
# 1. Make code changes
code apps\chroniclecore-core\internal\api\profiles.go

# 2. Build backend
cd apps\chroniclecore-core
.\build.bat

# 3. Create update package
cd ..\..
.\create_update.ps1 -Version "1.0.1" -Changes "Fixed currency bug"
```

**Output:** `updates/v1.0.1/` (~20 MB, just the binary)

### Distribute Update

**Option 1: Simple File Share (Recommended for 1 user)**

1. Upload `updates/v1.0.1/chroniclecore.exe` to Dropbox
2. Send her the link
3. Tell her: "Download, close ChronicleCore, replace the EXE, restart"

**Option 2: Auto-Update (Future)**

Implement auto-update checker (see `UPDATE_DISTRIBUTION.md`)

---

## 📦 What's in the Installer?

```
ChronicleCore_Setup_v1.0.0.exe (80 MB)
│
├── chroniclecore.exe (20 MB)      ← Backend
├── python/ (15 MB)                ← Embedded Python 3.11
├── python_packages/ (40 MB)       ← ML dependencies
├── ml/ (1 MB)                     ← ML sidecar code
├── spec/schema.sql                ← Database schema
└── docs/                          ← User guides
```

**Key benefit:** User doesn't install Python separately. Everything is bundled.

---

## 🎯 Your Workflow

### First Deployment

```
1. build_installer.ps1 -Version "1.0.0"
2. Test locally
3. Send ChronicleCore_Setup_v1.0.0.exe to her
4. She installs (one-click)
```

### Future Updates

```
1. Make changes
2. build.bat (just backend)
3. create_update.ps1 -Version "1.0.X"
4. Send 20 MB update, not 80 MB installer
```

---

## ⚙️ Configuration

### Change App Name/Publisher

Edit `chroniclecore_installer.iss`:

```iss
#define MyAppName "ChronicleCore"
#define MyAppPublisher "Your Name"
#define MyAppURL "https://github.com/yourname/chroniclecore"
```

### Add Custom Icon

1. Create `resources/app_icon.ico` (256x256)
2. Uncomment in `chroniclecore_installer.iss`:
   ```iss
   SetupIconFile=resources\app_icon.ico
   ```

### Add License

1. Create `LICENSE.txt` in root
2. Installer will show it during setup

---

## 🐛 Troubleshooting

### "Inno Setup not found"

Install from: https://jrsoftware.org/isdl.php

Default path: `C:\Program Files (x86)\Inno Setup 6\ISCC.exe`

### "Python download failed"

Check internet connection. Script downloads from python.org

Or download manually:
```
https://www.python.org/ftp/python/3.11.9/python-3.11.9-embed-amd64.zip
```
Extract to `dist/python/`

### "Backend build failed"

Check:
- Go installed: `go version`
- GCC installed: `gcc --version`
- Run `build.bat` directly first

### "ML dependencies failed"

Run manually:
```powershell
cd dist
.\python\python.exe -m pip install -r ..\apps\chronicle-ml\requirements.txt -t .\python_packages
```

### "Installer is huge (>100 MB)"

Normal! Includes:
- Backend (20 MB)
- Python runtime (15 MB)
- ML packages (40 MB)
- UI assets (5 MB)

Updates are only 20 MB.

---

## 📊 Size Comparison

| Method | First Install | Updates | User Steps |
|--------|--------------|---------|------------|
| **This Method** | 80 MB EXE | 20 MB | 1-click |
| ZIP + Manual | 20 MB + Python | 20 MB | 10+ steps |
| MSI Installer | 80 MB MSI | 80 MB | 3 clicks |

---

## ✅ Checklist

Before sending to user:

- [ ] Installer builds successfully
- [ ] Test install on your machine
- [ ] Backend starts and responds on :8080
- [ ] ML sidecar starts (check logs)
- [ ] Can create client/profile
- [ ] Can assign blocks
- [ ] CSV export works
- [ ] Desktop shortcut works
- [ ] Uninstall works

---

## 🎓 Advanced

### Include UI (If Built)

```powershell
# Build UI first
cd apps\chroniclecore-ui
npm run build

# Then build installer
cd ..\..
.\build_installer.ps1 -Version "1.0.0"
```

Installer will automatically include UI if `build/` folder exists.

### Custom Install Location

User can change during install wizard. Default: `%LocalAppData%\Programs\ChronicleCore`

### Silent Install

For automated deployment:
```cmd
ChronicleCore_Setup_v1.0.0.exe /SILENT
```

### Create Portable Version

Skip installer, use ZIP:
```powershell
.\build_installer.ps1 -Version "1.0.0" -SkipInno
```

Output: `dist/` folder (just ZIP it)

---

## 📚 Related Docs

- **UPDATE_DISTRIBUTION.md** - Auto-update system
- **PACKAGING_GUIDE.md** - Technical details
- **DEPLOYMENT_CHECKLIST.md** - Pre-deployment checks
- **QUICK_START.md** - User guide (include in installer)

---

## 🆘 Still Stuck?

### Quick Test

```powershell
# Test just the build, skip installer
.\build_installer.ps1 -Version "1.0.0"

# If it fails, check the error message
# Most common: Inno Setup not installed
```

### Manual Build

If PowerShell script fails, build manually:

1. Build backend: `cd apps\chroniclecore-core && build.bat`
2. Download Python embed: https://www.python.org/downloads/windows/
3. Install ML deps: `python -m pip install -r requirements.txt -t dist\python_packages`
4. Run Inno Setup: `iscc chroniclecore_installer.iss`

---

## 🎉 Summary

**You:** Run `build_installer.ps1` once

**Her:** Double-click `ChronicleCore_Setup_v1.0.0.exe`

**Result:** Fully working app with all dependencies!

**Updates:** Just send 20 MB binary, not full installer

---

**Next:** Run `.\build_installer.ps1 -Version "1.0.0"` and test!
