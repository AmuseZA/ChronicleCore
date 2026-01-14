# ChronicleCore v1.8.0

## 🎉 New Features

### Keyword Blacklist
- **Block by keyword** - Hide activities containing specific words in their title (e.g., "Facebook", "Netflix")
- **Settings > Blacklist** - New "Keywords" section to manage keyword filters
- **Review Page Modal** - Click "Blacklist" to choose between blocking the entire app OR a specific keyword

### ML Suggestions UI
- **AI Suggested badges** - ML predictions now display with purple "AI Suggested" badges on the Review page
- **Backend integration** - ML suggestions are joined with blocks for seamless display

### Smart Tagging
Improved title extraction for browser-based productivity apps:
- **Xero** - Extracts invoice/document names from browser tabs
- **Gmail, Outlook** - Properly categorized as email activities
- **Google Docs/Sheets** - Recognized and labeled correctly
- **Excel, Word** - Document names extracted from window titles

### Data Integrity
- **MinBlockDuration reduced** to 10 seconds - Captures shorter activities more accurately

## 📦 API Changes

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/blacklist/keywords` | GET | List all keyword blacklist entries |
| `/api/v1/blacklist/keywords` | POST | Add a keyword to the blacklist |
| `/api/v1/blacklist/keywords/{id}` | DELETE | Remove a keyword from the blacklist |

## 🛠️ Technical Changes

- `blacklist.go` - Added `ListKeywordBlacklist`, `RemoveFromKeywordBlacklist` handlers
- `main.go` - Registered keyword blacklist API routes
- `blocks.go` - Enhanced `extractTitleContext` for browser apps
- `aggregator.go` - Reduced `MinBlockDuration` to 10 seconds
- `store_schema.go` - Added `keyword_blacklist` table migration
- Review page Svelte - New blacklist modal component
- Settings blacklist page - Keyword management UI

## 📥 Download

**ChronicleCore_Setup_v1.8.0.exe** - Windows installer with embedded Python and all dependencies
