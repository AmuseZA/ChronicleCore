package store

import (
	"database/sql"
	"log"
	"strings"
)

// ensureSchema handles database migrations and table creation for updates
func (s *Store) ensureSchema(db *sql.DB) error {
	queries := []string{
		// 1.5.0 Migration: Add 'metadata' to block if missing
		`ALTER TABLE block ADD COLUMN metadata TEXT`,

		// 1.5.0 Migration: Add 'metadata' to raw_event if missing
		`ALTER TABLE raw_event ADD COLUMN metadata TEXT`,

		// 1.5.0 New Tables (ML Pipeline)
		// Model Registry
		`CREATE TABLE IF NOT EXISTS ml_model_registry (
		  model_id          INTEGER PRIMARY KEY,
		  model_type        TEXT NOT NULL CHECK (model_type IN ('PROFILE_CLASSIFIER', 'SESSION_CLUSTERER')),
		  version           TEXT NOT NULL,
		  algorithm         TEXT NOT NULL,
		  metrics_json      TEXT,
		  status            TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'ARCHIVED', 'FAILED')),
		  trained_at        TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		  trained_samples   INTEGER NOT NULL DEFAULT 0,
		  created_at        TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
		);`,

		// Label Events
		`CREATE TABLE IF NOT EXISTS ml_label_event (
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
		);`,

		// ML Suggestions
		`CREATE TABLE IF NOT EXISTS ml_suggestion (
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
		);`,

		// ML Run Log
		`CREATE TABLE IF NOT EXISTS ml_run_log (
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
		);`,

		// 1.6.0 Migration: App Blacklist table
		`CREATE TABLE IF NOT EXISTS app_blacklist (
		  blacklist_id    INTEGER PRIMARY KEY,
		  app_id          INTEGER NOT NULL UNIQUE,
		  reason          TEXT,
		  created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		  FOREIGN KEY (app_id) REFERENCES dict_app(app_id) ON DELETE CASCADE
		);`,

		// 1.6.0 Index for app_blacklist
		`CREATE INDEX IF NOT EXISTS idx_app_blacklist_app ON app_blacklist (app_id);`,

		// 1.7.0 Migration: Manual entries support
		`ALTER TABLE block ADD COLUMN is_manual INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE block ADD COLUMN manual_title TEXT`,

		// 1.7.0 Migration: Enhanced activity tracking
		`ALTER TABLE block ADD COLUMN action_type TEXT`,
		`ALTER TABLE block ADD COLUMN entity_context TEXT`,

		// 1.7.0 Index for manual entries
		`CREATE INDEX IF NOT EXISTS idx_block_is_manual ON block (is_manual);`,

		// 1.8.0 Migration: Keyword Blacklist
		`CREATE TABLE IF NOT EXISTS keyword_blacklist (
		  keyword_id      INTEGER PRIMARY KEY,
		  keyword_text    TEXT NOT NULL UNIQUE,
		  reason          TEXT,
		  created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
		);`,

		// 1.8.1 Migration: ML deletion learning
		`ALTER TABLE ml_label_event ADD COLUMN action_type TEXT DEFAULT 'ASSIGN'`,

		// 1.8.1 Migration: Add DELETE_SUGGEST to suggestion types (handled via INSERT not ALTER)
		// Note: SQLite doesn't support ALTER CHECK constraint, so we create a new table to track deletion training
		`CREATE TABLE IF NOT EXISTS ml_deletion_event (
		  deletion_event_id INTEGER PRIMARY KEY,
		  app_name          TEXT NOT NULL,
		  title_text        TEXT,
		  domain_text       TEXT,
		  ts_start          TEXT,
		  ts_end            TEXT,
		  actor             TEXT NOT NULL DEFAULT 'USER',
		  created_at        TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
		);`,
	}

	for _, query := range queries {
		// Attempt to run the query
		// For ALTER TABLE, it will fail if column exists. We ignore "duplicate column" errors.
		_, err := db.Exec(query)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate column") {
				continue // Already exists, safe to ignore
			}
			// For CREATE TABLE IF NOT EXISTS, it shouldn't error unless schema invalid
			// Log but don't stop, to maximize chance of success
			log.Printf("Schema migration note (ignorable): %v", err)
		}
	}

	return nil
}
