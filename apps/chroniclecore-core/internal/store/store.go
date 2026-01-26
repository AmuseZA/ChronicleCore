package store

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Store manages database connections and operations
type Store struct {
	DB          *sql.DB // Exported for ML handler access
	dbPath      string
	mu          sync.RWMutex
	dictCache   *DictCache
	initialized bool
}

// DictCache stores in-memory cache of dictionary tables to avoid lookups
type DictCache struct {
	mu      sync.RWMutex
	apps    map[string]int64 // app_name -> app_id
	titles  map[string]int64 // title_text -> title_id
	domains map[string]int64 // domain_text -> domain_id
}

// RawEvent represents an activity event
type RawEvent struct {
	EventID       int64
	TsStart       time.Time
	TsEnd         *time.Time
	AppID         int64
	TitleID       *int64
	DomainID      *int64
	State         string // ACTIVE, IDLE, PAUSED
	Source        string // OS, EXTENSION
	Metadata      *string
	HashSignature *string
}

// Block represents an aggregated time block
type Block struct {
	BlockID          int64
	TsStart          time.Time
	TsEnd            time.Time
	PrimaryAppID     int64
	PrimaryDomainID  *int64
	TitleSummaryID   *int64
	ProfileID        *int64
	Confidence       string // HIGH, MEDIUM, LOW
	Billable         bool
	Locked           bool
	Notes            *string
	Description      *string
	Metadata         *string
	ActivityScore    float64 // 0.0-1.0 representing active work percentage for billing
}

// NewStore creates a new store instance (not yet initialized)
func NewStore(dbPath string) *Store {
	return &Store{
		dbPath: dbPath,
		dictCache: &DictCache{
			apps:    make(map[string]int64),
			titles:  make(map[string]int64),
			domains: make(map[string]int64),
		},
	}
}

// Init initializes the database connection and schema
func (s *Store) Init() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.initialized {
		return fmt.Errorf("store already initialized")
	}

	// Open database
	db, err := sql.Open("sqlite3", s.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	s.DB = db

	// Configure SQLite for optimal performance
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA cache_size = -64000", // 64MB cache
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return fmt.Errorf("failed to set pragma %s: %w", pragma, err)
		}
	}

	// Verify schema exists (check for install table)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='install'").Scan(&count)
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to verify schema: %w", err)
	}

	if count == 0 {
		db.Close()
		return fmt.Errorf("database schema not initialized - run schema.sql first")
	}

	// Ensure schema is up to date (auto-migrate v1.4 -> v1.5)
	if err := s.ensureSchema(db); err != nil {
		log.Printf("Warning: Schema migration failed: %v", err)
		// Proceed anyway, critical tables likely exist
	}

	s.initialized = true
	log.Printf("Store initialized: %s (WAL mode enabled)", s.dbPath)

	return nil
}

// Close closes the database connection
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized || s.DB == nil {
		return nil
	}

	err := s.DB.Close()
	s.initialized = false
	return err
}

// GetOrCreateDictApp gets or creates an app dictionary entry
func (s *Store) GetOrCreateDictApp(appName string) (int64, error) {
	// Check cache first
	s.dictCache.mu.RLock()
	if id, exists := s.dictCache.apps[appName]; exists {
		s.dictCache.mu.RUnlock()
		return id, nil
	}
	s.dictCache.mu.RUnlock()

	// Try to get from DB
	var appID int64
	err := s.DB.QueryRow(
		"SELECT app_id FROM dict_app WHERE app_name = ?",
		appName,
	).Scan(&appID)

	if err == nil {
		// Found - update cache
		s.dictCache.mu.Lock()
		s.dictCache.apps[appName] = appID
		s.dictCache.mu.Unlock()
		return appID, nil
	}

	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query dict_app: %w", err)
	}

	// Not found - insert new
	result, err := s.DB.Exec(
		"INSERT INTO dict_app (app_name) VALUES (?)",
		appName,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert dict_app: %w", err)
	}

	appID, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Update cache
	s.dictCache.mu.Lock()
	s.dictCache.apps[appName] = appID
	s.dictCache.mu.Unlock()

	return appID, nil
}

// GetOrCreateDictTitle gets or creates a title dictionary entry
func (s *Store) GetOrCreateDictTitle(titleText string) (int64, error) {
	// Check cache first
	s.dictCache.mu.RLock()
	if id, exists := s.dictCache.titles[titleText]; exists {
		s.dictCache.mu.RUnlock()
		return id, nil
	}
	s.dictCache.mu.RUnlock()

	// Try to get from DB
	var titleID int64
	err := s.DB.QueryRow(
		"SELECT title_id FROM dict_title WHERE title_text = ?",
		titleText,
	).Scan(&titleID)

	if err == nil {
		// Found - update cache
		s.dictCache.mu.Lock()
		s.dictCache.titles[titleText] = titleID
		s.dictCache.mu.Unlock()
		return titleID, nil
	}

	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query dict_title: %w", err)
	}

	// Not found - insert new
	result, err := s.DB.Exec(
		"INSERT INTO dict_title (title_text) VALUES (?)",
		titleText,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert dict_title: %w", err)
	}

	titleID, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Update cache
	s.dictCache.mu.Lock()
	s.dictCache.titles[titleText] = titleID
	s.dictCache.mu.Unlock()

	return titleID, nil
}

// GetOrCreateDictDomain gets or creates a domain dictionary entry
func (s *Store) GetOrCreateDictDomain(domainText string) (int64, error) {
	// Check cache first
	s.dictCache.mu.RLock()
	if id, exists := s.dictCache.domains[domainText]; exists {
		s.dictCache.mu.RUnlock()
		return id, nil
	}
	s.dictCache.mu.RUnlock()

	// Try to get from DB
	var domainID int64
	err := s.DB.QueryRow(
		"SELECT domain_id FROM dict_domain WHERE domain_text = ?",
		domainText,
	).Scan(&domainID)

	if err == nil {
		// Found - update cache
		s.dictCache.mu.Lock()
		s.dictCache.domains[domainText] = domainID
		s.dictCache.mu.Unlock()
		return domainID, nil
	}

	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query dict_domain: %w", err)
	}

	// Not found - insert new
	result, err := s.DB.Exec(
		"INSERT INTO dict_domain (domain_text) VALUES (?)",
		domainText,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert dict_domain: %w", err)
	}

	domainID, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Update cache
	s.dictCache.mu.Lock()
	s.dictCache.domains[domainText] = domainID
	s.dictCache.mu.Unlock()

	return domainID, nil
}

// InsertRawEvent inserts a raw activity event
func (s *Store) InsertRawEvent(event *RawEvent) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return fmt.Errorf("store not initialized")
	}

	query := `
		INSERT INTO raw_event (
			ts_start, ts_end, app_id, title_id, domain_id,
			state, source, metadata, hash_signature
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var tsEnd *string
	if event.TsEnd != nil {
		ts := event.TsEnd.UTC().Format(time.RFC3339)
		tsEnd = &ts
	}

	result, err := s.DB.Exec(
		query,
		event.TsStart.UTC().Format(time.RFC3339),
		tsEnd,
		event.AppID,
		event.TitleID,
		event.DomainID,
		event.State,
		event.Source,
		event.Metadata,
		event.HashSignature,
	)

	if err != nil {
		return fmt.Errorf("failed to insert raw_event: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		event.EventID = id
	}

	return nil
}

// GetRawEventsSince retrieves raw events since a given timestamp
func (s *Store) GetRawEventsSince(since time.Time) ([]*RawEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, fmt.Errorf("store not initialized")
	}

	query := `
		SELECT
			event_id, ts_start, ts_end, app_id, title_id,
			domain_id, state, source, metadata, hash_signature
		FROM raw_event
		WHERE ts_start >= ?
		ORDER BY ts_start ASC
	`

	rows, err := s.DB.Query(query, since.UTC().Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("failed to query raw_event: %w", err)
	}
	defer rows.Close()

	var events []*RawEvent

	for rows.Next() {
		var event RawEvent
		var tsStartStr, tsEndStr sql.NullString
		var titleID, domainID sql.NullInt64
		var metadata, hashSig sql.NullString

		err := rows.Scan(
			&event.EventID,
			&tsStartStr,
			&tsEndStr,
			&event.AppID,
			&titleID,
			&domainID,
			&event.State,
			&event.Source,
			&metadata,
			&hashSig,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan raw_event: %w", err)
		}

		// Parse timestamps
		if tsStartStr.Valid {
			ts, _ := time.Parse(time.RFC3339, tsStartStr.String)
			event.TsStart = ts
		}

		if tsEndStr.Valid {
			ts, _ := time.Parse(time.RFC3339, tsEndStr.String)
			event.TsEnd = &ts
		}

		if titleID.Valid {
			id := titleID.Int64
			event.TitleID = &id
		}

		if domainID.Valid {
			id := domainID.Int64
			event.DomainID = &id
		}

		if metadata.Valid {
			m := metadata.String
			event.Metadata = &m
		}

		if hashSig.Valid {
			sig := hashSig.String
			event.HashSignature = &sig
		}

		events = append(events, &event)
	}

	return events, rows.Err()
}

// DeleteRawEventsBefore deletes raw events older than a given timestamp
func (s *Store) DeleteRawEventsBefore(before time.Time) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		return 0, fmt.Errorf("store not initialized")
	}

	result, err := s.DB.Exec(
		"DELETE FROM raw_event WHERE ts_start < ?",
		before.UTC().Format(time.RFC3339),
	)

	if err != nil {
		return 0, fmt.Errorf("failed to delete raw_events: %w", err)
	}

	count, _ := result.RowsAffected()
	return count, nil
}

// InsertBlock inserts an aggregated time block
func (s *Store) InsertBlock(block *Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		return fmt.Errorf("store not initialized")
	}

	query := `
		INSERT INTO block (
			ts_start, ts_end, primary_app_id, primary_domain_id,
			title_summary_id, profile_id, confidence, billable,
			locked, notes, description, metadata, activity_score
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.DB.Exec(
		query,
		block.TsStart.UTC().Format(time.RFC3339),
		block.TsEnd.UTC().Format(time.RFC3339),
		block.PrimaryAppID,
		block.PrimaryDomainID,
		block.TitleSummaryID,
		block.ProfileID,
		block.Confidence,
		block.Billable,
		block.Locked,
		block.Notes,
		block.Description,
		block.Metadata,
		block.ActivityScore,
	)

	if err != nil {
		return fmt.Errorf("failed to insert block: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		block.BlockID = id
	}

	return nil
}

// DeleteBlock deletes a block by ID
func (s *Store) DeleteBlock(blockID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		return fmt.Errorf("store not initialized")
	}

	// Manually delete related ML events first (safer than relying on CASCADE)
	if _, err := s.DB.Exec("DELETE FROM ml_label_event WHERE block_id = ?", blockID); err != nil {
		log.Printf("Warning: Failed to delete ml_label_event for block %d: %v", blockID, err)
	}

	res, err := s.DB.Exec("DELETE FROM block WHERE block_id = ?", blockID)
	if err != nil {
		log.Printf("Error deleting block %d: %v", blockID, err)
		return fmt.Errorf("failed to delete block: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("block not found")
	}

	return nil
}

// GetDB returns the underlying database connection (for advanced queries)
func (s *Store) GetDB() *sql.DB {
	return s.DB
}

// ============================================================
// Settings Management
// ============================================================

// GetSetting retrieves a setting value by key
func (s *Store) GetSetting(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return "", fmt.Errorf("store not initialized")
	}

	var value string
	err := s.DB.QueryRow(
		"SELECT value FROM settings WHERE key = ?",
		key,
	).Scan(&value)

	if err == sql.ErrNoRows {
		return "", nil // Return empty string for missing keys
	}

	if err != nil {
		return "", fmt.Errorf("failed to get setting %s: %w", key, err)
	}

	return value, nil
}

// SetSetting sets a setting value by key
func (s *Store) SetSetting(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		return fmt.Errorf("store not initialized")
	}

	_, err := s.DB.Exec(`
		INSERT INTO settings (key, value, is_encrypted)
		VALUES (?, ?, 0)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)

	if err != nil {
		return fmt.Errorf("failed to set setting %s: %w", key, err)
	}

	return nil
}

// GetSettingBool retrieves a boolean setting (returns false if not set)
func (s *Store) GetSettingBool(key string) (bool, error) {
	value, err := s.GetSetting(key)
	if err != nil {
		return false, err
	}

	return value == "true" || value == "1", nil
}

// SetSettingBool sets a boolean setting
func (s *Store) SetSettingBool(key string, value bool) error {
	strValue := "false"
	if value {
		strValue = "true"
	}
	return s.SetSetting(key, strValue)
}

// GetAllSettings retrieves all settings as a map
func (s *Store) GetAllSettings() (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, fmt.Errorf("store not initialized")
	}

	rows, err := s.DB.Query("SELECT key, value FROM settings WHERE is_encrypted = 0")
	if err != nil {
		return nil, fmt.Errorf("failed to query settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}
		settings[key] = value
	}

	return settings, rows.Err()
}
