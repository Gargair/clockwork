# Build Docker image for clockwork
# Usage: scripts/docker-build.ps1 [version]

param(
    [string]$Version = "latest"
)

$ErrorActionPreference = "Stop"

Write-Host "Building Docker image: clockwork:$Version" -ForegroundColor Cyan

# Build the image
docker build -t "clockwork:$Version" .

if ($LASTEXITCODE -ne 0) {
    Write-Error "Docker build failed"
    exit 1
}

Write-Host "Successfully built clockwork:$Version" -ForegroundColor Green

# Optionally tag with version if not "latest"
if ($Version -ne "latest") {
    docker tag "clockwork:$Version" "clockwork:latest"
    Write-Host "Tagged clockwork:$Version as clockwork:latest" -ForegroundColor Green
}

