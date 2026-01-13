package engine

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"chroniclecore/internal/store"
)

// Aggregator handles rollup of raw events into blocks
type Aggregator struct {
	store          *store.Store
	retentionDays  int
	ruleEngine     *RuleEngine
	templateEngine *TemplateEngine
	ctx            context.Context
	cancel         context.CancelFunc
}

// Config holds aggregator configuration
type AggregatorConfig struct {
	Store         *store.Store
	RetentionDays int           // How many days to keep raw events (default: 14)
	RollupInterval time.Duration // How often to run rollup (default: 5 minutes)
}

// NewAggregator creates a new aggregator
func NewAggregator(config AggregatorConfig) *Aggregator {
	if config.RetentionDays == 0 {
		config.RetentionDays = 14
	}

	if config.RollupInterval == 0 {
		config.RollupInterval = 5 * time.Minute
	}

	ctx, cancel := context.WithCancel(context.Background())

	agg := &Aggregator{
		store:          config.Store,
		retentionDays:  config.RetentionDays,
		ruleEngine:     NewRuleEngine(config.Store),
		templateEngine: NewTemplateEngine(config.Store),
		ctx:            ctx,
		cancel:         cancel,
	}

	// Load rules on startup
	if err := agg.ruleEngine.LoadRules(); err != nil {
		log.Printf("Warning: Failed to load rules: %v", err)
	}
	if err := agg.ruleEngine.LoadDictionaries(); err != nil {
		log.Printf("Warning: Failed to load dictionaries: %v", err)
	}

	// Start rollup scheduler
	go agg.scheduleRollup(config.RollupInterval)

	return agg
}

// Stop stops the aggregator
func (a *Aggregator) Stop() {
	if a.cancel != nil {
		a.cancel()
	}
}

// scheduleRollup runs rollup on a fixed interval
func (a *Aggregator) scheduleRollup(interval time.Duration) {
	// Run once immediately on startup
	if err := a.Rollup(); err != nil {
		log.Printf("Initial rollup failed: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			if err := a.Rollup(); err != nil {
				log.Printf("Rollup failed: %v", err)
			}
		}
	}
}

// Rollup aggregates raw events into blocks and purges old raw events
func (a *Aggregator) Rollup() error {
	log.Println("Starting rollup...")

	// Get last rollup timestamp from settings
	lastRollup, err := a.getLastRollupTime()
	if err != nil {
		return fmt.Errorf("failed to get last rollup time: %w", err)
	}

	// Get raw events since last rollup
	events, err := a.store.GetRawEventsSince(lastRollup)
	if err != nil {
		return fmt.Errorf("failed to get raw events: %w", err)
	}

	if len(events) == 0 {
		log.Println("No new events to process")
		return a.setLastRollupTime(time.Now())
	}

	log.Printf("Processing %d raw events", len(events))

	// Aggregate events into blocks
	blocks := a.aggregateEvents(events)
	log.Printf("Created %d blocks", len(blocks))

	// Insert blocks
	for _, block := range blocks {
		if err := a.store.InsertBlock(block); err != nil {
			log.Printf("Failed to insert block: %v", err)
		}
	}

	// Update last rollup time
	if err := a.setLastRollupTime(time.Now()); err != nil {
		return fmt.Errorf("failed to update last rollup time: %w", err)
	}

	// Apply rules to assign profiles to blocks
	if len(blocks) > 0 {
		if err := a.ruleEngine.AssignBlocksInRange(); err != nil {
			log.Printf("Failed to assign profiles: %v", err)
		}

		// Generate descriptions for new blocks
		if err := a.templateEngine.GenerateDescriptionsForBlocks(); err != nil {
			log.Printf("Failed to generate descriptions: %v", err)
		}
	}

	// Purge old raw events (beyond retention period)
	cutoff := time.Now().Add(-time.Duration(a.retentionDays) * 24 * time.Hour)
	deleted, err := a.store.DeleteRawEventsBefore(cutoff)
	if err != nil {
		log.Printf("Failed to purge old events: %v", err)
	} else if deleted > 0 {
		log.Printf("Purged %d old raw events (retention: %d days)", deleted, a.retentionDays)
	}

	log.Println("Rollup complete")
	return nil
}

// aggregateEvents groups sequential events into blocks
func (a *Aggregator) aggregateEvents(events []*store.RawEvent) []*store.Block {
	if len(events) == 0 {
		return nil
	}

	var blocks []*store.Block
	var currentBlock *blockBuilder

	for _, event := range events {
		// Skip events without end time (still open)
		if event.TsEnd == nil {
			continue
		}

		// Check if event should be grouped with current block
		if currentBlock != nil && currentBlock.canMerge(event) {
			currentBlock.merge(event)
		} else {
			// Finalize previous block
			if currentBlock != nil {
				if block := currentBlock.build(); block != nil {
					blocks = append(blocks, block)
				}
			}

			// Start new block
			currentBlock = newBlockBuilder(event)
		}
	}

	// Finalize last block
	if currentBlock != nil {
		if block := currentBlock.build(); block != nil {
			blocks = append(blocks, block)
		}
	}

	return blocks
}

// blockBuilder helps construct blocks from events
type blockBuilder struct {
	tsStart        time.Time
	tsEnd          time.Time
	primaryAppID   int64
	titleIDs       map[int64]bool
	domainIDs      map[int64]bool
	hasActiveTime  bool
	totalIdleTime  time.Duration
	activityScores []float64 // Store scores to calculate average
}

func newBlockBuilder(event *store.RawEvent) *blockBuilder {
	bb := &blockBuilder{
		tsStart:      event.TsStart,
		tsEnd:        *event.TsEnd,
		primaryAppID: event.AppID,
		titleIDs:     make(map[int64]bool),
		domainIDs:    make(map[int64]bool),
		activityScores: []float64{},
	}

	if event.TitleID != nil {
		bb.titleIDs[*event.TitleID] = true
	}

	if event.DomainID != nil {
		bb.domainIDs[*event.DomainID] = true
	}

	if event.State == "ACTIVE" {
		bb.hasActiveTime = true
		// Extract activity score
		if event.Metadata != nil {
			var meta map[string]interface{}
			if err := json.Unmarshal([]byte(*event.Metadata), &meta); err == nil {
				if score, ok := meta["activity_score"].(float64); ok {
					bb.activityScores = append(bb.activityScores, score)
				}
			}
		}
	} else if event.State == "IDLE" {
		duration := event.TsEnd.Sub(event.TsStart)
		bb.totalIdleTime += duration
	}

	return bb
}

// canMerge checks if an event can be merged into this block
func (bb *blockBuilder) canMerge(event *store.RawEvent) bool {
	// Same app and contiguous time (within 1 minute gap)
	gap := event.TsStart.Sub(bb.tsEnd)
	return event.AppID == bb.primaryAppID && gap < 1*time.Minute
}

// merge adds an event to the current block
func (bb *blockBuilder) merge(event *store.RawEvent) {
	if event.TsEnd == nil {
		return
	}

	bb.tsEnd = *event.TsEnd

	if event.TitleID != nil {
		bb.titleIDs[*event.TitleID] = true
	}

	if event.DomainID != nil {
		bb.domainIDs[*event.DomainID] = true
	}

	if event.State == "ACTIVE" {
		bb.hasActiveTime = true
		// Extract activity score
		if event.Metadata != nil {
			var meta map[string]interface{}
			if err := json.Unmarshal([]byte(*event.Metadata), &meta); err == nil {
				if score, ok := meta["activity_score"].(float64); ok {
					bb.activityScores = append(bb.activityScores, score)
				}
			}
		}
	} else if event.State == "IDLE" {
		duration := event.TsEnd.Sub(event.TsStart)
		bb.totalIdleTime += duration
	}
}

// build creates a block from accumulated events
func (bb *blockBuilder) build() *store.Block {
	// Skip blocks with no active time (pure idle)
	if !bb.hasActiveTime {
		return nil
	}

	// Pick primary title (first one for MVP)
	var titleID *int64
	for tid := range bb.titleIDs {
		t := tid
		titleID = &t
		break
	}

	// Pick primary domain (first one for MVP)
	var domainID *int64
	for did := range bb.domainIDs {
		d := did
		domainID = &d
		break
	}

	// Calculate average activity score
	var metadata *string
	if len(bb.activityScores) > 0 {
		var sum float64
		for _, s := range bb.activityScores {
			sum += s
		}
		avg := sum / float64(len(bb.activityScores))
		jsonStr := fmt.Sprintf(`{"avg_activity_score": %.2f}`, avg)
		metadata = &jsonStr
	}

	// Create block
	block := &store.Block{
		TsStart:         bb.tsStart,
		TsEnd:           bb.tsEnd,
		PrimaryAppID:    bb.primaryAppID,
		TitleSummaryID:  titleID,
		PrimaryDomainID: domainID,
		Confidence:      "LOW", // Will be assigned by rules engine
		Billable:        true,  // Default to billable (idle time excluded)
		Locked:          false,
		Metadata:        metadata,
	}

	return block
}

// getLastRollupTime retrieves the last rollup timestamp from settings
func (a *Aggregator) getLastRollupTime() (time.Time, error) {
	db := a.store.GetDB()

	var value string
	err := db.QueryRow(`
		SELECT value FROM settings
		WHERE key = 'last_rollup_ts'
	`).Scan(&value)

	if err == sql.ErrNoRows {
		// First run - use epoch or 7 days ago
		return time.Now().Add(-7 * 24 * time.Hour), nil
	}

	if err != nil {
		return time.Time{}, err
	}

	return time.Parse(time.RFC3339, value)
}

// setLastRollupTime updates the last rollup timestamp
func (a *Aggregator) setLastRollupTime(t time.Time) error {
	db := a.store.GetDB()

	_, err := db.Exec(`
		INSERT INTO settings (key, value, is_encrypted)
		VALUES ('last_rollup_ts', ?, 0)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, t.UTC().Format(time.RFC3339))

	return err
}
