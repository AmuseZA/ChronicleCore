package tracker

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

type UIAResult struct {
	Title   string `json:"title"`
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

var (
	lastUIATitle string
	lastUIATime  time.Time
	uiaCacheTTL  = 3 * time.Second // Don't spam PowerShell more than once every 3s
)

// GetUIATitle executes a PowerShell script to get the UIA Document title
// This is used for apps like Opera Sidebar that hide their true title from Win32 APIs
func GetUIATitle() string {
	if time.Since(lastUIATime) < uiaCacheTTL && lastUIATitle != "" {
		return lastUIATitle
	}

	// Locate script
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	scriptPath := filepath.Join(exeDir, "scripts", "get_uia_title.ps1")

	// Dev fallback
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		scriptPath = filepath.Join("..", "..", "scripts", "get_uia_title.ps1")
	}

	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath)

	// Create new window creation flags to hide the console window flashing
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("UIA Error: %v\n", err)
		return ""
	}

	var result UIAResult
	if err := json.Unmarshal(out, &result); err != nil {
		// fmt.Printf("JSON Error: %v | Output: %s\n", err, string(out))
		return ""
	}

	if result.Success {
		lastUIATitle = result.Title
		lastUIATime = time.Now()
		return result.Title
	}

	return ""
}
