package store

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestDB creates a test database with schema
func setupTestDB(t *testing.T) (*Store, string) {
	// Create temp directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Read schema from project root
	schemaPath := filepath.Join("..", "..", "..", "..", "spec", "schema.sql")
	schemaSQL, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("Failed to read schema: %v", err)
	}

	// Create database and apply schema
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	_, err = db.Exec(string(schemaSQL))
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}
	db.Close()

	// Create store
	store := NewStore(dbPath)
	if err := store.Init(); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}

	return store, dbPath
}

func TestStoreInit(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	if !store.initialized {
		t.Error("Store should be initialized")
	}

	if store.db == nil {
		t.Error("Database connection should not be nil")
	}
}

func TestGetOrCreateDictApp(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	appName := "EXCEL.EXE"

	// First call - should create
	id1, err := store.GetOrCreateDictApp(appName)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	if id1 == 0 {
		t.Error("App ID should not be zero")
	}

	// Second call - should return same ID (cached)
	id2, err := store.GetOrCreateDictApp(appName)
	if err != nil {
		t.Fatalf("Failed to get app: %v", err)
	}

	if id1 != id2 {
		t.Errorf("Expected same ID, got %d and %d", id1, id2)
	}

	// Verify cache is working
	store.dictCache.mu.RLock()
	cachedID, exists := store.dictCache.apps[appName]
	store.dictCache.mu.RUnlock()

	if !exists {
		t.Error("App should be in cache")
	}

	if cachedID != id1 {
		t.Errorf("Cached ID mismatch: expected %d, got %d", id1, cachedID)
	}
}

func TestGetOrCreateDictTitle(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	titleText := "Budget 2026 - Excel"

	// First call - should create
	id1, err := store.GetOrCreateDictTitle(titleText)
	if err != nil {
		t.Fatalf("Failed to create title: %v", err)
	}

	if id1 == 0 {
		t.Error("Title ID should not be zero")
	}

	// Second call - should return same ID
	id2, err := store.GetOrCreateDictTitle(titleText)
	if err != nil {
		t.Fatalf("Failed to get title: %v", err)
	}

	if id1 != id2 {
		t.Errorf("Expected same ID, got %d and %d", id1, id2)
	}
}

func TestInsertRawEvent(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	// Create dependencies
	appID, _ := store.GetOrCreateDictApp("WINWORD.EXE")
	titleID, _ := store.GetOrCreateDictTitle("Document1 - Word")

	// Create event
	now := time.Now().UTC()
	end := now.Add(5 * time.Minute)

	event := &RawEvent{
		TsStart:  now,
		TsEnd:    &end,
		AppID:    appID,
		TitleID:  &titleID,
		State:    "ACTIVE",
		Source:   "OS",
	}

	err := store.InsertRawEvent(event)
	if err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	if event.EventID == 0 {
		t.Error("Event ID should be set after insert")
	}
}

func TestGetRawEventsSince(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	appID, _ := store.GetOrCreateDictApp("CODE.EXE")
	titleID, _ := store.GetOrCreateDictTitle("main.go - VSCode")

	// Insert test events
	baseTime := time.Now().UTC().Add(-1 * time.Hour)

	for i := 0; i < 5; i++ {
		start := baseTime.Add(time.Duration(i*10) * time.Minute)
		end := start.Add(5 * time.Minute)

		event := &RawEvent{
			TsStart:  start,
			TsEnd:    &end,
			AppID:    appID,
			TitleID:  &titleID,
			State:    "ACTIVE",
			Source:   "OS",
		}

		store.InsertRawEvent(event)
	}

	// Query events since 30 minutes ago
	since := baseTime.Add(30 * time.Minute)
	events, err := store.GetRawEventsSince(since)

	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	// Should get last 3 events (30m, 40m, 50m offsets)
	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}

	// Verify order (should be ascending by ts_start)
	for i := 1; i < len(events); i++ {
		if events[i].TsStart.Before(events[i-1].TsStart) {
			t.Error("Events should be ordered by ts_start ascending")
		}
	}
}

func TestDeleteRawEventsBefore(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	appID, _ := store.GetOrCreateDictApp("EXCEL.EXE")
	titleID, _ := store.GetOrCreateDictTitle("Spreadsheet.xlsx")

	// Insert events at different times
	now := time.Now().UTC()

	oldEvent := &RawEvent{
		TsStart: now.Add(-20 * 24 * time.Hour), // 20 days ago
		AppID:   appID,
		TitleID: &titleID,
		State:   "ACTIVE",
		Source:  "OS",
	}
	store.InsertRawEvent(oldEvent)

	recentEvent := &RawEvent{
		TsStart: now.Add(-5 * 24 * time.Hour), // 5 days ago
		AppID:   appID,
		TitleID: &titleID,
		State:   "ACTIVE",
		Source:  "OS",
	}
	store.InsertRawEvent(recentEvent)

	// Delete events older than 14 days
	cutoff := now.Add(-14 * 24 * time.Hour)
	count, err := store.DeleteRawEventsBefore(cutoff)

	if err != nil {
		t.Fatalf("Failed to delete events: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 deleted event, got %d", count)
	}

	// Verify recent event still exists
	events, _ := store.GetRawEventsSince(now.Add(-7 * 24 * time.Hour))
	if len(events) != 1 {
		t.Errorf("Expected 1 remaining event, got %d", len(events))
	}
}

func TestInsertBlock(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	appID, _ := store.GetOrCreateDictApp("msedge.exe")
	titleID, _ := store.GetOrCreateDictTitle("Gmail - Inbox")

	now := time.Now().UTC()
	end := now.Add(30 * time.Minute)

	block := &Block{
		TsStart:        now,
		TsEnd:          end,
		PrimaryAppID:   appID,
		TitleSummaryID: &titleID,
		Confidence:     "HIGH",
		Billable:       true,
		Locked:         false,
	}

	err := store.InsertBlock(block)
	if err != nil {
		t.Fatalf("Failed to insert block: %v", err)
	}

	if block.BlockID == 0 {
		t.Error("Block ID should be set after insert")
	}

	// Verify block was inserted
	var count int
	err = store.db.QueryRow("SELECT COUNT(*) FROM block").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query blocks: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 block, got %d", count)
	}
}

func TestConcurrentDictAccess(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	// Test concurrent access to dictionary caching
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(n int) {
			appName := "TEST.EXE"
			id, err := store.GetOrCreateDictApp(appName)
			if err != nil {
				t.Errorf("Goroutine %d failed: %v", n, err)
			}
			if id == 0 {
				t.Errorf("Goroutine %d got zero ID", n)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// All goroutines should have gotten the same ID
	store.dictCache.mu.RLock()
	id, exists := store.dictCache.apps["TEST.EXE"]
	store.dictCache.mu.RUnlock()

	if !exists {
		t.Error("App should be in cache after concurrent access")
	}

	// Verify only one entry in database
	var count int
	store.db.QueryRow("SELECT COUNT(*) FROM dict_app WHERE app_name = 'TEST.EXE'").Scan(&count)
	if count != 1 {
		t.Errorf("Expected 1 app entry, got %d (cache: %d)", count, id)
	}
}
