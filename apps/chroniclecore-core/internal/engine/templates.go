package engine

import (
	"fmt"
	"strings"

	"chroniclecore/internal/store"
)

// TemplateEngine generates descriptions for blocks
type TemplateEngine struct {
	store *store.Store
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine(store *store.Store) *TemplateEngine {
	return &TemplateEngine{store: store}
}

// GenerateDescription creates a human-readable description for a block
func (te *TemplateEngine) GenerateDescription(block *store.Block) string {
	// Get app name
	var appName string
	err := te.store.GetDB().QueryRow(
		"SELECT app_name FROM dict_app WHERE app_id = ?",
		block.PrimaryAppID,
	).Scan(&appName)

	if err != nil {
		appName = "Unknown App"
	}

	// Clean up app name (remove .exe extension)
	appName = strings.TrimSuffix(appName, ".exe")
	appName = strings.TrimSuffix(appName, ".EXE")

	// Get title if available
	var titleText string
	if block.TitleSummaryID != nil {
		te.store.GetDB().QueryRow(
			"SELECT title_text FROM dict_title WHERE title_id = ?",
			*block.TitleSummaryID,
		).Scan(&titleText)
	}

	// Get domain if available (for browser context)
	var domainText string
	if block.PrimaryDomainID != nil {
		te.store.GetDB().QueryRow(
			"SELECT domain_text FROM dict_domain WHERE domain_id = ?",
			*block.PrimaryDomainID,
		).Scan(&domainText)
	}

	// Build description using template
	var description string

	if domainText != "" {
		// Browser with domain: "browsing gmail.com"
		description = fmt.Sprintf("browsing %s", domainText)
	} else if titleText != "" {
		// App with title: "Excel - Budget 2026.xlsx"
		description = fmt.Sprintf("%s - %s", appName, titleText)
	} else {
		// Just app name: "Excel"
		description = appName
	}

	// Capitalize first letter
	if len(description) > 0 {
		description = strings.ToUpper(description[:1]) + description[1:]
	}

	return description
}

// GenerateDescriptionsForBlocks applies descriptions to blocks without them
func (te *TemplateEngine) GenerateDescriptionsForBlocks() error {
	// Get blocks without descriptions
	query := `
		SELECT block_id, ts_start, ts_end, primary_app_id, primary_domain_id,
		       title_summary_id, profile_id, confidence, billable, locked, description
		FROM block
		WHERE description IS NULL OR description = ''
		ORDER BY ts_start DESC
		LIMIT 1000
	`

	rows, err := te.store.GetDB().Query(query)
	if err != nil {
		return fmt.Errorf("failed to query blocks: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var block store.Block
		var titleID, domainID, profileID *int64

		err := rows.Scan(
			&block.BlockID,
			&block.TsStart,
			&block.TsEnd,
			&block.PrimaryAppID,
			&domainID,
			&titleID,
			&profileID,
			&block.Confidence,
			&block.Billable,
			&block.Locked,
			&block.Description,
		)
		if err != nil {
			continue
		}

		if titleID != nil {
			block.TitleSummaryID = titleID
		}
		if domainID != nil {
			block.PrimaryDomainID = domainID
		}
		if profileID != nil {
			block.ProfileID = profileID
		}

		// Generate description
		desc := te.GenerateDescription(&block)

		// Update block
		_, err = te.store.GetDB().Exec(
			"UPDATE block SET description = ? WHERE block_id = ?",
			desc,
			block.BlockID,
		)
		if err == nil {
			count++
		}
	}

	if count > 0 {
		fmt.Printf("Generated descriptions for %d blocks\n", count)
	}

	return nil
}
