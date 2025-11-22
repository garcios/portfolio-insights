-- Create investments schema for portfolio and holdings-related tables
CREATE SCHEMA IF NOT EXISTS investments;

-- Create holdings table in investments schema
CREATE TABLE IF NOT EXISTS investments.holdings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL DEFAULT 0,
    average_cost_basis DECIMAL(20, 8) NOT NULL DEFAULT 0, -- Average price paid per share
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, symbol)
);

-- Create portfolio_history table in investments schema
CREATE TABLE IF NOT EXISTS investments.portfolio_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    total_value DECIMAL(20, 8) NOT NULL,
    total_cost_basis DECIMAL(20, 8) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_holdings_user_id ON investments.holdings(user_id);
CREATE INDEX idx_portfolio_history_user_id_timestamp ON investments.portfolio_history(user_id, timestamp DESC);
