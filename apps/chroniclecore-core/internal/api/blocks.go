package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"chroniclecore/internal/store"
)

// BlockHandler manages block-related endpoints
type BlockHandler struct {
	store *store.Store
}

func NewBlockHandler(store *store.Store) *BlockHandler {
	return &BlockHandler{store: store}
}

// BlockDTO represents a block with enriched data for API responses
type BlockDTO struct {
	BlockID         int64   `json:"block_id"`
	AppID           int64   `json:"app_id"` // Added for grouping
	TsStart         string  `json:"ts_start"`
	TsEnd           string  `json:"ts_end"`
	DurationSeconds float64 `json:"duration_seconds"` // Added for precision
	DurationMinutes float64 `json:"duration_minutes"`
	DurationHours   float64 `json:"duration_hours"`

	// App/Title/Domain info
	PrimaryAppName string  `json:"primary_app_name"`
	PrimaryDomain  *string `json:"primary_domain,omitempty"`
	TitleSummary   *string `json:"title_summary,omitempty"`

	// Profile assignment
	ProfileID   *int64  `json:"profile_id,omitempty"`
	ClientName  *string `json:"client_name,omitempty"`
	ProjectName *string `json:"project_name,omitempty"`
	ServiceName *string `json:"service_name,omitempty"`

	// Status
	Confidence string `json:"confidence"`
	Billable   bool   `json:"billable"`
	Locked     bool   `json:"locked"`

	// Metadata
	Notes         *string  `json:"notes,omitempty"`
	Description   *string  `json:"description,omitempty"`
	ActivityScore *float64 `json:"activity_score,omitempty"` // Derived from metadata
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// GroupedBlock represents blocks grouped by app+title context
type GroupedBlock struct {
	GroupKey       string     `json:"group_key"` // Unique key for this group
	PrimaryAppName string     `json:"primary_app_name"`
	AppID          int64      `json:"app_id"`
	TitleContext   string     `json:"title_context"` // Common title/context
	TotalMinutes   float64    `json:"total_minutes"`
	TotalHours     float64    `json:"total_hours"`
	BlockCount     int        `json:"block_count"`
	FirstTs        string     `json:"first_ts"` // Earliest block
	LastTs         string     `json:"last_ts"`  // Latest block
	Blocks         []BlockDTO `json:"blocks"`   // Individual blocks in this group
}

// PaginatedResponse wraps results with pagination info
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ListBlocks handles GET /api/v1/blocks with filters
func (h *BlockHandler) ListBlocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// ... [existing param parsing code omitted for brevity but I must keep it if I replace the whole function, or use partial replace] ...
	// Since I can't partial replace nicely inside a big function without context, I'll assume I'm editing around the query.

	// Params parsing logic
	params := r.URL.Query()

	// Date filter (default: today)
	dateStr := params.Get("date")
	var filterDate time.Time
	if dateStr != "" {
		var err error
		filterDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			respondError(w, "Invalid date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	} else {
		filterDate = time.Now()
	}

	// Date range filter
	startDateStr := params.Get("start_date")
	endDateStr := params.Get("end_date")

	// Profile filter
	profileIDStr := params.Get("profile_id")
	var profileID *int64
	if profileIDStr != "" {
		pid, err := strconv.ParseInt(profileIDStr, 10, 64)
		if err != nil {
			respondError(w, "Invalid profile_id", http.StatusBadRequest)
			return
		}
		profileID = &pid
	}

	// Unassigned filter
	unassignedOnly := params.Get("unassigned") == "true"

	// Low confidence filter (needs review)
	needsReview := params.Get("needs_review") == "true"

	// Build query
	query := `
		SELECT
			b.block_id,
			b.ts_start,
			b.ts_end,
			b.confidence,
			b.billable,
			b.locked,
			b.notes,
			b.description,
			b.metadata,
			b.created_at,
			b.updated_at,
			b.profile_id,
			da.app_name as primary_app_name,
			dd.domain_text as primary_domain,
			dt.title_text as title_summary,
			c.name as client_name,
			pr.name as project_name,
			s.name as service_name
		FROM block b
		JOIN dict_app da ON b.primary_app_id = da.app_id
		LEFT JOIN dict_domain dd ON b.primary_domain_id = dd.domain_id
		LEFT JOIN dict_title dt ON b.title_summary_id = dt.title_id
		LEFT JOIN profile p ON b.profile_id = p.profile_id
		LEFT JOIN client c ON p.client_id = c.client_id
		LEFT JOIN project pr ON p.project_id = pr.project_id
		LEFT JOIN service s ON p.service_id = s.service_id
		WHERE 1=1
	`

	var args []interface{}

	// Apply date filter
	if startDateStr != "" && endDateStr != "" {
		query += " AND DATE(b.ts_start) >= ? AND DATE(b.ts_start) <= ?"
		args = append(args, startDateStr, endDateStr)
	} else if dateStr != "" {
		query += " AND DATE(b.ts_start) = ?"
		args = append(args, filterDate.Format("2006-01-02"))
	}

	// Apply profile filter
	if profileID != nil {
		query += " AND b.profile_id = ?"
		args = append(args, *profileID)
	}

	// Apply unassigned filter
	if unassignedOnly {
		query += " AND b.profile_id IS NULL"
	}

	// Apply needs review filter
	if needsReview {
		query += " AND (b.profile_id IS NULL OR b.confidence = 'LOW')"
	}

	query += " ORDER BY b.ts_start DESC"

	// Add limit
	limit := params.Get("limit")
	// ... existing limit logic ...
	if limit != "" {
		limitInt, err := strconv.Atoi(limit)
		if err == nil && limitInt > 0 && limitInt <= 1000 {
			query += " LIMIT ?"
			args = append(args, limitInt)
		}
	} else {
		query += " LIMIT 100" // Default limit
	}

	// Execute query
	rows, err := h.store.GetDB().Query(query, args...)
	if err != nil {
		log.Printf("Failed to query blocks: %v", err)
		respondError(w, "Failed to query blocks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var blocks []BlockDTO

	for rows.Next() {
		var b BlockDTO
		var tsStart, tsEnd, createdAt, updatedAt string
		var profileID sql.NullInt64
		var primaryDomain, titleSummary, clientName, projectName, serviceName sql.NullString
		var notes, description, metadata sql.NullString

		err := rows.Scan(
			&b.BlockID,
			&tsStart,
			&tsEnd,
			&b.Confidence,
			&b.Billable,
			&b.Locked,
			&notes,
			&description,
			&metadata,
			&createdAt,
			&updatedAt,
			&profileID,
			&b.PrimaryAppName,
			&primaryDomain,
			&titleSummary,
			&clientName,
			&projectName,
			&serviceName,
		)

		if err != nil {
			log.Printf("Failed to scan block: %v", err)
			continue
		}

		// Parse timestamps
		start, _ := time.Parse(time.RFC3339, tsStart)
		end, _ := time.Parse(time.RFC3339, tsEnd)

		// Calculate duration
		durationSeconds := end.Sub(start).Seconds()
		b.DurationSeconds = durationSeconds
		b.DurationMinutes = durationSeconds / 60.0
		b.DurationHours = durationSeconds / 3600.0

		// Set timestamps
		b.TsStart = tsStart
		b.TsEnd = tsEnd
		b.CreatedAt = createdAt
		b.UpdatedAt = updatedAt

		// Set nullable fields
		if profileID.Valid {
			pid := profileID.Int64
			b.ProfileID = &pid
		}
		if primaryDomain.Valid {
			b.PrimaryDomain = &primaryDomain.String
		}
		if titleSummary.Valid {
			b.TitleSummary = &titleSummary.String
		}
		if clientName.Valid {
			b.ClientName = &clientName.String
		}
		if projectName.Valid {
			b.ProjectName = &projectName.String
		}
		if serviceName.Valid {
			b.ServiceName = &serviceName.String
		}
		if notes.Valid {
			b.Notes = &notes.String
		}
		if description.Valid {
			b.Description = &description.String
		}

		// Parse Activity Score from metadata
		if metadata.Valid {
			var metaMap map[string]interface{}
			if err := json.Unmarshal([]byte(metadata.String), &metaMap); err == nil {
				if score, ok := metaMap["avg_activity_score"].(float64); ok {
					b.ActivityScore = &score
				}
			}
		}

		blocks = append(blocks, b)
	}

	if blocks == nil {
		blocks = []BlockDTO{} // Return empty array, not null
	}

	respondJSON(w, blocks, http.StatusOK)
}

// ListGroupedBlocks handles GET /api/v1/blocks/grouped - returns blocks grouped by app+title context with pagination
func (h *BlockHandler) ListGroupedBlocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	params := r.URL.Query()

	// Pagination params
	page := 1
	perPage := 20
	if p := params.Get("page"); p != "" {
		if pInt, err := strconv.Atoi(p); err == nil && pInt > 0 {
			page = pInt
		}
	}
	if pp := params.Get("per_page"); pp != "" {
		if ppInt, err := strconv.Atoi(pp); err == nil && ppInt > 0 && ppInt <= 100 {
			perPage = ppInt
		}
	}

	// Filter: needs_review only
	needsReview := params.Get("needs_review") == "true"

	// Date range filter
	startDate := params.Get("start_date")
	endDate := params.Get("end_date")

	// First, get all blocks matching the filter (we'll group them in Go)
	query := `
		SELECT
			b.block_id,
			b.ts_start,
			b.ts_end,
			b.confidence,
			b.billable,
			b.locked,
			b.notes,
			b.description,
			b.metadata,
			b.created_at,
			b.updated_at,
			b.profile_id,
			b.primary_app_id,
			da.app_name as primary_app_name,
			dd.domain_text as primary_domain,
			dt.title_text as title_summary,
			c.name as client_name,
			pr.name as project_name,
			s.name as service_name,
			ms.payload_json,
			ms.confidence as ml_confidence
		FROM block b
		JOIN dict_app da ON b.primary_app_id = da.app_id
		LEFT JOIN dict_domain dd ON b.primary_domain_id = dd.domain_id
		LEFT JOIN dict_title dt ON b.title_summary_id = dt.title_id
		LEFT JOIN profile p ON b.profile_id = p.profile_id
		LEFT JOIN client c ON p.client_id = c.client_id
		LEFT JOIN project pr ON p.project_id = pr.project_id
		LEFT JOIN service s ON p.service_id = s.service_id
		LEFT JOIN app_blacklist abl ON b.primary_app_id = abl.app_id
		LEFT JOIN ml_suggestion ms ON b.block_id = ms.entity_id AND ms.entity_type = 'BLOCK' AND ms.status = 'PENDING'
		WHERE abl.app_id IS NULL
	`

	var args []interface{}

	// Apply needs_review filter
	if needsReview {
		query += " AND (b.profile_id IS NULL OR b.confidence = 'LOW')"
	}

	// Apply date filter
	if startDate != "" && endDate != "" {
		query += " AND DATE(b.ts_start) >= ? AND DATE(b.ts_start) <= ?"
		args = append(args, startDate, endDate)
	}

	query += " ORDER BY b.ts_start DESC"

	rows, err := h.store.GetDB().Query(query, args...)
	if err != nil {
		log.Printf("Failed to query blocks: %v", err)
		respondError(w, "Failed to query blocks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Group blocks by app_name + title_context
	groupsMap := make(map[string]*GroupedBlock)
	var groupOrder []string // Maintain order by first occurrence

	for rows.Next() {
		var b BlockDTO
		var tsStart, tsEnd, createdAt, updatedAt string
		var profileID sql.NullInt64
		var appID int64
		var primaryDomain, titleSummary, clientName, projectName, serviceName sql.NullString
		var notes, description, metadata sql.NullString
		var mlPayload sql.NullString
		var mlConfidence sql.NullFloat64

		err := rows.Scan(
			&b.BlockID,
			&tsStart,
			&tsEnd,
			&b.Confidence,
			&b.Billable,
			&b.Locked,
			&notes,
			&description,
			&metadata,
			&createdAt,
			&updatedAt,
			&profileID,
			&appID,
			&b.PrimaryAppName,
			&primaryDomain,
			&titleSummary,
			&clientName,
			&projectName,
			&serviceName,
			&mlPayload,
			&mlConfidence,
		)

		if err != nil {
			log.Printf("Failed to scan block: %v", err)
			continue
		}

		// Parse timestamps and calculate duration
		start, _ := time.Parse(time.RFC3339, tsStart)
		end, _ := time.Parse(time.RFC3339, tsEnd)
		durationSeconds := end.Sub(start).Seconds()
		b.DurationSeconds = durationSeconds
		b.DurationMinutes = durationSeconds / 60.0
		b.DurationHours = durationSeconds / 3600.0
		b.TsStart = tsStart
		b.TsEnd = tsEnd
		b.CreatedAt = createdAt
		b.UpdatedAt = updatedAt
		b.AppID = appID

		// Set nullable fields
		if profileID.Valid {
			pid := profileID.Int64
			b.ProfileID = &pid
		} else if mlPayload.Valid {
			// Apply ML Suggestion if no profile assigned
			var payload map[string]interface{}
			if err := json.Unmarshal([]byte(mlPayload.String), &payload); err == nil {
				if predictedID, ok := payload["predicted_profile_id"].(float64); ok {
					pid := int64(predictedID)
					b.ProfileID = &pid

					// Use ML confidence (convert to string enum)
					confVal := 0.0
					if mlConfidence.Valid {
						confVal = mlConfidence.Float64
					}

					if confVal >= 0.8 {
						b.Confidence = "ML_HIGH"
					} else if confVal >= 0.6 {
						b.Confidence = "ML_MEDIUM"
					} else {
						b.Confidence = "ML_LOW"
					}
				}
			}
		}
		if primaryDomain.Valid {
			b.PrimaryDomain = &primaryDomain.String
		}
		titleContext := ""
		if titleSummary.Valid {
			b.TitleSummary = &titleSummary.String
			titleContext = titleSummary.String
		}
		if clientName.Valid {
			b.ClientName = &clientName.String
		}
		if projectName.Valid {
			b.ProjectName = &projectName.String
		}
		if serviceName.Valid {
			b.ServiceName = &serviceName.String
		}
		if notes.Valid {
			b.Notes = &notes.String
		}
		if description.Valid {
			b.Description = &description.String
		}

		// Parse Activity Score from metadata
		if metadata.Valid {
			var metaMap map[string]interface{}
			if err := json.Unmarshal([]byte(metadata.String), &metaMap); err == nil {
				if score, ok := metaMap["avg_activity_score"].(float64); ok {
					b.ActivityScore = &score
				}
			}
		}

		// Generate group key: app_name + extracted context from title
		// For Xero, we extract company name from title like "White Cat Studios | Xero"
		groupKey := generateGroupKey(b.PrimaryAppName, titleContext)

		if group, exists := groupsMap[groupKey]; exists {
			// Add to existing group
			group.Blocks = append(group.Blocks, b)
			group.TotalMinutes += b.DurationMinutes
			group.TotalHours += b.DurationHours
			group.BlockCount++
			// Update time range
			if b.TsStart < group.FirstTs {
				group.FirstTs = b.TsStart
			}
			if b.TsEnd > group.LastTs {
				group.LastTs = b.TsEnd
			}
		} else {
			// Create new group
			groupsMap[groupKey] = &GroupedBlock{
				GroupKey:       groupKey,
				PrimaryAppName: b.PrimaryAppName,
				AppID:          appID,
				TitleContext:   extractTitleContext(b.PrimaryAppName, titleContext),
				TotalMinutes:   b.DurationMinutes,
				TotalHours:     b.DurationHours,
				BlockCount:     1,
				FirstTs:        b.TsStart,
				LastTs:         b.TsEnd,
				Blocks:         []BlockDTO{b},
			}
			groupOrder = append(groupOrder, groupKey)
		}
	}

	// Convert map to slice in order
	var groups []GroupedBlock
	for _, key := range groupOrder {
		groups = append(groups, *groupsMap[key])
	}

	// Calculate pagination
	total := len(groups)
	totalPages := (total + perPage - 1) / perPage
	startIdx := (page - 1) * perPage
	endIdx := startIdx + perPage
	if endIdx > total {
		endIdx = total
	}

	var pagedGroups []GroupedBlock
	if startIdx < total {
		pagedGroups = groups[startIdx:endIdx]
	} else {
		pagedGroups = []GroupedBlock{}
	}

	response := PaginatedResponse{
		Data: pagedGroups,
		Pagination: Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	respondJSON(w, response, http.StatusOK)
}

// generateGroupKey creates a unique key for grouping blocks
func generateGroupKey(appName, titleContext string) string {
	// Extract meaningful context from title
	context := extractTitleContext(appName, titleContext)
	return appName + "|" + context
}

// extractTitleContext extracts the meaningful context from a window title
func extractTitleContext(appName, title string) string {
	if title == "" {
		return "General"
	}

	appLower := strings.ToLower(appName)

	// Special handling for common apps
	// Special handling for browser apps to extract web app context
	isBrowser := strings.Contains(appLower, "chrome") || strings.Contains(appLower, "edge") || strings.Contains(appLower, "firefox") || strings.Contains(appLower, "opera") || strings.Contains(appLower, "brave")

	if isBrowser {
		// PRODUCTIVITY APPS DETECTION
		// Detect Xero
		if strings.Contains(strings.ToLower(title), "xero") {
			return strings.TrimSpace(strings.Split(title, "|")[0]) // "Invoice 123 | Xero" -> "Invoice 123"
		}
		// Detect Gmail
		if strings.Contains(strings.ToLower(title), "gmail") {
			return "Gmail"
		}
		// Detect Outlook
		if strings.Contains(strings.ToLower(title), "outlook") {
			return "Outlook"
		}
		// Detect Google Docs/Sheets
		if strings.Contains(strings.ToLower(title), "google docs") {
			return "Google Docs"
		}
		if strings.Contains(strings.ToLower(title), "google sheets") {
			return "Google Sheets"
		}

		// Generic browser extraction
		parts := strings.Split(title, "-")
		if len(parts) > 1 {
			// Last part is usually the browser name, take the second to last part or first part
			candidate := strings.TrimSpace(parts[len(parts)-2])
			if candidate != "" {
				return candidate
			}
			return strings.TrimSpace(parts[0])
		}
		return title
	}

	// EXCEL
	if strings.Contains(appLower, "excel") {
		// "WorkbookName.xlsx - Excel"
		parts := strings.Split(title, "-")
		if len(parts) > 0 {
			name := strings.TrimSpace(parts[0])
			name = strings.TrimSuffix(name, ".xlsx")
			name = strings.TrimSuffix(name, ".xls")
			return name
		}
		return "Excel - General"
	}

	// WORD
	if strings.Contains(appLower, "word") {
		parts := strings.Split(title, "-")
		if len(parts) > 0 {
			name := strings.TrimSpace(parts[0])
			name = strings.TrimSuffix(name, ".docx")
			name = strings.TrimSuffix(name, ".doc")
			return name
		}
		return "Word - General"
	}

	// GENERIC FALLBACK
	// For other apps, use first part before " - " if present
	if idx := strings.Index(title, " - "); idx > 0 {
		return strings.TrimSpace(title[:idx])
	}
	// Truncate long titles
	if len(title) > 80 {
		return title[:80] + "..."
	}
	return title
}

// ReassignRequest represents a block reassignment request
type ReassignRequest struct {
	ProfileID  *int64 `json:"profile_id"` // Null to unassign
	Confidence string `json:"confidence"` // Optional: HIGH, MEDIUM, LOW (default: HIGH)
}

// ReassignBlock handles POST /api/v1/blocks/{id}/reassign
func (h *BlockHandler) ReassignBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract block_id from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		respondError(w, "Invalid path", http.StatusBadRequest)
		return
	}

	blockID, err := strconv.ParseInt(pathParts[3], 10, 64)
	if err != nil {
		respondError(w, "Invalid block_id", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req ReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Set default confidence
	if req.Confidence == "" {
		req.Confidence = "HIGH" // Manual assignment = HIGH confidence
	}

	// Validate confidence
	if req.Confidence != "HIGH" && req.Confidence != "MEDIUM" && req.Confidence != "LOW" {
		respondError(w, "confidence must be HIGH, MEDIUM, or LOW", http.StatusBadRequest)
		return
	}

	// Get current block state for audit log
	var oldProfileID sql.NullInt64
	var oldConfidence string
	err = h.store.GetDB().QueryRow(
		"SELECT profile_id, confidence FROM block WHERE block_id = ?",
		blockID,
	).Scan(&oldProfileID, &oldConfidence)

	if err == sql.ErrNoRows {
		respondError(w, "Block not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Failed to get block: %v", err)
		respondError(w, "Failed to get block", http.StatusInternalServerError)
		return
	}

	// Update block
	_, err = h.store.GetDB().Exec(
		"UPDATE block SET profile_id = ?, confidence = ? WHERE block_id = ?",
		req.ProfileID,
		req.Confidence,
		blockID,
	)

	if err != nil {
		log.Printf("Failed to reassign block: %v", err)
		respondError(w, "Failed to reassign block", http.StatusInternalServerError)
		return
	}

	// Write audit log
	var oldPID *int64
	if oldProfileID.Valid {
		pid := oldProfileID.Int64
		oldPID = &pid
	}

	auditDetails := map[string]interface{}{
		"block_id":       blockID,
		"old_profile_id": oldPID,
		"new_profile_id": req.ProfileID,
		"old_confidence": oldConfidence,
		"new_confidence": req.Confidence,
	}

	h.writeAuditLog("REASSIGN_BLOCK", auditDetails)

	// Create ML label event for training feedback loop
	// This allows the ML system to learn from user corrections
	if req.ProfileID != nil {
		_, labelErr := h.store.GetDB().Exec(`
			INSERT INTO ml_label_event (block_id, old_profile_id, new_profile_id, actor, confidence_after)
			VALUES (?, ?, ?, 'USER', ?)
		`, blockID, oldPID, *req.ProfileID, req.Confidence)

		if labelErr != nil {
			log.Printf("Warning: Failed to create ML label event: %v", labelErr)
		} else {
			log.Printf("[ML] Label event created: block %d assigned to profile %d (training data recorded)", blockID, *req.ProfileID)

			// Check if we have enough training data to auto-trigger training
			var labelCount int
			h.store.GetDB().QueryRow("SELECT COUNT(*) FROM ml_label_event WHERE new_profile_id IS NOT NULL").Scan(&labelCount)

			// Log training data status every 5 corrections
			if labelCount%5 == 0 {
				log.Printf("[ML] Training data: %d labeled samples. Need 10+ for training.", labelCount)
			}
		}
	}

	// Fetch updated block
	blocks := h.getBlocksByIDs([]int64{blockID})
	if len(blocks) == 0 {
		respondError(w, "Failed to fetch updated block", http.StatusInternalServerError)
		return
	}

	respondJSON(w, blocks[0], http.StatusOK)
}

// LockBlock handles POST /api/v1/blocks/{id}/lock
func (h *BlockHandler) LockBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract block_id from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		respondError(w, "Invalid path", http.StatusBadRequest)
		return
	}

	blockID, err := strconv.ParseInt(pathParts[3], 10, 64)
	if err != nil {
		respondError(w, "Invalid block_id", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		Locked bool `json:"locked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Update block
	result, err := h.store.GetDB().Exec(
		"UPDATE block SET locked = ? WHERE block_id = ?",
		req.Locked,
		blockID,
	)

	if err != nil {
		log.Printf("Failed to lock block: %v", err)
		respondError(w, "Failed to update block", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		respondError(w, "Block not found", http.StatusNotFound)
		return
	}

	// Write audit log
	action := "LOCK_BLOCK"
	if !req.Locked {
		action = "UNLOCK_BLOCK"
	}

	h.writeAuditLog(action, map[string]interface{}{
		"block_id": blockID,
		"locked":   req.Locked,
	})

	// Fetch updated block
	blocks := h.getBlocksByIDs([]int64{blockID})
	if len(blocks) == 0 {
		respondError(w, "Failed to fetch updated block", http.StatusInternalServerError)
		return
	}

	respondJSON(w, blocks[0], http.StatusOK)
}

// DeleteBlock deletes a block
func (h *BlockHandler) DeleteBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse block ID
	// Path is /api/v1/blocks/{id}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/blocks/")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid block ID", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteBlock(id); err != nil {
		log.Printf("API DeleteBlock failed: %v", err)
		http.Error(w, "Failed to delete block", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]bool{"success": true}, http.StatusOK)
}

// getBlocksByIDs fetches blocks by IDs (helper for returning updated blocks)
func (h *BlockHandler) getBlocksByIDs(blockIDs []int64) []BlockDTO {
	if len(blockIDs) == 0 {
		return []BlockDTO{}
	}

	placeholders := make([]string, len(blockIDs))
	args := make([]interface{}, len(blockIDs))
	for i, id := range blockIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `
		SELECT
			b.block_id, b.ts_start, b.ts_end, b.confidence, b.billable, b.locked,
			b.notes, b.description, b.created_at, b.updated_at, b.profile_id,
			da.app_name, dd.domain_text, dt.title_text,
			c.name, pr.name, s.name
		FROM block b
		JOIN dict_app da ON b.primary_app_id = da.app_id
		LEFT JOIN dict_domain dd ON b.primary_domain_id = dd.domain_id
		LEFT JOIN dict_title dt ON b.title_summary_id = dt.title_id
		LEFT JOIN profile p ON b.profile_id = p.profile_id
		LEFT JOIN client c ON p.client_id = c.client_id
		LEFT JOIN project pr ON p.project_id = pr.project_id
		LEFT JOIN service s ON p.service_id = s.service_id
		WHERE b.block_id IN (` + strings.Join(placeholders, ",") + `)
	`

	rows, err := h.store.GetDB().Query(query, args...)
	if err != nil {
		log.Printf("getBlocksByIDs query error: %v", err)
		return []BlockDTO{}
	}
	defer rows.Close()

	var blocks []BlockDTO

	for rows.Next() {
		var b BlockDTO
		var tsStart, tsEnd, createdAt, updatedAt string
		var profileID sql.NullInt64
		var primaryDomain, titleSummary, clientName, projectName, serviceName sql.NullString
		var notes, description sql.NullString

		err := rows.Scan(
			&b.BlockID,
			&tsStart,
			&tsEnd,
			&b.Confidence,
			&b.Billable,
			&b.Locked,
			&notes,
			&description,
			&createdAt,
			&updatedAt,
			&profileID,
			&b.PrimaryAppName,
			&primaryDomain,
			&titleSummary,
			&clientName,
			&projectName,
			&serviceName,
		)

		if err != nil {
			log.Printf("getBlocksByIDs scan error: %v", err)
			continue
		}

		// Parse timestamps and calculate duration
		start, _ := time.Parse(time.RFC3339, tsStart)
		end, _ := time.Parse(time.RFC3339, tsEnd)
		durationSeconds := end.Sub(start).Seconds()
		b.DurationMinutes = durationSeconds / 60.0
		b.DurationHours = durationSeconds / 3600.0

		// Set timestamps
		b.TsStart = tsStart
		b.TsEnd = tsEnd
		b.CreatedAt = createdAt
		b.UpdatedAt = updatedAt

		// Set nullable fields
		if profileID.Valid {
			pid := profileID.Int64
			b.ProfileID = &pid
		}
		if primaryDomain.Valid {
			b.PrimaryDomain = &primaryDomain.String
		}
		if titleSummary.Valid {
			b.TitleSummary = &titleSummary.String
		}
		if clientName.Valid {
			b.ClientName = &clientName.String
		}
		if projectName.Valid {
			b.ProjectName = &projectName.String
		}
		if serviceName.Valid {
			b.ServiceName = &serviceName.String
		}
		if notes.Valid {
			b.Notes = &notes.String
		}
		if description.Valid {
			b.Description = &description.String
		}

		blocks = append(blocks, b)
	}

	return blocks
}

// writeAuditLog writes an audit log entry
func (h *BlockHandler) writeAuditLog(action string, details interface{}) {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		log.Printf("Failed to marshal audit details: %v", err)
		return
	}

	_, err = h.store.GetDB().Exec(
		"INSERT INTO audit_log (actor, action, details_json) VALUES ('USER', ?, ?)",
		action,
		string(detailsJSON),
	)

	if err != nil {
		log.Printf("Failed to write audit log: %v", err)
	}
}

// ManualEntryRequest represents a request to create a manual time entry
type ManualEntryRequest struct {
	ProfileID   int64  `json:"profile_id"`
	TsStart     string `json:"ts_start"`    // ISO-8601 format
	TsEnd       string `json:"ts_end"`      // ISO-8601 format
	Title       string `json:"title"`       // e.g., "Phone call with Client ABC"
	Description string `json:"description"` // Optional detailed description
	Billable    bool   `json:"billable"`
}

// CreateManualEntry handles POST /api/v1/blocks/manual
func (h *BlockHandler) CreateManualEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ManualEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ProfileID == 0 {
		respondError(w, "profile_id is required", http.StatusBadRequest)
		return
	}
	if req.Title == "" {
		respondError(w, "title is required", http.StatusBadRequest)
		return
	}
	if req.TsStart == "" || req.TsEnd == "" {
		respondError(w, "ts_start and ts_end are required", http.StatusBadRequest)
		return
	}

	// Parse timestamps
	tsStart, err := time.Parse(time.RFC3339, req.TsStart)
	if err != nil {
		respondError(w, "Invalid ts_start format. Use ISO-8601 (e.g., 2026-01-13T09:00:00Z)", http.StatusBadRequest)
		return
	}
	tsEnd, err := time.Parse(time.RFC3339, req.TsEnd)
	if err != nil {
		respondError(w, "Invalid ts_end format. Use ISO-8601 (e.g., 2026-01-13T10:00:00Z)", http.StatusBadRequest)
		return
	}

	// Validate time range
	if !tsEnd.After(tsStart) {
		respondError(w, "ts_end must be after ts_start", http.StatusBadRequest)
		return
	}

	// Don't allow times more than 1 hour in the future
	if tsStart.After(time.Now().Add(1 * time.Hour)) {
		respondError(w, "ts_start cannot be more than 1 hour in the future", http.StatusBadRequest)
		return
	}

	// Verify profile exists
	var profileExists int
	err = h.store.GetDB().QueryRow("SELECT 1 FROM profile WHERE profile_id = ?", req.ProfileID).Scan(&profileExists)
	if err != nil {
		respondError(w, "Profile not found", http.StatusBadRequest)
		return
	}

	// Get or create a "Manual Entry" app in dict_app
	var manualAppID int64
	err = h.store.GetDB().QueryRow("SELECT app_id FROM dict_app WHERE app_name = 'Manual Entry'").Scan(&manualAppID)
	if err == sql.ErrNoRows {
		result, err := h.store.GetDB().Exec("INSERT INTO dict_app (app_name) VALUES ('Manual Entry')")
		if err != nil {
			log.Printf("Failed to create Manual Entry app: %v", err)
			respondError(w, "Failed to create manual entry", http.StatusInternalServerError)
			return
		}
		manualAppID, _ = result.LastInsertId()
	} else if err != nil {
		log.Printf("Failed to lookup Manual Entry app: %v", err)
		respondError(w, "Failed to create manual entry", http.StatusInternalServerError)
		return
	}

	// Get or create title in dict_title
	var titleID int64
	err = h.store.GetDB().QueryRow("SELECT title_id FROM dict_title WHERE title_text = ?", req.Title).Scan(&titleID)
	if err == sql.ErrNoRows {
		result, err := h.store.GetDB().Exec("INSERT INTO dict_title (title_text) VALUES (?)", req.Title)
		if err != nil {
			log.Printf("Failed to create title: %v", err)
			respondError(w, "Failed to create manual entry", http.StatusInternalServerError)
			return
		}
		titleID, _ = result.LastInsertId()
	} else if err != nil {
		log.Printf("Failed to lookup title: %v", err)
		respondError(w, "Failed to create manual entry", http.StatusInternalServerError)
		return
	}

	// Build description
	description := req.Description
	if description == "" {
		description = req.Title
	}

	// Insert manual block
	result, err := h.store.GetDB().Exec(`
		INSERT INTO block (
			ts_start, ts_end, primary_app_id, title_summary_id, profile_id,
			confidence, billable, locked, description, is_manual, manual_title
		) VALUES (?, ?, ?, ?, ?, 'HIGH', ?, 0, ?, 1, ?)
	`,
		tsStart.Format(time.RFC3339),
		tsEnd.Format(time.RFC3339),
		manualAppID,
		titleID,
		req.ProfileID,
		req.Billable,
		description,
		req.Title,
	)

	if err != nil {
		log.Printf("Failed to insert manual entry: %v", err)
		respondError(w, "Failed to create manual entry", http.StatusInternalServerError)
		return
	}

	blockID, _ := result.LastInsertId()

	// Write audit log
	h.writeAuditLog("CREATE_MANUAL_ENTRY", map[string]interface{}{
		"block_id":   blockID,
		"profile_id": req.ProfileID,
		"title":      req.Title,
		"ts_start":   req.TsStart,
		"ts_end":     req.TsEnd,
		"billable":   req.Billable,
	})

	log.Printf("Created manual entry: block_id=%d, profile=%d, title=%s", blockID, req.ProfileID, req.Title)

	// Return created block
	blocks := h.getBlocksByIDs([]int64{blockID})
	if len(blocks) > 0 {
		respondJSON(w, blocks[0], http.StatusCreated)
	} else {
		respondJSON(w, map[string]interface{}{"block_id": blockID, "success": true}, http.StatusCreated)
	}
}
