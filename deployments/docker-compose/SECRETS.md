# Docker Compose Secrets Configuration

This directory uses Docker Compose secrets to manage sensitive data like passwords.

## Setup

1. **Copy the example environment file:**
   ```bash
   cp .env.example .env
   ```

2. **Edit `.env` with your actual passwords:**
   ```bash
   # Database Configuration
   DB_PASSWORD=your_secure_password_here
   POSTGRES_PASSWORD=your_secure_password_here

   # MinIO Configuration
   MINIO_ROOT_PASSWORD=your_minio_password_here
   ```

3. **Start the services:**
   ```bash
   podman-compose up -d
   # or from the project root:
   make podman-up
   ```

## How It Works

The `docker-compose.yml` file uses environment variables to inject sensitive data into containers, which is fully compatible with Podman on macOS:

- **Environment variables are defined** in `docker-compose.yml`:
  ```yaml
  services:
    postgres:
      environment:
        POSTGRES_PASSWORD: ${DB_PASSWORD}
  ```

- **Values are loaded** from the `.env` file using the `--env-file` flag in `make podman-up`:
  ```makefile
  podman-up:
      podman-compose --env-file deployments/docker-compose/.env ...
  ```

- **Application code reads from environment variables**: The Go services are configured to read passwords directly from these environment variables.

## Security Benefits

1. **No hardcoded passwords** in docker-compose.yml
2. **`.env` file is gitignored** to prevent accidental commits
3. **Secrets are mounted as files** in `/run/secrets/` (in-memory filesystem)
4. **Easy to rotate** passwords by updating `.env` and restarting services

## Services Using Secrets

- **postgres**: Uses `db_password` secret via `POSTGRES_PASSWORD_FILE`
- **user-service**: Uses `db_password` secret via `DB_PASSWORD_FILE`
- **portfolio-service**: Uses `db_password` secret via `DB_PASSWORD_FILE`
- **transaction-service**: Uses `db_password` secret via `DB_PASSWORD_FILE`
- **marketdata-service**: Uses `db_password` secret via `DB_PASSWORD_FILE`
- **minio**: Uses `minio_password` secret via `MINIO_ROOT_PASSWORD_FILE`
- **migrations**: Uses `DB_PASSWORD` environment variable (interpolated from `.env`)

## Production Deployment

For production, consider using:
- **Docker Swarm secrets** for orchestrated deployments
- **Kubernetes secrets** for Kubernetes deployments
- **External secret managers** like HashiCorp Vault, AWS Secrets Manager, etc.

## Troubleshooting

If services fail to start:

1. **Check that `.env` file exists:**
   ```bash
   ls -la .env
   ```

2. **Verify environment variables are set:**
   ```bash
   podman-compose config
   ```

3. **Check service logs:**
   ```bash
   podman-compose logs <service-name>
   ```
