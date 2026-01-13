package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"chroniclecore/internal/ml"
)

// MLHandler handles ML-related API endpoints
type MLHandler struct {
	db            *sql.DB
	sidecar       *ml.SidecarManager
	sidecarClient *ml.Client
}

// NewMLHandler creates a new ML handler
func NewMLHandler(db *sql.DB, sidecar *ml.SidecarManager) *MLHandler {
	return &MLHandler{
		db:            db,
		sidecar:       sidecar,
		sidecarClient: ml.NewClient(sidecar.GetPort(), sidecar.GetToken()),
	}
}

// GetTrainingData exports training data (features + labels) from database
func (h *MLHandler) GetTrainingData(w http.ResponseWriter, r *http.Request) {
	log.Println("Fetching training data from ml_label_event...")

	// Query label events (user corrections)
	query := `
		SELECT
			le.block_id,
			le.new_profile_id,
			b.ts_start,
			b.ts_end,
			da.app_name,
			dt.title_text,
			dd.domain_text
		FROM ml_label_event le
		JOIN block b ON le.block_id = b.block_id
		JOIN dict_app da ON b.primary_app_id = da.app_id
		LEFT JOIN dict_title dt ON b.title_summary_id = dt.title_id
		LEFT JOIN dict_domain dd ON b.primary_domain_id = dd.domain_id
		WHERE le.new_profile_id IS NOT NULL
		ORDER BY le.ts DESC
		LIMIT 1000
	`

	rows, err := h.db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database query failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type TrainingData struct {
		Features []map[string]interface{} `json:"features"`
		Labels   []int                    `json:"labels"`
		Count    int                      `json:"count"`
	}

	var features []map[string]interface{}
	var labels []int

	for rows.Next() {
		var blockID, profileID int
		var tsStart, tsEnd string
		var appName string
		var title, domain sql.NullString

		if err := rows.Scan(&blockID, &profileID, &tsStart, &tsEnd, &appName, &title, &domain); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Build feature dict
		feature := map[string]interface{}{
			"block_id": blockID,
			"app_name": appName,
			"ts_start": tsStart,
			"ts_end":   tsEnd,
		}

		if title.Valid {
			feature["title"] = title.String
		}

		if domain.Valid {
			feature["domain"] = domain.String
		}

		features = append(features, feature)
		labels = append(labels, profileID)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Row iteration error: %v", err), http.StatusInternalServerError)
		return
	}

	data := TrainingData{
		Features: features,
		Labels:   labels,
		Count:    len(features),
	}

	log.Printf("Exported %d training samples", data.Count)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// TriggerTraining initiates model training via sidecar
func (h *MLHandler) TriggerTraining(w http.ResponseWriter, r *http.Request) {
	log.Println("Triggering ML training...")

	// Ensure sidecar is running
	if !h.sidecar.IsRunning() {
		log.Println("ML sidecar not running, attempting to start...")
		if err := h.sidecar.Restart(); err != nil {
			log.Printf("Failed to start ML sidecar: %v", err)
			http.Error(w, fmt.Sprintf("ML service unavailable: %v", err), http.StatusServiceUnavailable)
			return
		}
		log.Println("ML sidecar started successfully")
	}

	// Get training data
	query := `
		SELECT
			le.block_id,
			le.new_profile_id,
			da.app_name,
			dt.title_text,
			dd.domain_text
		FROM ml_label_event le
		JOIN block b ON le.block_id = b.block_id
		JOIN dict_app da ON b.primary_app_id = da.app_id
		LEFT JOIN dict_title dt ON b.title_summary_id = dt.title_id
		LEFT JOIN dict_domain dd ON b.primary_domain_id = dd.domain_id
		WHERE le.new_profile_id IS NOT NULL
		ORDER BY le.ts DESC
		LIMIT 1000
	`

	rows, err := h.db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database query failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var features []map[string]interface{}
	var labels []int

	for rows.Next() {
		var blockID, profileID int
		var appName string
		var title, domain sql.NullString

		if err := rows.Scan(&blockID, &profileID, &appName, &title, &domain); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		feature := map[string]interface{}{
			"block_id": blockID,
			"app_name": appName,
		}

		if title.Valid {
			feature["title"] = title.String
		}

		if domain.Valid {
			feature["domain"] = domain.String
		}

		features = append(features, feature)
		labels = append(labels, profileID)
	}

	if len(features) < 10 {
		http.Error(w, fmt.Sprintf("Insufficient training data: need at least 10 samples, have %d", len(features)), http.StatusBadRequest)
		return
	}

	// Send training request to sidecar
	startTime := time.Now()
	trainReq := ml.TrainRequest{
		Features:  features,
		Labels:    labels,
		ModelType: "PROFILE_CLASSIFIER",
	}

	trainResp, err := h.sidecarClient.Train(trainReq)
	if err != nil {
		// Log to ml_run_log
		h.logMLRun("TRAIN", 0, false, err.Error(), time.Since(startTime), len(features), 0)
		http.Error(w, fmt.Sprintf("Training failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Persist model to registry
	modelID, err := h.persistModel(trainResp)
	if err != nil {
		log.Printf("Failed to persist model: %v", err)
		// Not a fatal error, training still succeeded
	}

	// Log successful run
	h.logMLRun("TRAIN", modelID, true, "", time.Since(startTime), len(features), 0)

	log.Printf("Training complete: %s (accuracy: %.3f)", trainResp.ModelVersion, trainResp.Metrics["accuracy"])

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trainResp)
}

// PredictBlocks generates profile predictions for unassigned blocks
func (h *MLHandler) PredictBlocks(w http.ResponseWriter, r *http.Request) {
	log.Println("Generating predictions for unassigned blocks...")

	// Get unassigned blocks
	query := `
		SELECT
			b.block_id,
			b.ts_start,
			b.ts_end,
			da.app_name,
			dt.title_text,
			dd.domain_text
		FROM block b
		JOIN dict_app da ON b.primary_app_id = da.app_id
		LEFT JOIN dict_title dt ON b.title_summary_id = dt.title_id
		LEFT JOIN dict_domain dd ON b.primary_domain_id = dd.domain_id
		WHERE b.profile_id IS NULL
		  AND b.confidence = 'LOW'
		ORDER BY b.ts_start DESC
		LIMIT 100
	`

	rows, err := h.db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database query failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var features []map[string]interface{}
	var blockIDs []int

	for rows.Next() {
		var blockID int
		var tsStart, tsEnd, appName string
		var title, domain sql.NullString

		if err := rows.Scan(&blockID, &tsStart, &tsEnd, &appName, &title, &domain); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		feature := map[string]interface{}{
			"block_id": blockID,
			"app_name": appName,
			"ts_start": tsStart,
			"ts_end":   tsEnd,
		}

		if title.Valid {
			feature["title"] = title.String
		}

		if domain.Valid {
			feature["domain"] = domain.String
		}

		features = append(features, feature)
		blockIDs = append(blockIDs, blockID)
	}

	if len(features) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			"predictions": 0,
			"message":     "No unassigned blocks to predict",
		})
		return
	}

	// Send prediction request to sidecar
	startTime := time.Now()
	predictReq := ml.PredictRequest{
		Features:  features,
		Threshold: 0.6, // MEDIUM confidence minimum
	}

	predictResp, err := h.sidecarClient.Predict(predictReq)
	if err != nil {
		h.logMLRun("PREDICT", 0, false, err.Error(), time.Since(startTime), len(features), 0)
		http.Error(w, fmt.Sprintf("Prediction failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Persist suggestions to database
	suggestionsCreated := 0
	for _, pred := range predictResp.Predictions {
		blockID := blockIDs[pred.BlockIndex]

		payloadJSON, _ := json.Marshal(map[string]interface{}{
			"predicted_profile_id": pred.PredictedProfileID,
			"confidence_level":     pred.ConfidenceLevel,
		})

		_, err := h.db.Exec(`
			INSERT INTO ml_suggestion (entity_type, entity_id, suggestion_type, payload_json, confidence, status)
			VALUES (?, ?, ?, ?, ?, ?)
		`, "BLOCK", blockID, "PROFILE_ASSIGN", string(payloadJSON), pred.Confidence, "PENDING")

		if err != nil {
			log.Printf("Failed to persist suggestion for block %d: %v", blockID, err)
			continue
		}

		suggestionsCreated++
	}

	h.logMLRun("PREDICT", 0, true, "", time.Since(startTime), len(features), suggestionsCreated)

	log.Printf("Created %d suggestions from %d predictions", suggestionsCreated, len(predictResp.Predictions))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":            true,
		"predictions":        len(predictResp.Predictions),
		"suggestions_created": suggestionsCreated,
		"model_version":      predictResp.ModelVersion,
	})
}

// GetSuggestions retrieves pending ML suggestions
func (h *MLHandler) GetSuggestions(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT
			s.suggestion_id,
			s.entity_type,
			s.entity_id,
			s.suggestion_type,
			s.payload_json,
			s.confidence,
			s.status,
			s.created_at
		FROM ml_suggestion s
		WHERE s.status = 'PENDING'
		ORDER BY s.confidence DESC, s.created_at DESC
		LIMIT 50
	`

	rows, err := h.db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database query failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Suggestion struct {
		SuggestionID   int                    `json:"suggestion_id"`
		EntityType     string                 `json:"entity_type"`
		EntityID       int                    `json:"entity_id"`
		SuggestionType string                 `json:"suggestion_type"`
		Payload        map[string]interface{} `json:"payload"`
		Confidence     float64                `json:"confidence"`
		Status         string                 `json:"status"`
		CreatedAt      string                 `json:"created_at"`
	}

	var suggestions []Suggestion

	for rows.Next() {
		var s Suggestion
		var payloadJSON string

		if err := rows.Scan(&s.SuggestionID, &s.EntityType, &s.EntityID, &s.SuggestionType, &payloadJSON, &s.Confidence, &s.Status, &s.CreatedAt); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Parse payload JSON
		json.Unmarshal([]byte(payloadJSON), &s.Payload)

		suggestions = append(suggestions, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}

// AcceptSuggestion accepts an ML suggestion and applies it
func (h *MLHandler) AcceptSuggestion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SuggestionID int `json:"suggestion_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get suggestion
	var entityType, suggestionType, payloadJSON string
	var entityID int
	var confidence float64

	err := h.db.QueryRow(`
		SELECT entity_type, entity_id, suggestion_type, payload_json, confidence
		FROM ml_suggestion
		WHERE suggestion_id = ? AND status = 'PENDING'
	`, req.SuggestionID).Scan(&entityType, &entityID, &suggestionType, &payloadJSON, &confidence)

	if err == sql.ErrNoRows {
		http.Error(w, "Suggestion not found or already processed", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse payload
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse suggestion payload: %v", err), http.StatusInternalServerError)
		return
	}

	// Apply suggestion based on type
	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to begin transaction: %v", err), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	if suggestionType == "PROFILE_ASSIGN" && entityType == "BLOCK" {
		// Safely extract profile ID with type checking
		predictedID, ok := payload["predicted_profile_id"].(float64)
		if !ok {
			http.Error(w, "Invalid suggestion payload: missing predicted_profile_id", http.StatusInternalServerError)
			return
		}
		profileID := int(predictedID)

		// Safely extract confidence level with fallback
		confidenceLevel, ok := payload["confidence_level"].(string)
		if !ok {
			confidenceLevel = "MEDIUM" // Fallback
		}

		// Update block
		_, err = tx.Exec(`
			UPDATE block
			SET profile_id = ?, confidence = ?, updated_at = datetime('now')
			WHERE block_id = ?
		`, profileID, confidenceLevel, entityID)

		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update block: %v", err), http.StatusInternalServerError)
			return
		}

		// Create label event for feedback loop
		_, err = tx.Exec(`
			INSERT INTO ml_label_event (block_id, old_profile_id, new_profile_id, actor, confidence_after)
			VALUES (?, NULL, ?, 'SYSTEM', ?)
		`, entityID, profileID, confidenceLevel)

		if err != nil {
			log.Printf("Failed to create label event: %v", err)
		}
	}

	// Mark suggestion as accepted
	_, err = tx.Exec(`
		UPDATE ml_suggestion
		SET status = 'ACCEPTED', resolved_at = datetime('now')
		WHERE suggestion_id = ?
	`, req.SuggestionID)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update suggestion: %v", err), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to commit transaction: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Accepted suggestion %d: assigned block %d to profile %v", req.SuggestionID, entityID, payload["predicted_profile_id"])

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Suggestion accepted and applied",
	})
}

// GetMLStatus returns the current ML system status
func (h *MLHandler) GetMLStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get label count
	var labelCount int
	h.db.QueryRow("SELECT COUNT(*) FROM ml_label_event WHERE new_profile_id IS NOT NULL").Scan(&labelCount)

	// Get latest model
	var latestModel struct {
		Version   string
		Algorithm string
		Accuracy  float64
		TrainedAt string
		Samples   int
	}

	err := h.db.QueryRow(`
		SELECT version, algorithm,
		       COALESCE(json_extract(metrics_json, '$.accuracy'), 0) as accuracy,
		       created_at, trained_samples
		FROM ml_model_registry
		WHERE status = 'ACTIVE'
		ORDER BY created_at DESC
		LIMIT 1
	`).Scan(&latestModel.Version, &latestModel.Algorithm, &latestModel.Accuracy, &latestModel.TrainedAt, &latestModel.Samples)

	hasModel := err == nil

	// Get pending suggestions count
	var pendingSuggestions int
	h.db.QueryRow("SELECT COUNT(*) FROM ml_suggestion WHERE status = 'PENDING'").Scan(&pendingSuggestions)

	// Get recent runs
	var recentRuns []map[string]interface{}
	rows, _ := h.db.Query(`
		SELECT run_type, success, input_samples, output_count, duration_ms, ts
		FROM ml_run_log
		ORDER BY ts DESC
		LIMIT 5
	`)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var runType, ts string
			var success, inputSamples, outputCount int
			var durationMs int64
			if rows.Scan(&runType, &success, &inputSamples, &outputCount, &durationMs, &ts) == nil {
				recentRuns = append(recentRuns, map[string]interface{}{
					"run_type":      runType,
					"success":       success == 1,
					"input_samples": inputSamples,
					"output_count":  outputCount,
					"duration_ms":   durationMs,
					"timestamp":     ts,
				})
			}
		}
	}

	sidecarReady := false
	if h.sidecar != nil {
		sidecarReady = h.sidecar.IsRunning()
	}

	status := map[string]interface{}{
		"sidecar_running":     sidecarReady,
		"training_samples":    labelCount,
		"ready_for_training":  labelCount >= 10,
		"has_trained_model":   hasModel,
		"pending_suggestions": pendingSuggestions,
		"recent_runs":         recentRuns,
	}

	if hasModel {
		status["latest_model"] = map[string]interface{}{
			"version":    latestModel.Version,
			"algorithm":  latestModel.Algorithm,
			"accuracy":   latestModel.Accuracy,
			"trained_at": latestModel.TrainedAt,
			"samples":    latestModel.Samples,
		}
	}

	log.Printf("[ML Status] Samples: %d, Model: %v, Sidecar: %v", labelCount, hasModel, sidecarReady)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// persistModel saves model metadata to ml_model_registry
func (h *MLHandler) persistModel(resp *ml.TrainResponse) (int, error) {
	metricsJSON, _ := json.Marshal(resp.Metrics)

	result, err := h.db.Exec(`
		INSERT INTO ml_model_registry (model_type, version, algorithm, metrics_json, status, trained_samples)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "PROFILE_CLASSIFIER", resp.ModelVersion, resp.Algorithm, string(metricsJSON), "ACTIVE", resp.SamplesTrained)

	if err != nil {
		return 0, err
	}

	modelID, _ := result.LastInsertId()
	return int(modelID), nil
}

// logMLRun logs training/prediction runs to ml_run_log
func (h *MLHandler) logMLRun(runType string, modelID int, success bool, errorSummary string, duration time.Duration, inputSamples, outputCount int) {
	var modelIDPtr *int
	if modelID > 0 {
		modelIDPtr = &modelID
	}

	successInt := 0
	if success {
		successInt = 1
	}

	_, err := h.db.Exec(`
		INSERT INTO ml_run_log (run_type, model_id, success, error_summary, duration_ms, input_samples, output_count)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, runType, modelIDPtr, successInt, errorSummary, duration.Milliseconds(), inputSamples, outputCount)

	if err != nil {
		log.Printf("Failed to log ML run: %v", err)
	}
}
