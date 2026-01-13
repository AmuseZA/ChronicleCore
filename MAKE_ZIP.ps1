# Creates final ZIP for distribution

if (-not (Test-Path "deploy\ChronicleCore")) {
    Write-Host "ERROR: Package not found!" -ForegroundColor Red
    Write-Host "Run CREATE_PACKAGE.ps1 first" -ForegroundColor Yellow
    exit 1
}

$zipPath = "ChronicleCore_v1.0.0.zip"

if (Test-Path $zipPath) {
    Remove-Item $zipPath
}

Write-Host "Creating ZIP package..." -ForegroundColor Cyan
Compress-Archive -Path "deploy\ChronicleCore" -DestinationPath $zipPath

$size = (Get-Item $zipPath).Length / 1MB
Write-Host ""
Write-Host "✅ ZIP created!" -ForegroundColor Green
Write-Host ""
Write-Host "File: $zipPath" -ForegroundColor Cyan
Write-Host "Size: $([math]::Round($size, 2)) MB" -ForegroundColor Cyan
Write-Host ""
Write-Host "Ready to send to your fiancé!" -ForegroundColor Green
