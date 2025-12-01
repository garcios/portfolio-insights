# Auth Server Helm Chart

This Helm chart deploys an Ory Hydra OAuth2/OIDC server with a custom login and consent provider.

## Overview

This chart includes the following components:

- **PostgreSQL**: Database for Hydra
- **Hydra Admin API**: Internal API for managing OAuth2 clients and configuration
- **Hydra Public API**: Public-facing OAuth2/OIDC endpoints
- **Login & Consent Provider**: Custom UI for user authentication and consent
- **Hydra Migration Job**: Database migration job that runs before deployment

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- PersistentVolume provisioner support in the underlying infrastructure (optional, for PostgreSQL persistence)

## Installation

### Quick Start

1. **Clone or navigate to the chart directory:**

```bash
cd deployments/helm-charts/auth-server
```

2. **Install the chart with default values:**

```bash
helm install auth-server . --namespace auth --create-namespace
```

3. **Install with custom values:**

```bash
helm install auth-server . --namespace auth --create-namespace -f custom-values.yaml
```

### Customizing the Installation

Create a `custom-values.yaml` file to override default values:

```yaml
# Example custom-values.yaml

# PostgreSQL configuration
postgres:
  auth:
    username: myuser
    password: mypassword
    database: hydra
  persistence:
    enabled: true
    size: 2Gi
    storageClass: "standard"

# Hydra secrets (IMPORTANT: Change these in production!)
secrets:
  hydra:
    system: "your-secure-random-string-32-chars"
    cookie: "your-secure-random-string-32-chars"
    oidcPairwiseSalt: "your-secure-random-string-32-chars"
  sessionSecret: "your-session-secret-32-chars"

# Login Consent Provider
loginConsentProvider:
  image:
    repository: your-registry.com/login-consent-provider
    tag: "v1.0.0"
  config:
    databaseUrl: "postgres://user:pass@external-postgres:5432/portfolio?sslmode=disable"

# Enable ingress for external access
hydraPublic:
  ingress:
    enabled: true
    className: "nginx"
    hosts:
      - host: auth.yourdomain.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: auth-tls
        hosts:
          - auth.yourdomain.com

loginConsentProvider:
  ingress:
    enabled: true
    className: "nginx"
    hosts:
      - host: login.yourdomain.com
        paths:
          - path: /
            pathType: Prefix
```

## Configuration

### Key Configuration Options

#### PostgreSQL

| Parameter | Description | Default |
|-----------|-------------|---------|
| `postgres.enabled` | Enable PostgreSQL deployment | `true` |
| `postgres.image.repository` | PostgreSQL image repository | `postgres` |
| `postgres.image.tag` | PostgreSQL image tag | `16-alpine` |
| `postgres.auth.username` | PostgreSQL username | `hydra` |
| `postgres.auth.password` | PostgreSQL password | `hydra_secret` |
| `postgres.auth.database` | PostgreSQL database name | `hydra` |
| `postgres.persistence.enabled` | Enable persistent storage | `true` |
| `postgres.persistence.size` | PVC size | `1Gi` |
| `postgres.persistence.storageClass` | Storage class | `""` |

#### Hydra Admin

| Parameter | Description | Default |
|-----------|-------------|---------|
| `hydraAdmin.enabled` | Enable Hydra Admin deployment | `true` |
| `hydraAdmin.replicaCount` | Number of replicas | `1` |
| `hydraAdmin.image.repository` | Hydra image repository | `oryd/hydra` |
| `hydraAdmin.image.tag` | Hydra image tag | `v2.2.0` |
| `hydraAdmin.service.type` | Service type | `ClusterIP` |
| `hydraAdmin.service.port` | Service port | `4445` |

#### Hydra Public

| Parameter | Description | Default |
|-----------|-------------|---------|
| `hydraPublic.enabled` | Enable Hydra Public deployment | `true` |
| `hydraPublic.replicaCount` | Number of replicas | `1` |
| `hydraPublic.image.repository` | Hydra image repository | `oryd/hydra` |
| `hydraPublic.image.tag` | Hydra image tag | `v2.2.0` |
| `hydraPublic.service.type` | Service type | `ClusterIP` |
| `hydraPublic.service.port` | Service port | `4444` |
| `hydraPublic.ingress.enabled` | Enable ingress | `false` |

#### Login Consent Provider

| Parameter | Description | Default |
|-----------|-------------|---------|
| `loginConsentProvider.enabled` | Enable login-consent-provider | `true` |
| `loginConsentProvider.replicaCount` | Number of replicas | `1` |
| `loginConsentProvider.image.repository` | Image repository | `your-registry/login-consent-provider` |
| `loginConsentProvider.image.tag` | Image tag | `latest` |
| `loginConsentProvider.service.port` | Service port | `3002` |
| `loginConsentProvider.config.databaseUrl` | Database connection URL | See values.yaml |

#### Secrets

| Parameter | Description | Default |
|-----------|-------------|---------|
| `secrets.hydra.system` | Hydra system secret | `youReallyNeedToChangeThis` |
| `secrets.hydra.cookie` | Hydra cookie secret | `youReallyNeedToChangeThisToo` |
| `secrets.hydra.oidcPairwiseSalt` | OIDC pairwise salt | `youReallyNeedToChangeThisThree` |
| `secrets.sessionSecret` | Session secret for login provider | `changeThisToASecureRandomString` |

**⚠️ IMPORTANT**: Always change the default secrets in production!

## Usage

### Accessing the Services

After installation, you can access the services:

1. **Hydra Public API** (OAuth2/OIDC endpoints):
   ```bash
   kubectl port-forward -n auth svc/auth-server-hydra-public 4444:4444
   ```
   Access at: http://localhost:4444

2. **Hydra Admin API** (Client management):
   ```bash
   kubectl port-forward -n auth svc/auth-server-hydra-admin 4445:4445
   ```
   Access at: http://localhost:4445

3. **Login & Consent Provider**:
   ```bash
   kubectl port-forward -n auth svc/auth-server-login-consent-provider 3002:3002
   ```
   Access at: http://localhost:3002

### Creating an OAuth2 Client

Once the services are running, you can create an OAuth2 client:

```bash
# Port-forward the admin API
kubectl port-forward -n auth svc/auth-server-hydra-admin 4445:4445

# Create a client
curl -X POST http://localhost:4445/admin/clients \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "my-client",
    "client_secret": "my-secret",
    "grant_types": ["authorization_code", "refresh_token"],
    "redirect_uris": ["http://localhost:5173/callback"],
    "response_types": ["code"],
    "scope": "openid offline_access"
  }'
```

### Testing the OAuth2 Flow

1. Navigate to the authorization endpoint:
   ```
   http://localhost:4444/oauth2/auth?client_id=my-client&response_type=code&scope=openid&redirect_uri=http://localhost:5173/callback
   ```

2. You'll be redirected to the login page at `http://localhost:3002/login`

3. After successful authentication and consent, you'll receive an authorization code

4. Exchange the code for tokens:
   ```bash
   curl -X POST http://localhost:4444/oauth2/token \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "grant_type=authorization_code" \
     -d "code=YOUR_AUTH_CODE" \
     -d "redirect_uri=http://localhost:5173/callback" \
     -d "client_id=my-client" \
     -d "client_secret=my-secret"
   ```

## Upgrading

To upgrade the chart:

```bash
helm upgrade auth-server . --namespace auth -f custom-values.yaml
```

## Uninstalling

To uninstall/delete the deployment:

```bash
helm uninstall auth-server --namespace auth
```

This will remove all resources associated with the chart, except for:
- PersistentVolumeClaims (if persistence is enabled)

To delete PVCs as well:

```bash
kubectl delete pvc -n auth -l app.kubernetes.io/instance=auth-server
```

## Architecture

### Component Interaction

```
┌─────────────────────────────────────────────────────────────┐
│                         External Users                       │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
         ┌────────────────────────┐
         │   Ingress (Optional)   │
         └────────┬───────────────┘
                  │
        ┌─────────┴──────────┐
        │                    │
        ▼                    ▼
┌──────────────┐    ┌──────────────────────┐
│ Hydra Public │    │ Login Consent        │
│   (4444)     │◄───┤ Provider (3002)      │
└──────┬───────┘    └──────────────────────┘
       │                     │
       │                     │
       ▼                     ▼
┌──────────────┐    ┌──────────────────────┐
│ Hydra Admin  │    │   External Postgres  │
│   (4445)     │    │   (portfolio DB)     │
└──────┬───────┘    └──────────────────────┘
       │
       ▼
┌──────────────┐
│  PostgreSQL  │
│   (5432)     │
└──────────────┘
```

### Startup Sequence

1. **PostgreSQL** starts first
2. **Hydra Migration Job** runs (waits for PostgreSQL)
3. **Hydra Admin** starts (waits for PostgreSQL and migration)
4. **Hydra Public** starts (waits for PostgreSQL and migration)
5. **Login Consent Provider** starts (waits for Hydra Admin)

This ordering is enforced using:
- Helm hooks for the migration job
- Init containers in deployments
- Readiness probes

## Troubleshooting

### Common Issues

1. **Migration job fails**
   - Check PostgreSQL is running: `kubectl get pods -n auth`
   - View migration logs: `kubectl logs -n auth job/auth-server-hydra-migrate`

2. **Hydra services can't connect to database**
   - Verify the DSN is correct in the secret
   - Check PostgreSQL service is accessible
   - View logs: `kubectl logs -n auth deployment/auth-server-hydra-admin`

3. **Login consent provider can't reach Hydra Admin**
   - Verify Hydra Admin service is running
   - Check the `HYDRA_ADMIN_URL` environment variable
   - View logs: `kubectl logs -n auth deployment/auth-server-login-consent-provider`

### Viewing Logs

```bash
# PostgreSQL logs
kubectl logs -n auth deployment/auth-server-postgres

# Hydra Admin logs
kubectl logs -n auth deployment/auth-server-hydra-admin

# Hydra Public logs
kubectl logs -n auth deployment/auth-server-hydra-public

# Login Consent Provider logs
kubectl logs -n auth deployment/auth-server-login-consent-provider

# Migration job logs
kubectl logs -n auth job/auth-server-hydra-migrate
```

### Debugging

To get a shell in a running pod:

```bash
kubectl exec -it -n auth deployment/auth-server-postgres -- /bin/sh
```

## Security Considerations

1. **Change all default secrets** before deploying to production
2. **Use TLS/SSL** for all external communications (enable ingress with TLS)
3. **Restrict network access** using NetworkPolicies
4. **Use strong passwords** for PostgreSQL
5. **Enable RBAC** and use service accounts with minimal permissions
6. **Regularly update** Hydra and PostgreSQL images
7. **Store secrets externally** using tools like HashiCorp Vault or AWS Secrets Manager

## Production Recommendations

1. **External Database**: Use a managed PostgreSQL service (AWS RDS, Google Cloud SQL, etc.)
   ```yaml
   postgres:
     enabled: false
   
   # Update DSN in secrets to point to external database
   ```

2. **High Availability**: Increase replica counts
   ```yaml
   hydraAdmin:
     replicaCount: 3
   hydraPublic:
     replicaCount: 3
   loginConsentProvider:
     replicaCount: 2
   ```

3. **Resource Limits**: Set appropriate resource requests and limits
   ```yaml
   hydraPublic:
     resources:
       requests:
         memory: "512Mi"
         cpu: "250m"
       limits:
         memory: "1Gi"
         cpu: "500m"
   ```

4. **Monitoring**: Integrate with Prometheus/Grafana
5. **Backup**: Implement regular database backups
6. **Secrets Management**: Use external secrets management

## License

This chart is provided as-is for the Portfolio Insights project.

## Support

For issues and questions, please refer to the main project repository.
