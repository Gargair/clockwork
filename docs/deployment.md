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

Clockwork can be deployed to Kubernetes using Helm charts. The Helm chart provides a complete, production-ready deployment configuration with support for development and production environments.

### Helm Chart Overview

The Helm chart is located in `deploy/helm/clockwork/` and includes:
- **Templates**: Namespace, ConfigMap, Secret, Deployment, Service, Migration Job, Ingress
- **Values files**: Default values, development overrides, production overrides
- **Flexibility**: Supports creating resources or using existing ones

**Quick start:**
```bash
# Development
helm install clockwork ./deploy/helm/clockwork \
  -f ./deploy/helm/clockwork/values-dev.yaml \
  -n clockwork \
  --create-namespace

# Production
helm install clockwork ./deploy/helm/clockwork \
  -f ./deploy/helm/clockwork/values-prod.yaml \
  -n clockwork \
  --create-namespace
```

For detailed deployment instructions, see [`deploy/README.md`](../deploy/README.md).

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

### Helm Chart Usage

**Installation:**
```bash
helm install <release-name> ./deploy/helm/clockwork \
  -f ./deploy/helm/clockwork/values-prod.yaml \
  -n <namespace> \
  --create-namespace
```

**Upgrade:**
```bash
helm upgrade <release-name> ./deploy/helm/clockwork \
  -f ./deploy/helm/clockwork/values-prod.yaml \
  -n <namespace>
```

**Uninstall:**
```bash
helm uninstall <release-name> -n <namespace>
```

**Rollback:**
```bash
helm rollback <release-name> <revision-number> -n <namespace>
```

### Development Cluster Setup

**kind (Kubernetes in Docker) - Recommended:**
```bash
# Install kind
choco install kind  # Windows
brew install kind   # macOS

# Create cluster
kind create cluster --name clockwork

# Load Docker image
kind load docker-image clockwork:latest --name clockwork
```

**minikube:**
```bash
# Install minikube
# Start cluster
minikube start

# Load Docker image
minikube image load clockwork:latest
```

### Production Deployment Checklist

Before deploying to production, ensure:

- [ ] **Image Management**
  - [ ] Image pushed to container registry (Docker Hub, GCR, ECR, ACR)
  - [ ] Image tagged with semantic version (avoid `latest` in production)
  - [ ] Image scanned for vulnerabilities

- [ ] **Secrets Management**
  - [ ] `DATABASE_URL` secret created using secure method
  - [ ] Secrets not committed to git
  - [ ] Secrets rotated regularly
  - [ ] Secret management tool integrated (if applicable)

- [ ] **Resource Configuration**
  - [ ] Resource limits appropriate for workload
  - [ ] Resource requests set for scheduling
  - [ ] Replica count set for high availability (minimum 2)
  - [ ] Horizontal Pod Autoscaler configured (optional)

- [ ] **Database**
  - [ ] Managed PostgreSQL service configured
  - [ ] Database backups enabled
  - [ ] Connection pooling configured (if needed)
  - [ ] SSL/TLS enabled for database connections

- [ ] **Networking**
  - [ ] Ingress configured with TLS/SSL
  - [ ] Certificates managed (cert-manager or manual)
  - [ ] Network policies configured (optional, for security)
  - [ ] CORS origins configured correctly

- [ ] **Migrations**
  - [ ] Migration strategy chosen (Job recommended)
  - [ ] Migration Job tested
  - [ ] Rollback plan documented

- [ ] **Monitoring & Observability**
  - [ ] Health checks configured (liveness/readiness probes)
  - [ ] Log aggregation configured
  - [ ] Metrics collection enabled
  - [ ] Alerting rules configured

- [ ] **Security**
  - [ ] Non-root user configured
  - [ ] Security context set
  - [ ] Pod Security Policies/Standards applied
  - [ ] RBAC configured (if needed)

### Secrets Management Best Practices

**Never commit secrets to git.** Use one of these methods:

1. **kubectl create secret** (manual):
   ```bash
   kubectl create secret generic clockwork-secrets \
     --from-literal=DATABASE_URL='postgres://...' \
     -n clockwork
   ```

2. **Sealed Secrets** (for GitOps):
   - Encrypt secrets for git storage
   - Automatically decrypts in cluster

3. **External Secrets Operator**:
   - Syncs from cloud secret managers
   - Supports AWS Secrets Manager, GCP Secret Manager, Azure Key Vault, HashiCorp Vault

4. **Cloud-native secret management**:
   - **GCP**: Secret Manager with Workload Identity
   - **AWS**: Secrets Manager with IAM roles
   - **Azure**: Key Vault with Managed Identity

**Secret rotation:**
- Rotate database passwords regularly
- Update secrets without redeploying (use `kubectl create secret --dry-run -o yaml | kubectl apply -f -`)
- Monitor secret age and set rotation policies

### Migration Strategy

**Job approach (recommended for production):**
- Set `migration.enabled: true` and `migration.strategy: job`
- Set `configMap.values.dbAutoMigrate: false`
- Migration Job runs before deployment (Helm pre-install hook)
- Ensures migrations run once, with proper error handling
- Can be integrated into CI/CD pipeline

**InitContainer approach:**
- Set `migration.enabled: true` and `migration.strategy: initContainer`
- Migrations run in InitContainer before main container starts
- Simpler but less control
- Migrations run on every pod start (inefficient for multiple replicas)
- **Note**: Requires migration image with goose installed

**Auto-migrate approach:**
- Set `configMap.values.dbAutoMigrate: true`
- Migrations run automatically on pod startup
- Convenient for development
- **Not recommended for production** (risk of concurrent migrations)

### Scaling Considerations

**Horizontal Pod Autoscaler (HPA):**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: clockwork-hpa
  namespace: clockwork
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: clockwork-server
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

**Resource limits:**
- Set appropriate CPU/memory limits based on workload
- Start conservative and adjust based on metrics
- Use resource requests for scheduling
- Monitor actual usage and adjust accordingly

**Pod Disruption Budget:**
```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: clockwork-pdb
  namespace: clockwork
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: clockwork
      component: server
```

### Using Existing Resources

The Helm chart supports using existing resources instead of creating new ones:

**Using existing namespace:**
```bash
helm install clockwork ./deploy/helm/clockwork \
  -f values.yaml \
  --set namespace.create=false \
  --set namespace.name=existing-namespace \
  -n existing-namespace
```

**Using existing ConfigMap:**
```bash
helm install clockwork ./deploy/helm/clockwork \
  -f values.yaml \
  --set configMap.create=false \
  --set configMap.name=existing-configmap \
  -n clockwork
```

**Using existing Secret:**
```bash
helm install clockwork ./deploy/helm/clockwork \
  -f values.yaml \
  --set secret.name=existing-secret \
  -n clockwork
```

**Using existing PostgreSQL:**
- Create secret with `DATABASE_URL` pointing to existing PostgreSQL
- Chart defaults to `postgresql.enabled: false`
- Configure connection string in secret (see PostgreSQL integration section)

This is useful for:
- Shared infrastructure (PostgreSQL managed by platform team)
- Pre-configured secrets (from secret management systems)
- Multi-tenant deployments (shared namespace with other apps)
