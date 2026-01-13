package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"chroniclecore/internal/store"
)

// BlacklistHandler manages app blacklist endpoints
type BlacklistHandler struct {
	store *store.Store
}

func NewBlacklistHandler(store *store.Store) *BlacklistHandler {
	return &BlacklistHandler{store: store}
}

// BlacklistEntry represents a blacklisted app
type BlacklistEntry struct {
	BlacklistID int64   `json:"blacklist_id"`
	AppID       int64   `json:"app_id"`
	AppName     string  `json:"app_name"`
	Reason      *string `json:"reason,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

// BlacklistCreate is the request body for adding to blacklist
type BlacklistCreate struct {
	AppName string  `json:"app_name"` // Can blacklist by app name
	AppID   *int64  `json:"app_id"`   // Or by app_id directly
	Reason  *string `json:"reason"`
}

// ListBlacklist returns all blacklisted apps
func (h *BlacklistHandler) ListBlacklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := `
		SELECT
			b.blacklist_id,
			b.app_id,
			da.app_name,
			b.reason,
			b.created_at
		FROM app_blacklist b
		JOIN dict_app da ON b.app_id = da.app_id
		ORDER BY da.app_name ASC
	`

	rows, err := h.store.GetDB().Query(query)
	if err != nil {
		log.Printf("Failed to query blacklist: %v", err)
		respondError(w, "Failed to query blacklist", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var entries []BlacklistEntry
	for rows.Next() {
		var e BlacklistEntry
		var reason sql.NullString
		err := rows.Scan(&e.BlacklistID, &e.AppID, &e.AppName, &reason, &e.CreatedAt)
		if err != nil {
			log.Printf("Failed to scan blacklist entry: %v", err)
			continue
		}
		if reason.Valid {
			e.Reason = &reason.String
		}
		entries = append(entries, e)
	}

	if entries == nil {
		entries = []BlacklistEntry{}
	}

	respondJSON(w, entries, http.StatusOK)
}

// AddToBlacklist adds an app to the blacklist
func (h *BlacklistHandler) AddToBlacklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input BlacklistCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var appID int64

	// Get app_id - either from input or by looking up app_name
	if input.AppID != nil {
		appID = *input.AppID
	} else if strings.TrimSpace(input.AppName) != "" {
		// Look up by app name
		err := h.store.GetDB().QueryRow(
			"SELECT app_id FROM dict_app WHERE app_name = ?",
			strings.TrimSpace(input.AppName),
		).Scan(&appID)
		if err == sql.ErrNoRows {
			respondError(w, "App not found", http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("Failed to find app: %v", err)
			respondError(w, "Failed to find app", http.StatusInternalServerError)
			return
		}
	} else {
		respondError(w, "app_id or app_name is required", http.StatusBadRequest)
		return
	}

	// Insert into blacklist
	result, err := h.store.GetDB().Exec(
		"INSERT INTO app_blacklist (app_id, reason) VALUES (?, ?)",
		appID,
		input.Reason,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			respondError(w, "App already blacklisted", http.StatusConflict)
			return
		}
		log.Printf("Failed to add to blacklist: %v", err)
		respondError(w, "Failed to add to blacklist", http.StatusInternalServerError)
		return
	}

	blacklistID, _ := result.LastInsertId()

	// Fetch created entry
	var entry BlacklistEntry
	var reason sql.NullString
	err = h.store.GetDB().QueryRow(`
		SELECT b.blacklist_id, b.app_id, da.app_name, b.reason, b.created_at
		FROM app_blacklist b
		JOIN dict_app da ON b.app_id = da.app_id
		WHERE b.blacklist_id = ?
	`, blacklistID).Scan(&entry.BlacklistID, &entry.AppID, &entry.AppName, &reason, &entry.CreatedAt)

	if err != nil {
		respondError(w, "Failed to fetch created entry", http.StatusInternalServerError)
		return
	}

	if reason.Valid {
		entry.Reason = &reason.String
	}

	respondJSON(w, entry, http.StatusCreated)
}

// RemoveFromBlacklist removes an app from the blacklist
func (h *BlacklistHandler) RemoveFromBlacklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract blacklist_id from path: /api/v1/blacklist/{id}
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		respondError(w, "Invalid path", http.StatusBadRequest)
		return
	}

	blacklistID, err := strconv.ParseInt(pathParts[3], 10, 64)
	if err != nil {
		respondError(w, "Invalid blacklist_id", http.StatusBadRequest)
		return
	}

	result, err := h.store.GetDB().Exec(
		"DELETE FROM app_blacklist WHERE blacklist_id = ?",
		blacklistID,
	)
	if err != nil {
		log.Printf("Failed to remove from blacklist: %v", err)
		respondError(w, "Failed to remove from blacklist", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		respondError(w, "Entry not found", http.StatusNotFound)
		return
	}

	respondJSON(w, map[string]bool{"success": true}, http.StatusOK)
}

// BlacklistAndDeleteBlocks blacklists an app AND deletes all associated blocks
func (h *BlacklistHandler) BlacklistAndDeleteBlocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input BlacklistCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var appID int64

	// Get app_id
	if input.AppID != nil {
		appID = *input.AppID
	} else if strings.TrimSpace(input.AppName) != "" {
		err := h.store.GetDB().QueryRow(
			"SELECT app_id FROM dict_app WHERE app_name = ?",
			strings.TrimSpace(input.AppName),
		).Scan(&appID)
		if err == sql.ErrNoRows {
			respondError(w, "App not found", http.StatusNotFound)
			return
		}
		if err != nil {
			respondError(w, "Failed to find app", http.StatusInternalServerError)
			return
		}
	} else {
		respondError(w, "app_id or app_name is required", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := h.store.GetDB().Begin()
	if err != nil {
		respondError(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// 1. Add to blacklist (ignore if already exists)
	_, err = tx.Exec(
		"INSERT OR IGNORE INTO app_blacklist (app_id, reason) VALUES (?, ?)",
		appID,
		input.Reason,
	)
	if err != nil {
		log.Printf("Failed to add to blacklist: %v", err)
		respondError(w, "Failed to add to blacklist", http.StatusInternalServerError)
		return
	}

	// 2. Delete all blocks with this app
	result, err := tx.Exec(
		"DELETE FROM block WHERE primary_app_id = ?",
		appID,
	)
	if err != nil {
		log.Printf("Failed to delete blocks: %v", err)
		respondError(w, "Failed to delete blocks", http.StatusInternalServerError)
		return
	}

	blocksDeleted, _ := result.RowsAffected()

	// 3. Delete raw events with this app
	result, err = tx.Exec(
		"DELETE FROM raw_event WHERE app_id = ?",
		appID,
	)
	if err != nil {
		log.Printf("Failed to delete raw events: %v", err)
		respondError(w, "Failed to delete raw events", http.StatusInternalServerError)
		return
	}

	eventsDeleted, _ := result.RowsAffected()

	// Commit transaction
	if err = tx.Commit(); err != nil {
		respondError(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Get app name for response
	var appName string
	h.store.GetDB().QueryRow("SELECT app_name FROM dict_app WHERE app_id = ?", appID).Scan(&appName)

	respondJSON(w, map[string]interface{}{
		"success":        true,
		"app_name":       appName,
		"blocks_deleted": blocksDeleted,
		"events_deleted": eventsDeleted,
	}, http.StatusOK)
}

// IsBlacklisted checks if an app is blacklisted (used internally)
func (h *BlacklistHandler) IsBlacklisted(appID int64) bool {
	var count int
	h.store.GetDB().QueryRow(
		"SELECT COUNT(*) FROM app_blacklist WHERE app_id = ?",
		appID,
	).Scan(&count)
	return count > 0
}
