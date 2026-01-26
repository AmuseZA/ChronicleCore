# ChronicleCore Browser Extension

A browser extension that automatically tracks your browsing activity and sends it to the local ChronicleCore server for time tracking and billing.

## Features

- **Automatic Tab Tracking**: Tracks when you switch tabs or load new pages
- **Human-Readable Descriptions**: Generates meaningful descriptions like:
  - "Chatted with John Smith on WhatsApp"
  - "Reviewed PR #42 on org/repo"
  - "Edited document: Q4 Report"
  - "Email: Re: Project Update"
- **Privacy-First**: Only sends data to localhost (127.0.0.1:8080)
- **Pause/Resume**: Toggle tracking from the extension popup

## Supported Browsers

- Google Chrome
- Microsoft Edge
- Brave
- Opera
- Firefox (requires manifest.v2.json)

## Installation

### Prerequisites

1. ChronicleCore server must be running on `http://127.0.0.1:8080`
2. PNG icon files (see "Creating Icons" section below)

### Chrome / Edge / Brave / Opera

1. Open your browser's extension page:
   - **Chrome**: Navigate to `chrome://extensions`
   - **Edge**: Navigate to `edge://extensions`
   - **Brave**: Navigate to `brave://extensions`
   - **Opera**: Navigate to `opera://extensions`

2. Enable **Developer mode** (toggle in the top-right corner)

3. Click **"Load unpacked"**

4. Select this folder (`chroniclecore-extension`)

5. The extension icon should appear in your toolbar

### Firefox

Firefox requires Manifest V2 format. To use this extension in Firefox:

#### Temporary Installation (for testing)

1. Navigate to `about:debugging#/runtime/this-firefox`
2. Click **"Load Temporary Add-on"**
3. Select the `manifest.v2.json` file from this folder
4. **Note**: Temporary add-ons are removed when Firefox closes

#### Permanent Installation (self-signed)

1. Install web-ext tool:
   ```bash
   npm install -g web-ext
   ```

2. Get Mozilla API credentials from https://addons.mozilla.org/developers/addon/api/key/

3. Sign the extension:
   ```bash
   cd chroniclecore-extension
   web-ext sign --api-key=$AMO_JWT_ISSUER --api-secret=$AMO_JWT_SECRET
   ```

4. Install the generated `.xpi` file via `about:addons` → gear icon → "Install Add-on From File"

## Creating Icons

The extension requires PNG icons. These are generated from the SVG templates in `icons/` using the `tools/icon-converter` script.
To regenerate icons:
1. Navigate to `tools/icon-converter`
2. Run `npm install`
3. Run `node convert.js`

## Usage

1. **Start ChronicleCore**: Make sure the ChronicleCore server is running

2. **Install the extension**: Follow the installation steps above

3. **Check status**: Click the extension icon to see:
   - Server connection status
   - Tracking status (active/paused)
   - Last recorded activity

4. **Pause/Resume**: Click the button in the popup to toggle tracking

5. **View data**: Open ChronicleCore dashboard at http://127.0.0.1:8080 to see tracked activities

## Privacy & Security

- **Localhost only**: The extension ONLY communicates with `127.0.0.1:8080`
- **No external servers**: No data is sent outside your machine
- **Description summaries**: Only human-readable descriptions are stored, not full URLs
- **Contact names from titles**: Names come from page titles (publicly visible), not message content
- **Minimal permissions**: Only requests `tabs`, `activeTab`, and `storage` permissions

## Supported Sites (Activity Descriptions)

The extension generates smart descriptions for many popular sites:

| Site | Example Description |
|------|---------------------|
| WhatsApp Web | "Chatted with John Smith on WhatsApp" |
| Slack | "Chatted in #general on Slack" |
| Gmail | "Email: Re: Project Update" |
| Google Docs | "Edited document: Q4 Report" |
| GitHub | "Reviewed PR #42 on org/repo" |
| Figma | "Designed: App Redesign in Figma" |
| Notion | "Edited: Meeting Notes in Notion" |
| YouTube | "Watched: How to Code in Go" |
| LinkedIn | "Viewed LinkedIn profile" |
| And more... | Falls back to "Browsed domain.com: Page Title" |

## Troubleshooting

### "Server not reachable" error

1. Make sure ChronicleCore is running
2. Check that it's accessible at http://127.0.0.1:8080/health
3. Verify no firewall is blocking localhost connections

### Extension not tracking

1. Check if tracking is paused (click extension icon)
2. Verify the server is connected
3. Internal browser pages (chrome://, about:, etc.) are not tracked

### Icons not showing

1. Convert the SVG files to PNG (see "Creating Icons" section)
2. Reload the extension after adding PNG files

## Development

### File Structure

```
chroniclecore-extension/
├── manifest.json       # Chrome/Edge/Brave/Opera (MV3)
├── manifest.v2.json    # Firefox (MV2)
├── background.js       # Service worker - main tracking logic
├── patterns.js         # Activity description patterns
├── popup/
│   ├── popup.html      # Extension popup UI
│   ├── popup.css       # Popup styles
│   └── popup.js        # Popup interaction logic
├── icons/
│   ├── icon16.svg      # SVG templates
│   ├── icon48.svg
│   └── icon128.svg
└── README.md           # This file
```

### Adding New Patterns

Edit `patterns.js` to add new activity description patterns:

```javascript
{
  match: (url) => url.hostname.includes('example.com'),
  describe: (url, title) => {
    // Extract relevant info from URL/title
    return `Did something on Example`;
  }
}
```

## License

Part of ChronicleCore. See main repository for license information.
