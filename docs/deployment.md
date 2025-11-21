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

## Kubernetes Deployment

### Prerequisites

**Required:**
- Kubernetes cluster (v1.24+)
- `kubectl` configured to access the cluster
- Helm 3.x installed
- Docker image `clockwork:latest` built and available (or pushed to container registry)

**For development:**
- Local Kubernetes cluster: kind or minikube
- Image loaded into cluster (kind: `kind load docker-image clockwork:latest`, minikube: `minikube image load clockwork:latest`)

**For production:**
- Managed Kubernetes cluster (GKE, EKS, AKS) or self-hosted
- Container registry access (Docker Hub, GCR, ECR, ACR)
- Image pushed to registry with appropriate tags

### Server Configuration Requirements

The application requires the following environment variables:

**Required:**
- `DATABASE_URL` (required): PostgreSQL connection string
  - Format: `postgres://user:password@host:port/database`
  - Must be provided via Kubernetes Secret (never in ConfigMap)

**Optional (with defaults):**
- `DB_AUTO_MIGRATE` (default `false`): Run migrations on startup
  - Set to `true` for automatic migrations (convenient but less control)
  - Set to `false` for manual migrations via Job (recommended for production)
- `MIGRATIONS_DIR` (default `server/migrations`): Path to SQL migrations
  - In container: `/app/migrations` (set by Dockerfile)
- `PORT` (default `8080`): HTTP port to bind
  - Container exposes port 8080
- `ENV` (default `development`): Environment mode
  - Values: `development` or `production`
  - Set to `production` for production deployments
- `STATIC_DIR` (default `client/dist`): Path to built client assets
  - In container: `/app/static` (set by Dockerfile)
- `ALLOWED_ORIGINS` (CSV): CORS allowed origins
  - Default: `*` in development when not set
  - In production: Specify exact origins (e.g., `https://app.example.com,https://api.example.com`)

### Deployment Strategy

**Development:**
- **kind (Kubernetes in Docker)**: Recommended for local testing
  - Install: `choco install kind` (Windows) or `brew install kind` (Mac)
  - Create cluster: `kind create cluster --name clockwork`
  - Load image: `kind load docker-image clockwork:latest --name clockwork`
- **minikube**: Alternative local cluster option
  - Install minikube
  - Start: `minikube start`
  - Load image: `minikube image load clockwork:latest`

**Production:**
- **Managed Kubernetes** (recommended):
  - Google Kubernetes Engine (GKE)
  - Amazon Elastic Kubernetes Service (EKS)
  - Azure Kubernetes Service (AKS)
- **Self-hosted Kubernetes**: For organizations with existing infrastructure

### PostgreSQL Deployment Options

**Option 1: Managed PostgreSQL Service (Recommended for Production)**
- Use cloud-managed PostgreSQL:
  - Google Cloud SQL
  - Amazon RDS for PostgreSQL
  - Azure Database for PostgreSQL
- Benefits:
  - Automated backups and point-in-time recovery
  - High availability and automatic failover
  - Managed updates and patches
  - Monitoring and alerting
  - SSL/TLS encryption
- Configuration:
  - Create Kubernetes Secret with `DATABASE_URL` pointing to managed service
  - Configure network policies/security groups for cluster access
  - Use connection pooling for high-traffic scenarios

**Option 2: In-Cluster StatefulSet (For Development/Testing)**
- Deploy PostgreSQL as StatefulSet with PersistentVolumeClaim
- Benefits:
  - Self-contained cluster (no external dependencies)
  - Useful for development and testing environments
- Considerations:
  - Requires persistent storage
  - Manual backup and maintenance
  - Not recommended for production workloads
- Can use Bitnami PostgreSQL Helm chart or similar

**Decision:**
- **Development**: Option 2 (in-cluster) for simplicity
- **Production**: Option 1 (managed service) for reliability and operational excellence

### Kubernetes Resources

**Required resources:**
- **Namespace**: Isolated environment for the application (default: `clockwork`)
- **ConfigMap**: Non-sensitive configuration (PORT, ENV, MIGRATIONS_DIR, STATIC_DIR, ALLOWED_ORIGINS)
- **Secret**: Sensitive configuration (DATABASE_URL)
- **Deployment**: Application pods with replicas, resource limits, health checks
- **Service**: Internal/external access to the application

**Optional resources:**
- **Job**: Database migration execution (if not using auto-migrate)
- **Ingress**: External access with TLS/SSL termination
- **HorizontalPodAutoscaler**: Automatic scaling based on metrics
- **PodDisruptionBudget**: Ensure availability during cluster maintenance

### Database Migrations

**Strategy options:**
1. **Auto-migrate** (`DB_AUTO_MIGRATE=true`):
   - Migrations run automatically on pod startup
   - Convenient but less control
   - Risk of multiple pods running migrations simultaneously
   - Suitable for development or single-replica deployments

2. **Migration Job** (recommended for production):
   - Set `DB_AUTO_MIGRATE=false`
   - Run migrations as separate Kubernetes Job before deployment
   - Ensures migrations run once, with proper error handling
   - Can be integrated into CI/CD pipeline

3. **InitContainer** (alternative):
   - Run migrations in InitContainer before main container starts
   - Automatic but less control over migration execution
   - Migrations run on every pod start (inefficient for multiple replicas)

**Migration state:**
- Stored in PostgreSQL database via goose migration tool
- Tracks applied migrations in `goose_db_version` table

### Environments

**Development:**
- Local Kubernetes cluster (kind/minikube)
- Single replica for resource efficiency
- Auto-migrate enabled for convenience
- Development configuration values

**Production:**
- Managed Kubernetes cluster
- Multiple replicas for high availability (minimum 2)
- Manual migrations via Job
- Production configuration values
- Resource limits and requests configured
- Monitoring and alerting enabled

