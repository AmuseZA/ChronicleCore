# Directive D03: Windows Activity Capture

**Goal:** Implement the Win32 tracking loop to capture `ProcessName` and `WindowTitle` safely.

**Scope:**
- `chroniclecore/internal/tracker` (Go package)
- `GetForegroundWindow`, `GetWindowThreadProcessId`, `GetWindowTextW` (Win32 APIs)
- `execution/generate_fixtures.py` (Mock data alignment)

**Inputs:**
- `spec/rules_default.json` (Redaction regexes - if available, else hardcode safe defaults)

**Outputs:**
- A `Tracker` struct that runs a loop (1s interval).
- Emits `RawEvent` struct to a channel.
- **Privacy:** `RawEvent.Title` MUST be run through a Redactor function.

## 1. Win32 Implementation (Go)
- Use `golang.org/x/sys/windows` or `syscall`.
- Function `GetActiveWindowInfo() (app string, title string, err error)`
  - Call `user32.GetForegroundWindow()`.
  - Call `user32.GetWindowThreadProcessId()` to get PID.
  - Open Process with `PROCESS_QUERY_INFORMATION` rights.
  - Call `GetModuleFileNameEx` or `QueryFullProcessImageName` to get `.exe` name (e.g. `chrome.exe`).
  - Call `user32.GetWindowTextLengthW` and `user32.GetWindowTextW` to get title.

## 2. Redaction Logic
- Implement `RedactTitle(original string) string`:
  - If "Privacy Mode" is ON (default): return "Redacted".
  - Else:
    - Apply Regex scrubbers (e.g. Credit Card patterns, Email patterns).
    - If `block_title_capture` for this app is true: return "Blocked".

## 3. Idle Detection Integration
- Call `user32.GetLastInputInfo()` every loop.
- If `(CurrentTick - LastInputTick) > IdleThreshold` (e.g. 5 mins):
  - Mark event state as `IDLE`.
  - Pause standard polling or emit special "Idle" event.

## Acceptance Criteria
- [ ] Running tracker prints `Active: chrome.exe - Google search`.
- [ ] Switching windows updates the output within 2 seconds.
- [ ] Stopping mouse/keyboard for >300s triggers `IDLE` state.
