ChronicleCore UI Style Guide (Svelte + Tailwind v1)
1) Design Principles

Clarity over decoration: tables and structured layouts first; no heavy charts in v1.

High information density, low visual noise: whitespace + subtle separators.

Action-first: “Needs Review” and “Export readiness” should always be obvious.

Trust + privacy cues: UI should visually communicate local-first and privacy-safe defaults.

2) Layout System
2.1 Grid & Page Width

Max content width: max-w-7xl (or max-w-[1200px]) for readability.

Page padding: px-4 sm:px-6 lg:px-8

Vertical spacing between sections: space-y-6 (tight: space-y-4)

2.2 App Shell

Sidebar width: w-64 (collapsed w-16)

Topbar height: h-14 or h-16

Primary layout: sidebar fixed, content scrolls.

2.3 Cards and Panels

Use cards sparingly; the dominant pattern is table inside a card.

Card baseline:

rounded-xl

border border-slate-200

bg-white

shadow-sm (never heavy shadows)

3) Typography
3.1 Font

Use a system-friendly sans stack:

font-sans (Tailwind default stack is fine)

Avoid mixing families in v1.

3.2 Type Scale

Page title: text-xl font-semibold text-slate-900

Section title: text-sm font-semibold text-slate-900 tracking-tight

Table header: text-xs font-medium text-slate-500 uppercase tracking-wide

Body text: text-sm text-slate-700

Muted/meta text: text-xs text-slate-500

3.3 Numbers

Time and currency should be:

tabular-nums for alignment

Slightly stronger colour: text-slate-900 font-medium for primary totals

4) Colour & Status Language (Minimal Palette)
4.1 Base Neutrals

Background: bg-slate-50

Surface: bg-white

Border: border-slate-200

Primary text: text-slate-900

Secondary text: text-slate-600

Muted: text-slate-500

4.2 Semantic Colours (Use sparingly)

Success / Reviewed: green

Warning / Needs Review: amber

Error / Unassigned / Blocked: red

Info: blue (only for informational chips/links)

Rule: No more than two semantic colours visible in a single table row.

5) Core UI Components Styling
5.1 Tables (Primary Pattern)

Tables should feel like your references: wide rows, subtle dividers, easy scanning.

Table Container

rounded-xl border border-slate-200 bg-white overflow-hidden

Table Header Row

bg-slate-50

text-xs font-medium text-slate-500 uppercase tracking-wide

Divider: border-b border-slate-200

Table Rows

Height/padding: py-4 (compact: py-3)

Row divider: border-b border-slate-100

Hover: hover:bg-slate-50

Selected row: bg-slate-50 ring-1 ring-slate-200 (subtle)

Column Alignment

Left align labels; right align numeric totals.

Use tabular-nums on numeric cells.

5.2 Avatars (Initials)

Circle: h-10 w-10 rounded-full

Background: neutral tint or soft category tint

Initials: text-sm font-semibold text-slate-700

In tables: align avatar + name in a single cell with flex items-center gap-3

Colour approach:

Deterministic colour per client/profile using a small palette of soft background tints:

e.g. bg-rose-100, bg-amber-100, bg-lime-100, bg-sky-100, bg-violet-100

Keep text neutral (don’t colour the initials heavily).

5.3 Badges and Chips
Confidence Badge (HIGH/MEDIUM/LOW)

Size: text-xs font-medium px-2 py-1 rounded-md

HIGH: neutral-success styling (subtle)

MEDIUM: neutral-warning styling (subtle)

LOW: warning/error lean, but not alarming

Context Chips (App/Domain)

text-xs px-2 py-1 rounded-md border border-slate-200 bg-white text-slate-600

Keep them uniform; do not add icons unless highly recognisable (Word/Excel/Browser).

“Needs Review” Pill

More prominent than confidence:

bg-amber-50 text-amber-800 border border-amber-200

5.4 Progress Bars (Reviewed / Unreviewed / Unassigned)

Inspired by the capacity bars in your reference.

Track: h-2 rounded-full bg-slate-100 overflow-hidden

Segments:

Reviewed (green-ish)

Unreviewed (slate)

Unassigned (amber/red-ish depending on severity)

Always include a small legend in tooltips or a condensed label.

5.5 Buttons
Primary button (rare)

Used for: “Generate Export”, “Save Profile”, “Start Tracking”

rounded-lg px-3 py-2 text-sm font-medium

Solid fill (single accent)

Secondary button (default)

border border-slate-200 bg-white text-slate-700 hover:bg-slate-50

Danger button (rare)

For destructive actions only (purge, reset)

Must require confirmation modal.

Icon buttons

h-9 w-9 rounded-md border border-slate-200 bg-white hover:bg-slate-50

5.6 Inputs

Text input: rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm

Focus ring: focus:outline-none focus:ring-2 focus:ring-slate-300

Dropdown: same style as input

Search inputs should have a subtle icon and placeholder:text-slate-400

5.7 Drawers and Modals
Drawer (Block Detail)

Right-side slide in

Width: w-[420px] desktop; full screen on mobile

Header: title + close

Tabs: low-height, simple underline active state

Content uses space-y-4 and avoids nested cards

Modal (Split/Merge)

Max width: max-w-lg

Footer actions aligned right

Confirm is primary; cancel is secondary

6) Screen-Specific Style Rules
6.1 Today / Timeline

Timeline rows should be slightly taller (py-4) for readability.

Context chips must wrap cleanly; avoid multi-line chaos by truncating with tooltips.

6.2 Needs Review

Items should be visually “queue-like”:

Leading indicator dot or small badge

Stronger “Assign” CTA within each row

Selected state for batch actions must be obvious but subtle (no neon highlight).

6.3 Calendar (Week View)

Keep the heatmap understated:

small cells with rounded corners

intensity increases but remains within soft tints

Day headers: text-xs font-medium text-slate-500 uppercase

6.4 Exports

Wizard steps should be clean and linear:

“Step 1 / Step 2 / Step 3”

Export readiness panel should use:

green/amber indicators with clear language

no aggressive red unless export would be materially wrong

7) Icons & Visuals

Keep icons minimal:

Play/Pause

Calendar

Filter

Download

Lock

Warning

Use one icon set consistently (e.g., Lucide style equivalents—Svelte ports exist, or inline SVGs).

8) Density Modes (Optional Toggle)

Offer two density modes for different user preferences:

Comfortable (default): py-4, more whitespace

Compact: py-3, smaller chips

Store preference in localStorage only.

9) Microcopy & Tone (Trust-building)

Use short, clear, non-technical labels:

“Needs Review”

“Unassigned”

“Reviewed”

“Domain-only mode (recommended)”

“Titles captured: On (with redaction)”

Avoid:

“Telemetry”

“Monitoring”

“Surveillance”

10) Minimal Tailwind Token Set (Consistency)

Define consistent utility “recipes” (either as Tailwind component classes or simple shared strings):

card: rounded-xl border border-slate-200 bg-white shadow-sm

tableWrap: rounded-xl border border-slate-200 bg-white overflow-hidden

tableHead: bg-slate-50 text-xs font-medium text-slate-500 uppercase tracking-wide

btnPrimary: rounded-lg px-3 py-2 text-sm font-medium

btnSecondary: rounded-lg px-3 py-2 text-sm font-medium border border-slate-200 bg-white text-slate-700 hover:bg-slate-50

input: rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-300

11) Recommended v1 Theme (Default)

Light theme by default (matches your references and feels “admin-grade”)

Dark theme can be a v2 enhancement—dark themes are easy to get wrong and add complexity.