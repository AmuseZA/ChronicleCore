# ChronicleCore

**Local-first Windows time tracking with deterministic correctness.**

ChronicleCore is a privacy-focused desktop application for tracking work activity, managing client profiles, and generating invoice-ready exports. Built on a 3-layer architecture (Directive → Orchestration → Execution), it prioritizes deterministic business logic over AI complexity.

## 🚀 Features

### Core Tracking
- **Local-first**: All data stays on your machine. No cloud sync, no remote servers.
- **Privacy by default**: Captures window titles and process names locally.
- **Deep Activity Tracking**: (v2.0.0) Optional detailed content extraction for emails, chats, and documents.
- **Idle Detection**: Smartly excludes idle time from billing calculations.

### Intelligence
- **ML Suggestions**: Learned profile suggestions based on your history.
- **Smart Tagging**: Automatic categorization of browser-based apps (Gmail, Jira, etc.).
- **Keyword Blacklisting**: Hide activities containing specific words (e.g., "Facebook").

### Browser Extension (v2.0.0)
- **Cross-browser support**: Chrome, Edge, Brave, Firefox.
- **Detailed Insights**: Captures specific URLs, page titles, and generates human-readable activity descriptions (e.g., "Chatted with John on WhatsApp").
- **Privacy**: Only sends data to the local ChronicleCore instance (127.0.0.1).

## 📥 Installation

1. Go to the [Releases](https://github.com/AmuseZA/ChronicleCore/releases) page.
2. Download the latest installer: **ChronicleCore_Setup_v2.0.0.exe**.
3. Run the installer (includes embedded Python and all dependencies).
4. Launch via Desktop shortcut or Start Menu.

## 📜 Version History

### [v2.0.0] - 2026-01-27
- **Major Release**
- **Browser Extension**: Added cross-browser extension for detailed web activity tracking.
- **Deep Activity Tracking**: Enhanced desktop tracking for detailed content (Outlook, VS Code, etc.).
- **Settings API**: New endpoints for managing granular tracking preferences.

### [v1.8.10]
- **Critical Fix**: Resolved "Failed to scan profile" errors by correctly handling NULL profile names.

### [v1.8.9]
- **Hotfix**: Fixed silent migration failures for profile name columns.

### [v1.8.8]
- **Bug Fix**: Fixed issue where rejected ML suggestions would disappear from the review page.

### [v1.8.1]
- **ML Improvements**: Auto-predictions on stop, learning from deletions/rejections.
- **UX**: "Delete Group" button, scroll preservation.

### [v1.8.0]
- **Features**: Keyword Blacklist, UI-integrated ML suggestions.

## Architecture

### 3-Layer Model

```
┌─────────────────────────────────────────────┐
│ Layer 1: Directives (What to do)           │
│ - SOPs in directives/                       │
│ - Feature-scoped, explicit acceptance       │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│ Layer 2: Orchestration (Decision making)   │
│ - LLM routes tasks to scripts               │
│ - Enforces constraints & validates outputs  │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│ Layer 3: Execution (Doing the work)        │
│ - Deterministic scripts in execution/      │
│ - Schema validation, contract tests, etc.  │
└─────────────────────────────────────────────┘
```

## Security & Privacy

### Binding Policy
- Server MUST bind to `127.0.0.1` only (enforced in code).
- No remote access permitted.
- CORS restricted to localhost origins.

### Data Capture Policy
- ✅ Window titles (with redaction support)
- ✅ Process names (e.g., EXCEL.EXE)
- ✅ Idle detection
- ✅ Local-only Browser Extension (v2.0.0)
- ❌ Screenshots (prohibited)
- ❌ Keystrokes (prohibited)

## License

[To be determined - proprietary/commercial expected]

## Contact

ChronicleCore Project
Built with the 3-layer DOE/SOP methodology.
