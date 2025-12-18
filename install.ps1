# PODX Installer for Windows
# Run in PowerShell as Administrator:
# iwr -useb https://raw.githubusercontent.com/dwirx/podx/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo = "dwirx/podx"
$BinaryName = "podx.exe"
$InstallDir = "$env:LOCALAPPDATA\podx"

Write-Host "🔐 Installing PODX..." -ForegroundColor Cyan

# Create install directory
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Get latest release
$ReleasesUrl = "https://api.github.com/repos/$Repo/releases/latest"
$Release = Invoke-RestMethod -Uri $ReleasesUrl -UseBasicParsing

$Asset = $Release.assets | Where-Object { $_.name -like "*windows-amd64*" }

if (!$Asset) {
    Write-Host "❌ Could not find Windows release" -ForegroundColor Red
    exit 1
}

$DownloadUrl = $Asset.browser_download_url
Write-Host "⬇️  Downloading from: $DownloadUrl" -ForegroundColor Yellow

# Download
$TmpFile = "$env:TEMP\podx-download.exe"
Invoke-WebRequest -Uri $DownloadUrl -OutFile $TmpFile -UseBasicParsing

# Install
Move-Item -Path $TmpFile -Destination "$InstallDir\$BinaryName" -Force

# Add to PATH if not already
$CurrentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($CurrentPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$CurrentPath;$InstallDir", "User")
    $env:PATH = "$env:PATH;$InstallDir"
    Write-Host "📁 Added $InstallDir to PATH" -ForegroundColor Green
}

Write-Host ""
Write-Host "✅ PODX installed successfully!" -ForegroundColor Green
Write-Host "   Location: $InstallDir\$BinaryName" -ForegroundColor Gray
Write-Host ""
Write-Host "🚀 Quick Start:" -ForegroundColor Cyan
Write-Host "   podx keygen -t age    # Generate key"
Write-Host "   podx init             # Init project"
Write-Host "   podx encrypt-all      # Encrypt secrets"
Write-Host ""
Write-Host "⚠️  Restart your terminal for PATH changes to take effect" -ForegroundColor Yellow
