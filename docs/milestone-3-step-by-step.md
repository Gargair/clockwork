## Milestone 3: Step-by-step implementation plan

 - [x] 1: Extend server configuration (`internal/config`)
  - [x] Add fields to `Config` with sensible defaults and validation:
    - [x] `Port int` (env `PORT`, default `8080`)
    - [x] `Env string` (env `ENV`, values: `development|production`, default `development`)
    - [x] `StaticDir string` (env `STATIC_DIR`, default `client/dist`)
    - [x] `AllowedOrigins []string` (env `ALLOWED_ORIGINS`, CSV; default `*` in dev, empty in prod)
  - [x] Update `Load()` to parse and validate new fields (port > 0, known env)
  - [x] Document env vars in `docs/development.md` briefly

 - [x] 2: Add a clock abstraction (`internal/clock`)
  - [x] Define `type Clock interface { Now() time.Time }`
  - [x] Implement `type SystemClock struct{}` with `Now()` returning `time.Now().UTC()`
  - [x] Provide a `NewSystemClock()` constructor for clarity

- [x] 3: Add HTTP router and middleware (`internal/http`)
  - [x] Add dependencies:
    - [x] `github.com/go-chi/chi/v5`
    - [x] `github.com/go-chi/cors`
  - [x] Implement `func NewRouter(cfg config.Config, dbConn *sql.DB, clk clock.Clock, logger *slog.Logger) http.Handler`
    - [x] Base router: `chi.NewRouter()`
    - [x] Middleware:
      - [x] `middleware.RequestID` (unique per request)
      - [x] `middleware.RealIP`
      - [x] `middleware.Recoverer`
      - [x] Request logging middleware using `slog` (method, path, status, duration, request ID)
      - [x] `cors.Handler` with:
        - [x] `AllowedOrigins`: from `cfg.AllowedOrigins` (default `*` in dev, explicit in prod)
        - [x] `AllowedMethods`: `GET,POST,PUT,PATCH,DELETE,OPTIONS`
        - [x] `AllowedHeaders`: common headers (`Accept, Authorization, Content-Type, X-Request-ID`)
        - [x] `ExposedHeaders`: `X-Request-ID`
        - [x] `AllowCredentials`: false (tighten later as needed)
        - [x] `MaxAge`: 300
    - [x] Routes:
      - [x] `GET /healthz`: returns JSON `{ "ok": true, "db": "up|down", "time": <utc now> }`
        - [x] Use a short timeout (e.g., 1s) to ping the DB via `db.Health`
        - [x] `200 OK` when up, `503 Service Unavailable` when DB down

- [x] 4: Serve static files in production
  - [x] If `cfg.Env == "production"` and `cfg.StaticDir` exists:
    - [x] Mount file server at `/` to serve files under `cfg.StaticDir`
    - [x] Add SPA fallback for unknown GET routes to `index.html`
    - [x] Cache headers: long cache for hashed assets under `assets/`; `no-store` for `index.html`
  - [x] In development, skip static serving (handled by Vite dev server)

- [x] 5: Wire HTTP server in `cmd/server/main.go`
  - [x] Initialize `slog` logger
    - [x] JSON handler in production; text handler in development
  - [x] Load config and run migrations when `cfg.AutoMigrate` is true (already present)
  - [x] Open DB connection: `db.Open(ctx, cfg.DatabaseURL)`; defer `Close()`
  - [x] Create router: `http.NewRouter(cfg, dbConn, clock.NewSystemClock(), logger)`
  - [x] Configure `http.Server` with timeouts (`ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`)
  - [x] Start server on `fmt.Sprintf(":%d", cfg.Port)` in a goroutine
  - [x] Implement graceful shutdown on `SIGINT/SIGTERM` with a 10s timeout context
  - [x] Log startup details (env, port, static enabled, migrations applied)

- [X] 6: Local verification
  - [X] Start Postgres (from Milestone 2): `docker compose up -d postgres`
  - [X] Set env and run the server
    ```powershell
    cd server
    $env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/clockwork?sslmode=disable"
    $env:PORT = "8080"
    go run ./cmd/server
    ```
  - [X] Health check returns 200 and `db: "up"`
    ```powershell
    curl http://localhost:8080/healthz
    ```
  - [X] (Optional) Verify static serving in production mode
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

- [X] 7: Acceptance checklist (map to Implementation Plan)
  - [X] Server starts and logs env/port
  - [X] `GET /healthz` returns `200 OK` with `{ ok: true, db: "up" }` when DB reachable
  - [X] In production mode with built assets present, server serves the SPA (index and assets)

### Notes for M3
- Keep handlers minimal; business logic lands in services in later milestones
- Use strict CORS: allow `*` only in development; enumerate origins in production
- Prefer UTC for all timestamps (`clock.SystemClock` returns `time.Now().UTC()`)


