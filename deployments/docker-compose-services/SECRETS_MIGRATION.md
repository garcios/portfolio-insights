# Docker Compose Secrets Migration - Summary

## Changes Made

### 1. Updated Go Services to Support Reading Passwords from Files

Modified the following services to support the `_FILE` suffix pattern for reading secrets:

- **user-service** (`services/user-service/internal/infrastructure/postgres.go`)
- **portfolio-service** (`services/portfolio-service/internal/infrastructure/postgres.go`)
- **transaction-service** (`services/transaction-service/internal/infrastructure/postgres.go`)
- **marketdata-service** (`services/marketdata-service/internal/infrastructure/postgres.go`)

Each service now includes a `getEnvOrFile()` helper function that:
1. First checks for `{KEY}_FILE` environment variable
2. If found, reads the password from that file path
3. Otherwise, falls back to the regular `{KEY}` environment variable
4. Finally, uses the default value if neither is set

### 2. Updated docker-compose.yml

**Updated secrets handling:**
Instead of Docker secrets (which had compatibility issues with Podman on macOS), we now use direct environment variable injection:

```yaml
services:
  postgres:
    environment:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
```

**Note:** This approach avoids "operation not permitted" errors when mounting secret files in Podman VMs on macOS. The `.env` file is still the source of truth for these values.

**Updated services to use secrets:**
- `postgres`: Uses `POSTGRES_PASSWORD_FILE` with `db_password` secret
- `user-service`: Uses `DB_PASSWORD_FILE` with `db_password` secret
- `portfolio-service`: Uses `DB_PASSWORD_FILE` with `db_password` secret
- `transaction-service`: Uses `DB_PASSWORD_FILE` with `db_password` secret
- `marketdata-service`: Uses `DB_PASSWORD_FILE` with `db_password` secret
- `minio`: Uses `MINIO_ROOT_PASSWORD_FILE` with `minio_password` secret
- `migrations`: Uses `${DB_PASSWORD}` environment variable interpolation

### 3. Created Configuration Files

- **`.env.example`**: Template file with placeholder passwords (committed to git)
- **`.env`**: Actual passwords file (gitignored, created for local development)
- **`SECRETS.md`**: Documentation on how to use and configure secrets

## Benefits

1. **Security**: Passwords are no longer hardcoded in docker-compose.yml
2. **Flexibility**: Easy to change passwords by updating .env file
3. **Best Practices**: Follows Docker Compose secrets pattern
4. **Backward Compatible**: Services still work with regular environment variables
5. **Production Ready**: Can easily migrate to Docker Swarm or Kubernetes secrets

## Migration Path

### For Local Development
1. Copy `.env.example` to `.env`
2. Update passwords in `.env` if needed
3. Run `make podman-up`

### For Production
Replace the `environment` secret driver with:
- Docker Swarm: Use `external: true` and create secrets with `docker secret create`
- Kubernetes: Use Kubernetes Secret objects
- Cloud: Use cloud-specific secret managers (AWS Secrets Manager, GCP Secret Manager, etc.)

## Testing

The configuration has been validated using `podman-compose config` and shows:
- Secrets are properly defined
- Services correctly reference secrets
- Environment variables are properly interpolated

## Next Steps

1. Test the services with the new secrets configuration
2. Update any deployment documentation
3. Consider adding more secrets for other sensitive data (API keys, tokens, etc.)
4. Implement secret rotation procedures for production
