# ChronicleCore v2.0.0 Release Notes

## Major New Features

### Browser Extension
A cross-browser extension for enhanced web activity tracking with intelligent activity descriptions.

**Features:**
- Captures detailed browsing activity (URLs, page titles, domains)
- Generates human-readable activity descriptions (e.g., "Chatted with John on WhatsApp", "Edited document: Q4 Report")
- Works with Chrome, Edge, Brave, Opera, and Firefox
- Self-hosted distribution - no app store required
- Privacy-focused: only sends data to localhost

**Supported Sites with Smart Descriptions:**
- WhatsApp Web - extracts contact names
- Slack - extracts channel/DM context
- Gmail/Outlook - extracts email subjects
- Google Docs/Notion - extracts document names
- GitHub - extracts PR/issue numbers and repo info
- Figma - extracts design project names
- Microsoft Teams - extracts meeting/channel context
- YouTube - extracts video titles
- LinkedIn - detects messaging vs browsing

**Installation:**
- Chrome/Edge/Brave: Load unpacked from `apps/chroniclecore-extension`
- Firefox: Load as temporary add-on or sign with web-ext

### Deep Activity Tracking
Enhanced desktop tracking that extracts detailed content from applications.

**Tracked Content Types:**
| App Category | Information Extracted |
|-------------|----------------------|
| Microsoft Outlook | Email subject, sender, activity type |
| MS Word/Excel/PowerPoint | Document name, file name |
| Browsers | URL, domain, page title |
| VS Code | File name, project name |
| Visual Studio | File name, solution name |
| JetBrains IDEs | Project name, file name |
| Microsoft Teams | Channel/contact, meeting detection |
| Slack | Channel or DM contact |
| Discord | Server, channel, DM contact |
| WhatsApp | Contact name |
| Telegram | Contact/group name |
| Figma | Design document name |
| File Explorer | Current folder path |

**Privacy Features:**
- All tracking happens locally - no cloud calls
- Privacy Mode for redacting sensitive content
- Granular controls - enable/disable specific content types
- Excluded apps list

### Settings API
New API endpoints for managing tracking settings.

```
GET  /api/v1/settings        - Get all tracking settings
PUT  /api/v1/settings        - Update tracking settings
GET  /api/v1/settings/{key}  - Get single setting value
```

### Updated Settings UI
New "Deep Activity Tracking" section in Settings page with toggles for:
- Deep Tracking Enable/Disable
- Browser content tracking
- Email content tracking
- Document content tracking
- Chat content tracking
- Privacy Mode

## API Changes

### New Endpoint: Event Ingestion
```
POST /api/v1/events/ingest
```
Accepts browser extension events with intelligent activity descriptions.

## Technical Details

- Uses Windows UI Automation for deep content extraction
- Activity score-based idle time exclusion for accurate billing
- Metadata stored in `raw_event.metadata` as JSON
- Settings stored in SQLite `settings` table

## Upgrade Notes

- This is a major version upgrade
- All existing data and configurations are preserved
- New settings default to disabled - enable in Settings page
- Browser extension is optional - install separately if needed

## Installation

Download `ChronicleCore_Setup_v2.0.0.exe` and run the installer.

For browser extension installation instructions, see:
`apps/chroniclecore-extension/README.md`
