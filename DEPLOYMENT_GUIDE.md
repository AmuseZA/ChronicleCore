# ChronicleCore - Deployment & Update Guide

## For Your Fiancé (End User)

### First-Time Installation (One-Click)

**What you need to send her:**
- Entire `ChronicleCore` folder (or zip it)

**What she does:**
1. Extract the zip (if zipped)
2. **Double-click `install.bat`**
3. Follow the prompts
4. Click desktop shortcut when done

**That's it!** The installer handles:
- ✅ Python dependency check (opens download if missing)
- ✅ ML library installation
- ✅ Desktop shortcut creation
- ✅ Startup script creation
- ✅ Data directory setup

**Time:** 5 minutes (or 15 if Python needs installing)

---

## For You (Developer)

### How to Deploy Updates

When you make changes to the backend, here's how to get them to her:

#### Option 1: Full Package (First Time Only)

1. **Build the backend:**
   ```bash
   cd apps\chroniclecore-core
   build.bat
   ```

2. **Package everything:**
   ```
   ChronicleCore/
   ├── apps/
   │   ├── chroniclecore-core/
   │   │   └── chroniclecore.exe     ← New binary
   │   └── chronicle-ml/              ← ML sidecar
   ├── install.bat                    ← Installer
   ├── update.bat                     ← Updater
   ├── QUICK_START.md
   └── VERSION.txt
   ```

3. **Zip it:**
   ```bash
   Compress-Archive -Path ChronicleCore -DestinationPath ChronicleCore-v1.0.0.zip
   ```

4. **Send via:** Email, OneDrive, Dropbox, USB, etc.

5. **She extracts and runs `install.bat`**

#### Option 2: Update Only (After First Install)

**For subsequent updates, only send the new binary:**

1. **Build new version:**
   ```bash
   cd apps\chroniclecore-core
   build.bat
   ```

2. **Update VERSION.txt:**
   ```
   1.0.1
   2026-01-10
   Fixed currency validation bug
   ```

3. **Package update:**
   ```
   ChronicleCore-Update-v1.0.1/
   ├── apps/chroniclecore-core/chroniclecore.exe  ← New binary
   ├── update.bat                                  ← Updater script
   └── VERSION.txt                                 ← Version info
   ```

4. **Send her the update package**

5. **She does:**
   - Extract to her existing ChronicleCore folder (overwrite)
   - Double-click `update.bat`
   - Script handles backup, stop, update, restart

**Time:** 2 minutes

---

## Update Workflow Details

### What `update.bat` Does

1. **Checks if running** → Offers to stop it
2. **Backs up old binary** → `%LOCALAPPDATA%\ChronicleCore\backups\chroniclecore_TIMESTAMP.exe`
3. **Updates ML dependencies** → `pip install --upgrade`
4. **Verifies new binary** → Checks file exists and size
5. **Offers to restart** → Starts new version

### Database Migrations

If you change the database schema:

1. **Create migration file:**
   ```sql
   -- spec/migrations/002_add_new_feature.sql
   ALTER TABLE ...
   ```

2. **Include in update package:**
   ```
   ChronicleCore-Update/
   ├── apps/chroniclecore-core/chroniclecore.exe
   ├── spec/migrations/002_add_new_feature.sql  ← New
   ├── migrate.bat                               ← Migration script
   └── update.bat
   ```

3. **She runs:** `migrate.bat` (or you build auto-migration into backend)

---

## Development Workflow

### Making Changes

1. **Edit code** in your dev environment
2. **Test locally** - verify it works
3. **Build binary:**
   ```bash
   cd apps\chroniclecore-core
   build.bat
   ```
4. **Bump version** in `VERSION.txt`
5. **Package update** (see Option 2 above)
6. **Send to her**

### Testing Before Deployment

**Run these checks:**

```bash
# Build succeeds
cd apps\chroniclecore-core
build.bat

# Binary runs
chroniclecore.exe
# Should start without errors

# Health check
curl http://localhost:8080/health
# Should return {"status":"ok"}

# ML sidecar running
# Check logs for "✓ ML sidecar running"
```

### Version Numbering

Use semantic versioning:
- `1.0.0` - Major release (breaking changes)
- `1.0.1` - Minor release (bug fixes)
- `1.1.0` - Feature release (new features)

---

## Distribution Options

### Option A: Direct File Share (Current)
**Pros:** Simple, no infrastructure
**Cons:** Manual process
**Best for:** Single user (your fiancé)

**Steps:**
1. Build binary
2. Zip update package
3. Send via email/OneDrive/Dropbox
4. She extracts and runs `update.bat`

### Option B: Shared Folder (Recommended)
**Pros:** Automatic updates possible
**Cons:** Requires shared storage
**Best for:** Regular updates

**Setup:**
1. Create OneDrive/Dropbox shared folder
2. She installs from there
3. You update the binary in shared folder
4. She runs `update.bat` when notified

### Option C: GitHub Releases (Future)
**Pros:** Professional, version history
**Cons:** Public or requires GitHub account
**Best for:** Multiple users or open-source

**Setup:**
1. Create private GitHub repo
2. Push code
3. Create releases with binaries
4. She downloads from Releases page
5. Built-in update checker in app

### Option D: Auto-Updater (Future)
**Pros:** Zero-friction updates
**Cons:** Requires backend changes
**Best for:** Production app

**Implementation:**
- Backend checks version API
- Shows "Update available" notification
- Downloads and applies update
- Restarts automatically

---

## File Size Considerations

**Full Package:** ~30-50 MB
- Backend binary: ~20 MB
- Python ML sidecar: ~5 KB (code)
- Documentation: ~1 MB

**Update Package:** ~20 MB
- Just the binary

**For large files:** Use cloud storage links instead of email attachments

---

## Backup & Rollback

### Automatic Backups

`update.bat` automatically backs up to:
```
%LOCALAPPDATA%\ChronicleCore\backups\chroniclecore_TIMESTAMP.exe
```

### Manual Rollback

If new version has issues:

1. **Stop ChronicleCore**
2. **Restore backup:**
   ```bash
   copy "%LOCALAPPDATA%\ChronicleCore\backups\chroniclecore_20260109_143000.exe" ^
        "C:\Path\To\ChronicleCore\apps\chroniclecore-core\chroniclecore.exe"
   ```
3. **Restart**

### Database Backup

**Before major updates:**
```bash
# Backup database
copy "%LOCALAPPDATA%\ChronicleCore\chronicle.db" ^
     "%LOCALAPPDATA%\ChronicleCore\chronicle.db.backup"
```

**Restore if needed:**
```bash
copy "%LOCALAPPDATA%\ChronicleCore\chronicle.db.backup" ^
     "%LOCALAPPDATA%\ChronicleCore\chronicle.db"
```

---

## Troubleshooting Updates

### Issue: Update script fails

**Solution:**
1. Manually stop ChronicleCore (Task Manager)
2. Copy new `chroniclecore.exe` over old one
3. Restart via desktop shortcut

### Issue: New version won't start

**Solution:**
1. Check logs in console window
2. Rollback to previous version (see above)
3. Send logs to you for debugging

### Issue: Database schema mismatch

**Error:** `no such column: xyz`

**Solution:**
1. Run migration script if provided
2. Or delete database (loses data!)
3. Or rollback to previous version

### Issue: ML sidecar not working after update

**Solution:**
```bash
cd apps\chronicle-ml
pip install --upgrade -r requirements.txt
```

---

## Communication Template

### Sending First Install

**Email Subject:** ChronicleCore Time Tracker - Ready to Install!

**Email Body:**
```
Hi [Name],

ChronicleCore is ready for you to try!

INSTALLATION (5 minutes):
1. Extract the attached zip file
2. Double-click "install.bat"
3. Follow the prompts (it will check for Python)
4. Click the desktop shortcut when done

WHAT IT DOES:
- Tracks your time automatically
- Learns which work belongs to which client
- Generates accurate invoices with proper currency

FIRST WEEK:
- Use it normally throughout your day
- Manually assign 50+ blocks to different clients
- After that, it starts auto-suggesting (80-90% accurate!)

DOCUMENTATION:
- Quick start guide: QUICK_START.md
- Full guide: workflow/ml_user_guide.md

Let me know if you have any questions!
```

### Sending Updates

**Email Subject:** ChronicleCore Update - Version 1.0.1

**Email Body:**
```
Hi [Name],

New update available with improvements!

WHAT'S NEW:
- Fixed currency validation bug
- Improved ML accuracy
- Performance optimizations

TO UPDATE (2 minutes):
1. Extract the attached zip to your ChronicleCore folder
2. Double-click "update.bat"
3. It will backup old version and install new one

That's it!
```

---

## Quick Reference

### First Install
```bash
1. Send: Full ChronicleCore.zip (30-50 MB)
2. She: Extracts + runs install.bat
3. Done: Desktop shortcut created, ready to use
```

### Updates
```bash
1. You: Build new binary (build.bat)
2. You: Update VERSION.txt
3. You: Package apps/chroniclecore-core/chroniclecore.exe + update.bat
4. You: Send update package (~20 MB)
5. She: Extracts to ChronicleCore folder (overwrite)
6. She: Runs update.bat
7. Done: New version running, old backed up
```

### Rollback
```bash
1. Stop ChronicleCore
2. Copy backup from %LOCALAPPDATA%\ChronicleCore\backups\
3. Paste to apps\chroniclecore-core\chroniclecore.exe
4. Restart
```

---

## Future Improvements

### Phase 1 (Current)
- ✅ One-click installer
- ✅ Update script with backup
- ✅ Desktop shortcut
- ✅ Manual update process

### Phase 2 (Future)
- ⏳ Auto-update checker (built into app)
- ⏳ In-app update notifications
- ⏳ One-click update from UI
- ⏳ Automatic database migrations

### Phase 3 (Future)
- ⏳ MSI/NSIS installer (Windows native)
- ⏳ Auto-start on Windows boot
- ⏳ System tray icon
- ⏳ Silent updates

---

## Summary

**For her:** Double-click `install.bat` (first time) or `update.bat` (updates)

**For you:** Build, zip, send. She runs the script.

**Simple, reliable, and no manual commands needed!** 🎉
