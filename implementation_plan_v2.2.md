# Implementation Plan - ChronicleCore v2.2 (Polishing & Enhancements)

This plan outlines the roadmap for version 2.2, focusing on user experience polish, installation improvements, and technical debt reduction.

## Completed Features

### [DONE] Custom Date Range Reports
- **Goal**: Allow users to generate billing reports for specific date ranges (e.g., 19th - 23rd January)
- **Features Implemented**:
    - New "Custom Range" view mode in Profile detail page
    - Date pickers for start/end date selection
    - Per-day billing breakdown with amounts calculated from hourly rate
    - "Copy as Text" - Formatted timesheet report for invoicing
    - "Download CSV" - Spreadsheet export with all billable items
    - Auto-expand all dates in custom range view for easy review

## In Progress

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

#### [DEFERRED] [Event Architecture]
- **Goal**: Move towards a true asynchronous event bus.
- **Status**: Deferred to v2.3 - requires significant refactoring
- **Approach**: Refactor `windows.go` loop to decouple detection from enrichment completely, potentially using channels for "DetectedWindow" events.

### UI / UX

#### [MODIFY] [Light Mode Aesthetics]
- **Goal**: Address "too much white" feedback and improve visual hierarchy.
- **Specific Tasks**:
    - **Dashboard Cards**: Add subtle background colors, stronger borders, or accents to summary cards so they stand out against the page background.
    - **Page Backgrounds**: Use subtle gray background (slate-50) instead of pure white
    - **General Layout**: Audit high-level pages (Settings, Profiles) to reduce clinical white space.

#### [MODIFY] [Dark Mode Comprehensive Audit]
- **Goal**: Fix remaining unstyled or broken elements in Dark Mode.
- **Identified Issues**:
    - **Live Activity**: Currently "too white" or clashing in dark mode. Needs specific dark styling update.
    - **Review Page**: Main content cards remain white.
    - **History & Reports**: Summary cards and list backgrounds remain white.
    - **ML Suggestions**: Suggestion cards are white.
    - **Profiles & Rates**: List items and modal backgrounds remain white.
    - **Shadows/Gradients**: Fix white gradients/shadows appearing on dark backgrounds (likely `from-white` or `shadow-xl` without dark mode overrides).
    - **Form Inputs**: Date pickers, select dropdowns, and text inputs need dark styling.
    - **Modals**: Profile selector, confirmation dialogs need dark backgrounds.

## Release Checklist
- [ ] Update version to 2.2.0 in main.go and chroniclecore.iss
- [ ] Build backend with embedded icon
- [ ] Build frontend with UI fixes
- [ ] Run prepare_installer.ps1
- [ ] Run build_inno_installer.ps1
- [ ] Create release notes
- [ ] Commit and tag v2.2.0
- [ ] Push to GitHub
