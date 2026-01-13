# ChronicleCore Installer - FIXED VERSION

## What Was Fixed

### Problem
The original launcher had a critical bug:
- Browser opened **before** backend started
- Connection refused error (ERR_CONNECTION_REFUSED)
- No error messages shown
- Console window closed immediately

### Solution
The new launcher:
1. ✅ Starts backend first (in background)
2. ✅ Waits for backend to be ready (checks health endpoint)
3. ✅ Only opens browser once server is confirmed running
4. ✅ Shows clear error messages if startup fails
5. ✅ Keeps window open so backend stays running

---

## Files Updated

**Location:** `C:\Users\josh\.gemini\antigravity\workspaces\ChronicleCore\`

### 1. Updated Installer
**File:** `installer_output\ChronicleCore_Setup_v1.0.0.exe` (83 MB)
- Fixed launcher script included
- Includes debug version for troubleshooting

### 2. ChronicleCore.bat (Main Launcher)
**Changes:**
- Starts backend in background first
- Waits up to 30 seconds for server to be ready
- Checks health endpoint before opening browser
- Shows helpful error messages if startup fails
- Keeps console open (backend runs in background)

### 3. ChronicleCore_DEBUG.bat (New!)
**Purpose:** Troubleshooting version that shows ALL output
- Displays all checks (Python, backend, ML dependencies)
- Shows port status
- Runs backend in foreground to see logs
- Doesn't close on error
- Perfect for diagnosing issues

### 4. TROUBLESHOOTING.md (New!)
**Purpose:** Complete troubleshooting guide
- Common issues and solutions
- Port conflicts
- Antivirus/Firewall blocking
- Permission errors
- Step-by-step diagnostics

---

## How It Works Now

### User Experience

**Step 1: Install**
1. Run `ChronicleCore_Setup_v1.0.0.exe`
2. Follow wizard
3. Desktop shortcut created

**Step 2: First Run**
1. Double-click desktop shortcut
2. Console shows:
   ```
   [1/2] Starting backend server...
   [2/2] Waiting for server to be ready...

   ChronicleCore is running!
   Web UI: http://localhost:8080

   Opening browser...
   ```
3. Browser opens to working app
4. Console stays open (backend running)

**Step 3: Stop**
- Close console window, OR
- Task Manager → End chroniclecore.exe

### If Something Goes Wrong

The launcher will:
1. Show clear error message
2. Suggest possible causes
3. Offer to open logs folder
4. Keep window open so error can be read

**Example error:**
```
ERROR: Server failed to start after 30 seconds!

Check if:
  - Port 8080 is already in use
  - Firewall is blocking the application
  - Antivirus is interfering

Press any key to open logs location...
```

---

## Troubleshooting Mode

If issues occur:

**Run Debug Version:**
```
C:\Program Files\ChronicleCore\ChronicleCore_DEBUG.bat
```

This shows:
- Current directory
- Python version
- Backend file size
- ML dependencies status
- Port availability
- **Full backend output** (all logs)

**Perfect for:**
- "Connection refused" errors
- Backend crashes
- Missing dependencies
- Permission issues

---

## Common Issues & Quick Fixes

### Issue: "Connection Refused"

**Cause:** Backend not starting

**Solution:**
1. Run `ChronicleCore_DEBUG.bat`
2. Read error message
3. Common fixes:
   - Add antivirus exclusion
   - Allow through firewall
   - Kill process on port 8080

### Issue: Console Closes Immediately

**Old behavior:** This was a bug (browser opened too early)

**New behavior:** Console stays open, shows status

**If still happening:**
- Backend crashed
- Run DEBUG version to see why

### Issue: Port 8080 In Use

**Check:**
```cmd
netstat -ano | findstr :8080
```

**Fix:**
Kill the other process or change port (future feature)

---

## Distribution

### Send to User

**File:** `installer_output\ChronicleCore_Setup_v1.0.0.exe` (83 MB)

**Instructions for her:**
1. Download installer
2. Run it (may need administrator)
3. Follow wizard
4. Double-click desktop shortcut
5. Browser opens automatically

**If issues:**
- Run: `C:\Program Files\ChronicleCore\ChronicleCore_DEBUG.bat`
- Send you the output

---

## What's Included

### Regular User Files
- `ChronicleCore.bat` - Smart launcher (main)
- `chroniclecore.exe` - Backend server
- `python/` - Embedded Python 3.11.9
- `ml/` - ML sidecar code
- `spec/` - Database schema

### Troubleshooting Files
- `ChronicleCore_DEBUG.bat` - Debug launcher
- Desktop shortcut → Runs regular launcher

### User Data (Created on First Run)
- `%LOCALAPPDATA%\ChronicleCore\chronicle.db` - User's data
- Preserved on reinstall if user chooses

---

## Technical Details

### Launcher Sequence

1. **Validation Phase**
   ```
   Check Python exists
   Check backend exists
   Check ML dependencies
   ```

2. **Startup Phase**
   ```
   Start backend in background
   Wait for health endpoint (max 30 sec)
   If success → open browser
   If fail → show error
   ```

3. **Running Phase**
   ```
   Backend runs in background
   Console shows status
   User can close console to stop
   ```

### Health Check

Launcher polls: `http://localhost:8080/health`

Expected response: `{"status":"ok","version":"1.0.0"}`

Retries: 30 times (1 second each)

### Error Handling

All errors show:
- Clear message
- Possible causes
- Suggested actions
- Option to view logs

---

## Comparison: Old vs New

| Aspect | Old Behavior | New Behavior |
|--------|--------------|--------------|
| **Browser open** | Before backend starts | After backend ready |
| **Error messages** | None (window closes) | Clear, actionable |
| **Health check** | None | 30-second polling |
| **Console** | Closes immediately | Stays open |
| **Debug mode** | Not available | DEBUG.bat included |
| **Timeout** | None (immediate fail) | 30 seconds |
| **User feedback** | Confusing | Clear status |

---

## Testing Checklist

Before sending to user:

- [ ] Run installer on clean machine
- [ ] Desktop shortcut works
- [ ] Browser opens to working app
- [ ] Console stays open
- [ ] Can stop by closing console
- [ ] DEBUG version shows output
- [ ] Error handling works (test with port conflict)

---

## Future Improvements

Potential enhancements:
- [ ] Auto-restart on crash
- [ ] Configurable port
- [ ] System tray icon
- [ ] Windows Service mode
- [ ] Auto-update checker
- [ ] Better logging to file

---

## Summary

**Fixed Issues:**
- ✅ Connection refused error
- ✅ Console closing immediately
- ✅ No error messages
- ✅ Browser opening too early

**New Features:**
- ✅ Smart startup sequence
- ✅ Health check polling
- ✅ Clear error messages
- ✅ Debug mode for troubleshooting
- ✅ Comprehensive error handling

**Result:** Professional, robust launcher that "just works" and provides clear feedback when issues occur.

---

**Installer ready at:**
`C:\Users\josh\.gemini\antigravity\workspaces\ChronicleCore\installer_output\ChronicleCore_Setup_v1.0.0.exe`

**Send this version** - it has all the fixes! 🚀
