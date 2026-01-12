# PODX Installer for Windows
# Run in PowerShell (Admin recommended):
#   iwr -useb https://raw.githubusercontent.com/dwirx/podx/main/install.ps1 | iex
#   iwr -useb https://raw.githubusercontent.com/dwirx/podx/main/install.ps1 | iex -Args "--beta"

param(
    [switch]$Beta,
    [string]$Version,
    [string]$InstallDir
)

$ErrorActionPreference = "Stop"

$Repo = "dwirx/podx"
$BinaryName = "podx.exe"

# Default install directory
if (-not $InstallDir) {
    $InstallDir = "$env:LOCALAPPDATA\podx"
}

# Detect architecture
$Arch = if ([Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    Write-Host "❌ 32-bit Windows is not supported" -ForegroundColor Red
    exit 1
}

$AssetName = "podx-windows-${Arch}.exe"

Write-Host ""
Write-Host "🔐 PODX Installer" -ForegroundColor Cyan
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray
Write-Host "   Platform: windows/$Arch"
Write-Host "   Install:  $InstallDir"

if ($Beta) {
    Write-Host "   Channel:  BETA (development)" -ForegroundColor Yellow
} else {
    Write-Host "   Channel:  Stable"
}
Write-Host ""

# Create install directory
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Determine release URL
if ($Version) {
    $ReleaseUrl = "https://api.github.com/repos/$Repo/releases/tags/$Version"
    Write-Host "📦 Fetching version: $Version" -ForegroundColor Cyan
} elseif ($Beta) {
    $ReleaseUrl = "https://api.github.com/repos/$Repo/releases/tags/beta"
    Write-Host "📦 Fetching beta release..." -ForegroundColor Yellow
} else {
    $ReleaseUrl = "https://api.github.com/repos/$Repo/releases/latest"
    Write-Host "📦 Fetching latest stable release..." -ForegroundColor Cyan
}

# Get release info
try {
    $Release = Invoke-RestMethod -Uri $ReleaseUrl -UseBasicParsing
} catch {
    Write-Host "❌ Failed to fetch release info" -ForegroundColor Red
    if ($Beta) {
        Write-Host "   Beta release may not exist yet. Try stable version." -ForegroundColor Gray
    }
    exit 1
}

# Find asset
$Asset = $Release.assets | Where-Object { $_.name -like "*$AssetName*" -or $_.name -eq $AssetName }

if (!$Asset) {
    Write-Host "❌ Could not find release for $AssetName" -ForegroundColor Red
    Write-Host "   Available assets:" -ForegroundColor Gray
    $Release.assets | ForEach-Object { Write-Host "   - $($_.name)" }
    exit 1
}

$DownloadUrl = $Asset.browser_download_url
$TagName = $Release.tag_name

Write-Host "   Version: $TagName" -ForegroundColor Green
Write-Host ""
Write-Host "⬇️  Downloading..." -ForegroundColor Cyan
Write-Host "   URL: $DownloadUrl" -ForegroundColor Gray

# Download
$TmpFile = "$env:TEMP\podx-download.exe"
try {
    $ProgressPreference = 'SilentlyContinue'  # Faster download
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TmpFile -UseBasicParsing
    $ProgressPreference = 'Continue'
} catch {
    Write-Host "❌ Download failed: $_" -ForegroundColor Red
    exit 1
}

# Verify download
if (!(Test-Path $TmpFile) -or (Get-Item $TmpFile).Length -eq 0) {
    Write-Host "❌ Download failed or file is empty" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "📁 Installing to $InstallDir..." -ForegroundColor Cyan

# Install
Move-Item -Path $TmpFile -Destination "$InstallDir\$BinaryName" -Force

# Verify installation
if (!(Test-Path "$InstallDir\$BinaryName")) {
    Write-Host "❌ Installation failed" -ForegroundColor Red
    exit 1
}

# Add to PATH if not already
$CurrentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($CurrentPath -notlike "*$InstallDir*") {
    try {
        [Environment]::SetEnvironmentVariable("PATH", "$CurrentPath;$InstallDir", "User")
        $env:PATH = "$env:PATH;$InstallDir"
        Write-Host "📁 Added $InstallDir to PATH" -ForegroundColor Green
    } catch {
        Write-Host "⚠️  Could not add to PATH automatically" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray
Write-Host "✅ PODX installed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "   Location: $InstallDir\$BinaryName" -ForegroundColor Gray

# Try to get version
try {
    $InstalledVersion = & "$InstallDir\$BinaryName" version 2>$null | Select-Object -First 1
    Write-Host "   Version:  $InstalledVersion" -ForegroundColor Gray
} catch {
    Write-Host "   Version:  $TagName" -ForegroundColor Gray
}

Write-Host ""
Write-Host "🚀 Quick Start:" -ForegroundColor Cyan
Write-Host "   podx                  # Open interactive TUI"
Write-Host "   podx keygen -t age    # Generate encryption key"
Write-Host "   podx init             # Initialize project"
Write-Host "   podx encrypt-all      # Encrypt all secrets"
Write-Host ""

if ($Beta) {
    Write-Host "⚠️  You installed a BETA version. Report bugs at:" -ForegroundColor Yellow
    Write-Host "   https://github.com/$Repo/issues"
    Write-Host ""
}

Write-Host "⚠️  Restart your terminal for PATH changes to take effect" -ForegroundColor Yellow
Write-Host ""

# Optional: Create desktop shortcut
$CreateShortcut = Read-Host "Create desktop shortcut? (y/N)"
if ($CreateShortcut -eq "y" -or $CreateShortcut -eq "Y") {
    try {
        $WshShell = New-Object -ComObject WScript.Shell
        $Shortcut = $WshShell.CreateShortcut("$env:USERPROFILE\Desktop\PODX.lnk")
        $Shortcut.TargetPath = "cmd.exe"
        $Shortcut.Arguments = "/k `"$InstallDir\$BinaryName`""
        $Shortcut.WorkingDirectory = "$env:USERPROFILE"
        $Shortcut.Description = "PODX - Secure Encryption CLI"
        $Shortcut.Save()
        Write-Host "✅ Desktop shortcut created" -ForegroundColor Green
    } catch {
        Write-Host "⚠️  Could not create desktop shortcut" -ForegroundColor Yellow
    }
}
