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

