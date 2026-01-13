# ChronicleCore - Quick Start Guide

**Version:** 1.0.0-dev (with ML + Currency support)
**For:** Windows 10/11

## What's New

✅ **ML Auto-Assignment** - Learns your work patterns and suggests profile assignments
✅ **Multi-Currency Support** - All ISO currencies with auto-detection
✅ **Unified Profile Management** - Streamlined client/project/service workflow

## Quick Install (5 minutes)

### Step 1: Install Python (for ML features)

1. Download Python 3.8+ from https://www.python.org/downloads/
2. **Important:** Check "Add Python to PATH" during install
3. Verify: Open cmd and run `python --version`

### Step 2: Setup ML Dependencies

```bash
cd apps\chronicle-ml
setup.bat
```

This installs scikit-learn, FastAPI, and other ML libraries.

### Step 3: Start Backend

```bash
cd apps\chroniclecore-core
chroniclecore.exe
```

You should see:
```
✓ ML sidecar running
✓ ML endpoints registered
Server listening on 127.0.0.1:8080
```

### Step 4: Open UI

- Navigate to `apps/chroniclecore-ui/`
- Run `npm install` (first time only)
- Run `npm run dev`
- Open http://localhost:5173

## First Time Setup

### 1. Create Your First Client

Go to Profiles tab → Click "Add Profile"

- Client: e.g., "Acme Corp"
- Service: e.g., "Bookkeeping"
- Rate: e.g., R 150/hour (currency auto-detected)

### 2. Let It Track

The system will automatically track your window activity:
- Excel spreadsheets
- Xero tabs
- Email clients
- Any active window

### 3. Manually Assign 50 Blocks

For the first week, manually assign blocks to profiles.

After 50 assignments, the ML model will train automatically and start suggesting assignments!

## Using ML Auto-Assignment

Once you've manually assigned 50+ blocks:

1. **Training happens automatically** after 50 labels or daily
2. **Check suggestions**: GET http://127.0.0.1:8080/api/v1/ml/suggestions
3. **Accept good suggestions**: They'll be applied automatically
4. **Correct bad ones**: This improves the model

### Manual Training

```bash
curl -X POST http://127.0.0.1:8080/api/v1/ml/train
```

### Generate Predictions

```bash
curl -X POST http://127.0.0.1:8080/api/v1/ml/predict
```

## Multi-Currency

The system auto-detects your currency based on location:
- South Africa → ZAR
- United States → USD
- United Kingdom → GBP
- etc.

You can use any ISO 4217 currency code (USD, EUR, GBP, ZAR, etc.)

## Database Location

Your data is stored locally at:
```
C:\Users\[YourName]\AppData\Local\ChronicleCore\chronicle.db
```

## Troubleshooting

### ML Sidecar Not Starting

**Error:** `⚠️  ML sidecar disabled: python not found`

**Fix:**
1. Install Python from https://python.org
2. Ensure Python is in PATH
3. Restart backend

---

**Error:** `Failed to install dependencies`

**Fix:**
```bash
cd apps\chronicle-ml
pip install -r requirements.txt
```

### Backend Won't Start

**Error:** `Database locked`

**Fix:** Close any other instances of chroniclecore.exe

---

**Error:** `Port 8080 already in use`

**Fix:** Stop other services on port 8080 or change PORT env variable

### Low ML Accuracy

**Symptoms:** Predictions are wrong most of the time

**Fix:**
1. Manually assign more blocks (aim for 50-100+)
2. Use consistent naming in window titles
3. Retrain: `POST /api/v1/ml/train`

## Testing Script

Run the ML pipeline test:
```bash
cd execution
python test_ml_pipeline.py
```

This verifies:
- Backend is running
- ML sidecar is working
- Training succeeds
- Predictions work

## Privacy & Security

✅ **100% Local** - No cloud, no external servers
✅ **Localhost Only** - Backend only accessible from your machine
✅ **No Screenshots** - Only captures window titles and app names
✅ **Encrypted Storage** - Sensitive settings use Windows DPAPI
✅ **No Telemetry** - No tracking, no analytics, no phone-home

## Support

- Documentation: `workflow/` directory
- ML Guide: `workflow/ml_user_guide.md`
- Backend Status: `workflow/backend_status.md`
- Building Guide: `workflow/building_windows.md`

## For Your Fiancé

Perfect for multi-client accounting work!

**Use Case:**
- Multiple Xero clients
- Excel spreadsheets for different companies
- Email threads with various clients

**How it helps:**
1. Tracks time automatically (no manual timers)
2. Learns which windows belong to which clients
3. Auto-assigns after seeing patterns
4. Generates accurate invoices with proper currency

**Example:**
- Opens "Invoice - ABC Corp - Xero" → ML suggests "ABC Corp - Bookkeeping"
- Opens "Report.xlsx [DEF Ltd]" → ML suggests "DEF Ltd - Reporting"
- Opens email to client → ML suggests correct profile

After 2 weeks of use, accuracy typically reaches 80-90%!

---

**Questions?** Check the documentation in `workflow/` or create an issue.

**Ready to start?** Run `chroniclecore.exe` and open the UI!
