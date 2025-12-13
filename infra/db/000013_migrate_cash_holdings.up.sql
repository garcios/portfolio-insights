-- Migrate existing CASH-* holdings to cash_balances table
INSERT INTO investments.cash_balances (user_id, currency, balance, created_at, updated_at)
SELECT 
    user_id,
    SUBSTRING(symbol FROM 6) AS currency,  -- Extract 'USD' from 'CASH-USD'
    quantity AS balance,
    updated_at AS created_at,
    updated_at
FROM investments.holdings
WHERE symbol LIKE 'CASH-%'
ON CONFLICT (user_id, currency) DO UPDATE
    SET balance = EXCLUDED.balance,
        updated_at = EXCLUDED.updated_at;

-- Delete migrated CASH-* holdings
DELETE FROM investments.holdings WHERE symbol LIKE 'CASH-%';
