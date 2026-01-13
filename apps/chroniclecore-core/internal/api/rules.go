package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"chroniclecore/internal/store"
)

// RuleHandler manages rule-related endpoints
type RuleHandler struct {
	store *store.Store
}

func NewRuleHandler(store *store.Store) *RuleHandler {
	return &RuleHandler{store: store}
}

// RuleDTO represents a rule for API responses
type RuleDTO struct {
	RuleID          int64   `json:"rule_id"`
	Name            string  `json:"name"`
	Priority        int     `json:"priority"`
	MatchType       string  `json:"match_type"`
	MatchValue      string  `json:"match_value"`
	TargetProfileID int64   `json:"target_profile_id"`
	TargetServiceID *int64  `json:"target_service_id,omitempty"`
	ConfidenceBoost int     `json:"confidence_boost"`
	Enabled         bool    `json:"enabled"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`

	// Enriched fields
	ProfileName     *string `json:"profile_name,omitempty"`
	ClientName      *string `json:"client_name,omitempty"`
	ServiceName     *string `json:"service_name,omitempty"`
}

// CreateRuleRequest represents a rule creation request
type CreateRuleRequest struct {
	Name            string  `json:"name"`
	Priority        int     `json:"priority"`
	MatchType       string  `json:"match_type"`
	MatchValue      string  `json:"match_value"`
	TargetProfileID int64   `json:"target_profile_id"`
	TargetServiceID *int64  `json:"target_service_id,omitempty"`
	ConfidenceBoost int     `json:"confidence_boost"`
	Enabled         *bool   `json:"enabled,omitempty"`
}

// UpdateRuleRequest represents a rule update request
type UpdateRuleRequest struct {
	Name            *string `json:"name,omitempty"`
	Priority        *int    `json:"priority,omitempty"`
	MatchType       *string `json:"match_type,omitempty"`
	MatchValue      *string `json:"match_value,omitempty"`
	TargetProfileID *int64  `json:"target_profile_id,omitempty"`
	TargetServiceID *int64  `json:"target_service_id,omitempty"`
	ConfidenceBoost *int    `json:"confidence_boost,omitempty"`
	Enabled         *bool   `json:"enabled,omitempty"`
}

// ListRules handles GET /api/v1/rules
func (h *RuleHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Query parameters
	params := r.URL.Query()
	enabledOnly := params.Get("enabled") != "false" // Default: true

	query := `
		SELECT
			r.rule_id,
			r.name,
			r.priority,
			r.match_type,
			r.match_value,
			r.target_profile_id,
			r.target_service_id,
			r.confidence_boost,
			r.enabled,
			r.created_at,
			r.updated_at,
			c.name as client_name,
			s.name as service_name
		FROM rule r
		JOIN profile p ON r.target_profile_id = p.profile_id
		JOIN client c ON p.client_id = c.client_id
		LEFT JOIN service s ON p.service_id = s.service_id
	`

	if enabledOnly {
		query += " WHERE r.enabled = 1"
	}

	query += " ORDER BY r.priority DESC, r.rule_id ASC"

	rows, err := h.store.GetDB().Query(query)
	if err != nil {
		respondError(w, "Failed to query rules", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var rules []RuleDTO

	for rows.Next() {
		var r RuleDTO
		var targetServiceID sql.NullInt64
		var clientName, serviceName sql.NullString
		var enabled int

		err := rows.Scan(
			&r.RuleID,
			&r.Name,
			&r.Priority,
			&r.MatchType,
			&r.MatchValue,
			&r.TargetProfileID,
			&targetServiceID,
			&r.ConfidenceBoost,
			&enabled,
			&r.CreatedAt,
			&r.UpdatedAt,
			&clientName,
			&serviceName,
		)

		if err != nil {
			continue
		}

		r.Enabled = enabled == 1

		if targetServiceID.Valid {
			sid := targetServiceID.Int64
			r.TargetServiceID = &sid
		}

		if clientName.Valid {
			r.ClientName = &clientName.String
		}

		if serviceName.Valid {
			r.ServiceName = &serviceName.String
		}

		rules = append(rules, r)
	}

	if rules == nil {
		rules = []RuleDTO{}
	}

	respondJSON(w, rules, http.StatusOK)
}

// CreateRule handles POST /api/v1/rules
func (h *RuleHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		respondError(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.MatchType == "" {
		respondError(w, "match_type is required", http.StatusBadRequest)
		return
	}

	if req.MatchValue == "" {
		respondError(w, "match_value is required", http.StatusBadRequest)
		return
	}

	if req.TargetProfileID == 0 {
		respondError(w, "target_profile_id is required", http.StatusBadRequest)
		return
	}

	// Validate match_type
	validMatchTypes := []string{"APP", "DOMAIN", "TITLE_REGEX", "KEYWORD", "COMPOSITE"}
	if !contains(validMatchTypes, req.MatchType) {
		respondError(w, "match_type must be one of: APP, DOMAIN, TITLE_REGEX, KEYWORD, COMPOSITE", http.StatusBadRequest)
		return
	}

	// Validate regex if match_type is TITLE_REGEX
	if req.MatchType == "TITLE_REGEX" {
		if _, err := regexp.Compile(req.MatchValue); err != nil {
			respondError(w, "Invalid regex pattern: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Verify target_profile_id exists
	var profileExists int
	err := h.store.GetDB().QueryRow(
		"SELECT COUNT(*) FROM profile WHERE profile_id = ? AND is_active = 1",
		req.TargetProfileID,
	).Scan(&profileExists)

	if err != nil || profileExists == 0 {
		respondError(w, "target_profile_id does not exist or is inactive", http.StatusBadRequest)
		return
	}

	// Verify target_service_id if provided
	if req.TargetServiceID != nil {
		var serviceExists int
		err := h.store.GetDB().QueryRow(
			"SELECT COUNT(*) FROM service WHERE service_id = ? AND is_active = 1",
			*req.TargetServiceID,
		).Scan(&serviceExists)

		if err != nil || serviceExists == 0 {
			respondError(w, "target_service_id does not exist or is inactive", http.StatusBadRequest)
			return
		}
	}

	// Default enabled to true if not specified
	enabled := 1
	if req.Enabled != nil && !*req.Enabled {
		enabled = 0
	}

	// Insert rule
	result, err := h.store.GetDB().Exec(`
		INSERT INTO rule (name, priority, match_type, match_value, target_profile_id, target_service_id, confidence_boost, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		req.Name,
		req.Priority,
		req.MatchType,
		req.MatchValue,
		req.TargetProfileID,
		req.TargetServiceID,
		req.ConfidenceBoost,
		enabled,
	)

	if err != nil {
		respondError(w, "Failed to create rule", http.StatusInternalServerError)
		return
	}

	ruleID, _ := result.LastInsertId()

	// Fetch created rule
	rules := h.getRulesByIDs([]int64{ruleID})
	if len(rules) == 0 {
		respondError(w, "Failed to fetch created rule", http.StatusInternalServerError)
		return
	}

	respondJSON(w, rules[0], http.StatusCreated)
}

// UpdateRule handles PUT /api/v1/rules/{id}
func (h *RuleHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract rule_id from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		respondError(w, "Invalid path", http.StatusBadRequest)
		return
	}

	ruleID, err := strconv.ParseInt(pathParts[3], 10, 64)
	if err != nil {
		respondError(w, "Invalid rule_id", http.StatusBadRequest)
		return
	}

	var req UpdateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Check if rule exists
	var exists int
	err = h.store.GetDB().QueryRow("SELECT COUNT(*) FROM rule WHERE rule_id = ?", ruleID).Scan(&exists)
	if err != nil || exists == 0 {
		respondError(w, "Rule not found", http.StatusNotFound)
		return
	}

	// Build dynamic UPDATE query
	updates := []string{}
	args := []interface{}{}

	if req.Name != nil {
		updates = append(updates, "name = ?")
		args = append(args, *req.Name)
	}

	if req.Priority != nil {
		updates = append(updates, "priority = ?")
		args = append(args, *req.Priority)
	}

	if req.MatchType != nil {
		validMatchTypes := []string{"APP", "DOMAIN", "TITLE_REGEX", "KEYWORD", "COMPOSITE"}
		if !contains(validMatchTypes, *req.MatchType) {
			respondError(w, "match_type must be one of: APP, DOMAIN, TITLE_REGEX, KEYWORD, COMPOSITE", http.StatusBadRequest)
			return
		}
		updates = append(updates, "match_type = ?")
		args = append(args, *req.MatchType)
	}

	if req.MatchValue != nil {
		// Validate regex if updating to TITLE_REGEX
		currentMatchType := ""
		if req.MatchType != nil {
			currentMatchType = *req.MatchType
		} else {
			h.store.GetDB().QueryRow("SELECT match_type FROM rule WHERE rule_id = ?", ruleID).Scan(&currentMatchType)
		}

		if currentMatchType == "TITLE_REGEX" {
			if _, err := regexp.Compile(*req.MatchValue); err != nil {
				respondError(w, "Invalid regex pattern: "+err.Error(), http.StatusBadRequest)
				return
			}
		}

		updates = append(updates, "match_value = ?")
		args = append(args, *req.MatchValue)
	}

	if req.TargetProfileID != nil {
		// Verify profile exists
		var profileExists int
		err := h.store.GetDB().QueryRow(
			"SELECT COUNT(*) FROM profile WHERE profile_id = ? AND is_active = 1",
			*req.TargetProfileID,
		).Scan(&profileExists)

		if err != nil || profileExists == 0 {
			respondError(w, "target_profile_id does not exist or is inactive", http.StatusBadRequest)
			return
		}

		updates = append(updates, "target_profile_id = ?")
		args = append(args, *req.TargetProfileID)
	}

	if req.TargetServiceID != nil {
		if *req.TargetServiceID == 0 {
			updates = append(updates, "target_service_id = NULL")
		} else {
			// Verify service exists
			var serviceExists int
			err := h.store.GetDB().QueryRow(
				"SELECT COUNT(*) FROM service WHERE service_id = ? AND is_active = 1",
				*req.TargetServiceID,
			).Scan(&serviceExists)

			if err != nil || serviceExists == 0 {
				respondError(w, "target_service_id does not exist or is inactive", http.StatusBadRequest)
				return
			}

			updates = append(updates, "target_service_id = ?")
			args = append(args, *req.TargetServiceID)
		}
	}

	if req.ConfidenceBoost != nil {
		updates = append(updates, "confidence_boost = ?")
		args = append(args, *req.ConfidenceBoost)
	}

	if req.Enabled != nil {
		enabled := 0
		if *req.Enabled {
			enabled = 1
		}
		updates = append(updates, "enabled = ?")
		args = append(args, enabled)
	}

	if len(updates) == 0 {
		respondError(w, "No fields to update", http.StatusBadRequest)
		return
	}

	// Add rule_id to args
	args = append(args, ruleID)

	query := "UPDATE rule SET " + strings.Join(updates, ", ") + " WHERE rule_id = ?"
	_, err = h.store.GetDB().Exec(query, args...)

	if err != nil {
		respondError(w, "Failed to update rule", http.StatusInternalServerError)
		return
	}

	// Fetch updated rule
	rules := h.getRulesByIDs([]int64{ruleID})
	if len(rules) == 0 {
		respondError(w, "Failed to fetch updated rule", http.StatusInternalServerError)
		return
	}

	respondJSON(w, rules[0], http.StatusOK)
}

// DeleteRule handles DELETE /api/v1/rules/{id}
func (h *RuleHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract rule_id from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		respondError(w, "Invalid path", http.StatusBadRequest)
		return
	}

	ruleID, err := strconv.ParseInt(pathParts[3], 10, 64)
	if err != nil {
		respondError(w, "Invalid rule_id", http.StatusBadRequest)
		return
	}

	// Hard delete (can change to soft delete by setting enabled=0 if preferred)
	result, err := h.store.GetDB().Exec("DELETE FROM rule WHERE rule_id = ?", ruleID)
	if err != nil {
		respondError(w, "Failed to delete rule", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		respondError(w, "Rule not found", http.StatusNotFound)
		return
	}

	respondJSON(w, map[string]bool{"success": true}, http.StatusOK)
}

// getRulesByIDs fetches rules by IDs (helper for returning created/updated rules)
func (h *RuleHandler) getRulesByIDs(ruleIDs []int64) []RuleDTO {
	if len(ruleIDs) == 0 {
		return []RuleDTO{}
	}

	placeholders := make([]string, len(ruleIDs))
	args := make([]interface{}, len(ruleIDs))
	for i, id := range ruleIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `
		SELECT
			r.rule_id,
			r.name,
			r.priority,
			r.match_type,
			r.match_value,
			r.target_profile_id,
			r.target_service_id,
			r.confidence_boost,
			r.enabled,
			r.created_at,
			r.updated_at,
			c.name as client_name,
			s.name as service_name
		FROM rule r
		JOIN profile p ON r.target_profile_id = p.profile_id
		JOIN client c ON p.client_id = c.client_id
		LEFT JOIN service s ON p.service_id = s.service_id
		WHERE r.rule_id IN (` + strings.Join(placeholders, ",") + `)
	`

	rows, err := h.store.GetDB().Query(query, args...)
	if err != nil {
		return []RuleDTO{}
	}
	defer rows.Close()

	var rules []RuleDTO

	for rows.Next() {
		var r RuleDTO
		var targetServiceID sql.NullInt64
		var clientName, serviceName sql.NullString
		var enabled int

		err := rows.Scan(
			&r.RuleID,
			&r.Name,
			&r.Priority,
			&r.MatchType,
			&r.MatchValue,
			&r.TargetProfileID,
			&targetServiceID,
			&r.ConfidenceBoost,
			&enabled,
			&r.CreatedAt,
			&r.UpdatedAt,
			&clientName,
			&serviceName,
		)

		if err != nil {
			continue
		}

		r.Enabled = enabled == 1

		if targetServiceID.Valid {
			sid := targetServiceID.Int64
			r.TargetServiceID = &sid
		}

		if clientName.Valid {
			r.ClientName = &clientName.String
		}

		if serviceName.Valid {
			r.ServiceName = &serviceName.String
		}

		rules = append(rules, r)
	}

	return rules
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
