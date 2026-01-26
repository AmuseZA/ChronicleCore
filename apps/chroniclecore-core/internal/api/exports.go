package api

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"chroniclecore/internal/store"
)

// ExportHandler manages export endpoints
type ExportHandler struct {
	store *store.Store
}

func NewExportHandler(store *store.Store) *ExportHandler {
	return &ExportHandler{store: store}
}

// ExportRequest represents the invoice lines export request
type ExportRequest struct {
	StartDate              string  `json:"start_date"`               // YYYY-MM-DD
	EndDate                string  `json:"end_date"`                 // YYYY-MM-DD
	ProfileIDs             []int64 `json:"profile_ids,omitempty"`    // Optional filter
	RoundingMinutes        int     `json:"rounding_minutes"`         // 6 or 15
	MinimumBillableMinutes int     `json:"minimum_billable_minutes"` // Default: 0
}

// InvoiceLine represents a single invoice line item
type InvoiceLine struct {
	Client      string
	Project     string
	Service     string
	Date        string
	StartTime   string
	EndTime     string
	Duration    float64 // Hours
	RawDuration float64 // Hours (before rounding)
	Rate        float64
	Currency    string
	Amount      float64
	Description string
	Confidence  string
}

// ExportInvoiceLines handles POST /api/v1/export/invoice-lines
func (h *ExportHandler) ExportInvoiceLines(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate and set defaults
	if req.RoundingMinutes == 0 {
		req.RoundingMinutes = 6 // Default 6-minute rounding
	}
	if req.RoundingMinutes != 6 && req.RoundingMinutes != 15 {
		respondError(w, "rounding_minutes must be 6 or 15", http.StatusBadRequest)
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		respondError(w, "Invalid start_date format (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		respondError(w, "Invalid end_date format (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	// Query blocks in date range
	lines, err := h.queryInvoiceLines(startDate, endDate, req.ProfileIDs)
	if err != nil {
		log.Printf("Failed to query invoice lines: %v", err)
		respondError(w, "Failed to query blocks", http.StatusInternalServerError)
		return
	}

	// Apply rounding and minimum billing
	for i := range lines {
		lines[i].RawDuration = lines[i].Duration

		// Apply rounding
		durationMinutes := lines[i].Duration * 60
		roundedMinutes := roundDuration(durationMinutes, float64(req.RoundingMinutes))

		// Apply minimum billing
		if req.MinimumBillableMinutes > 0 && roundedMinutes < float64(req.MinimumBillableMinutes) {
			roundedMinutes = float64(req.MinimumBillableMinutes)
		}

		lines[i].Duration = roundedMinutes / 60.0 // Convert back to hours
		lines[i].Amount = lines[i].Duration * lines[i].Rate
	}

	// Generate CSV
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="invoice_lines.csv"`)

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	writer.Write([]string{
		"Client",
		"Project",
		"Service",
		"Date",
		"Start Time",
		"End Time",
		"Hours (Rounded)",
		"Hours (Actual)",
		"Rate",
		"Currency",
		"Amount",
		"Description",
		"Confidence",
	})

	// Write rows
	for _, line := range lines {
		writer.Write([]string{
			line.Client,
			line.Project,
			line.Service,
			line.Date,
			line.StartTime,
			line.EndTime,
			fmt.Sprintf("%.2f", line.Duration),
			fmt.Sprintf("%.2f", line.RawDuration),
			fmt.Sprintf("%.2f", line.Rate),
			line.Currency,
			fmt.Sprintf("%.2f", line.Amount),
			line.Description,
			line.Confidence,
		})
	}
}

// queryInvoiceLines retrieves blocks from database
// Uses activity-weighted billing: duration is multiplied by activity_score
func (h *ExportHandler) queryInvoiceLines(startDate, endDate time.Time, profileIDs []int64) ([]InvoiceLine, error) {
	query := `
		SELECT
			b.ts_start,
			b.ts_end,
			b.description,
			b.confidence,
			c.name as client_name,
			COALESCE(pr.name, '') as project_name,
			s.name as service_name,
			r.hourly_minor_units,
			r.currency_code,
			COALESCE(b.activity_score, 1.0) as activity_score
		FROM block b
		JOIN profile p ON b.profile_id = p.profile_id
		JOIN client c ON p.client_id = c.client_id
		LEFT JOIN project pr ON p.project_id = pr.project_id
		JOIN service s ON p.service_id = s.service_id
		JOIN rate r ON p.rate_id = r.rate_id
		WHERE b.billable = 1
		  AND DATE(b.ts_start) >= ?
		  AND DATE(b.ts_start) <= ?
	`

	args := []interface{}{
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	}

	// Add profile filter if specified
	if len(profileIDs) > 0 {
		query += " AND b.profile_id IN ("
		for i := range profileIDs {
			if i > 0 {
				query += ","
			}
			query += "?"
			args = append(args, profileIDs[i])
		}
		query += ")"
	}

	query += " ORDER BY b.ts_start ASC"

	rows, err := h.store.GetDB().Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []InvoiceLine
	for rows.Next() {
		var tsStart, tsEnd string
		var description sql.NullString
		var confidence, client, project, service, currency string
		var minorUnits int64
		var activityScore float64

		err := rows.Scan(
			&tsStart,
			&tsEnd,
			&description,
			&confidence,
			&client,
			&project,
			&service,
			&minorUnits,
			&currency,
			&activityScore,
		)
		if err != nil {
			return nil, err
		}

		// Parse timestamps
		start, _ := time.Parse(time.RFC3339, tsStart)
		end, _ := time.Parse(time.RFC3339, tsEnd)

		// Calculate duration in hours with activity weighting
		// This ensures only actual active work time is billed
		durationSeconds := end.Sub(start).Seconds()
		durationHours := (durationSeconds / 3600.0) * activityScore

		// Convert rate to major units
		rate := float64(minorUnits) / 100.0

		line := InvoiceLine{
			Client:      client,
			Project:     project,
			Service:     service,
			Date:        start.Format("2006-01-02"),
			StartTime:   start.Format("15:04"),
			EndTime:     end.Format("15:04"),
			Duration:    durationHours,
			RawDuration: durationHours,
			Rate:        rate,
			Currency:    currency,
			Amount:      0, // Will be calculated after rounding
			Description: description.String,
			Confidence:  confidence,
		}

		lines = append(lines, line)
	}

	return lines, nil
}

// roundDuration rounds duration to nearest increment
func roundDuration(minutes, increment float64) float64 {
	if increment <= 0 {
		return minutes
	}
	return math.Ceil(minutes/float64(increment)) * float64(increment)
}
