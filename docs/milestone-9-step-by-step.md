## Milestone 9: Step-by-step implementation plan

- [x] 1: Preparations and validation
  - [x] Confirm server endpoints and payloads in `docs/api.md` for categories.
  - [x] Ensure `CategorySchema` and `CategoryListSchema` in `client/src/types/schemas.ts` match server responses:
    - `id`, `projectId`: UUID
    - `parentCategoryId`: UUID or `null`
    - `name`: string
    - `description`: optional string
    - `createdAt`, `updatedAt`: RFC3339 strings

- [x] 2: Implement typed Categories API client
  - [x] Fill in `client/src/api/categories.ts` using the shared HTTP helper and Zod schemas
    - [x] `listCategories(projectId: string): Promise<Category[]>`
      - GET `/api/projects/{projectId}/categories` with `CategoryListSchema`
    - [x] `createCategory(projectId: string, input: CreateCategoryInput): Promise<Category>`
      - POST `/api/projects/{projectId}/categories` with `CategorySchema`
    - [x] `getCategory(projectId: string, categoryId: string): Promise<Category>`
      - GET `/api/projects/{projectId}/categories/{categoryId}` with `CategorySchema`
    - [x] `updateCategory(projectId: string, categoryId: string, input: UpdateCategoryInput): Promise<Category>`
      - PATCH `/api/projects/{projectId}/categories/{categoryId}` with `CategorySchema`
    - [x] `deleteCategory(projectId: string, categoryId: string): Promise<void>`
      - DELETE `/api/projects/{projectId}/categories/{categoryId}` with `z.undefined()`
  - Notes:
    - Prefer explicit parameter and return types for all exported functions.
    - Inputs support `description?: string | null` and `parentCategoryId?: string | null`.

- [ ] 3: Create a reusable `useCategories` hook
  - [ ] Add `client/src/hooks/useCategories.ts`
    - Return type (explicit):
      - `status: 'idle' | 'loading' | 'success' | 'error'`
      - `categories: Category[]`
      - `error: { message: string; code?: string; requestId?: string } | null`
      - CRUD: `refresh()`, `create(input)`, `update(id, input)`, `remove(id)`
    - Behavior:
      - Load list on mount via `listCategories(projectId)`
      - On mutations, call API and then `refresh()` to reconcile
      - Capture `ApiError` details (`code`, `requestId`) when available
    - [ ] Add `buildCategoryTree(categories: Category[]): CategoryNode[]` pure helper
      - Converts flat list to a tree, ensures stable ordering (e.g., by `name`)
      - Defensive: gracefully handle orphans if any (should not occur with server constraints)

- [ ] 4: Build UI components
  - [ ] `client/src/components/CategoryForm.tsx`
    - Props (explicit types):
      - `initial?: { name: string; description?: string | null; parentCategoryId?: string | null }`
      - `onSubmit: (values: { name: string; description?: string | null; parentCategoryId?: string | null }) => void | Promise<void>`
      - `submitLabel?: string`
      - `disabled?: boolean`
      - `parentOptions: Array<{ value: string | null; label: string }>`
    - Controlled inputs for name, description, and a Parent select (includes “None”).
  - [ ] `client/src/components/CategoryTree.tsx`
    - Props (explicit types):
      - `tree: CategoryNode[]`
      - Callbacks: `onAdd(parentId | null)`, `onEdit(categoryId)`, `onDelete(categoryId)`
      - UI state: `loading?: boolean`, `error?: string | null`
    - Behavior:
      - Expand/collapse per node
      - Actions per node: Add subcategory, Edit, Delete
      - ARIA roles for a11y (`role="tree"`, `role="treeitem"`) and focus handling

- [ ] 5: Categories page and routing
  - [ ] Add `client/src/pages/Categories.tsx`
    - Read `projectId` from route params
    - Use `useCategories(projectId)` to load data; display loading/error states
    - Render:
      - Create form for top-level category (Parent = None)
      - `CategoryTree` listing with inline edit/delete and “add child” flows
    - Display API constraint errors in an alert region:
      - `invalid_parent`, `cross_project_parent`, `category_cycle`
  - [ ] Update routing in `client/src/main.tsx`:
    - Add `<Route path="/projects/:projectId/categories" element={<Categories />} />`
  - [ ] Update `client/src/pages/Projects.tsx`:
    - Add a “Manage categories” link per project row that navigates to `/projects/{id}/categories`

- [ ] 6: Tests (Vitest + RTL)
  - [ ] `client/src/pages/Categories.test.tsx`
    - Mock `../api/categories` like `Projects.test.tsx` does for projects
    - Cover:
      - Listing renders as a tree
      - Create top-level and child category
      - Update category name and parent (re-parenting)
      - Delete category
      - Error paths surface readable messages and `requestId` when available
      - Cycle error (`category_cycle`) shows user-friendly message
  - [ ] `CategoryTree` component tests
    - Expand/collapse behavior
    - Actions invoke callbacks with correct IDs
  - [ ] `buildCategoryTree` helper tests
    - Acyclic nesting builds correctly
    - Orphan nodes are handled defensively (if encountered)

- [ ] 7: Verification steps
  - [ ] Start the server; verify `GET /api/projects/{projectId}/categories` returns 200 with data
  - [ ] `cd client; npm run dev`; navigate to `/projects/{id}/categories`
    - See tree (empty state if none)
    - Create top-level category → appears
    - Add child category → nested correctly
    - Edit and re-parent → updates persist after reload
    - Delete → removed from tree

- [ ] 8: Acceptance checklist (aligns with Implementation Plan)
  - [ ] `client/src/api/categories.ts` implements typed list/create/get/update/delete with Zod validation
  - [ ] `useCategories` hook provides typed state and CRUD helpers
  - [ ] `CategoryForm` provides validated inputs including Parent select
  - [ ] `CategoryTree` renders hierarchy with core actions
  - [ ] `CategoriesPage` supports listing, creating, updating, re-parenting, and deleting
  - [ ] Routing to `/projects/:projectId/categories` exists and works
  - [ ] Component tests cover happy paths and key error scenarios and pass

### Notes for M9
- Prefer explicit types for all exported APIs, hooks, and complex props.
- Keep the UI accessible; use appropriate ARIA roles and keyboard navigation cues.
- Surface server error codes directly to help correlate with logs, similar to Projects.


