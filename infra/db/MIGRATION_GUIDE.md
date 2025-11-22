# Database Schema Migration Guide

## Quick Reference: Schema Changes

This guide provides a quick reference for the database schema reorganization.

## Schema Mapping

| Old Table Name | New Fully Qualified Name | Schema |
|----------------|--------------------------|--------|
| `users` | `customers.users` | customers |
| `transactions` | `txn.transactions` | txn |
| `assets` | `marketdata.assets` | marketdata |
| `asset_prices` | `marketdata.asset_prices` | marketdata |
| `holdings` | `investments.holdings` | investments |
| `portfolio_history` | `investments.portfolio_history` | investments |

## Migration Checklist

### 1. Database Migrations
- [x] Updated `000001_create_users_table.up.sql` - Added `customers` schema
- [x] Updated `000001_create_users_table.down.sql` - Drop schema on rollback
- [x] Updated `000002_create_transactions_table.up.sql` - Added `txn` schema
- [x] Updated `000002_create_transactions_table.down.sql` - Drop schema on rollback
- [x] Updated `000003_create_market_data_tables.up.sql` - Added `marketdata` schema
- [x] Updated `000003_create_market_data_tables.down.sql` - Drop schema on rollback
- [x] Updated `000004_create_portfolio_tables.up.sql` - Added `investments` schema
- [x] Updated `000004_create_portfolio_tables.down.sql` - Drop schema on rollback

### 2. Application Code Updates Needed

The following services need to be updated to use fully qualified table names:

#### User Service
```go
// Update all queries from:
"SELECT * FROM users WHERE id = $1"

// To:
"SELECT * FROM customers.users WHERE id = $1"
```

#### Transaction Service
```go
// Update all queries from:
"INSERT INTO transactions (...) VALUES (...)"

// To:
"INSERT INTO txn.transactions (...) VALUES (...)"
```

#### MarketData Service
```go
// Update all queries from:
"SELECT * FROM assets WHERE symbol = $1"
"SELECT * FROM asset_prices WHERE asset_id = $1"

// To:
"SELECT * FROM marketdata.assets WHERE symbol = $1"
"SELECT * FROM marketdata.asset_prices WHERE asset_id = $1"
```

#### Portfolio Service
```go
// Update all queries from:
"SELECT * FROM holdings WHERE user_id = $1"
"SELECT * FROM portfolio_history WHERE user_id = $1"

// To:
"SELECT * FROM investments.holdings WHERE user_id = $1"
"SELECT * FROM investments.portfolio_history WHERE user_id = $1"
```

### 3. Search Patterns for Code Updates

Use these search patterns to find all occurrences that need updating:

```bash
# Find all SQL queries that need updating
grep -r "FROM users" services/
grep -r "INTO users" services/
grep -r "UPDATE users" services/

grep -r "FROM transactions" services/
grep -r "INTO transactions" services/
grep -r "UPDATE transactions" services/

grep -r "FROM assets" services/
grep -r "INTO assets" services/
grep -r "UPDATE assets" services/

grep -r "FROM asset_prices" services/
grep -r "INTO asset_prices" services/

grep -r "FROM holdings" services/
grep -r "INTO holdings" services/
grep -r "UPDATE holdings" services/

grep -r "FROM portfolio_history" services/
grep -r "INTO portfolio_history" services/
```

## Testing the Migration

### 1. Test in Development Environment

```bash
# Navigate to db directory
cd infra/db

# Run migrations
make migrate-up

# Verify schemas were created
psql -U postgres -d portfolio_insights -c "\dn"

# Verify tables in each schema
psql -U postgres -d portfolio_insights -c "\dt customers.*"
psql -U postgres -d portfolio_insights -c "\dt txn.*"
psql -U postgres -d portfolio_insights -c "\dt marketdata.*"
psql -U postgres -d portfolio_insights -c "\dt investments.*"

# Test rollback
make migrate-down

# Re-apply
make migrate-up
```

### 2. Verify Schema Structure

```sql
-- List all schemas
SELECT schema_name 
FROM information_schema.schemata 
WHERE schema_name IN ('customers', 'txn', 'marketdata', 'investments');

-- List tables in each schema
SELECT table_schema, table_name 
FROM information_schema.tables 
WHERE table_schema IN ('customers', 'txn', 'marketdata', 'investments')
ORDER BY table_schema, table_name;

-- Verify indexes
SELECT schemaname, tablename, indexname 
FROM pg_indexes 
WHERE schemaname IN ('customers', 'txn', 'marketdata', 'investments')
ORDER BY schemaname, tablename;
```

### 3. Test Sample Queries

```sql
-- Test customers schema
SELECT COUNT(*) FROM customers.users;

-- Test txn schema
SELECT COUNT(*) FROM txn.transactions;

-- Test marketdata schema
SELECT COUNT(*) FROM marketdata.assets;
SELECT COUNT(*) FROM marketdata.asset_prices;

-- Test investments schema
SELECT COUNT(*) FROM investments.holdings;
SELECT COUNT(*) FROM investments.portfolio_history;
```

## Deployment Steps

### Development
1. ✅ Update migration files (COMPLETED)
2. ⏳ Update application code to use schema-qualified names
3. ⏳ Run migrations: `make migrate-up`
4. ⏳ Test all services
5. ⏳ Verify data integrity

### Staging
1. Backup database
2. Run migrations
3. Deploy updated services
4. Run integration tests
5. Verify all functionality

### Production
1. Schedule maintenance window
2. Backup database
3. Run migrations
4. Deploy updated services
5. Monitor for errors
6. Verify all functionality
7. Keep rollback plan ready

## Rollback Plan

If issues are encountered:

```bash
# Rollback migrations
cd infra/db
make migrate-down

# Redeploy previous version of services
# Restore database from backup if necessary
```

## Common Issues and Solutions

### Issue: Schema does not exist
```
ERROR: schema "customers" does not exist
```
**Solution**: Run migrations: `cd infra/db && make migrate-up`

### Issue: Table not found
```
ERROR: relation "users" does not exist
```
**Solution**: Update query to use `customers.users` instead of `users`

### Issue: Permission denied
```
ERROR: permission denied for schema customers
```
**Solution**: Grant permissions:
```sql
GRANT USAGE ON SCHEMA customers TO your_user;
GRANT ALL ON ALL TABLES IN SCHEMA customers TO your_user;
```

### Issue: Foreign key constraint violation
```
ERROR: foreign key constraint violation
```
**Solution**: Ensure referenced tables use schema-qualified names in foreign key definitions

## Files Modified

### Migration Files
- `infra/db/000001_create_users_table.up.sql`
- `infra/db/000001_create_users_table.down.sql`
- `infra/db/000002_create_transactions_table.up.sql`
- `infra/db/000002_create_transactions_table.down.sql`
- `infra/db/000003_create_market_data_tables.up.sql`
- `infra/db/000003_create_market_data_tables.down.sql`
- `infra/db/000004_create_portfolio_tables.up.sql`
- `infra/db/000004_create_portfolio_tables.down.sql`

### Documentation Files
- `infra/db/SCHEMA_ORGANIZATION.md` (NEW)
- `infra/db/MIGRATION_GUIDE.md` (THIS FILE)

## Next Steps

1. **Review the changes**: Examine all modified migration files
2. **Update application code**: Search for and update all table references
3. **Test locally**: Run migrations and test all services
4. **Update documentation**: Document any service-specific changes
5. **Deploy to staging**: Test in staging environment
6. **Deploy to production**: Follow deployment checklist

## Support

For questions or issues:
- Review `SCHEMA_ORGANIZATION.md` for detailed schema documentation
- Check application logs for specific error messages
- Verify database connection strings include correct schema search paths
- Ensure all SQL queries use fully qualified table names

---

**Last Updated**: 2025-11-22  
**Migration Status**: ✅ Database scripts updated, ⏳ Application code updates pending
