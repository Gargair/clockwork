# Clockwork Kubernetes Deployment

This directory contains Helm charts and deployment configurations for deploying Clockwork to Kubernetes clusters.

## Prerequisites

**Required:**
- Kubernetes cluster (v1.24+)
- `kubectl` configured to access the cluster
- Helm 3.x installed
- Docker image `clockwork:latest` built and available (or pushed to container registry)

**For development:**
- Local Kubernetes cluster: kind or minikube
- Image loaded into cluster (see below)

**For production:**
- Managed Kubernetes cluster (GKE, EKS, AKS) or self-hosted
- Container registry access (Docker Hub, GCR, ECR, ACR)
- Image pushed to registry with appropriate tags

## Helm Installation

**Windows (using Chocolatey):**
```powershell
choco install kubernetes-helm
```

**macOS (using Homebrew):**
```bash
brew install helm
```

**Linux:**
```bash
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

**Verify installation:**
```bash
helm version
```

## Development Deployment

### 1. Set up local cluster

**Option A: kind (Kubernetes in Docker)**
```powershell
# Install kind (Windows)
choco install kind

# Create cluster
kind create cluster --name clockwork

# Load Docker image into cluster
kind load docker-image clockwork:latest --name clockwork
```

**Option B: minikube**
```powershell
# Install minikube
# Start cluster
minikube start

# Load Docker image into cluster
minikube image load clockwork:latest
```

### 2. Create namespace (if not using --create-namespace)

```bash
kubectl create namespace clockwork
```

### 3. Create Secret with DATABASE_URL

**Important:** Never commit secrets to git. Create the secret manually:

```bash
kubectl create secret generic clockwork-secrets \
  --from-literal=DATABASE_URL='postgres://user:password@host:port/database' \
  -n clockwork
```

For development with in-cluster PostgreSQL:
```bash
kubectl create secret generic clockwork-secrets \
  --from-literal=DATABASE_URL='postgres://postgres:postgres@postgres-service:5432/clockwork?sslmode=disable' \
  -n clockwork
```

### 4. Deploy with Helm

```bash
helm install clockwork ./helm/clockwork \
  -f ./helm/clockwork/values-dev.yaml \
  -n clockwork \
  --create-namespace
```

### 5. Verify deployment

```bash
# Check namespace
kubectl get namespace clockwork

# Check ConfigMap
kubectl get configmap -n clockwork

# Check Secret
kubectl get secret clockwork-secrets -n clockwork

# Check Deployment
kubectl get deployment -n clockwork

# Wait for pods to be ready
kubectl wait --for=condition=available deployment/clockwork-server -n clockwork --timeout=300s

# Check pods
kubectl get pods -n clockwork

# Check logs
kubectl logs -l app=clockwork,component=server -n clockwork
```

### 6. Port forward for testing

```bash
kubectl port-forward service/clockwork-service 8080:80 -n clockwork
```

Test endpoints:
- Health: `curl http://localhost:8080/healthz`
- API: `curl http://localhost:8080/api/health`
- Static assets: `curl http://localhost:8080/`

## Production Deployment

### 1. Push image to container registry

**Docker Hub:**
```bash
docker tag clockwork:latest yourusername/clockwork:v1.0.0
docker push yourusername/clockwork:v1.0.0
```

**Google Container Registry (GCR):**
```bash
docker tag clockwork:latest gcr.io/PROJECT_ID/clockwork:v1.0.0
docker push gcr.io/PROJECT_ID/clockwork:v1.0.0
```

**Amazon ECR:**
```bash
aws ecr get-login-password --region REGION | docker login --username AWS --password-stdin ACCOUNT_ID.dkr.ecr.REGION.amazonaws.com
docker tag clockwork:latest ACCOUNT_ID.dkr.ecr.REGION.amazonaws.com/clockwork:v1.0.0
docker push ACCOUNT_ID.dkr.ecr.REGION.amazonaws.com/clockwork:v1.0.0
```

**Azure Container Registry (ACR):**
```bash
az acr login --name REGISTRY_NAME
docker tag clockwork:latest REGISTRY_NAME.azurecr.io/clockwork:v1.0.0
docker push REGISTRY_NAME.azurecr.io/clockwork:v1.0.0
```

### 2. Update values file

Edit `values-prod.yaml` to:
- Set image repository and tag
- Configure production environment variables
- Set appropriate resource limits
- Configure replica count (minimum 2 for HA)
- Set `ENV=production`
- Configure `ALLOWED_ORIGINS` with specific domains

### 3. Create Secret with production DATABASE_URL

```bash
kubectl create secret generic clockwork-secrets \
  --from-literal=DATABASE_URL='postgres://user:password@managed-postgres-host:5432/clockwork' \
  -n clockwork
```

**Note:** For production, use managed PostgreSQL services (Cloud SQL, RDS, Azure Database) with SSL/TLS enabled.

### 4. Run database migrations (if not using auto-migrate)

If `migration.enabled: true` in values:
```bash
# Wait for migration job to complete
kubectl wait --for=condition=complete job/clockwork-migrate -n clockwork --timeout=300s

# Check migration logs
kubectl logs job/clockwork-migrate -n clockwork
```

### 5. Deploy with Helm

```bash
helm install clockwork ./helm/clockwork \
  -f ./helm/clockwork/values-prod.yaml \
  -n clockwork \
  --create-namespace
```

### 6. Verify deployment

```bash
# Check all resources
kubectl get all -n clockwork

# Check pod status
kubectl get pods -n clockwork

# Check service endpoints
kubectl get endpoints clockwork-service -n clockwork

# View logs
kubectl logs -l app=clockwork,component=server -n clockwork -f
```

## Using Existing Resources

### Using existing namespace

```bash
helm install clockwork ./helm/clockwork \
  -f ./helm/clockwork/values.yaml \
  --set namespace.create=false \
  --set namespace.name=existing-namespace \
  -n existing-namespace
```

### Using existing ConfigMap

```bash
helm install clockwork ./helm/clockwork \
  -f ./helm/clockwork/values.yaml \
  --set configMap.create=false \
  --set configMap.name=existing-configmap \
  -n clockwork
```

### Using existing Secret

```bash
helm install clockwork ./helm/clockwork \
  -f ./helm/clockwork/values.yaml \
  --set secret.name=existing-secret \
  -n clockwork
```

**Note:** The chart defaults to `secret.create: false`, so you only need to set `secret.name` if it differs from the default `clockwork-secrets`.

### Using existing PostgreSQL

The chart defaults to `postgresql.enabled: false`, assuming PostgreSQL exists externally. This is the recommended approach for production deployments.

#### Connection String Format

The `DATABASE_URL` must be provided in a Kubernetes Secret. The connection string format is:

```
postgres://[user]:[password]@[host]:[port]/[database][?parameters]
```

**Examples:**

**Basic connection:**
```
postgres://postgres:mypassword@postgres.example.com:5432/clockwork
```

**With SSL (recommended for production):**
```
postgres://postgres:mypassword@postgres.example.com:5432/clockwork?sslmode=require
```

**With SSL and certificate verification:**
```
postgres://postgres:mypassword@postgres.example.com:5432/clockwork?sslmode=verify-full&sslcert=/path/to/cert&sslkey=/path/to/key&sslrootcert=/path/to/ca-cert
```

**Connection string parameters:**
- `sslmode`: `disable`, `allow`, `prefer`, `require`, `verify-ca`, `verify-full`
- `connect_timeout`: Connection timeout in seconds
- `application_name`: Application identifier (optional)

#### Creating the Secret

**1. Create secret with DATABASE_URL:**

```bash
kubectl create secret generic clockwork-secrets \
  --from-literal=DATABASE_URL='postgres://user:password@host:5432/clockwork?sslmode=require' \
  -n clockwork
```

**2. Verify the secret:**

```bash
# View secret (base64 encoded)
kubectl get secret clockwork-secrets -n clockwork -o jsonpath='{.data.DATABASE_URL}' | base64 -d

# Or decode on Windows PowerShell:
[System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String((kubectl get secret clockwork-secrets -n clockwork -o jsonpath='{.data.DATABASE_URL}')))
```

**3. Deploy with existing secret:**

The chart automatically uses the secret specified in `postgresql.existingSecret` (default: `clockwork-secrets`). No additional configuration needed if using the default name.

```bash
helm install clockwork ./helm/clockwork \
  -f ./helm/clockwork/values-prod.yaml \
  -n clockwork \
  --create-namespace
```

#### Managed PostgreSQL Services

**Google Cloud SQL:**

1. **Using Cloud SQL Proxy (recommended):**
   - Deploy Cloud SQL Proxy as a sidecar or separate deployment
   - Use connection string: `postgres://user:password@127.0.0.1:5432/clockwork`
   - Configure with Workload Identity for authentication

2. **Using Private IP:**
   - Enable Private IP on Cloud SQL instance
   - Configure VPC peering between GKE and Cloud SQL
   - Use connection string: `postgres://user:password@10.x.x.x:5432/clockwork?sslmode=require`

3. **Using Public IP (less secure):**
   - Enable authorized networks in Cloud SQL
   - Add cluster node IPs to authorized networks
   - Use connection string: `postgres://user:password@PUBLIC_IP:5432/clockwork?sslmode=require`

**Amazon RDS:**

1. **Configure security groups:**
   - Allow inbound traffic from Kubernetes cluster on port 5432
   - Use cluster security group or node security groups

2. **Connection string:**
   ```
   postgres://user:password@your-instance.region.rds.amazonaws.com:5432/clockwork?sslmode=require
   ```

3. **Using IAM database authentication (optional):**
   - Configure RDS for IAM authentication
   - Use IAM role for database access
   - Connection string format differs (see AWS documentation)

**Azure Database for PostgreSQL:**

1. **Configure firewall rules:**
   - Add Kubernetes cluster IP ranges to firewall rules
   - Or enable "Allow access to Azure services"

2. **Connection string:**
   ```
   postgres://user@servername:password@servername.postgres.database.azure.com:5432/clockwork?sslmode=require
   ```

3. **Using Managed Identity (recommended):**
   - Configure Azure AD authentication
   - Use Managed Identity for passwordless authentication

#### SSL/TLS Requirements

**For production deployments, always use SSL/TLS:**

1. **Minimum requirement:** `sslmode=require`
   - Encrypts connection but doesn't verify certificate
   - Suitable for managed services with trusted certificates

2. **Recommended:** `sslmode=verify-full`
   - Encrypts connection and verifies server certificate
   - Requires CA certificate (provided by managed service)

3. **Certificate files:**
   - For `verify-full`, you may need to mount certificate files
   - Managed services typically provide CA certificates
   - Mount certificates as ConfigMap or Secret and reference in connection string

**Example with certificate verification:**

```bash
# Create ConfigMap with CA certificate
kubectl create configmap postgres-ca-cert \
  --from-file=ca-cert.pem=/path/to/ca-cert.pem \
  -n clockwork

# Update deployment to mount certificate
# Then use connection string: postgres://...?sslmode=verify-full&sslrootcert=/etc/ssl/certs/ca-cert.pem
```

#### Testing Database Connectivity

After deployment, verify database connectivity:

```bash
# Check pod logs for database connection
kubectl logs -l app=clockwork,component=server -n clockwork | grep -i database

# Test from within a pod
kubectl exec -it deployment/clockwork-server -n clockwork -- /app/healthcheck http://localhost:8080/healthz

# Check health endpoint (includes database status)
curl http://localhost:8080/healthz
# (after port-forwarding: kubectl port-forward service/clockwork-service 8080:80 -n clockwork)
```

## Secrets Management

**Never commit secrets to git.** Use one of these methods:

### 1. kubectl create secret (manual)

```bash
kubectl create secret generic clockwork-secrets \
  --from-literal=DATABASE_URL='postgres://...' \
  -n clockwork
```

### 2. Sealed Secrets (for GitOps)

Install Sealed Secrets Controller, then:
```bash
kubectl create secret generic clockwork-secrets \
  --from-literal=DATABASE_URL='postgres://...' \
  --dry-run=client -o yaml | kubeseal -o yaml > sealed-secret.yaml
```

### 3. External Secrets Operator

Configure External Secrets Operator to sync from:
- AWS Secrets Manager
- Google Secret Manager
- Azure Key Vault
- HashiCorp Vault

### 4. Cloud-native secret management

- **GCP**: Use Secret Manager with Workload Identity
- **AWS**: Use Secrets Manager with IAM roles
- **Azure**: Use Key Vault with Managed Identity

## Migration Execution

### Auto-migrate (DB_AUTO_MIGRATE=true)

Migrations run automatically on pod startup. This is convenient but less control.

### Migration Job (recommended for production)

Set `migration.enabled: true` and `configMap.values.dbAutoMigrate: false` in values:

```bash
# Deploy with migration job
helm install clockwork ./helm/clockwork -f values-prod.yaml -n clockwork

# Wait for migration to complete
kubectl wait --for=condition=complete job/clockwork-migrate -n clockwork --timeout=300s

# Check migration logs
kubectl logs job/clockwork-migrate -n clockwork

# Then deploy application
helm upgrade clockwork ./helm/clockwork -f values-prod.yaml -n clockwork
```

### Manual migration

Run migrations manually using a one-off job:

```bash
kubectl create job --from=cronjob/clockwork-migrate clockwork-migrate-manual -n clockwork
```

Or create a manual job:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: clockwork-migrate-manual
  namespace: clockwork
spec:
  template:
    spec:
      containers:
      - name: migrate
        image: clockwork:latest
        command: ["/app/server", "--migrate"]
        envFrom:
        - configMapRef:
            name: clockwork-config
        - secretRef:
            name: clockwork-secrets
      restartPolicy: Never
```

## Upgrade Instructions

```bash
# Upgrade with new values
helm upgrade clockwork ./helm/clockwork \
  -f ./helm/clockwork/values-prod.yaml \
  -n clockwork

# Upgrade with specific image tag
helm upgrade clockwork ./helm/clockwork \
  -f ./helm/clockwork/values-prod.yaml \
  --set deployment.image.tag=v1.0.1 \
  -n clockwork

# Check upgrade status
helm status clockwork -n clockwork

# Watch rollout
kubectl rollout status deployment/clockwork-server -n clockwork
```

## Rollback Procedures

```bash
# List release history
helm history clockwork -n clockwork

# Rollback to previous revision
helm rollback clockwork -n clockwork

# Rollback to specific revision
helm rollback clockwork <revision-number> -n clockwork

# Rollback Kubernetes deployment directly
kubectl rollout undo deployment/clockwork-server -n clockwork
```

## Uninstall Instructions

```bash
# Uninstall Helm release
helm uninstall clockwork -n clockwork

# Delete namespace (if created by Helm)
kubectl delete namespace clockwork

# Note: Secrets and PVCs are not automatically deleted for safety
# Manually delete if needed:
kubectl delete secret clockwork-secrets -n clockwork
```

## Troubleshooting

### Pods not starting

```bash
# Check pod status
kubectl get pods -n clockwork

# Describe pod for events
kubectl describe pod <pod-name> -n clockwork

# Check logs
kubectl logs <pod-name> -n clockwork

# Check previous container logs (if crashed)
kubectl logs <pod-name> -n clockwork --previous
```

### Database connection issues

```bash
# Verify Secret exists
kubectl get secret clockwork-secrets -n clockwork

# Check DATABASE_URL format (base64 encoded)
kubectl get secret clockwork-secrets -n clockwork -o jsonpath='{.data.DATABASE_URL}' | base64 -d

# Test database connectivity from pod
kubectl exec -it <pod-name> -n clockwork -- /app/healthcheck http://localhost:8080/healthz
```

### Image pull errors

```bash
# Verify image exists
docker images clockwork:latest

# For kind: reload image
kind load docker-image clockwork:latest --name clockwork

# For minikube: reload image
minikube image load clockwork:latest

# For production: verify registry credentials
kubectl get secret <registry-secret> -n clockwork
```

### Service not accessible

```bash
# Check service
kubectl get service clockwork-service -n clockwork

# Check endpoints
kubectl get endpoints clockwork-service -n clockwork

# Check ingress (if configured)
kubectl get ingress -n clockwork

# Port forward for testing
kubectl port-forward service/clockwork-service 8080:80 -n clockwork
```

### Migration failures

```bash
# Check migration job status
kubectl get job clockwork-migrate -n clockwork

# View migration logs
kubectl logs job/clockwork-migrate -n clockwork

# Check migration state in database
kubectl exec -it <postgres-pod> -n clockwork -- psql -U postgres -d clockwork -c "SELECT * FROM goose_db_version;"
```

### Resource constraints

```bash
# Check resource usage
kubectl top pods -n clockwork

# Check resource limits
kubectl describe pod <pod-name> -n clockwork | grep -A 5 "Limits"

# Adjust resources in values.yaml and upgrade
helm upgrade clockwork ./helm/clockwork -f values-prod.yaml -n clockwork
```

## Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Helm Documentation](https://helm.sh/docs/)
- [Clockwork Deployment Guide](../docs/deployment.md)
- [Development Guide](../docs/development.md)

