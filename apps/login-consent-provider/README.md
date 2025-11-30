# Login and Consent Provider for Ory Hydra

This is a standalone Go application that handles the login and consent flows for Ory Hydra OAuth2/OIDC implementation.

## Overview

This application serves as the user-facing identity layer for the Portfolio Insights authentication system. It handles:

- **Login Flow**: Authenticates users against the PostgreSQL database
- **Consent Flow**: Requests user permission for OAuth2 scopes
- **Logout Flow**: Terminates user sessions
- **Session Management**: Maintains user sessions across requests

## Architecture

```
User Browser
     ↓
Ory Hydra (Public)
     ↓
Login/Consent Provider (this app)
     ↓
Ory Hydra (Admin API)
     ↓
PostgreSQL (user database)
```

## Features

- ✅ Modern, responsive UI with gradient styling
- ✅ Secure password hashing with bcrypt
- ✅ Session management with secure cookies
- ✅ Integration with Hydra Admin API
- ✅ Remember me functionality
- ✅ CSRF protection
- ✅ Error handling and user feedback

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check endpoint |
| GET | `/login` | Display login form |
| POST | `/login` | Process login credentials |
| GET | `/consent` | Display consent screen |
| POST | `/consent` | Process consent decision |
| GET | `/logout` | Display logout confirmation |
| POST | `/logout` | Process logout |
| GET | `/error` | Display error page |

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `3001` |
| `HYDRA_ADMIN_URL` | Hydra Admin API URL | `http://localhost:4445` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://postgres:postgres@localhost:5432/investments?sslmode=disable` |
| `SESSION_SECRET` | Secret for session encryption | `change-this-secret` |
| `LOG_LEVEL` | Logging level (debug/info) | `info` |

## Development

### Prerequisites

- Go 1.21+
- PostgreSQL with users table
- Ory Hydra running

### Run Locally

```bash
# Install dependencies
go mod download

# Set environment variables
export PORT=3001
export HYDRA_ADMIN_URL=http://localhost:4445
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/investments?sslmode=disable
export SESSION_SECRET=my-secret-key
export LOG_LEVEL=debug

# Run the application
go run .
```

### Build

```bash
# Build binary
go build -o login-consent-provider .

# Run binary
./login-consent-provider
```

### Docker

```bash
# Build image
docker build -t login-consent-provider .

# Run container
docker run -p 3001:3001 \
  -e HYDRA_ADMIN_URL=http://hydra-admin:4445 \
  -e DATABASE_URL=postgres://postgres:postgres@postgres:5432/investments?sslmode=disable \
  -e SESSION_SECRET=my-secret-key \
  login-consent-provider
```

## Database Schema

The application expects a `users` table with the following structure:

```sql
CREATE TABLE customers.users (
    id VARCHAR(255) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Security

### Password Hashing

Passwords are hashed using bcrypt with a cost factor of 10.

### Session Management

Sessions are stored in encrypted cookies with the following settings:
- HttpOnly: true
- Secure: true (in production)
- SameSite: Strict
- Max-Age: 3600 seconds (1 hour)

### CSRF Protection

CSRF protection is built into the Gin framework and session middleware.

## Testing

### Manual Testing

1. **Start the application:**
   ```bash
   go run .
   ```

2. **Trigger login flow:**
   ```bash
   # Get login challenge from Hydra
   curl "http://localhost:3001/login?login_challenge=<challenge>"
   ```

3. **Submit login form:**
   ```bash
   curl -X POST http://localhost:3001/login \
     -d "challenge=<challenge>&email=user@example.com&password=password123"
   ```

### Integration Testing

The application is designed to work with Ory Hydra. Test the complete flow:

1. Start Hydra
2. Start Login/Consent Provider
3. Initiate OAuth2 flow from client
4. Verify login and consent screens appear
5. Verify successful token exchange

## Troubleshooting

### "Unable to connect to database"

- Verify PostgreSQL is running
- Check DATABASE_URL is correct
- Ensure users table exists

### "Hydra returned status 404"

- Verify HYDRA_ADMIN_URL is correct
- Ensure Hydra Admin API is accessible
- Check Hydra is running

### "Invalid password"

- Verify password is hashed with bcrypt
- Check password_hash column in database
- Ensure bcrypt cost factor matches

### "Session not persisting"

- Check SESSION_SECRET is set
- Verify cookies are enabled in browser
- Check cookie settings (HttpOnly, Secure, SameSite)

## Production Deployment

### Checklist

- [ ] Set strong SESSION_SECRET
- [ ] Enable HTTPS
- [ ] Set Secure cookie flag
- [ ] Use production database
- [ ] Enable rate limiting
- [ ] Add monitoring and logging
- [ ] Configure CORS properly
- [ ] Use environment-specific Hydra URLs

### Docker Compose

See `deployments/docker-compose/docker-compose.hydra.yml` for production configuration.

## License

Part of Portfolio Insights project.

## Contributing

This is a personal project. Contributions are welcome via pull requests.

---

*Last Updated: 2025-11-30*
