-- Drop indexes
DROP INDEX IF EXISTS txn.idx_transactions_type;
DROP INDEX IF EXISTS txn.idx_transactions_executed_at;
DROP INDEX IF EXISTS txn.idx_transactions_user_type;

-- Remove amount column
ALTER TABLE txn.transactions 
  DROP COLUMN IF EXISTS amount;

-- Restore NOT NULL constraints on equity-specific fields
ALTER TABLE txn.transactions 
  ALTER COLUMN symbol SET NOT NULL,
  ALTER COLUMN quantity SET NOT NULL,
  ALTER COLUMN price_per_share SET NOT NULL;
