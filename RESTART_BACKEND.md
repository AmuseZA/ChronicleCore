# Restart Backend with New Rules API

The Rules Management API has been successfully built into the backend binary.

## To Enable Rules API

**You need to manually restart the backend** since the old process is still running:

### Step 1: Stop Current Backend

In the terminal where `chroniclecore.exe` is running, press **Ctrl+C**

### Step 2: Start New Backend

```bash
cd apps/chroniclecore-core
./chroniclecore.exe
```

### Step 3: Verify Rules API

```bash
curl http://127.0.0.1:8080/api/v1/rules
```

**Expected**: `[]` (empty array, not 404)

## What's New

### 4 New Endpoints Added

1. **GET /api/v1/rules** - List all rules
2. **POST /api/v1/rules** - Create new rule
3. **PUT /api/v1/rules/{id}** - Update existing rule
4. **DELETE /api/v1/rules/{id}** - Delete rule

### Quick Test

After restarting, run this to verify:

```bash
# Should return empty array
curl http://127.0.0.1:8080/api/v1/rules

# Create a test rule (use existing profile_id)
curl -X POST http://127.0.0.1:8080/api/v1/rules \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Rule","priority":10,"match_type":"APP","match_value":"Code.exe","target_profile_id":1}'
```

## Full Testing Guide

See [workflow/rules_api_testing.md](workflow/rules_api_testing.md) for comprehensive testing scenarios.

---

**Status**: ✅ Backend built successfully with Rules API
**Action Required**: Manual restart to load new binary
