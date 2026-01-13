# Directive D04: Raw Event Storage and Retention

**Goal:** Implement efficient writing of `raw_event` rows and enforcing the short-term retention policy.

**Scope:**
- `chroniclecore/internal/store` (Go package)
- `spec/schema.sql` (Reference)
- `execution/generate_fixtures.py` (Validation)

**Inputs:**
- `chroniclecore.db` (SQLite)

**Outputs:**
- Function `RecordEvent(e RawEvent) error`
- Function `PurgeOldEvents(context.Context, retentionDays int) (int64, error)`
- WAL Mode verified enabled.

## 1. Storage Implementation (Go)
- **WAL Mode:** Ensure `PRAGMA journal_mode = WAL;` is executed on connection open.
- **Synchronous:** `PRAGMA synchronous = NORMAL;` (Faster writes, safe enough for this app).
- **Batching:** If possible, buffer events in memory and flush every 5s or 100 events to reduce IOPS, OR just use standard `INSERT` (SQLite WAL is fast enough for 1 Hz).
- **Table:** Insert into `raw_event` (ts_start, app_id, title_id, etc).

## 2. Retention Logic
- Implement `PurgeOldEvents`.
- SQL: `DELETE FROM raw_event WHERE ts_start < datetime('now', '-14 days')`.
- This should be called:
  - On Startup.
  - After every successful Rollup job.

## 3. Maintenance (Vacuum)
- Implement `RunVacuum()`.
- SQL: `PRAGMA incremental_vacuum;` or `VACUUM;`.
- Policy: DO NOT run automatically. Expose as an Admin function or run only if DB size > 1GB (unlikely with this retention).

## Acceptance Criteria
- [ ] `RecordEvent` inserts a row.
- [ ] `PurgeOldEvents` removes rows older than 14 days.
- [ ] `execution/generate_fixtures.py` can insert 10,000 rows quickly.
