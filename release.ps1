# ChronicleCore Release Script
# Usage: .\release.ps1 -Version "1.6.0" [-Notes "Release notes here"]

param(
    [Parameter(Mandatory=$true)]
    [string]$Version,

    [string]$Notes = "ChronicleCore v$Version release"
)

$ErrorActionPreference = "Stop"

$RepoOwner = "AmuseZA"
$RepoName = "ChronicleCore"
$InstallerPath = "installer_output\ChronicleCore_Setup_v$Version.exe"

Write-Host "=== ChronicleCore Release Script ===" -ForegroundColor Cyan
Write-Host "Version: $Version"
Write-Host "Installer: $InstallerPath"
Write-Host ""

# Check if installer exists
if (-not (Test-Path $InstallerPath)) {
    Write-Host "ERROR: Installer not found at $InstallerPath" -ForegroundColor Red
    Write-Host "Make sure you've built the installer first." -ForegroundColor Yellow
    exit 1
}

# Check gh auth
Write-Host "Checking GitHub CLI authentication..." -ForegroundColor Yellow
$authStatus = & "C:\Program Files\GitHub CLI\gh.exe" auth status 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Not logged in to GitHub CLI" -ForegroundColor Red
    Write-Host "Run: gh auth login" -ForegroundColor Yellow
    exit 1
}
Write-Host "Authenticated!" -ForegroundColor Green

# Create the release
Write-Host ""
Write-Host "Creating release v$Version..." -ForegroundColor Yellow

$releaseNotes = @"
## ChronicleCore v$Version

$Notes

### Installation
1. Download `ChronicleCore_Setup_v$Version.exe` below
2. Run the installer
3. Launch from Start Menu or Desktop shortcut

### What's New
- See commit history for changes
"@

try {
    & "C:\Program Files\GitHub CLI\gh.exe" release create "v$Version" `
        --repo "$RepoOwner/$RepoName" `
        --title "ChronicleCore v$Version" `
        --notes $releaseNotes `
        $InstallerPath

    if ($LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-Host "=== Release Created Successfully! ===" -ForegroundColor Green
        Write-Host "URL: https://github.com/$RepoOwner/$RepoName/releases/tag/v$Version" -ForegroundColor Cyan
    } else {
        Write-Host "Release creation failed" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
    exit 1
}
