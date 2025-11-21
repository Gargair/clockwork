# Deployment

## Containerization

The application is containerized using a multi-stage Dockerfile that produces a minimal, secure production image.

### Docker Build Process

**Build the image:**
```powershell
scripts/docker-build.ps1
# Or manually:
docker build -t clockwork:latest .
```

**Build stages:**
1. **Server builder**: Compiles the Go server as a statically linked binary
   - Uses `golang:1.25-alpine` as base
   - Builds with `CGO_ENABLED=0` for pure Go binary
   - Strips debug symbols with `-ldflags '-w -s'` to reduce size
   - Also builds a healthcheck binary for container health checks

2. **Client builder**: Builds the React SPA
   - Uses `node:25-alpine` as base
   - Runs `npm ci` for reproducible builds
   - Produces optimized production build in `dist/`

3. **Runtime image**: Minimal final image
   - Uses `gcr.io/distroless/static:nonroot` (no shell, no package manager)
   - Includes CA certificates for TLS verification
   - Runs as non-root user automatically
   - Contains only: server binary, client assets, migrations, healthcheck binary

**Image size optimization:**
- Multi-stage builds eliminate build dependencies from final image
- Static linking enables use of distroless base (no OS overhead)
- Debug symbols stripped (`-ldflags '-w -s'`)
- Client assets are pre-built and minified by Vite
- Target size: < 60MB for server binary + client assets

### Docker Compose Usage

**Development/Testing:**
```powershell
# Start all services
docker compose up -d

# View logs
docker compose logs -f

# Stop services
docker compose down
```

**Configuration:**
- Environment variables are set in `docker-compose.yml`
- Override with `.env` file or environment variables
- Port mappings use environment variable defaults (e.g., `SERVER_PORT`, `POSTGRES_PORT`)

**Health checks:**
- Postgres: Uses `pg_isready` command
- Server: Uses custom `/app/healthcheck` binary (checks `/healthz` endpoint)
- Services wait for dependencies to be healthy before starting

### Production Considerations

**Security:**
- **Non-root user**: Image runs as non-root user (distroless handles this automatically)
- **Minimal attack surface**: No shell, no package manager, no unnecessary tools
- **Secrets management**: Never hardcode secrets in docker-compose.yml or Dockerfile
  - Use Docker secrets, Kubernetes secrets, or environment variable injection
  - Consider using `.env` files (excluded from git) or secret management services

**Configuration:**
- Set `ENV=production` for production deployments
- Configure `ALLOWED_ORIGINS` with specific domains (avoid `*` in production)
- Use managed PostgreSQL services when possible
- Set appropriate resource limits (CPU, memory) in orchestration platform

**Migrations:**
- **Auto-migrate**: Convenient but less control
  - Set `DB_AUTO_MIGRATE=true` in production if acceptable
  - Migrations run on every container start
- **Manual migrations**: Recommended for production
  - Set `DB_AUTO_MIGRATE=false`
  - Run migrations as a separate job or init container
  - Use `scripts/docker-migrate.ps1` or similar tooling
  - Ensures migrations run once, with proper rollback procedures

**Static assets:**
- Client SPA is built into the image at `/app/static`
- Server serves static files when `ENV=production`
- For high-traffic scenarios, consider serving from CDN/object storage
- Update `STATIC_DIR` environment variable if using external storage

**Monitoring and observability:**
- Health check endpoint: `GET /healthz`
- Logs: Configure log aggregation (e.g., ELK, Loki, CloudWatch)
- Metrics: Add Prometheus metrics endpoint if needed
- Tracing: Consider distributed tracing for production debugging

## Kubernetes
- Deployment + Service for the server
- ConfigMap/Secret for configuration (include `DATABASE_URL`)
- PostgreSQL database
  - Prefer managed PostgreSQL in production
  - Alternatively run in-cluster via StatefulSet with PersistentVolumeClaim

## Database migrations
- Apply schema migrations on startup or via a separate job/tool
- Store migration state in the PostgreSQL database

## Environments
- dev: local cluster (kind/minikube)
- prod: managed Kubernetes

