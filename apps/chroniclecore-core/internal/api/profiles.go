package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"chroniclecore/internal/store"
)

// ProfileHandler manages profile-related endpoints
type ProfileHandler struct {
	store *store.Store
}

func NewProfileHandler(store *store.Store) *ProfileHandler {
	return &ProfileHandler{store: store}
}

// Client models

type Client struct {
	ClientID  int64  `json:"client_id"`
	Name      string `json:"name"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ClientCreate struct {
	Name string `json:"name"`
}

// Service models

type Service struct {
	ServiceID int64  `json:"service_id"`
	Name      string `json:"name"`
	IsActive  bool   `json:"is_active"`
}

type ServiceCreate struct {
	Name string `json:"name"`
}

// Rate models

type Rate struct {
	RateID           int64   `json:"rate_id"`
	Name             string  `json:"name"`
	CurrencyCode     string  `json:"currency_code"` // ISO 4217 3-letter code
	HourlyAmount     float64 `json:"hourly_amount"` // Converted from minor units
	HourlyMinorUnits int64   `json:"hourly_minor_units"`
	EffectiveFrom    *string `json:"effective_from,omitempty"`
	EffectiveTo      *string `json:"effective_to,omitempty"`
	IsActive         bool    `json:"is_active"`
}

type RateCreate struct {
	Name          string  `json:"name"`
	CurrencyCode  string  `json:"currency_code"` // ISO 4217 3-letter code
	HourlyAmount  float64 `json:"hourly_amount"` // Will be converted to minor units
	EffectiveFrom *string `json:"effective_from,omitempty"`
	EffectiveTo   *string `json:"effective_to,omitempty"`
}

// Profile models

type Profile struct {
	ProfileID    int64   `json:"profile_id"`
	Name         *string `json:"name,omitempty"` // Changed to pointer to handle NULL
	ClientName   string  `json:"client_name"`
	ProjectName  *string `json:"project_name,omitempty"`
	ServiceName  string  `json:"service_name"`
	RateName     string  `json:"rate_name"`
	RateAmount   float64 `json:"rate_amount"`
	CurrencyCode string  `json:"currency_code"` // ISO 4217 3-letter code
	IsActive     bool    `json:"is_active"`
}

type ProfileCreate struct {
	ClientID  int64   `json:"client_id"`
	ProjectID *int64  `json:"project_id,omitempty"`
	ServiceID int64   `json:"service_id"`
	RateID    int64   `json:"rate_id"`
	Name      *string `json:"name,omitempty"`
}

// Client endpoints

func (h *ProfileHandler) ListClients(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Query parameter: active_only (default: true)
	activeOnly := true
	if r.URL.Query().Get("active_only") == "false" {
		activeOnly = false
	}

	query := "SELECT client_id, name, is_active, created_at, updated_at FROM client"
	if activeOnly {
		query += " WHERE is_active = 1"
	}
	query += " ORDER BY name ASC"

	rows, err := h.store.GetDB().Query(query)
	if err != nil {
		respondError(w, "Failed to query clients", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var clients []Client
	for rows.Next() {
		var c Client
		err := rows.Scan(&c.ClientID, &c.Name, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			respondError(w, "Failed to scan client", http.StatusInternalServerError)
			return
		}
		clients = append(clients, c)
	}

	if clients == nil {
		clients = []Client{} // Return empty array, not null
	}

	respondJSON(w, clients, http.StatusOK)
}

func (h *ProfileHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input ClientCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate
	if strings.TrimSpace(input.Name) == "" {
		respondError(w, "Client name is required", http.StatusBadRequest)
		return
	}

	// Insert
	result, err := h.store.GetDB().Exec(
		"INSERT INTO client (name) VALUES (?)",
		strings.TrimSpace(input.Name),
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			respondError(w, "Client name already exists", http.StatusConflict)
			return
		}
		respondError(w, "Failed to create client", http.StatusInternalServerError)
		return
	}

	clientID, _ := result.LastInsertId()

	// Fetch created client
	var client Client
	err = h.store.GetDB().QueryRow(
		"SELECT client_id, name, is_active, created_at, updated_at FROM client WHERE client_id = ?",
		clientID,
	).Scan(&client.ClientID, &client.Name, &client.IsActive, &client.CreatedAt, &client.UpdatedAt)

	if err != nil {
		respondError(w, "Failed to fetch created client", http.StatusInternalServerError)
		return
	}

	respondJSON(w, client, http.StatusCreated)
}

// Service endpoints

func (h *ProfileHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := "SELECT service_id, name, is_active FROM service WHERE is_active = 1 ORDER BY name ASC"

	rows, err := h.store.GetDB().Query(query)
	if err != nil {
		respondError(w, "Failed to query services", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var services []Service
	for rows.Next() {
		var s Service
		err := rows.Scan(&s.ServiceID, &s.Name, &s.IsActive)
		if err != nil {
			respondError(w, "Failed to scan service", http.StatusInternalServerError)
			return
		}
		services = append(services, s)
	}

	if services == nil {
		services = []Service{}
	}

	respondJSON(w, services, http.StatusOK)
}

func (h *ProfileHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input ServiceCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(input.Name) == "" {
		respondError(w, "Service name is required", http.StatusBadRequest)
		return
	}

	result, err := h.store.GetDB().Exec(
		"INSERT INTO service (name) VALUES (?)",
		strings.TrimSpace(input.Name),
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			respondError(w, "Service name already exists", http.StatusConflict)
			return
		}
		respondError(w, "Failed to create service", http.StatusInternalServerError)
		return
	}

	serviceID, _ := result.LastInsertId()

	var service Service
	err = h.store.GetDB().QueryRow(
		"SELECT service_id, name, is_active FROM service WHERE service_id = ?",
		serviceID,
	).Scan(&service.ServiceID, &service.Name, &service.IsActive)

	if err != nil {
		respondError(w, "Failed to fetch created service", http.StatusInternalServerError)
		return
	}

	respondJSON(w, service, http.StatusCreated)
}

// Rate endpoints

func (h *ProfileHandler) ListRates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := `
		SELECT rate_id, name, currency_code, hourly_minor_units,
		       effective_from, effective_to, is_active
		FROM rate
		WHERE is_active = 1
		ORDER BY name ASC
	`

	rows, err := h.store.GetDB().Query(query)
	if err != nil {
		respondError(w, "Failed to query rates", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var rates []Rate
	for rows.Next() {
		var r Rate
		var effectiveFrom, effectiveTo sql.NullString
		err := rows.Scan(
			&r.RateID, &r.Name, &r.CurrencyCode, &r.HourlyMinorUnits,
			&effectiveFrom, &effectiveTo, &r.IsActive,
		)
		if err != nil {
			respondError(w, "Failed to scan rate", http.StatusInternalServerError)
			return
		}

		// Convert minor units to major units (e.g., cents to dollars)
		r.HourlyAmount = float64(r.HourlyMinorUnits) / 100.0

		if effectiveFrom.Valid {
			r.EffectiveFrom = &effectiveFrom.String
		}
		if effectiveTo.Valid {
			r.EffectiveTo = &effectiveTo.String
		}

		rates = append(rates, r)
	}

	if rates == nil {
		rates = []Rate{}
	}

	respondJSON(w, rates, http.StatusOK)
}

func (h *ProfileHandler) CreateRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input RateCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate
	if strings.TrimSpace(input.Name) == "" {
		respondError(w, "Rate name is required", http.StatusBadRequest)
		return
	}
	if input.CurrencyCode == "" {
		input.CurrencyCode = "USD" // Default to USD
	} else {
		// Validate currency code format
		if !ValidateCurrencyCode(input.CurrencyCode) {
			respondError(w, "Invalid currency code - must be 3-letter ISO 4217 code (e.g., USD, ZAR, EUR)", http.StatusBadRequest)
			return
		}
	}
	if input.HourlyAmount <= 0 {
		respondError(w, "Hourly amount must be positive", http.StatusBadRequest)
		return
	}

	// Convert to minor units (cents)
	minorUnits := int64(input.HourlyAmount * 100)

	result, err := h.store.GetDB().Exec(
		"INSERT INTO rate (name, currency_code, hourly_minor_units, effective_from, effective_to) VALUES (?, ?, ?, ?, ?)",
		strings.TrimSpace(input.Name),
		input.CurrencyCode,
		minorUnits,
		input.EffectiveFrom,
		input.EffectiveTo,
	)
	if err != nil {
		respondError(w, "Failed to create rate", http.StatusInternalServerError)
		return
	}

	rateID, _ := result.LastInsertId()

	var rate Rate
	var effectiveFrom, effectiveTo sql.NullString
	err = h.store.GetDB().QueryRow(
		"SELECT rate_id, name, currency_code, hourly_minor_units, effective_from, effective_to, is_active FROM rate WHERE rate_id = ?",
		rateID,
	).Scan(&rate.RateID, &rate.Name, &rate.CurrencyCode, &rate.HourlyMinorUnits, &effectiveFrom, &effectiveTo, &rate.IsActive)

	if err != nil {
		respondError(w, "Failed to fetch created rate", http.StatusInternalServerError)
		return
	}

	rate.HourlyAmount = float64(rate.HourlyMinorUnits) / 100.0
	if effectiveFrom.Valid {
		rate.EffectiveFrom = &effectiveFrom.String
	}
	if effectiveTo.Valid {
		rate.EffectiveTo = &effectiveTo.String
	}

	respondJSON(w, rate, http.StatusCreated)
}

// Profile endpoints

func (h *ProfileHandler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := `
		SELECT
			p.profile_id,
			p.name,
			c.name as client_name,
			pr.name as project_name,
			s.name as service_name,
			r.name as rate_name,
			r.hourly_minor_units,
			r.currency_code,
			p.is_active
		FROM profile p
		JOIN client c ON p.client_id = c.client_id
		LEFT JOIN project pr ON p.project_id = pr.project_id
		JOIN service s ON p.service_id = s.service_id
		JOIN rate r ON p.rate_id = r.rate_id
		WHERE p.is_active = 1
		ORDER BY c.name, pr.name, s.name
	`

	rows, err := h.store.GetDB().Query(query)
	if err != nil {
		respondError(w, "Failed to query profiles", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var profiles []Profile
	for rows.Next() {
		var p Profile
		var profileName sql.NullString
		var projectName sql.NullString
		var minorUnits int64

		err := rows.Scan(
			&p.ProfileID,
			&profileName,
			&p.ClientName,
			&projectName,
			&p.ServiceName,
			&p.RateName,
			&minorUnits,
			&p.CurrencyCode,
			&p.IsActive,
		)
		if err != nil {
			respondError(w, "Failed to scan profile", http.StatusInternalServerError)
			return
		}

		if profileName.Valid {
			p.Name = &profileName.String
		}
		if projectName.Valid {
			p.ProjectName = &projectName.String
		}

		p.RateAmount = float64(minorUnits) / 100.0

		profiles = append(profiles, p)
	}

	if profiles == nil {
		profiles = []Profile{}
	}

	respondJSON(w, profiles, http.StatusOK)
}

func (h *ProfileHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input ProfileCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate
	if input.ClientID <= 0 || input.ServiceID <= 0 || input.RateID <= 0 {
		respondError(w, "client_id, service_id, and rate_id are required", http.StatusBadRequest)
		return
	}

	// Insert
	result, err := h.store.GetDB().Exec(
		"INSERT INTO profile (client_id, project_id, service_id, rate_id, name) VALUES (?, ?, ?, ?, ?)",
		input.ClientID,
		input.ProjectID,
		input.ServiceID,
		input.RateID,
		input.Name,
	)
	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY") {
			respondError(w, "Invalid client_id, service_id, or rate_id", http.StatusBadRequest)
			return
		}
		respondError(w, "Failed to create profile", http.StatusInternalServerError)
		return
	}

	profileID, _ := result.LastInsertId()

	// Fetch created profile with joins
	var profile Profile
	var profileName sql.NullString
	var projectName sql.NullString
	var minorUnits int64

	err = h.store.GetDB().QueryRow(`
		SELECT
			p.profile_id,
			p.name,
			c.name,
			pr.name,
			s.name,
			r.name,
			r.hourly_minor_units,
			r.currency_code,
			p.is_active
		FROM profile p
		JOIN client c ON p.client_id = c.client_id
		LEFT JOIN project pr ON p.project_id = pr.project_id
		JOIN service s ON p.service_id = s.service_id
		JOIN rate r ON p.rate_id = r.rate_id
		WHERE p.profile_id = ?
	`, profileID).Scan(
		&profile.ProfileID,
		&profileName,
		&profile.ClientName,
		&projectName,
		&profile.ServiceName,
		&profile.RateName,
		&minorUnits,
		&profile.CurrencyCode,
		&profile.IsActive,
	)

	if err != nil {
		respondError(w, "Failed to fetch created profile", http.StatusInternalServerError)
		return
	}

	if profileName.Valid {
		profile.Name = &profileName.String
	}
	if projectName.Valid {
		profile.ProjectName = &projectName.String
	}
	profile.RateAmount = float64(minorUnits) / 100.0

	respondJSON(w, profile, http.StatusCreated)
}

// ProfileStats contains detailed statistics for a profile
type ProfileStats struct {
	ProfileID    int64   `json:"profile_id"`
	ClientName   string  `json:"client_name"`
	ProjectName  *string `json:"project_name,omitempty"`
	ServiceName  string  `json:"service_name"`
	RateName     string  `json:"rate_name"`
	RateAmount   float64 `json:"rate_amount"`
	CurrencyCode string  `json:"currency_code"`

	// Statistics
	TotalBlocks       int     `json:"total_blocks"`
	TotalMinutes      float64 `json:"total_minutes"`
	TotalHours        float64 `json:"total_hours"`
	BillableMinutes   float64 `json:"billable_minutes"`
	BillableHours     float64 `json:"billable_hours"`
	EstimatedBillable float64 `json:"estimated_billable"` // hours * rate
	LockedMinutes     float64 `json:"locked_minutes"`
	LockedHours       float64 `json:"locked_hours"`
	LockedBillable    float64 `json:"locked_billable"` // locked hours * rate

	// Recent blocks for detail view
	RecentBlocks []ProfileBlock `json:"recent_blocks,omitempty"`
}

// ProfileBlock represents a block in profile stats
type ProfileBlock struct {
	BlockID         int64   `json:"block_id"`
	TsStart         string  `json:"ts_start"`
	TsEnd           string  `json:"ts_end"`
	DurationMinutes float64 `json:"duration_minutes"`
	PrimaryAppName  string  `json:"primary_app_name"`
	TitleSummary    *string `json:"title_summary,omitempty"`
	Confidence      string  `json:"confidence"`
	Billable        bool    `json:"billable"`
	Locked          bool    `json:"locked"`
}

// GetProfileStats returns detailed statistics for a specific profile
func (h *ProfileHandler) GetProfileStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract profile_id from path: /api/v1/profiles/{id}/stats
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		respondError(w, "Invalid path", http.StatusBadRequest)
		return
	}

	profileID, err := strconv.ParseInt(pathParts[3], 10, 64)
	if err != nil {
		respondError(w, "Invalid profile_id", http.StatusBadRequest)
		return
	}

	// Get date range from query params (optional)
	params := r.URL.Query()
	startDate := params.Get("start_date")
	endDate := params.Get("end_date")
	includeBlocks := params.Get("include_blocks") == "true"

	// Fetch profile info
	var stats ProfileStats
	var projectName sql.NullString
	var minorUnits int64

	err = h.store.GetDB().QueryRow(`
		SELECT
			p.profile_id,
			c.name,
			pr.name,
			s.name,
			r.name,
			r.hourly_minor_units,
			r.currency_code
		FROM profile p
		JOIN client c ON p.client_id = c.client_id
		LEFT JOIN project pr ON p.project_id = pr.project_id
		JOIN service s ON p.service_id = s.service_id
		JOIN rate r ON p.rate_id = r.rate_id
		WHERE p.profile_id = ?
	`, profileID).Scan(
		&stats.ProfileID,
		&stats.ClientName,
		&projectName,
		&stats.ServiceName,
		&stats.RateName,
		&minorUnits,
		&stats.CurrencyCode,
	)

	if err == sql.ErrNoRows {
		respondError(w, "Profile not found", http.StatusNotFound)
		return
	}
	if err != nil {
		respondError(w, "Failed to fetch profile", http.StatusInternalServerError)
		return
	}

	if projectName.Valid {
		stats.ProjectName = &projectName.String
	}
	stats.RateAmount = float64(minorUnits) / 100.0

	// Build stats query
	// Activity-weighted billing: billable_minutes = duration * activity_score
	// This ensures only actual active work time is billed, not idle time within blocks
	statsQuery := `
		SELECT
			COUNT(*) as total_blocks,
			COALESCE(SUM((strftime('%s', ts_end) - strftime('%s', ts_start)) / 60.0), 0) as total_minutes,
			COALESCE(SUM(CASE WHEN billable = 1 THEN ((strftime('%s', ts_end) - strftime('%s', ts_start)) / 60.0) * COALESCE(activity_score, 1.0) ELSE 0 END), 0) as billable_minutes,
			COALESCE(SUM(CASE WHEN locked = 1 THEN ((strftime('%s', ts_end) - strftime('%s', ts_start)) / 60.0) * COALESCE(activity_score, 1.0) ELSE 0 END), 0) as locked_minutes
		FROM block
		WHERE profile_id = ?
	`
	var args []interface{}
	args = append(args, profileID)

	if startDate != "" && endDate != "" {
		statsQuery += " AND DATE(ts_start) >= ? AND DATE(ts_start) <= ?"
		args = append(args, startDate, endDate)
	}

	err = h.store.GetDB().QueryRow(statsQuery, args...).Scan(
		&stats.TotalBlocks,
		&stats.TotalMinutes,
		&stats.BillableMinutes,
		&stats.LockedMinutes,
	)
	if err != nil {
		respondError(w, "Failed to calculate stats", http.StatusInternalServerError)
		return
	}

	// Calculate derived values
	stats.TotalHours = stats.TotalMinutes / 60.0
	stats.BillableHours = stats.BillableMinutes / 60.0
	stats.LockedHours = stats.LockedMinutes / 60.0
	stats.EstimatedBillable = stats.BillableHours * stats.RateAmount
	stats.LockedBillable = stats.LockedHours * stats.RateAmount

	// Optionally include recent blocks
	if includeBlocks {
		blocksQuery := `
			SELECT
				b.block_id,
				b.ts_start,
				b.ts_end,
				(strftime('%s', b.ts_end) - strftime('%s', b.ts_start)) / 60.0 as duration_minutes,
				da.app_name,
				dt.title_text,
				b.confidence,
				b.billable,
				b.locked
			FROM block b
			JOIN dict_app da ON b.primary_app_id = da.app_id
			LEFT JOIN dict_title dt ON b.title_summary_id = dt.title_id
			WHERE b.profile_id = ?
		`
		var blockArgs []interface{}
		blockArgs = append(blockArgs, profileID)

		if startDate != "" && endDate != "" {
			blocksQuery += " AND DATE(b.ts_start) >= ? AND DATE(b.ts_start) <= ?"
			blockArgs = append(blockArgs, startDate, endDate)
		}

		blocksQuery += " ORDER BY b.ts_start DESC LIMIT 100"

		rows, err := h.store.GetDB().Query(blocksQuery, blockArgs...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var block ProfileBlock
				var titleSummary sql.NullString
				err := rows.Scan(
					&block.BlockID,
					&block.TsStart,
					&block.TsEnd,
					&block.DurationMinutes,
					&block.PrimaryAppName,
					&titleSummary,
					&block.Confidence,
					&block.Billable,
					&block.Locked,
				)
				if err != nil {
					continue
				}
				if titleSummary.Valid {
					block.TitleSummary = &titleSummary.String
				}
				stats.RecentBlocks = append(stats.RecentBlocks, block)
			}
		}

		if stats.RecentBlocks == nil {
			stats.RecentBlocks = []ProfileBlock{}
		}
	}

	respondJSON(w, stats, http.StatusOK)
}

func (h *ProfileHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract profile_id from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		respondError(w, "Invalid path", http.StatusBadRequest)
		return
	}

	profileID, err := strconv.ParseInt(pathParts[3], 10, 64)
	if err != nil {
		respondError(w, "Invalid profile_id", http.StatusBadRequest)
		return
	}

	// Soft delete (set is_active = 0)
	result, err := h.store.GetDB().Exec(
		"UPDATE profile SET is_active = 0 WHERE profile_id = ?",
		profileID,
	)
	if err != nil {
		respondError(w, "Failed to delete profile", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		respondError(w, "Profile not found", http.StatusNotFound)
		return
	}

	respondJSON(w, map[string]bool{"success": true}, http.StatusOK)
}

// Helper functions

func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    getErrorCode(status),
			"message": message,
		},
	})
}

func getErrorCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "INVALID_REQUEST"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	default:
		return "INTERNAL_ERROR"
	}
}
