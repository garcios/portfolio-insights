-- Rollback: Migrate cash_balances back to holdings
INSERT INTO investments.holdings (user_id, symbol, quantity, average_cost_basis, currency, created_at, updated_at)
SELECT 
    user_id,
    'CASH-' || currency AS symbol,
    balance AS quantity,
    1.0 AS average_cost_basis,
    currency,
    created_at,
    updated_at
FROM investments.cash_balances
ON CONFLICT (user_id, symbol) DO UPDATE
    SET quantity = EXCLUDED.quantity,
        updated_at = EXCLUDED.updated_at;

-- Delete cash_balances after rollback
DELETE FROM investments.cash_balances;
