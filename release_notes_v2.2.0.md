# ChronicleCore v2.2.0 Release Notes

## New Features

### Custom Date Range Reports
Generate billing reports for specific date ranges with per-day itemized breakdown.

- **Custom Range View**: Select specific start/end dates (e.g., Jan 19-23) for detailed billing
- **Per-Day Breakdown**: View billable hours and amounts for each day
- **Export Options**:
  - **Copy as Text**: Formatted timesheet report with daily breakdown, task summaries, and totals
  - **Download CSV**: Spreadsheet export with every billable item (Date, Task, Hours, Rate, Amount)

### Improved Update Checker
- Proper semantic version comparison (SemVer)
- Better messaging for development builds: "Running development build (v2.2.0 > latest release v2.1.0)"
- No more confusing "No update available" messages when local version is newer

## UI/UX Improvements

### Dark Mode Comprehensive Fix
Fixed dark mode styling across all pages:
- Review page cards and modals
- History page timeline and stats
- Suggestions page cards
- Profiles list and detail pages
- Settings and Blacklist pages
- Clients, Services, and Rates pages
- Create Profile page

### Elements Fixed
- Card backgrounds now properly use dark slate colors
- Borders adapt to dark theme
- Text colors maintain readability in dark mode
- Modal overlays work correctly

## Technical Details

- Updated `compareVersions()` function for proper semver comparison
- Added `parseVersion()` helper for extracting version components
- Comprehensive dark mode class additions across 12+ page templates

## Upgrade Notes

- All existing data is preserved
- New date range picker defaults to current date
- Dark mode toggle in top-right corner of all pages

## Installation

Download `ChronicleCore_Setup_v2.2.0.exe` and run the installer.
