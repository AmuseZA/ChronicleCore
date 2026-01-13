package engine

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"

	"chroniclecore/internal/store"
)

// RuleEngine handles automatic profile assignment based on rules
type RuleEngine struct {
	store *store.Store
	cache *ruleCache
}

type ruleCache struct {
	rules       []*Rule
	appNameMap  map[string]int64 // app_name -> app_id
	titleCache  map[int64]string // title_id -> title_text
	domainCache map[int64]string // domain_id -> domain_text
}

// Rule represents an assignment rule
type Rule struct {
	RuleID           int64
	Name             string
	Priority         int
	MatchType        string // APP, DOMAIN, TITLE_REGEX, KEYWORD, COMPOSITE
	MatchValue       string
	TargetProfileID  int64
	TargetServiceID  *int64
	ConfidenceBoost  int
	Enabled          bool
	compiledRegex    *regexp.Regexp // For TITLE_REGEX match type
}

// NewRuleEngine creates a new rule engine
func NewRuleEngine(store *store.Store) *RuleEngine {
	return &RuleEngine{
		store: store,
		cache: &ruleCache{
			appNameMap:  make(map[string]int64),
			titleCache:  make(map[int64]string),
			domainCache: make(map[int64]string),
		},
	}
}

// LoadRules loads and caches all active rules from database
func (re *RuleEngine) LoadRules() error {
	query := `
		SELECT rule_id, name, priority, match_type, match_value,
		       target_profile_id, target_service_id, confidence_boost, enabled
		FROM rule
		WHERE enabled = 1
		ORDER BY priority DESC, rule_id ASC
	`

	rows, err := re.store.GetDB().Query(query)
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}
	defer rows.Close()

	var rules []*Rule
	for rows.Next() {
		var r Rule
		var targetServiceID sql.NullInt64

		err := rows.Scan(
			&r.RuleID,
			&r.Name,
			&r.Priority,
			&r.MatchType,
			&r.MatchValue,
			&r.TargetProfileID,
			&targetServiceID,
			&r.ConfidenceBoost,
			&r.Enabled,
		)
		if err != nil {
			return fmt.Errorf("failed to scan rule: %w", err)
		}

		if targetServiceID.Valid {
			sid := targetServiceID.Int64
			r.TargetServiceID = &sid
		}

		// Compile regex for TITLE_REGEX match type
		if r.MatchType == "TITLE_REGEX" {
			compiled, err := regexp.Compile(r.MatchValue)
			if err != nil {
				log.Printf("Warning: Invalid regex in rule %d (%s): %v", r.RuleID, r.Name, err)
				continue // Skip invalid regex rules
			}
			r.compiledRegex = compiled
		}

		rules = append(rules, &r)
	}

	re.cache.rules = rules
	log.Printf("Loaded %d active rules", len(rules))

	return nil
}

// LoadDictionaries loads app/title/domain dictionaries into cache
func (re *RuleEngine) LoadDictionaries() error {
	// Load app names
	rows, err := re.store.GetDB().Query("SELECT app_id, app_name FROM dict_app")
	if err != nil {
		return fmt.Errorf("failed to load app dictionary: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var appID int64
		var appName string
		if err := rows.Scan(&appID, &appName); err != nil {
			return err
		}
		re.cache.appNameMap[appName] = appID
	}

	// Load titles (for regex matching)
	rows, err = re.store.GetDB().Query("SELECT title_id, title_text FROM dict_title")
	if err != nil {
		return fmt.Errorf("failed to load title dictionary: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var titleID int64
		var titleText string
		if err := rows.Scan(&titleID, &titleText); err != nil {
			return err
		}
		re.cache.titleCache[titleID] = titleText
	}

	log.Printf("Loaded dictionaries: %d apps, %d titles", len(re.cache.appNameMap), len(re.cache.titleCache))

	return nil
}

// AssignProfile matches a block against rules and returns profile ID + confidence
func (re *RuleEngine) AssignProfile(block *store.Block) (profileID *int64, confidence string) {
	// Get app name from dictionary
	var appName string
	err := re.store.GetDB().QueryRow(
		"SELECT app_name FROM dict_app WHERE app_id = ?",
		block.PrimaryAppID,
	).Scan(&appName)

	if err != nil {
		log.Printf("Failed to get app name for block %d: %v", block.BlockID, err)
		return nil, "LOW"
	}

	// Get title text if present
	var titleText string
	if block.TitleSummaryID != nil {
		titleText = re.cache.titleCache[*block.TitleSummaryID]
		if titleText == "" {
			// Not in cache, fetch from DB
			re.store.GetDB().QueryRow(
				"SELECT title_text FROM dict_title WHERE title_id = ?",
				*block.TitleSummaryID,
			).Scan(&titleText)
		}
	}

	// Match against rules (ordered by priority DESC)
	for _, rule := range re.cache.rules {
		matched := false

		switch rule.MatchType {
		case "APP":
			// Exact app name match
			matched = (appName == rule.MatchValue)

		case "TITLE_REGEX":
			// Regex match on title
			if rule.compiledRegex != nil && titleText != "" {
				matched = rule.compiledRegex.MatchString(titleText)
			}

		case "KEYWORD":
			// Simple substring match (case-insensitive) on title
			if titleText != "" {
				// Simple implementation for MVP
				matched = contains(titleText, rule.MatchValue)
			}

		case "DOMAIN":
			// Domain match (requires browser extension - Phase 2)
			// For MVP, skip domain rules
			continue

		case "COMPOSITE":
			// Complex rules (future enhancement)
			continue
		}

		if matched {
			pid := rule.TargetProfileID
			conf := "HIGH" // Automatic assignment with rule match

			// Apply confidence boost if specified
			if rule.ConfidenceBoost < 0 {
				conf = "MEDIUM" // Lower confidence
			}

			return &pid, conf
		}
	}

	// No match - return unassigned with LOW confidence
	return nil, "LOW"
}

// AssignBlocksInRange applies rules to all unassigned blocks in a time range
func (re *RuleEngine) AssignBlocksInRange() error {
	// Reload rules to catch any changes
	if err := re.LoadRules(); err != nil {
		return err
	}

	// Reload dictionaries
	if err := re.LoadDictionaries(); err != nil {
		return err
	}

	// Get all unassigned or LOW confidence blocks
	query := `
		SELECT block_id, ts_start, ts_end, primary_app_id, primary_domain_id,
		       title_summary_id, profile_id, confidence, billable, locked
		FROM block
		WHERE (profile_id IS NULL OR confidence = 'LOW')
		  AND locked = 0
		ORDER BY ts_start DESC
		LIMIT 1000
	`

	rows, err := re.store.GetDB().Query(query)
	if err != nil {
		return fmt.Errorf("failed to query blocks: %w", err)
	}
	defer rows.Close()

	var blocks []*store.Block
	for rows.Next() {
		var b store.Block
		var profileID sql.NullInt64
		var domainID, titleID sql.NullInt64

		err := rows.Scan(
			&b.BlockID,
			&b.TsStart,
			&b.TsEnd,
			&b.PrimaryAppID,
			&domainID,
			&titleID,
			&profileID,
			&b.Confidence,
			&b.Billable,
			&b.Locked,
		)
		if err != nil {
			return fmt.Errorf("failed to scan block: %w", err)
		}

		if domainID.Valid {
			did := domainID.Int64
			b.PrimaryDomainID = &did
		}
		if titleID.Valid {
			tid := titleID.Int64
			b.TitleSummaryID = &tid
		}
		if profileID.Valid {
			pid := profileID.Int64
			b.ProfileID = &pid
		}

		blocks = append(blocks, &b)
	}

	if len(blocks) == 0 {
		log.Println("No blocks to assign")
		return nil
	}

	log.Printf("Assigning profiles to %d blocks...", len(blocks))

	// Process each block
	assigned := 0
	for _, block := range blocks {
		profileID, confidence := re.AssignProfile(block)

		// Update block if assignment changed
		if profileID != nil || confidence != block.Confidence {
			_, err := re.store.GetDB().Exec(
				"UPDATE block SET profile_id = ?, confidence = ? WHERE block_id = ?",
				profileID,
				confidence,
				block.BlockID,
			)
			if err != nil {
				log.Printf("Failed to update block %d: %v", block.BlockID, err)
				continue
			}

			if profileID != nil {
				assigned++
			}
		}
	}

	log.Printf("Assigned %d blocks to profiles", assigned)
	return nil
}

// Helper function for case-insensitive substring matching
func contains(text, substr string) bool {
	// Simple case-insensitive contains for MVP
	// Could be improved with proper Unicode normalization
	return regexp.MustCompile(`(?i)` + regexp.QuoteMeta(substr)).MatchString(text)
}
