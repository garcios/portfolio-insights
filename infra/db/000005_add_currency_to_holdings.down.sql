-- Remove currency column from holdings table
DROP INDEX IF EXISTS investments.idx_holdings_currency;
ALTER TABLE investments.holdings DROP COLUMN IF EXISTS currency;
