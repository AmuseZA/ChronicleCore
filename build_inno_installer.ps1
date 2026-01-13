# Build Inno Setup Installer
param(
    [string]$Version = "1.0.0"
)

$ErrorActionPreference = "Stop"

Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "  Building ChronicleCore Installer" -ForegroundColor Cyan
Write-Host "  Version: $Version" -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host ""

# Check if files are prepared
if (-not (Test-Path "dist_installer\chroniclecore.exe")) {
    Write-Host "ERROR: Files not prepared!" -ForegroundColor Red
    Write-Host "Run prepare_installer.ps1 first" -ForegroundColor Yellow
    exit 1
}

# Check for Inno Setup
$innoSetupPath = "C:\Users\josh\AppData\Local\Programs\Inno Setup 6\ISCC.exe"
if (-not (Test-Path $innoSetupPath)) {
    Write-Host "ERROR: Inno Setup not found!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install Inno Setup from:" -ForegroundColor Yellow
    Write-Host "https://jrsoftware.org/isdl.php" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "It's free and takes 2 minutes to install." -ForegroundColor Gray
    Write-Host ""
    pause
    exit 1
}

# Update version in ISS file
Write-Host "Updating version in installer script..." -ForegroundColor Yellow
$issFile = "chroniclecore.iss"
$issContent = Get-Content $issFile -Raw
$issContent = $issContent -replace '#define MyAppVersion ".*"', "#define MyAppVersion `"$Version`""
$issContent | Set-Content $issFile
Write-Host "   Version updated to $Version" -ForegroundColor Green
Write-Host ""

# Create output directory
if (-not (Test-Path "installer_output")) {
    New-Item -ItemType Directory -Path "installer_output" | Out-Null
}

# Build installer
Write-Host "Building installer with Inno Setup..." -ForegroundColor Yellow
Write-Host "   This may take 2-3 minutes..." -ForegroundColor Gray
& $innoSetupPath $issFile

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "ERROR: Installer build failed!" -ForegroundColor Red
    exit 1
}

# Success
$installerPath = "installer_output\ChronicleCore_Setup_v$Version.exe"
$installerSize = (Get-Item $installerPath).Length / 1MB

Write-Host ""
Write-Host "============================================================" -ForegroundColor Green
Write-Host "  INSTALLER BUILT SUCCESSFULLY!" -ForegroundColor Green
Write-Host "============================================================" -ForegroundColor Green
Write-Host ""
Write-Host "Installer:" -ForegroundColor White
Write-Host "  $installerPath" -ForegroundColor Cyan
Write-Host ""
Write-Host "Size:" -ForegroundColor White
Write-Host "  $([math]::Round($installerSize, 2)) MB" -ForegroundColor Cyan
Write-Host ""
Write-Host "What's included:" -ForegroundColor White
Write-Host "  [+] Backend binary (chroniclecore.exe)" -ForegroundColor Green
Write-Host "  [+] Embedded Python 3.11.9" -ForegroundColor Green
Write-Host "  [+] All ML dependencies pre-installed" -ForegroundColor Green
Write-Host "  [+] ML sidecar source code" -ForegroundColor Green
Write-Host "  [+] Database schema" -ForegroundColor Green
Write-Host "  [+] Launcher script" -ForegroundColor Green
Write-Host "  [+] Desktop shortcut" -ForegroundColor Green
Write-Host ""
Write-Host "Ready to distribute." -ForegroundColor White
Write-Host ""
