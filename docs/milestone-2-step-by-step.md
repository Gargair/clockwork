## Milestone 2: Step-by-step implementation plan

 - [x] 1: Select migration tooling and add dependencies
   - [x] Choose `pressly/goose` for SQL-first migrations and simple CLI/Go API integration.
   - [x] Add dependency in `server/go.mod` and pin a recent version of `github.com/pressly/goose/v3`.
   - [x] Plan to use the CLI for day-to-day ops; optional Go runner can be added later for in-app auto-migrate in dev.

- [x] 2: Create migrations directory and baseline scripts
  - [x] Create `server/migrations/`.
  - [x] Establish naming convention: `YYYYMMDD\d{4}_*.sql` (e.g., `202501150001_init.sql`, `202501150002_indexes.sql`).
  - [x] Add a README comment at top of the directory explaining how to create/apply/roll back migrations.

- [x] 3: Define initial schema migration (tables + constraints)
  - [x] Create `YYYYMMDD0001_init.sql` implementing the tables from `docs/domain-model.md`:
    - [x] `project (id uuid pk, name text not null, description text, created_at timestamptz not null, updated_at timestamptz not null)`
    - [x] `category (id uuid pk, project_id uuid not null fk → project.id, parent_category_id uuid null fk → category.id on delete set null, name text not null, description text, created_at timestamptz not null, updated_at timestamptz not null)`
    - [x] `time_entry (id uuid pk, category_id uuid not null fk → category.id, started_at timestamptz not null, stopped_at timestamptz null, duration_seconds integer null, created_at timestamptz not null, updated_at timestamptz not null)`
  - [x] FKs: `category.project_id` (on delete restrict), `category.parent_category_id` (on delete set null), `time_entry.category_id` (on delete restrict)
  - [x] Basic audit columns default to `now()` where appropriate.

- [x] 4: Add indexes and uniqueness in a follow-up migration
  - [x] Create `YYYYMMDD0002_indexes.sql` with:
    - [x] Unique `(project_id, name)` on `category` to prevent duplicate names within a project
    - [x] Indexes: `category(project_id)`, `category(parent_category_id)`, `time_entry(category_id, started_at)`
    - [x] Optional partial index to accelerate active timer queries: `CREATE INDEX IF NOT EXISTS time_entry_active_idx ON time_entry (category_id) WHERE stopped_at IS NULL;`

- [x] 5: Add local Postgres via Docker
  - [x] Add a minimal `docker-compose.yml` at the repo root with a `postgres` service (e.g., image `postgres:16`, user/password/db from env, healthcheck, and a named volume).
  - [x] Document default connection URL (e.g., `postgres://postgres:postgres@localhost:5432/clockwork?sslmode=disable`).

- [x] 6: Implement typed configuration for database connection
  - [x] In `server/internal/config`, add a `Config` struct with `DatabaseURL string` and loader that reads `DATABASE_URL` (and validates it).
  - [x] Provide sane defaults for local dev (optional): allow `.env` support or fall back to `postgres://postgres:postgres@localhost:5432/clockwork?sslmode=disable`.

- [x] 7: Implement connection pool in `internal/db`
  - [x] In `server/internal/db`, implement a constructor (e.g., `Open(ctx context.Context, databaseURL string) (*sql.DB, error)`) using `database/sql` and `pgx` driver.
  - [x] Configure pool settings (max open/idle conns, conn max lifetime) suitable for dev.
  - [x] Add a simple `Ping`/`Health` helper.

- [x] 8: Add migration commands (developer workflow)
  - [x] Provide scripts/commands to run migrations locally using goose:
    - [x] Create: `go run github.com/pressly/goose/v3/cmd/goose@latest create init sql`
    - [x] Up: `go run github.com/pressly/goose/v3/cmd/goose@latest -dir ./server/migrations postgres "$env:DATABASE_URL" up`
    - [x] Down (careful): `go run github.com/pressly/goose/v3/cmd/goose@latest -dir ./server/migrations postgres "$env:DATABASE_URL" down`
  - [x] Optionally add script wrapper for convenience: `scripts/goose.ps1`

- [X] 9: Bring up database and apply migrations
   - [X] `docker compose up -d postgres`
   - [X] Set `DATABASE_URL` (PowerShell): `$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/clockwork?sslmode=disable"`
   - [X] Run `goose up` as above and verify it reports to the latest migration.

- [x] 10: Add a minimal smoke test (round-trip)
  - [x] Create `server/internal/db/db_smoke_test.go` flagged as an integration test (e.g., build tag `integration`).
  - [x] Test outline:
    - [x] Connect using `DATABASE_URL` and `db.Open`.
    - [x] Insert a row into `project` with a generated UUID and name.
    - [x] Read it back by ID and assert fields match; clean up by deleting the row.
  - [x] Document how to run it locally: `go test ./server/... -tags=integration` (with DB running and `DATABASE_URL` set).

- [x] 11: Wire automatic migrations on server startup (config gated)
  - [x] Add `DB_AUTO_MIGRATE` (or `AUTO_MIGRATE`) boolean to `internal/config.Config` with default `false`.
  - [x] Add `MIGRATIONS_DIR` to config with default `server/migrations` (allow override for containers/CI).
  - [x] Implement `internal/db/migrate.go` with `RunMigrations(ctx, databaseURL, migrationsDir)` using `goose`.
  - [ ] Optionally support embedded migrations via `go:embed` and `goose.WithBaseFS` to avoid filesystem path issues in containers.
  - [x] In `cmd/server/main.go`, before starting the HTTP server, if `DB_AUTO_MIGRATE` is true: run `RunMigrations`; log from/to version; fail fast on error.
  - [x] Production safety: default `DB_AUTO_MIGRATE=false`.
  - [x] Add readiness wait with backoff if DB is not yet available in local compose.

- [x] 12: Update docs and CI notes
  - [x] Add a short section to `docs/development.md` explaining how to start Postgres, set `DATABASE_URL`, and run migrations/tests.
  - [x] CI: keep DB-backed tests under the `integration` tag so baseline CI (`go test ./...`) continues to pass without a DB service; integration tests can be wired later when adding services/handlers.

- [X] 13: Acceptance checklist
   - [X] `docker compose up` starts a healthy local Postgres.
   - [X] `goose up` applies `YYYYMMDD0001` and `YYYYMMDD0002` cleanly on a fresh database.
  - [x] `go run ./...` (server build) still succeeds.
   - [X] Integration smoke test passes locally against the running DB.


