## Milestone 6: Step-by-step implementation plan

 - [x] 1: Establish handler scaffolding and shared HTTP helpers
  - [x] Create `server/internal/http/models.go` with explicit request/response types and JSON tags:
  - [x] Projects: `ProjectCreateRequest`, `ProjectUpdateRequest`, `ProjectResponse`
  - [x] Categories: `CategoryCreateRequest`, `CategoryUpdateRequest`, `CategoryResponse`
  - [x] Time: `TimeStartRequest`, `TimeEntryResponse`, `ActiveTimerResponse`
  - [x] Errors: `ErrorResponse { code string, message string, requestId string }`
  - [x] Create `server/internal/http/json.go` helpers:
  - [x] `decodeJSON(r *http.Request, dst any) error` (strict decoder: `DisallowUnknownFields`)
  - [x] `writeJSON(w http.ResponseWriter, status int, v any)`
  - [x] `writeError(w http.ResponseWriter, r *http.Request, status int, code, msg string)`
  - [x] `parseUUID(str string) (uuid.UUID, error)` and `parseTimeRFC3339(str string) (time.Time, error)`

 - [X] 2: Define consistent error mapping from services to HTTP
  - [x] Create `server/internal/http/errors.go` with:
  - [x] `type apiErrorCode string`
  - [x] Map `service.ErrInvalidProjectName` → `400 Bad Request`
  - [x] Map `service.ErrInvalidParent`, `service.ErrCrossProjectParent` → `400 Bad Request`
  - [x] Map `service.ErrCategoryCycle` → `409 Conflict`
  - [x] Map `service.ErrNoActiveTimer` (on stop) → `409 Conflict`
  - [x] Repository not found → `404 Not Found`
  - [x] Unknown → `500 Internal Server Error`
  - [x] Ensure `writeError` includes `X-Request-ID` value in `ErrorResponse.requestId`

 - [X] 3: Implement `ProjectHandler`
  - [x] Add `server/internal/http/project_handler.go` with methods using `service.ProjectService`:
  - [x] `POST /api/projects` → validate name, create → `201 Created` with `ProjectResponse`
  - [x] `GET /api/projects` → list → `200 OK` `[ProjectResponse]`
  - [x] `GET /api/projects/{projectId}` → `200 OK` or `404`
  - [x] `PATCH /api/projects/{projectId}` → validate, update → `200 OK`
  - [x] `DELETE /api/projects/{projectId}` → `204 No Content`
  - [x] Convert path params to `uuid.UUID`, handle invalid UUID as `400`
  - [x] Tests in `server/internal/http/project_handler_test.go` (httptest + fake service):
  - [x] Happy paths for all endpoints
  - [x] Invalid UUID returns `400`
  - [x] Empty/whitespace name returns `400`
  - [x] Not found returns `404`

 - [X] 4: Implement `CategoryHandler`
  - [x] Add `server/internal/http/category_handler.go` with `service.CategoryService`:
   - [x] `POST /api/projects/{projectId}/categories`
   - [x] `GET /api/projects/{projectId}/categories`
   - [x] `GET /api/projects/{projectId}/categories/{categoryId}`
   - [x] `PATCH /api/projects/{projectId}/categories/{categoryId}`
   - [x] `DELETE /api/projects/{projectId}/categories/{categoryId}`
  - [x] Validate `projectId`, `categoryId` UUIDs; map service errors per step 2
  - [x] Tests in `server/internal/http/category_handler_test.go`:
   - [x] Create with valid parent (same project) → `201`
   - [x] Cross-project parent → `400`
   - [x] Cycle on update → `409`
   - [x] Not found → `404`

 - [X] 5: Implement `TimeHandler`
  - [x] Add `server/internal/http/time_handler.go` with `service.TimeTrackingService`:
   - [x] `POST /api/time/start` with body `{ "categoryId": "..." }` → `201 Created` returns `TimeEntryResponse`
   - [x] `POST /api/time/stop` → `200 OK` with final entry, or `409` if no active
   - [x] `GET /api/time/active` → `200 OK` with active or `200 OK` with `null` payload
   - [x] `GET /api/time/entries?projectId=&categoryId=&from=&to=` → `200 OK` list
  - [x] Parse `projectId`/`categoryId` as UUIDs; parse `from`/`to` as RFC3339
  - [x] If both `from` and `to` provided and `from > to`, return `400`
  - [x] Tests in `server/internal/http/time_handler_test.go`:
   - [x] Start returns `201` and echoes category
   - [x] Stop without active → `409` with error code
   - [x] Active returns `null` when none
   - [x] Entries filter parsing and validation

 - [X] 6: Wire routes into the main router
  - [x] In `server/internal/http/http.go`, under `/api` mount sub-routers:
   - [x] `/api/projects` → `ProjectHandler`
   - [x] `/api/projects/{projectId}/categories` → `CategoryHandler`
   - [x] `/api/time` → `TimeHandler`
  - [x] Ensure JSON `Content-Type` responses and correct status codes
  - [x] Keep existing middleware (RequestID, logging, CORS) applied

 - [X] 7: End-to-end handler tests (no DB) with fake services
  - [x] Provide minimal fake implementations of service interfaces within `_test.go`
  - [x] Use `httptest.NewRecorder` + real `chi.Mux`
  - [x] Verify response schemas match `docs/api.md` (fields, casing, nullability)
  - [x] Command: `cd server; go test ./...` (no `-tags=integration` needed for handler tests)

 - [X] 8: Logging and observability
  - [x] Ensure each handler logs at start/end with `X-Request-ID` and outcome
  - [x] Include `requestId` in all `ErrorResponse`s via `writeError`

 - [X] 9: Documentation updates
  - [x] Expand `docs/api.md` with example request/response bodies for all endpoints
  - [x] Note validation rules (UUIDs, RFC3339 times) and error codes

 - [X] 10: Acceptance checklist (aligns with Implementation Plan)
  - [x] `ProjectHandler`, `CategoryHandler`, and `TimeHandler` implemented with explicit types
  - [x] Error mapping consistent and documented; structured `ErrorResponse` returned
  - [x] Routes registered under `/api/*` and served by `NewRouter`
  - [x] Handler tests cover happy paths and error cases and pass: `cd server; go test ./...`
  - [x] `docs/api.md` reflects implemented behavior with examples

### Notes for M6
- Prefer explicit types and annotations for all public handler APIs and DTOs.
- Keep business rules in services; handlers focus on validation, translation, and HTTP concerns.
- Use strict JSON decoding (`DisallowUnknownFields`) to fail fast on client mistakes.


