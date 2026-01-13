# Directive D10: Export Logic (CSV)

**Goal:** Generate CSV exports compatible with accounting software.

**Scope:**
- `chroniclecore/internal/api`
- `POST /api/v1/export/invoice-lines`

**Inputs:**
- Date Range (Start, End)
- Filter (Clients/Projects)

**Outputs:**
- A CSV file download.

## 1. Rounding Logic
- **Rules per Profile:**
  - `RoundingIncrement`: (e.g. 6 mins, 15 mins).
  - `MinimumBlock`: (e.g. 6 mins).
- **Algorithm:**
  - `RoundedDuration = Ceil(ActualDuration / Increment) * Increment`.
  - If `RoundedDuration < Min`, use `Min`.

## 2. CSV Columns
- Date
- Client
- Project
- Service
- Description (The deterministic string)
- Duration (Hours)
- Duration (Minutes)
- Rate
- Amount (Calculated)

## Acceptance Criteria
- [ ] Export requests returns `text/csv`.
- [ ] 5 min block rounded to 15m (if rule set).
- [ ] 61 min block rounded to 75m (if 15m rule) or 1.25h.
- [ ] `execution/verify_exports.py` passes.
