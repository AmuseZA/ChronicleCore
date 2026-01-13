# Directive D11: Packaging (MSIX)

**Goal:** Package the binary and UI assets into an installable Windows package.

**Scope:**
- `go-msi` or `wix` or `powershell` build scripts.
- `chroniclecore.exe` + `dist/` folder.

**Inputs:**
- `chroniclecore.exe` (Built binary)
- `internal/assets/dist` (Embedded UI? If embedded, single binary is easier. If not, folder structure).
  - *Recommendation:* Embed `dist` into Go binary using `embed` package. Single EXE is best for local tools.

## 1. Embedding UI
- In `internal/assets/assets.go`:
  ```go
  //go:embed dist/*
  var UIAssets embed.FS
  ```
- Serve from HTTP: `http.FileServer(http.FS(UIAssets))`.

## 2. Windows Installer
- Create `execution/build_msix.ps1` (or Setup.exe via InnoSetup).
- **MSIX:**
  - Requires Manifest.
  - Requires Self-Signing Cert (generated in script).
- **InnoSetup (Alternative):**
  - Simpler for "Check for updates" later.
- **Decision:** Use **InnoSetup** or simple `zip` for MVP if MSIX is too complex for Dev Agent.
  - *Constraint:* Plan said MSIX. Stick to MSIX if possible, but fallback to Inno is acceptable if noted.

## Acceptance Criteria
- [ ] Script produces `ChronicleCore.msix` (or setup.exe).
- [ ] Install places binary in Key path.
- [ ] Uninstall removes it.
