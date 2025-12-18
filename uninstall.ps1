# PODX Uninstaller for Windows
# Run in PowerShell as Administrator:
# iwr -useb https://raw.githubusercontent.com/dwirx/podx/main/uninstall.ps1 | iex

$ErrorActionPreference = "Stop"

$BinaryName = "podx.exe"
$InstallDir = "$env:LOCALAPPDATA\podx"
$ConfigDir = "$env:USERPROFILE\.config\podx"

Write-Host "🗑️  Uninstalling PODX..." -ForegroundColor Cyan

# Remove binary
if (Test-Path "$InstallDir\$BinaryName") {
    Remove-Item -Path "$InstallDir\$BinaryName" -Force
    Write-Host "✓ Removed binary: $InstallDir\$BinaryName" -ForegroundColor Green
} else {
    Write-Host "⚠️  Binary not found: $InstallDir\$BinaryName" -ForegroundColor Yellow
}

# Remove install directory if empty
if (Test-Path $InstallDir) {
    $items = Get-ChildItem -Path $InstallDir
    if ($items.Count -eq 0) {
        Remove-Item -Path $InstallDir -Force
        Write-Host "✓ Removed empty directory: $InstallDir" -ForegroundColor Green
    }
}

# Remove from PATH
$CurrentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($CurrentPath -like "*$InstallDir*") {
    $NewPath = ($CurrentPath.Split(';') | Where-Object { $_ -ne $InstallDir }) -join ';'
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    Write-Host "✓ Removed from PATH" -ForegroundColor Green
}

# Ask about config
if (Test-Path $ConfigDir) {
    Write-Host ""
    $response = Read-Host "Remove config directory ($ConfigDir)? [y/N]"
    if ($response -eq 'y' -or $response -eq 'Y') {
        Remove-Item -Path $ConfigDir -Recurse -Force
        Write-Host "✓ Removed config: $ConfigDir" -ForegroundColor Green
    } else {
        Write-Host "⚠️  Kept config: $ConfigDir" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "✅ PODX uninstalled successfully!" -ForegroundColor Green
