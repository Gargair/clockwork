# API

All endpoints return and accept JSON. Timestamps are RFC3339. IDs are UUID strings.

- Request header: `Content-Type: application/json`
- Response header: `Content-Type: application/json`
- Errors conform to `ErrorResponse` with a machine-readable `code` and the `requestId` from `X-Request-ID`.

ErrorResponse
```json
{
  "code": "invalid_id",
  "message": "invalid projectId",
  "requestId": "abc123"
}
```

Common error codes
- invalid_json
- invalid_id
- invalid_time, invalid_time_range
- invalid_project_name
- invalid_parent, cross_project_parent
- category_cycle
- no_active_timer
- not_found
- internal

## Projects

POST /api/projects
- Create a project
- Request
```json
{
  "name": "My Project",
  "description": "Optional description"
}
```
- 201 Created
```json
{
  "id": "b7c1a5a8-6fcb-4d19-9e76-8f2d2e1d2b2b",
  "name": "My Project",
  "description": "Optional description",
  "createdAt": "2025-11-02T12:34:56Z",
  "updatedAt": "2025-11-02T12:34:56Z"
}
```
- 400: invalid_json | invalid_project_name

GET /api/projects
- 200 OK
```json
[
  {
    "id": "...",
    "name": "My Project",
    "description": "Optional description",
    "createdAt": "2025-11-02T12:34:56Z",
    "updatedAt": "2025-11-02T12:34:56Z"
  }
]
```

GET /api/projects/{projectId}
- 200 OK returns `ProjectResponse`
- 400: invalid_id
- 404: not_found

PATCH /api/projects/{projectId}
- Update name and/or description. Name is required.
- Request
```json
{
  "name": "Renamed",
  "description": "Optional"
}
```
- 200 OK returns `ProjectResponse`
- 400: invalid_id | invalid_json | invalid_project_name
- 404: not_found

DELETE /api/projects/{projectId}
- 204 No Content
- 400: invalid_id
- 404: not_found

## Categories (scoped to project)

POST /api/projects/{projectId}/categories
- Create a category in a project.
- Request
```json
{
  "name": "Frontend",
  "description": "Optional",
  "parentCategoryId": null
}
```
- 201 Created
```json
{
  "id": "...",
  "projectId": "...",
  "parentCategoryId": null,
  "name": "Frontend",
  "description": "Optional",
  "createdAt": "2025-11-02T12:34:56Z",
  "updatedAt": "2025-11-02T12:34:56Z"
}
```
- 400: invalid_id (bad projectId/parentCategoryId) | invalid_json | invalid_parent | cross_project_parent

GET /api/projects/{projectId}/categories
- 200 OK: `CategoryResponse[]`
- 400: invalid_id

GET /api/projects/{projectId}/categories/{categoryId}
- 200 OK: `CategoryResponse`
- 400: invalid_id
- 404: not_found

PATCH /api/projects/{projectId}/categories/{categoryId}
- Update name/description/parent.
- Request
```json
{
  "name": "Frontend v2",
  "description": "Optional",
  "parentCategoryId": null
}
```
- 200 OK: `CategoryResponse`
- 400: invalid_id | invalid_json | invalid_parent | cross_project_parent
- 409: category_cycle
- 404: not_found

DELETE /api/projects/{projectId}/categories/{categoryId}
- 204 No Content
- 400: invalid_id
- 404: not_found

## Time Tracking

POST /api/time/start
- Starts a new entry for the given category.
- Request
```json
{ "categoryId": "..." }
```
- 201 Created
```json
{
  "id": "...",
  "categoryId": "...",
  "startedAt": "2025-11-02T12:34:56Z",
  "stoppedAt": null,
  "durationSeconds": null,
  "createdAt": "2025-11-02T12:34:56Z",
  "updatedAt": "2025-11-02T12:34:56Z"
}
```
- 400: invalid_json | invalid_id

POST /api/time/stop
- Stops the current active entry.
- 200 OK: `TimeEntryResponse` (with final `stoppedAt` and `durationSeconds`)
- 409: no_active_timer

GET /api/time/active
- 200 OK: `TimeEntryResponse` or `null` when none

GET /api/time/entries?categoryId=&from=&to=
- Lists entries for a category, optionally within a time range.
- Query params
  - categoryId: UUID (required)
  - from, to: RFC3339 timestamps (optional; if both provided, `from` must be <= `to`)
- 200 OK: `TimeEntryResponse[]`
- 400: invalid_id | invalid_time | invalid_time_range

## Validation rules
- UUID path/query params must be valid UUID strings → 400 `invalid_id`.
- JSON bodies are decoded strictly with `DisallowUnknownFields` → 400 `invalid_json`.
- Project name must be non-empty/trimmed → 400 `invalid_project_name`.
- `from`/`to` must be valid RFC3339; if both present, `from <= to` → 400 `invalid_time`/`invalid_time_range`.
- Category parent must exist and belong to the same project → 400 `invalid_parent`/`cross_project_parent`.
- Updating category to create a cycle → 409 `category_cycle`.
- Stopping without an active timer → 409 `no_active_timer`.
- Missing entities → 404 `not_found`.
