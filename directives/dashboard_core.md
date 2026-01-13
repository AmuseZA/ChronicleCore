# Directive D09: Dashboard Core (Svelte)

**Goal:** Implement the primary "Today" view and App Shell.

**Scope:**
- `apps/chroniclecore-ui`
- `src/routes/`
- `src/lib/`

**Inputs:**
- `../workflow/ui_style_guide.md` (Visual Ref)
- `../workflow/frontend_architecture.md` (Components)

**Outputs:**
- A functional Web App running on `http://127.0.0.1:[PORT]`.

## 1. App Shell
- **Sidebar:** Navigation (Today, Review, Profiles, Settings).
- **Top Bar:** "Tracking Status" pill (Green/Amber), "Sync/Storage" status.
- **Layout:** Tailwind grid. Max-width container.

## 2. Today View (`/today`)
- **Components:**
  - `DateNavigator` (Day picker).
  - `Timeline` (List of Blocks).
  - `BlockRow` (Start-End, Duration, Title, Confidence Badge).
  - `BlockDrawer` (Edit details).
- **Interactions:**
  - Click row -> Open Drawer.
  - Toggle "Billable".
  - "Split Block" modal.
  - "Merge" selection mode.

## 3. State Management
- Use Svelte Stores (`writable`) for:
  - `trackingStatus` (Polling `/api/v1/tracking/status`).
  - `blocks` (List for current day).

## Acceptance Criteria
- [ ] UI loads and fetches blocks for Today.
- [ ] Can start/stop tracking via Top Bar.
- [ ] Clicking a block opens details.
- [ ] Visuals match Style Guide (Slate colors, rounded corners).
