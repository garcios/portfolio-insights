# Auth Server Helm Chart - Quick Reference

## Installation

### Basic Installation
```bash
helm install auth-server ./deployments/helm-charts/auth-server \
  --namespace auth \
  --create-namespace
```

### Installation with Custom Values
```bash
helm install auth-server ./deployments/helm-charts/auth-server \
  --namespace auth \
  --create-namespace \
  --values custom-values.yaml
```

### Dry Run (Test without installing)
```bash
helm install auth-server ./deployments/helm-charts/auth-server \
  --namespace auth \
  --dry-run \
  --debug
```

## Upgrade

```bash
helm upgrade auth-server ./deployments/helm-charts/auth-server \
  --namespace auth \
  --values custom-values.yaml
```

## Uninstall

```bash
# Uninstall the release
helm uninstall auth-server --namespace auth

# Also delete PVCs
kubectl delete pvc -n auth -l app.kubernetes.io/instance=auth-server
```

## Validation

### Lint the Chart
```bash
helm lint ./deployments/helm-charts/auth-server
```

### Template Rendering (see what will be deployed)
```bash
helm template auth-server ./deployments/helm-charts/auth-server
```

### Validate Against Kubernetes
```bash
helm template auth-server ./deployments/helm-charts/auth-server | kubectl apply --dry-run=client -f -
```

## Common Customizations

### Example: Production Values

Create a `production-values.yaml`:

```yaml
# Use external PostgreSQL
postgres:
  enabled: false

# Update secrets (REQUIRED for production!)
secrets:
  hydra:
    system: "CHANGE-ME-32-CHARS-MINIMUM-XXXXX"
    cookie: "CHANGE-ME-32-CHARS-MINIMUM-XXXXX"
    oidcPairwiseSalt: "CHANGE-ME-32-CHARS-MINIMUM-XXXXX"
  sessionSecret: "CHANGE-ME-32-CHARS-MINIMUM-XXXXX"

# Update DSN to point to external database
# (This would typically be done via secret management)

# Enable ingress
hydraPublic:
  replicaCount: 3
  ingress:
    enabled: true
    className: "nginx"
    annotations:
      cert-manager.io/cluster-issuer: "letsencrypt-prod"
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
  replicaCount: 2
  image:
    repository: your-registry.com/login-consent-provider
    tag: "v1.0.0"
  ingress:
    enabled: true
    className: "nginx"
    annotations:
      cert-manager.io/cluster-issuer: "letsencrypt-prod"
    hosts:
      - host: login.yourdomain.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: login-tls
        hosts:
          - login.yourdomain.com

# Increase resources
hydraAdmin:
  resources:
    requests:
      memory: "512Mi"
      cpu: "250m"
    limits:
      memory: "1Gi"
      cpu: "500m"

hydraPublic:
  resources:
    requests:
      memory: "512Mi"
      cpu: "250m"
    limits:
      memory: "1Gi"
      cpu: "500m"
```

Then install:
```bash
helm install auth-server ./deployments/helm-charts/auth-server \
  --namespace auth \
  --create-namespace \
  --values production-values.yaml
```

## Monitoring

### Check Release Status
```bash
helm status auth-server -n auth
```

### List All Releases
```bash
helm list -n auth
```

### Get Release Values
```bash
helm get values auth-server -n auth
```

### Get All Release Information
```bash
helm get all auth-server -n auth
```

## Debugging

### View Rendered Templates
```bash
helm get manifest auth-server -n auth
```

### Check Pod Status
```bash
kubectl get pods -n auth -l app.kubernetes.io/instance=auth-server
```

### View Logs
```bash
# Postgres
kubectl logs -n auth -l app.kubernetes.io/component=postgres

# Hydra Admin
kubectl logs -n auth -l app.kubernetes.io/component=hydra-admin

# Hydra Public
kubectl logs -n auth -l app.kubernetes.io/component=hydra-public

# Login Consent Provider
kubectl logs -n auth -l app.kubernetes.io/component=login-consent-provider

# Migration Job
kubectl logs -n auth job/auth-server-hydra-migrate
```

### Port Forward for Local Access
```bash
# Hydra Public
kubectl port-forward -n auth svc/auth-server-hydra-public 4444:4444

# Hydra Admin
kubectl port-forward -n auth svc/auth-server-hydra-admin 4445:4445

# Login Consent Provider
kubectl port-forward -n auth svc/auth-server-login-consent-provider 3002:3002
```

## Package the Chart

### Create a Chart Package
```bash
helm package ./deployments/helm-charts/auth-server
```

This creates `auth-server-0.1.0.tgz`

### Install from Package
```bash
helm install auth-server auth-server-0.1.0.tgz --namespace auth --create-namespace
```

## Rollback

### View Release History
```bash
helm history auth-server -n auth
```

### Rollback to Previous Version
```bash
helm rollback auth-server -n auth
```

### Rollback to Specific Revision
```bash
helm rollback auth-server 1 -n auth
```

## Testing OAuth2 Flow

### 1. Create an OAuth2 Client
```bash
kubectl port-forward -n auth svc/auth-server-hydra-admin 4445:4445

curl -X POST http://localhost:4445/admin/clients \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "test-client",
    "client_secret": "test-secret",
    "grant_types": ["authorization_code", "refresh_token"],
    "redirect_uris": ["http://localhost:5173/callback"],
    "response_types": ["code"],
    "scope": "openid offline_access email profile"
  }'
```

### 2. Start Authorization Flow
Open in browser:
```
http://localhost:4444/oauth2/auth?client_id=test-client&response_type=code&scope=openid&redirect_uri=http://localhost:5173/callback&state=random-state
```

### 3. Exchange Code for Token
```bash
curl -X POST http://localhost:4444/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=YOUR_AUTH_CODE" \
  -d "redirect_uri=http://localhost:5173/callback" \
  -d "client_id=test-client" \
  -d "client_secret=test-secret"
```

## Security Best Practices

1. **Always change default secrets in production**
2. **Use external secrets management** (e.g., HashiCorp Vault, AWS Secrets Manager)
3. **Enable TLS/HTTPS** via ingress
4. **Use NetworkPolicies** to restrict traffic
5. **Regularly update** images to get security patches
6. **Use RBAC** with minimal permissions
7. **Enable audit logging**
8. **Use external managed database** for production

## Troubleshooting

### Migration Job Fails
```bash
# Check job logs
kubectl logs -n auth job/auth-server-hydra-migrate

# Check if postgres is ready
kubectl get pods -n auth -l app.kubernetes.io/component=postgres

# Manually delete and retry
kubectl delete job -n auth auth-server-hydra-migrate
helm upgrade auth-server ./deployments/helm-charts/auth-server -n auth
```

### Pods Not Starting
```bash
# Describe pod to see events
kubectl describe pod -n auth <pod-name>

# Check init containers
kubectl logs -n auth <pod-name> -c wait-for-postgres
```

### Database Connection Issues
```bash
# Verify secret
kubectl get secret -n auth auth-server-secrets -o yaml

# Test database connectivity from a pod
kubectl run -it --rm debug --image=postgres:16-alpine --restart=Never -n auth -- \
  psql "postgres://hydra:hydra_secret@auth-server-postgres:5432/hydra?sslmode=disable"
```
