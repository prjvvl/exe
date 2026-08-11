#Requires -Version 5.1
# Installs exe from the latest GitHub release.
#
# Usage:
#   irm https://raw.githubusercontent.com/prjvvl/exe/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Owner = "prjvvl"
$Repo = "exe"
$InstallDir = Join-Path $env:LocalAppData "exe\bin"

$arch = if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq [System.Runtime.InteropServices.Architecture]::Arm64) {
    "arm64"
} else {
    "amd64"
}

Write-Host "Fetching latest release info for $Owner/$Repo..."
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Owner/$Repo/releases/latest"
$assetName = "exe_windows_$arch.zip"
$asset = $release.assets | Where-Object { $_.name -eq $assetName }

if (-not $asset) {
    Write-Error "Could not find asset '$assetName' in the latest release."
    exit 1
}

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$zipPath = Join-Path $env:TEMP "exe-install.zip"

Write-Host "Downloading $($asset.browser_download_url)..."
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $zipPath

Write-Host "Extracting to $InstallDir..."
Expand-Archive -Path $zipPath -DestinationPath $InstallDir -Force
Remove-Item $zipPath

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ([string]::IsNullOrEmpty($userPath) -or $userPath -notlike "*$InstallDir*") {
    Write-Host "Adding $InstallDir to your user PATH..."
    $newUserPath = if ([string]::IsNullOrEmpty($userPath)) { $InstallDir } else { "$userPath;$InstallDir" }
    [Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")
    $env:Path = "$env:Path;$InstallDir"
}

$hookMarker = "# exe shell hook"
$hookLine = "exe init powershell | Out-String | Invoke-Expression"

if (-not (Test-Path $PROFILE)) {
    New-Item -ItemType File -Path $PROFILE -Force | Out-Null
}

$profileContent = Get-Content -Raw -Path $PROFILE -ErrorAction SilentlyContinue
if ([string]::IsNullOrEmpty($profileContent) -or $profileContent -notlike "*$hookMarker*") {
    Add-Content -Path $PROFILE -Value "`n$hookMarker`n$hookLine"
    Write-Host "Added the exe shell hook to your PowerShell profile ($PROFILE)."
} else {
    Write-Host "exe shell hook already present in your PowerShell profile."
}

Write-Host ""
Write-Host "exe installed to $InstallDir"
Write-Host "Restart your terminal to start using it."
