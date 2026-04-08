param(
    [string]$FrontendApiHost = "127.0.0.1",
    [string]$FrontendApiPort = "7350",
    [switch]$SkipFrontendInstall
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

Write-Host "Building and starting backend containers (production-style)..." -ForegroundColor Cyan
docker compose up -d --build

$frontendDir = Join-Path $scriptDir "frontend"
Set-Location $frontendDir

if (-not $SkipFrontendInstall) {
    Write-Host "Installing frontend dependencies..." -ForegroundColor Cyan
    npm install
}

Write-Host "Building frontend for production..." -ForegroundColor Cyan
$env:VITE_NAKAMA_HOST = $FrontendApiHost
$env:VITE_NAKAMA_PORT = $FrontendApiPort
npm run build

Write-Host "Production build completed." -ForegroundColor Green
Write-Host "Frontend artifacts: frontend/dist" -ForegroundColor Green
Write-Host "Backend services are running in Docker." -ForegroundColor Green
