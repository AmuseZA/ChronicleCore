# ChronicleCore v2.3.0 Release Notes

## Improvements

### Polished Dark Mode
Completed dark mode styling across the application to ensure consistency and readability:
- **Main Dashboard**: Improved contrast for stats, dates, and text.
- **Profiles List**: Fixed table headers, rows, and empty states.
- **Suggestions Page**: Enhanced visibility for suggestion cards and headers.
- **Profile Details**: Refined styling for all elements.
- **Review & History Pages**: Applied final touches for seamless dark theme support.

### Profile View Enhancements
- **Collapsed Date Groups**: Date groups in the profile detail view now default to **collapsed**, making it easier to navigate large lists of activities without excessive scrolling. Click any date to expand and view details.

## Bug Fixes

### Data Integrity
- **Profile Stats Fix**: Resolved an issue where profile statistics could be inaccurate due to a database query limit. Removed the `LIMIT 100` restriction to ensure all blocks in the selected range are included in calculations.
- **Suggestions Page**: Fixed an HTML entity issue causing display errors in confidence levels.

## Upgrade Notes

- No database migration required.
- Existing dark mode preference will be respected.

## Installation

Download `ChronicleCore_Setup_v2.3.0.exe` and run the installer.
