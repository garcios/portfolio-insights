-- Create cash_balances table for cleaner separation from equity holdings
CREATE TABLE IF NOT EXISTS investments.cash_balances (
    user_id UUID NOT NULL,
    currency VARCHAR(3) NOT NULL,
    balance DECIMAL(20, 8) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (user_id, currency),
    CONSTRAINT fk_cash_user FOREIGN KEY (user_id) 
        REFERENCES customers.users(id) ON DELETE CASCADE
);

-- Create index for user lookups
CREATE INDEX idx_cash_balances_user_id ON investments.cash_balances(user_id);

-- Add comment
COMMENT ON TABLE investments.cash_balances IS 
    'Tracks cash balances per user per currency. Separates cash from equity holdings.';
