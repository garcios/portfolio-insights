# Portfolio Insights Helm Chart

A Helm chart for deploying the Portfolio Insights application to Kubernetes.

## Overview

This chart deploys a complete microservices-based portfolio management system with the following components:

### Application Services
- **gateway** - GraphQL API gateway (port 8080)
- **user-service** - User management service (gRPC: 50051)
- **portfolio-service** - Portfolio management service (gRPC: 50052)
- **transaction-service** - Transaction management service (gRPC: 50053, HTTP: 8081)
- **marketdata-service** - Market data service (gRPC: 50054)

### Infrastructure Services
- **postgres** - PostgreSQL database (port 5432)
- **nats** - NATS messaging with JetStream (ports 4222, 8222)
- **redis** - Redis cache (port 6379)
- **minio** - MinIO object storage (ports 9000, 9001)

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- PV provisioner support in the underlying infrastructure (for persistent volumes)

## Installation

### Install the chart

```bash
helm install portfolio-insights ./deployments/helm-charts/portfolio-insights
```

### Install with custom values

```bash
helm install portfolio-insights ./deployments/helm-charts/portfolio-insights \
  --set services.postgres.persistence.size=10Gi \
  --set services.gateway.ingress.hosts[0].host=myapp.example.com
```

### Install from a values file

```bash
helm install portfolio-insights ./deployments/helm-charts/portfolio-insights \
  -f custom-values.yaml
```

## Uninstallation

```bash
helm uninstall portfolio-insights
```

## Configuration

The following table lists the configurable parameters and their default values.

### Global Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `global.imagePullPolicy` | Image pull policy for all services | `IfNotPresent` |

### Service-Specific Parameters

Each service follows this structure in `values.yaml`:

```yaml
services:
  <service-name>:
    enabled: true
    image:
      repository: <image-repo>
      tag: <image-tag>
    ports: [...]
    env: {...}
    secrets: {...}
    persistence: {...}
    resources: {...}
```

### PostgreSQL Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `services.postgres.enabled` | Enable PostgreSQL | `true` |
| `services.postgres.image.repository` | PostgreSQL image | `postgres` |
| `services.postgres.image.tag` | PostgreSQL version | `16-alpine` |
| `services.postgres.env.POSTGRES_USER` | Database user | `garcios` |
| `services.postgres.env.POSTGRES_DB` | Database name | `portfolio` |
| `services.postgres.secrets.POSTGRES_PASSWORD` | Database password | `Password123` |
| `services.postgres.persistence.enabled` | Enable persistence | `true` |
| `services.postgres.persistence.size` | PVC size | `1Gi` |

### Gateway Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `services.gateway.enabled` | Enable gateway | `true` |
| `services.gateway.image.repository` | Gateway image | `portfolio-insights/gateway` |
| `services.gateway.image.tag` | Gateway version | `latest` |
| `services.gateway.ingress.enabled` | Enable ingress | `true` |
| `services.gateway.ingress.hosts[0].host` | Ingress hostname | `gateway.local` |

### Redis Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `services.redis.enabled` | Enable Redis | `true` |
| `services.redis.persistence.enabled` | Enable persistence | `true` |
| `services.redis.persistence.size` | PVC size | `1Gi` |

### MinIO Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `services.minio.enabled` | Enable MinIO | `true` |
| `services.minio.env.MINIO_ROOT_USER` | MinIO admin user | `minioadmin` |
| `services.minio.secrets.MINIO_ROOT_PASSWORD` | MinIO admin password | `minioadmin` |
| `services.minio.persistence.size` | PVC size | `5Gi` |

## Accessing the Application

### Via Ingress (Gateway)

If ingress is enabled for the gateway service, you can access it at:

```
http://gateway.local
```

Make sure to add the hostname to your `/etc/hosts` file if testing locally:

```bash
echo "127.0.0.1 gateway.local" | sudo tee -a /etc/hosts
```

### Via Port Forwarding

Forward the gateway service to your local machine:

```bash
kubectl port-forward svc/gateway 8080:8080
```

Then access at: `http://localhost:8080`

### Via LoadBalancer (Cloud)

Change the gateway service type to LoadBalancer:

```yaml
services:
  gateway:
    service:
      type: LoadBalancer
```

## Persistence

The following services use persistent volumes:

- **postgres** - Database data (`/var/lib/postgresql/data`)
- **redis** - Cache data (`/data`)
- **minio** - Object storage (`/data`)

By default, PVCs are created with the default storage class. To use a specific storage class:

```yaml
services:
  postgres:
    persistence:
      storageClass: "fast-ssd"
```

## Secrets Management

Sensitive values are stored in Kubernetes Secrets. For production deployments, consider:

1. **Using external secret management:**
   - HashiCorp Vault
   - AWS Secrets Manager
   - Azure Key Vault
   - Google Secret Manager

2. **Encrypting secrets at rest:**
   - Enable Kubernetes encryption at rest
   - Use sealed-secrets or similar tools

3. **Updating default passwords:**

```bash
helm install portfolio-insights ./deployments/helm-charts/portfolio-insights \
  --set services.postgres.secrets.POSTGRES_PASSWORD=<secure-password> \
  --set services.minio.secrets.MINIO_ROOT_PASSWORD=<secure-password>
```

## Resource Limits

By default, no resource limits are set. For production, configure appropriate limits:

```yaml
services:
  gateway:
    resources:
      requests:
        memory: "256Mi"
        cpu: "250m"
      limits:
        memory: "512Mi"
        cpu: "500m"
```

## Health Checks

All services include:
- **Liveness probes** - Restart unhealthy containers
- **Readiness probes** - Route traffic only to ready containers

Default configuration:
- Initial delay: 5-10 seconds
- Period: 10 seconds
- Probe type: TCP socket check

## Service Dependencies

The chart includes dependency annotations but does not enforce startup ordering (Kubernetes does not support native ordering). Services should implement retry logic for dependent services.

Documented dependencies:
- `gateway` → `user-service`, `portfolio-service`, `transaction-service`
- `user-service` → `postgres`
- `portfolio-service` → `postgres`, `nats`, `redis`, `marketdata-service`
- `marketdata-service` → `postgres`, `minio`
- `transaction-service` → `postgres`, `nats`

## Monitoring

Metrics endpoints are exposed on the following ports:
- `user-service`: 9096
- `portfolio-service`: 9098
- `transaction-service`: 9097
- `marketdata-service`: 9099

These can be scraped by Prometheus or other monitoring tools.

## Troubleshooting

### Check pod status

```bash
kubectl get pods
```

### View pod logs

```bash
kubectl logs <pod-name>
```

### Describe a pod

```bash
kubectl describe pod <pod-name>
```

### Check persistent volumes

```bash
kubectl get pvc
kubectl get pv
```

### Common Issues

**Pods stuck in Pending:**
- Check if PVCs can be provisioned
- Verify sufficient cluster resources

**ImagePullBackOff:**
- Ensure images are built and available
- Check image repository and tag names

**CrashLoopBackOff:**
- Check pod logs for application errors
- Verify environment variables and secrets
- Ensure dependent services are ready

## Development

### Building Images

Before deploying, build all service images:

```bash
# From repository root
docker build -t portfolio-insights/gateway:latest -f apps/gateway/Dockerfile .
docker build -t portfolio-insights/user-service:latest -f services/user-service/Dockerfile .
docker build -t portfolio-insights/portfolio-service:latest -f services/portfolio-service/Dockerfile .
docker build -t portfolio-insights/transaction-service:latest -f services/transaction-service/Dockerfile .
docker build -t portfolio-insights/marketdata-service:latest -f services/marketdata-service/Dockerfile .
```

### Testing the Chart

Render templates without installing:

```bash
helm template portfolio-insights ./deployments/helm-charts/portfolio-insights
```

Dry run installation:

```bash
helm install portfolio-insights ./deployments/helm-charts/portfolio-insights --dry-run --debug
```

### Upgrading

After making changes to the chart:

```bash
helm upgrade portfolio-insights ./deployments/helm-charts/portfolio-insights
```


## Support

For issues and questions, please open an issue in the repository.
