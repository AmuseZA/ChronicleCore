# ChronicleCore Update Distribution System

## Overview

This system allows you to push updates to users **without rebuilding the entire installer**. Users get notified of updates and can install them with one click.

## Architecture

```
Your Machine                  Web Server              User's Machine
┌──────────────┐            ┌──────────────┐        ┌──────────────┐
│ Build update │───────────>│ update.json  │<───────│ Auto-checker │
│ chroniclecore│            │ + binary     │        │ (every 24h)  │
│    .exe      │            └──────────────┘        └──────────────┘
└──────────────┘
```

## Quick Start

### 1. Initial Distribution

**Build full installer:**
```powershell
.\build_installer.ps1 -Version "1.0.0"
```

**Result:** `installer_output/ChronicleCore_Setup_v1.0.0.exe`

**Send to user:** She runs the installer once.

---

### 2. Making Updates

**When you make code changes:**

```powershell
# 1. Build new backend binary
cd apps\chroniclecore-core
.\build.bat

# 2. Create update package
.\create_update.ps1 -Version "1.0.1" -Changes "Fixed currency bug, improved ML accuracy"
```

**Result:** `updates/v1.0.1/update_package.zip` (just the binary + manifest)

---

### 3. Distributing Updates

**Option A: Simple File Hosting (Recommended)**

Upload to Dropbox/OneDrive/Google Drive:
```
/ChronicleCore-Updates/
  ├── update.json          ← Points to latest version
  └── v1.0.1/
      └── chroniclecore.exe
```

**Option B: GitHub Releases**

```bash
gh release create v1.0.1 --title "v1.0.1" --notes "Bug fixes" updates/v1.0.1/chroniclecore.exe
```

**Option C: Your Own Server**

Upload to any web server with public HTTPS:
```
https://yourdomain.com/chroniclecore/updates/
```

---

## Update Manifest Format

**File:** `update.json`

```json
{
  "latest_version": "1.0.1",
  "release_date": "2026-01-09",
  "download_url": "https://your-hosting.com/updates/v1.0.1/chroniclecore.exe",
  "changelog": "- Fixed currency rounding\n- Improved ML accuracy to 85%\n- Added ZAR currency support",
  "mandatory": false,
  "min_version": "1.0.0"
}
```

---

## How Users Get Updates

### Automatic (Recommended)

Backend checks for updates every 24 hours:

1. User opens ChronicleCore
2. Backend fetches `update.json` silently
3. If new version exists:
   - Shows notification in UI: "Update available: v1.0.1"
   - User clicks "Update"
   - Backend downloads new binary
   - Replaces itself (using helper script)
   - Restarts

### Manual

User can check manually:
- UI button: "Check for Updates"
- Or: `GET /api/v1/system/check-update`

---

## Implementation Steps

### Step 1: Create Update Scripts

**Already created for you:**
- `create_update.ps1` - Packages updates
- `internal/api/updates.go` - Update checker endpoint (needs implementation)

### Step 2: Choose Hosting

**Easiest: Dropbox Public Folder**

1. Create folder: `ChronicleCore-Updates`
2. Put `update.json` and binaries there
3. Get public share link
4. Configure in backend:

```go
const UpdateCheckURL = "https://www.dropbox.com/s/YOUR_SHARE_LINK/update.json?dl=1"
```

### Step 3: Configure Auto-Update

Edit `internal/config/config.go`:

```go
const (
    AppVersion = "1.0.0"
    UpdateCheckURL = "https://your-hosting.com/updates/update.json"
    UpdateCheckInterval = 24 * time.Hour
)
```

### Step 4: Test Update Flow

```powershell
# Simulate update
.\test_update.ps1
```

---

## Update Workflow (Day-to-Day)

### When You Fix a Bug

```powershell
# 1. Make code changes
code apps\chroniclecore-core\internal\api\profiles.go

# 2. Build
cd apps\chroniclecore-core
.\build.bat

# 3. Create update package
cd ..\..
.\create_update.ps1 -Version "1.0.2" -Changes "Fixed profile deletion bug"

# 4. Upload
# Copy updates/v1.0.2/chroniclecore.exe to your hosting
# Update update.json with new version

# 5. Done! Users auto-notified within 24h
```

**Time:** 5 minutes

---

## Update Package Contents

**Small update (~20 MB):**
```
updates/v1.0.1/
├── chroniclecore.exe     ← New backend binary
├── CHANGELOG.md          ← What changed
└── manifest.json         ← Version info
```

**Database migration (~20 MB + SQL):**
```
updates/v1.0.1/
├── chroniclecore.exe
├── migrations/
│   └── 002_add_tags.sql  ← SQL migration
├── CHANGELOG.md
└── manifest.json
```

**UI update (~25 MB):**
```
updates/v1.0.1/
├── chroniclecore.exe
├── web/                  ← New UI build
│   ├── index.html
│   └── assets/...
├── CHANGELOG.md
└── manifest.json
```

---

## Security

### Signed Updates (Recommended for Production)

```powershell
# Generate signing key (once)
.\generate_signing_key.ps1

# Sign update package
.\create_update.ps1 -Version "1.0.1" -Sign
```

Updates include SHA256 signature in `update.json`:

```json
{
  "latest_version": "1.0.1",
  "download_url": "...",
  "sha256": "abc123def456...",
  "signature": "xyz789..."
}
```

Backend verifies before applying.

### HTTPS Only

All update checks and downloads use HTTPS.

---

## Rollback

If an update breaks:

```powershell
# User runs
.\rollback.bat

# Or from UI
POST /api/v1/system/rollback
```

Backend keeps last 2 versions for rollback.

---

## Testing Updates

**Test on clean VM:**

1. Install v1.0.0 from installer
2. Configure test update server
3. Trigger update check
4. Verify update downloads and applies
5. Check functionality post-update

**Test script:**
```powershell
.\test_update_flow.ps1
```

---

## Hosting Options Comparison

| Option | Cost | Ease | Reliability | Best For |
|--------|------|------|-------------|----------|
| **Dropbox Public** | Free | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Small user base |
| **GitHub Releases** | Free | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Open source |
| **OneDrive Share** | Free | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Small user base |
| **AWS S3** | ~$0.50/mo | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Production |
| **Your Own Server** | Varies | ⭐⭐ | ⭐⭐⭐ | Full control |

**Recommendation for your use case:** Dropbox or OneDrive (free, easy, reliable)

---

## FAQ

**Q: Does she need to reinstall from scratch for updates?**
A: No! Updates replace just the binary. First install uses the full installer.

**Q: How big are update downloads?**
A: ~20 MB (just the backend binary). First install is ~80 MB (includes Python).

**Q: Can updates fail?**
A: Yes. Backend backs up old version and can rollback automatically.

**Q: Do updates require admin rights?**
A: No, because app installs to user directory.

**Q: Can I push updates while she's using it?**
A: App will notify and prompt to restart. Won't interrupt current work.

**Q: How do I test updates before she gets them?**
A: Use a separate update channel (e.g., `update_beta.json`)

---

## Advanced: Multi-Channel Updates

Support beta/stable channels:

```json
// update_stable.json
{
  "latest_version": "1.0.1",
  "channel": "stable"
}

// update_beta.json
{
  "latest_version": "1.1.0-beta",
  "channel": "beta"
}
```

Users choose channel in settings.

---

## Monitoring

Track update adoption:

```go
// Backend logs
log.Printf("Update check: current=%s, latest=%s, action=%s", currentVer, latestVer, action)
```

Optionally send anonymous telemetry (with user consent).

---

## Summary

### First Distribution
1. Run `build_installer.ps1`
2. Send `ChronicleCore_Setup_v1.0.0.exe` to user
3. She installs once

### Every Update After
1. Make changes
2. Run `create_update.ps1`
3. Upload `chroniclecore.exe` to hosting
4. Update `update.json`
5. Users auto-notified within 24h

**No more full installers!** Just 20 MB binary updates.

---

## Next Steps

1. **Now:** Choose hosting (Dropbox recommended)
2. **Today:** Implement update checker endpoint
3. **Tomorrow:** Test update flow end-to-end
4. **This week:** Deploy v1.0.0 to your fiancé
5. **Next week:** Push first update to verify system works

---

**Questions?** Check the scripts:
- `build_installer.ps1` - Full installer creation
- `create_update.ps1` - Update package creation
- `apply_update.ps1` - Update application (runs on user machine)
