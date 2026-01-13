//go:build windows
// +build windows

package tracker

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                        = windows.NewLazySystemDLL("user32.dll")
	kernel32                      = windows.NewLazySystemDLL("kernel32.dll")
	procGetForegroundWindow       = user32.NewProc("GetForegroundWindow")
	procGetWindowText             = user32.NewProc("GetWindowTextW")
	procGetWindowThreadProcessId  = user32.NewProc("GetWindowThreadProcessId")
	procOpenProcess               = kernel32.NewProc("OpenProcess")
	procQueryFullProcessImageName = kernel32.NewProc("QueryFullProcessImageNameW")
	procGetLastInputInfo          = user32.NewProc("GetLastInputInfo")
	procGetTickCount              = kernel32.NewProc("GetTickCount")
)

const (
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
	MAX_PATH                  = 260
)

type LASTINPUTINFO struct {
	cbSize uint32
	dwTime uint32
}

// WindowInfo contains information about the active window
type WindowInfo struct {
	ProcessName string
	WindowTitle string
	IsIdle      bool
	IdleSeconds int
	AppID       int64  // Resolved App ID
	TitleID     *int64 // Resolved Title ID
	State       string // ACTIVE or IDLE
}

// GetForegroundWindow returns the handle of the foreground window
func GetForegroundWindow() (windows.HWND, error) {
	ret, _, err := procGetForegroundWindow.Call()
	if ret == 0 {
		return 0, fmt.Errorf("GetForegroundWindow failed: %v", err)
	}
	return windows.HWND(ret), nil
}

// GetWindowText retrieves the title of a window
func GetWindowText(hwnd windows.HWND) (string, error) {
	buf := make([]uint16, 512)
	ret, _, _ := procGetWindowText.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)

	if ret == 0 {
		return "", nil // Empty title is not an error
	}

	return syscall.UTF16ToString(buf), nil
}

// GetWindowProcessName retrieves the process name for a window
func GetWindowProcessName(hwnd windows.HWND) (string, error) {
	var processID uint32

	// Get process ID
	ret, _, _ := procGetWindowThreadProcessId.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&processID)),
	)

	if ret == 0 {
		return "", fmt.Errorf("GetWindowThreadProcessId failed")
	}

	// Open process handle
	handle, _, err := procOpenProcess.Call(
		PROCESS_QUERY_INFORMATION|PROCESS_VM_READ,
		0,
		uintptr(processID),
	)

	if handle == 0 {
		return "", fmt.Errorf("OpenProcess failed: %v", err)
	}
	defer windows.CloseHandle(windows.Handle(handle))

	// Get process image name
	buf := make([]uint16, MAX_PATH)
	size := uint32(MAX_PATH)

	ret, _, _ = procQueryFullProcessImageName.Call(
		handle,
		0, // Win32 path format
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)

	if ret == 0 {
		// Fallback: try to get just the executable name
		return getProcessNameFallback(processID)
	}

	fullPath := syscall.UTF16ToString(buf)

	// Extract just the filename
	return extractFileName(fullPath), nil
}

// getProcessNameFallback attempts to get process name using alternative method
func getProcessNameFallback(processID uint32) (string, error) {
	// This is a simplified fallback - in production you might want psapi.dll
	return fmt.Sprintf("PID_%d", processID), nil
}

// extractFileName extracts the filename from a full path
func extractFileName(path string) string {
	// Find last backslash
	lastSlash := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '\\' || path[i] == '/' {
			lastSlash = i
			break
		}
	}

	if lastSlash >= 0 && lastSlash < len(path)-1 {
		return path[lastSlash+1:]
	}

	return path
}

// GetIdleTime returns the number of milliseconds the system has been idle
func GetIdleTime() (int64, error) {
	var lii LASTINPUTINFO
	lii.cbSize = uint32(unsafe.Sizeof(lii))

	ret, _, err := procGetLastInputInfo.Call(uintptr(unsafe.Pointer(&lii)))
	if ret == 0 {
		return 0, fmt.Errorf("GetLastInputInfo failed: %v", err)
	}

	// Get current tick count
	tickCount, _, _ := procGetTickCount.Call()

	// Calculate idle time in milliseconds
	idleMillis := uint32(tickCount) - lii.dwTime

	return int64(idleMillis), nil
}

// GetIdleTimeSec returns the number of seconds the system has been idle
func GetIdleTimeSec() (int, error) {
	idleMillis, err := GetIdleTime()
	if err != nil {
		return 0, err
	}
	return int(idleMillis / 1000), nil
}

// GetCurrentWindowInfo retrieves information about the current foreground window
func GetCurrentWindowInfo(idleThresholdSeconds int) (*WindowInfo, error) {
	// Check idle time first
	idleSeconds, err := GetIdleTimeSec()
	if err != nil {
		return nil, fmt.Errorf("failed to get idle time: %w", err)
	}

	isIdle := idleSeconds >= idleThresholdSeconds

	// If idle, return minimal info
	if isIdle {
		return &WindowInfo{
			ProcessName: "System",
			WindowTitle: "Idle",
			IsIdle:      true,
			IdleSeconds: idleSeconds,
		}, nil
	}

	// Get foreground window
	hwnd, err := GetForegroundWindow()
	if err != nil {
		return nil, fmt.Errorf("failed to get foreground window: %w", err)
	}

	// Get window title
	title, err := GetWindowText(hwnd)
	if err != nil {
		title = "" // Non-fatal
	}

	// Get process name
	processName, err := GetWindowProcessName(hwnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get process name: %w", err)
	}

	return &WindowInfo{
		ProcessName: processName,
		WindowTitle: title,
		IsIdle:      false,
		IdleSeconds: idleSeconds,
	}, nil
}
