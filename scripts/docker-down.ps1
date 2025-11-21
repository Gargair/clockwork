# Stop Docker Compose services
# Usage: scripts/docker-down.ps1 [-RemoveVolumes]

param(
    [switch]$RemoveVolumes = $false
)

$ErrorActionPreference = "Stop"

if ($RemoveVolumes) {
    Write-Warning "WARNING: This will remove all volumes and delete all data!"
    $confirmation = Read-Host "Are you sure you want to continue? (yes/no)"
    if ($confirmation -ne "yes") {
        Write-Host "Aborted." -ForegroundColor Yellow
        exit 0
    }
    Write-Host "Stopping services and removing volumes..." -ForegroundColor Cyan
    docker compose down -v
} else {
    Write-Host "Stopping services..." -ForegroundColor Cyan
    docker compose down
}

if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to stop services"
    exit 1
}

Write-Host "Services stopped." -ForegroundColor Green

