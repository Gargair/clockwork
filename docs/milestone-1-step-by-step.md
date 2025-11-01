## Milestone 1: Step-by-step implementation plan

- [X] 0: Confirm repository context: `https://github.com/Gargair/clockwork`
  - [X] Clone the repo and checkout the working branch
  - [X] Pull latest changes from `main`

- [X] 1: Initialize repository layout and shared configs
  - [X] Create top-level directories per architecture
    - [X] `client/`
    - [X] `server/`
    - [X] `.github/workflows/`
  - [X] Add root configs
    - [X] `.gitignore` (Go, Node, OS/IDE ignores)
    - [X] `.editorconfig` (UTF-8, LF, 2 spaces for TS/JS, tabs/spaces per language as preferred)
    - [X] Confirm `docs/` remains source of truth for dev workflow

- [X] 2: Scaffold the client (Vite + React + TypeScript)
  - [X] Create the app skeleton non-interactively
    ```bash
    npm create vite@latest client -- --template react-ts
    ```
  - [X] Inside `client/`, install baseline deps
    ```bash
    cd client
    npm i react-router-dom
    npm i -D @types/react-router-dom eslint @typescript-eslint/parser @typescript-eslint/eslint-plugin eslint-plugin-react eslint-plugin-react-hooks prettier eslint-config-prettier eslint-plugin-prettier
    ```
  - [X] Configure TypeScript strictness in `client/tsconfig.json`
    - [X] "strict": true
    - [X] "noImplicitAny": true
    - [X] "noUncheckedIndexedAccess": true
  - [X] Add ESLint config `client/.eslintrc.cjs` (TS + React + Prettier)
  - [X] Add Prettier config `client/.prettierrc` and `client/.prettierignore`
  - [X] Add basic router and layout
    - [X] `client/src/app/App.tsx` (shell with header/footer and an outlet)
    - [X] Wire router in `client/src/main.tsx` with a root `App` route and a placeholder `Home` page
  - [X] Update `client/package.json` scripts
    - [X] `dev`, `build`, `preview` (from Vite)
    - [X] `lint`: eslint "src/**/*.{ts,tsx}"
    - [X] `format`: prettier --write "src/**/*.{ts,tsx,css,json,md}"

- [X] 3: Scaffold the server (Go module and package skeletons)
  - [X] Create module and folder structure
    ```bash
    mkdir -p server/cmd/server
    mkdir -p server/internal/{config,http,service,repository,db,domain,clock}
    cd server
    go mod init github.com/Gargair/clockwork/server
    go fmt ./...
    ```
  - [X] Add minimal compilable placeholders
    - [X] `server/cmd/server/main.go` with `package main` and empty `main()`
    - [X] `server/internal/config/config.go` with package `config`
    - [X] `server/internal/http/http.go` with package `http`
    - [X] `server/internal/service/service.go` with package `service`
    - [X] `server/internal/repository/repository.go` with package `repository`
    - [X] `server/internal/db/db.go` with package `db`
    - [X] `server/internal/domain/domain.go` with package `domain`
    - [X] `server/internal/clock/clock.go` with package `clock`
  - [X] Verify module builds
    ```bash
    cd server
    go build ./...
    ```

- [X] 4: Set up Go linting (`golangci-lint`)
  - [X] Add `server/.golangci.yml` with sensible defaults (gosimple, govet, staticcheck, errcheck)
  - [X] Local run (optional during M1)
    ```bash
    golangci-lint run ./...
    ```

- [X] 5: Set up JavaScript/TypeScript linting and formatting
  - [X] Ensure ESLint + Prettier are configured (from step 2)
  - [X] Verify locally
    ```bash
    cd client
    npm run lint && npm run format
    ```

- [X] 6: Continuous Integration (GitHub Actions)
  - [X] Create `.github/workflows/ci.yml` with two jobs: `server` and `client`
    - [X] Server job: `ubuntu-latest`
    - [X] Setup Go 1.25+
    - [X] Run `go test ./...` and `go build ./...`
    - [X] Client job: `ubuntu-latest`
    - [X] Setup Node.js 25+
    - [X] Run `npm ci` and `npm run build` in `client/`
  - [X] Trigger on `pull_request` and `push` to main branches

- [ ] 7: Local verification and first commit
  - [X] Run locally
    ```bash
    # Server
    cd server && go test ./... && go build ./...

    # Client
    cd ../client && npm install && npm run build
    ```
  - [X] Commit structured scaffolding
    - [X] Include `client/`, `server/`, `.github/workflows/ci.yml`, `.gitignore`, `.editorconfig`, and lint configs
  - [ ] Push and confirm CI passes

### Notes and conventions for M1
- Keep server packages as empty, compilable placeholders; actual HTTP wiring lands in Milestone 3
- Keep client UI minimal: router skeleton and base layout only
- Prefer strict TypeScript settings and explicit typing for public APIs
- Do not introduce DB or migrations yet (Milestone 2)

### Verification checklist (maps to acceptance)
- [X] Server: `go test ./...` passes
- [X] Server: `go build ./...` succeeds
- [X] Client: `npm run build` succeeds
- [ ] CI: Both jobs complete successfully on PR and push


