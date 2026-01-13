# Simple deployment - creates ZIP with everything ready
# No Inno Setup required!

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "Creating deployment package..." -ForegroundColor Cyan
Write-Host ""

# Check if files are prepared
if (-not (Test-Path "dist_installer")) {
    Write-Host "ERROR: Files not prepared" -ForegroundColor Red
    Write-Host "Run: .\prepare_installer.ps1 -Version '1.0.0'" -ForegroundColor Yellow
    exit 1
}

# Create ZIP
$zipPath = "ChronicleCore_Portable_v1.0.0.zip"
if (Test-Path $zipPath) {
    Remove-Item $zipPath
}

Write-Host "Creating ZIP..." -ForegroundColor Yellow
Compress-Archive -Path "dist_installer\*" -DestinationPath $zipPath

$size = (Get-Item $zipPath).Length / 1MB

Write-Host ""
Write-Host "SUCCESS!" -ForegroundColor Green
Write-Host ""
Write-Host "File: $zipPath" -ForegroundColor Cyan
Write-Host "Size: $([math]::Round($size, 2)) MB" -ForegroundColor Cyan
Write-Host ""
Write-Host "What to do:" -ForegroundColor White
Write-Host "  1. Send this ZIP to your user" -ForegroundColor Gray
Write-Host "  2. They extract it anywhere" -ForegroundColor Gray
Write-Host "  3. They double-click ChronicleCore.bat" -ForegroundColor Gray
Write-Host "  4. Browser opens automatically" -ForegroundColor Gray
Write-Host ""
Write-Host "Ready to use!" -ForegroundColor Green
Write-Host ""
