## Milestone 7: Step-by-step implementation plan

- [x] 1: Create shared client schemas and types (with Zod)
  - [x] Add `client/src/types/schemas.ts` defining Zod schemas mirroring `docs/api.md`:
    - [x] `ProjectSchema`
    - [x] `CategorySchema`
    - [x] `TimeEntrySchema`
    - [x] `ErrorResponseSchema`
  - [x] Add `client/src/types/index.ts` to export explicit types inferred from schemas:
    - [x] `export type Project = z.infer<typeof ProjectSchema>`
    - [x] `export type Category = z.infer<typeof CategorySchema>`
    - [x] `export type TimeEntry = z.infer<typeof TimeEntrySchema>`
    - [x] `export type ErrorResponse = z.infer<typeof ErrorResponseSchema>`
  - [x] Include array helpers where needed: `export const ProjectListSchema = z.array(ProjectSchema)` (same for others)

- [x] 2: Add Zod for runtime validation
  - [x] Install dependency in `client`: `npm i zod`
  - [x] Ensure TypeScript settings are strict enough to avoid `any` leakage

- [x] 3: Configure API base URL (dev/prod)
  - [x] Add `client/src/api/config.ts`:
    - [x] `export const API_BASE_URL: string = (import.meta as any).env?.VITE_API_BASE_URL ?? (window.location.origin || '');`
    - In development, set `VITE_API_BASE_URL` to `http://localhost:8080` via `.env.development.local` (optional; CORS defaults to `*` in dev per server config).
    - In production, the SPA is served by the Go server; relative paths work.

- [x] 4: Implement a typed HTTP helper with schema validation
  - [x] Add `client/src/api/http.ts`:
    - [x] `export interface RequestOptions { method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'; headers?: Record<string, string>; body?: unknown; signal?: AbortSignal }`
    - [x] `export async function requestJson<T>(path: string, options: RequestOptions | undefined, schema: z.ZodSchema<T>): Promise<T>`
      - Prefix `path` with `API_BASE_URL` when absolute URL not provided
      - Set `Accept: application/json` and `Content-Type: application/json` when sending a body
      - Serialize `body` as JSON when provided
      - Parse successful responses with the provided Zod `schema` and return typed data
      - On non-2xx, attempt to parse `ErrorResponse` via `ErrorResponseSchema` and throw an `Error` that includes `code` and `requestId`; if parsing fails, include raw text

- [x] 5: Add minimal endpoint wrappers (using schemas)
  - [x] Add `client/src/api/health.ts`:
    - [x] `export async function getHealth(): Promise<{ status: string }>` calling `GET /healthz` and validating with `z.object({ status: z.string() })`
  - [x] Prepare modules (stubs) for upcoming milestones with typed signatures (implement later):
    - [x] `client/src/api/projects.ts` (list/create/get/update/delete) using `ProjectSchema`/`ProjectListSchema`
    - [x] `client/src/api/categories.ts` (list/create/get/update/delete) using `CategorySchema`/`z.array(CategorySchema)`
    - [x] `client/src/api/time.ts` (start/stop/getActive/listEntries) using `TimeEntrySchema`/`z.array(TimeEntrySchema)`

- [x] 6: Introduce an application error boundary
  - [x] Add `client/src/app/ErrorBoundary.tsx` with an explicit `Props` and `State`:
    - Catches render errors and displays a fallback with `requestId` if present
  - [x] Wrap the router tree in `ErrorBoundary` in `client/src/main.tsx`

- [x] 7: Establish theme tokens and base layout polish
  - [x] Augment `client/src/style.css` with CSS custom properties for colors/spacing/typography (light/dark), keeping existing styles
  - [x] Ensure `App` uses semantic HTML structure and respects the tokens

- [x] 8: Wire a simple health check on Home
  - [x] Add `client/src/pages/Home.tsx` logic to call `getHealth()` on mount and display the status (OK/error)
    - Show error message and `requestId` when available
    - Keep types explicit for local component state

- [x] 9: Developer experience (DX) basics
  - [x] Add `client/.env.development.example` with `VITE_API_BASE_URL=http://localhost:8080`
  - [x] Document in `docs/development.md` (Client section) how to point the SPA to the server in dev using `VITE_API_BASE_URL`
  - [x] Note Zod installation and the rationale for runtime validation of API responses

- [ ] 10: Verification steps
  - [ ] Start the server (see `docs/development.md`) and ensure it exposes `/healthz`
  - [ ] In another terminal: `cd client; npm run dev`
  - [ ] Load the SPA at the Vite URL; confirm:
    - App shell renders with header/footer and base styles
    - Home shows health status fetched via the API client and validated by Zod
    - Error boundary presents a readable fallback when forced
    - Intentionally corrupting the response in dev (e.g., via proxy/mock) surfaces a Zod validation error

- [ ] 11: Acceptance checklist (aligns with Implementation Plan)
  - [ ] `client/src/api` provides a typed `requestJson<T>` that validates with Zod schemas, and `getHealth()` uses a Zod schema
  - [ ] `client/src/types` exports Zod schemas and explicit types for `Project`, `Category`, `TimeEntry`, and `ErrorResponse`
  - [ ] Error boundary in place for the app tree
  - [ ] Theme tokens exist and are used in the base layout
  - [ ] Dev app loads and successfully calls `/healthz` via the API client, with runtime validation enforced

### Notes for M7
- Prefer explicit types for public functions, params, and return values.
- Do not introduce implicit `any`; enable strict TypeScript settings if not already enabled.
- Use Zod schemas as the single source of truth for response shapes; derive TypeScript types from them for safety and consistency.
- Keep business logic out of the client for now; this milestone focuses on foundations only.


