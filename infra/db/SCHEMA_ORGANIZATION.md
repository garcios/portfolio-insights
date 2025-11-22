# Database Schema Organization

This document describes the database schema organization for the Portfolio Insights application.

## Overview

The database is organized into **4 separate schemas**, each representing a distinct domain within the application. This separation provides better organization, security, and maintainability.

## Schema Structure

### 1. **customers** Schema
**Purpose**: User account and authentication data

**Tables**:
- `customers.users`

**Description**: Contains all user-related information including authentication credentials and profile data.

```sql
CREATE SCHEMA IF NOT EXISTS customers;

CREATE TABLE customers.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

---

### 2. **txn** Schema
**Purpose**: Transaction records and trading activity

**Tables**:
- `txn.transactions`

**Description**: Stores all buy/sell transactions executed by users.

```sql
CREATE SCHEMA IF NOT EXISTS txn;

CREATE TABLE txn.transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    type VARCHAR(10) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL,
    price_per_share DECIMAL(20, 8) NOT NULL,
    executed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_user_id ON txn.transactions(user_id);
```

---

### 3. **marketdata** Schema
**Purpose**: Market data, asset information, and pricing

**Tables**:
- `marketdata.assets`
- `marketdata.asset_prices`

**Description**: Contains asset definitions and historical price data.

```sql
CREATE SCHEMA IF NOT EXISTS marketdata;

CREATE TABLE marketdata.assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- e.g., 'STOCK', 'CRYPTO'
    exchange VARCHAR(50),
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE marketdata.asset_prices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID NOT NULL REFERENCES marketdata.assets(id) ON DELETE CASCADE,
    price DECIMAL(20, 8) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_assets_symbol ON marketdata.assets(symbol);
CREATE INDEX idx_asset_prices_asset_id_timestamp ON marketdata.asset_prices(asset_id, timestamp DESC);
```

---

### 4. **investments** Schema
**Purpose**: Portfolio holdings and performance tracking

**Tables**:
- `investments.holdings`
- `investments.portfolio_history`

**Description**: Tracks current holdings and historical portfolio performance.

```sql
CREATE SCHEMA IF NOT EXISTS investments;

CREATE TABLE investments.holdings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL DEFAULT 0,
    average_cost_basis DECIMAL(20, 8) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, symbol)
);

CREATE TABLE investments.portfolio_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    total_value DECIMAL(20, 8) NOT NULL,
    total_cost_basis DECIMAL(20, 8) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_holdings_user_id ON investments.holdings(user_id);
CREATE INDEX idx_portfolio_history_user_id_timestamp ON investments.portfolio_history(user_id, timestamp DESC);
```

---

## Schema Relationships

```
┌─────────────────┐
│ customers.users │
└────────┬────────┘
         │
         │ user_id (FK reference)
         │
    ┌────┴────────────────────────────┐
    │                                 │
    ▼                                 ▼
┌──────────────────┐        ┌────────────────────┐
│ txn.transactions │        │ investments.holdings│
└──────────────────┘        └────────────────────┘
                                     │
                            ┌────────┴────────┐
                            │                 │
                            ▼                 ▼
                  ┌──────────────────┐  ┌─────────────────────────┐
                  │ marketdata.assets│  │ investments.portfolio_  │
                  └────────┬─────────┘  │      history            │
                           │            └─────────────────────────┘
                           │
                           ▼
                  ┌────────────────────┐
                  │ marketdata.asset_  │
                  │     prices         │
                  └────────────────────┘
```

**Note**: Foreign key relationships are currently implemented at the application level, not at the database level. This is a common pattern in microservices architectures where each service owns its schema.

---

## Benefits of Schema Separation

### 1. **Domain Isolation**
Each schema represents a clear business domain, making it easier to understand and maintain the database structure.

### 2. **Security & Access Control**
Different schemas can have different access permissions:
- Read-only access to `marketdata` for most services
- Restricted write access to `customers` for authentication service only
- Controlled access to `txn` for transaction service

### 3. **Microservices Alignment**
Each schema aligns with a microservice:
- `customers` → User Service
- `txn` → Transaction Service
- `marketdata` → MarketData Service
- `investments` → Portfolio Service

### 4. **Easier Maintenance**
- Clear separation makes it easier to backup specific domains
- Simpler to understand data ownership
- Reduces risk of accidental cross-domain queries

### 5. **Scalability**
- Each schema can potentially be moved to a separate database in the future
- Easier to implement schema-specific optimizations
- Better prepared for database sharding if needed

---

## Migration Files

The database migrations are organized as follows:

| Migration | Schema | Tables | Description |
|-----------|--------|--------|-------------|
| `000001_create_users_table` | `customers` | `users` | User accounts and authentication |
| `000002_create_transactions_table` | `txn` | `transactions` | Buy/sell transaction records |
| `000003_create_market_data_tables` | `marketdata` | `assets`, `asset_prices` | Asset definitions and pricing |
| `000004_create_portfolio_tables` | `investments` | `holdings`, `portfolio_history` | Portfolio holdings and history |

---

## Querying Across Schemas

When querying tables, always use the fully qualified table name:

```sql
-- Good: Fully qualified
SELECT * FROM customers.users WHERE email = 'user@example.com';

-- Bad: Unqualified (may fail or use wrong schema)
SELECT * FROM users WHERE email = 'user@example.com';
```

### Example Queries

**Get user with their holdings:**
```sql
SELECT 
    u.id,
    u.name,
    u.email,
    h.symbol,
    h.quantity,
    h.average_cost_basis
FROM customers.users u
LEFT JOIN investments.holdings h ON u.id = h.user_id
WHERE u.id = 'user-uuid-here';
```

**Get transaction history with asset details:**
```sql
SELECT 
    t.id,
    t.type,
    t.quantity,
    t.price_per_share,
    t.executed_at,
    a.name as asset_name,
    a.symbol
FROM txn.transactions t
JOIN marketdata.assets a ON t.symbol = a.symbol
WHERE t.user_id = 'user-uuid-here'
ORDER BY t.executed_at DESC;
```

**Get portfolio value with current prices:**
```sql
SELECT 
    h.symbol,
    h.quantity,
    h.average_cost_basis,
    ap.price as current_price,
    (h.quantity * ap.price) as current_value,
    ((ap.price - h.average_cost_basis) / h.average_cost_basis * 100) as gain_loss_percent
FROM investments.holdings h
JOIN marketdata.assets a ON h.symbol = a.symbol
JOIN LATERAL (
    SELECT price 
    FROM marketdata.asset_prices 
    WHERE asset_id = a.id 
    ORDER BY timestamp DESC 
    LIMIT 1
) ap ON true
WHERE h.user_id = 'user-uuid-here';
```

---

## Running Migrations

### Apply All Migrations
```bash
cd infra/db
make migrate-up
```

### Rollback All Migrations
```bash
cd infra/db
make migrate-down
```

### Check Migration Status
```bash
cd infra/db
make migrate-status
```

---

## Application Code Updates

When updating application code to use the new schemas, ensure all table references include the schema prefix:

### Go Example
```go
// Before
query := "SELECT * FROM users WHERE id = $1"

// After
query := "SELECT * FROM customers.users WHERE id = $1"
```

### SQL Repository Pattern
```go
const (
    UsersTable           = "customers.users"
    TransactionsTable    = "txn.transactions"
    AssetsTable          = "marketdata.assets"
    AssetPricesTable     = "marketdata.asset_prices"
    HoldingsTable        = "investments.holdings"
    PortfolioHistoryTable = "investments.portfolio_history"
)
```

---

## Best Practices

1. **Always use fully qualified table names** in queries
2. **Set search_path per connection** if needed for a specific schema
3. **Grant minimal permissions** to each service for their schema
4. **Document cross-schema queries** as they indicate service dependencies
5. **Consider views** for common cross-schema queries
6. **Use transactions** when modifying multiple schemas

---

## Future Considerations

### Potential Enhancements
- Add foreign key constraints where appropriate
- Create views for common cross-schema queries
- Implement row-level security policies
- Add audit tables per schema
- Consider partitioning for large tables (e.g., `asset_prices`, `portfolio_history`)

### Database Sharding Strategy
If the application grows, schemas can be moved to separate databases:
- `customers` → User Database
- `txn` → Transactions Database
- `marketdata` → Market Data Database (read replicas)
- `investments` → Portfolio Database

---

## Troubleshooting

### Schema Not Found Error
```
ERROR: schema "customers" does not exist
```
**Solution**: Run migrations: `make migrate-up`

### Permission Denied
```
ERROR: permission denied for schema customers
```
**Solution**: Grant appropriate permissions to the database user

### Table Not Found
```
ERROR: relation "users" does not exist
```
**Solution**: Use fully qualified name: `customers.users`

---

## Summary

The database is now organized into 4 logical schemas:

| Schema | Purpose | Tables |
|--------|---------|--------|
| **customers** | User management | users |
| **txn** | Transactions | transactions |
| **marketdata** | Market data | assets, asset_prices |
| **investments** | Portfolio tracking | holdings, portfolio_history |

This organization provides better separation of concerns, improved security, and easier maintenance as the application scales.
