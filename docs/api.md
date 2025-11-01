# API (initial sketch)

This is an initial outline. Details will be refined during implementation and testing.

## Projects
- POST /api/projects
- GET /api/projects
- GET /api/projects/{projectId}
- PATCH /api/projects/{projectId}
- DELETE /api/projects/{projectId}

## Categories (scoped to project)
- POST /api/projects/{projectId}/categories
- GET /api/projects/{projectId}/categories
- GET /api/projects/{projectId}/categories/{categoryId}
- PATCH /api/projects/{projectId}/categories/{categoryId}
- DELETE /api/projects/{projectId}/categories/{categoryId}

## Time Tracking
- POST /api/time/start { categoryId }
  - Starts a new active TimeEntry; stops any currently active entry first
- POST /api/time/stop
  - Stops the current active entry (if any) and returns it with final duration
- GET /api/time/active
  - Returns the currently active TimeEntry if one exists
- GET /api/time/entries?projectId=&categoryId=&from=&to=
  - Lists time entries with optional filters

## Example payloads
```json
// Create project
{
  "name": "My Project",
  "description": "Optional"
}
```

```json
// Create category
{
  "name": "Frontend",
  "description": "Optional",
  "parentCategoryId": null
}
```

```json
// Start time tracking
{
  "categoryId": "cat_123"
}
```

