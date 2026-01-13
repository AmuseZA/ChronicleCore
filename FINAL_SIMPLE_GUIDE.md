# ChronicleCore - Dead Simple Deployment

## ✅ What You Have Now

A folder ready to ZIP and send: `deploy\ChronicleCore\`

## 🚀 How to Deploy (2 Steps)

### Step 1: Create ZIP

```
Right-click deploy\ChronicleCore folder
→ Send to → Compressed (zipped) folder
```

Result: `ChronicleCore.zip` (~12 MB)

### Step 2: Send to Her

Upload to Dropbox/OneDrive/Email and send link.

**That's it from your side!**

---

## 👤 What She Does (3 Steps)

### Step 1: Install Python (One-Time, 5 min)

- Download: https://www.python.org/downloads/
- Run installer
- **IMPORTANT:** Check "Add Python to PATH"
- Click Install

### Step 2: Extract ZIP

- Download ChronicleCore.zip
- Right-click → Extract All
- Choose location (e.g., Desktop or Documents)

### Step 3: Run

- Open extracted folder
- Double-click `START.bat`
- First run: Installs dependencies automatically (1-2 min)
- Browser opens to http://localhost:8080/health
- If shows `{"status":"ok"}` → Working!

**Done!** Leave the window open while working.

---

## 🔄 Future Updates (Easy)

### When You Make Changes

```powershell
# 1. Build
cd apps\chroniclecore-core
.\build.bat

# 2. Package
cd ..\..
.\CREATE_PACKAGE.ps1

# 3. Send her just the new chroniclecore.exe (11 MB)
```

### She Updates

1. Close ChronicleCore
2. Replace old `chroniclecore.exe` with new one
3. Double-click START.bat again

**No reinstall needed!**

---

## 📦 What's in the Package

```
ChronicleCore/
├── START.bat            ← Double-click this
├── chroniclecore.exe    ← Backend (11 MB)
├── ml/                  ← ML sidecar code
│   ├── src/
│   └── requirements.txt
├── spec/
│   └── schema.sql
└── README.md
```

---

## 🐛 Common Issues

### "Python not installed"

START.bat will detect this and open python.org automatically.

### "Dependencies failed"

Check internet connection. Dependencies install from PyPI.

### "Port 8080 in use"

Another program is using that port. Close it or change port in code.

### Backend doesn't start

1. Check console window for errors
2. Try running `chroniclecore.exe` directly
3. Check firewall isn't blocking

---

## ✨ Advantages

| Task | Old Way | New Way |
|------|---------|---------|
| **Initial setup** | Manual commands | Double-click START.bat |
| **Dependencies** | Manual pip install | Auto-installed first run |
| **Updates** | Full reinstall | Replace 11 MB EXE |
| **User steps** | 10+ | 3 |

---

## 📊 File Sizes

- **Full package:** ~12 MB (without Python)
- **Update:** ~11 MB (just EXE)
- **With Python installed:** Auto-downloads ~40 MB dependencies first run

---

## 🎯 Complete Workflow

### You (First Time)

```powershell
# Already done!
.\CREATE_PACKAGE.ps1
```

### Her (First Time)

1. Install Python (5 min)
2. Extract ZIP (30 sec)
3. Double-click START.bat (2 min first run)

### You (Updates)

```powershell
cd apps\chroniclecore-core
.\build.bat
```

Send her new `chroniclecore.exe`

### Her (Updates)

1. Close app
2. Replace EXE
3. Start again

---

## 🔒 Security Notes

- ✅ Everything runs locally (localhost:8080)
- ✅ No external network calls
- ✅ Data stored in %LOCALAPPDATA%\ChronicleCore
- ✅ No admin rights needed

---

## ✅ Ready to Send

Your package is at: `deploy\ChronicleCore\`

**Just ZIP it and send!**

---

## 💡 Tips

1. **Test first:** Extract and run START.bat on your machine
2. **Include README:** It's already in the package
3. **Check Python:** Make sure she checks "Add to PATH"
4. **First run:** Tell her first run takes 1-2 min (installs deps)
5. **Keep window open:** Console window must stay open while using app

---

## 🆘 If Something Goes Wrong

The package includes START.bat which:
- Checks for Python before starting
- Shows clear error messages
- Auto-installs dependencies
- Opens browser to health check

If she has issues:
1. Check console window for errors
2. Verify Python installed correctly
3. Check firewall isn't blocking
4. Try running chroniclecore.exe directly to see full error

---

**YOU'RE DONE!** Just ZIP and send the deploy\ChronicleCore folder. 🎉
