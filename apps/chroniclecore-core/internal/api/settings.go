package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"chroniclecore/internal/store"
)

// SettingsHandler manages settings endpoints
type SettingsHandler struct {
	store *store.Store
}

func NewSettingsHandler(store *store.Store) *SettingsHandler {
	return &SettingsHandler{store: store}
}

// Setting keys constants
const (
	SettingFullTrackingMode     = "full_tracking_mode"
	SettingDeepTrackingEnabled  = "deep_tracking_enabled"
	SettingTrackBrowserContent  = "track_browser_content"
	SettingTrackEmailContent    = "track_email_content"
	SettingTrackDocumentContent = "track_document_content"
	SettingTrackChatContent     = "track_chat_content"
	SettingExcludedApps         = "excluded_apps"
	SettingIdleThreshold        = "idle_threshold_seconds"
	SettingPrivacyMode          = "privacy_mode" // Redacts sensitive content
)

// SettingsResponse represents the full settings object
type SettingsResponse struct {
	FullTrackingMode     bool     `json:"full_tracking_mode"`
	DeepTrackingEnabled  bool     `json:"deep_tracking_enabled"`
	TrackBrowserContent  bool     `json:"track_browser_content"`
	TrackEmailContent    bool     `json:"track_email_content"`
	TrackDocumentContent bool     `json:"track_document_content"`
	TrackChatContent     bool     `json:"track_chat_content"`
	ExcludedApps         []string `json:"excluded_apps"`
	IdleThresholdSeconds int      `json:"idle_threshold_seconds"`
	PrivacyMode          bool     `json:"privacy_mode"`
}

// GetSettings handles GET /api/v1/settings
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	settings, err := h.loadSettings()
	if err != nil {
		log.Printf("Failed to load settings: %v", err)
		respondError(w, "Failed to load settings", http.StatusInternalServerError)
		return
	}

	respondJSON(w, settings, http.StatusOK)
}

// UpdateSettings handles PUT /api/v1/settings
func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SettingsResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Save each setting
	if err := h.store.SetSettingBool(SettingFullTrackingMode, req.FullTrackingMode); err != nil {
		log.Printf("Failed to save %s: %v", SettingFullTrackingMode, err)
	}

	if err := h.store.SetSettingBool(SettingDeepTrackingEnabled, req.DeepTrackingEnabled); err != nil {
		log.Printf("Failed to save %s: %v", SettingDeepTrackingEnabled, err)
	}

	if err := h.store.SetSettingBool(SettingTrackBrowserContent, req.TrackBrowserContent); err != nil {
		log.Printf("Failed to save %s: %v", SettingTrackBrowserContent, err)
	}

	if err := h.store.SetSettingBool(SettingTrackEmailContent, req.TrackEmailContent); err != nil {
		log.Printf("Failed to save %s: %v", SettingTrackEmailContent, err)
	}

	if err := h.store.SetSettingBool(SettingTrackDocumentContent, req.TrackDocumentContent); err != nil {
		log.Printf("Failed to save %s: %v", SettingTrackDocumentContent, err)
	}

	if err := h.store.SetSettingBool(SettingTrackChatContent, req.TrackChatContent); err != nil {
		log.Printf("Failed to save %s: %v", SettingTrackChatContent, err)
	}

	if err := h.store.SetSettingBool(SettingPrivacyMode, req.PrivacyMode); err != nil {
		log.Printf("Failed to save %s: %v", SettingPrivacyMode, err)
	}

	// Save excluded apps as JSON array
	if excludedJSON, err := json.Marshal(req.ExcludedApps); err == nil {
		h.store.SetSetting(SettingExcludedApps, string(excludedJSON))
	}

	// Save idle threshold
	if req.IdleThresholdSeconds > 0 {
		h.store.SetSetting(SettingIdleThreshold, intToString(req.IdleThresholdSeconds))
	}

	log.Printf("Settings updated: full_tracking=%v, deep_tracking=%v",
		req.FullTrackingMode, req.DeepTrackingEnabled)

	// Return updated settings
	settings, _ := h.loadSettings()
	respondJSON(w, settings, http.StatusOK)
}

// GetSingleSetting handles GET /api/v1/settings/{key}
func (h *SettingsHandler) GetSingleSetting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract key from path /api/v1/settings/{key}
	path := r.URL.Path
	key := path[len("/api/v1/settings/"):]

	if key == "" {
		respondError(w, "Setting key required", http.StatusBadRequest)
		return
	}

	value, err := h.store.GetSetting(key)
	if err != nil {
		respondError(w, "Failed to get setting", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{
		"key":   key,
		"value": value,
	}, http.StatusOK)
}

// loadSettings loads all settings with defaults
func (h *SettingsHandler) loadSettings() (*SettingsResponse, error) {
	settings := &SettingsResponse{
		FullTrackingMode:     false,
		DeepTrackingEnabled:  false,
		TrackBrowserContent:  true, // Default: track browser content
		TrackEmailContent:    true,
		TrackDocumentContent: true,
		TrackChatContent:     true,
		ExcludedApps:         []string{},
		IdleThresholdSeconds: 300, // Default: 5 minutes
		PrivacyMode:          false,
	}

	// Load from database
	if val, err := h.store.GetSettingBool(SettingFullTrackingMode); err == nil {
		settings.FullTrackingMode = val
	}

	if val, err := h.store.GetSettingBool(SettingDeepTrackingEnabled); err == nil {
		settings.DeepTrackingEnabled = val
	}

	if val, err := h.store.GetSettingBool(SettingTrackBrowserContent); err == nil {
		settings.TrackBrowserContent = val
	}

	if val, err := h.store.GetSettingBool(SettingTrackEmailContent); err == nil {
		settings.TrackEmailContent = val
	}

	if val, err := h.store.GetSettingBool(SettingTrackDocumentContent); err == nil {
		settings.TrackDocumentContent = val
	}

	if val, err := h.store.GetSettingBool(SettingTrackChatContent); err == nil {
		settings.TrackChatContent = val
	}

	if val, err := h.store.GetSettingBool(SettingPrivacyMode); err == nil {
		settings.PrivacyMode = val
	}

	// Load excluded apps
	if excludedJSON, err := h.store.GetSetting(SettingExcludedApps); err == nil && excludedJSON != "" {
		var excluded []string
		if err := json.Unmarshal([]byte(excludedJSON), &excluded); err == nil {
			settings.ExcludedApps = excluded
		}
	}

	// Load idle threshold
	if thresholdStr, err := h.store.GetSetting(SettingIdleThreshold); err == nil && thresholdStr != "" {
		if threshold := stringToInt(thresholdStr); threshold > 0 {
			settings.IdleThresholdSeconds = threshold
		}
	}

	return settings, nil
}

// Helper functions
func intToString(i int) string {
	return strconv.Itoa(i)
}

func stringToInt(s string) int {
	result, _ := strconv.Atoi(s)
	return result
}
