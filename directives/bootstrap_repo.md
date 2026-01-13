# Directive D01: Bootstrap Repository and Tooling

**Goal:** Initialize the project repositories for Go (Backend) and Svelte (Frontend), and establish the `execution/` tooling baseline for schema validation.

**Scope:**
- `apps/chroniclecore-core` (Go module `chroniclecore`)
- `apps/chroniclecore-ui` (Svelte Kit or Vite+Svelte, TypeScript)
- `execution/validate_schema.py` (Tooling)
- `go.mod` and `package.json` initialization.

**Inputs:**
- `spec/schema.sql` (Canonical Schema)

**Outputs:**
- Working Go module (builds `cmd/server/main.go` hello world).
- Working Svelte app (builds `dist/`).
- `execution/validate_schema.py` passes and confirms schema hash.

## 1. Backend Bootstrap (Go)
1. Initialize Go module:
   ```bash
   cd apps/chroniclecore-core
   go mod init chroniclecore
   ```
2. Create entry point `cmd/server/main.go`:
   - Must print "ChronicleCore v0.1.0 (Bootstrap)".
   - Must NOT bind port yet (just print and exit is fine for bootstrap, or bind 127.0.0.1:8080 hello world).
   - Use `github.com/mattn/go-sqlite3` (CGO required) or `modernc.org/sqlite` (Pure Go). **Decision: Use `github.com/mattn/go-sqlite3`** if CGO is acceptable on Windows (standard), else `modernc.org/sqlite`. *Recommendation: `modernc.org/sqlite` avoids CGO requirements which simplifies Windows builds significantly.*
   - Import `chroniclecore/internal/store`.

## 2. Frontend Bootstrap (Svelte)
1. Initialize Vite + Svelte (TypeScript):
   ```bash
   cd apps/chroniclecore-ui
   npm create vite@latest . -- --template svelte-ts
   npm install
   npm install -D tailwindcss postcss autoprefixer
   npx tailwindcss init -p
   ```
2. Configure Tailwind to match `../workflow/ui_style_guide.md` colors (Slate-50, etc).
3. Ensure `npm run build` outputs to `dist/`.

## 3. Execution Tooling (Python)
1. Create `execution/requirements.txt`:
   ```
   pytest
   requests
   jsonschema
   ```
2. Create `execution/validate_schema.py`:
   - Connects to `:memory:` SQLite DB.
   - Reads `spec/schema.sql` and executes `executescript()`.
   - Checks for `raw_event`, `block`, `profile` tables.
   - Prints "Schema Validation: PASS".

## Acceptance Criteria
- [ ] `go build ./cmd/server` succeeds in `apps/chroniclecore-core`.
- [ ] `npm run build` succeeds in `apps/chroniclecore-ui`.
- [ ] `python execution/validate_schema.py` prints PASS.
