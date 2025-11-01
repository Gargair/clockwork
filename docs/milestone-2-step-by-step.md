## Milestone 2: Step-by-step implementation plan

 - [x] 1: Select migration tooling and add dependencies
   - [x] Choose `pressly/goose` for SQL-first migrations and simple CLI/Go API integration.
   - [x] Add dependency in `server/go.mod` and pin a recent version of `github.com/pressly/goose/v3`.
   - [x] Plan to use the CLI for day-to-day ops; optional Go runner can be added later for in-app auto-migrate in dev.

- [x] 2: Create migrations directory and baseline scripts
  - [x] Create `server/migrations/`.
  - [x] Establish naming convention: `YYYYMMDD_\d{4}_*.sql` (e.g., `20250115_0001_init.sql`, `20250115_0002_indexes.sql`).
  - [x] Add a README comment at top of the directory explaining how to create/apply/roll back migrations.

- [x] 3: Define initial schema migration (tables + constraints)
  - [x] Create `YYYYMMDD_0001_init.sql` implementing the tables from `docs/domain-model.md`:
    - [x] `project (id uuid pk, name text not null, description text, created_at timestamptz not null, updated_at timestamptz not null)`
    - [x] `category (id uuid pk, project_id uuid not null fk → project.id, parent_category_id uuid null fk → category.id on delete set null, name text not null, description text, created_at timestamptz not null, updated_at timestamptz not null)`
    - [x] `time_entry (id uuid pk, category_id uuid not null fk → category.id, started_at timestamptz not null, stopped_at timestamptz null, duration_seconds integer null, created_at timestamptz not null, updated_at timestamptz not null)`
  - [x] FKs: `category.project_id` (on delete restrict), `category.parent_category_id` (on delete set null), `time_entry.category_id` (on delete restrict)
  - [x] Basic audit columns default to `now()` where appropriate.

- [x] 4: Add indexes and uniqueness in a follow-up migration
  - [x] Create `YYYYMMDD_0002_indexes.sql` with:
    - [x] Unique `(project_id, name)` on `category` to prevent duplicate names within a project
    - [x] Indexes: `category(project_id)`, `category(parent_category_id)`, `time_entry(category_id, started_at)`
    - [x] Optional partial index to accelerate active timer queries: `CREATE INDEX IF NOT EXISTS time_entry_active_idx ON time_entry (category_id) WHERE stopped_at IS NULL;`

- [ ] 5: Add local Postgres via Docker
   - [ ] Add a minimal `docker-compose.yml` at the repo root with a `postgres` service (e.g., image `postgres:16`, user/password/db from env, healthcheck, and a named volume).
   - [ ] Document default connection URL (e.g., `postgres://postgres:postgres@localhost:5432/clockwork?sslmode=disable`).

- [ ] 6: Implement typed configuration for database connection
   - [ ] In `server/internal/config`, add a `Config` struct with `DatabaseURL string` and loader that reads `DATABASE_URL` (and validates it).
   - [ ] Provide sane defaults for local dev (optional): allow `.env` support or fall back to `postgres://postgres:postgres@localhost:5432/clockwork?sslmode=disable`.

- [ ] 7: Implement connection pool in `internal/db`
   - [ ] In `server/internal/db`, implement a constructor (e.g., `Open(ctx context.Context, databaseURL string) (*sql.DB, error)`) using `database/sql` and `pgx` driver.
   - [ ] Configure pool settings (max open/idle conns, conn max lifetime) suitable for dev.
   - [ ] Add a simple `Ping`/`Health` helper.

- [ ] 8: Add migration commands (developer workflow)
   - [ ] Provide scripts/commands to run migrations locally using goose:
     - [ ] Create: `go run github.com/pressly/goose/v3/cmd/goose@latest create init sql`
     - [ ] Up: `go run github.com/pressly/goose/v3/cmd/goose@latest -dir ./server/migrations postgres "$env:DATABASE_URL" up`
     - [ ] Down (careful): `go run github.com/pressly/goose/v3/cmd/goose@latest -dir ./server/migrations postgres "$env:DATABASE_URL" down`
   - [ ] Optionally add `make` or `npm` scripts to wrap these commands for convenience and CI.

- [ ] 9: Bring up database and apply migrations
   - [ ] `docker compose up -d postgres`
   - [ ] Set `DATABASE_URL` (PowerShell): `$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/clockwork?sslmode=disable"`
   - [ ] Run `goose up` as above and verify it reports to the latest migration.

- [ ] 10: Add a minimal smoke test (round-trip)
   - [ ] Create `server/internal/db/db_smoke_test.go` flagged as an integration test (e.g., build tag `integration`).
   - [ ] Test outline:
     - [ ] Connect using `DATABASE_URL` and `db.Open`.
     - [ ] Insert a row into `project` with a generated UUID and name.
     - [ ] Read it back by ID and assert fields match; clean up by deleting the row.
   - [ ] Document how to run it locally: `go test ./server/... -tags=integration` (with DB running and `DATABASE_URL` set).

- [ ] 11: Wire automatic migrations on server startup (config gated)
   - [ ] Add `DB_AUTO_MIGRATE` (or `AUTO_MIGRATE`) boolean to `internal/config.Config` with default `false`.
   - [ ] Add `MIGRATIONS_DIR` to config with default `server/migrations` (allow override for containers/CI).
   - [ ] Implement `internal/db/migrate.go` with `RunMigrations(ctx, databaseURL, migrationsDir) error` using `goose` (`sql.Open`/`goose.SetDialect("postgres")`/`goose.Up`).
   - [ ] Optionally support embedded migrations via `go:embed` and `goose.WithBaseFS` to avoid filesystem path issues in containers.
   - [ ] In `cmd/server/main.go`, before starting the HTTP server, if `DB_AUTO_MIGRATE` is true: open a connection and run `RunMigrations`; log from/to version; fail fast on error.
   - [ ] Production safety: keep default `DB_AUTO_MIGRATE=false`; document enabling only for dev/staging or controlled prod rollouts.
   - [ ] Ensure retries/backoff or readiness wait if DB is not yet available in local compose.

- [ ] 12: Update docs and CI notes
   - [ ] Add a short section to `docs/development.md` explaining how to start Postgres, set `DATABASE_URL`, and run migrations/tests.
   - [ ] CI: keep DB-backed tests under the `integration` tag so baseline CI (`go test ./...`) continues to pass without a DB service; integration tests can be wired later when adding services/handlers.

- [ ] 13: Acceptance checklist
   - [ ] `docker compose up` starts a healthy local Postgres.
   - [ ] `goose up` applies `YYYYMMDD_0001` and `YYYYMMDD_0002` cleanly on a fresh database.
   - [ ] `go run ./...` (server build) still succeeds.
   - [ ] Integration smoke test passes locally against the running DB.


