# Implementation Plan - ChronicleCore v2.2 (Polishing & Enhancements)

This plan outlines the roadmap for version 2.2, focusing on user experience polish, installation improvements, and technical debt reduction.

## User Review Required
> [!NOTE]
> This plan will be updated as new findings occur during the v2.1 rollout.

## Proposed Changes

### Installer & Packaging
Application identity and system integration improvements.

#### [NEW] [Embed Application Icon]
- **Goal**: Ensure the generated Desktop Shortcut and Taskbar entry display the correct ChronicleCore logo instead of the default Go/Windows executable icon.
- **Approach**:
    - Use `rsrc` or `go-winres` to embed `app_icon.ico` directly into the `chroniclecore.exe` binary structure during build.
    - Update `prepare_installer.ps1` to ensure the icon resource is generated before `go build`.
    - Verify Inno Setup uses the embedded icon for shortcuts.

### Server & Backend
#### [MODIFY] [Update Checker Logic]
- **Goal**: Fix confusing log messages when local version is newer than remote (dev scenarios).
- **Current Behavior**: Logs "No update available (current: 2.1.0, latest: 2.0.0)" which implies a downgrade is an update or is just confusing.
- **Approach**: Implement semantic version comparison (SemVer) and handle `local > remote` specifically (e.g., "You are running a pre-release version").

#### [MODIFY] [Event Architecture]
- **Goal**: Move towards a true asynchronous event bus.
- **Approach**: Refactor `windows.go` loop to decouple detection from enrichment completely, potentially using channels for "DetectedWindow" events.

### UI / UX
#### [MODIFY] [Dark Mode Polish]
- **Goal**: Ensure comprehensive dark mode support across all new components (Sidebars, Modals, Cards).
- **Status**: Sidebar and Dashboard Cards fixes applied in post-v2.1 hotfix.
