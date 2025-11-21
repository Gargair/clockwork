# View Docker Compose logs
# Usage: scripts/docker-logs.ps1 [service]

param(
    [string]$Service = ""
)

$ErrorActionPreference = "Stop"

if ($Service) {
    Write-Host "Displaying logs for service: $Service" -ForegroundColor Cyan
    docker compose logs -f $Service
} else {
    Write-Host "Displaying logs for all services" -ForegroundColor Cyan
    docker compose logs -f
}

