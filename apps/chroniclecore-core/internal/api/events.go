package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"chroniclecore/internal/store"
)

// EventHandler manages extension event endpoints
type EventHandler struct {
	store *store.Store
}

func NewEventHandler(store *store.Store) *EventHandler {
	return &EventHandler{store: store}
}

// ExtensionEvent represents an event from the browser extension
type ExtensionEvent struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Domain      string `json:"domain"`
	Description string `json:"description"`
	TabID       int    `json:"tab_id"`
	Timestamp   string `json:"timestamp"`
	EventType   string `json:"event_type"`   // TAB_ACTIVATED, PAGE_LOADED
	DurationMs  *int   `json:"duration_ms"`  // Time spent on previous page
}

// IngestExtensionEvent handles POST /api/v1/events/ingest
// Receives events from the browser extension and stores them as raw_events
func (h *EventHandler) IngestExtensionEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event ExtensionEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		respondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if event.URL == "" || event.EventType == "" {
		respondError(w, "url and event_type are required", http.StatusBadRequest)
		return
	}

	// Parse timestamp or use current time
	var tsStart time.Time
	if event.Timestamp != "" {
		parsed, err := time.Parse(time.RFC3339, event.Timestamp)
		if err != nil {
			tsStart = time.Now()
		} else {
			tsStart = parsed
		}
	} else {
		tsStart = time.Now()
	}

	// Get or create dictionary entries
	// Use "Browser" as the app name since this comes from extension
	appName := "Browser"
	if event.Domain != "" {
		appName = "Browser (" + event.Domain + ")"
	}

	appID, err := h.store.GetOrCreateDictApp(appName)
	if err != nil {
		log.Printf("Failed to get/create app dict: %v", err)
		respondError(w, "Failed to process event", http.StatusInternalServerError)
		return
	}

	// Store title (includes description context)
	var titleID *int64
	titleText := event.Title
	if event.Description != "" {
		titleText = event.Description // Use the human-readable description as title
	}
	if titleText != "" {
		id, err := h.store.GetOrCreateDictTitle(titleText)
		if err != nil {
			log.Printf("Failed to get/create title dict: %v", err)
		} else {
			titleID = &id
		}
	}

	// Store domain
	var domainID *int64
	if event.Domain != "" {
		id, err := h.store.GetOrCreateDictDomain(event.Domain)
		if err != nil {
			log.Printf("Failed to get/create domain dict: %v", err)
		} else {
			domainID = &id
		}
	}

	// Build metadata JSON
	metadata := buildExtensionMetadata(event)

	// Create raw event
	rawEvent := &store.RawEvent{
		TsStart:  tsStart,
		TsEnd:    nil, // Extension events are point-in-time; rollup will handle duration
		AppID:    appID,
		TitleID:  titleID,
		DomainID: domainID,
		State:    "ACTIVE",
		Source:   "EXTENSION",
		Metadata: &metadata,
	}

	// If we have duration from previous page, set end time
	if event.DurationMs != nil && *event.DurationMs > 0 {
		endTime := tsStart
		startTime := tsStart.Add(-time.Duration(*event.DurationMs) * time.Millisecond)
		rawEvent.TsStart = startTime
		rawEvent.TsEnd = &endTime
	}

	if err := h.store.InsertRawEvent(rawEvent); err != nil {
		log.Printf("Failed to insert extension event: %v", err)
		respondError(w, "Failed to store event", http.StatusInternalServerError)
		return
	}

	log.Printf("Extension event: %s - %s", event.EventType, event.Description)

	respondJSON(w, map[string]interface{}{
		"success":  true,
		"event_id": rawEvent.EventID,
	}, http.StatusOK)
}

// buildExtensionMetadata creates JSON metadata for extension events
func buildExtensionMetadata(event ExtensionEvent) string {
	meta := map[string]interface{}{
		"source":      "extension",
		"event_type":  event.EventType,
		"tab_id":      event.TabID,
		"url":         event.URL,
		"description": event.Description,
	}

	if event.DurationMs != nil {
		meta["duration_ms"] = *event.DurationMs
	}

	jsonBytes, err := json.Marshal(meta)
	if err != nil {
		return `{"source": "extension"}`
	}
	return string(jsonBytes)
}
