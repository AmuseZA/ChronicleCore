//go:build windows
// +build windows

package tracker

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                        = windows.NewLazySystemDLL("user32.dll")
	kernel32                      = windows.NewLazySystemDLL("kernel32.dll")
	oleaut32                      = windows.NewLazySystemDLL("oleaut32.dll")
	ole32                         = windows.NewLazySystemDLL("ole32.dll")
	procGetForegroundWindow       = user32.NewProc("GetForegroundWindow")
	procGetWindowText             = user32.NewProc("GetWindowTextW")
	procGetWindowThreadProcessId  = user32.NewProc("GetWindowThreadProcessId")
	procOpenProcess               = kernel32.NewProc("OpenProcess")
	procQueryFullProcessImageName = kernel32.NewProc("QueryFullProcessImageNameW")
	procGetLastInputInfo          = user32.NewProc("GetLastInputInfo")
	procGetTickCount              = kernel32.NewProc("GetTickCount")
	procGetClassName              = user32.NewProc("GetClassNameW")
	procFindWindowEx              = user32.NewProc("FindWindowExW")
	procSendMessage               = user32.NewProc("SendMessageW")
	procCoInitialize              = ole32.NewProc("CoInitialize")
	procCoUninitialize            = ole32.NewProc("CoUninitialize")
	procSysFreeString             = oleaut32.NewProc("SysFreeString")
)

const (
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
	MAX_PATH                  = 260
	WM_GETTEXT                = 0x000D
	WM_GETTEXTLENGTH          = 0x000E
)

type LASTINPUTINFO struct {
	cbSize uint32
	dwTime uint32
}

// WindowInfo contains information about the active window
type WindowInfo struct {
	ProcessName string
	WindowTitle string
	BrowserURL  string // URL extracted from browser address bar
	IsIdle      bool
	IdleSeconds int
	AppID       int64  // Resolved App ID
	TitleID     *int64 // Resolved Title ID
	DomainID    *int64 // Resolved Domain ID (for browser URLs)
	State       string // ACTIVE or IDLE
}

// Browser detection patterns
var browserPatterns = map[string]string{
	"chrome.exe":  "chrome",
	"msedge.exe":  "edge",
	"firefox.exe": "firefox",
	"opera.exe":   "opera",
	"brave.exe":   "brave",
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

// isBrowser checks if the process is a known browser
func isBrowser(processName string) (bool, string) {
	lower := strings.ToLower(processName)
	if browserType, ok := browserPatterns[lower]; ok {
		return true, browserType
	}
	return false, ""
}

// GetBrowserURL attempts to extract the URL from a browser's address bar
// This uses a combination of window title parsing and accessibility APIs
func GetBrowserURL(hwnd windows.HWND, processName string) string {
	isBrowserApp, browserType := isBrowser(processName)
	if !isBrowserApp {
		return ""
	}

	// Method 1: Extract URL from window title (works for most browsers when they show URL in title)
	// Many browsers show "Page Title - Browser Name" or "Page Title - URL - Browser Name"
	title, _ := GetWindowText(hwnd)
	if url := extractURLFromTitle(title, browserType); url != "" {
		return url
	}

	// Method 2: Try to get URL from address bar using UI Automation
	// This is more reliable but more complex
	url := getURLFromAddressBar(hwnd, browserType)
	if url != "" {
		return url
	}

	return ""
}

// extractURLFromTitle attempts to extract a URL from the window title
func extractURLFromTitle(title, browserType string) string {
	// Some browsers include the URL in the title, especially for certain pages
	// Look for common URL patterns
	if strings.Contains(title, "://") {
		// Find URL-like patterns
		parts := strings.Fields(title)
		for _, part := range parts {
			if strings.HasPrefix(part, "http://") || strings.HasPrefix(part, "https://") {
				// Clean up the URL (remove trailing characters that might be part of title formatting)
				url := strings.TrimRight(part, "|-–—")
				return url
			}
		}
	}
	return ""
}

// getURLFromAddressBar uses FindWindowEx to locate and read the address bar
// This is browser-specific and may not work for all versions
func getURLFromAddressBar(hwnd windows.HWND, browserType string) string {
	switch browserType {
	case "chrome", "edge", "brave", "opera":
		// Chromium-based browsers use similar window structures
		return getChromiumURL(hwnd)
	case "firefox":
		return getFirefoxURL(hwnd)
	}
	return ""
}

// getChromiumURL extracts URL from Chromium-based browsers (Chrome, Edge, Brave, Opera)
// Uses window hierarchy traversal to find the address bar
func getChromiumURL(hwnd windows.HWND) string {
	// Chromium browsers have a complex window hierarchy
	// The address bar is typically in a child window with class "Chrome_OmniboxView" or similar

	// Try to find address bar by enumerating child windows
	var addressBarHwnd windows.HWND

	// Look for known Chromium address bar class names
	classNames := []string{
		"Chrome_OmniboxView",
		"OmniboxViewViews",
	}

	for _, className := range classNames {
		classNamePtr, _ := syscall.UTF16PtrFromString(className)
		child, _, _ := procFindWindowEx.Call(
			uintptr(hwnd),
			0,
			uintptr(unsafe.Pointer(classNamePtr)),
			0,
		)
		if child != 0 {
			addressBarHwnd = windows.HWND(child)
			break
		}
	}

	if addressBarHwnd == 0 {
		// Try deeper search - Chromium has nested window structure
		addressBarHwnd = findChildWindowByClass(hwnd, "Chrome_OmniboxView", 5)
	}

	if addressBarHwnd == 0 {
		return ""
	}

	// Get text from the address bar
	return getWindowTextFromHwnd(addressBarHwnd)
}

// getFirefoxURL extracts URL from Firefox
func getFirefoxURL(hwnd windows.HWND) string {
	// Firefox uses different class names
	// The URL bar is typically "MozillaWindowClass" with specific child structure

	// Try common Firefox address bar patterns
	addressBarHwnd := findChildWindowByClass(hwnd, "MozillaCompositorWindowClass", 3)
	if addressBarHwnd == 0 {
		addressBarHwnd = findChildWindowByClass(hwnd, "MozillaWindowClass", 3)
	}

	if addressBarHwnd == 0 {
		return ""
	}

	return getWindowTextFromHwnd(addressBarHwnd)
}

// findChildWindowByClass recursively searches for a child window with the given class name
func findChildWindowByClass(parent windows.HWND, className string, maxDepth int) windows.HWND {
	if maxDepth <= 0 {
		return 0
	}

	classNamePtr, _ := syscall.UTF16PtrFromString(className)

	// First, try direct child
	child, _, _ := procFindWindowEx.Call(
		uintptr(parent),
		0,
		uintptr(unsafe.Pointer(classNamePtr)),
		0,
	)

	if child != 0 {
		return windows.HWND(child)
	}

	// If not found, search children recursively
	var foundHwnd windows.HWND
	enumChildWindows(parent, func(hwnd windows.HWND) bool {
		// Check this child's class
		actualClass := getWindowClass(hwnd)
		if actualClass == className {
			foundHwnd = hwnd
			return false // Stop enumeration
		}

		// Search this child's children
		if result := findChildWindowByClass(hwnd, className, maxDepth-1); result != 0 {
			foundHwnd = result
			return false // Stop enumeration
		}

		return true // Continue enumeration
	})

	return foundHwnd
}

// EnumChildProc is the callback type for EnumChildWindows
type EnumChildProc func(hwnd windows.HWND) bool

var (
	procEnumChildWindows = user32.NewProc("EnumChildWindows")
)

// enumChildWindows enumerates child windows
func enumChildWindows(parent windows.HWND, callback EnumChildProc) {
	// Note: This is a simplified version. For production, you'd use proper callback handling
	var child uintptr = 0
	for {
		child, _, _ = procFindWindowEx.Call(
			uintptr(parent),
			child,
			0,
			0,
		)
		if child == 0 {
			break
		}
		if !callback(windows.HWND(child)) {
			break
		}
	}
}

// getWindowClass gets the class name of a window
func getWindowClass(hwnd windows.HWND) string {
	buf := make([]uint16, 256)
	ret, _, _ := procGetClassName.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if ret == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

// getWindowTextFromHwnd gets text from a window handle using SendMessage
func getWindowTextFromHwnd(hwnd windows.HWND) string {
	// Get text length
	length, _, _ := procSendMessage.Call(
		uintptr(hwnd),
		WM_GETTEXTLENGTH,
		0,
		0,
	)

	if length == 0 {
		return ""
	}

	// Get text
	buf := make([]uint16, length+1)
	procSendMessage.Call(
		uintptr(hwnd),
		WM_GETTEXT,
		uintptr(length+1),
		uintptr(unsafe.Pointer(&buf[0])),
	)

	return syscall.UTF16ToString(buf)
}

// ExtractDomainFromURL extracts the domain from a URL
func ExtractDomainFromURL(url string) string {
	if url == "" {
		return ""
	}

	// Remove protocol
	domain := url
	if idx := strings.Index(domain, "://"); idx != -1 {
		domain = domain[idx+3:]
	}

	// Remove path
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	// Remove port
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	// Remove www. prefix for cleaner grouping
	domain = strings.TrimPrefix(domain, "www.")

	return domain
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

	// Try to get browser URL if this is a browser
	browserURL := GetBrowserURL(hwnd, processName)

	return &WindowInfo{
		ProcessName: processName,
		WindowTitle: title,
		BrowserURL:  browserURL,
		IsIdle:      false,
		IdleSeconds: idleSeconds,
	}, nil
}
