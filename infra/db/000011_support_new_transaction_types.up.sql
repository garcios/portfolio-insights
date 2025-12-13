-- Make equity-specific fields nullable to support cash-only transactions
ALTER TABLE txn.transactions 
  ALTER COLUMN symbol DROP NOT NULL,
  ALTER COLUMN quantity DROP NOT NULL,
  ALTER COLUMN price_per_share DROP NOT NULL;

-- Add amount field for cash transactions (INT, DIV, DEP, WIT)
ALTER TABLE txn.transactions 
  ADD COLUMN amount DECIMAL(20, 8);

-- Add composite index for filtering by user and transaction type
CREATE INDEX idx_transactions_user_type ON txn.transactions(user_id, type);

-- Add index on executed_at for time-based queries
CREATE INDEX idx_transactions_executed_at ON txn.transactions(executed_at);

-- Add index on type for transaction type filtering
CREATE INDEX idx_transactions_type ON txn.transactions(type);

-- Note: Supported transaction types are now: BUY, SELL, INT, DIV, DEP, WIT
-- VARCHAR(10) is sufficient for all transaction type codes
