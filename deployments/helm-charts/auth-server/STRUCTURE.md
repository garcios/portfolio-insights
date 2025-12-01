# Auth Server Helm Chart - File Structure

## Chart Structure

```
auth-server/
├── .helmignore                                    # Files to exclude from chart package
├── Chart.yaml                                     # Chart metadata
├── README.md                                      # Comprehensive documentation
├── QUICK_REFERENCE.md                            # Quick command reference
├── values.yaml                                    # Default configuration values
└── templates/                                     # Kubernetes manifest templates
    ├── NOTES.txt                                 # Post-installation notes
    ├── _helpers.tpl                              # Template helper functions
    ├── configmap.yaml                            # Non-sensitive configuration
    ├── secret.yaml                               # Sensitive configuration
    ├── serviceaccount.yaml                       # Service account for pods
    ├── pvc.yaml                                  # Persistent volume claim for PostgreSQL
    ├── job-migrate.yaml                          # Database migration job (pre-install hook)
    ├── deployment-postgres.yaml                  # PostgreSQL deployment
    ├── deployment-hydra-admin.yaml               # Hydra Admin API deployment
    ├── deployment-hydra-public.yaml              # Hydra Public API deployment
    ├── deployment-login-consent-provider.yaml    # Login & Consent Provider deployment
    ├── service-postgres.yaml                     # PostgreSQL service
    ├── service-hydra-admin.yaml                  # Hydra Admin service
    ├── service-hydra-public.yaml                 # Hydra Public service
    ├── service-login-consent-provider.yaml       # Login & Consent Provider service
    ├── ingress-hydra-public.yaml                 # Ingress for Hydra Public (optional)
    └── ingress-login-consent-provider.yaml       # Ingress for Login Provider (optional)
```

## Component Mapping from Docker Compose

### Source: docker-compose.hydra.yml

| Docker Compose Service | Helm Template(s) | Notes |
|------------------------|------------------|-------|
| `hydra-postgres` | `deployment-postgres.yaml`<br>`service-postgres.yaml`<br>`pvc.yaml` | PostgreSQL database for Hydra |
| `hydra-migrate` | `job-migrate.yaml` | Converted to Kubernetes Job with pre-install hook |
| `hydra-admin` | `deployment-hydra-admin.yaml`<br>`service-hydra-admin.yaml` | Hydra Admin API (internal) |
| `hydra-public` | `deployment-hydra-public.yaml`<br>`service-hydra-public.yaml`<br>`ingress-hydra-public.yaml` | Hydra Public API (external) |
| `login-consent-provider` | `deployment-login-consent-provider.yaml`<br>`service-login-consent-provider.yaml`<br>`ingress-login-consent-provider.yaml` | Custom login/consent UI |

## Key Features Implemented

### ✅ Complete Translation
- All docker-compose services converted to Kubernetes resources
- Environment variables mapped to ConfigMaps and Secrets
- Volumes mapped to PersistentVolumeClaims
- Ports mapped to Services
- External access via optional Ingress

### ✅ Best Practices
- **API Versions**: Using `apps/v1` for Deployments, `v1` for Services/ConfigMaps/Secrets
- **Health Checks**: Liveness and readiness probes configured
- **Resource Management**: Resource requests and limits defined
- **Labels**: Consistent labeling using helper templates
- **Secrets Management**: Sensitive data in Secrets, non-sensitive in ConfigMaps
- **Startup Ordering**: Init containers to handle dependencies (replaces docker-compose `depends_on`)
- **Helm Hooks**: Migration job runs before deployments

### ✅ Configurability
- All values parameterized in `values.yaml`
- Easy customization via custom values files
- Support for enabling/disabling components
- Ingress configuration for external access
- Flexible resource allocation

### ✅ Documentation
- **README.md**: Comprehensive guide with installation, configuration, usage, troubleshooting
- **QUICK_REFERENCE.md**: Quick command reference for common operations
- **NOTES.txt**: Post-installation instructions displayed to users
- **_helpers.tpl**: Well-documented template functions

## Translation Highlights

### Environment Variables
- **Sensitive** → `secret.yaml` with references in deployments
- **Non-sensitive** → `configmap.yaml` with references in deployments

### Volumes
- `hydra-postgres-data` → PersistentVolumeClaim (`pvc.yaml`)

### Networking
- `hydra-network` → Kubernetes Services (automatic DNS)
- `default` network → Cross-namespace service references (configurable)

### Dependencies (depends_on)
Converted to:
- **Init containers** that wait for dependent services
- **Helm hooks** for migration job (runs before deployments)
- **Readiness probes** to ensure services are healthy before routing traffic

### Ports
| Service | Docker Port | K8s Service Port | Ingress |
|---------|-------------|------------------|---------|
| hydra-postgres | 5433:5432 | 5432 (ClusterIP) | No |
| hydra-admin | 4445:4445 | 4445 (ClusterIP) | No |
| hydra-public | 4444:4444 | 4444 (ClusterIP) | Optional |
| login-consent-provider | 3002:3002 | 3002 (ClusterIP) | Optional |

## Usage Examples

### Install with defaults
```bash
helm install auth-server ./deployments/helm-charts/auth-server -n auth --create-namespace
```

### Install with custom values
```bash
helm install auth-server ./deployments/helm-charts/auth-server \
  -n auth --create-namespace \
  -f custom-values.yaml
```

### Validate before installing
```bash
helm lint ./deployments/helm-charts/auth-server
helm template auth-server ./deployments/helm-charts/auth-server --debug
```

## Customization Points

### Common Customizations
1. **External Database**: Disable postgres, update DSN
2. **Secrets**: Override all default secrets (REQUIRED for production)
3. **Ingress**: Enable and configure for external access
4. **Resources**: Adjust CPU/memory limits
5. **Replicas**: Scale up for high availability
6. **Image**: Update login-consent-provider image repository/tag

See `README.md` and `QUICK_REFERENCE.md` for detailed examples.

## Validation

Chart has been validated with:
- ✅ `helm lint` - No errors
- ✅ `helm template` - Generates valid manifests
- ✅ Proper YAML syntax
- ✅ All required fields present
- ✅ Helper templates working correctly

## Next Steps

1. **Update image repository** for login-consent-provider in `values.yaml`
2. **Generate secure secrets** for production use
3. **Configure ingress** if external access is needed
4. **Test installation** in a development cluster
5. **Customize values** for your environment
6. **Set up CI/CD** to automate deployments

## Support

For detailed documentation, see:
- `README.md` - Full documentation
- `QUICK_REFERENCE.md` - Command reference
- `values.yaml` - Configuration options with comments
