# Podman-Compose Secrets Fix

## Problem
The original `docker-compose.yml` used environment-based secrets:
```yaml
secrets:
  db_password:
    environment: DB_PASSWORD
```

This syntax is not supported by `podman-compose`, which resulted in the error:
```
ValueError: ERROR: unparsable secret: "db_password", service: "postgres"
```

## Solution
Switched to using environment variables directly, which avoids file mounting issues on macOS:

```yaml
services:
  postgres:
    environment:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
```

## How It Works

1. **`.env` file**: Contains the actual password values (gitignored)
2. **`Makefile`**: Uses `podman-compose --env-file ...` to load these values
3. **`docker-compose.yml`**: References variables like `${DB_PASSWORD}`
4. **Podman**: Injects these values as environment variables into containers

## Usage

The `make podman-up` command now automatically loads the `.env` file:
```bash
make podman-up
```

## Files Modified

- `docker-compose.yml`: Removed `secrets` section, updated services to use `${VAR}` syntax
- `Makefile`: Added `--env-file deployments/docker-compose/.env` to podman commands
- `SECRETS.md`: Updated documentation
- `SECRETS_MIGRATION.md`: Updated migration notes

## Security

- The `.env` file is gitignored
- Only `.env.example` is committed to the repository
- Passwords are passed as environment variables (standard container practice)

## Testing

Verify the configuration:
```bash
podman-compose --env-file deployments/docker-compose/.env -f deployments/docker-compose/docker-compose.yml config
```

Start the services:
```bash
make podman-up
```
