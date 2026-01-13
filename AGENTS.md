ChronicleCore DOE/SOP (Built on Your Default)
Agent Instructions (ChronicleCore Edition)

This file is mirrored across CLAUDE.md, AGENTS.md, and GEMINI.md so the same instructions load in any AI environment.

You operate within a 3-layer architecture that separates concerns to maximize reliability. ChronicleCore is a local-first Windows application with deterministic business logic (tracking, aggregation, assignment, invoicing) and optional AI. Your job is to keep the build reliable by pushing complexity into deterministic code and structured directives, while using LLMs primarily for orchestration, planning, review, and documentation.

The 3-Layer Architecture
Layer 1: Directive (What to do)

SOPs in Markdown live in directives/

Define goals, scope, constraints, inputs, outputs, tools/scripts, edge cases, acceptance criteria

Written like instructions to a mid-level engineer

ChronicleCore directives must be feature-scoped, not broad: e.g. directives/track_windows_foreground.md, not directives/build_app.md.

Layer 2: Orchestration (Decision making)

This is you (the LLM).

You route tasks to deterministic scripts, validate outputs, and maintain continuity across backend/frontend/spec.

You enforce constraints: Windows-only, local-first, no sensitive capture by default, retention, localhost-only server.

You keep scope tight: MVP first, optional AI later.

Layer 3: Execution (Doing the work)

Deterministic code in execution/ (Python by default; Go/Rust tooling can live under tools/ if needed)

Environment variables and tokens in .env (cloud AI keys never live in client—only in future gateway)

Scripts perform:

Schema validation

OpenAPI schema generation/checking

Contract tests against localhost API

UI build checks

Release packaging validation steps

Test-data generation

Lint/format/test runners

Scripts are reliable, testable, and repeatable.

Why this works for ChronicleCore:
ChronicleCore depends on deterministic correctness (time blocks, billing rounding, retention, security). LLMs can help plan and document, but correctness must be enforced by tests and scripts.

ChronicleCore Operating Principles
1) Contract-first, local-first

All UI work is driven by the local API contract and the SQLite schema.

The backend must bind only to 127.0.0.1 and must not expose data remotely.

The frontend must never assume cloud availability.

2) Deterministic baseline before AI

The deterministic assignment engine + template description engine must be production-grade before any AI is introduced.

AI features are opt-in enhancements:

Local AI: LM Studio / Ollama (user-installed)

Cloud AI: later paid option via gateway only (no API keys in desktop client)

3) Privacy and security by default

No keystrokes, no screenshots, no full URLs by default.

Title capture must support exclusions + redaction.

Browser extension uses a per-install secret token.

Sensitive config is encrypted at rest (Windows DPAPI).

4) Storage stays small

Raw events are short-term only; blocks are long-term.

Rollup + retention is mandatory for MVP.

SQLite maintenance (VACUUM policy) is defined and testable.

5) Self-anneal when things break

When errors occur:

Read stack trace

Fix deterministic tools/tests first

Re-run tests

Update the directive with the failure mode + fix

The system gets stronger

Repository/Workspace Conventions (ChronicleCore)
Directories

directives/ — SOPs for each feature and workflow

execution/ — deterministic scripts (validation, test generation, contract checks)

apps/chronicalcore-ui/ — Svelte + Vite + Tailwind dashboard

apps/chronicalcore-core/ — backend + tracker core (Go or Rust)

extensions/chronicalcore-browser/ — Chrome/Edge extension (optional in MVP, planned in Phase 2)

.tmp/ — intermediates: generated fixtures, test exports, logs; never committed

spec/ — canonical technical spec artifacts:

spec/schema.sql (DDL)

spec/api-contract.md (endpoints + examples)

spec/ui-wire.md (wire-level spec)

spec/acceptance-tests.md

spec/threat-model.md

Deliverables vs Intermediates

Deliverables: build artifacts (installer, release zip), docs, exported sample outputs

Intermediates: any generated fixture, logs, local DB copies, export previews in .tmp/

ChronicleCore Directive Standards

Every directive must include these headings:

Goal

Scope (In/Out)

Inputs

Outputs

Constraints (privacy/security/storage)

Dependencies (other directives, scripts, modules)

Procedure (steps)

Acceptance Criteria (testable, explicit)

Edge Cases

Execution Tools (scripts to run)

Update Notes (what to append when self-annealing)

ChronicleCore DOE (Definition of Execution) Workflow
Phase Gate Model

Work is executed in gated phases. Each phase has a Definition of Done (DoD) and validation scripts.

Phase 0: Project Bootstrap

Objective: repo scaffolding, spec pinning, CI skeleton.

Freeze canonical spec files in spec/

Confirm schema DDL is stored as spec/schema.sql

Establish backend stack decision (choose Go or Rust) and document why

Set up execution/ scripts for:

schema checks

API contract checks

UI build checks

DoD: CI can run “lint + tests + UI build + schema validation”

Phase 1 (MVP): Local Tracking + Timeline + Profiles + Exports

Objective: sellable local-first MVP, no extension required.
Must include:

Foreground app/window tracking, idle detection

Raw events → aggregated blocks

Profiles CRUD

Rules engine v1

Deterministic description engine

Localhost dashboard (Svelte)

Export wizard (CSV)

Retention + cleanup + SQLite maintenance

DoD: acceptance tests pass + export outputs correct + storage remains stable

Phase 2: Browser Extension + Rule Builder Depth

Objective: domain-level browser context, improved assignments.

Extension ingestion endpoint + secret token

Domain-based rules

Rule test panel and better UX

Phase 3: Local AI

Objective: AI polish without cloud.

LM Studio/Ollama provider detection

Sanitised structured prompts only

Toggle-based AI features

Phase 4: Paid Cloud AI Gateway

Objective: monetisation + model choice.

Auth + subscription + metering

Model selection UI

Gateway routing to providers

No client-stored provider keys

Mandatory Deterministic Tooling (Execution Layer)

Create and maintain these scripts early:

1) Schema Validation

execution/validate_schema.py

Confirms spec/schema.sql executes cleanly on SQLite

Verifies required tables/indexes/triggers exist

Outputs a schema fingerprint hash for version tracking

2) API Contract Validator

execution/validate_api_contract.py

Checks that backend exposes required endpoints

Calls /health, /tracking/status, /blocks, core write endpoints

Verifies response JSON shapes against a stored JSON schema

Runs as a local integration test (localhost-only)

3) Fixture Generator

execution/generate_fixtures.py

Generates realistic raw_event sequences for:

Office work

Browser work (domain known/unknown)

Idle segments

Frequent context switching

Produces .tmp/fixtures/*.json and optionally seeds a .tmp/test.db

4) Export Verification

execution/verify_exports.py

Runs export endpoints and checks:

rounding increment rules

minimum billing rules

totals match timeline totals

unassigned warnings behave correctly

Produces .tmp/exports/*.csv for regression snapshots

5) Privacy Regression Checks

execution/privacy_checks.py

Confirms settings defaults:

domain-only ON

full URL OFF

no screenshot/keystroke features exist

Validates redaction rules apply to stored titles in dict_title

6) UI Build & Smoke Test

execution/ui_build_check.py

Runs Svelte build

Optionally runs Playwright smoke tests later (Phase 2+)

Orchestration Rules (How the Planning Agent Should Behave)
A) Always start with directives

Before suggesting implementation steps, identify which directive applies.
If directive doesn’t exist:

Create a new directive draft in the response (do not overwrite existing ones without instruction).

Add the directive to a “Directives to Add” list.

B) Write plans as handoff-ready

Plans must include:

File/module map

Endpoint list + request/response examples

DB touchpoints

Tests and scripts to run

Acceptance criteria

C) No ambiguous handoffs

If something is undecided (Go vs Rust), decide based on constraints:

Lightweight, single binary, Windows friendly, developer availability.
Document the decision rationale and proceed.

ChronicleCore DoD (Definition of Done) Checklist — MVP

A change is “done” only if:

Build & Run

Backend builds to a single Windows binary

UI builds and is served by backend

App runs and opens dashboard via localhost

Security & Privacy

Server binds to 127.0.0.1 only

No sensitive capture features exist by default

Extension ingestion requires per-install secret (even if extension not shipped yet)

DPAPI encryption used for secrets/config

Storage

Raw retention works (default 14 days)

Rollup runs and purges old raw events

VACUUM policy defined and tested

DB size remains stable under fixture load tests

Functionality

Blocks created correctly with idle segments excluded from billable totals

Profiles CRUD works

Rules engine assigns profile with confidence scoring

Manual reassign/split/merge/lock works and writes audit logs

Exports produce correct CSV outputs with rounding/min rules

Tooling

execution/validate_schema.py passes

execution/validate_api_contract.py passes

execution/verify_exports.py passes

UI build check passes

Recommended Directives to Create (ChronicleCore)

Create these as individual files in directives/:

directives/bootstrap_repo.md

directives/backend_stack_decision.md

directives/windows_activity_capture.md

directives/idle_detection.md

directives/raw_event_storage.md

directives/block_aggregation.md

directives/profile_crud.md

directives/rules_engine.md

directives/deterministic_description_engine.md

directives/dashboard_today_view.md

directives/dashboard_needs_review.md

directives/dashboard_exports.md

directives/retention_and_maintenance.md

directives/local_api_security.md

directives/audit_logging.md

directives/browser_extension_phase2.md

directives/local_ai_phase3.md

directives/cloud_ai_gateway_phase4.md