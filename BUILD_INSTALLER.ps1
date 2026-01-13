# Master script - builds complete installer in one command
# Usage: .\BUILD_INSTALLER.ps1 -Version "1.0.0"

param(
    [string]$Version = "1.0.0",
    [switch]$SkipPrepare = $false
)

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "  ChronicleCore Complete Installer Builder" -ForegroundColor Cyan
Write-Host "  Version: $Version" -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "This will:" -ForegroundColor White
Write-Host "  1. Verify backend is built" -ForegroundColor Gray
Write-Host "  2. Download embedded Python" -ForegroundColor Gray
Write-Host "  3. Install all ML dependencies" -ForegroundColor Gray
Write-Host "  4. Create Inno Setup installer EXE" -ForegroundColor Gray
Write-Host ""
Write-Host "Time: ~5-10 minutes (mostly downloading)" -ForegroundColor Yellow
Write-Host ""

$continue = Read-Host "Continue? (Y/N)"
if ($continue -ne "Y" -and $continue -ne "y") {
    Write-Host "Cancelled." -ForegroundColor Yellow
    exit 0
}

Write-Host ""

# Step 1: Prepare files
if (-not $SkipPrepare) {
    Write-Host "============================================================" -ForegroundColor Cyan
    Write-Host "  Step 1/2: Preparing Files" -ForegroundColor Cyan
    Write-Host "============================================================" -ForegroundColor Cyan
    Write-Host ""

    & .\prepare_installer.ps1 -Version $Version

    if ($LASTEXITCODE -ne 0) {
        Write-Host ""
        Write-Host "ERROR: Preparation failed!" -ForegroundColor Red
        exit 1
    }

    Write-Host ""
}

# Step 2: Build installer
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "  Step 2/2: Building Installer" -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host ""

& .\build_inno_installer.ps1 -Version $Version

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "ERROR: Installer build failed!" -ForegroundColor Red
    exit 1
}

# Final summary
Write-Host ""
Write-Host "============================================================" -ForegroundColor Green
Write-Host "  COMPLETE - READY TO SEND!" -ForegroundColor Green
Write-Host "============================================================" -ForegroundColor Green
Write-Host ""
Write-Host "File: installer_output\ChronicleCore_Setup_v$Version.exe" -ForegroundColor Cyan
Write-Host ""
Write-Host "User experience:" -ForegroundColor White
Write-Host "  1. Double-click installer → 30 sec wizard" -ForegroundColor Gray
Write-Host "  2. Double-click desktop shortcut → App starts + browser opens" -ForegroundColor Gray
Write-Host "  3. Done! No Python, no commands, no hassle" -ForegroundColor Gray
Write-Host ""
