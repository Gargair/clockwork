# Run database migrations manually
# Usage: scripts/docker-migrate.ps1 [up|down|status]
# Note: This is only needed if DB_AUTO_MIGRATE=false in docker-compose.yml
# Uses the local goose.ps1 script with containerized database

param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$Action = @("up")
)

$ErrorActionPreference = "Stop"

Write-Host "Running migrations..." -ForegroundColor Cyan

# Check if postgres container is running
$postgresStatus = docker compose ps postgres --format json | ConvertFrom-Json | Select-Object -ExpandProperty State -ErrorAction SilentlyContinue
if ($postgresStatus -ne "running") {
    Write-Error "Postgres container is not running. Start it first with: scripts/docker-up.ps1"
    exit 1
}

# Set DATABASE_URL to point to containerized postgres
# Get the postgres port from docker compose
$postgresPort = docker compose port postgres 5432 | ForEach-Object { ($_ -split ":")[1] }
if (-not $postgresPort) {
    $postgresPort = "5432"
}

$dbUrl = "postgres://postgres:postgres@localhost:$postgresPort/clockwork?sslmode=disable"

Write-Host "Using DATABASE_URL: postgres://postgres:***@localhost:$postgresPort/clockwork?sslmode=disable" -ForegroundColor Cyan

# Set environment variable and run goose script
$env:DATABASE_URL = $dbUrl
& "$PSScriptRoot/goose.ps1" @Action

exit $LASTEXITCODE

