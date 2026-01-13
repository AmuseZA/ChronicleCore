# ChronicleCore API Reference

Complete reference for all API endpoints in ChronicleCore v1.0.

**Base URL**: `http://127.0.0.1:8080`

---

## Table of Contents

1. [System](#system)
2. [Tracking Control](#tracking-control)
3. [Blocks](#blocks)
4. [Profiles](#profiles)
5. [Rules](#rules)
6. [Exports](#exports)
7. [Common Patterns](#common-patterns)
8. [Error Codes](#error-codes)

---

## System

### Health Check

**GET** `/health`

Returns server health and version information.

**Response**:
```json
{
  "status": "ok",
  "version": "1.0.0-dev",
  "uptime_seconds": 3600
}
```

**Status Codes**:
- `200 OK` - Server is healthy

---

## Tracking Control

### Get Tracking Status

**GET** `/api/v1/tracking/status`

Returns current tracking state.

**Response**:
```json
{
  "state": "ACTIVE",
  "last_active_at": "2026-01-09T15:30:00Z",
  "idle_seconds": 0,
  "current_window": {
    "app_name": "EXCEL.EXE",
    "title": "Budget 2026.xlsx"
  }
}
```

**States**: `STOPPED`, `ACTIVE`, `PAUSED`

### Start Tracking

**POST** `/api/v1/tracking/start`

Begins activity capture.

**Response**: Same as status endpoint

**Status Codes**:
- `200 OK` - Tracking started
- `400 Bad Request` - Already running

### Pause Tracking

**POST** `/api/v1/tracking/pause`

Temporarily pauses capture.

**Status Codes**:
- `200 OK` - Tracking paused
- `400 Bad Request` - Not active

### Resume Tracking

**POST** `/api/v1/tracking/resume`

Resumes from paused state.

**Status Codes**:
- `200 OK` - Tracking resumed
- `400 Bad Request` - Not paused

### Stop Tracking

**POST** `/api/v1/tracking/stop`

Fully stops tracking.

**Status Codes**:
- `200 OK` - Tracking stopped

---

## Blocks

### List Blocks

**GET** `/api/v1/blocks`

Query time blocks with filters.

**Query Parameters**:

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `date` | string | Filter by date (YYYY-MM-DD) | Today |
| `start_date` | string | Date range start (YYYY-MM-DD) | - |
| `end_date` | string | Date range end (YYYY-MM-DD) | - |
| `profile_id` | integer | Filter by profile | - |
| `unassigned` | boolean | Show only unassigned | false |
| `needs_review` | boolean | Show LOW confidence or unassigned | false |
| `limit` | integer | Max results (1-1000) | 100 |

**Example Requests**:
```bash
# Today's blocks
GET /api/v1/blocks

# Specific date
GET /api/v1/blocks?date=2026-01-08

# Date range
GET /api/v1/blocks?start_date=2026-01-01&end_date=2026-01-31

# Unassigned only
GET /api/v1/blocks?unassigned=true

# Needs review
GET /api/v1/blocks?needs_review=true

# By profile
GET /api/v1/blocks?profile_id=1

# With limit
GET /api/v1/blocks?limit=50
```

**Response**:
```json
[
  {
    "block_id": 1,
    "ts_start": "2026-01-09T14:30:00Z",
    "ts_end": "2026-01-09T15:15:00Z",
    "duration_minutes": 45.0,
    "duration_hours": 0.75,
    "primary_app_name": "EXCEL.EXE",
    "primary_domain": null,
    "title_summary": "Budget 2026.xlsx",
    "profile_id": 1,
    "client_name": "Acme Corp",
    "project_name": null,
    "service_name": "Bookkeeping",
    "confidence": "HIGH",
    "billable": true,
    "locked": false,
    "notes": null,
    "description": "Excel - Budget 2026.xlsx",
    "created_at": "2026-01-09T15:15:05Z",
    "updated_at": "2026-01-09T15:15:05Z"
  }
]
```

**Status Codes**:
- `200 OK` - Success (returns empty array if no matches)
- `400 Bad Request` - Invalid date format or parameters

### Reassign Block

**POST** `/api/v1/blocks/{id}/reassign`

Assign or unassign a profile to a block.

**Request Body**:
```json
{
  "profile_id": 1,           // or null to unassign
  "confidence": "HIGH"       // Optional: HIGH, MEDIUM, LOW
}
```

**Response**: Updated block (same format as list)

**Status Codes**:
- `200 OK` - Reassignment successful
- `400 Bad Request` - Invalid JSON or confidence value
- `404 Not Found` - Block not found

**Examples**:
```bash
# Assign to profile
curl -X POST http://127.0.0.1:8080/api/v1/blocks/1/reassign \
  -H "Content-Type: application/json" \
  -d '{"profile_id":1}'

# Assign with specific confidence
curl -X POST http://127.0.0.1:8080/api/v1/blocks/1/reassign \
  -H "Content-Type: application/json" \
  -d '{"profile_id":1,"confidence":"MEDIUM"}'

# Unassign
curl -X POST http://127.0.0.1:8080/api/v1/blocks/1/reassign \
  -H "Content-Type: application/json" \
  -d '{"profile_id":null,"confidence":"LOW"}'
```

### Lock/Unlock Block

**POST** `/api/v1/blocks/{id}/lock`

Lock or unlock a block to prevent/allow auto-reassignment.

**Request Body**:
```json
{
  "locked": true   // or false to unlock
}
```

**Response**: Updated block (same format as list)

**Status Codes**:
- `200 OK` - Lock status changed
- `400 Bad Request` - Invalid JSON
- `404 Not Found` - Block not found

**Examples**:
```bash
# Lock block
curl -X POST http://127.0.0.1:8080/api/v1/blocks/1/lock \
  -H "Content-Type: application/json" \
  -d '{"locked":true}'

# Unlock block
curl -X POST http://127.0.0.1:8080/api/v1/blocks/1/lock \
  -H "Content-Type: application/json" \
  -d '{"locked":false}'
```

---

## Profiles

### Clients

#### List Clients

**GET** `/api/v1/clients`

**Query Parameters**:
- `active_only` - boolean (default: true)

**Response**:
```json
[
  {
    "client_id": 1,
    "name": "Acme Corp",
    "is_active": true,
    "created_at": "2026-01-09T10:00:00Z",
    "updated_at": "2026-01-09T10:00:00Z"
  }
]
```

#### Create Client

**POST** `/api/v1/clients/create`

**Request Body**:
```json
{
  "name": "Acme Corp"
}
```

**Response**: Created client

**Status Codes**:
- `201 Created` - Client created
- `400 Bad Request` - Name is empty
- `409 Conflict` - Name already exists

### Services

#### List Services

**GET** `/api/v1/services`

**Response**:
```json
[
  {
    "service_id": 1,
    "name": "Bookkeeping",
    "is_active": true
  }
]
```

#### Create Service

**POST** `/api/v1/services/create`

**Request Body**:
```json
{
  "name": "Bookkeeping"
}
```

**Status Codes**:
- `201 Created` - Service created
- `400 Bad Request` - Name is empty
- `409 Conflict` - Name already exists

### Rates

#### List Rates

**GET** `/api/v1/rates`

**Response**:
```json
[
  {
    "rate_id": 1,
    "name": "Standard",
    "currency": "ZAR",
    "hourly_amount": 150.00,
    "hourly_minor_units": 15000,
    "effective_from": null,
    "effective_to": null,
    "is_active": true
  }
]
```

**Note**: `hourly_amount` is in major units (dollars/rands), `hourly_minor_units` is stored value (cents).

#### Create Rate

**POST** `/api/v1/rates/create`

**Request Body**:
```json
{
  "name": "Standard",
  "currency": "ZAR",
  "hourly_amount": 150.00,
  "effective_from": "2026-01-01",  // Optional
  "effective_to": null              // Optional
}
```

**Status Codes**:
- `201 Created` - Rate created
- `400 Bad Request` - Invalid amount or missing fields

### Profiles

#### List Profiles

**GET** `/api/v1/profiles`

**Response**:
```json
[
  {
    "profile_id": 1,
    "client_name": "Acme Corp",
    "project_name": null,
    "service_name": "Bookkeeping",
    "rate_name": "Standard",
    "rate_amount": 150.00,
    "currency": "ZAR",
    "is_active": true
  }
]
```

#### Create Profile

**POST** `/api/v1/profiles`

**Request Body**:
```json
{
  "client_id": 1,
  "project_id": null,     // Optional
  "service_id": 1,
  "rate_id": 1,
  "name": "Custom Name"   // Optional display name
}
```

**Status Codes**:
- `201 Created` - Profile created
- `400 Bad Request` - Missing required fields or invalid foreign keys

#### Delete Profile

**DELETE** `/api/v1/profiles/{id}`

Soft deletes (sets is_active=0).

**Status Codes**:
- `204 No Content` - Profile deleted
- `404 Not Found` - Profile not found

---

## Rules

### List Rules

**GET** `/api/v1/rules`

Query rules with optional filtering.

**Query Parameters**:

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `enabled` | boolean | Filter by enabled status | true |

**Example Requests**:
```bash
# Get all enabled rules (default)
GET /api/v1/rules

# Get all rules (enabled and disabled)
GET /api/v1/rules?enabled=false
```

**Response**:
```json
[
  {
    "rule_id": 1,
    "name": "VS Code Development",
    "priority": 10,
    "match_type": "APP",
    "match_value": "Code.exe",
    "target_profile_id": 1,
    "confidence_boost": 10,
    "enabled": true,
    "client_name": "Acme Corp",
    "service_name": "Development",
    "created_at": "2026-01-09T10:00:00Z",
    "updated_at": "2026-01-09T10:00:00Z"
  }
]
```

**Status Codes**:
- `200 OK` - Success (returns empty array if no matches)

### Create Rule

**POST** `/api/v1/rules`

Create a new profile assignment rule.

**Request Body**:
```json
{
  "name": "VS Code Development",
  "priority": 10,
  "match_type": "APP",
  "match_value": "Code.exe",
  "target_profile_id": 1,
  "target_service_id": null,
  "confidence_boost": 10,
  "enabled": true
}
```

**Field Descriptions**:
- `name` (required): Rule name
- `priority` (required): Higher values matched first
- `match_type` (required): One of: `APP`, `DOMAIN`, `TITLE_REGEX`, `KEYWORD`, `COMPOSITE`
- `match_value` (required): Pattern to match (exact string, regex, or JSON)
- `target_profile_id` (required): Profile to assign when matched
- `target_service_id` (optional): Override profile's service
- `confidence_boost` (optional): Confidence adjustment (-100 to +100), default: 0
- `enabled` (optional): Enable rule, default: true

**Match Types**:
- `APP`: Exact match on application name (e.g., "Code.exe")
- `DOMAIN`: Exact match on domain (e.g., "github.com")
- `TITLE_REGEX`: Regular expression on window title (e.g., ".*GitHub.*")
- `KEYWORD`: Substring match on title (case-insensitive)
- `COMPOSITE`: JSON with multiple conditions

**Response**: Created rule (same format as list)

**Status Codes**:
- `201 Created` - Rule created
- `400 Bad Request` - Invalid match_type, bad regex, or missing fields
- `400 Bad Request` - target_profile_id doesn't exist

**Examples**:
```bash
# Match specific app
curl -X POST http://127.0.0.1:8080/api/v1/rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "VS Code Work",
    "priority": 10,
    "match_type": "APP",
    "match_value": "Code.exe",
    "target_profile_id": 1
  }'

# Match with regex
curl -X POST http://127.0.0.1:8080/api/v1/rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "GitHub PRs",
    "priority": 8,
    "match_type": "TITLE_REGEX",
    "match_value": ".*GitHub.*Pull Request.*",
    "target_profile_id": 1,
    "confidence_boost": 5
  }'

# Match with keyword
curl -X POST http://127.0.0.1:8080/api/v1/rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Meeting Notes",
    "priority": 5,
    "match_type": "KEYWORD",
    "match_value": "meeting",
    "target_profile_id": 2
  }'
```

### Update Rule

**PUT** `/api/v1/rules/{id}`

Update an existing rule (partial updates supported).

**Request Body** (all fields optional):
```json
{
  "name": "VS Code Development (Updated)",
  "priority": 15,
  "match_type": "APP",
  "match_value": "Code.exe",
  "target_profile_id": 1,
  "confidence_boost": 15,
  "enabled": false
}
```

**Response**: Updated rule (same format as list)

**Status Codes**:
- `200 OK` - Rule updated
- `400 Bad Request` - Invalid fields or no fields provided
- `404 Not Found` - Rule not found

**Examples**:
```bash
# Change priority
curl -X PUT http://127.0.0.1:8080/api/v1/rules/1 \
  -H "Content-Type: application/json" \
  -d '{"priority": 20}'

# Disable rule
curl -X PUT http://127.0.0.1:8080/api/v1/rules/2 \
  -H "Content-Type: application/json" \
  -d '{"enabled": false}'

# Update multiple fields
curl -X PUT http://127.0.0.1:8080/api/v1/rules/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Name",
    "priority": 25,
    "confidence_boost": 20
  }'
```

### Delete Rule

**DELETE** `/api/v1/rules/{id}`

Permanently delete a rule.

**Status Codes**:
- `204 No Content` - Rule deleted
- `404 Not Found` - Rule not found

**Example**:
```bash
curl -X DELETE http://127.0.0.1:8080/api/v1/rules/1
```

---

## Exports

### Export Invoice Lines

**POST** `/api/v1/export/invoice-lines`

Generates CSV export with rounding and billing rules.

**Request Body**:
```json
{
  "start_date": "2026-01-01",
  "end_date": "2026-01-31",
  "profile_ids": [1, 2],              // Optional filter
  "rounding_minutes": 6,               // 6 or 15
  "minimum_billable_minutes": 0        // Optional
}
```

**Response**: CSV file

**CSV Format**:
```csv
Client,Project,Service,Date,Start Time,End Time,Hours (Rounded),Hours (Actual),Rate,Currency,Amount,Description,Confidence
Acme Corp,,Bookkeeping,2026-01-15,09:00,10:30,1.60,1.50,150.00,ZAR,240.00,"Excel - Budget 2026.xlsx",HIGH
```

**Status Codes**:
- `200 OK` - CSV generated
- `400 Bad Request` - Invalid dates or rounding value

**Example**:
```bash
curl -X POST http://127.0.0.1:8080/api/v1/export/invoice-lines \
  -H "Content-Type: application/json" \
  -d '{
    "start_date": "2026-01-01",
    "end_date": "2026-01-31",
    "rounding_minutes": 6
  }' \
  > invoice.csv
```

---

## Common Patterns

### Date Formats

All dates use ISO-8601 format:
- **Input (query params)**: `YYYY-MM-DD` (e.g., `2026-01-09`)
- **Output (JSON)**: `YYYY-MM-DDTHH:MM:SSZ` (e.g., `2026-01-09T15:30:00Z`)

### Pagination

Most list endpoints support `limit` parameter:
```
GET /api/v1/blocks?limit=50
```

Default limits:
- Blocks: 100 (max: 1000)
- Others: No default limit (returns all)

### Filtering Unassigned

Two approaches:
```bash
# Only unassigned (profile_id IS NULL)
GET /api/v1/blocks?unassigned=true

# Unassigned OR low confidence
GET /api/v1/blocks?needs_review=true
```

### Null Values

Nullable fields may be:
- `null` in JSON
- Omitted from response (for optional fields)

Example:
```json
{
  "profile_id": null,      // Explicitly null (unassigned)
  "notes": null,           // Explicitly null (no notes)
  "project_name": null     // Explicitly null (no project)
}
```

---

## Error Codes

### Standard HTTP Status Codes

| Code | Meaning | Common Causes |
|------|---------|---------------|
| 200 | OK | Success |
| 201 | Created | Resource created |
| 204 | No Content | Deletion successful |
| 400 | Bad Request | Invalid JSON, missing fields, bad format |
| 404 | Not Found | Resource doesn't exist |
| 405 | Method Not Allowed | Wrong HTTP method |
| 409 | Conflict | Unique constraint violation |
| 500 | Internal Server Error | Database error, server bug |

### Error Response Format

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Invalid date format (use YYYY-MM-DD)",
    "details": {}
  }
}
```

**Error Codes**:
- `INVALID_REQUEST` - Bad input (400)
- `NOT_FOUND` - Resource not found (404)
- `CONFLICT` - Unique constraint violated (409)
- `INTERNAL_ERROR` - Server error (500)

### Common Errors

#### Invalid Date
```bash
GET /api/v1/blocks?date=2026-13-01
# 400: Invalid date format (use YYYY-MM-DD)
```

#### Block Not Found
```bash
POST /api/v1/blocks/99999/reassign
# 404: Block not found
```

#### Invalid Confidence
```bash
POST /api/v1/blocks/1/reassign
Body: {"profile_id":1,"confidence":"INVALID"}
# 400: confidence must be HIGH, MEDIUM, or LOW
```

#### Duplicate Name
```bash
POST /api/v1/clients/create
Body: {"name":"Existing Client"}
# 409: Client name already exists
```

---

## CORS Policy

**Allowed Origins**: `http://127.0.0.1`, `http://localhost`

**Allowed Methods**: `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`

**Allowed Headers**: `Content-Type`

**Preflight Requests**: Supported via `OPTIONS` method

---

## Rate Limiting

**Current Status**: None

API is localhost-only and not rate-limited.

---

## Authentication

**Current Status**: None

API is localhost-only and does not require authentication.

**Future**: May add token-based auth for multi-user scenarios (Phase 4+).

---

## Versioning

**Current Version**: v1

API version is in the URL path: `/api/v1/...`

Breaking changes will increment the version: `/api/v2/...`

---

## Changelog

### v1.0.0-dev (2026-01-09)

**Added**:
- System: Health check
- Tracking: Status, start/pause/resume/stop
- Blocks: List with filters, reassign, lock
- Profiles: Full CRUD for clients, services, rates, profiles
- Exports: Invoice lines CSV with rounding

---

## Support

**Documentation**:
- [BUILDING.md](BUILDING.md) - Build and installation
- [TESTING_BLOCKS_API.md](TESTING_BLOCKS_API.md) - Testing guide
- [workflow/phase1_completion.md](workflow/phase1_completion.md) - Implementation details

**Issues**: GitHub repository (to be added)

---

**Last Updated**: 2026-01-09
**API Version**: 1.0.0-dev
