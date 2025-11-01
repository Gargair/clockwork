# Database migrations (goose)

This directory contains SQL migrations managed by `pressly/goose` v3.

## Naming convention
- Files are named: `YYYYMMDD_NNNN_description.sql`
  - `YYYYMMDD` is the UTC date.
  - `NNNN` is a 4-digit sequence starting at `0001` for the day.
  - Example: `20250115_0001_init.sql`, `20250115_0002_indexes.sql`.

## Common commands (PowerShell)
Set your database URL (example for local Postgres):

```powershell
$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/clockwork?sslmode=disable"
```

- Create a new migration (SQL-first):
```powershell
go run github.com/pressly/goose/v3/cmd/goose@latest create <name> sql -dir ./server/migrations
```

- Apply all pending migrations:
```powershell
go run github.com/pressly/goose/v3/cmd/goose@latest -dir ./server/migrations postgres "$env:DATABASE_URL" up
```

- Roll back the last migration (use with care):
```powershell
go run github.com/pressly/goose/v3/cmd/goose@latest -dir ./server/migrations postgres "$env:DATABASE_URL" down
```

- Show migration status:
```powershell
go run github.com/pressly/goose/v3/cmd/goose@latest -dir ./server/migrations postgres "$env:DATABASE_URL" status
```

## Notes
- Do not modify applied migrations; create a new migration instead.
- Step 3 will introduce the initial schema; Step 4 adds indexes.
- Step 11 may enable auto-running migrations at server startup (config-gated).

