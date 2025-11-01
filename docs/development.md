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
  - `go test ./server/... -tags=integration`

