# Portfolio Service - PostgreSQL Refactoring Summary

## ✅ What Was Done

Successfully refactored the portfolio-service to use PostgreSQL persistent storage instead of in-memory storage.

## 📁 Files Created/Modified

### Created Files:
1. **`internal/infrastructure/postgres.go`**
   - Database connection setup
   - Uses environment variables for configuration
   - Defaults: localhost:5432, user: garcios, db: portfolio

2. **`internal/repository/postgres_holding_repo.go`**
   - PostgreSQL implementation of `HoldingRepository` interface
   - Uses existing `investments.holdings` table
   - Methods: Upsert, GetByUserAndSymbol, ListByUser, Count, DeleteZeroQuantityHoldings

3. **`POSTGRES_MIGRATION.md`**
   - Documentation for the PostgreSQL migration
   - Schema reference
   - Configuration guide

### Modified Files:
1. **`cmd/server/main.go`**
   - Changed from `NewInMemoryHoldingRepository()` to `NewPostgresHoldingRepository(db)`
   - Added database connection setup
   - Added proper cleanup with `defer db.Close()`

2. **`go.mod`**
   - Added `github.com/lib/pq v1.10.9` dependency

## 🗄️ Database Schema

Uses the existing `investments.holdings` table:

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

**Key Features:**
- **UPSERT support**: Uses `ON CONFLICT (user_id, symbol)` for efficient updates
- **UUID primary key**: Auto-generated unique identifier
- **Unique constraint**: One holding per user per symbol
- **Indexed**: `idx_holdings_user_id` for fast user lookups

## 🔄 Repository Implementation

### PostgresHoldingRepository Methods

```go
// Insert or update a holding
Upsert(holding *domain.Holding) error

// Get a specific holding
GetByUserAndSymbol(userID, symbol string) (*domain.Holding, error)

// Get all holdings for a user
ListByUser(userID string) ([]*domain.Holding, error)

// Get total count (for metrics)
Count() (int, error)

// Clean up zero quantity holdings
DeleteZeroQuantityHoldings() error
```

### Domain Mapping

| Domain Field | Database Column |
|--------------|-----------------|
| `UserID` | `user_id` |
| `Symbol` | `symbol` |
| `Quantity` | `quantity` |
| `AverageCost` | `average_cost_basis` |
| `LastUpdated` | `updated_at` |

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `garcios` | Database user |
| `DB_PASSWORD` | `Password123` | Database password |
| `DB_NAME` | `portfolio` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |

### Example Configuration

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=garcios
export DB_PASSWORD=Password123
export DB_NAME=portfolio
```

## 🚀 How to Use

### 1. Ensure Database is Running

```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Check if investments.holdings table exists
psql -h localhost -U garcios -d portfolio -c "\d investments.holdings"
```

### 2. Start the Service

```bash
cd services/portfolio-service
go run cmd/server/main.go
```

You should see:
```
Portfolio Service starting...
Successfully connected to PostgreSQL database
Portfolio Service is running and listening for events...
```

### 3. Test the Integration

Create a transaction in the transaction-service, and it will:
1. Publish a NATS event
2. Portfolio-service receives the event
3. Updates the `investments.holdings` table
4. Data persists in PostgreSQL

### 4. Verify Data

```bash
# Check holdings in database
psql -h localhost -U garcios -d portfolio -c "SELECT * FROM investments.holdings;"

# Count holdings
psql -h localhost -U garcios -d portfolio -c "SELECT COUNT(*) FROM investments.holdings;"
```

## 🔍 Key Differences from In-Memory

| Aspect | In-Memory | PostgreSQL |
|--------|-----------|------------|
| **Persistence** | Lost on restart | Persists across restarts |
| **Concurrency** | Mutex locks | Database transactions |
| **Scalability** | Single instance | Multiple instances possible |
| **Querying** | Limited | Full SQL capabilities |
| **Data Loss Risk** | High | Low (with backups) |

## 📊 Benefits

1. **Data Persistence**: Holdings survive service restarts
2. **Reliability**: ACID transactions ensure data integrity
3. **Scalability**: Can run multiple service instances
4. **Auditability**: Can query historical data
5. **Backup/Recovery**: Standard PostgreSQL backup tools work
6. **Production Ready**: Suitable for production deployment

## 🧪 Testing

### Manual Test

```bash
# 1. Create a transaction (via transaction-service)
# 2. Check the database
psql -h localhost -U garcios -d portfolio -c "
SELECT user_id, symbol, quantity, average_cost_basis, updated_at 
FROM investments.holdings 
ORDER BY updated_at DESC 
LIMIT 5;
"
```

### Expected Flow

```
Transaction Created (transaction-service)
    ↓
NATS Event Published
    ↓
Portfolio Service Receives Event
    ↓
PostgreSQL UPSERT (investments.holdings)
    ↓
Data Persisted ✓
```

## 🔧 Troubleshooting

### Service Won't Start

**Error**: `failed to connect to database`

**Solution**: Check PostgreSQL is running and credentials are correct
```bash
pg_isready -h localhost -p 5432
```

### Table Not Found

**Error**: `relation "investments.holdings" does not exist`

**Solution**: Run the database migrations from `infra/db`
```bash
# Check if migrations have been run
psql -h localhost -U garcios -d portfolio -c "\dt investments.*"
```

### UPSERT Fails

**Error**: `duplicate key value violates unique constraint`

**Solution**: This shouldn't happen with ON CONFLICT clause. Check the query syntax.

## 📝 Next Steps

1. ✅ PostgreSQL repository implemented
2. ✅ Service updated to use PostgreSQL
3. ⏳ Add database connection health checks
4. ⏳ Add metrics for database operations
5. ⏳ Add connection pooling configuration
6. ⏳ Add database query logging
7. ⏳ Consider adding a gRPC endpoint to query holdings

## 🔄 Rollback (if needed)

To rollback to in-memory storage:

```go
// In cmd/server/main.go
// Comment out PostgreSQL setup
// db, err := infrastructure.NewPostgresDB()
// ...

// Use in-memory repository
repo := repository.NewInMemoryHoldingRepository()
```

---

**Status**: ✅ Portfolio Service successfully migrated to PostgreSQL!
**Database**: `investments.holdings` table
**Ready for**: Production deployment with persistent storage
