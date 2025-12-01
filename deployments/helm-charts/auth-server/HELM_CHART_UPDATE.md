# Helm Chart Update: Service-to-Service Authentication

## Overview

The `auth-server` Helm chart has been updated to reflect the migration from direct database access to service-to-service gRPC authentication for the `login-consent-provider`.

## Changes Made

### 1. Deployment Configuration (`templates/deployment-login-consent-provider.yaml`)

#### Environment Variables
**Before:**
```yaml
- name: DATABASE_URL
  valueFrom:
    configMapKeyRef:
      name: {{ include "auth-server.fullname" . }}-config
      key: login-consent-database-url
```

**After:**
```yaml
- name: USER_SERVICE_ADDR
  valueFrom:
    configMapKeyRef:
      name: {{ include "auth-server.fullname" . }}-config
      key: login-consent-user-service-addr
```

#### Init Containers
Added a new init container to wait for user-service availability:

```yaml
initContainers:
- name: wait-for-hydra-admin
  image: busybox:1.36
  command: ['sh', '-c', 'until nc -z {{ include "auth-server.hydraAdmin.fullname" . }} {{ .Values.hydraAdmin.service.port }}; do echo waiting for hydra-admin; sleep 2; done;']
- name: wait-for-user-service
  image: busybox:1.36
  command: ['sh', '-c', 'until nc -z {{ .Values.loginConsentProvider.config.userServiceAddr | replace ":50051" "" }} 50051; do echo waiting for user-service; sleep 2; done;']
```

### 2. ConfigMap (`templates/configmap.yaml`)

**Before:**
```yaml
login-consent-database-url: {{ .Values.loginConsentProvider.config.databaseUrl | quote }}
```

**After:**
```yaml
login-consent-user-service-addr: {{ .Values.loginConsentProvider.config.userServiceAddr | quote }}
```

### 3. Values (`values.yaml`)

**Before:**
```yaml
loginConsentProvider:
  config:
    port: "3002"
    hydraAdminUrl: "http://hydra-admin:4445"
    databaseUrl: "postgres://garcios:Password123@postgres:5432/portfolio?sslmode=disable"
    logLevel: debug
```

**After:**
```yaml
loginConsentProvider:
  config:
    port: "3002"
    hydraAdminUrl: "http://hydra-admin:4445"
    # User Service Address for service-to-service authentication
    # Replaces direct database access for improved security
    userServiceAddr: "user-service:50051"
    logLevel: debug
```

## Deployment Architecture

### Previous Architecture
```
┌─────────────────────────┐
│ Login-Consent-Provider  │
│   Pod                   │
└────────────┬────────────┘
             │
             │ DATABASE_URL
             ▼
   ┌─────────────────┐
   │   PostgreSQL    │
   │   (External)    │
   └─────────────────┘
```

### New Architecture
```
┌─────────────────────────┐
│ Login-Consent-Provider  │
│   Pod                   │
└────────────┬────────────┘
             │
             │ gRPC (USER_SERVICE_ADDR)
             ▼
   ┌─────────────────────┐
   │   User-Service      │
   │   Service           │
   └──────────┬──────────┘
              │
              │ DB Connection
              ▼
    ┌─────────────────┐
    │   PostgreSQL    │
    └─────────────────┘
```

## Installation

### Prerequisites

1. **User-Service must be deployed** in the same namespace or accessible via service name
2. **User-Service must expose port 50051** for gRPC connections
3. **Network policies** should allow traffic from login-consent-provider to user-service

### Install/Upgrade Command

```bash
# Install
helm install auth-server ./deployments/helm-charts/auth-server \
  --namespace auth \
  --create-namespace

# Upgrade existing installation
helm upgrade auth-server ./deployments/helm-charts/auth-server \
  --namespace auth
```

### Custom Values

Create a `custom-values.yaml` file:

```yaml
loginConsentProvider:
  config:
    # Update if user-service is in a different namespace
    userServiceAddr: "user-service.default.svc.cluster.local:50051"
    
  # Enable ingress for external access
  ingress:
    enabled: true
    className: "nginx"
    hosts:
      - host: login.example.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: login-tls
        hosts:
          - login.example.com

# Production secrets
secrets:
  hydra:
    system: "CHANGE_THIS_IN_PRODUCTION"
    cookie: "CHANGE_THIS_IN_PRODUCTION"
    oidcPairwiseSalt: "CHANGE_THIS_IN_PRODUCTION"
  sessionSecret: "CHANGE_THIS_IN_PRODUCTION"
```

Install with custom values:
```bash
helm install auth-server ./deployments/helm-charts/auth-server \
  --namespace auth \
  --create-namespace \
  --values custom-values.yaml
```

## Configuration Options

### User Service Address Formats

| Environment | Format | Example |
|-------------|--------|---------|
| Same namespace | `service-name:port` | `user-service:50051` |
| Different namespace | `service-name.namespace.svc.cluster.local:port` | `user-service.default.svc.cluster.local:50051` |
| External service | `hostname:port` | `user-service.example.com:50051` |

### Security Considerations

#### 1. Network Policies
Create a NetworkPolicy to restrict access:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: login-consent-provider-egress
  namespace: auth
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: auth-server
      app.kubernetes.io/component: login-consent-provider
  policyTypes:
  - Egress
  egress:
  # Allow DNS
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: UDP
      port: 53
  # Allow to user-service
  - to:
    - namespaceSelector:
        matchLabels:
          name: default
      podSelector:
        matchLabels:
          app: user-service
    ports:
    - protocol: TCP
      port: 50051
  # Allow to hydra-admin
  - to:
    - podSelector:
        matchLabels:
          app.kubernetes.io/name: auth-server
          app.kubernetes.io/component: hydra-admin
    ports:
    - protocol: TCP
      port: 4445
```

#### 2. Service Mesh (Istio/Linkerd)
For production, consider using a service mesh for automatic mTLS:

```yaml
# Istio PeerAuthentication
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: auth
spec:
  mtls:
    mode: STRICT
```

#### 3. Pod Security Standards
Apply pod security standards:

```yaml
podSecurityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 1000
  seccompProfile:
    type: RuntimeDefault

securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true
```

## Verification

### 1. Check Pod Status
```bash
kubectl get pods -n auth -l app.kubernetes.io/component=login-consent-provider
```

### 2. Check Init Containers
```bash
kubectl describe pod -n auth -l app.kubernetes.io/component=login-consent-provider
```

Look for:
- `wait-for-hydra-admin` - should complete successfully
- `wait-for-user-service` - should complete successfully

### 3. Check Logs
```bash
# Login-consent-provider logs
kubectl logs -n auth -l app.kubernetes.io/component=login-consent-provider

# Should see:
# "Connected to user-service at user-service:50051"
```

### 4. Test gRPC Connection
```bash
# Port-forward to login-consent-provider
kubectl port-forward -n auth svc/login-consent-provider 3002:3002

# Test login flow
curl http://localhost:3002/health
```

### 5. Test OAuth Flow
```bash
# Get Hydra public URL
kubectl port-forward -n auth svc/hydra-public 4444:4444

# Create OAuth client
kubectl run -it --rm oauth-client --image=curlimages/curl --restart=Never -- \
  curl -X POST http://hydra-admin:4445/admin/clients \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "test-client",
    "client_secret": "test-secret",
    "grant_types": ["authorization_code", "refresh_token"],
    "redirect_uris": ["http://localhost:5173/callback"],
    "response_types": ["code"],
    "scope": "openid profile email"
  }'

# Test authorization flow
# Navigate to: http://localhost:4444/oauth2/auth?client_id=test-client&response_type=code&scope=openid&redirect_uri=http://localhost:5173/callback
```

## Troubleshooting

### Issue: Init container stuck on "waiting for user-service"

**Cause:** User-service is not deployed or not accessible

**Solution:**
```bash
# Check if user-service exists
kubectl get svc user-service -n default

# Check if user-service is ready
kubectl get pods -l app=user-service -n default

# Check network connectivity
kubectl run -it --rm debug --image=busybox --restart=Never -- \
  nc -zv user-service.default.svc.cluster.local 50051
```

### Issue: gRPC connection errors in logs

**Cause:** Incorrect user-service address or port

**Solution:**
```bash
# Verify service address
kubectl get svc user-service -n default -o yaml

# Update values.yaml with correct address
helm upgrade auth-server ./deployments/helm-charts/auth-server \
  --namespace auth \
  --set loginConsentProvider.config.userServiceAddr="user-service.default.svc.cluster.local:50051"
```

### Issue: Authentication fails

**Cause:** User-service not responding or database issues

**Solution:**
```bash
# Check user-service logs
kubectl logs -n default -l app=user-service

# Check user-service health
kubectl exec -it -n default deployment/user-service -- \
  grpcurl -plaintext localhost:50051 list

# Test VerifyUser RPC
kubectl exec -it -n default deployment/user-service -- \
  grpcurl -plaintext -d '{"email":"test@example.com","password":"password"}' \
  localhost:50051 user.UserService/VerifyUser
```

## Rollback

If you need to rollback to the previous version:

```bash
# Rollback to previous release
helm rollback auth-server -n auth

# Or rollback to specific revision
helm rollback auth-server 1 -n auth
```

## Migration from Direct DB Access

If you're upgrading from a version that used direct database access:

### 1. Backup Current Configuration
```bash
helm get values auth-server -n auth > backup-values.yaml
```

### 2. Update Values
```bash
# Remove databaseUrl
# Add userServiceAddr
```

### 3. Upgrade
```bash
helm upgrade auth-server ./deployments/helm-charts/auth-server \
  --namespace auth \
  --values custom-values.yaml
```

### 4. Verify
```bash
# Check that DATABASE_URL is no longer in environment
kubectl exec -it -n auth deployment/login-consent-provider -- env | grep DATABASE_URL
# Should return nothing

# Check that USER_SERVICE_ADDR is set
kubectl exec -it -n auth deployment/login-consent-provider -- env | grep USER_SERVICE_ADDR
# Should show: USER_SERVICE_ADDR=user-service:50051
```

## Production Recommendations

### 1. Use Secrets for Service Addresses
Instead of hardcoding in values.yaml:

```yaml
# Create secret
kubectl create secret generic user-service-config \
  --from-literal=address=user-service.default.svc.cluster.local:50051 \
  -n auth

# Update deployment to use secret
env:
- name: USER_SERVICE_ADDR
  valueFrom:
    secretKeyRef:
      name: user-service-config
      key: address
```

### 2. Enable mTLS
Use cert-manager to generate certificates:

```bash
# Install cert-manager
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set installCRDs=true

# Create certificate for login-consent-provider
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: login-consent-provider-cert
  namespace: auth
spec:
  secretName: login-consent-provider-tls
  issuerRef:
    name: ca-issuer
    kind: ClusterIssuer
  dnsNames:
  - login-consent-provider.auth.svc.cluster.local
EOF
```

### 3. Add Resource Limits
```yaml
loginConsentProvider:
  resources:
    requests:
      memory: "128Mi"
      cpu: "100m"
    limits:
      memory: "256Mi"
      cpu: "500m"
```

### 4. Enable Horizontal Pod Autoscaling
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: login-consent-provider
  namespace: auth
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: login-consent-provider
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Summary

The Helm chart has been successfully updated to use service-to-service gRPC authentication instead of direct database access. This provides:

- ✅ **Better Security**: No database credentials in login-consent-provider
- ✅ **Separation of Concerns**: User-service owns all user operations
- ✅ **Scalability**: Services can scale independently
- ✅ **Maintainability**: Easier to update and maintain
- ✅ **Production Ready**: Supports mTLS, network policies, and service mesh

For questions or issues, refer to the main [SECURITY_ARCHITECTURE.md](../../apps/login-consent-provider/SECURITY_ARCHITECTURE.md) documentation.
