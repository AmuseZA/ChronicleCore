package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// SystemHandler handles system-level endpoints
type SystemHandler struct{}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

// LocaleResponse contains system locale and currency information
type LocaleResponse struct {
	Country      string `json:"country"`       // e.g., "ZA", "US"
	Locale       string `json:"locale"`        // e.g., "en-ZA", "en-US"
	CurrencyCode string `json:"currency_code"` // ISO 4217 code, e.g., "ZAR", "USD"
}

// GetLocale detects system locale and returns currency information
func (h *SystemHandler) GetLocale(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Try to detect locale from environment variables
	locale := detectLocale()
	country := extractCountry(locale)
	currencyCode := getCurrencyForCountry(country)

	response := LocaleResponse{
		Country:      country,
		Locale:       locale,
		CurrencyCode: currencyCode,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// detectLocale attempts to detect the system locale
func detectLocale() string {
	// Try various environment variables in order of preference
	localeVars := []string{
		"LC_ALL",
		"LC_MONETARY",
		"LANG",
	}

	for _, varName := range localeVars {
		if locale := os.Getenv(varName); locale != "" {
			// Clean up locale (remove encoding, e.g., "en_US.UTF-8" -> "en_US")
			if idx := strings.Index(locale, "."); idx != -1 {
				locale = locale[:idx]
			}
			// Convert underscore to hyphen (en_US -> en-US)
			locale = strings.ReplaceAll(locale, "_", "-")
			return locale
		}
	}

	// Default to South African locale (primary user base)
	return "en-ZA"
}

// extractCountry extracts the country code from locale
func extractCountry(locale string) string {
	// Locale format: language-COUNTRY (e.g., "en-ZA", "en-US")
	parts := strings.Split(locale, "-")
	if len(parts) >= 2 {
		return strings.ToUpper(parts[1])
	}

	// If no country specified, default to US
	return "US"
}

// getCurrencyForCountry maps country codes to ISO 4217 currency codes
func getCurrencyForCountry(country string) string {
	// Common country -> currency mappings
	currencyMap := map[string]string{
		"US": "USD", // United States
		"ZA": "ZAR", // South Africa
		"GB": "GBP", // United Kingdom
		"EU": "EUR", // European Union (not a real country, but common)
		"DE": "EUR", // Germany
		"FR": "EUR", // France
		"IT": "EUR", // Italy
		"ES": "EUR", // Spain
		"NL": "EUR", // Netherlands
		"BE": "EUR", // Belgium
		"AT": "EUR", // Austria
		"PT": "EUR", // Portugal
		"IE": "EUR", // Ireland
		"CA": "CAD", // Canada
		"AU": "AUD", // Australia
		"NZ": "NZD", // New Zealand
		"JP": "JPY", // Japan
		"CN": "CNY", // China
		"IN": "INR", // India
		"BR": "BRL", // Brazil
		"MX": "MXN", // Mexico
		"CH": "CHF", // Switzerland
		"SE": "SEK", // Sweden
		"NO": "NOK", // Norway
		"DK": "DKK", // Denmark
		"PL": "PLN", // Poland
		"CZ": "CZK", // Czech Republic
		"HU": "HUF", // Hungary
		"RU": "RUB", // Russia
		"TR": "TRY", // Turkey
		"KR": "KRW", // South Korea
		"SG": "SGD", // Singapore
		"HK": "HKD", // Hong Kong
		"TH": "THB", // Thailand
		"MY": "MYR", // Malaysia
		"ID": "IDR", // Indonesia
		"PH": "PHP", // Philippines
		"VN": "VND", // Vietnam
		"AR": "ARS", // Argentina
		"CL": "CLP", // Chile
		"CO": "COP", // Colombia
		"PE": "PEN", // Peru
		"IL": "ILS", // Israel
		"SA": "SAR", // Saudi Arabia
		"AE": "AED", // UAE
		"EG": "EGP", // Egypt
		"NG": "NGN", // Nigeria
		"KE": "KES", // Kenya
		"GH": "GHS", // Ghana
	}

	if currency, found := currencyMap[country]; found {
		return currency
	}

	// Default to USD if country not recognized
	return "USD"
}

// ValidateCurrencyCode validates that a currency code is a valid 3-letter uppercase ISO 4217 code
func ValidateCurrencyCode(code string) bool {
	// Must be exactly 3 characters
	if len(code) != 3 {
		return false
	}

	// Must be uppercase letters
	for _, char := range code {
		if char < 'A' || char > 'Z' {
			return false
		}
	}

	// Common ISO 4217 codes (not exhaustive, but covers major currencies)
	validCodes := map[string]bool{
		"USD": true, "EUR": true, "GBP": true, "JPY": true, "CHF": true,
		"CAD": true, "AUD": true, "NZD": true, "ZAR": true, "CNY": true,
		"INR": true, "BRL": true, "MXN": true, "SEK": true, "NOK": true,
		"DKK": true, "PLN": true, "CZK": true, "HUF": true, "RUB": true,
		"TRY": true, "KRW": true, "SGD": true, "HKD": true, "THB": true,
		"MYR": true, "IDR": true, "PHP": true, "VND": true, "ARS": true,
		"CLP": true, "COP": true, "PEN": true, "ILS": true, "SAR": true,
		"AED": true, "EGP": true, "NGN": true, "KES": true, "GHS": true,
		"PKR": true, "BDT": true, "LKR": true, "TWD": true, "UAH": true,
		"RON": true, "BGN": true, "HRK": true, "ISK": true, "CRC": true,
		"BOB": true, "PYG": true, "UYU": true, "VEF": true, "DOP": true,
		"GTQ": true, "HNL": true, "NIO": true, "PAB": true, "JMD": true,
		"TTD": true, "BZD": true, "XCD": true, "BSD": true, "BBD": true,
		"FJD": true, "PGK": true, "WST": true, "TOP": true, "SBD": true,
		"VUV": true, "BWP": true, "MUR": true, "SCR": true, "MWK": true,
		"ZMW": true, "TZS": true, "UGX": true, "RWF": true, "MGA": true,
		"ETB": true, "ERN": true, "DJF": true, "SOS": true, "KMF": true,
		"AOA": true, "MZN": true, "LSL": true, "SZL": true, "NAD": true,
		"BIF": true, "CVE": true, "GMD": true, "GNF": true, "LRD": true,
		"SLL": true, "STD": true, "XOF": true, "XAF": true, "CDF": true,
	}

	return validCodes[code]
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	UpdateAvailable bool   `json:"update_available"`
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version,omitempty"`
	ReleaseNotes    string `json:"release_notes,omitempty"`
	DownloadURL     string `json:"download_url,omitempty"`
	ReleaseURL      string `json:"release_url,omitempty"`
	PublishedAt     string `json:"published_at,omitempty"`
}

// GitHub API response structure
type githubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// Update configuration - GitHub repo for auto-updates
const (
	GitHubOwner = "AmuseZA"         // GitHub username/org
	GitHubRepo  = "ChronicleCore"   // Repo name
)

// CheckForUpdate checks GitHub releases for a newer version
func (h *SystemHandler) CheckForUpdate(w http.ResponseWriter, r *http.Request, currentVersion string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info := UpdateInfo{
		UpdateAvailable: false,
		CurrentVersion:  currentVersion,
	}

	// Fetch latest release from GitHub
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", GitHubOwner, GitHubRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
		return
	}

	// GitHub API requires User-Agent header
	req.Header.Set("User-Agent", "ChronicleCore/"+currentVersion)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to fetch GitHub releases: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("GitHub API returned status %d", resp.StatusCode)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
		return
	}

	var release githubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		log.Printf("Failed to parse GitHub response: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
		return
	}

	// Parse version from tag (remove 'v' prefix if present)
	latestVersion := strings.TrimPrefix(release.TagName, "v")

	// Compare versions
	if compareVersions(latestVersion, currentVersion) > 0 {
		info.UpdateAvailable = true
		info.LatestVersion = latestVersion
		info.ReleaseNotes = release.Body
		info.ReleaseURL = release.HTMLURL
		info.PublishedAt = release.PublishedAt

		// Find the installer asset (.exe)
		for _, asset := range release.Assets {
			if strings.HasSuffix(strings.ToLower(asset.Name), ".exe") {
				info.DownloadURL = asset.BrowserDownloadURL
				break
			}
		}

		log.Printf("Update available: %s -> %s", currentVersion, latestVersion)
	} else {
		log.Printf("No update available (current: %s, latest: %s)", currentVersion, latestVersion)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// compareVersions compares two semantic version strings
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Pad shorter version with zeros
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var num1, num2 int
		if i < len(parts1) {
			num1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			num2, _ = strconv.Atoi(parts2[i])
		}

		if num1 > num2 {
			return 1
		}
		if num1 < num2 {
			return -1
		}
	}

	return 0
}
