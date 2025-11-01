## Milestone 4: Step-by-step implementation plan

 - [x] 1: Define domain entities in `server/internal/domain`
  - [x] Add explicit types reflecting the schema in `docs/domain-model.md`:
  - [x] `type Project struct { ID uuid.UUID; Name string; Description *string; CreatedAt time.Time; UpdatedAt time.Time }`
  - [x] `type Category struct { ID uuid.UUID; ProjectID uuid.UUID; ParentCategoryID *uuid.UUID; Name string; Description *string; CreatedAt time.Time; UpdatedAt time.Time }`
  - [x] `type TimeEntry struct { ID uuid.UUID; CategoryID uuid.UUID; StartedAt time.Time; StoppedAt *time.Time; DurationSeconds *int32; CreatedAt time.Time; UpdatedAt time.Time }`
  - [x] Import explicit dependencies: `github.com/google/uuid`, `time`
  - [x] Keep domain free of DB concerns; no JSON/DB tags yet (DTOs can be added later if needed)

- [x] 2: Define repository interfaces in `server/internal/repository/repository.go`
  - [x] Add package-level errors (explicit vars) for common conditions (e.g., `ErrNotFound`, `ErrDuplicate`, `ErrForeignKeyViolation`)
  - [x] `ProjectRepository` (CRUD + list)
   - [x] `Create(ctx context.Context, project domain.Project) (domain.Project, error)`
   - [x] `GetByID(ctx context.Context, id uuid.UUID) (domain.Project, error)`
   - [x] `List(ctx context.Context) ([]domain.Project, error)`
   - [x] `Update(ctx context.Context, id uuid.UUID, name string, description *string) (domain.Project, error)`
   - [x] `Delete(ctx context.Context, id uuid.UUID) error`
  - [x] `CategoryRepository`
   - [x] `Create(ctx context.Context, category domain.Category) (domain.Category, error)`
   - [x] `GetByID(ctx context.Context, id uuid.UUID) (domain.Category, error)`
   - [x] `ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Category, error)`
   - [x] `ListChildren(ctx context.Context, parentID uuid.UUID) ([]domain.Category, error)`
   - [x] `Update(ctx context.Context, id uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error)`
   - [x] `Delete(ctx context.Context, id uuid.UUID) error`
   - [x] Note: Do not allow changing `ProjectID` via repository Update (service-level invariant)
  - [x] `TimeEntryRepository`
   - [x] `Create(ctx context.Context, entry domain.TimeEntry) (domain.TimeEntry, error)`
   - [x] `GetByID(ctx context.Context, id uuid.UUID) (domain.TimeEntry, error)`
   - [x] `ListByCategory(ctx context.Context, categoryID uuid.UUID) ([]domain.TimeEntry, error)`
   - [x] `ListByCategoryAndRange(ctx context.Context, categoryID uuid.UUID, start time.Time, end time.Time) ([]domain.TimeEntry, error)`
   - [x] `FindActive(ctx context.Context) (*domain.TimeEntry, error)`
   - [x] `Stop(ctx context.Context, id uuid.UUID, stoppedAt time.Time, durationSeconds *int32) (domain.TimeEntry, error)`
  - [x] Ensure explicit parameter and return types for all methods

- [x] 3: Create Postgres adapters in `server/internal/repository/postgres`
  - [x] Layout
   - [x] `postgres/project_repository.go`
   - [x] `postgres/category_repository.go`
   - [x] `postgres/timeentry_repository.go`
   - [x] `postgres/errors.go` (translate PG errors → repository errors)
  - [x] Constructors (explicit types):
   - [x] `func NewProjectRepository(db *sql.DB) ProjectRepository`
   - [x] `func NewCategoryRepository(db *sql.DB) CategoryRepository`
   - [x] `func NewTimeEntryRepository(db *sql.DB) TimeEntryRepository`
  - [x] Implement SQL using `database/sql` with the `pgx` driver (already in use)
   - [x] Use explicit column lists in `INSERT`/`UPDATE`
   - [x] Use `RETURNING` to map rows back to domain types
   - [x] Set/update timestamps (`updated_at = now()`) in SQL
  - [x] Error mapping in `errors.go`
   - [x] Inspect `*pgconn.PgError` codes (e.g., 23505 unique violation, 23503 foreign key)
   - [x] Map to `repository.ErrDuplicate`, `repository.ErrForeignKeyViolation`, or pass through

- [x] 4: Test utilities for integration tests (tag: `integration`)
  - [x] Helper to open DB from `DATABASE_URL` (reuse `internal/db.Open`)
  - [x] `truncateAll(t *testing.T, db *sql.DB)` that deletes from `time_entry`, `category`, `project` in the right order
  - [x] Small builders for domain entities (generate UUIDs, names) to keep tests readable

 - [x] 5: ProjectRepository integration tests (`server/internal/repository/postgres/project_repository_integration_test.go`)
  - [x] `Create` then `GetByID` returns same fields
  - [x] `List` returns created projects
  - [x] `Update` changes `name`/`description` and bumps `updated_at`
  - [x] `Delete` removes the row; `GetByID` → `ErrNotFound`

 - [x] 6: CategoryRepository integration tests (`.../category_repository_integration_test.go`)
  - [x] `Create` with valid `project_id` succeeds
  - [x] Unique `(project_id, name)` enforced → duplicate insert maps to `ErrDuplicate`
  - [x] Parent/child relationships: `ListChildren` returns children; deleting parent sets children `parent_category_id` to `NULL`
  - [x] `Update` can change `name`/`description`/`parentCategoryID` but not `projectID`
  - [x] `ListByProject` returns only categories in that project

 - [x] 7: TimeEntryRepository integration tests (`.../timeentry_repository_integration_test.go`)
  - [x] `Create` inserts an active entry (with `stopped_at = NULL`)
  - [x] `FindActive` returns the created entry; when stopped, returns `nil`
  - [x] `Stop` sets `stopped_at` and `duration_seconds`; `GetByID` reflects changes
  - [x] `ListByCategory` returns entries in descending `started_at` (define and test an order)
  - [x] `ListByCategoryAndRange` filters by inclusive range

 - [x] 8: Wire repositories where helpful (non-invasive)
  - [x] Add constructors to a central wiring point if needed (e.g., later in services or handlers)
  - [x] Do not expose repositories via HTTP yet (reserved for Milestone 6)

- [ ] 9: Docs and developer workflow
  - [ ] Add a short note to `docs/development.md` about running repository integration tests (`go test ./server/... -tags=integration`)
  - [ ] Ensure `docker compose up -d postgres` and `DATABASE_URL` instructions are referenced

- [ ] 10: Acceptance checklist (map to Implementation Plan)
  - [ ] Domain entities defined with explicit types in `internal/domain`
  - [ ] Repository interfaces defined with explicit parameter/return types
  - [ ] Postgres implementations for Project/Category/TimeEntry complete
  - [ ] Error mapping from Postgres → repository errors implemented
  - [ ] Integration tests for all three repositories pass locally (`-tags=integration`)

### Notes for M4
- Keep business rules (e.g., single active timer, category move restrictions) in services (Milestone 5), not in repositories. Repositories should be thin data mappers with clear, explicit APIs.
- Prefer UTC timestamps. Let the DB set `created_at`/`updated_at` via defaults and SQL `now()` where appropriate; read back with `RETURNING`.
- Use explicit typing for all public APIs, parameters, and return values.


