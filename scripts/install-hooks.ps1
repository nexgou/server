# scripts/install-hooks.ps1
# Configures Git to use the project's .githooks directory.
# Run once after cloning: .\scripts\install-hooks.ps1

$ErrorActionPreference = 'Stop'

$hooksDir = '.githooks'

if (-not (Test-Path -LiteralPath $hooksDir -PathType Container)) {
    Write-Error "ERROR: $hooksDir directory not found. Run this script from the repository root."
    exit 1
}

git config core.hooksPath $hooksDir

if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to configure git hooks path."
    exit 1
}

Write-Host "Git hooks installed successfully."
Write-Host "Hook path set to: $(git config core.hooksPath)"
