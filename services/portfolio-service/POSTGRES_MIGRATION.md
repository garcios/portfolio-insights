# Portfolio Service - PostgreSQL Migration

## Overview

The portfolio-service has been refactored to use PostgreSQL for persistent storage instead of in-memory storage. It uses the existing `investments.holdings` table that was already defined in the database migrations.

## Database Schema

### Holdings Table

The service uses the existing `investments.holdings` table defined in `infra/db/000004_create_portfolio_tables.up.sql`:

```sql
CREATE TABLE investments.holdings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL DEFAULT 0,
    average_cost_basis DECIMAL(20, 8) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, symbol)
);
```

**Indexes:**
- `idx_holdings_user_id` - Fast lookups by user

**Constraints:**
- `UNIQUE(user_id, symbol)` - Ensures one holding per user per symbol

## Database Migrations

The database schema is managed centrally in the `infra/db` directory. The holdings table was already created by migration `000004_create_portfolio_tables.up.sql`.

**No additional migrations are needed** - the table already exists in your database.

## Environment Variables

The service uses the following environment variables for database connection:

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `garcios` | Database user |
| `DB_PASSWORD` | `Password123` | Database password |
| `DB_NAME` | `portfolio` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode (disable/require) |

## Repository Implementation

### PostgresHoldingRepository

Located in `internal/repository/postgres_holding_repo.go`

**Methods:**
- `Upsert(holding *domain.Holding) error` - Insert or update a holding
- `GetByUserAndSymbol(userID, symbol string) (*domain.Holding, error)` - Get specific holding
- `ListByUser(userID string) ([]*domain.Holding, error)` - Get all holdings for a user
- `Count() (int, error)` - Get total number of holdings
- `DeleteZeroQuantityHoldings() error` - Clean up holdings with zero quantity

### Key Features

1. **UPSERT Support**: Uses PostgreSQL's `ON CONFLICT` clause for efficient upserts
2. **Proper Error Handling**: Wraps errors with context
3. **SQL Injection Protection**: Uses parameterized queries
4. **Indexing**: Optimized for common query patterns

## Migration from In-Memory

The old in-memory implementation (`inmemory_holding_repo.go`) is still available but no longer used. 

**Changes made:**
1. Created `postgres.go` for database connection
2. Created `postgres_holding_repo.go` for PostgreSQL repository
3. Updated `main.go` to use PostgreSQL repository
4. Added `github.com/lib/pq` dependency

## Testing

### Verify Database Connection

```bash
# Check if service can connect
go run cmd/server/main.go
# Should see: "Successfully connected to PostgreSQL database"
```

### Check Holdings Table

```bash
psql -h localhost -U garcios -d portfolio -c "SELECT * FROM holdings;"
```

### Insert Test Data

```bash
psql -h localhost -U garcios -d portfolio -c "
INSERT INTO holdings (user_id, symbol, quantity, average_cost, last_updated)
VALUES ('user-123', 'AAPL', 10.5, 150.25, NOW())
ON CONFLICT (user_id, symbol) DO UPDATE SET
    quantity = EXCLUDED.quantity,
    average_cost = EXCLUDED.average_cost,
    last_updated = EXCLUDED.last_updated;
"
```

## Rollback Plan

If you need to rollback to in-memory storage:

1. Edit `cmd/server/main.go`:
   ```go
   // Comment out PostgreSQL setup
   // db, err := infrastructure.NewPostgresDB()
   // ...
   
   // Use in-memory repository
   repo := repository.NewInMemoryHoldingRepository()
   ```

2. Restart the service

## Performance Considerations

- **Connection Pooling**: The `sql.DB` object manages a connection pool automatically
- **Indexes**: All common query patterns are indexed
- **UPSERT**: Efficient single-query upsert using `ON CONFLICT`
- **Batch Operations**: Consider adding batch upsert for high-volume scenarios

## Next Steps

1. ✅ Run migrations to create holdings table
2. ✅ Start portfolio-service
3. ✅ Create a transaction to test NATS → Portfolio flow
4. ⏳ Add metrics for database operations
5. ⏳ Add database connection health checks
6. ⏳ Consider adding migration versioning (e.g., golang-migrate)

## Troubleshooting

### Connection Refused

```
Error: failed to connect to database: dial tcp [::1]:5432: connect: connection refused
```

**Solution**: Ensure PostgreSQL is running
```bash
# Check if PostgreSQL is running
pg_isready -h localhost -p 5432

# Start PostgreSQL (macOS with Homebrew)
brew services start postgresql
```

### Authentication Failed

```
Error: pq: password authentication failed for user "garcios"
```

**Solution**: Check credentials or create user
```bash
psql postgres -c "CREATE USER garcios WITH PASSWORD 'Password123';"
psql postgres -c "GRANT ALL PRIVILEGES ON DATABASE portfolio TO garcios;"
```

### Database Does Not Exist

```
Error: pq: database "portfolio" does not exist
```

**Solution**: Create the database
```bash
psql postgres -c "CREATE DATABASE portfolio OWNER garcios;"
```
