## Milestone 8: Step-by-step implementation plan

- [x] 1: Implement typed Projects API client
  - [x] Fill in `client/src/api/projects.ts` using the shared HTTP helper and Zod schemas
    - [x] `listProjects(): Promise<Project[]>` → `GET /api/projects` with `ProjectListSchema`
    - [x] `createProject(input: CreateProjectInput): Promise<Project>` → `POST /api/projects` with `ProjectSchema`
    - [x] `getProject(projectId: string): Promise<Project>` → `GET /api/projects/{projectId}` with `ProjectSchema`
    - [x] `updateProject(projectId: string, input: UpdateProjectInput): Promise<Project>` → `PATCH /api/projects/{projectId}` with `ProjectSchema`
    - [x] `deleteProject(projectId: string): Promise<void>` → `DELETE /api/projects/{projectId}` with `z.undefined()`
  - Notes:
    - Use `requestJson<T>()` from `client/src/api/http.ts` for all calls and validate responses against `ProjectSchema`/`ProjectListSchema`.
    - Keep all function parameters and return types explicit. Do not rely on inference for exported functions.

- [x] 2: Create a reusable `useProjects` hook
  - [x] Add `client/src/hooks/useProjects.ts`
    - Expose explicit types:
      - `status: 'idle' | 'loading' | 'success' | 'error'`
      - `projects: Project[]`
      - `error: { message: string; code?: string; requestId?: string } | null`
      - CRUD methods: `refresh()`, `create(input)`, `update(id, input)`, `remove(id)`
    - Behavior:
      - Load list on mount via `listProjects()`
      - On mutations, either optimistically update local state then reconcile, or call API and then `refresh()` (start simple with refetch)
      - Capture `ApiError` details (`code`, `requestId`) when present

- [x] 3: Build `ProjectForm` component
  - [x] Add `client/src/components/ProjectForm.tsx`
    - Props (explicit types):
      - `initial?: { name: string; description?: string | null }`
      - `onSubmit: (values: { name: string; description?: string | null }) => void | Promise<void>`
      - `submitLabel?: string`
      - `disabled?: boolean`
    - Controlled inputs for `name` (required, trimmed) and `description` (optional)
    - Basic inline validation (e.g., show message if name is empty)

- [x] 4: Create `ProjectsPage` UI
  - [x] Add `client/src/pages/Projects.tsx`
    - Uses `useProjects()` to:
      - Render list of projects (name, description, created/updated timestamps)
      - Provide a create form using `ProjectForm`
      - Provide edit and delete controls per row (edit can be inline or a simple toggle)
    - Handle and display loading/error states, including `ApiError` info when present

- [x] 5: Wire routing and navigation
  - [x] Update `client/src/app/App.tsx` to include a `Link` to `/projects`
  - [x] Update `client/src/main.tsx` to add a route:
    - `<Route path="/projects" element={<Projects />} />`

- [x] 6: Tests (component)
  - [x] Add client test tooling (if not present):
    - Dev dependencies: `vitest`, `@testing-library/react`, `@testing-library/user-event`, `@testing-library/jest-dom`, `jsdom`
    - Add scripts in `client/package.json`:
      - `"test": "vitest --run"`, `"test:watch": "vitest"`
    - Minimal `vitest` config to use `jsdom`
  - [x] Write tests for `ProjectsPage` (mock API layer):
    - Renders list from `listProjects()`
    - Creating a project calls `createProject()` and updates UI
    - Updating a project calls `updateProject()` and updates UI
    - Deleting a project calls `deleteProject()` and removes from UI
    - Error paths surface readable messages (and `requestId` if available)

- [x] 7: Verification steps
  - [x] Start the server (per `docs/development.md`); confirm `GET /api/projects` returns 200
  - [x] In another terminal: `cd client; npm run dev`
  - [x] Navigate to `/projects`:
    - See list (empty state if none)
    - Create project → appears in list
    - Edit project name/description → changes persist after reload
    - Delete project → removed, and 404s handled gracefully if racing

- [x] 8: Acceptance checklist (aligns with Implementation Plan)
  - [x] `client/src/api/projects.ts` implements typed list/create/get/update/delete with Zod validation
  - [x] `useProjects` hook provides typed state and CRUD helpers
  - [x] `ProjectForm` provides validated inputs and explicit props
  - [x] `ProjectsPage` supports listing, creating, updating, and deleting
  - [x] Navigation to `/projects` exists and works
  - [x] Component tests cover happy paths and key error scenarios and pass

### Notes for M8
- Prefer explicit types for all exported APIs, including hook return shapes.
- Keep UI lean; prioritize correctness and accessibility over styling.
- Defer advanced UX (optimistic updates, inline validation with Zod) to later milestones unless trivial.
- Reuse `ApiError` details in error banners to help correlate with server logs.


