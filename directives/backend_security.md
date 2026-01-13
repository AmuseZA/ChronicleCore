# Directive D02: Backend Security & Networking

**Goal:** Implement the HTTP server core with strict security defaults (Localhost binding, Headers, DPAPI).

**Scope:**
- `chroniclecore/internal/api` (Server setup)
- `chroniclecore/internal/security` (DPAPI wrapper)
- `execution/privacy_checks.py` (Validation)

**Inputs:**
- `apps/chroniclecore-core` (initialized in D01)

**Outputs:**
- HTTP Server running on `127.0.0.1`.
- CORS middleware restricted.
- Helper function for DPAPI encryption.

## 1. HTTP Server (Go)
- Use standard `net/http` or a lightweight router (Chi/Echo).
- **CRITICAL:** Listen Address must be hardcoded to `127.0.0.1:0` (random port) or specific port (e.g. 8090) if needed for dev, but *never* `0.0.0.0`.
- **Headers:**
  - Middleware to reject requests where `Host` header is not `127.0.0.1` or `localhost`. (Prevents DNS rebinding).
  - CORS: AllowOrigin `http://127.0.0.1:[FrontendPort]`.

## 2. DPAPI Encryption (Windows)
- Implement `internal/security/dpapi_windows.go`.
- Use `syscall` or `golang.org/x/sys/windows` to call `CryptProtectData`.
- Create interface:
  ```go
  func Encrypt(data []byte) ([]byte, error)
  func Decrypt(data []byte) ([]byte, error)
  ```
- Storage of secrets (Extension Token, etc.) in `sqlite` must use this.

## 3. Privacy Checks (Validation)
- Create `execution/privacy_checks.py`:
  - Python script that `imports socket`.
  - Attempts to connect to the app port from an "external" IP (mocked) or checks config files.
  - Verifies that `spec/defaults.json` (if exists) has `capture_urls: false`.

## Acceptance Criteria
- [ ] Server refuses connections on LAN IP (e.g. 192.168.x.x).
- [ ] `curl -H "Host: evil.com" http://127.0.0.1:PORT` returns 403 Forbidden.
- [ ] DPAPI Encrypt -> Decrypt roundtrip works on Windows.
