## Milestone 1: Step-by-step implementation plan

- [ ] Confirm repository context: `https://github.com/Gargair/clockwork`
  - [ ] Clone the repo and checkout the working branch
  - [ ] Pull latest changes from `main`

- [ ] Initialize repository layout and shared configs
  - [ ] Create top-level directories per architecture
    - [ ] `client/`
    - [ ] `server/`
    - [ ] `.github/workflows/`
  - [ ] Add root configs
    - [ ] `.gitignore` (Go, Node, OS/IDE ignores)
    - [ ] `.editorconfig` (UTF-8, LF, 2 spaces for TS/JS, tabs/spaces per language as preferred)
    - [ ] Confirm `docs/` remains source of truth for dev workflow

- [ ] Scaffold the client (Vite + React + TypeScript)
  - [ ] Create the app skeleton non-interactively
    ```bash
    npm create vite@latest client -- --template react-ts
    ```
  - [ ] Inside `client/`, install baseline deps
    ```bash
    cd client
    npm i react-router-dom
    npm i -D @types/react-router-dom eslint @typescript-eslint/parser @typescript-eslint/eslint-plugin eslint-plugin-react eslint-plugin-react-hooks prettier eslint-config-prettier eslint-plugin-prettier
    ```
  - [ ] Configure TypeScript strictness in `client/tsconfig.json`
    - [ ] "strict": true
    - [ ] "noImplicitAny": true
    - [ ] "noUncheckedIndexedAccess": true
  - [ ] Add ESLint config `client/.eslintrc.cjs` (TS + React + Prettier)
  - [ ] Add Prettier config `client/.prettierrc` and `client/.prettierignore`
  - [ ] Add basic router and layout
    - [ ] `client/src/app/App.tsx` (shell with header/footer and an outlet)
    - [ ] Wire router in `client/src/main.tsx` with a root `App` route and a placeholder `Home` page
  - [ ] Update `client/package.json` scripts
    - [ ] `dev`, `build`, `preview` (from Vite)
    - [ ] `lint`: eslint "src/**/*.{ts,tsx}"
    - [ ] `format`: prettier --write "src/**/*.{ts,tsx,css,json,md}"

- [ ] Scaffold the server (Go module and package skeletons)
  - [ ] Create module and folder structure
    ```bash
    mkdir -p server/cmd/server
    mkdir -p server/internal/{config,http,service,repository,db,domain,clock}
    cd server
    go mod init github.com/Gargair/clockwork/server
    go fmt ./...
    ```
  - [ ] Add minimal compilable placeholders
    - [ ] `server/cmd/server/main.go` with `package main` and empty `main()`
    - [ ] `server/internal/config/config.go` with package `config`
    - [ ] `server/internal/http/http.go` with package `http`
    - [ ] `server/internal/service/service.go` with package `service`
    - [ ] `server/internal/repository/repository.go` with package `repository`
    - [ ] `server/internal/db/db.go` with package `db`
    - [ ] `server/internal/domain/domain.go` with package `domain`
    - [ ] `server/internal/clock/clock.go` with package `clock`
  - [ ] Verify module builds
    ```bash
    cd server
    go build ./...
    ```

- [ ] Set up Go linting (`golangci-lint`)
  - [ ] Add `server/.golangci.yml` with sensible defaults (gosimple, govet, staticcheck, errcheck)
  - [ ] Local run (optional during M1)
    ```bash
    golangci-lint run ./...
    ```

- [ ] Set up JavaScript/TypeScript linting and formatting
  - [ ] Ensure ESLint + Prettier are configured (from step 2)
  - [ ] Verify locally
    ```bash
    cd client
    npm run lint && npm run format
    ```

- [ ] Continuous Integration (GitHub Actions)
  - [ ] Create `.github/workflows/ci.yml` with two jobs: `server` and `client`
    - [ ] Server job: `ubuntu-latest`
    - [ ] Setup Go 1.22+
    - [ ] Run `go test ./...` and `go build ./...`
    - [ ] Client job: `ubuntu-latest`
    - [ ] Setup Node.js 20+
    - [ ] Run `npm ci` and `npm run build` in `client/`
  - [ ] Trigger on `pull_request` and `push` to main branches

- [ ] Local verification and first commit
  - [ ] Run locally
    ```bash
    # Server
    cd server && go test ./... && go build ./...

    # Client
    cd ../client && npm install && npm run build
    ```
  - [ ] Commit structured scaffolding
    - [ ] Include `client/`, `server/`, `.github/workflows/ci.yml`, `.gitignore`, `.editorconfig`, and lint configs
  - [ ] Push and confirm CI passes

### Notes and conventions for M1
- Keep server packages as empty, compilable placeholders; actual HTTP wiring lands in Milestone 3
- Keep client UI minimal: router skeleton and base layout only
- Prefer strict TypeScript settings and explicit typing for public APIs
- Do not introduce DB or migrations yet (Milestone 2)

### Verification checklist (maps to acceptance)
- [ ] Server: `go test ./...` passes
- [ ] Server: `go build ./...` succeeds
- [ ] Client: `npm run build` succeeds
- [ ] CI: Both jobs complete successfully on PR and push


