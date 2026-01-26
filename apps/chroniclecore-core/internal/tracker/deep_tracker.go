//go:build windows
// +build windows

package tracker

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// DeepTrackingInfo contains detailed information extracted from an application
type DeepTrackingInfo struct {
	// Basic info (always populated)
	AppName     string `json:"app_name"`
	WindowTitle string `json:"window_title"`

	// Deep tracking info (when enabled)
	DocumentName    string   `json:"document_name,omitempty"`    // File being edited
	DocumentPath    string   `json:"document_path,omitempty"`    // Full path if available
	EmailSubject    string   `json:"email_subject,omitempty"`    // Email subject line
	EmailSender     string   `json:"email_sender,omitempty"`     // Email sender
	EmailRecipients []string `json:"email_recipients,omitempty"` // Email recipients
	ChatContact     string   `json:"chat_contact,omitempty"`     // Chat partner name
	ChatChannel     string   `json:"chat_channel,omitempty"`     // Slack/Teams channel
	BrowserURL      string   `json:"browser_url,omitempty"`      // Full URL
	BrowserDomain   string   `json:"browser_domain,omitempty"`   // Domain only
	PageTitle       string   `json:"page_title,omitempty"`       // Web page title
	ProjectName     string   `json:"project_name,omitempty"`     // IDE project
	FileName        string   `json:"file_name,omitempty"`        // Current file in IDE
	ContentSummary  string   `json:"content_summary,omitempty"`  // Brief content summary
	ActivityType    string   `json:"activity_type,omitempty"`    // Type: editing, reading, composing, etc.
}

// DeepTrackerConfig holds configuration for deep tracking
type DeepTrackerConfig struct {
	TrackBrowserContent  bool
	TrackEmailContent    bool
	TrackDocumentContent bool
	TrackChatContent     bool
	PrivacyMode          bool // Redacts sensitive content
	ExcludedApps         []string
}

// DeepTracker extracts detailed information from applications
type DeepTracker struct {
	config DeepTrackerConfig
}

// NewDeepTracker creates a new deep tracker instance
func NewDeepTracker(config DeepTrackerConfig) *DeepTracker {
	return &DeepTracker{config: config}
}

// ExtractDeepInfo extracts detailed information from the current window
func (dt *DeepTracker) ExtractDeepInfo(hwnd windows.HWND, processName, windowTitle string) *DeepTrackingInfo {
	info := &DeepTrackingInfo{
		AppName:     processName,
		WindowTitle: windowTitle,
	}

	// Check if app is excluded
	if dt.isExcluded(processName) {
		return info
	}

	// Normalize process name for matching
	procLower := strings.ToLower(processName)

	// Route to appropriate extractor based on application
	switch {
	// Microsoft Office
	case strings.Contains(procLower, "outlook"):
		if dt.config.TrackEmailContent {
			dt.extractOutlookInfo(hwnd, info)
		}
	case strings.Contains(procLower, "winword") || strings.Contains(procLower, "word"):
		if dt.config.TrackDocumentContent {
			dt.extractWordInfo(hwnd, windowTitle, info)
		}
	case strings.Contains(procLower, "excel"):
		if dt.config.TrackDocumentContent {
			dt.extractExcelInfo(hwnd, windowTitle, info)
		}
	case strings.Contains(procLower, "powerpnt") || strings.Contains(procLower, "powerpoint"):
		if dt.config.TrackDocumentContent {
			dt.extractPowerPointInfo(hwnd, windowTitle, info)
		}

	// Browsers
	case strings.Contains(procLower, "chrome"), strings.Contains(procLower, "msedge"),
		strings.Contains(procLower, "firefox"), strings.Contains(procLower, "brave"),
		strings.Contains(procLower, "opera"):
		if dt.config.TrackBrowserContent {
			dt.extractBrowserInfo(hwnd, processName, windowTitle, info)
		}

	// IDEs and Code Editors
	case strings.Contains(procLower, "code") || strings.Contains(procLower, "vscode"):
		if dt.config.TrackDocumentContent {
			dt.extractVSCodeInfo(hwnd, windowTitle, info)
		}
	case strings.Contains(procLower, "devenv"): // Visual Studio
		if dt.config.TrackDocumentContent {
			dt.extractVisualStudioInfo(hwnd, windowTitle, info)
		}
	case strings.Contains(procLower, "idea"), strings.Contains(procLower, "pycharm"),
		strings.Contains(procLower, "webstorm"), strings.Contains(procLower, "rider"):
		if dt.config.TrackDocumentContent {
			dt.extractJetBrainsInfo(hwnd, windowTitle, info)
		}

	// Communication Apps
	case strings.Contains(procLower, "teams"):
		if dt.config.TrackChatContent {
			dt.extractTeamsInfo(hwnd, windowTitle, info)
		}
	case strings.Contains(procLower, "slack"):
		if dt.config.TrackChatContent {
			dt.extractSlackInfo(hwnd, windowTitle, info)
		}
	case strings.Contains(procLower, "discord"):
		if dt.config.TrackChatContent {
			dt.extractDiscordInfo(hwnd, windowTitle, info)
		}
	case strings.Contains(procLower, "whatsapp"):
		if dt.config.TrackChatContent {
			dt.extractWhatsAppInfo(hwnd, windowTitle, info)
		}
	case strings.Contains(procLower, "telegram"):
		if dt.config.TrackChatContent {
			dt.extractTelegramInfo(hwnd, windowTitle, info)
		}

	// Design Tools
	case strings.Contains(procLower, "figma"):
		if dt.config.TrackDocumentContent {
			dt.extractFigmaInfo(hwnd, windowTitle, info)
		}

	// Other common apps
	case strings.Contains(procLower, "notepad"):
		if dt.config.TrackDocumentContent {
			dt.extractNotepadInfo(hwnd, windowTitle, info)
		}
	case strings.Contains(procLower, "explorer"):
		dt.extractExplorerInfo(hwnd, windowTitle, info)

	default:
		// Generic extraction from window title
		dt.extractGenericInfo(windowTitle, info)
	}

	// Apply privacy mode if enabled
	if dt.config.PrivacyMode {
		dt.applyPrivacyRedaction(info)
	}

	return info
}

// isExcluded checks if an app is in the exclusion list
func (dt *DeepTracker) isExcluded(processName string) bool {
	procLower := strings.ToLower(processName)
	for _, excluded := range dt.config.ExcludedApps {
		if strings.Contains(procLower, strings.ToLower(excluded)) {
			return true
		}
	}
	return false
}

// ============================================================
// Microsoft Office Extractors
// ============================================================

func (dt *DeepTracker) extractOutlookInfo(hwnd windows.HWND, info *DeepTrackingInfo) {
	info.ActivityType = "email"

	// Try to extract email info from window title
	// Outlook titles: "Subject - Sender - Outlook" or "Inbox - email@domain.com - Outlook"
	title := info.WindowTitle

	// Check if reading/composing email
	if strings.Contains(title, " - Message") {
		info.ActivityType = "composing_email"
	}

	// Try to extract subject from title
	parts := strings.Split(title, " - ")
	if len(parts) >= 2 {
		// First part is usually the subject or folder name
		subject := strings.TrimSpace(parts[0])
		if subject != "Inbox" && subject != "Sent Items" && subject != "Drafts" {
			info.EmailSubject = subject
		}

		// Second part might be sender email
		if len(parts) >= 3 && strings.Contains(parts[1], "@") {
			info.EmailSender = strings.TrimSpace(parts[1])
		}
	}

	// Try UI Automation to get more details
	dt.extractOutlookUIAutomation(hwnd, info)
}

func (dt *DeepTracker) extractOutlookUIAutomation(hwnd windows.HWND, info *DeepTrackingInfo) {
	// Try to find the reading pane or compose window
	// Look for elements with known Outlook UI classes

	// Find subject field
	subjectHwnd := findChildWindowByClass(hwnd, "RichEdit20WPT", 3)
	if subjectHwnd != 0 {
		subject := getWindowTextFromHwnd(subjectHwnd)
		if subject != "" && info.EmailSubject == "" {
			info.EmailSubject = truncateString(subject, 100)
		}
	}
}

func (dt *DeepTracker) extractWordInfo(hwnd windows.HWND, title string, info *DeepTrackingInfo) {
	info.ActivityType = "editing_document"

	// Word title format: "Document Name - Word" or "Document Name.docx - Word"
	if idx := strings.LastIndex(title, " - "); idx > 0 {
		docName := strings.TrimSpace(title[:idx])
		info.DocumentName = docName
		info.FileName = docName
	}
}

func (dt *DeepTracker) extractExcelInfo(hwnd windows.HWND, title string, info *DeepTrackingInfo) {
	info.ActivityType = "editing_spreadsheet"

	// Excel title format: "Workbook Name - Excel"
	if idx := strings.LastIndex(title, " - "); idx > 0 {
		docName := strings.TrimSpace(title[:idx])
		info.DocumentName = docName
		info.FileName = docName
	}
}

func (dt *DeepTracker) extractPowerPointInfo(hwnd windows.HWND, title string, info *DeepTrackingInfo) {
	info.ActivityType = "editing_presentation"

	// PowerPoint title format: "Presentation Name - PowerPoint"
	if idx := strings.LastIndex(title, " - "); idx > 0 {
		docName := strings.TrimSpace(title[:idx])
		info.DocumentName = docName
		info.FileName = docName
	}
}

// ============================================================
// Browser Extractors
// ============================================================

func (dt *DeepTracker) extractBrowserInfo(hwnd windows.HWND, processName, title string, info *DeepTrackingInfo) {
	info.ActivityType = "browsing"

	// Try to get URL from address bar
	url := GetBrowserURL(hwnd, processName)
	if url != "" {
		info.BrowserURL = url
		info.BrowserDomain = ExtractDomainFromURL(url)
	}

	// Extract page title from window title
	// Format: "Page Title - Browser Name" or "Page Title — Browser Name"
	browserSuffixes := []string{
		" - Google Chrome", " - Chrome",
		" - Microsoft Edge", " - Edge",
		" - Mozilla Firefox", " - Firefox",
		" - Brave", " - Opera",
		" — Google Chrome", " — Mozilla Firefox",
	}

	pageTitle := title
	for _, suffix := range browserSuffixes {
		if idx := strings.LastIndex(pageTitle, suffix); idx > 0 {
			pageTitle = strings.TrimSpace(pageTitle[:idx])
			break
		}
	}
	info.PageTitle = pageTitle

	// Detect specific activity types based on URL/domain
	if info.BrowserDomain != "" {
		dt.detectBrowserActivity(info)
	}
}

func (dt *DeepTracker) detectBrowserActivity(info *DeepTrackingInfo) {
	domain := strings.ToLower(info.BrowserDomain)

	switch {
	case strings.Contains(domain, "mail.google.com"), strings.Contains(domain, "outlook."):
		info.ActivityType = "webmail"
	case strings.Contains(domain, "docs.google.com"):
		info.ActivityType = "editing_document"
	case strings.Contains(domain, "sheets.google.com"):
		info.ActivityType = "editing_spreadsheet"
	case strings.Contains(domain, "github.com"):
		if strings.Contains(info.BrowserURL, "/pull/") {
			info.ActivityType = "reviewing_code"
		} else if strings.Contains(info.BrowserURL, "/issues/") {
			info.ActivityType = "tracking_issues"
		} else {
			info.ActivityType = "coding"
		}
	case strings.Contains(domain, "slack.com"):
		info.ActivityType = "messaging"
	case strings.Contains(domain, "youtube.com"):
		info.ActivityType = "watching_video"
	case strings.Contains(domain, "linkedin.com"):
		info.ActivityType = "networking"
	case strings.Contains(domain, "figma.com"):
		info.ActivityType = "designing"
	}
}

// ============================================================
// IDE/Code Editor Extractors
// ============================================================

func (dt *DeepTracker) extractVSCodeInfo(hwnd windows.HWND, title string, info *DeepTrackingInfo) {
	info.ActivityType = "coding"

	// VS Code title format: "filename - folder - Visual Studio Code"
	// Or: "filename - project - Visual Studio Code"
	parts := strings.Split(title, " - ")
	if len(parts) >= 3 {
		info.FileName = strings.TrimSpace(parts[0])
		info.ProjectName = strings.TrimSpace(parts[1])
	} else if len(parts) >= 2 {
		info.FileName = strings.TrimSpace(parts[0])
	}

	// Detect file type from extension
	if info.FileName != "" {
		ext := getFileExtension(info.FileName)
		if ext != "" {
			info.ContentSummary = fmt.Sprintf("Editing %s file", ext)
		}
	}
}

func (dt *DeepTracker) extractVisualStudioInfo(hwnd windows.HWND, title string, info *DeepTrackingInfo) {
	info.ActivityType = "coding"

	// Visual Studio title: "Solution Name - Microsoft Visual Studio"
	// Or: "File - Solution - Microsoft Visual Studio"
	if idx := strings.Index(title, " - Microsoft Visual Studio"); idx > 0 {
		projectInfo := title[:idx]
		parts := strings.Split(projectInfo, " - ")
		if len(parts) >= 2 {
			info.FileName = strings.TrimSpace(parts[0])
			info.ProjectName = strings.TrimSpace(parts[len(parts)-1])
		} else {
			info.ProjectName = strings.TrimSpace(projectInfo)
		}
	}
}

func (dt *DeepTracker) extractJetBrainsInfo(hwnd windows.HWND, title string, info *DeepTrackingInfo) {
	info.ActivityType = "coding"

	// JetBrains IDE title: "project – file – IDE Name"
	parts := strings.Split(title, " – ")
	if len(parts) >= 2 {
		info.ProjectName = strings.TrimSpace(parts[0])
		if len(parts) >= 3 {
			info.FileName = strings.TrimSpace(parts[1])
		}
	}
}

// ============================================================
// Communication App Extractors
// ============================================================

func (dt *DeepTracker) extractTeamsInfo(hwnd windows.HWND, title string, info *DeepTrackingInfo) {
	info.ActivityType = "messaging"

	// Teams title: "Channel/Contact | Microsoft Teams" or "Meeting Name | Microsoft Teams"
	if idx := strings.Index(title, " | "); idx > 0 {
		context := strings.TrimSpace(title[:idx])

		// Check if it's a channel or direct message
		if strings.HasPrefix(context, "#") || strings.Contains(context, "General") {
			info.ChatChannel = context
		} else if strings.Contains(title, "Meeting") || strings.Contains(title, "Call") {
			info.ActivityType = "meeting"
			info.ContentSummary = context
		} else {
			info.ChatContact = context
		}
	}
}

func (dt *DeepTracker) extractSlackInfo(hwnd windows.HWND, title string, info *DeepTrackingInfo) {
	info.ActivityType = "messaging"

	// Slack title: "#channel | Workspace - Slack" or "Person | Workspace - Slack"
	parts := strings.Split(title, " | ")
	if len(parts) >= 2 {
		context := strings.TrimSpace(parts[0])
		if strings.HasPrefix(context, "#") {
			info.ChatChannel = context
		} else {
			info.ChatContact = context
		}
	}
}

func (dt *DeepTracker) extractDiscordInfo(hwnd windows.HWND, title string, info *DeepTrackingInfo) {
	info.ActivityType = "messaging"

	// Discord title: "Server | #channel - Discord" or "@user - Discord"
	if strings.Contains(title, " | ") {
		parts := strings.Split(title, " | ")
		if len(parts) >= 2 {
			// Server name first, then channel
			info.ChatChannel = strings.TrimSpace(parts[1])
			if strings.Contains(info.ChatChannel, " - Discord") {
				info.ChatChannel = strings.Replace(info.ChatChannel, " - Discord", "", 1)
			}
		}
	} else if strings.HasPrefix(title, "@") {
		// DM format
		if idx := strings.Index(title, " - Discord"); idx > 0 {
			info.ChatContact = strings.TrimSpace(title[:idx])
		}
	}
}

func (dt *DeepTracker) extractWhatsAppInfo(hwnd windows.HWND, title string, info *DeepTrackingInfo) {
	info.ActivityType = "messaging"

	// WhatsApp title: "Contact Name" or "Contact Name - WhatsApp"
	contact := title
	if idx := strings.Index(title, " - WhatsApp"); idx > 0 {
		contact = strings.TrimSpace(title[:idx])
	}

	if contact != "" && contact != "WhatsApp" {
		info.ChatContact = contact
	}
}

func (dt *DeepTracker) extractTelegramInfo(hwnd windows.HWND, title string, info *DeepTrackingInfo) {
	info.ActivityType = "messaging"

	// Telegram title: "Contact/Group Name – Telegram" or "Contact/Group Name - Telegram"
	for _, sep := range []string{" – ", " - "} {
		if idx := strings.Index(title, sep+"Telegram"); idx > 0 {
			info.ChatContact = strings.TrimSpace(title[:idx])
			break
		}
	}
}

// ============================================================
// Other App Extractors
// ============================================================

func (dt *DeepTracker) extractFigmaInfo(hwnd windows.HWND, title string, info *DeepTrackingInfo) {
	info.ActivityType = "designing"

	// Figma title: "Design Name – Figma" or "Design Name - Figma"
	for _, sep := range []string{" – ", " - "} {
		if idx := strings.Index(title, sep+"Figma"); idx > 0 {
			info.DocumentName = strings.TrimSpace(title[:idx])
			break
		}
	}
}

func (dt *DeepTracker) extractNotepadInfo(hwnd windows.HWND, title string, info *DeepTrackingInfo) {
	info.ActivityType = "editing_text"

	// Notepad title: "filename - Notepad" or "*filename - Notepad" (unsaved)
	if idx := strings.Index(title, " - Notepad"); idx > 0 {
		filename := strings.TrimSpace(title[:idx])
		filename = strings.TrimPrefix(filename, "*") // Remove unsaved indicator
		info.FileName = filename
		info.DocumentName = filename
	}
}

func (dt *DeepTracker) extractExplorerInfo(hwnd windows.HWND, title string, info *DeepTrackingInfo) {
	info.ActivityType = "file_management"

	// Explorer title is usually the folder name or path
	if title != "" && title != "File Explorer" {
		info.DocumentPath = title
	}
}

func (dt *DeepTracker) extractGenericInfo(title string, info *DeepTrackingInfo) {
	// Try to extract useful information from generic window titles

	// Common pattern: "Document - Application"
	if idx := strings.LastIndex(title, " - "); idx > 0 {
		possibleDoc := strings.TrimSpace(title[:idx])

		// Check if it looks like a filename
		if strings.Contains(possibleDoc, ".") {
			info.FileName = possibleDoc
		} else {
			info.DocumentName = possibleDoc
		}
	}
}

// ============================================================
// Privacy and Utility Functions
// ============================================================

func (dt *DeepTracker) applyPrivacyRedaction(info *DeepTrackingInfo) {
	// Redact email addresses
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)

	if info.EmailSender != "" {
		info.EmailSender = emailRegex.ReplaceAllString(info.EmailSender, "[email]")
	}
	if info.ChatContact != "" {
		info.ChatContact = emailRegex.ReplaceAllString(info.ChatContact, "[email]")
	}

	// Truncate long subjects/titles
	if len(info.EmailSubject) > 50 {
		info.EmailSubject = info.EmailSubject[:47] + "..."
	}
	if len(info.ContentSummary) > 100 {
		info.ContentSummary = info.ContentSummary[:97] + "..."
	}

	// Redact URLs (keep only domain)
	if info.BrowserURL != "" && info.BrowserDomain != "" {
		info.BrowserURL = "https://" + info.BrowserDomain + "/[redacted]"
	}
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// getFileExtension extracts the file extension from a filename
func getFileExtension(filename string) string {
	if idx := strings.LastIndex(filename, "."); idx >= 0 && idx < len(filename)-1 {
		return filename[idx+1:]
	}
	return ""
}

// ============================================================
// Windows UI Automation Helpers
// ============================================================

var (
	procCoInitializeEx = ole32.NewProc("CoInitializeEx")
)

const (
	COINIT_APARTMENTTHREADED = 0x2
	COINIT_MULTITHREADED     = 0x0
)

// initCOM initializes COM for UI Automation
func initCOM() error {
	ret, _, _ := procCoInitializeEx.Call(0, COINIT_MULTITHREADED)
	if ret != 0 && ret != 1 { // S_OK or S_FALSE (already initialized)
		return fmt.Errorf("CoInitializeEx failed: %x", ret)
	}
	return nil
}

// GetUIElementText attempts to get text from a UI element using accessibility APIs
func GetUIElementText(hwnd windows.HWND, maxLength int) string {
	// First try WM_GETTEXT
	text := getWindowTextFromHwnd(hwnd)
	if text != "" {
		if len(text) > maxLength {
			return text[:maxLength]
		}
		return text
	}

	// Could add UI Automation fallback here
	return ""
}

// FindEditControl finds an edit control within a window
func FindEditControl(hwnd windows.HWND) windows.HWND {
	// Common edit control class names
	editClasses := []string{
		"Edit",
		"RichEdit",
		"RichEdit20A",
		"RichEdit20W",
		"RichEdit20WPT",
		"RICHEDIT50W",
		"TextBox",
	}

	for _, className := range editClasses {
		child := findChildWindowByClass(hwnd, className, 5)
		if child != 0 {
			return child
		}
	}

	return 0
}

// GetEditControlText gets text from an edit control
func GetEditControlText(hwnd windows.HWND) string {
	editHwnd := FindEditControl(hwnd)
	if editHwnd == 0 {
		return ""
	}

	// Get text length
	length, _, _ := procSendMessage.Call(
		uintptr(editHwnd),
		WM_GETTEXTLENGTH,
		0,
		0,
	)

	if length == 0 || length > 10000 { // Limit to 10KB
		return ""
	}

	// Get text
	buf := make([]uint16, length+1)
	procSendMessage.Call(
		uintptr(editHwnd),
		WM_GETTEXT,
		uintptr(length+1),
		uintptr(unsafe.Pointer(&buf[0])),
	)

	return syscall.UTF16ToString(buf)
}

// GenerateContentSummary creates a human-readable summary of the tracking info
func (info *DeepTrackingInfo) GenerateContentSummary() string {
	var parts []string

	switch info.ActivityType {
	case "email", "composing_email", "webmail":
		if info.EmailSubject != "" {
			parts = append(parts, fmt.Sprintf("Email: %s", info.EmailSubject))
		}
		if info.EmailSender != "" {
			parts = append(parts, fmt.Sprintf("from %s", info.EmailSender))
		}

	case "messaging":
		if info.ChatChannel != "" {
			parts = append(parts, fmt.Sprintf("Chat in %s", info.ChatChannel))
		} else if info.ChatContact != "" {
			parts = append(parts, fmt.Sprintf("Chat with %s", info.ChatContact))
		}

	case "coding":
		if info.FileName != "" {
			parts = append(parts, fmt.Sprintf("Editing %s", info.FileName))
		}
		if info.ProjectName != "" {
			parts = append(parts, fmt.Sprintf("in %s", info.ProjectName))
		}

	case "editing_document", "editing_spreadsheet", "editing_presentation":
		if info.DocumentName != "" {
			parts = append(parts, fmt.Sprintf("Editing: %s", info.DocumentName))
		}

	case "browsing":
		if info.PageTitle != "" {
			parts = append(parts, info.PageTitle)
		}
		if info.BrowserDomain != "" {
			parts = append(parts, fmt.Sprintf("on %s", info.BrowserDomain))
		}

	case "designing":
		if info.DocumentName != "" {
			parts = append(parts, fmt.Sprintf("Designing: %s", info.DocumentName))
		}

	case "meeting":
		if info.ContentSummary != "" {
			parts = append(parts, fmt.Sprintf("In meeting: %s", info.ContentSummary))
		}

	default:
		if info.DocumentName != "" {
			parts = append(parts, info.DocumentName)
		} else if info.FileName != "" {
			parts = append(parts, info.FileName)
		} else if info.PageTitle != "" {
			parts = append(parts, info.PageTitle)
		}
	}

	if len(parts) == 0 {
		return info.WindowTitle
	}

	return strings.Join(parts, " ")
}

// ToMetadataJSON converts the deep tracking info to a JSON-compatible map
func (info *DeepTrackingInfo) ToMetadataMap() map[string]interface{} {
	meta := make(map[string]interface{})

	if info.ActivityType != "" {
		meta["activity_type"] = info.ActivityType
	}
	if info.DocumentName != "" {
		meta["document_name"] = info.DocumentName
	}
	if info.FileName != "" {
		meta["file_name"] = info.FileName
	}
	if info.ProjectName != "" {
		meta["project_name"] = info.ProjectName
	}
	if info.EmailSubject != "" {
		meta["email_subject"] = info.EmailSubject
	}
	if info.EmailSender != "" {
		meta["email_sender"] = info.EmailSender
	}
	if info.ChatContact != "" {
		meta["chat_contact"] = info.ChatContact
	}
	if info.ChatChannel != "" {
		meta["chat_channel"] = info.ChatChannel
	}
	if info.BrowserDomain != "" {
		meta["browser_domain"] = info.BrowserDomain
	}
	if info.PageTitle != "" {
		meta["page_title"] = info.PageTitle
	}

	summary := info.GenerateContentSummary()
	if summary != "" && summary != info.WindowTitle {
		meta["content_summary"] = summary
	}

	return meta
}

// DebugLog logs deep tracking info for debugging
func (info *DeepTrackingInfo) DebugLog() {
	log.Printf("DeepTracking: app=%s, type=%s, summary=%s",
		info.AppName, info.ActivityType, info.GenerateContentSummary())
}
