# Directive D08: Deterministic Description Templates

**Goal:** Generate "Invoice Ready" line items from Activity Blocks without using AI.

**Scope:**
- `chroniclecore/internal/engine` (Go package)
- `block.description` field

**Inputs:**
- Block data (App, Title, Profile, Duration)

**Outputs:**
- A string `description` written to the block.

## 1. Template Engine
- Implement a simple string interpolator.
- **Templates:**
  - Default: `"{App}: {Title}"`
  - Per-Profile Overrides: Allow user to set specific templates per Profile? (MVP: Global default is fine, maybe "By Service" later).
- **Sanitization:**
  - If title is redacted, use "Work on {App}".
  - If multiple small events merged: "Multiple Tasks in {App}".

## 2. Logic
- When a Block is finalized (at end of Rollup or User Edit):
  - Generate Description.
  - Write to `block.description`.
- **Constraint:** Description MUST be deterministic. Same input -> Same output.

## Acceptance Criteria
- [ ] Block "Chrome - Google" -> Desc "Chrome: Google".
- [ ] Block "Word - Contract" -> Desc "Word: Contract".
- [ ] Edit tests: changing block content updates description (unless locked).
