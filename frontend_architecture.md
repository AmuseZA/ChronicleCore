1) Frontend Architecture (Svelte v1)
1.1 App Structure

src/routes/ (or src/pages/ depending on routing choice)

src/components/ reusable UI

src/stores/ global state

src/lib/api.ts API client

src/lib/types.ts shared types

src/lib/format.ts formatting helpers

1.2 Routing

Simple client-side routing (recommended):

/today

/review

/calendar

/profiles

/rules

/exports

/settings

/audit

If you prefer zero router complexity, you can use a single-page layout and manage views via a store; routing is still recommended for clarity.

2) Global State (Svelte Stores)
2.1 Stores

trackingStore

status: 'ON'|'PAUSED'|'STOPPED'|'IDLE'

idleSince?: string

currentBlockId?: number

lastUpdated: number

uiStore

sidebarOpen: boolean

activeDate: string (YYYY-MM-DD)

toastQueue: Toast[]

drawer: { open: boolean; blockId?: number; tab?: 'summary'|'assignment'|'description' }

profilesStore

caches clients/projects/services/rates/profiles

exposes quick “recent profiles” list

settingsStore

privacy + storage + AI provider settings snapshot

2.2 Polling Strategy (v1)

Poll /tracking/status every 5–10 seconds

Poll blocks only when:

date changes

user navigates to a date

after edit operations

optional: light background refresh every 30–60 seconds on Today

3) Shared Types (Frontend)
3.1 Core Types

Block

block_id, ts_start, ts_end, duration_minutes

profile_id?: number

confidence: 'HIGH'|'MEDIUM'|'LOW'

billable: boolean

locked: boolean

app: { name }

domain?: { text }

title_summary?: string

notes?: string

description?: string

Profile

profile_id, client_name, service_name, project_name?, rate_name?

Rule

rule_id, name, enabled, priority, match_type, match_value, target_profile_id, confidence_boost

4) Core UI Components (Svelte)
4.1 App Shell

<AppShell>

Layout grid: sidebar + topbar + content

Slots: topbar, sidebar, content

<SidebarNav>

Nav list + active highlighting

Storage indicator + version footer

<TopBar>

Page title slot

Right side: <TrackingPill>, quick actions

<TrackingPill>
Props:

status, idleSince?
Events:

start, stop, pause, resume, addNote

4.2 Timeline Components

<DateNavigator>

Prev/next day buttons

Date picker input

<FilterBar>

Billable filter

Confidence filter

Unassigned-only toggle

<TimelineList>
Props:

blocks: Block[]

selectedIds: Set<number>
Events:

select(blockId, multi)

open(blockId)

bulkMerge(selectedIds)

<BlockRow>
Props:

block: Block

selected: boolean
Events:

toggleSelect

open

quickBillableToggle

openActionsMenu

<ConfidenceBadge>
Props: confidence

<ContextChips>
Props: appName, domain?, titleSummary?

4.3 Editing Components

<BlockDrawer>
Props:

blockId

open

tab
Events:

close

updated (refresh caller)

<AssignmentPicker>
Props:

profiles: Profile[]

recent: Profile[]

suggested?: Profile[]
Events:

select(profileId)

<SplitModal>
Props:

block: Block

open: boolean
Events:

confirm(splitTime)

close

<MergeModal>
Props:

blocks: Block[]

open
Events:

confirm(resolutionChoice)

close

<NotesEditor>
Props: notes
Events: save(notes)

<DescriptionEditor>
Props:

description

aiEnabled
Events:

generateDeterministic

polishWithAI

save(description)

4.4 Tables and CRUD Components

<TabbedManager>

Tabs: Clients/Projects/Services/Rates/Profiles

Each tab uses:

<EntityTable>

<EntityFormDrawer>

<RuleTable> + <RuleEditor>

Includes <RuleTestPanel>

<ExportWizard>

Steps + preview panel

<SettingsPanel>

Sub-panels: Privacy, Storage, AI Providers, About

<AuditTable> + <AuditDetailDrawer>

5) Screen-by-Screen Implementation Notes
5.1 Today Screen (/today)

Components:

<DateNavigator>

<FilterBar>

<TimelineList>

<BlockDrawer>

<SplitModal>, <MergeModal>

Key behaviours:

Multi-select blocks (shift-click)

Reassign from drawer (typeahead)

Quick billable toggle in row

API calls:

GET /api/v1/blocks?date=...

edits: /blocks/{id}/... etc.

5.2 Needs Review (/review)

Components:

<ReviewFilterBar>

<ReviewQueueList>

Batch assign modal (reuse <AssignmentPicker>)

Logic:

Default filter: confidence=LOW OR profile_id IS NULL

Batch operations: assign, mark non-billable, merge (adjacent only)

5.3 Calendar (/calendar)

Week view in v1:

Fetch range

Client-side summarise per day

Click day opens Today

5.4 Profiles (/profiles)

Tabbed CRUD:

Create client/service/rate quickly

Create “Profile” mapping entity as primary

UX rule:

Profiles tab is promoted as “Billing Profiles” so it’s obvious.

5.5 Rules (/rules)

Core:

Rules list + editor

Rule test panel is mandatory (reduces support burden)

5.6 Exports (/exports)

Export wizard:

Show warnings if:

unassigned time exists

low confidence time exists

Provide “Go to Needs Review” CTA

5.7 Settings (/settings)

Privacy:

capture titles toggle

redaction regex list (with test input)

app/domain exclusions

Storage:

retention dropdown

“cleanup now”

“vacuum now”

AI Providers:

mode Off/Local/Cloud

local provider selection + base URL + model

cloud sign-in + model picker + budget caps

5.8 Audit (/audit)

Pull last 30 days by default

Detail drawer shows JSON prettified

6) Styling (Tailwind Guidelines)

To keep it professional without heavy UI kits:

Use a restrained palette (neutral + 1 accent)

Use consistent spacing scale

Use cards with subtle borders/shadows

Badges for confidence with clear contrast

Avoid dense tables; use row padding and truncation with tooltips

If you want a kit: Skeleton UI (Svelte + Tailwind) is a sensible add, but you can also keep it pure Tailwind for minimal dependency.

7) Keeping the “Upgrade to React” option open

To ensure you can swap Svelte later without pain:

Keep API contracts stable (already defined)

Keep all UI logic in the frontend; do not bake UI assumptions into backend

Keep descriptions/rules/exports server-driven (so UI is replaceable)

Store “view preferences” (filters, last date) in localStorage only