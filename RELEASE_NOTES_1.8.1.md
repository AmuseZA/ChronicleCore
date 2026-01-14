# ChronicleCore v1.8.1

**Release Date:** January 14, 2026

## What's New

### ML Intelligence Improvements
- **Automatic ML Predictions on Session Stop**: When you stop a tracking session, ML predictions now trigger automatically (if trained model exists)
- **ML Learns from Deletions**: The system now learns from blocks you delete and suggests similar items for deletion in the future
- **ML Learns from Rejections**: When you reject a suggestion, it's recorded as negative training data to improve future predictions
- **ML Suggestions Page**: New dedicated `/suggestions` page for reviewing AI-generated profile suggestions grouped by confidence (HIGH/MEDIUM/LOW)
- **Deletion Suggestions**: ML Suggestions page now shows "Suggested Deletions" section based on your deletion patterns
- **ML Visibility Fix**: Fixed issue where ML-suggested blocks would disappear from review page

### Blacklist Enhancements  
- **Keyword Blacklist Filtering**: Blocks matching blacklisted keywords are now correctly filtered from the review page

### Review Page UX Improvements
- **Delete Group**: New button to delete all items in a group at once
- **Scroll Preservation**: Page no longer jumps to top after delete operations
- **Add Entry Link**: Quick access to add manual entries from review page

### Settings
- **Skip Delete Confirmations**: New toggle in Settings → User Preferences to disable delete confirmation dialogs

---

## Installation

**ChronicleCore_Setup_v1.8.1.exe** - Windows installer with embedded Python and all dependencies

1. Download the installer below
2. Run the installer (30 second wizard)
3. Launch from Start Menu or Desktop shortcut

---

## Technical Changes

### Backend
- `apps/chroniclecore-core/cmd/server/main.go`: Auto-trigger predictions on session stop
- `apps/chroniclecore-core/internal/api/blocks.go`: Fixed `needs_review` filter, added keyword blacklist filtering
- `apps/chroniclecore-core/internal/api/ml.go`: Added `RejectSuggestion` endpoint

### Frontend
- `apps/chroniclecore-ui/src/routes/suggestions/+page.svelte`: New ML Suggestions page
- `apps/chroniclecore-ui/src/routes/review/+page.svelte`: Delete group, scroll fix, Add Entry link
- `apps/chroniclecore-ui/src/routes/settings/+page.svelte`: Skip delete confirmations toggle
- `apps/chroniclecore-ui/src/lib/components/Sidebar.svelte`: ML Suggestions navigation link
