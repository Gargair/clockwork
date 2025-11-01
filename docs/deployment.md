# Deployment

## Containerization
- Build a container image for the Go server
- Include static client build in the image or serve client from an external object store/CDN

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

