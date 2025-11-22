-- Create txn schema for transaction-related tables
CREATE SCHEMA IF NOT EXISTS txn;

-- Create transactions table in txn schema
CREATE TABLE IF NOT EXISTS txn.transactions (
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
