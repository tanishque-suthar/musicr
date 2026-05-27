# musicr installation script for Windows PowerShell
# Usage: iex (iwr https://raw.githubusercontent.com/yourusername/musicr/main/install.ps1).Content

param(
    [string]$Version = "latest",
    [string]$InstallDir = "$env:USERPROFILE\AppData\Local\Programs\musicr"
)

$ErrorActionPreference = "Stop"

Write-Host "Downloading musicr..." -ForegroundColor Cyan

# Get the latest release if version is "latest"
if ($Version -eq "latest") {
    try {
        $releases = Invoke-RestMethod -Uri "https://api.github.com/repos/yourusername/musicr/releases"
        $latestRelease = $releases[0]
        $Version = $latestRelease.tag_name -replace '^v', ''
    } catch {
        Write-Host "Failed to fetch latest version. Using v1.0.0" -ForegroundColor Yellow
        $Version = "1.0.0"
    }
}

# Download URL
$DownloadUrl = "https://github.com/yourusername/musicr/releases/download/v${Version}/musicr_${Version}_windows_x86_64.zip"

# Create temp directory
$TempDir = New-Item -ItemType Directory -Path "$env:TEMP\musicr-install-$([System.Guid]::NewGuid())" -Force

try {
    # Download the binary
    Write-Host "Downloading from: $DownloadUrl"
    $ZipPath = Join-Path $TempDir "musicr.zip"
    
    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath
    } catch {
        Write-Host "Failed to download musicr" -ForegroundColor Red
        Write-Host "Error: $_" -ForegroundColor Red
        exit 1
    }

    # Extract zip
    Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force

    # Create installation directory
    $null = New-Item -ItemType Directory -Path $InstallDir -Force

    # Copy binary
    Copy-Item -Path (Join-Path $TempDir "musicr.exe") -Destination $InstallDir -Force

    Write-Host "✓ musicr installed to $InstallDir" -ForegroundColor Green
    Write-Host ""
    Write-Host "Installation Notes:" -ForegroundColor Cyan
    Write-Host "  1. Add to PATH (optional): [Environment]::SetEnvironmentVariable('Path', `"$InstallDir;`$env:Path`", 'User')" 
    Write-Host "  2. Install mpv:"
    Write-Host "     choco install mpv"
    Write-Host "     OR download from https://sourceforge.net/projects/mpv-player-windows/"
    Write-Host "  3. Run: musicr 'your search query'"
    Write-Host ""
    Write-Host "To add to PATH for current session:"
    Write-Host "  `$env:Path = '$InstallDir;' + `$env:Path"
    Write-Host ""
    Write-Host "For more info: https://github.com/yourusername/musicr"

} finally {
    # Cleanup
    Remove-Item -Path $TempDir -Recurse -Force
}
