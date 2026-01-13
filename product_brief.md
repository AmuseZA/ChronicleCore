Product Brief: Lightweight Windows Activity Tracker with Local Dashboard and Optional AI Enhancements
1) Product Summary

A standalone Windows application that automatically tracks a user’s work activity while the program is running, groups it into meaningful “work blocks,” assigns those blocks to invoiceable profiles (Client / Project / Service / Rate), and provides a clean localhost dashboard for review, corrections, reporting, and invoice-ready exports.

The product is local-first by default (no cloud required), with optional Local AI (LM Studio / Ollama) and optional Cloud AI (paid tier via your hosted gateway) for advanced descriptions, summarisation, and higher-accuracy classification—without compromising privacy or installing large models by default.

2) Target User and Primary Use Case

Primary user: A professional (e.g., admin/bookkeeper/agency operator) working independently on one Windows laptop.
Core workflow:

User launches the app at the start of work.

App tracks active apps and web domains (and optionally browser tab metadata via extension).

App groups activity into billable blocks and suggests Client/Service allocation.

User reviews and corrects quickly in the dashboard.

User exports invoice-ready time data (CSV/PDF) or generates line-item narratives.

3) Key Outcomes (What “Success” Looks Like)

Accurate client allocation with minimal manual input.

Invoice-ready outputs (rounded, minimum billing rules, narrative descriptions).

Low disk usage over long-term use via rollups and retention.

Strong privacy and security by design (local-only storage; sanitised AI inputs).

Sellable architecture with optional paid AI and integrations (without bloating installer).

4) Core Features (MVP)
4.1 Activity Capture (while app is running)

Captures at a configurable sampling interval (adaptive recommended):

Foreground application/process name (e.g., Word, Excel, Browser)

Active window title (with redaction controls)

Idle/AFK detection (keyboard/mouse inactivity)

Basic session events: Start/Stop tracking, Pause/Resume

Browser tracking:

Baseline: domain-level tracking (e.g., go.xero.com, mail.google.com)

Optional (recommended): browser extension for Chrome/Edge to send tab title/domain/URL metadata to the local app over localhost (secure token).

4.2 Work Block Aggregation (“Smart Grouping”)

Transforms raw events into meaningful time blocks:

Merges consecutive similar activity into a single block

Detects context switching and splits appropriately

Stitches related sequences (e.g., Email → Excel → Xero) into a single “work unit” when signals indicate the same client/workstream

4.3 Profiles and Billing Logic

Profiles represent invoice allocation:

Client

Project (optional)

Service Category (e.g., Bookkeeping, Admin, Reporting, Comms)

Rate (hourly)

Billing rules per profile/client:

rounding increments (e.g., 6/10/15 minutes)

minimum billable increments

billable vs non-billable defaults

4.4 Automatic Assignment Engine (Rules + Confidence)

A deterministic engine assigns blocks to profiles using:

Domain rules (e.g., go.xero.com → Xero category; optionally tenant mapping)

App rules (Word/Excel/Outlook)

Title keyword rules (VAT, reconcile, invoice, statement, quote)

Folder/file naming cues (where available via title, not file system scanning by default)

Every assignment gets a confidence score:

High (auto-approve candidate)

Medium (review suggested)

Low (goes to “Needs Review”)

4.5 Review Dashboard (Localhost Web UI)

Dashboard served on 127.0.0.1 only. App opens default browser to the dashboard on launch.

Core screens:

Timeline view (day/week): blocks with labels, durations, confidence

Needs Review inbox: low-confidence items first

Edit tools: split/merge/reassign/tag/add notes

Profiles management: clients/projects/services/rates/rules

Exports: invoice-ready CSV, timesheet summary, narrative line items

4.6 Invoice-Ready Description System (No-LLM baseline)

Generates clean professional line items using:

a controlled taxonomy (Client/Service/Activity/Outcome)

templates (deterministic phrasing)

block context (apps/domains/keywords)

consistent formatting rules (tone, tense, bullet vs sentence, etc.)

This ensures the product is useful without any AI.

5) Optional Enhancements (Post-MVP, Still Local-First)
5.1 Learning from Corrections (Non-LLM)

When a user reassigns a block, the system captures the correction as training data and improves future suggestions using lightweight approaches:

keyword-weighting updates

nearest-neighbour match on prior labelled blocks

optional small classical ML model (on-device) for classification only

5.2 Advanced Privacy Controls

Per-app tracking exclusions (denylist)

Title redaction patterns (regex)

“Private mode” pause toggle

Schedule-based tracking windows (optional)

6) AI Strategy (Local AI + Cloud AI, Pluggable Provider Layer)
6.1 Design Principle

AI is not required for core value. AI improves:

narrative quality

weekly summaries

classification accuracy in ambiguous cases

AI inputs must default to sanitised structured summaries, not raw titles/URLs.

6.2 Local AI Providers (User-Managed)

Support local providers without bundling model weights:

LM Studio local server (preferred Windows UX; OpenAI-compatible endpoint)

Ollama (local-only; commonly localhost)

Integration approach:

Provider detection (ping localhost endpoint)

Configurable base URL + model name

If unavailable, auto-fallback to deterministic template engine

6.3 Cloud AI (Paid Option via Your Gateway)

To generate revenue and keep keys secure:

App authenticates to your SaaS gateway

Gateway routes to external LLM providers (OpenAI, Anthropic, Google, etc.)

Users choose model from your allowed list in-app

Usage metering and plan limits enforced server-side

Two sellable modes:

Pro BYOK: user supplies their own API key(s) (optional tier)

Pro + Credits: user pays you; you provide metered AI usage via gateway

6.4 AI Use Cases

“Polish my invoice line items” from structured block summaries

“Weekly summary by client/service”

“Suggest profile assignment” for low-confidence blocks

“Detect duplicate/overlapping work units” and propose merges

7) Architecture Overview (Recommended)
7.1 Local Components (Installed)

Tracker App (User-launched)

collects OS-level activity

runs a local API + web server on 127.0.0.1 with random/fallback port

opens dashboard URL in browser

Local Database (SQLite)

stores raw events short-term

stores aggregated blocks long-term

(Optional) Browser Extension

sends domain/tab metadata to local API using a per-install secret token

No Docker required for single-laptop mode.

7.2 Cloud Components (Only for Paid Cloud AI)

Auth + Billing service

AI Gateway/router

Usage metering + budgets

Provider adapters (OpenAI/Anthropic/etc.)

8) Security and Privacy Requirements (Non-Negotiables)
8.1 Local Security

Local API bound to 127.0.0.1 only

Per-install secret token for extension/API calls

CSRF protections for dashboard interactions

Least privilege execution (no admin requirement for normal operation)

Encrypt sensitive settings at rest using Windows-native mechanisms (DPAPI recommended)

Clear audit of Pause/Resume actions

8.2 Data Handling Defaults

Default to domain-only tracking (not full URLs)

Window-title capture enabled with redaction tools and exclusions

No screenshots / keystrokes / content scraping by default

8.3 Cloud AI Safety

Never send raw activity logs by default

Sanitise before sending (client aliasing optional)

Enforce usage limits and budgets

No client-side embedded provider keys

9) Storage Management (Long-Term Lightweight Operation)
9.1 Retention Strategy

Raw events retained short-term (default: 14 days; configurable)

Aggregated blocks retained long-term (default: indefinite)

Periodic rollup job converts raw → blocks, then deletes expired raw events

9.2 Storage Optimisations

Dictionary tables to deduplicate repeated strings (titles/domains/apps)

Adaptive sampling (capture on change; slower during stable focus)

SQLite maintenance:

WAL/checkpoints

vacuum/compaction on a schedule or threshold

9.3 User Controls

Dashboard “Storage & Privacy” page:

current disk usage breakdown

retention settings

“Lightweight mode” toggle

“Purge now” control

10) Installer and Packaging (Windows)

Single installer (MSIX recommended for modern Windows packaging; alternatives possible)

Installs tracker app + local dashboard server + database location

First-run wizard:

create install ID + local secret

privacy defaults (domain-only, redaction options)

optional browser extension install prompt

optional “Enable Local AI” instructions (detect existing LM Studio/Ollama)

The app runs only when user launches it (per your requirement), with an optional “Launch at login” toggle as a future convenience feature.

11) Reporting and Exports

Outputs should be designed for real billing workflows:

By client/service/day/week/month

Rounding and minimum billing applied

Narrative line items generated (deterministic; AI-enhanced optional)

CSV export compatible with common accounting/invoicing workflows

PDF timesheet summary (optional post-MVP)

12) Product Differentiators (Positioning)

Invoice-first design (not generic time tracking)

Low-admin “Needs Review” workflow with confidence scoring

Local-first privacy with optional AI polish

AI-provider choice (Local LLMs or paid Cloud AI via gateway)

Lightweight over years via rollups and retention

13) Proposed Roadmap
Phase 1: MVP (Local-only, sellable foundation)

Windows activity capture (apps + titles + idle)

Aggregation into blocks

Profiles + rules + confidence scoring

Localhost dashboard (review, split/merge, reassign)

Invoice-ready export + deterministic narrative templates

Storage rollup + retention + maintenance

Phase 2: Quality and Automation

Browser extension for richer metadata (Edge/Chrome)

Learning-from-corrections (non-LLM)

More robust rule builder (regex, priorities, per-client mapping)

Phase 3: Local AI

Provider layer: LM Studio + Ollama

AI for narrative polishing and weekly summaries (sanitised inputs)

Phase 4: Cloud AI (Paid)

Auth/billing

AI gateway + metering

Model selection UI and budgets

Tiering (BYOK and Credits plans)

14) Commercial Model (Aligned to Your Goals)

Free / Basic: local tracking + basic profiles + exports

Pro (Subscription): advanced rules, learning, automation, premium reporting

Pro + AI Credits: includes cloud AI via your gateway (model choice + usage limits)

Pro BYOK (optional): user supplies provider keys; you charge for premium product features

15) Deliverables for the Next Step (If You Proceed)

From this brief, the next artifacts to produce are:

Functional specification (screens, workflows, acceptance criteria)

Data model (raw events, blocks, profiles, rules, audit)

Local API contract (dashboard + extension + providers)

AI provider interface spec (local + cloud)

Installer/packaging plan (Windows deployment + update strategy)

Security threat model + mitigation checklist