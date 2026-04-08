param(
    [switch]$SkipFrontendInstall
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

Write-Host "Starting local backend (Docker)..." -ForegroundColor Cyan
docker compose up -d --build

Write-Host "Preparing local frontend..." -ForegroundColor Cyan
Set-Location (Join-Path $scriptDir "frontend")

if (-not $SkipFrontendInstall) {
    npm install
}

Write-Host "Launching frontend dev server on http://localhost:5173 ..." -ForegroundColor Green
npm run dev
