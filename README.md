# ChronicleCore

**Local-first Windows time tracking with deterministic correctness.**

ChronicleCore is a privacy-focused desktop application for tracking work activity, managing client profiles, and generating invoice-ready exports. Built on a 3-layer architecture (Directive → Orchestration → Execution), it prioritizes deterministic business logic over AI complexity.

## Core Principles

- **Local-first**: All data stays on your machine. No cloud sync, no remote servers.
- **Privacy by default**: Captures window titles only (no URLs, screenshots, or keystrokes).
- **Deterministic logic**: Rules-based assignment with confidence scoring.
- **Windows-only**: Single Go binary targeting Windows desktop.
- **Localhost-only API**: Server binds to 127.0.0.1 only for security.

## Architecture

### 3-Layer Model

```
┌─────────────────────────────────────────────┐
│ Layer 1: Directives (What to do)           │
│ - SOPs in directives/                       │
│ - Feature-scoped, explicit acceptance       │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│ Layer 2: Orchestration (Decision making)   │
│ - LLM routes tasks to scripts               │
│ - Enforces constraints & validates outputs  │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│ Layer 3: Execution (Doing the work)        │
│ - Deterministic scripts in execution/      │
│ - Schema validation, contract tests, etc.  │
└─────────────────────────────────────────────┘
```

### Technology Stack

- **Backend**: Go 1.22 (single binary: `chroniclecore.exe`)
- **Database**: SQLite with WAL mode
- **Frontend**: Svelte + Vite + Tailwind (embedded in binary)
- **Testing**: Python scripts for validation & fixture generation

## Project Structure

```
ChronicleCore/
├── spec/                      # Canonical specifications
│   ├── schema.sql             # SQLite DDL (frozen contract)
│   └── api_contract.yaml      # OpenAPI 3.0 spec
├── directives/                # Layer 1: Feature SOPs
│   ├── bootstrap_repo.md
│   ├── windows_activity_capture.md
│   └── ...
├── execution/                 # Layer 3: Validation scripts
│   ├── validate_schema.py
│   ├── validate_api_contract.py
│   └── generate_fixtures.py
├── apps/
│   ├── chroniclecore-core/    # Go backend
│   │   ├── cmd/server/
│   │   └── internal/
│   └── chroniclecore-ui/      # Svelte frontend
├── workflow/                  # Build plans & milestones
└── .tmp/                      # Generated test data (gitignored)
```

## Phase 0: Bootstrap ✅ COMPLETE

**Goal**: Establish foundational contracts and validation tooling.

### Deliverables

✅ **Canonical Schema** ([spec/schema.sql](spec/schema.sql))
   - 14 tables (install, settings, clients, profiles, rules, events, blocks, audit)
   - Foreign key constraints enforced
   - Triggers for updated_at timestamps
   - Views for duration calculations
   - **Validated**: `python execution/validate_schema.py`

✅ **API Contract** ([spec/api_contract.yaml](spec/api_contract.yaml))
   - OpenAPI 3.0 specification
   - System, Tracking, Blocks, Profiles, Exports endpoints
   - Localhost-only (127.0.0.1) binding requirement documented
   - **Validator**: `python execution/validate_api_contract.py`

✅ **Validation Scripts**
   - `validate_schema.py` - Confirms DDL correctness, foreign keys, objects
   - `validate_api_contract.py` - Integration tests against live API
   - `generate_fixtures.py` - Creates realistic 90-day test datasets

✅ **Backend Scaffolding**
   - Go module initialized ([apps/chroniclecore-core/go.mod](apps/chroniclecore-core/go.mod))
   - HTTP server stub with localhost-only binding ([cmd/server/main.go](apps/chroniclecore-core/cmd/server/main.go))
   - All API endpoints return stub responses (200 OK)
   - CORS restricted to localhost origins

### Quick Start (Phase 0 Validation)

```bash
# 1. Validate schema
python execution/validate_schema.py

# 2. Generate test fixtures (30 days)
python execution/generate_fixtures.py --days 30

# 3. Build and run server (requires Go 1.22+)
cd apps/chroniclecore-core
go run cmd/server/main.go

# 4. Validate API contract (in separate terminal)
python execution/validate_api_contract.py
```

## Next Steps: Phase 1

Phase 1 will implement the core tracking loop:

- Windows activity capture (Win32 API: GetForegroundWindow)
- Idle detection
- Raw event storage with 14-day retention
- Block aggregation (5-minute rollup cadence)
- Rules engine (process + title regex matching)
- Profile CRUD
- Deterministic description templates
- Svelte dashboard (Today view)
- CSV export with rounding/minimums

See [workflow/v1_build_plan.md](workflow/v1_build_plan.md) for full roadmap.

## Security & Privacy

### Binding Policy
- Server MUST bind to `127.0.0.1` only (enforced in code)
- No remote access permitted
- CORS restricted to localhost origins

### Data Capture Policy (MVP)
- ✅ Window titles (with redaction support)
- ✅ Process names (e.g., EXCEL.EXE)
- ✅ Idle detection
- ❌ Full URLs (deferred to Phase 2 with browser extension)
- ❌ Screenshots (prohibited)
- ❌ Keystrokes (prohibited)

### Secrets Management
- Sensitive config encrypted at rest via Windows DPAPI
- No cloud API keys stored in desktop client
- Browser extension uses per-install secret token

### Retention
- Raw events: 14 days (configurable)
- Blocks: indefinite (aggregated, no sensitive detail)
- Rollup job runs every 5 minutes to purge old raw events

## Development Workflow

### Self-Annealing Protocol

When errors occur:
1. Read the stack trace
2. Fix deterministic tools/tests first
3. Re-run validation scripts
4. Update the relevant directive with failure mode + fix
5. System gets stronger

### Definition of Done (DoD)

A change is "done" only if:
- ✅ Build succeeds (Go binary compiles)
- ✅ All validation scripts pass
- ✅ API contract tests pass
- ✅ Schema validation passes
- ✅ Security checks pass (localhost binding, DPAPI, no sensitive capture)
- ✅ Storage remains stable (retention + VACUUM tested)

## License

[To be determined - proprietary/commercial expected]

## Contact

ChronicleCore Project
Built with the 3-layer DOE/SOP methodology.

---

**Status**: Phase 0 Complete ✅ | Phase 1 Ready to Begin
