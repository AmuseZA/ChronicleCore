# Directive D06: Rules Engine for Assignment

**Goal:** Assign `ProfileID` (Client/Project) to Blocks based on Rules.

**Scope:**
- `chroniclecore/internal/engine` (Go package)
- `rule` table

**Inputs:**
- `block` (App, Title, Domain)
- `rule` (Match patterns)

**Outputs:**
- `block.profile_id` update.
- `block.confidence` update.

## 1. Match Logic (MVP)
- **Priority:** Rule with highest `priority` wins.
- **Match Types:**
  - `APP`: Exact match on Process Name (e.g. `WINWORD.EXE`).
  - `TITLE_REGEX`: Regex match on Window Title.
  - `DOMAIN`: Exact match (Phase 2 - stub for now).
- **Evaluation:**
  - When a block is created (during Rollup), fetch all `enabled` rules ordered by Priority DESC.
  - Iterate:
    - If `APP` matches: Apply Profile.
    - If `TITLE_REGEX` matches: Apply Profile.
  - Set `Confidence`:
    - "HIGH" if explicit rule match.
    - "LOW" if no match (Unassigned).

## 2. Auto-Assignment
- Function `ApplyRules(block *Block)`.
- If a block is manually assigned by user (`locked=1`), SKIP rules.

## Acceptance Criteria
- [ ] Rule "Word -> Client A" exists.
- [ ] New Block "WINWORD.EXE" triggers assignment to Client A (Confidence HIGH).
- [ ] Unknown Block stays Unassigned (Confidence LOW).
