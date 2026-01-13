-- Migration: Rename currency to currency_code in rate table
-- Description: Updates the rate table to use currency_code instead of currency
--              for consistency with ISO 4217 naming conventions
-- Date: 2026-01-09

-- SQLite doesn't support direct column rename in all versions,
-- so we need to recreate the table

BEGIN TRANSACTION;

-- Create new rate table with currency_code
CREATE TABLE IF NOT EXISTS rate_new (
  rate_id            INTEGER PRIMARY KEY,
  name               TEXT NOT NULL,
  currency_code      TEXT NOT NULL DEFAULT 'USD',  -- ISO 4217 3-letter code
  hourly_minor_units INTEGER NOT NULL,
  effective_from     TEXT,
  effective_to       TEXT,
  is_active          INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0,1)),
  created_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

-- Copy data from old table to new (handles both currency and currency_code)
INSERT INTO rate_new (rate_id, name, currency_code, hourly_minor_units, effective_from, effective_to, is_active, created_at, updated_at)
SELECT
  rate_id,
  name,
  currency,
  hourly_minor_units,
  effective_from,
  effective_to,
  is_active,
  created_at,
  updated_at
FROM rate;

-- Drop old table
DROP TABLE rate;

-- Rename new table
ALTER TABLE rate_new RENAME TO rate;

-- Recreate trigger
CREATE TRIGGER IF NOT EXISTS trg_rate_updated_at
AFTER UPDATE ON rate
FOR EACH ROW
BEGIN
  UPDATE rate
     SET updated_at = (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
   WHERE rate_id = OLD.rate_id;
END;

-- Create indexes if needed
CREATE INDEX IF NOT EXISTS idx_rate_active ON rate(is_active);

COMMIT;
