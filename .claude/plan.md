# ChronicleCore Development Roadmap

---

## v1.7.0 - COMPLETED

### Bug Fixes
- **ML Sidecar Status Bug**: Fixed `IsRunning()` to perform actual health checks
- **Auto-restart sidecar**: Added auto-restart logic in `TriggerTraining()`

### Features Implemented
- **Smart Noise Filtering**: 30s minimum threshold + 2-minute merge window
- **Manual Time Entries**: API endpoint + ManualEntryModal component
- **Timely-Style Grouping**: ActivityGroup component on Dashboard and History
- **Dashboard 4-Card Layout**: Time Worked, Billable, Focus, Needs Review

---

## v1.8.0 - Enhanced Activity Summaries (PLANNED)

### Overview
Implement intelligent window title parsing to generate rich, human-readable activity summaries like "Edited Invoice INV-1234 for Client ABC in Xero" instead of generic app names.

### Feature 1: Title Parser Engine

**New File:** `apps/chroniclecore-core/internal/engine/title_parser.go`

Parse window titles to extract:
- **Action type**: CREATE, EDIT, VIEW, COMMUNICATE, REVIEW
- **Document type**: invoice, quote, email, spreadsheet, document, presentation
- **Entity references**: invoice numbers (INV-1234), client names, file names

#### Pattern Library

```go
type TitlePattern struct {
    App       string         // App name pattern (regex)
    Regex     *regexp.Regexp // Title pattern
    Action    string         // Detected action type
    DocType   string         // Document type
    Extractor func(matches []string) map[string]string // Extract entities
}

var patterns = []TitlePattern{
    // === XERO ===
    {App: "xero", Regex: `Invoice (INV-\d+)`, Action: "EDIT", DocType: "invoice"},
    {App: "xero", Regex: `New Invoice`, Action: "CREATE", DocType: "invoice"},
    {App: "xero", Regex: `Quote (QU-\d+)`, Action: "EDIT", DocType: "quote"},
    {App: "xero", Regex: `New Quote`, Action: "CREATE", DocType: "quote"},
    {App: "xero", Regex: `Bill (BILL-\d+)`, Action: "EDIT", DocType: "bill"},
    {App: "xero", Regex: `Contact.*-\s*(.+)`, Action: "VIEW", DocType: "contact"},
    {App: "xero", Regex: `Bank Reconciliation`, Action: "REVIEW", DocType: "reconciliation"},

    // === MICROSOFT OFFICE ===
    {App: "WINWORD", Regex: `(.+\.docx?)`, Action: "EDIT", DocType: "document"},
    {App: "WINWORD", Regex: `Document\d+`, Action: "CREATE", DocType: "document"},
    {App: "EXCEL", Regex: `(.+\.xlsx?)`, Action: "EDIT", DocType: "spreadsheet"},
    {App: "EXCEL", Regex: `Book\d+`, Action: "CREATE", DocType: "spreadsheet"},
    {App: "POWERPNT", Regex: `(.+\.pptx?)`, Action: "EDIT", DocType: "presentation"},
    {App: "OUTLOOK", Regex: `(.+) - Message`, Action: "COMMUNICATE", DocType: "email"},
    {App: "OUTLOOK", Regex: `Inbox`, Action: "VIEW", DocType: "email"},
    {App: "OUTLOOK", Regex: `Calendar`, Action: "VIEW", DocType: "calendar"},

    // === BROWSERS (Gmail, etc.) ===
    {App: "chrome|edge|firefox", Regex: `Inbox.*Gmail`, Action: "VIEW", DocType: "email"},
    {App: "chrome|edge|firefox", Regex: `Compose.*Gmail`, Action: "CREATE", DocType: "email"},
    {App: "chrome|edge|firefox", Regex: `(.+) - Google Docs`, Action: "EDIT", DocType: "document"},
    {App: "chrome|edge|firefox", Regex: `(.+) - Google Sheets`, Action: "EDIT", DocType: "spreadsheet"},

    // === COMMUNICATION APPS ===
    {App: "slack", Regex: `(.+) \| Slack`, Action: "COMMUNICATE", DocType: "chat"},
    {App: "teams", Regex: `(.+) \| Microsoft Teams`, Action: "COMMUNICATE", DocType: "meeting"},
    {App: "zoom", Regex: `Zoom Meeting`, Action: "COMMUNICATE", DocType: "meeting"},

    // === CODE EDITORS ===
    {App: "code", Regex: `(.+) - Visual Studio Code`, Action: "EDIT", DocType: "code"},
    {App: "devenv", Regex: `(.+) - Microsoft Visual Studio`, Action: "EDIT", DocType: "code"},
}
```

#### Parser Implementation

```go
type ParsedTitle struct {
    Action      string            // CREATE, EDIT, VIEW, etc.
    DocType     string            // invoice, email, document, etc.
    Entities    map[string]string // invoice_number: "INV-1234", client: "ABC Corp"
    RawTitle    string
    Confidence  float64           // 0.0 - 1.0
}

func ParseWindowTitle(appName, title string) ParsedTitle {
    for _, pattern := range patterns {
        if !matchesApp(appName, pattern.App) {
            continue
        }
        if matches := pattern.Regex.FindStringSubmatch(title); matches != nil {
            entities := map[string]string{}
            if pattern.Extractor != nil {
                entities = pattern.Extractor(matches)
            }
            return ParsedTitle{
                Action:     pattern.Action,
                DocType:    pattern.DocType,
                Entities:   entities,
                RawTitle:   title,
                Confidence: 0.9,
            }
        }
    }
    // Fallback: generic parsing
    return ParsedTitle{
        Action:     inferAction(title),
        DocType:    "application",
        RawTitle:   title,
        Confidence: 0.5,
    }
}
```

### Feature 2: Group Summary Generation

**New File:** `apps/chroniclecore-core/internal/engine/summary_generator.go`

Generate Timely-style natural language summaries for activity groups.

```go
type SummaryContext struct {
    Actions   map[string]int      // CREATE: 2, EDIT: 5, VIEW: 3
    DocTypes  map[string]int      // invoice: 3, email: 5
    Entities  []string            // ["INV-1234", "Client ABC"]
    Apps      []string            // ["Xero", "Outlook"]
    Duration  time.Duration
}

func GenerateGroupSummary(activities []ParsedActivity) string {
    ctx := analyzeSummaryContext(activities)

    // Build natural language summary
    parts := []string{}

    // Primary action + document type
    primaryAction, primaryDoc := getPrimaryActivity(ctx)
    parts = append(parts, formatPrimaryActivity(primaryAction, primaryDoc, ctx.DocTypes[primaryDoc]))

    // Secondary activities
    if len(ctx.Actions) > 1 {
        parts = append(parts, formatSecondaryActivities(ctx))
    }

    // Entity references
    if len(ctx.Entities) > 0 {
        parts = append(parts, formatEntities(ctx.Entities))
    }

    return strings.Join(parts, ", ")
}

// Example outputs:
// "Edited 3 invoices in Xero, sent 5 emails"
// "Managed NIS accounts, issued quote and invoice communications"
// "Worked on Project ABC spreadsheet, reviewed documentation"
// "Phone call with Client XYZ, updated CRM records"
```

### Feature 3: Real-time Summary Display

Update block creation to include parsed summaries:

```go
// In aggregator.go
func (bb *blockBuilder) build() *store.Block {
    // ... existing logic ...

    // Parse and generate summary
    parsed := ParseWindowTitle(bb.primaryAppName, bb.dominantTitle)
    summary := GenerateSingleActivitySummary(parsed)

    return &store.Block{
        // ... existing fields ...
        TitleSummary:   summary,
        ActionType:     parsed.Action,
        EntityContext:  encodeEntities(parsed.Entities),
    }
}
```

### Database Changes

Already migrated in v1.7.0:
- `action_type TEXT` - Stores detected action (CREATE, EDIT, etc.)
- `entity_context TEXT` - JSON-encoded entity references

### Files to Create

| File | Purpose |
|------|---------|
| `internal/engine/title_parser.go` | Window title pattern matching |
| `internal/engine/summary_generator.go` | Natural language summary generation |
| `internal/engine/patterns/` | Extensible pattern definitions |

### Files to Modify

| File | Changes |
|------|---------|
| `internal/engine/aggregator.go` | Integrate title parser on block creation |
| `internal/api/blocks.go` | Return action_type and entity_context |
| `src/lib/components/ActivityGroup.svelte` | Display richer summaries |

---

## v1.9.0 - Profile Detail Page Grouping (PLANNED)

### Overview
Apply Timely-style grouping to the Profile detail page (`/profiles/[id]`), showing all activities for a specific profile in the same grouped format as Dashboard and History.

### Changes Required

**File:** `apps/chroniclecore-ui/src/routes/profiles/[id]/+page.svelte`

1. Import `ActivityGroup` component
2. Add `groupBlocks()` function (same as Dashboard/History)
3. Replace existing block list with grouped display
4. Add "Add Time" button for manual entries pre-selecting the profile

```svelte
<script lang="ts">
    import ActivityGroup from "$lib/components/ActivityGroup.svelte";
    import ManualEntryModal from "$lib/components/ManualEntryModal.svelte";

    // ... existing code ...

    let showManualModal = false;
    let activityGroups = [];

    async function loadProfile() {
        // ... existing load logic ...
        activityGroups = groupBlocks(blocks);
    }
</script>

<!-- Profile Header -->
<div class="flex justify-between items-center">
    <h1>{profile.client_name} - {profile.service_name}</h1>
    <button on:click={() => showManualModal = true}>
        Add Time
    </button>
</div>

<!-- Activity Groups -->
<div class="space-y-3">
    {#each activityGroups as group}
        <ActivityGroup {group} showActions={true} />
    {/each}
</div>

<!-- Manual Entry Modal (pre-selected profile) -->
<ManualEntryModal
    bind:isOpen={showManualModal}
    preselectedProfileId={profile.profile_id}
    on:created={loadProfile}
/>
```

---

## v2.0.0 - Invoicing & Reporting (PLANNED)

### Overview
Major release adding invoice generation, time reports, and client billing features.

### Feature 1: Invoice Generation

#### Database Schema

```sql
-- Invoice table
CREATE TABLE invoice (
    invoice_id INTEGER PRIMARY KEY,
    client_id INTEGER NOT NULL,
    invoice_number TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'draft', -- draft, sent, paid, overdue
    issue_date TEXT NOT NULL,
    due_date TEXT NOT NULL,
    subtotal REAL NOT NULL,
    tax_rate REAL DEFAULT 0,
    tax_amount REAL DEFAULT 0,
    total REAL NOT NULL,
    currency TEXT DEFAULT 'AUD',
    notes TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    FOREIGN KEY (client_id) REFERENCES dict_client(client_id)
);

-- Invoice line items (linked to blocks)
CREATE TABLE invoice_line (
    line_id INTEGER PRIMARY KEY,
    invoice_id INTEGER NOT NULL,
    block_id INTEGER,  -- NULL for manual line items
    description TEXT NOT NULL,
    quantity REAL NOT NULL,  -- Hours
    unit_price REAL NOT NULL,  -- Hourly rate
    amount REAL NOT NULL,
    FOREIGN KEY (invoice_id) REFERENCES invoice(invoice_id) ON DELETE CASCADE,
    FOREIGN KEY (block_id) REFERENCES block(block_id)
);
```

#### API Endpoints

```go
// Invoice management
POST   /api/v1/invoices              // Create invoice from time entries
GET    /api/v1/invoices              // List invoices
GET    /api/v1/invoices/:id          // Get invoice details
PUT    /api/v1/invoices/:id          // Update invoice
DELETE /api/v1/invoices/:id          // Delete draft invoice
POST   /api/v1/invoices/:id/send     // Mark as sent
POST   /api/v1/invoices/:id/paid     // Mark as paid

// Invoice generation
POST /api/v1/invoices/generate
{
    "client_id": 1,
    "profile_ids": [1, 2, 3],  // Which profiles to include
    "start_date": "2024-01-01",
    "end_date": "2024-01-31",
    "group_by": "profile",  // profile, day, none
    "include_descriptions": true
}
```

#### Frontend Components

| Component | Purpose |
|-----------|---------|
| `InvoiceList.svelte` | List/filter invoices |
| `InvoiceDetail.svelte` | View/edit invoice |
| `InvoiceGenerator.svelte` | Create invoice from time entries |
| `InvoicePDF.svelte` | PDF preview/download |

### Feature 2: Time Reports

Generate detailed time reports for clients or internal use.

```go
// Report types
type ReportConfig struct {
    Type       string   // summary, detailed, by-client, by-project
    DateRange  DateRange
    ClientIDs  []int64
    ProfileIDs []int64
    GroupBy    string   // day, week, month, profile, client
    Format     string   // json, csv, pdf
}

// Report endpoints
GET /api/v1/reports/summary?start=...&end=...
GET /api/v1/reports/detailed?start=...&end=...&client_id=...
GET /api/v1/reports/export/csv?...
GET /api/v1/reports/export/pdf?...
```

### Feature 3: Dashboard Reporting Widget

Add a new card/section to Dashboard showing:
- Unbilled hours by client
- Pending invoices
- Quick "Generate Invoice" button

---

## v2.1.0 - Xero Integration (PLANNED)

### Overview
Direct integration with Xero for syncing invoices and clients.

### Features

1. **OAuth2 Authentication**
   - Connect to Xero account
   - Store tokens securely
   - Auto-refresh tokens

2. **Client Sync**
   - Import clients from Xero
   - Map ChronicleCore clients to Xero contacts
   - Two-way sync option

3. **Invoice Push**
   - Create invoices in Xero from ChronicleCore
   - Sync invoice status (sent, paid)
   - Link payments

### API Endpoints

```go
// Xero connection
GET  /api/v1/integrations/xero/connect     // Start OAuth flow
GET  /api/v1/integrations/xero/callback    // OAuth callback
GET  /api/v1/integrations/xero/status      // Check connection status
POST /api/v1/integrations/xero/disconnect  // Disconnect

// Sync operations
POST /api/v1/integrations/xero/sync/clients    // Sync clients
POST /api/v1/integrations/xero/push/invoice    // Push invoice to Xero
GET  /api/v1/integrations/xero/invoices        // List Xero invoices
```

---

## v2.2.0 - Mobile Companion App (PLANNED)

### Overview
React Native app for iOS/Android to view time entries and add manual time on the go.

### Features

1. **View Dashboard**
   - Today's time worked
   - Activity groups
   - Quick stats

2. **Manual Time Entry**
   - Add time entries
   - Select profile/client
   - Set date/time range

3. **Push Notifications**
   - Reminder to log time
   - ML classification suggestions

### Tech Stack
- React Native
- Same Go backend (via REST API)
- Local SQLite cache for offline support

---

## Future Ideas (Backlog)

### Productivity Analytics
- Weekly/monthly productivity trends
- App usage breakdowns
- Focus time vs. meeting time analysis
- Comparison with previous periods

### Team Features
- Multi-user support
- Team dashboards
- Shared clients/profiles
- Time approval workflows

### Calendar Integration
- Google Calendar sync
- Outlook Calendar sync
- Auto-create blocks from calendar events
- Meeting detection

### Browser Extension
- Capture URL context for browser activities
- Auto-detect project context from URLs
- Quick profile assignment popup

### AI Enhancements
- GPT-powered activity summaries
- Automatic client/project detection from context
- Smart time suggestions based on patterns
- Anomaly detection (unusually long/short sessions)

---

## Version History

| Version | Status | Key Features |
|---------|--------|--------------|
| v1.6.0 | Released | Initial release, ML classification |
| v1.6.1 | Released | Hotfix: Database migration, Pydantic fix |
| v1.7.0 | Released | Noise filtering, Manual entries, Timely grouping |
| v1.8.0 | Planned | Enhanced activity summaries, Title parser |
| v1.9.0 | Planned | Profile page grouping |
| v2.0.0 | Planned | Invoicing & Reporting |
| v2.1.0 | Planned | Xero Integration |
| v2.2.0 | Planned | Mobile App |
