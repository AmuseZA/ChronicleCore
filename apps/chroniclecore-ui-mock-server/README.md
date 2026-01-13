# ChronicleCore Mock API Server

A lightweight Node.js mock server that implements all ChronicleCore API endpoints for UI development when the Go backend is unavailable.

## Purpose

This mock server allows UI development to proceed without requiring:
- Go installation
- Backend compilation
- SQLite database setup

It provides realistic mock data and stateful responses for all 19 API endpoints.

## Quick Start

```bash
# Install dependencies
npm install

# Start server (runs on http://127.0.0.1:8080)
npm start
```

## Mock Data

The server includes:
- **3 blocks**: Excel (assigned to Acme Corp), Edge (unassigned), Code (assigned to Internal)
- **2 clients**: Acme Corp, Internal
- **3 services**: Bookkeeping, Development, Consulting
- **2 rates**: Standard ($150/hr), Internal ($0/hr)
- **2 profiles**: Acme Corp + Bookkeeping, Internal + Development

## Stateful Behavior

- Tracking state persists across requests (STOPPED → ACTIVE → PAUSED)
- Block reassignment updates mock data in memory
- Profile/client/service creation adds to mock arrays

## API Endpoints

All endpoints from [API_REFERENCE.md](../../../API_REFERENCE.md) are implemented:

### System
- `GET /health`

### Tracking
- `GET /api/v1/tracking/status`
- `POST /api/v1/tracking/{start,pause,resume,stop}`

### Blocks
- `GET /api/v1/blocks?unassigned=true&needs_review=true&profile_id=1`
- `POST /api/v1/blocks/:id/reassign`
- `POST /api/v1/blocks/:id/lock`

### Profiles
- `GET /api/v1/clients`
- `POST /api/v1/clients/create`
- `GET /api/v1/services`
- `POST /api/v1/services/create`
- `GET /api/v1/rates`
- `POST /api/v1/rates/create`
- `GET /api/v1/profiles`
- `POST /api/v1/profiles`
- `DELETE /api/v1/profiles/:id`

### Exports
- `POST /api/v1/export/invoice-lines`

## Testing

```bash
# Health check
curl http://127.0.0.1:8080/health

# Get tracking status
curl http://127.0.0.1:8080/api/v1/tracking/status

# Start tracking
curl -X POST http://127.0.0.1:8080/api/v1/tracking/start

# Get all blocks
curl http://127.0.0.1:8080/api/v1/blocks

# Get unassigned blocks
curl "http://127.0.0.1:8080/api/v1/blocks?unassigned=true"

# Reassign block
curl -X POST http://127.0.0.1:8080/api/v1/blocks/2/reassign \
  -H "Content-Type: application/json" \
  -d '{"profile_id":1}'

# Get profiles
curl http://127.0.0.1:8080/api/v1/profiles
```

## Switching to Real Backend

When the Go backend is available:

1. Stop mock server (Ctrl+C)
2. Build and start Go backend:
   ```bash
   cd ../../chroniclecore-core
   go build -o chroniclecore.exe ./cmd/server
   ./chroniclecore.exe
   ```
3. UI will automatically connect to the same URL

No code changes needed in UI - both servers run on `http://127.0.0.1:8080`.

## Limitations

- **No persistence**: Data resets on server restart
- **Simplified filtering**: Date range filtering on blocks is not fully implemented
- **No validation**: Foreign key validation is basic
- **Static CSV**: Export endpoint returns fixed CSV data

These limitations don't affect UI development workflow.

## CORS

Configured to accept requests from:
- `http://localhost:5173` (Vite default)
- `http://127.0.0.1:5173`
- `http://localhost:3000`

Add more origins in `server.js` if needed.
