-- ============================================================
-- SQLite Schema DDL (v1)
-- Product: Windows Activity Tracker + Localhost Dashboard
-- Notes:
--  - Timestamps are stored as ISO-8601 TEXT in UTC (recommended).
--  - Enable foreign keys in every connection: PRAGMA foreign_keys = ON;
--  - For best performance: use WAL mode at runtime (PRAGMA journal_mode = WAL).
-- ============================================================

PRAGMA foreign_keys = ON;

-- ----------------------------
-- Install / Settings
-- ----------------------------

CREATE TABLE IF NOT EXISTS install (
  install_id   TEXT PRIMARY KEY,                  -- UUID-like string generated on first run
  created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  app_version  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS settings (
  key          TEXT PRIMARY KEY,
  value        BLOB NOT NULL,                      -- store JSON/text/bytes; encrypt sensitive values using DPAPI before writing
  is_encrypted INTEGER NOT NULL DEFAULT 0 CHECK (is_encrypted IN (0,1)),
  updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TRIGGER IF NOT EXISTS trg_settings_updated_at
AFTER UPDATE ON settings
FOR EACH ROW
BEGIN
  UPDATE settings
     SET updated_at = (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
   WHERE key = OLD.key;
END;

-- ----------------------------
-- Dictionary tables (de-dup strings)
-- ----------------------------

CREATE TABLE IF NOT EXISTS dict_app (
  app_id    INTEGER PRIMARY KEY,
  app_name  TEXT NOT NULL UNIQUE                   -- e.g. "WINWORD.EXE", "EXCEL.EXE", "msedge.exe"
);

CREATE TABLE IF NOT EXISTS dict_title (
  title_id    INTEGER PRIMARY KEY,
  title_text  TEXT NOT NULL UNIQUE                 -- may store redacted text if configured
);

CREATE TABLE IF NOT EXISTS dict_domain (
  domain_id    INTEGER PRIMARY KEY,
  domain_text  TEXT NOT NULL UNIQUE                -- e.g. "go.xero.com"
);

-- ----------------------------
-- Billing / Profile entities
-- ----------------------------

CREATE TABLE IF NOT EXISTS client (
  client_id   INTEGER PRIMARY KEY,
  name        TEXT NOT NULL UNIQUE,
  is_active   INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0,1)),
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TRIGGER IF NOT EXISTS trg_client_updated_at
AFTER UPDATE ON client
FOR EACH ROW
BEGIN
  UPDATE client
     SET updated_at = (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
   WHERE client_id = OLD.client_id;
END;

CREATE TABLE IF NOT EXISTS project (
  project_id  INTEGER PRIMARY KEY,
  client_id   INTEGER NOT NULL,
  name        TEXT NOT NULL,
  is_active   INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0,1)),
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  FOREIGN KEY (client_id) REFERENCES client(client_id) ON DELETE CASCADE,
  UNIQUE (client_id, name)
);

CREATE TRIGGER IF NOT EXISTS trg_project_updated_at
AFTER UPDATE ON project
FOR EACH ROW
BEGIN
  UPDATE project
     SET updated_at = (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
   WHERE project_id = OLD.project_id;
END;

CREATE TABLE IF NOT EXISTS service (
  service_id  INTEGER PRIMARY KEY,
  name        TEXT NOT NULL UNIQUE,                -- e.g. "Bookkeeping", "Admin", "Reporting", "Client Comms"
  is_active   INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0,1)),
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TRIGGER IF NOT EXISTS trg_service_updated_at
AFTER UPDATE ON service
FOR EACH ROW
BEGIN
  UPDATE service
     SET updated_at = (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
   WHERE service_id = OLD.service_id;
END;

CREATE TABLE IF NOT EXISTS rate (
  rate_id            INTEGER PRIMARY KEY,
  name               TEXT NOT NULL,                -- e.g. "Standard", "After-hours"
  currency_code      TEXT NOT NULL DEFAULT 'USD',  -- ISO 4217 3-letter code (e.g., USD, ZAR, EUR)
  hourly_minor_units INTEGER NOT NULL,             -- store in cents/minor units for precision
  effective_from     TEXT,                         -- nullable; ISO-8601
  effective_to       TEXT,                         -- nullable; ISO-8601
  is_active          INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0,1)),
  created_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
  -- Note: UNIQUE constraint relaxed - enforce uniqueness in application layer if needed
);

CREATE TRIGGER IF NOT EXISTS trg_rate_updated_at
AFTER UPDATE ON rate
FOR EACH ROW
BEGIN
  UPDATE rate
     SET updated_at = (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
   WHERE rate_id = OLD.rate_id;
END;

CREATE TABLE IF NOT EXISTS profile (
  profile_id  INTEGER PRIMARY KEY,
  client_id   INTEGER NOT NULL,
  project_id  INTEGER,                             -- optional
  service_id  INTEGER NOT NULL,
  rate_id     INTEGER NOT NULL,
  name        TEXT,                                -- optional display name override
  is_active   INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0,1)),
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  FOREIGN KEY (client_id)  REFERENCES client(client_id)   ON DELETE CASCADE,
  FOREIGN KEY (project_id) REFERENCES project(project_id) ON DELETE SET NULL,
  FOREIGN KEY (service_id) REFERENCES service(service_id) ON DELETE RESTRICT,
  FOREIGN KEY (rate_id)    REFERENCES rate(rate_id)       ON DELETE RESTRICT
  -- Note: UNIQUE constraint relaxed - enforce uniqueness in application layer if needed
);

CREATE TRIGGER IF NOT EXISTS trg_profile_updated_at
AFTER UPDATE ON profile
FOR EACH ROW
BEGIN
  UPDATE profile
     SET updated_at = (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
   WHERE profile_id = OLD.profile_id;
END;

-- ----------------------------
-- Rules engine
-- ----------------------------

CREATE TABLE IF NOT EXISTS rule (
  rule_id            INTEGER PRIMARY KEY,
  name               TEXT NOT NULL,
  priority           INTEGER NOT NULL DEFAULT 0,    -- higher wins
  match_type         TEXT NOT NULL CHECK (match_type IN (
                       'APP', 'DOMAIN', 'TITLE_REGEX', 'KEYWORD', 'COMPOSITE'
                     )),
  match_value        TEXT NOT NULL,                -- pattern/regex/json per match_type
  target_profile_id  INTEGER NOT NULL,
  target_service_id  INTEGER,                       -- optional override
  confidence_boost   INTEGER NOT NULL DEFAULT 0,     -- can be negative or positive
  enabled            INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0,1)),
  created_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  FOREIGN KEY (target_profile_id) REFERENCES profile(profile_id) ON DELETE CASCADE,
  FOREIGN KEY (target_service_id) REFERENCES service(service_id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_rule_enabled_priority
  ON rule (enabled, priority DESC);

CREATE TRIGGER IF NOT EXISTS trg_rule_updated_at
AFTER UPDATE ON rule
FOR EACH ROW
BEGIN
  UPDATE rule
     SET updated_at = (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
   WHERE rule_id = OLD.rule_id;
END;

-- ----------------------------
-- Raw events (short-term)
-- ----------------------------

CREATE TABLE IF NOT EXISTS raw_event (
  event_id        INTEGER PRIMARY KEY,
  ts_start        TEXT NOT NULL,                    -- ISO-8601 UTC
  ts_end          TEXT,                             -- ISO-8601 UTC (nullable while open)
  app_id          INTEGER NOT NULL,
  title_id        INTEGER,                          -- nullable if redacted/disabled
  domain_id       INTEGER,                          -- nullable; requires extension or other capture method
  state           TEXT NOT NULL CHECK (state IN ('ACTIVE','IDLE','PAUSED')),
  source          TEXT NOT NULL CHECK (source IN ('OS','EXTENSION')),
  metadata        TEXT,                             -- JSON payload for advanced tracking (activity score, browser context)
  hash_signature  TEXT,                             -- optional: quick grouping signature (e.g., app|domain|title)
  created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  FOREIGN KEY (app_id)    REFERENCES dict_app(app_id)       ON DELETE RESTRICT,
  FOREIGN KEY (title_id)  REFERENCES dict_title(title_id)   ON DELETE SET NULL,
  FOREIGN KEY (domain_id) REFERENCES dict_domain(domain_id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_raw_event_ts_start
  ON raw_event (ts_start);

CREATE INDEX IF NOT EXISTS idx_raw_event_app_ts
  ON raw_event (app_id, ts_start);

CREATE INDEX IF NOT EXISTS idx_raw_event_domain_ts
  ON raw_event (domain_id, ts_start);

CREATE INDEX IF NOT EXISTS idx_raw_event_state_ts
  ON raw_event (state, ts_start);

-- ----------------------------
-- Aggregated blocks (long-term)
-- ----------------------------

CREATE TABLE IF NOT EXISTS block (
  block_id           INTEGER PRIMARY KEY,
  ts_start           TEXT NOT NULL,
  ts_end             TEXT NOT NULL,
  primary_app_id     INTEGER NOT NULL,
  primary_domain_id  INTEGER,                       -- nullable
  title_summary_id   INTEGER,                       -- nullable; preferably redacted/summary form
  profile_id         INTEGER,                       -- nullable if unassigned
  confidence         TEXT NOT NULL DEFAULT 'LOW' CHECK (confidence IN ('HIGH','MEDIUM','LOW')),
  billable           INTEGER NOT NULL DEFAULT 1 CHECK (billable IN (0,1)),
  locked             INTEGER NOT NULL DEFAULT 0 CHECK (locked IN (0,1)),
  notes              TEXT,
  description        TEXT,                          -- final invoice line item text (deterministic or AI-polished)
  metadata           TEXT,                          -- JSON payload (e.g. {"avg_activity_score": 0.9})
  created_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  FOREIGN KEY (primary_app_id)    REFERENCES dict_app(app_id)       ON DELETE RESTRICT,
  FOREIGN KEY (primary_domain_id) REFERENCES dict_domain(domain_id) ON DELETE SET NULL,
  FOREIGN KEY (title_summary_id)  REFERENCES dict_title(title_id)   ON DELETE SET NULL,
  FOREIGN KEY (profile_id)        REFERENCES profile(profile_id)    ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_block_ts_start
  ON block (ts_start);

CREATE INDEX IF NOT EXISTS idx_block_profile_ts
  ON block (profile_id, ts_start);

CREATE INDEX IF NOT EXISTS idx_block_confidence_ts
  ON block (confidence, ts_start);

CREATE INDEX IF NOT EXISTS idx_block_billable_ts
  ON block (billable, ts_start);

CREATE TRIGGER IF NOT EXISTS trg_block_updated_at
AFTER UPDATE ON block
FOR EACH ROW
BEGIN
  UPDATE block
     SET updated_at = (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
   WHERE block_id = OLD.block_id;
END;

-- ----------------------------
-- Audit log (user/system actions)
-- ----------------------------

CREATE TABLE IF NOT EXISTS audit_log (
  audit_id     INTEGER PRIMARY KEY,
  ts           TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  actor        TEXT NOT NULL DEFAULT 'USER' CHECK (actor IN ('USER','SYSTEM')),
  action       TEXT NOT NULL,                       -- e.g., "REASSIGN_BLOCK", "SPLIT_BLOCK", "PAUSE_TRACKING"
  details_json TEXT                                -- JSON payload: before/after, ids, parameters
);

CREATE INDEX IF NOT EXISTS idx_audit_ts
  ON audit_log (ts);

CREATE INDEX IF NOT EXISTS idx_audit_action_ts
  ON audit_log (action, ts);

-- ----------------------------
-- Helpful Views (optional but recommended)
-- ----------------------------

-- Duration in minutes and hours for blocks, computed on demand.
-- NOTE: ISO-8601 parsing assumes 'Z' UTC format. Ensure consistent timestamp formatting.
CREATE VIEW IF NOT EXISTS v_block_duration AS
SELECT
  b.block_id,
  b.ts_start,
  b.ts_end,
  b.profile_id,
  b.confidence,
  b.billable,
  b.locked,
  b.description,
  -- duration in seconds
  (strftime('%s', b.ts_end) - strftime('%s', b.ts_start)) AS duration_seconds,
  -- duration in minutes (float)
  (strftime('%s', b.ts_end) - strftime('%s', b.ts_start)) / 60.0 AS duration_minutes,
  -- duration in hours (float)
  (strftime('%s', b.ts_end) - strftime('%s', b.ts_start)) / 3600.0 AS duration_hours
FROM block b;

-- ----------------------------
-- ML Pipeline Tables (Option C+)
-- ----------------------------

-- Model Registry: Track trained model versions
CREATE TABLE IF NOT EXISTS ml_model_registry (
  model_id          INTEGER PRIMARY KEY,
  model_type        TEXT NOT NULL CHECK (model_type IN ('PROFILE_CLASSIFIER', 'SESSION_CLUSTERER')),
  version           TEXT NOT NULL,
  algorithm         TEXT NOT NULL,
  metrics_json      TEXT,
  status            TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'ARCHIVED', 'FAILED')),
  trained_at        TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  trained_samples   INTEGER NOT NULL DEFAULT 0,
  created_at        TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE INDEX IF NOT EXISTS idx_ml_model_type_status
  ON ml_model_registry (model_type, status);

-- Label Events: Capture user corrections for retraining
CREATE TABLE IF NOT EXISTS ml_label_event (
  label_event_id    INTEGER PRIMARY KEY,
  block_id          INTEGER NOT NULL,
  old_profile_id    INTEGER,
  new_profile_id    INTEGER,
  actor             TEXT NOT NULL DEFAULT 'USER',
  confidence_before TEXT,
  confidence_after  TEXT,
  ts                TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  FOREIGN KEY (block_id) REFERENCES block(block_id) ON DELETE CASCADE,
  FOREIGN KEY (old_profile_id) REFERENCES profile(profile_id) ON DELETE SET NULL,
  FOREIGN KEY (new_profile_id) REFERENCES profile(profile_id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_ml_label_block
  ON ml_label_event (block_id);

CREATE INDEX IF NOT EXISTS idx_ml_label_ts
  ON ml_label_event (ts);

-- ML Suggestions: Predictions from the model  
CREATE TABLE IF NOT EXISTS ml_suggestion (
  suggestion_id     INTEGER PRIMARY KEY,
  entity_type       TEXT NOT NULL CHECK (entity_type IN ('BLOCK', 'SESSION', 'RULE')),
  entity_id         INTEGER NOT NULL,
  suggestion_type   TEXT NOT NULL CHECK (suggestion_type IN ('PROFILE_ASSIGN', 'MERGE_BLOCKS', 'CREATE_RULE')),
  payload_json      TEXT NOT NULL,
  confidence        REAL NOT NULL CHECK (confidence >= 0.0 AND confidence <= 1.0),
  model_id          INTEGER,
  status            TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'ACCEPTED', 'REJECTED', 'EXPIRED')),
  created_at        TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  resolved_at       TEXT,
  FOREIGN KEY (model_id) REFERENCES ml_model_registry(model_id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_ml_suggestion_entity
  ON ml_suggestion (entity_type, entity_id);

CREATE INDEX IF NOT EXISTS idx_ml_suggestion_status
  ON ml_suggestion (status, confidence DESC);

-- ML Run Log: Audit log for training/inference runs
CREATE TABLE IF NOT EXISTS ml_run_log (
  run_id            INTEGER PRIMARY KEY,
  run_type          TEXT NOT NULL CHECK (run_type IN ('TRAIN', 'PREDICT', 'CLUSTER', 'RETRAIN')),
  model_id          INTEGER,
  success           INTEGER NOT NULL DEFAULT 0 CHECK (success IN (0,1)),
  error_summary     TEXT,
  duration_ms       INTEGER,
  input_samples     INTEGER,
  output_count      INTEGER,
  triggered_by      TEXT NOT NULL DEFAULT 'SYSTEM',
  ts                TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  FOREIGN KEY (model_id) REFERENCES ml_model_registry(model_id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_ml_run_type_ts
  ON ml_run_log (run_type, ts);

-- ----------------------------
-- App Blacklist
-- ----------------------------

CREATE TABLE IF NOT EXISTS app_blacklist (
  blacklist_id    INTEGER PRIMARY KEY,
  app_id          INTEGER NOT NULL UNIQUE,
  reason          TEXT,                            -- optional reason for blacklisting
  created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  FOREIGN KEY (app_id) REFERENCES dict_app(app_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_app_blacklist_app
  ON app_blacklist (app_id);
