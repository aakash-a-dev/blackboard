# Install mockapi on Windows — downloads the latest release binary and adds it to PATH.
# Usage: irm https://raw.githubusercontent.com/aakash-a-dev/blackboard/main/blackboard-cli/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo    = "aakash-a-dev/blackboard"
$Binary  = "mockapi.exe"
$InstallDir = "$env:LOCALAPPDATA\mockapi"

# ── fetch latest version ────────────────────────────────────────────────────
Write-Host "Fetching latest mockapi release..."
$release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
$version = $release.tag_name

if (-not $version) {
  Write-Error "Could not determine latest version."
  exit 1
}

Write-Host "Installing mockapi $version (windows/amd64)..."

# ── download ────────────────────────────────────────────────────────────────
$archive = "mockapi_${version}_windows_amd64.zip"
$url     = "https://github.com/$Repo/releases/download/$version/$archive"
$tmp     = Join-Path $env:TEMP "mockapi-install"

New-Item -ItemType Directory -Force $tmp | Out-Null
$zipPath = Join-Path $tmp $archive

Invoke-WebRequest -Uri $url -OutFile $zipPath -UseBasicParsing

# ── extract ─────────────────────────────────────────────────────────────────
Expand-Archive -Path $zipPath -DestinationPath $tmp -Force

# ── install ──────────────────────────────────────────────────────────────────
New-Item -ItemType Directory -Force $InstallDir | Out-Null
Copy-Item (Join-Path $tmp $Binary) (Join-Path $InstallDir $Binary) -Force

# ── add to PATH (user-scoped, permanent) ────────────────────────────────────
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath -notlike "*$InstallDir*") {
  [Environment]::SetEnvironmentVariable("PATH", "$userPath;$InstallDir", "User")
  Write-Host "  Added $InstallDir to your PATH."
  Write-Host "  Restart your terminal for PATH to take effect."
}

# ── cleanup ──────────────────────────────────────────────────────────────────
Remove-Item $tmp -Recurse -Force

Write-Host ""
Write-Host "  mockapi $version installed → $InstallDir\$Binary"
Write-Host "  Run: mockapi init"
