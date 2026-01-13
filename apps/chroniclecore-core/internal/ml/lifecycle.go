package ml

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// SidecarManager manages the ML sidecar process lifecycle
type SidecarManager struct {
	cmd         *exec.Cmd
	ctx         context.Context
	cancel      context.CancelFunc
	port        int
	token       string
	pythonPath  string
	sidecarPath string
	isRunning   bool
}

// NewSidecarManager creates a new sidecar lifecycle manager
func NewSidecarManager(port int) (*SidecarManager, error) {
	// Generate secure token for authentication
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Find Python executable
	pythonPath, err := findPython()
	if err != nil {
		return nil, fmt.Errorf("python not found: %w", err)
	}

	// Find sidecar directory
	sidecarPath, err := findSidecarPath()
	if err != nil {
		return nil, fmt.Errorf("sidecar path not found: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &SidecarManager{
		ctx:         ctx,
		cancel:      cancel,
		port:        port,
		token:       token,
		pythonPath:  pythonPath,
		sidecarPath: sidecarPath,
		isRunning:   false,
	}, nil
}

// Start launches the ML sidecar process
func (sm *SidecarManager) Start() error {
	if sm.isRunning {
		return fmt.Errorf("sidecar already running")
	}

	log.Printf("Starting ML sidecar on port %d", sm.port)

	// Set up command
	sm.cmd = exec.CommandContext(
		sm.ctx,
		sm.pythonPath,
		"-m", "uvicorn",
		"src.main:app",
		"--host", "127.0.0.1",
		"--port", fmt.Sprintf("%d", sm.port),
		"--log-level", "info",
	)

	// Set working directory to sidecar path
	sm.cmd.Dir = sm.sidecarPath

	// Set environment variables
	sm.cmd.Env = append(os.Environ(),
		fmt.Sprintf("CC_ML_TOKEN=%s", sm.token),
		fmt.Sprintf("ML_PORT=%d", sm.port),
		"PYTHONUNBUFFERED=1",
	)

	// Capture output
	stdout, _ := sm.cmd.StdoutPipe()
	stderr, _ := sm.cmd.StderrPipe()

	// Start process
	if err := sm.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start sidecar: %w", err)
	}

	// Stream logs in background
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			log.Printf("[ML Sidecar] %s", scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("[ML Sidecar ERROR] %s", scanner.Text())
		}
	}()

	sm.isRunning = true
	log.Printf("ML sidecar started (PID: %d)", sm.cmd.Process.Pid)

	// Wait for sidecar to be ready
	if err := sm.waitForReady(30 * time.Second); err != nil {
		sm.Stop()
		return fmt.Errorf("sidecar failed to start: %w", err)
	}

	log.Println("ML sidecar is ready")

	// Monitor process in background
	go sm.monitor()

	return nil
}

// Stop gracefully stops the ML sidecar
func (sm *SidecarManager) Stop() error {
	if !sm.isRunning {
		return nil
	}

	log.Println("Stopping ML sidecar...")

	// Cancel context to signal shutdown
	sm.cancel()

	// Give process time to shut down gracefully
	done := make(chan error, 1)
	go func() {
		done <- sm.cmd.Wait()
	}()

	select {
	case <-time.After(10 * time.Second):
		// Force kill if not stopped
		log.Println("ML sidecar not responding, force killing...")
		if sm.cmd.Process != nil {
			sm.cmd.Process.Kill()
		}
	case err := <-done:
		if err != nil {
			log.Printf("ML sidecar stopped with error: %v", err)
		} else {
			log.Println("ML sidecar stopped cleanly")
		}
	}

	sm.isRunning = false
	return nil
}

// IsRunning returns whether the sidecar is currently running and healthy
func (sm *SidecarManager) IsRunning() bool {
	if !sm.isRunning {
		return false
	}
	// Actually verify with health check
	client := NewClient(sm.port, sm.token)
	if err := client.HealthCheck(); err != nil {
		log.Printf("ML sidecar health check failed: %v", err)
		return false
	}
	return true
}

// IsProcessRunning returns whether the process flag indicates running (without health check)
func (sm *SidecarManager) IsProcessRunning() bool {
	return sm.isRunning
}

// GetToken returns the authentication token for the sidecar
func (sm *SidecarManager) GetToken() string {
	return sm.token
}

// GetPort returns the port the sidecar is running on
func (sm *SidecarManager) GetPort() int {
	return sm.port
}

// Restart stops and restarts the sidecar
func (sm *SidecarManager) Restart() error {
	log.Println("Restarting ML sidecar...")

	if err := sm.Stop(); err != nil {
		return fmt.Errorf("failed to stop sidecar: %w", err)
	}

	// Wait a moment before restarting
	time.Sleep(2 * time.Second)

	// Create new context
	sm.ctx, sm.cancel = context.WithCancel(context.Background())

	return sm.Start()
}

// waitForReady polls the health endpoint until the sidecar is ready
func (sm *SidecarManager) waitForReady(timeout time.Duration) error {
	client := NewClient(sm.port, sm.token)
	deadline := time.Now().Add(timeout)

	// Check for early exit
	exitChan := make(chan error, 1)
	go func() {
		state, err := sm.cmd.Process.Wait()
		if err != nil {
			exitChan <- err
		} else if !state.Success() {
			exitChan <- fmt.Errorf("process exited with code %d", state.ExitCode())
		} else {
			exitChan <- fmt.Errorf("process exited unexpectedly")
		}
	}()

	for time.Now().Before(deadline) {
		// Check if process died
		select {
		case err := <-exitChan:
			return fmt.Errorf("process died during startup: %w", err)
		default:
			// Continue
		}

		if err := client.HealthCheck(); err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("sidecar did not become ready within %v", timeout)
}

// monitor watches the sidecar process and restarts if it crashes
func (sm *SidecarManager) monitor() {
	err := sm.cmd.Wait()

	if sm.ctx.Err() != nil {
		// Context was cancelled, intentional shutdown
		return
	}

	// Process crashed unexpectedly
	log.Printf("ML sidecar crashed: %v", err)
	sm.isRunning = false

	// Wait before attempting restart
	time.Sleep(5 * time.Second)

	log.Println("Attempting to restart ML sidecar...")
	if err := sm.Restart(); err != nil {
		log.Printf("Failed to restart ML sidecar: %v", err)
	}
}

// generateToken creates a secure random token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// findPython locates the Python executable
func findPython() (string, error) {
	// 1. Check for embedded python (packaging support)
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		embeddedPython := filepath.Join(exeDir, "python", "python.exe")
		if _, err := os.Stat(embeddedPython); err == nil {
			log.Printf("Found embedded Python: %s", embeddedPython)
			return embeddedPython, nil
		}
	}

	// 2. Try common Python paths on Windows
	candidates := []string{
		"python",
		"python3",
		"py",
	}

	for _, candidate := range candidates {
		path, err := exec.LookPath(candidate)
		if err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("python executable not found")
}

// findSidecarPath locates the chronicle-ml directory
func findSidecarPath() (string, error) {
	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	exeDir := filepath.Dir(exePath)

	// Try relative paths
	candidates := []string{
		// 1. Embedded Packaging (dist/ml)
		filepath.Join(exeDir, "ml"),
		// 2. Production path (adjacent to executable)
		filepath.Join(exeDir, "chronicle-ml"),
		// 3. Development path
		filepath.Join(exeDir, "..", "..", "apps", "chronicle-ml"),
		// 4. Alternative production path
		filepath.Join(exeDir, "..", "chronicle-ml"),
	}

	for _, candidate := range candidates {
		absPath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}

		// Check if main.py exists
		mainPy := filepath.Join(absPath, "src", "main.py")
		if _, err := os.Stat(mainPy); err == nil {
			log.Printf("Found ML sidecar at: %s", absPath)
			return absPath, nil
		}
	}

	return "", fmt.Errorf("chronicle-ml directory not found")
}
