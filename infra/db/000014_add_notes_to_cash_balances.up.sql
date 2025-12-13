ALTER TABLE investments.cash_balances
ADD COLUMN IF NOT EXISTS notes VARCHAR(100);
