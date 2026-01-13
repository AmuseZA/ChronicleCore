# ChronicleCore - User Testing Quick Start

**Status**: ✅ MVP Ready for Testing
**Version**: 1.0.0-dev

---

## Quick Start (5 Minutes)

### 1. Start the Backend

```bash
cd apps/chroniclecore-core
./chroniclecore.exe
```

**Expected**: Server starts on `http://127.0.0.1:8080`

### 2. Start the UI

```bash
cd apps/chroniclecore-ui
npm run dev
```

**Expected**: UI opens at `http://localhost:5173`

### 3. Begin Tracking

1. Open UI in browser
2. Click "Start Tracking"
3. Use your computer normally for 5+ minutes
4. Check Dashboard to see your activity blocks

---

## First-Time Setup Workflow

### Create Your First Profile

A **profile** = Client + Service + Rate

**Step 1: Create Client**
- Go to Profiles page
- Click "New Client"
- Enter name (e.g., "Acme Corp")
- Save

**Step 2: Create Service**
- Click "New Service"
- Enter name (e.g., "Development" or "Consulting")
- Save

**Step 3: Create Rate**
- Click "New Rate"
- Enter name (e.g., "Standard")
- Enter hourly rate (e.g., $150.00)
- Currency: USD
- Save

**Step 4: Create Profile**
- Click "New Profile"
- Select client, service, rate from dropdowns
- Save

**Result**: You can now assign time blocks to this profile!

---

## Create Rules (Auto-Assignment)

**Rules** automatically assign profiles to your activity.

### Example 1: Match Specific App

```
Name: VS Code Development
Priority: 10
Match Type: APP
Match Value: Code.exe
Target Profile: [Your Dev Profile]
```

**Result**: All VS Code time automatically assigned to dev profile

### Example 2: Match Window Title

```
Name: Client Meetings
Priority: 15
Match Type: KEYWORD
Match Value: meeting
Target Profile: [Your Client Profile]
```

**Result**: Any window with "meeting" in title assigned to client

### Example 3: Match with Regex

```
Name: GitHub Work
Priority: 8
Match Type: TITLE_REGEX
Match Value: .*github\.com.*
Target Profile: [Your Dev Profile]
```

**Result**: Any GitHub activity assigned to dev profile

---

## Review & Assign Time

### Manual Assignment

1. Go to "Needs Review" page
2. See blocks without profiles
3. Click block → Select profile → Save
4. Lock block to prevent auto-reassignment

### Bulk Review

- Filter by date
- Filter by unassigned
- Assign multiple blocks at once

---

## Export Your Time

1. Go to Export page
2. Select date range (today, this week, custom)
3. Choose rounding:
   - **6 minutes** - Standard billing increment
   - **15 minutes** - Quarter-hour billing
   - **No rounding** - Exact time
4. Set minimum billable (optional)
   - e.g., 30 minutes minimum
5. Download CSV

**CSV Columns**:
- Client, Project, Service
- Date, Start Time, End Time
- Hours (Rounded), Hours (Actual)
- Rate, Amount
- Description, Confidence

---

## Testing Scenarios

### Scenario 1: Freelancer (Single Client)

**Goal**: Track billable hours for one client

1. Create 1 client, 1 service, 1 rate, 1 profile
2. Create rule: Your main app → Profile
3. Work for 2 hours
4. Export CSV
5. **Verify**: Hours and amount correct

### Scenario 2: Agency (Multiple Clients)

**Goal**: Track time across 3 different clients

1. Create 3 profiles (3 clients)
2. Create 3 rules (different apps/keywords per client)
3. Work on each client's project
4. Review dashboard - time distributed correctly?
5. Export filtered by client

### Scenario 3: Manual Review

**Goal**: Handle activities without rules

1. Work without rules configured
2. Go to "Needs Review"
3. Manually assign blocks
4. Lock important blocks
5. Export includes manual assignments

---

## What to Test & Report

### Functionality
- [ ] Tracking starts/pauses/stops correctly
- [ ] Blocks appear after 5 minutes
- [ ] Can create profiles
- [ ] Can create rules
- [ ] Rules auto-assign correctly
- [ ] Can manually reassign
- [ ] Export generates CSV
- [ ] CSV opens in Excel

### Accuracy
- [ ] Time blocks match actual work
- [ ] Idle detection works (5 min threshold)
- [ ] Rounding calculation correct
- [ ] No missed activities

### Usability
- [ ] UI is intuitive
- [ ] Dashboard is clear
- [ ] Profile creation is easy
- [ ] Rules are understandable
- [ ] Export is straightforward

### Performance
- [ ] No slowness or lag
- [ ] Backend responsive
- [ ] UI loads quickly
- [ ] CSV exports fast

---

## Common Issues & Solutions

### Backend Won't Start

**Problem**: `chroniclecore.exe` fails to start
**Solution**:
1. Check port 8080 is not in use
2. Run from correct directory
3. Check database path permissions

### Rules Not Working

**Problem**: Blocks not auto-assigned
**Solution**:
1. Wait 5 minutes for rollup
2. Check rule priority (higher = first)
3. Check match_value is exact (for APP type)
4. Try KEYWORD instead of APP for broader matching

### No Blocks Appearing

**Problem**: Dashboard empty after tracking
**Solution**:
1. Wait 5+ minutes (rollup interval)
2. Check tracking is actually started
3. Verify you're using applications (not idle)
4. Check backend logs for errors

### Export CSV Empty

**Problem**: Export returns no data
**Solution**:
1. Check date range includes your activity
2. Verify blocks exist for that period
3. Check profile filter (try "All")

---

## Testing Scripts (Optional)

Run automated tests to verify everything works:

```bash
# Validate database schema
python execution/validate_schema.py

# Test API endpoints
python execution/validate_api_contract.py

# Test export logic
python execution/verify_exports.py

# Audit privacy settings
python execution/privacy_checks.py

# Verify UI builds
python execution/ui_build_check.py
```

---

## Feedback to Collect

### What Works Well?
- Easiest features to use
- Clearest UI elements
- Most useful functionality

### What's Confusing?
- Unclear workflows
- Terminology issues
- Missing instructions

### What's Missing?
- Features you expected but didn't find
- Integrations needed (other tools, formats)
- Workflow improvements

### What's Broken?
- Bugs encountered
- Incorrect calculations
- UI issues

---

## Support & Documentation

**Full Documentation**:
- [README.md](README.md) - Project overview
- [API_REFERENCE.md](API_REFERENCE.md) - API docs
- [workflow/backend_status.md](workflow/backend_status.md) - Setup
- [workflow/mvp_completion_report.md](workflow/mvp_completion_report.md) - Complete status

**Testing Guides**:
- [workflow/rules_api_testing.md](workflow/rules_api_testing.md) - Rules API
- [execution/verify_exports.py](execution/verify_exports.py) - Export validation

**Quick Restart**:
- Stop backend (Ctrl+C)
- Run: `./chroniclecore.exe`

---

## Data Location

Your data is stored locally:
- **Database**: `%LOCALAPPDATA%\ChronicleCore\chronicle.db`
- **Logs**: Console output only (not persisted)
- **Exports**: Downloads folder (user-initiated)

**Privacy**: No cloud sync, no external services, localhost-only.

---

## Ready to Test!

1. Start backend ✅
2. Start UI ✅
3. Create profile ✅
4. Create rule (optional) ✅
5. Start tracking ✅
6. Work for 5+ minutes ✅
7. Review dashboard ✅
8. Export CSV ✅

**Questions?** Check [workflow/mvp_completion_report.md](workflow/mvp_completion_report.md)

**Good luck with testing!** 🎯
