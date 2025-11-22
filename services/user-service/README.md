# User Service - PostgreSQL Integration

This document describes the PostgreSQL database integration for the User Service.

## Overview

The User Service now connects to a PostgreSQL database using the `customers` schema to store and manage user data.

## Database Schema

### Table: `customers.users`

```sql
CREATE TABLE customers.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## Configuration

The service uses environment variables for database configuration:

### Required Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` | `postgres` |
| `DB_PORT` | PostgreSQL port | `5432` | `5432` |
| `DB_USER` | Database user | `garcios` | `garcios` |
| `DB_PASSWORD` | Database password | `Password123` | `Password123` |
| `DB_NAME` | Database name | `portfolio` | `portfolio` |
| `DB_SSLMODE` | SSL mode | `disable` | `disable` or `require` |

### Optional Environment Variables (Connection Pool)

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_MAX_OPEN_CONNS` | Maximum open connections | `25` |
| `DB_MAX_IDLE_CONNS` | Maximum idle connections | `5` |
| `DB_CONN_MAX_LIFETIME` | Connection max lifetime | `5m` |

## Repository Methods

The `UserRepository` interface provides the following methods:

### GetByID
```go
GetByID(id string) (*User, error)
```
Retrieves a user by their UUID.

**Example:**
```go
user, err := repo.GetByID("550e8400-e29b-41d4-a716-446655440000")
```

### GetByEmail
```go
GetByEmail(email string) (*User, error)
```
Retrieves a user by their email address.

**Example:**
```go
user, err := repo.GetByEmail("john@example.com")
```

### Create
```go
Create(user *User) error
```
Creates a new user. The ID, CreatedAt, and UpdatedAt fields are automatically set by the database.

**Example:**
```go
user := &domain.User{
    Email:    "john@example.com",
    Name:     "John Doe",
    Password: "hashed_password_here",
}
err := repo.Create(user)
// user.ID, user.CreatedAt, user.UpdatedAt are now populated
```

### Update
```go
Update(user *User) error
```
Updates an existing user. The UpdatedAt field is automatically updated.

**Example:**
```go
user.Name = "Jane Doe"
err := repo.Update(user)
```

### Delete
```go
Delete(id string) error
```
Deletes a user by their UUID.

**Example:**
```go
err := repo.Delete("550e8400-e29b-41d4-a716-446655440000")
```

## Connection Management

### Connection Pool

The service configures a connection pool with the following defaults:
- **Max Open Connections**: 25
- **Max Idle Connections**: 5
- **Connection Max Lifetime**: 5 minutes

These can be customized via environment variables.

### Connection Verification

On startup, the service:
1. Attempts to connect to PostgreSQL
2. Pings the database to verify connectivity
3. Logs connection success or exits with an error

### Graceful Shutdown

The database connection is properly closed when the service shuts down using `defer db.Close()`.

## Error Handling

All repository methods return descriptive errors:

- **User not found**: `fmt.Errorf("user not found: %s", id)`
- **Database errors**: Wrapped with context using `fmt.Errorf("failed to ...: %w", err)`
- **No rows affected**: Indicates the resource doesn't exist

## Running the Service

### Local Development

```bash
# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=garcios
export DB_PASSWORD=Password123
export DB_NAME=portfolio
export DB_SSLMODE=disable

# Run the service
go run cmd/server/main.go
```

### Docker/Podman

The service is configured in `docker-compose.yml`:

```yaml
user-service:
  build:
    context: ../../
    dockerfile: services/user-service/Dockerfile
  environment:
    - DB_HOST=postgres
    - DB_USER=garcios
    - DB_PASSWORD=Password123
    - DB_NAME=portfolio
  depends_on:
    - postgres
```

Start with:
```bash
make podman-up
```

## Testing Database Connection

### Using psql

```bash
# Connect to the database
psql -h localhost -p 5432 -U garcios -d portfolio

# Check the users table
\dt customers.*
\d customers.users

# Query users
SELECT * FROM customers.users;
```

### Using the Service

The service will log connection status on startup:

```
Successfully connected to PostgreSQL database: garcios@postgres:5432/portfolio
```

## Migration

Database migrations are located in `/infra/db/`:

- `000001_create_users_table.up.sql` - Creates the `customers` schema and `users` table
- `000001_create_users_table.down.sql` - Drops the table and schema

Apply migrations:
```bash
cd infra/db
make migrate-up
```

## Dependencies

The service uses the following Go packages:

- `database/sql` - Standard library SQL interface
- `github.com/lib/pq` - PostgreSQL driver

Install dependencies:
```bash
go mod download
```

## Security Considerations

1. **Password Storage**: Always hash passwords before storing (use bcrypt or argon2)
2. **SQL Injection**: All queries use parameterized statements (`$1`, `$2`, etc.)
3. **SSL/TLS**: Use `DB_SSLMODE=require` in production
4. **Credentials**: Never commit credentials; use environment variables or secrets management
5. **Connection Limits**: Configure appropriate pool sizes for your workload

## Troubleshooting

### Connection Refused

```
failed to connect to database: connection refused
```

**Solutions:**
- Verify PostgreSQL is running: `podman ps | grep postgres`
- Check the host and port are correct
- Ensure the database accepts connections from your IP

### Role Does Not Exist

```
role "garcios" does not exist
```

**Solutions:**
- Verify the user exists in PostgreSQL
- Check the `DB_USER` environment variable
- Ensure you're connecting to the correct PostgreSQL instance

### Database Does Not Exist

```
database "portfolio" does not exist
```

**Solutions:**
- Create the database: `createdb -U garcios portfolio`
- Run migrations: `cd infra/db && make migrate-up`

### Schema Does Not Exist

```
schema "customers" does not exist
```

**Solutions:**
- Run migrations: `cd infra/db && make migrate-up`
- Verify migrations completed successfully

## Performance Tips

1. **Use Connection Pooling**: Already configured by default
2. **Index Frequently Queried Columns**: Email has a unique index
3. **Use Prepared Statements**: Consider using `db.Prepare()` for repeated queries
4. **Monitor Slow Queries**: Enable PostgreSQL query logging
5. **Use Transactions**: For operations that modify multiple tables

## Future Enhancements

- [ ] Add pagination for list operations
- [ ] Implement soft deletes with `deleted_at` column
- [ ] Add user search functionality
- [ ] Implement caching layer (Redis)
- [ ] Add database connection health checks
- [ ] Implement query timeouts
- [ ] Add database metrics and monitoring

## References

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Go database/sql Tutorial](https://go.dev/doc/database/sql-injection)
- [lib/pq Documentation](https://github.com/lib/pq)
- [Schema Organization](../../infra/db/SCHEMA_ORGANIZATION.md)

---

**Last Updated**: 2025-11-22  
**Service Version**: 1.0.0
