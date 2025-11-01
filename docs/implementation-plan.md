## Implementation Plan and Milestones

This plan translates the MVP scope, architecture, domain model, API, and deployment docs into concrete milestones with clear acceptance criteria and sequencing.

## Milestone 1: Repository scaffolding and tooling
- **Scope**: Establish repo layout, initialize Go and Node projects, baseline scripts and linters.
- **Deliverables**:
  - `client/` Vite + React scaffold; router and basic layout
  - `server/` with `cmd/server` and `internal/{config,http,service,repository,db,domain,clock}` skeletons
  - Lint/format: `golangci-lint`, `eslint`, `prettier`; CI for build + tests
- **Acceptance**: Both apps build locally; CI runs and passes baseline `go test ./...` and `npm run build`.

### Step-by-step implementation plan

See the detailed guide for this milestone: [milestone-1-step-by-step.md](milestone-1-step-by-step.md).

## Milestone 2: Database schema and migrations
- **Scope**: Implement PostgreSQL schema and migration tooling.
- **Deliverables**:
  - Migrations for `PROJECT`, `CATEGORY`, `TIME_ENTRY` with constraints and indexes (per domain model)
  - `internal/db` connection pool; typed config with `DATABASE_URL`
- **Acceptance**: Migrations apply cleanly; local DB up via Docker; simple round‑trip smoke test succeeds.

### Step-by-step implementation plan

See the detailed guide for this milestone: [milestone-2-step-by-step.md](milestone-2-step-by-step.md).

## Milestone 3: Server foundations
- **Scope**: HTTP server setup, middleware, static file serving, health check.
- **Deliverables**:
  - `internal/http` router, logging, recover, CORS, request-id; `/healthz`
  - `internal/clock` with interface and `SystemClock`
- **Acceptance**: Server runs; `/healthz` returns OK; serves placeholder static assets in prod mode.

## Milestone 4: Repository interfaces and Postgres implementations
- **Scope**: Define repository interfaces and implement PostgreSQL adapters.
- **Deliverables**:
  - `ProjectRepository`, `CategoryRepository`, `TimeEntryRepository` interfaces
  - Postgres implementations with CRUD/list and query helpers
- **Acceptance**: Integration tests for repositories pass (CRUD, listing, filters, constraints).

## Milestone 5: Domain services with invariants
- **Scope**: Implement `ProjectService`, `CategoryService`, `TimeTrackingService` with rules.
- **Deliverables**:
  - Enforce category tree constraints and single-active-timer invariant
  - Unit tests for invariants and edge cases
- **Acceptance**: TDD tests pass for start/stop/getActive/list and category hierarchy behaviors.

## Milestone 6: HTTP API handlers
- **Scope**: Implement REST endpoints per API sketch with validation and error handling.
- **Deliverables**:
  - `ProjectHandler`, `CategoryHandler`, `TimeHandler`
  - Request/response models and validation
- **Acceptance**: Handler integration tests cover happy paths and error cases and pass.

## Milestone 7: Client foundations
- **Scope**: Establish API client, shared types, app shell.
- **Deliverables**:
  - `client/src/api` HTTP client; typed endpoint wrappers
  - `client/src/types` for `Project`, `Category`, `TimeEntry`
  - Theme, error boundary, router, base layout
- **Acceptance**: App loads; API client can call health endpoint in dev.

## Milestone 8: Projects feature (client)
- **Scope**: CRUD for projects.
- **Deliverables**:
  - `ProjectsPage`, `ProjectForm`, `useProjects` hook
- **Acceptance**: List/create/update/delete projects works; component tests pass.

## Milestone 9: Categories feature (client)
- **Scope**: Hierarchical categories per project.
- **Deliverables**:
  - `CategoriesPage`, `CategoryTree`, `useCategories` hook
- **Acceptance**: Users manage nested categories; constraints respected; tests pass.

## Milestone 10: Time tracking feature (client)
- **Scope**: Start/Stop controls, active timer, entries list/filters.
- **Deliverables**:
  - `DashboardPage`, `TimerControls`, `EntryList`, `useActiveTimer`, `useTimeEntries`
- **Acceptance**: Single active timer enforced end‑to‑end; filtering works; tests pass.

## Milestone 11: Containerization and local orchestration
- **Scope**: Build images and compose for local dev.
- **Deliverables**:
  - Multi-stage Dockerfile bundling Go server and built SPA
  - docker-compose for server + Postgres; make targets or npm scripts
- **Acceptance**: `docker compose up` runs full app; migrations applied automatically or via command.

## Milestone 12: Kubernetes deployment (dev/prod)
- **Scope**: Manifests/Helm, configuration, secrets, migration strategy.
- **Deliverables**:
  - `deploy/` with Deployment, Service, ConfigMap/Secret; migration Job or init strategy
  - Instructions for dev (kind/minikube) and prod
- **Acceptance**: App deploys to a dev cluster; health checks pass; DB connectivity verified.

## Milestone 13: Observability and security hardening (MVP scope)
- **Scope**: Structured logs, basic metrics (optional), MVP security guardrails.
- **Deliverables**:
  - Correlated request/response logging; basic metrics endpoint if chosen
  - Input validation, strict CORS, single-user safeguards per security doc
- **Acceptance**: Logs structured and correlated; basic dashboards/alerts optionally scaffolded.

## Milestone 14: Testing maturity and docs
- **Scope**: Expand automated tests and developer documentation.
- **Deliverables**:
  - Server unit/integration tests; client component tests; smoke e2e
  - `docs/` updates (local dev, testing, deployment runbooks)
- **Acceptance**: CI runs all test suites; onboarding achievable in under 15 minutes.

## Dependencies and sequencing
- Repos/tooling → DB/migrations → Server foundations → Repositories → Services → Handlers → Client foundations → Client features → Containerization → Kubernetes → Observability/Security → Testing/Docs.
- After server endpoints stabilize, client features (Projects/Categories/Time) can proceed in parallel.

## Risks and notes
- Single-active-timer correctness depends on transactional semantics; add tests for races.
- Category tree operations can be complex; keep invariants in services, not handlers.
- Serving SPA from the Go image simplifies deployment; confirm asset pipeline early.


