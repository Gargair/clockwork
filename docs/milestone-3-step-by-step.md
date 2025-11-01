## Milestone 3: Step-by-step implementation plan

 - [x] 1: Extend server configuration (`internal/config`)
  - [x] Add fields to `Config` with sensible defaults and validation:
    - [x] `Port int` (env `PORT`, default `8080`)
    - [x] `Env string` (env `ENV`, values: `development|production`, default `development`)
    - [x] `StaticDir string` (env `STATIC_DIR`, default `client/dist`)
    - [x] `AllowedOrigins []string` (env `ALLOWED_ORIGINS`, CSV; default `*` in dev, empty in prod)
  - [x] Update `Load()` to parse and validate new fields (port > 0, known env)
  - [x] Document env vars in `docs/development.md` briefly

- [ ] 2: Add a clock abstraction (`internal/clock`)
  - [ ] Define `type Clock interface { Now() time.Time }`
  - [ ] Implement `type SystemClock struct{}` with `Now()` returning `time.Now().UTC()`
  - [ ] Provide a `NewSystemClock()` constructor for clarity

- [ ] 3: Add HTTP router and middleware (`internal/http`)
  - [ ] Add dependencies:
    - [ ] `github.com/go-chi/chi/v5`
    - [ ] `github.com/go-chi/cors`
  - [ ] Implement `func NewRouter(cfg config.Config, dbConn *sql.DB, clk clock.Clock, logger *slog.Logger) http.Handler`
    - [ ] Base router: `chi.NewRouter()`
    - [ ] Middleware:
      - [ ] `middleware.RequestID` (unique per request)
      - [ ] `middleware.RealIP`
      - [ ] `middleware.Recoverer`
      - [ ] Request logging middleware using `slog` (method, path, status, duration, request ID)
      - [ ] `cors.Handler` with:
        - [ ] `AllowedOrigins`: from `cfg.AllowedOrigins` (default `*` in dev, explicit in prod)
        - [ ] `AllowedMethods`: `GET,POST,PUT,PATCH,DELETE,OPTIONS`
        - [ ] `AllowedHeaders`: common headers (`Accept, Authorization, Content-Type, X-Request-ID`)
        - [ ] `ExposedHeaders`: `X-Request-ID`
        - [ ] `AllowCredentials`: false (tighten later as needed)
        - [ ] `MaxAge`: 300
    - [ ] Routes:
      - [ ] `GET /healthz`: returns JSON `{ "ok": true, "db": "up|down", "time": <utc now> }`
        - [ ] Use a short timeout (e.g., 1s) to ping the DB via `db.Health`
        - [ ] `200 OK` when up, `503 Service Unavailable` when DB down

- [ ] 4: Serve static files in production
  - [ ] If `cfg.Env == "production"` and `cfg.StaticDir` exists:
    - [ ] Mount `http.FileServer` at `/` to serve files under `cfg.StaticDir`
    - [ ] Add SPA fallback for unknown GET routes to `index.html`
    - [ ] Cache headers: long cache for hashed assets under `assets/`; `no-store` for `index.html`
  - [ ] In development, skip static serving (handled by Vite dev server)

- [ ] 5: Wire HTTP server in `cmd/server/main.go`
  - [ ] Initialize `slog` logger
    - [ ] JSON handler in production; text handler in development
  - [ ] Load config and run migrations when `cfg.AutoMigrate` is true (already present)
  - [ ] Open DB connection: `db.Open(ctx, cfg.DatabaseURL)`; defer `Close()`
  - [ ] Create router: `http.NewRouter(cfg, dbConn, clock.NewSystemClock(), logger)`
  - [ ] Configure `http.Server` with timeouts (`ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`)
  - [ ] Start server on `fmt.Sprintf(":%d", cfg.Port)` in a goroutine
  - [ ] Implement graceful shutdown on `SIGINT/SIGTERM` with a 10s timeout context
  - [ ] Log startup details (env, port, static enabled, migrations applied)

- [ ] 6: Local verification
  - [ ] Start Postgres (from Milestone 2): `docker compose up -d postgres`
  - [ ] Set env and run the server
    ```powershell
    cd server
    $env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/clockwork?sslmode=disable"
    $env:PORT = "8080"
    go run ./cmd/server
    ```
  - [ ] Health check returns 200 and `db: "up"`
    ```powershell
    curl http://localhost:8080/healthz
    ```
  - [ ] (Optional) Verify static serving in production mode
    ```powershell
    # Build client assets
    cd client
    npm install
    npm run build

    # Serve in production mode
    cd ../server
    $env:ENV = "production"
    $env:STATIC_DIR = "..\client\dist"
    go run ./cmd/server

    # Open browser
    Start-Process "http://localhost:8080/"
    ```

- [ ] 7: Acceptance checklist (map to Implementation Plan)
  - [ ] Server starts and logs env/port
  - [ ] `GET /healthz` returns `200 OK` with `{ ok: true, db: "up" }` when DB reachable
  - [ ] In production mode with built assets present, server serves the SPA (index and assets)

### Notes for M3
- Keep handlers minimal; business logic lands in services in later milestones
- Use strict CORS: allow `*` only in development; enumerate origins in production
- Prefer UTC for all timestamps (`clock.SystemClock` returns `time.Now().UTC()`)


