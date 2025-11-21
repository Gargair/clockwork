# Start Docker Compose services
# Usage: scripts/docker-up.ps1

$ErrorActionPreference = "Stop"

Write-Host "Starting Docker Compose services..." -ForegroundColor Cyan

# Start services in detached mode
docker compose up -d

if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to start services"
    exit 1
}

Write-Host "Waiting for services to be healthy..." -ForegroundColor Cyan

# Wait for postgres to be healthy
$maxWait = 60
$waited = 0
while ($waited -lt $maxWait) {
    $postgresHealth = docker compose ps postgres --format json | ConvertFrom-Json | Select-Object -ExpandProperty Health -ErrorAction SilentlyContinue
    if ($postgresHealth -eq "healthy") {
        Write-Host "Postgres is healthy" -ForegroundColor Green
        break
    }
    Start-Sleep -Seconds 2
    $waited += 2
    Write-Host "." -NoNewline
}

Write-Host ""

# Wait for server to be healthy
$waited = 0
while ($waited -lt $maxWait) {
    $serverHealth = docker compose ps server --format json | ConvertFrom-Json | Select-Object -ExpandProperty Health -ErrorAction SilentlyContinue
    if ($serverHealth -eq "healthy") {
        Write-Host "Server is healthy" -ForegroundColor Green
        break
    }
    Start-Sleep -Seconds 2
    $waited += 2
    Write-Host "." -NoNewline
}

Write-Host ""
Write-Host "Services started. Displaying logs (Ctrl+C to exit)..." -ForegroundColor Green
Write-Host ""

# Display logs
docker compose logs -f

