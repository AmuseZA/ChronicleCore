package ml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client communicates with the ML sidecar via HTTP
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new ML sidecar client
func NewClient(port int, token string) *Client {
	return &Client{
		baseURL: fmt.Sprintf("http://127.0.0.1:%d", port),
		token:   token,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// HealthCheck verifies the sidecar is running
func (c *Client) HealthCheck() error {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// TrainRequest represents a training request
type TrainRequest struct {
	Features  []map[string]interface{} `json:"features"`
	Labels    []int                    `json:"labels"`
	ModelType string                   `json:"model_type"`
}

// TrainResponse represents a training response
type TrainResponse struct {
	Success        bool               `json:"success"`
	ModelVersion   string             `json:"model_version"`
	Algorithm      string             `json:"algorithm"`
	Metrics        map[string]float64 `json:"metrics"`
	SamplesTrained int                `json:"samples_trained"`
	Message        string             `json:"message,omitempty"`
}

// Train sends a training request to the sidecar
func (c *Client) Train(req TrainRequest) (*TrainResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/train", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-CC-Token", c.token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("training request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("training failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result TrainResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// PredictRequest represents a prediction request
type PredictRequest struct {
	Features  []map[string]interface{} `json:"features"`
	Threshold float64                  `json:"threshold"`
}

// PredictionResult represents a single prediction
type PredictionResult struct {
	BlockIndex          int     `json:"block_index"`
	PredictedProfileID  int     `json:"predicted_profile_id"`
	Confidence          float64 `json:"confidence"`
	ConfidenceLevel     string  `json:"confidence_level"`
}

// PredictResponse represents a prediction response
type PredictResponse struct {
	Success          bool               `json:"success"`
	Predictions      []PredictionResult `json:"predictions"`
	ModelVersion     string             `json:"model_version"`
	TotalPredictions int                `json:"total_predictions"`
	Message          string             `json:"message,omitempty"`
}

// Predict sends a prediction request to the sidecar
func (c *Client) Predict(req PredictRequest) (*PredictResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/predict", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-CC-Token", c.token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("prediction request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("prediction failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result PredictResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// BlockData represents a block for clustering
type BlockData struct {
	BlockID int    `json:"block_id"`
	TsStart string `json:"ts_start"`
	TsEnd   string `json:"ts_end"`
	AppName string `json:"app_name,omitempty"`
	Title   string `json:"title,omitempty"`
}

// ClusterRequest represents a clustering request
type ClusterRequest struct {
	Blocks              []BlockData `json:"blocks"`
	GapThresholdMinutes int         `json:"gap_threshold_minutes"`
}

// SessionData represents a session cluster
type SessionData struct {
	SessionID       int     `json:"session_id"`
	BlockIDs        []int   `json:"block_ids"`
	StartTime       string  `json:"start_time"`
	EndTime         string  `json:"end_time"`
	DurationMinutes float64 `json:"duration_minutes"`
	BlockCount      int     `json:"block_count"`
}

// ClusterResponse represents a clustering response
type ClusterResponse struct {
	Success       bool          `json:"success"`
	Sessions      []SessionData `json:"sessions"`
	TotalBlocks   int           `json:"total_blocks"`
	TotalSessions int           `json:"total_sessions"`
	Message       string        `json:"message,omitempty"`
}

// Cluster sends a clustering request to the sidecar
func (c *Client) Cluster(req ClusterRequest) (*ClusterResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/cluster", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-CC-Token", c.token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("clustering request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("clustering failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result ClusterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
