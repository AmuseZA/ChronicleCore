# Directive D05: Block Aggregation Engine

**Goal:** Convert high-frequency `raw_event` data into consolidated `block` rows for billing.

**Scope:**
- `chroniclecore/internal/engine` (Go package)
- `rollup_job` (Routine)

**Inputs:**
- `raw_event` table
- `rules` (for assignment - see D06)

**Outputs:**
- Populated `block` table.
- "Rollup" log/results.

## 1. Aggregation Logic
- **Trigger:** Runs every 5 minutes.
- **Process:**
  1. Select un-aggregated `raw_event` rows (or process by time window).
  2. **Merge:** Combine sequential events where `AppID` and `TitleID` (and optionally `DomainID`) match.
  3. **Gap Handling:** If gap > `IdleThreshold`, insert GAP (or simply end block).
  4. **Persist:** Write `block` row.
  5. **Mark:** Note raw events as "processed" (or simply use time-watermark). *preferred: Time-watermark (e.g., "aggregated_until" timestamp).*

## 2. Smart Grouping Heuristics
- **Context Switch Filtering:** If user flips to "Spotify" for 5 seconds then back to "Word", ignore the Spotify glitch?
  - *Decision:* MVP = Strict logging. If 5s Spotify, it's a 5s block.
  - Users can merge later in UI.
- **Minimum Block Size:** Defaults. MVP = No minimum (log everything).

## 3. Rollup Job
- Implement `StartRollupLoop(ctx, interval)`.
- Respects App "Paused" state (don't aggregate if tracking is paused? No, aggregate whatever is in DB).

## Acceptance Criteria
- [ ] 10 sequential "Chrome" events (1s each) become 1 "Chrome" block (10s duration).
- [ ] Alternating "Chrome" -> "Word" -> "Chrome" creates 3 blocks.
- [ ] Rollup runs successfully without errors.
