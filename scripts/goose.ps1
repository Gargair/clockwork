Param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$Args
)

if (-not $Args -or $Args.Count -eq 0) {
    Write-Host "Usage: scripts/goose.ps1 <create|up|down|status|redo|reset|fix> [args...]"
    Write-Host "Examples:"
    Write-Host "  scripts/goose.ps1 create init sql"
    Write-Host "  scripts/goose.ps1 up"
    Write-Host "  scripts/goose.ps1 down"
    exit 1
}

$goose = "github.com/pressly/goose/v3/cmd/goose@latest"
$migrationsDir = "./server/migrations"

# Handle 'create' separately (no DB needed)
if ($Args[0] -eq "create") {
    $rest = @()
    if ($Args.Count -gt 1) { $rest = $Args[1..($Args.Count - 1)] }
    go run $goose create @rest -dir $migrationsDir
    exit $LASTEXITCODE
}

if (-not $env:DATABASE_URL -or [string]::IsNullOrWhiteSpace($env:DATABASE_URL)) {
    Write-Error "DATABASE_URL is not set. Example: `$env:DATABASE_URL = 'postgres://postgres:postgres@localhost:5432/clockwork?sslmode=disable'"
    exit 1
}

go run $goose -dir $migrationsDir postgres "$env:DATABASE_URL" @Args
exit $LASTEXITCODE


