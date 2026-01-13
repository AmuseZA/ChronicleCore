package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"chroniclecore/internal/api"
	"chroniclecore/internal/engine"
	"chroniclecore/internal/ml"
	"chroniclecore/internal/store"
	"chroniclecore/internal/tracker"

	_ "github.com/mattn/go-sqlite3"
)

const (
	AppVersion          = "1.7.2"
	DefaultPort         = "8080"
	MLPort              = 8081
	UpdateCheckInterval = 30 * time.Minute
)

var (
	appStore      *store.Store
	appTracker    *tracker.Tracker
	appAggregator *engine.Aggregator
	mlSidecar     *ml.SidecarManager
	startTime     time.Time
)

func main() {
	log.Printf("ChronicleCore v%s starting...", AppVersion)
	startTime = time.Now()

	// Configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	// Get database path (default: user's AppData/Local/ChronicleCore)
	dbPath := getDefaultDBPath()
	log.Printf("Database path: %s", dbPath)

	// Initialize database
	if err := initializeDatabase(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer appStore.Close()

	// Initialize tracker
	appTracker = tracker.NewTracker(tracker.Config{
		Store:                appStore,
		PollInterval:         5 * time.Second,
		IdleThresholdSeconds: 300, // 5 minutes
	})

	// Initialize aggregator (auto-starts rollup scheduler)
	appAggregator = engine.NewAggregator(engine.AggregatorConfig{
		Store:          appStore,
		RetentionDays:  14,
		RollupInterval: 5 * time.Minute,
	})
	defer appAggregator.Stop()

	// Initialize ML sidecar (optional - starts Python process)
	mlSidecar, err := ml.NewSidecarManager(MLPort)
	if err != nil {
		log.Printf("⚠️  ML sidecar disabled: %v", err)
	} else {
		// Start ML in background so it doesn't block HTTP server startup
		go func() {
			log.Println("Starting ML sidecar in background...")
			if err := mlSidecar.Start(); err != nil {
				log.Printf("⚠️  ML sidecar failed to start: %v", err)
				// We can't set mlSidecar = nil here safely due to race conditions,
				// but the handler checks for nil or readiness anyway.
			} else {
				log.Println("✓ ML sidecar running")
			}
		}()
		// Ensure we stop it on shutdown
		defer mlSidecar.Stop()
	}

	// Initialize API handlers
	profileHandler := api.NewProfileHandler(appStore)
	exportHandler := api.NewExportHandler(appStore)
	blockHandler := api.NewBlockHandler(appStore)
	ruleHandler := api.NewRuleHandler(appStore)
	systemHandler := api.NewSystemHandler()
	blacklistHandler := api.NewBlacklistHandler(appStore)

	// ML handler (only if sidecar is running)
	var mlHandler *api.MLHandler
	if mlSidecar != nil {
		mlHandler = api.NewMLHandler(appStore.DB, mlSidecar)
	}

	// Setup HTTP server - MUST bind to localhost only
	mux := http.NewServeMux()

	// ---------------------------------------------------------
	// Static File Handling (SPA)
	// ---------------------------------------------------------

	// Get execution directory
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	// Development: use run path; Production: use exe path
	webDir := filepath.Join(exeDir, "web")
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		// Fallback for development (running from cmd/server)
		webDir = filepath.Join("..", "..", "chroniclecore-ui", "build")
	}

	log.Printf("Serving UI from: %s", webDir)

	// File server for static assets requires stripping the prefix if mounted at a subpath,
	// but here we serve at root, so we check for files/API first.
	fileServer := http.FileServer(http.Dir(webDir))

	// SPA Handler: Serves index.html for non-API 404s
	spaHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// API routes are handled by specific handlers registered below.
		// If we get here, it means no specific API route matched yet
		// (wait, ServeMux matches longest pattern, so we need to be careful).
		// Actually, ServeMux doesn't support regex/wildcard fallbacks easily.
		// So we will register "/" to this handler, and it will serve files or index.html.

		// Check if file exists in webDir
		cleanPath := filepath.Clean(path)
		fullPath := filepath.Join(webDir, cleanPath)

		if fileExists(fullPath) {
			// Serve static file
			fileServer.ServeHTTP(w, r)
			return
		}

		// If not found and not an API call, serve index.html (SPA Fallback)
		if !strings.HasPrefix(path, "/api/") {
			http.ServeFile(w, r, filepath.Join(webDir, "index.html"))
			return
		}

		// If it IS an API call, return 404
		http.Error(w, "Not found", http.StatusNotFound)
	})

	// Register root handler for SPA
	mux.Handle("/", spaHandler)

	// Health endpoint
	mux.HandleFunc("/health", handleHealth)

	// Tracking control
	mux.HandleFunc("/api/v1/tracking/status", handleTrackingStatus)
	mux.HandleFunc("/api/v1/tracking/start", handleTrackingStart)
	mux.HandleFunc("/api/v1/tracking/pause", handleTrackingPause)
	mux.HandleFunc("/api/v1/tracking/resume", handleTrackingResume)
	mux.HandleFunc("/api/v1/tracking/stop", handleTrackingStop)

	// Data endpoints
	mux.HandleFunc("/api/v1/blocks", blockHandler.ListBlocks)
	mux.HandleFunc("/api/v1/blocks/grouped", blockHandler.ListGroupedBlocks)
	mux.HandleFunc("/api/v1/blocks/manual", blockHandler.CreateManualEntry)
	mux.HandleFunc("/api/v1/blocks/", func(w http.ResponseWriter, r *http.Request) {
		// Route based on path suffix
		path := r.URL.Path
		if strings.HasSuffix(path, "/reassign") {
			blockHandler.ReassignBlock(w, r)
		} else if strings.HasSuffix(path, "/lock") {
			blockHandler.LockBlock(w, r)
		} else if r.Method == http.MethodDelete {
			blockHandler.DeleteBlock(w, r)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	// Profile management endpoints
	mux.HandleFunc("/api/v1/clients", profileHandler.ListClients)
	mux.HandleFunc("/api/v1/clients/create", profileHandler.CreateClient)
	mux.HandleFunc("/api/v1/services", profileHandler.ListServices)
	mux.HandleFunc("/api/v1/services/create", profileHandler.CreateService)
	mux.HandleFunc("/api/v1/rates", profileHandler.ListRates)
	mux.HandleFunc("/api/v1/rates/create", profileHandler.CreateRate)
	mux.HandleFunc("/api/v1/profiles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			profileHandler.ListProfiles(w, r)
		} else if r.Method == http.MethodPost {
			profileHandler.CreateProfile(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/v1/profiles/", func(w http.ResponseWriter, r *http.Request) {
		// Route based on path suffix
		path := r.URL.Path
		if strings.HasSuffix(path, "/stats") {
			profileHandler.GetProfileStats(w, r)
		} else if r.Method == http.MethodDelete {
			profileHandler.DeleteProfile(w, r)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	// Export endpoints
	mux.HandleFunc("/api/v1/export/invoice-lines", exportHandler.ExportInvoiceLines)

	// Rules management endpoints
	mux.HandleFunc("/api/v1/rules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			ruleHandler.ListRules(w, r)
		} else if r.Method == http.MethodPost {
			ruleHandler.CreateRule(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/v1/rules/", func(w http.ResponseWriter, r *http.Request) {
		// Handles /api/v1/rules/{id} for PUT and DELETE
		if r.Method == http.MethodPut {
			ruleHandler.UpdateRule(w, r)
		} else if r.Method == http.MethodDelete {
			ruleHandler.DeleteRule(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// System endpoints
	mux.HandleFunc("/api/v1/system/locale", systemHandler.GetLocale)
	mux.HandleFunc("/api/v1/system/check-update", func(w http.ResponseWriter, r *http.Request) {
		systemHandler.CheckForUpdate(w, r, AppVersion)
	})

	// Blacklist endpoints
	mux.HandleFunc("/api/v1/blacklist", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			blacklistHandler.ListBlacklist(w, r)
		} else if r.Method == http.MethodPost {
			blacklistHandler.AddToBlacklist(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/v1/blacklist/", blacklistHandler.RemoveFromBlacklist)
	mux.HandleFunc("/api/v1/blacklist/with-delete", blacklistHandler.BlacklistAndDeleteBlocks)

	// ML endpoints (only if sidecar is available)
	if mlHandler != nil {
		mux.HandleFunc("/api/v1/ml/status", mlHandler.GetMLStatus)
		mux.HandleFunc("/api/v1/ml/training-data", mlHandler.GetTrainingData)
		mux.HandleFunc("/api/v1/ml/train", mlHandler.TriggerTraining)
		mux.HandleFunc("/api/v1/ml/predict", mlHandler.PredictBlocks)
		mux.HandleFunc("/api/v1/ml/suggestions", mlHandler.GetSuggestions)
		mux.HandleFunc("/api/v1/ml/suggestions/accept", mlHandler.AcceptSuggestion)
		log.Println("✓ ML endpoints registered")
	}

	// Server configuration - CRITICAL: bind to 127.0.0.1 only
	addr := fmt.Sprintf("127.0.0.1:%s", port)
	server := &http.Server{
		Addr:         addr,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server listening on %s", addr)
		log.Printf("⚠️  SECURITY: Server bound to localhost only")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Start periodic update checker (every 30 minutes after initial check)
	go startUpdateChecker()

	<-done
	log.Println("Server shutting down...")

	// Stop tracker gracefully
	if appTracker != nil {
		appTracker.Stop()
	}

	// Stop ML sidecar
	if mlSidecar != nil {
		mlSidecar.Stop()
	}

	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

// getDefaultDBPath returns the default database path
func getDefaultDBPath() string {
	// Check environment variable first
	if dbPath := os.Getenv("CHRONICLE_DB_PATH"); dbPath != "" {
		return dbPath
	}

	// Use AppData/Local/ChronicleCore/chronicle.db
	appData := os.Getenv("LOCALAPPDATA")
	if appData == "" {
		appData = "." // Fallback to current directory
	}

	dir := filepath.Join(appData, "ChronicleCore")
	os.MkdirAll(dir, 0755)

	return filepath.Join(dir, "chronicle.db")
}

// initializeDatabase sets up the database and store
func initializeDatabase(dbPath string) error {
	// Check if database exists
	dbExists := fileExists(dbPath)

	if !dbExists {
		// Create database with schema
		log.Println("Database not found, creating with schema...")

		// Look for schema in multiple locations
		candidates := []string{
			// Production: adjacent to executable
			"schema.sql",
			filepath.Join(".", "schema.sql"),
			// Development: repo structure
			filepath.Join(".", "..", "..", "spec", "schema.sql"),
			// Alternate Development
			filepath.Join("spec", "schema.sql"),
		}

		var schemaPath string
		for _, p := range candidates {
			if fileExists(p) {
				schemaPath = p
				break
			}
		}

		if schemaPath == "" {
			return fmt.Errorf("schema.sql not found in any candidate paths: %v", candidates)
		}

		log.Printf("Loading schema from: %s", schemaPath)
		schemaSQL, err := os.ReadFile(schemaPath)
		if err != nil {
			return fmt.Errorf("failed to read schema: %w", err)
		}

		// Create database
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}

		_, err = db.Exec(string(schemaSQL))
		db.Close()

		if err != nil {
			return fmt.Errorf("failed to apply schema: %w", err)
		}

		log.Println("Database created successfully")
	}

	// Initialize store
	appStore = store.NewStore(dbPath)
	if err := appStore.Init(); err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// corsMiddleware - STRICT localhost-only CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Only allow localhost origins
		if origin == "http://127.0.0.1" || origin == "http://localhost" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Health check handler
func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	uptime := int(time.Since(startTime).Seconds())
	fmt.Fprintf(w, `{"status":"ok","version":"%s","uptime_seconds":%d}`, AppVersion, uptime)
}

// Tracking status handler
func handleTrackingStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := appTracker.GetStatus()

	response := map[string]interface{}{
		"state":          string(status.State),
		"last_active_at": nil,
		"idle_seconds":   status.IdleSeconds,
		"current_window": nil,
	}

	if status.LastActiveAt != nil {
		response["last_active_at"] = status.LastActiveAt.Format(time.RFC3339)
	}

	if status.CurrentWindow != nil {
		response["current_window"] = map[string]interface{}{
			"app_name": status.CurrentWindow.ProcessName,
			"title":    status.CurrentWindow.WindowTitle,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Tracking control handlers
func handleTrackingStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := appTracker.Start(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return updated status
	handleTrackingStatus(w, &http.Request{Method: "GET"})
}

func handleTrackingPause(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := appTracker.Pause(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	handleTrackingStatus(w, &http.Request{Method: "GET"})
}

func handleTrackingResume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := appTracker.Resume(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	handleTrackingStatus(w, &http.Request{Method: "GET"})
}

func handleTrackingStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := appTracker.Stop(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	handleTrackingStatus(w, &http.Request{Method: "GET"})
}

// startUpdateChecker runs periodic update checks
func startUpdateChecker() {
	// Initial check after 1 minute (give server time to fully start)
	time.Sleep(1 * time.Minute)
	checkForUpdateBackground()

	// Then check every 30 minutes
	ticker := time.NewTicker(UpdateCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		checkForUpdateBackground()
	}
}

// checkForUpdateBackground checks GitHub for updates and logs if one is available
func checkForUpdateBackground() {
	client := &http.Client{Timeout: 10 * time.Second}
	url := "https://api.github.com/repos/AmuseZA/ChronicleCore/releases/latest"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", "ChronicleCore/"+AppVersion)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Update check failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var release struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")

	// Simple version comparison
	if latestVersion != AppVersion && latestVersion > AppVersion {
		log.Printf("🔔 Update available: v%s -> v%s (check Settings to download)", AppVersion, latestVersion)
	}
}
