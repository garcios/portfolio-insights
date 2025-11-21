# Portfolio Insights - Podman Setup Guide

## Prerequisites

Install Podman and podman-compose on macOS:

```bash
# Install Podman
brew install podman

# Install podman-compose
brew install podman-compose

# Initialize and start the Podman machine
podman machine init
podman machine start

# Verify installation
podman --version
podman-compose --version
```

## Running the Application

### Start all services
```bash
make podman-up
```

### Stop all services
```bash
make podman-down
```

### View logs
```bash
make podman-logs
```

## Podman vs Docker Differences

1. **Rootless by default**: Podman runs containers without root privileges
2. **No daemon**: Podman doesn't require a background daemon
3. **Pod support**: Native Kubernetes pod support
4. **Docker compatibility**: Most Docker commands work with Podman using aliases

## Troubleshooting

### Port binding issues
If you encounter port binding issues, ensure the Podman machine has sufficient resources:
```bash
podman machine stop
podman machine rm
podman machine init --cpus 4 --memory 8192 --disk-size 50
podman machine start
```

### Volume permissions
Podman handles volumes differently. If you have permission issues:
```bash
# Check volume
podman volume ls

# Inspect volume
podman volume inspect postgres_data

# Remove and recreate if needed
podman volume rm postgres_data
```

### Using Docker alias (optional)
If you want to use `docker` commands with Podman:
```bash
# Add to your ~/.zshrc
alias docker=podman
alias docker-compose=podman-compose
```

## Service URLs

- **Gateway**: http://localhost:8080
- **PostgreSQL**: localhost:5432
- **NATS**: localhost:4222 (client), localhost:8222 (monitoring)

## Database Credentials

- **User**: garcios
- **Password**: Password123
- **Database**: portfolio
