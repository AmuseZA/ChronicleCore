package tracker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"chroniclecore/internal/store"
)

// State represents the tracker state
type State string

const (
	StateStopped State = "STOPPED"
	StateActive  State = "ACTIVE"
	StatePaused  State = "PAUSED"
)

// Config holds tracker configuration
type Config struct {
	PollInterval         time.Duration // How often to poll window info (default: 5s)
	IdleThresholdSeconds int           // Idle threshold (default: 300s = 5 mins)
	Store                *store.Store
}

// Tracker manages activity tracking
type Tracker struct {
	config            Config
	state             State
	mu                sync.RWMutex
	ctx               context.Context
	cancel            context.CancelFunc
	lastActiveAt      *time.Time
	currentWindow     *WindowInfo
	currentEventID    int64 // Track current open event
	activitySamples   int   // Total samples in current poll interval
	activeSampleCount int   // Number of samples where user was active
}

// Status represents current tracker status
type Status struct {
	State         State
	LastActiveAt  *time.Time
	IdleSeconds   int
	CurrentWindow *WindowInfo
}

// NewTracker creates a new tracker instance
func NewTracker(config Config) *Tracker {
	// Set defaults
	if config.PollInterval == 0 {
		config.PollInterval = 5 * time.Second
	}
	if config.IdleThresholdSeconds == 0 {
		config.IdleThresholdSeconds = 300 // 5 minutes
	}

	return &Tracker{
		config: config,
		state:  StateStopped,
	}
}

// Start begins activity tracking
func (t *Tracker) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state != StateStopped {
		return fmt.Errorf("tracker already running (state: %s)", t.state)
	}

	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.state = StateActive

	// Start tracking goroutine
	go t.trackLoop()

	log.Println("Tracker started")
	return nil
}

// Pause pauses tracking (keeps state but stops capturing)
func (t *Tracker) Pause() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state != StateActive {
		return fmt.Errorf("tracker not active (state: %s)", t.state)
	}

	t.state = StatePaused

	// Close current event if open
	if t.currentEventID != 0 {
		t.closeCurrentEvent()
	}

	log.Println("Tracker paused")
	return nil
}

// Resume resumes tracking from paused state
func (t *Tracker) Resume() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state != StatePaused {
		return fmt.Errorf("tracker not paused (state: %s)", t.state)
	}

	t.state = StateActive
	log.Println("Tracker resumed")
	return nil
}

// Stop stops tracking completely
func (t *Tracker) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state == StateStopped {
		return nil
	}

	// Close current event if open
	if t.currentEventID != 0 {
		t.closeCurrentEvent()
	}

	t.state = StateStopped
	if t.cancel != nil {
		t.cancel()
	}

	log.Println("Tracker stopped")
	return nil
}

// GetStatus returns current tracker status
func (t *Tracker) GetStatus() Status {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return Status{
		State:         t.state,
		LastActiveAt:  t.lastActiveAt,
		IdleSeconds:   0, // TODO: Get from last capture
		CurrentWindow: t.currentWindow,
	}
}

// trackLoop is the main tracking loop (runs in goroutine)
func (t *Tracker) trackLoop() {
	// Main poll ticker (e.g. 5s)
	mainTicker := time.NewTicker(t.config.PollInterval)
	defer mainTicker.Stop()

	// High-frequency sampler ticker (e.g. 200ms)
	sampleTicker := time.NewTicker(200 * time.Millisecond)
	defer sampleTicker.Stop()

	var lastIdleMs int64 = 0

	for {
		select {
		case <-t.ctx.Done():
			return

		case <-sampleTicker.C:
			// High-freq sampling for activity meter
			if t.state != StateActive {
				continue
			}

			currentIdleMs, err := GetIdleTime()
			if err != nil {
				continue
			}

			t.mu.Lock()
			t.activitySamples++

			// If idle time DECREASED, user did something (moved mouse, typed)
			// Or if idle time is very small (< 100ms), they are currently doing something
			if currentIdleMs < lastIdleMs || currentIdleMs < 100 {
				t.activeSampleCount++
			}

			lastIdleMs = currentIdleMs
			t.mu.Unlock()

		case <-mainTicker.C:
			t.capture()
		}
	}
}

// capture captures the current window state
func (t *Tracker) capture() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Don't capture if paused
	if t.state != StateActive {
		return
	}

	// Calculate Activity Score for this interval
	var activityScore float64 = 0.0
	if t.activitySamples > 0 {
		activityScore = float64(t.activeSampleCount) / float64(t.activitySamples)
	}

	// Reset counters for next interval
	t.activitySamples = 0
	t.activeSampleCount = 0

	// Get current window info
	winInfo, err := GetCurrentWindowInfo(t.config.IdleThresholdSeconds)
	if err != nil {
		log.Printf("Failed to capture window info: %v", err)
		return
	}

	t.currentWindow = winInfo
	now := time.Now()

	// Determine event state
	eventState := "ACTIVE"
	if winInfo.IsIdle {
		eventState = "IDLE"
		activityScore = 0.0 // Force 0 if officially idle
	}

	// Get or create dictionary IDs
	appID, err := t.config.Store.GetOrCreateDictApp(winInfo.ProcessName)
	if err != nil {
		log.Printf("Failed to get/create app dict: %v", err)
		return
	}

	var titleID *int64
	if winInfo.WindowTitle != "" {
		id, err := t.config.Store.GetOrCreateDictTitle(winInfo.WindowTitle)
		if err != nil {
			log.Printf("Failed to get/create title dict: %v", err)
		} else {
			titleID = &id
		}
	}

	// Update current window info with resolved IDs and State BEFORE checking for change
	// This ensures t.currentWindow has the "new" state to compare against the "previous" state
	// Wait, t.currentWindow holds the window from the *previous* capture iteration.
	// We need to compare "previous" (t.currentWindow) vs "current" (winInfo).
	// So we should NOT overwrite t.currentWindow yet.

	// Check if context has changed (different app/title or state change)
	contextChanged := t.shouldStartNewEvent(appID, titleID, eventState)

	// Update t.currentWindow for next iteration
	// We need to store specific fields for comparison next time
	winInfo.AppID = appID
	winInfo.TitleID = titleID
	winInfo.State = eventState
	t.currentWindow = winInfo // Now it becomes the "last" window for next loop

	if contextChanged {
		// Close previous event if open
		if t.currentEventID != 0 {
			t.closeCurrentEvent()
		}

		// Prepare Metadata JSON
		metadata := fmt.Sprintf(`{"activity_score": %.2f}`, activityScore)

		// Create new event
		event := &store.RawEvent{
			TsStart:  now,
			TsEnd:    nil, // Open-ended
			AppID:    appID,
			TitleID:  titleID,
			State:    eventState,
			Source:   "OS",
			Metadata: &metadata,
		}

		err = t.config.Store.InsertRawEvent(event)
		if err != nil {
			log.Printf("Failed to insert raw event: %v", err)
			return
		}

		t.currentEventID = event.EventID
		t.lastActiveAt = &now
	} else {
		// If NOT changed, we should ideally update the lastActiveAt of the *current event*?
		// No, lastActiveAt is global for the tracker status.
	}

	// Update global last active time if not idle
	if !winInfo.IsIdle {
		t.lastActiveAt = &now
	}
}

// shouldStartNewEvent determines if a new event should be created
func (t *Tracker) shouldStartNewEvent(appID int64, titleID *int64, state string) bool {
	// If no current event, always start new one
	if t.currentEventID == 0 {
		return true
	}

	// If we have no previous window info, start new
	if t.currentWindow == nil {
		return true
	}

	// Helper to check if titleIDs differ (handling nil pointers)
	titlesDiffer := false
	if (t.currentWindow.TitleID == nil && titleID != nil) ||
		(t.currentWindow.TitleID != nil && titleID == nil) {
		titlesDiffer = true
	} else if t.currentWindow.TitleID != nil && titleID != nil {
		titlesDiffer = *t.currentWindow.TitleID != *titleID
	}

	// Check if context has changed
	if t.currentWindow.AppID != appID ||
		titlesDiffer ||
		t.currentWindow.State != state {
		return true
	}

	return false
}

// closeCurrentEvent closes the currently open event
func (t *Tracker) closeCurrentEvent() {
	if t.currentEventID == 0 {
		return
	}

	now := time.Now()

	// Update the event's ts_end
	db := t.config.Store.GetDB()
	_, err := db.Exec(
		"UPDATE raw_event SET ts_end = ? WHERE event_id = ?",
		now.UTC().Format(time.RFC3339),
		t.currentEventID,
	)

	if err != nil {
		log.Printf("Failed to close event %d: %v", t.currentEventID, err)
	}

	t.currentEventID = 0
}
