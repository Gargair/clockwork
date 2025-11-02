## Milestone 5: Step-by-step implementation plan

 - [x] 1: Establish service package structure and common errors
  - [x] Create `server/internal/service/errors.go`
  - [x] Define explicit error variables (public, for handler mapping later):
   - [x] `var ErrNoActiveTimer = errors.New("service: no active timer")`
   - [x] `var ErrCategoryCycle = errors.New("service: category cycle detected")`
   - [x] `var ErrCrossProjectParent = errors.New("service: parent category belongs to a different project")`
   - [x] `var ErrInvalidParent = errors.New("service: invalid parent category")`

 - [x] 2: Define service interfaces and constructors in `server/internal/service/service.go`
  - [x] Add interfaces with explicit parameter and return types and context:
   - [x] `type ProjectService interface { Create(ctx context.Context, name string, description *string) (domain.Project, error); Update(ctx context.Context, id uuid.UUID, name string, description *string) (domain.Project, error); Delete(ctx context.Context, id uuid.UUID) error; GetByID(ctx context.Context, id uuid.UUID) (domain.Project, error); List(ctx context.Context) ([]domain.Project, error) }`
   - [x] `type CategoryService interface { Create(ctx context.Context, projectID uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error); Update(ctx context.Context, id uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error); Delete(ctx context.Context, id uuid.UUID) error; GetByID(ctx context.Context, id uuid.UUID) (domain.Category, error); ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Category, error); ListChildren(ctx context.Context, parentID uuid.UUID) ([]domain.Category, error) }`
   - [x] `type TimeTrackingService interface { Start(ctx context.Context, categoryID uuid.UUID) (domain.TimeEntry, error); StopActive(ctx context.Context) (domain.TimeEntry, error); GetActive(ctx context.Context) (*domain.TimeEntry, error); ListByCategory(ctx context.Context, categoryID uuid.UUID) ([]domain.TimeEntry, error); ListByCategoryAndRange(ctx context.Context, categoryID uuid.UUID, start time.Time, end time.Time) ([]domain.TimeEntry, error) }`
  - [x] Define concrete constructors that accept dependencies explicitly:
   - [x] `func NewProjectService(repo repository.ProjectRepository) ProjectService`
   - [x] `func NewCategoryService(repo repository.CategoryRepository) CategoryService`
   - [x] `func NewTimeTrackingService(repo repository.TimeEntryRepository, categoryRepo repository.CategoryRepository, clk clock.Clock) TimeTrackingService`

 - [x] 3: Implement `ProjectService` in `server/internal/service/project_service.go`
  - [x] Validate `name` is non-empty after `strings.TrimSpace`
  - [x] Implement pass-through CRUD/list using `repository.ProjectRepository`
  - [x] Keep timestamps DB-driven; avoid time logic here

 - [x] 4: Implement `CategoryService` in `server/internal/service/category_service.go`
  - [x] On Create: if `parentCategoryID != nil`, fetch parent and assert:
   - [x] Parent exists or return `ErrInvalidParent`
   - [x] `parent.ProjectID == projectID` else return `ErrCrossProjectParent`
  - [x] On Update (including potential parent change):
   - [x] Load current category to know `ProjectID`
   - [x] If parent provided, validate parent belongs to same project
   - [x] Detect cycles: parent cannot be the category itself or any of its descendants
   - [x] Fetch descendants via repeated `ListChildren` calls (BFS/DFS) and detect membership; if cycle → `ErrCategoryCycle`
  - [x] On Delete: call repository `Delete`; DB will `SET NULL` on children per schema
  - [x] Disallow cross-project moves implicitly by not exposing a way to change `ProjectID`

 - [x] 5: Implement `TimeTrackingService` in `server/internal/service/time_service.go`
  - [x] `Start`:
  - [x] Ensure category exists via `CategoryRepository.GetByID`
  - [x] Capture a single `now := clk.Now()`; if an active entry exists, stop it with `stoppedAt = now` and compute duration using `now`
  - [x] Create the new entry with `StartedAt = now`, `StoppedAt = nil`, `DurationSeconds = nil`
  - [x] `StopActive`:
  - [x] Look up active via `FindActive`; if nil, return `ErrNoActiveTimer`
  - [x] Compute `durationSeconds` = `int32(clk.Now().Sub(active.StartedAt).Seconds())`, clamp to `>=0`
  - [x] Call `repo.Stop(active.ID, stoppedAt, &durationSeconds)` and return updated entry
  - [x] `GetActive`, `ListByCategory`, `ListByCategoryAndRange`: thin pass-throughs to repository

 - [x] 6: Add explicit service-level types and wiring
  - [x] Define unexported structs `projectService`, `categoryService`, `timeTrackingService` implementing the interfaces
  - [x] Store required dependencies as fields with explicit types (e.g., repositories, `clock.Clock`)
  - [x] Add compile-time interface assertions: `var _ ProjectService = (*projectService)(nil)` (and similarly for others)

 - [x] 7: Unit tests (TDD) for invariants and edge cases in `server/internal/service`
  - [x] Create `clock_test.go` fake clock implementing `clock.Clock` with controllable `Now()`
  - [x] Create minimal in-memory fakes for repositories inside tests (hand-rolled) to avoid DB
  - [x] `ProjectService` tests:
   - [x] Reject empty/whitespace names
   - [x] Create → GetByID roundtrip passes through fields
  - [x] `CategoryService` tests:
   - [x] Create with parent in same project succeeds
   - [x] Create with parent from different project → `ErrCrossProjectParent`
   - [x] Update changing parent to descendant → `ErrCategoryCycle`
   - [x] Update name/description only succeeds
  - [x] `TimeTrackingService` tests:
   - [x] `Start` when no active exists creates active entry with exact `StartedAt`
   - [x] `Start` when active exists stops the previous entry and starts a new one; assert `prev.stoppedAt == new.startedAt` and equals fake clock `now`
   - [x] `StopActive` computes `durationSeconds` correctly with fake clock and clears active
   - [x] `GetActive` reflects state transitions
  - [x] Run: `go test ./server/...`

 - [x] 8: Cross-cutting error mapping and consistency
  - [x] Map repository errors to service semantics where appropriate (e.g., `repository.ErrNotFound` → pass-through for Get/Delete)
  - [x] Ensure all public functions return explicit errors from `errors.go` for invariant violations

 - [ ] 9: Light wiring for future handlers (no HTTP yet)
  - [ ] Provide simple factory in `service/service.go` or a small `service/wire.go` with constructors
  - [ ] Do not register HTTP routes (reserved for Milestone 6)

 - [ ] 10: Documentation and developer workflow
  - [ ] Add a note to `docs/development.md` about running service unit tests (`go test ./server/...`)
  - [ ] Reference invariants in `docs/domain-model.md` to keep behavior aligned

 - [ ] 11: Acceptance checklist (aligns with Implementation Plan)
  - [ ] Category tree constraints enforced in services (same project parent, no cycles)
  - [ ] Single-active-timer invariant enforced in `TimeTrackingService`
  - [ ] Auto-stop on `Start` uses the same timestamp as the new `StartedAt`
  - [ ] Unit tests for invariants and edge cases pass locally: `go test ./server/...`
  - [ ] Explicit types for all public service APIs and constructors

### Notes for M5
- Keep repositories as thin mappers; enforce business rules only in the services.
- Use `clock.Clock` for all time decisions to make tests deterministic.
- Prefer table-driven tests and descriptive test names. Avoid global state in fakes.


