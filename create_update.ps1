# Create Update Package for ChronicleCore
# Usage: .\create_update.ps1 -Version "1.0.1" -Changes "Bug fixes"

param(
    [Parameter(Mandatory=$true)]
    [string]$Version,

    [Parameter(Mandatory=$true)]
    [string]$Changes,

    [switch]$IncludeUI = $false,
    [switch]$IncludeMigration = $false,
    [string]$MigrationFile = ""
)

$ErrorActionPreference = "Stop"

Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "  Creating Update Package v$Version" -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host ""

# Paths
$RootDir = $PSScriptRoot
$UpdatesDir = Join-Path $RootDir "updates"
$VersionDir = Join-Path $UpdatesDir "v$Version"
$BackendBinary = Join-Path $RootDir "apps\chroniclecore-core\chroniclecore.exe"

# Create update directory
if (Test-Path $VersionDir) {
    Write-Host "WARNING: Update v$Version already exists. Overwriting..." -ForegroundColor Yellow
    Remove-Item $VersionDir -Recurse -Force
}
New-Item -ItemType Directory -Path $VersionDir | Out-Null

# Verify backend binary exists
if (-not (Test-Path $BackendBinary)) {
    Write-Host "ERROR: Backend binary not found!" -ForegroundColor Red
    Write-Host "Please build first: cd apps\chroniclecore-core && .\build.bat" -ForegroundColor Yellow
    exit 1
}

Write-Host "[1/5] Copying backend binary..." -ForegroundColor Yellow
Copy-Item $BackendBinary (Join-Path $VersionDir "chroniclecore.exe")
$binarySize = (Get-Item $BackendBinary).Length / 1MB
Write-Host "   Binary size: $([math]::Round($binarySize, 2)) MB" -ForegroundColor Green
Write-Host ""

# Include UI if requested
if ($IncludeUI) {
    Write-Host "[2/5] Including UI assets..." -ForegroundColor Yellow
    $UIBuildDir = Join-Path $RootDir "apps\chroniclecore-ui\build"
    if (-not (Test-Path $UIBuildDir)) {
        Write-Host "   WARNING: UI build not found at $UIBuildDir" -ForegroundColor Yellow
        Write-Host "   Run: cd apps\chroniclecore-ui && npm run build" -ForegroundColor Yellow
    } else {
        $UIDestDir = Join-Path $VersionDir "web"
        Copy-Item $UIBuildDir $UIDestDir -Recurse
        Write-Host "   UI included!" -ForegroundColor Green
    }
} else {
    Write-Host "[2/5] Skipping UI (backend-only update)..." -ForegroundColor Yellow
}
Write-Host ""

# Include migration if requested
if ($IncludeMigration -and $MigrationFile) {
    Write-Host "[3/5] Including database migration..." -ForegroundColor Yellow
    if (-not (Test-Path $MigrationFile)) {
        Write-Host "   ERROR: Migration file not found: $MigrationFile" -ForegroundColor Red
        exit 1
    }
    $MigrationsDir = Join-Path $VersionDir "migrations"
    New-Item -ItemType Directory -Path $MigrationsDir | Out-Null
    Copy-Item $MigrationFile $MigrationsDir
    Write-Host "   Migration included!" -ForegroundColor Green
} else {
    Write-Host "[3/5] No database migration..." -ForegroundColor Yellow
}
Write-Host ""

# Create manifest
Write-Host "[4/5] Creating manifest..." -ForegroundColor Yellow
$manifest = @{
    version = $Version
    release_date = (Get-Date -Format "yyyy-MM-dd")
    changes = $Changes
    includes_ui = $IncludeUI
    includes_migration = $IncludeMigration
    binary_size_mb = [math]::Round($binarySize, 2)
}
$manifest | ConvertTo-Json | Set-Content (Join-Path $VersionDir "manifest.json")
Write-Host "   Manifest created!" -ForegroundColor Green
Write-Host ""

# Create changelog
Write-Host "[5/5] Creating changelog..." -ForegroundColor Yellow
$changelog = @"
# ChronicleCore v$Version

**Release Date:** $(Get-Date -Format "yyyy-MM-dd")

## Changes

$Changes

## Update Instructions

1. Download chroniclecore.exe
2. Close ChronicleCore if running
3. Replace the old chroniclecore.exe in your installation folder
4. Restart ChronicleCore

## Files Included

- chroniclecore.exe ($([math]::Round($binarySize, 2)) MB)
$(if ($IncludeUI) { "- web/ (UI assets)" } else { "" })
$(if ($IncludeMigration) { "- migrations/ (Database updates)" } else { "" })

## Rollback

If this update causes issues, restore the backup:
1. Navigate to your installation folder
2. Delete chroniclecore.exe
3. Rename chroniclecore.exe.backup to chroniclecore.exe
4. Restart
"@
$changelog | Set-Content (Join-Path $VersionDir "CHANGELOG.md")
Write-Host "   Changelog created!" -ForegroundColor Green
Write-Host ""

# Calculate SHA256 for security
Write-Host "Calculating SHA256 checksum..." -ForegroundColor Yellow
$hash = Get-FileHash (Join-Path $VersionDir "chroniclecore.exe") -Algorithm SHA256
Write-Host "   SHA256: $($hash.Hash)" -ForegroundColor Gray
Write-Host ""

# Create update.json for distribution
Write-Host "Creating update.json..." -ForegroundColor Yellow
$updateJson = @{
    latest_version = $Version
    release_date = (Get-Date -Format "yyyy-MM-dd")
    download_url = "https://YOUR-HOSTING-URL/updates/v$Version/chroniclecore.exe"
    changelog = $Changes
    mandatory = $false
    min_version = "1.0.0"
    sha256 = $hash.Hash
    size_mb = [math]::Round($binarySize, 2)
    includes_ui = $IncludeUI
    includes_migration = $IncludeMigration
}
$updateJson | ConvertTo-Json | Set-Content (Join-Path $UpdatesDir "update.json")
Write-Host "   update.json created!" -ForegroundColor Green
Write-Host ""

# Create ZIP for easy distribution
Write-Host "Creating distribution ZIP..." -ForegroundColor Yellow
$zipPath = Join-Path $UpdatesDir "ChronicleCore_Update_v$Version.zip"
Compress-Archive -Path $VersionDir -DestinationPath $zipPath -Force
$zipSize = (Get-Item $zipPath).Length / 1MB
Write-Host "   ZIP created: $([math]::Round($zipSize, 2)) MB" -ForegroundColor Green
Write-Host ""

# Success summary
Write-Host "============================================================" -ForegroundColor Green
Write-Host "  UPDATE PACKAGE CREATED!" -ForegroundColor Green
Write-Host "============================================================" -ForegroundColor Green
Write-Host ""
Write-Host "Version:" -ForegroundColor White
Write-Host "  $Version" -ForegroundColor Cyan
Write-Host ""
Write-Host "Changes:" -ForegroundColor White
Write-Host "  $Changes" -ForegroundColor Cyan
Write-Host ""
Write-Host "Package location:" -ForegroundColor White
Write-Host "  $VersionDir" -ForegroundColor Cyan
Write-Host ""
Write-Host "Distribution ZIP:" -ForegroundColor White
Write-Host "  $zipPath" -ForegroundColor Cyan
Write-Host ""
Write-Host "SHA256:" -ForegroundColor White
Write-Host "  $($hash.Hash)" -ForegroundColor Gray
Write-Host ""
Write-Host "============================================================" -ForegroundColor White
Write-Host "  NEXT STEPS" -ForegroundColor White
Write-Host "============================================================" -ForegroundColor White
Write-Host ""
Write-Host "1. Upload to hosting:" -ForegroundColor Yellow
Write-Host "   - Upload: updates/v$Version/chroniclecore.exe" -ForegroundColor Gray
Write-Host "   - Upload: updates/update.json" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Update update.json with your hosting URL:" -ForegroundColor Yellow
Write-Host "   - Edit: updates/update.json" -ForegroundColor Gray
Write-Host "   - Replace: https://YOUR-HOSTING-URL/..." -ForegroundColor Gray
Write-Host ""
Write-Host "3. Test update:" -ForegroundColor Yellow
Write-Host "   - Run: .\test_update.ps1" -ForegroundColor Gray
Write-Host ""
Write-Host "4. Notify users (optional):" -ForegroundColor Yellow
Write-Host "   - They'll auto-check within 24h" -ForegroundColor Gray
Write-Host "   - Or send email with changelog" -ForegroundColor Gray
Write-Host ""
Write-Host "For detailed instructions, see: UPDATE_DISTRIBUTION.md" -ForegroundColor White
Write-Host ""
