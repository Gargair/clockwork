# Development

## Prerequisites
- Go 1.22+
- Node.js 20+
- npm or pnpm
 - PostgreSQL 15+ (or Docker for local Postgres)

## Suggested tooling
- Go: standard `testing`, `httptest`, optionally `testify`
- Client: Vite + React and Vitest/Jest
- Linting/formatting: `golangci-lint`, `eslint`, `prettier`

## Local workflow
- Server
  - TDD loop: write failing test → implement → refactor
  - Run tests: `go test ./...`
  - Run server: `go run ./cmd/server` (exact path TBD)
- Client
  - Dev server: `npm run dev` (scripts TBD)
  - Build: `npm run build`
  - API base URL (Vite): set `VITE_API_BASE_URL` to point the SPA at the Go server in development. Create `client/.env.development.local` with:
    - `VITE_API_BASE_URL=http://localhost:8080`
    - Note: In production the SPA is served by the Go server; relative paths work.
  - Zod runtime validation: the client validates API responses with Zod schemas to catch contract drift early and keep types accurate at runtime.
 - Database
  - Run PostgreSQL locally (e.g., Docker) and set `DATABASE_URL`
  - Create a dev database and run migrations

## Database (local)
- Start Postgres with Docker Compose:
  - `docker compose up -d postgres`
- Default connection URL (PowerShell):
  - `$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/clockwork?sslmode=disable"`
- Change external port without editing compose (uses env var in ports mapping):
  - PowerShell: `$env:POSTGRES_PORT = 5433`
  - Then start: `docker compose up -d postgres`
  - Update URL accordingly: `$env:DATABASE_URL = "postgres://postgres:postgres@localhost:$env:POSTGRES_PORT/clockwork?sslmode=disable"`
- Stop and remove the container (data persists in the named volume):
  - `docker compose down`

## Migrations
- Wrapper script (PowerShell):
  - Create: `scripts/goose.ps1 create init sql`
  - Up: `scripts/goose.ps1 up`
  - Down: `scripts/goose.ps1 down`
  - Status: `scripts/goose.ps1 status`
- Direct commands (alternative):
  - Create: `go run github.com/pressly/goose/v3/cmd/goose@latest create <name> sql -dir ./server/migrations`
  - Up: `go run github.com/pressly/goose/v3/cmd/goose@latest -dir ./server/migrations postgres "$env:DATABASE_URL" up`
  - Down: `go run github.com/pressly/goose/v3/cmd/goose@latest -dir ./server/migrations postgres "$env:DATABASE_URL" down`

## Server configuration (environment variables)
- `DATABASE_URL` (required): Postgres connection string
- `DB_AUTO_MIGRATE` (default `false`): Run migrations on startup
- `MIGRATIONS_DIR` (default `server/migrations`): Path to SQL migrations
- `PORT` (default `8080`): HTTP port to bind
- `ENV` (default `development`): `development` or `production`
- `STATIC_DIR` (default `client/dist`): Path to built client assets (served in production)
- `ALLOWED_ORIGINS` (CSV): CORS allowed origins; defaults to `*` in development when unset

## Integration tests
- Ensure Postgres is running and `DATABASE_URL` is set (see above)
- Run integration tests (PowerShell):
  - `cd server; go test ./... -tags=integration`
 - Repository integration tests live under `server/internal/repository/postgres` and are gated with the `integration` build tag.

## Containerized Development

The application can be run entirely in Docker containers using Docker Compose, which is useful for:
- Consistent development environments
- Testing the full stack without local dependencies
- Simulating production-like deployments

### Quick Start

1. **Build the Docker image:**
   ```powershell
   scripts/docker-build.ps1
   ```
   Or manually:
   ```powershell
   docker build -t clockwork:latest .
   ```

2. **Start all services:**
   ```powershell
   scripts/docker-up.ps1
   ```
   Or manually:
   ```powershell
   docker compose up -d
   ```

3. **Access the application:**
   - Web UI: http://localhost:8080
   - API: http://localhost:8080/api
   - Health check: http://localhost:8080/healthz

4. **View logs:**
   ```powershell
   scripts/docker-logs.ps1
   # Or for a specific service:
   scripts/docker-logs.ps1 server
   ```

5. **Stop services:**
   ```powershell
   scripts/docker-down.ps1
   ```

### Environment Variable Overrides

You can override environment variables when using Docker Compose:

**PowerShell:**
```powershell
$env:SERVER_PORT = "9090"
$env:POSTGRES_PORT = "5433"
docker compose up -d
```

**Or create a `.env` file** in the repository root:
```
SERVER_PORT=9090
POSTGRES_PORT=5433
```

**Or use docker-compose.override.yml** (not tracked in git):
```yaml
services:
  server:
    environment:
      ENV: development
      ALLOWED_ORIGINS: "http://localhost:3000,http://localhost:5173"
```

### Migration Strategy

The server supports two migration strategies:

**Auto-migrate (default in docker-compose.yml):**
- Set `DB_AUTO_MIGRATE=true` in docker-compose.yml
- Migrations run automatically on server startup
- Convenient for development and testing
- **Warning:** In production, consider manual migrations for better control

**Manual migrations:**
1. Set `DB_AUTO_MIGRATE=false` in docker-compose.yml
2. Run migrations manually:
   ```powershell
   scripts/docker-migrate.ps1 up
   ```
   Or use the local goose script:
   ```powershell
   $env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/clockwork?sslmode=disable"
   scripts/goose.ps1 up
   ```

### Helper Scripts

The following PowerShell scripts are available in the `scripts/` directory:

- **`docker-build.ps1 [version]`**: Build Docker image (default: `latest`)
- **`docker-up.ps1`**: Start services, wait for health checks, and show logs
- **`docker-down.ps1 [-RemoveVolumes]`**: Stop services (optionally remove volumes)
- **`docker-logs.ps1 [service]`**: View logs for all services or a specific service
- **`docker-migrate.ps1 [action]`**: Run migrations manually (requires `DB_AUTO_MIGRATE=false`)

### Troubleshooting

**Services won't start:**
- Check if ports are already in use: `netstat -ano | findstr :8080`
- Verify Docker is running: `docker ps`
- Check logs: `scripts/docker-logs.ps1`

**Database connection issues:**
- Ensure postgres service is healthy: `docker compose ps postgres`
- Check DATABASE_URL in docker-compose.yml matches postgres service name
- Verify postgres is accessible: `docker compose exec postgres pg_isready`

**Migrations not running:**
- Check `DB_AUTO_MIGRATE` setting in docker-compose.yml
- Verify migrations directory exists: `docker compose exec server ls -la /app/migrations`
- Run migrations manually if needed: `scripts/docker-migrate.ps1 up`
