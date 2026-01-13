# ChronicleCore Troubleshooting Guide

## Issue: "Connection Refused" when browser opens

### Symptoms
- Desktop shortcut runs
- Browser opens to http://localhost:8080
- Shows "ERR_CONNECTION_REFUSED" or "Can't connect"
- Console window closes immediately

### Causes & Solutions

#### 1. Backend Not Starting

**Check:**
Run the DEBUG version to see actual error messages:
- Go to: `C:\Program Files\ChronicleCore\`
- Double-click: `ChronicleCore_DEBUG.bat`
- Read the error messages

**Common issues:**
- Missing DLL files
- Antivirus blocking
- Firewall blocking

#### 2. Port 8080 Already in Use

**Check:**
```cmd
netstat -ano | findstr :8080
```

**Solution:**
- Close the other application using port 8080, OR
- Kill the process:
  ```cmd
  taskkill /PID <process_id> /F
  ```

#### 3. Antivirus/Windows Defender Blocking

**Symptoms:**
- EXE starts briefly then closes
- No error message
- Process disappears from Task Manager

**Solution:**
Add exclusion to Windows Defender:
1. Open Windows Security
2. Virus & threat protection → Manage settings
3. Exclusions → Add exclusion → Folder
4. Add: `C:\Program Files\ChronicleCore\`

#### 4. Firewall Blocking

**Check:**
```cmd
netsh advfirewall firewall show rule name="ChronicleCore"
```

**Solution:**
Allow through firewall:
```cmd
netsh advfirewall firewall add rule name="ChronicleCore" dir=in action=allow program="C:\Program Files\ChronicleCore\chroniclecore.exe" enable=yes
```

Or use GUI:
1. Windows Defender Firewall
2. Advanced settings → Inbound Rules
3. New Rule → Program
4. Browse to: `C:\Program Files\ChronicleCore\chroniclecore.exe`
5. Allow connection

#### 5. Missing Visual C++ Runtime

**Symptoms:**
- Error about missing VCRUNTIME140.dll or similar

**Solution:**
Install Visual C++ Redistributable:
- Download from: https://aka.ms/vs/17/release/vc_redist.x64.exe
- Run installer
- Restart ChronicleCore

#### 6. Wrong Working Directory

**Check:**
Open ChronicleCore_DEBUG.bat and verify it shows:
```
Current directory: C:\Program Files\ChronicleCore
```

**If wrong, the shortcut may be misconfigured.**

**Solution:**
Recreate shortcut:
1. Right-click `ChronicleCore.bat`
2. Send to → Desktop (create shortcut)
3. Right-click shortcut → Properties
4. Start in: `C:\Program Files\ChronicleCore\`

---

## Issue: Backend Starts But Crashes Immediately

### Check Logs

Backend creates database in:
```
%LOCALAPPDATA%\ChronicleCore\chronicle.db
```

**To view:**
```cmd
explorer %LOCALAPPDATA%\ChronicleCore
```

### Common Causes

#### Database Corruption

**Solution:**
Delete and recreate:
```cmd
del "%LOCALAPPDATA%\ChronicleCore\chronicle.db"
```

Restart ChronicleCore (it will recreate the database)

#### Missing Schema File

**Check:**
`C:\Program Files\ChronicleCore\spec\schema.sql` exists

**If missing:**
Reinstall ChronicleCore

---

## Issue: ML Sidecar Not Starting

### Symptoms
- Backend starts fine
- API works (http://localhost:8080/health returns OK)
- But ML features don't work
- Logs show: "ML sidecar disabled"

### Check ML Dependencies

Run in PowerShell:
```powershell
cd "C:\Program Files\ChronicleCore"
.\python\python.exe -c "import fastapi; import sklearn; import pandas; print('ML deps OK')"
```

**If error:**
```powershell
.\python\python.exe -m pip install -r ml\requirements.txt
```

### Check Python Path

Run:
```cmd
cd "C:\Program Files\ChronicleCore"
dir python\python.exe
```

Should show the file exists.

---

## Issue: Browser Opens But Shows Blank Page

### Cause
UI not built or not included

### Check
Visit: http://localhost:8080/health

**If shows:** `{"status":"ok"}` → Backend working, UI missing

**Solution:**
Backend is working! The issue is the UI needs to be built separately.

For now, you can test the API directly:
- Health: http://localhost:8080/health
- Tracking status: http://localhost:8080/api/v1/tracking/status

---

## Issue: "Access Denied" or "Permission Error"

### Cause
Installer needs admin rights or files are in protected location

### Solution

**Option A: Run as Administrator**
- Right-click ChronicleCore.bat
- Run as administrator

**Option B: Install to User Directory**
- Uninstall current version
- Reinstall and choose install location:
  `C:\Users\<YourName>\AppData\Local\Programs\ChronicleCore\`

---

## Debug Mode - Full Output

To see **all** output and error messages:

1. Go to: `C:\Program Files\ChronicleCore\`
2. Run: `ChronicleCore_DEBUG.bat`
3. Read all messages
4. Backend output will show in console

This version:
- Shows all checks
- Displays all error messages
- Runs backend in foreground
- Doesn't close on error

---

## Manual Start (for testing)

To test backend directly:

```cmd
cd "C:\Program Files\ChronicleCore"
chroniclecore.exe
```

Leave console open. Check for errors.

To test in different directory:
```cmd
cd C:\Temp
"C:\Program Files\ChronicleCore\chroniclecore.exe"
```

Should still work (backend finds its files automatically).

---

## Check What's Running

### Is backend running?

```cmd
tasklist | findstr chroniclecore
```

**If found:** Backend is running

**To kill:**
```cmd
taskkill /IM chroniclecore.exe /F
```

### Is port in use?

```cmd
netstat -ano | findstr :8080
```

---

## Reinstall

If all else fails:

1. **Uninstall:**
   - Settings → Apps → ChronicleCore → Uninstall
   - Choose: Keep data (YES) or Delete data (NO)

2. **Clean install:**
   - Delete: `C:\Program Files\ChronicleCore\` (if exists)
   - Delete: `%LOCALAPPDATA%\ChronicleCore\` (optional, loses data)

3. **Reinstall:**
   - Run: `ChronicleCore_Setup_v1.0.0.exe`
   - Try different install location if permission issues

---

## Getting Help

When asking for help, provide:

1. **Error message** from ChronicleCore_DEBUG.bat
2. **Windows version:** `winver` command
3. **Antivirus:** What antivirus is running?
4. **Firewall:** Is Windows Firewall enabled?
5. **Port check:** Output of `netstat -ano | findstr :8080`
6. **Install location:** Where did you install?
7. **Logs:** Contents of `%LOCALAPPDATA%\ChronicleCore\`

---

## Quick Diagnostics

Run this in Command Prompt:

```cmd
@echo off
echo === ChronicleCore Diagnostics ===
echo.

echo Windows Version:
ver

echo.
echo Port 8080 status:
netstat -ano | findstr :8080

echo.
echo ChronicleCore process:
tasklist | findstr chroniclecore

echo.
echo Installation check:
dir "C:\Program Files\ChronicleCore\chroniclecore.exe"

echo.
echo Python check:
dir "C:\Program Files\ChronicleCore\python\python.exe"

echo.
echo Database check:
dir "%LOCALAPPDATA%\ChronicleCore\chronicle.db"

echo.
echo === End Diagnostics ===
pause
```

Save as `diagnostics.bat` and run it.

---

## Common Solutions Summary

| Issue | Quick Fix |
|-------|-----------|
| Connection refused | Run ChronicleCore_DEBUG.bat to see error |
| Port in use | `netstat -ano \| findstr :8080` then kill process |
| Antivirus blocking | Add exclusion for ChronicleCore folder |
| Firewall blocking | Allow chroniclecore.exe through firewall |
| Missing DLL | Install VC++ Redistributable |
| Permission denied | Run as administrator or reinstall to user folder |
| Crashes immediately | Check %LOCALAPPDATA%\ChronicleCore for logs |

---

**Still stuck?** Run `ChronicleCore_DEBUG.bat` and send the output!
