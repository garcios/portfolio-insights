CREATE TABLE IF NOT EXISTS investments.portfolio_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- State Snapshots
    holdings_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    cash_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    realized_gains_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    net_invested TEXT NOT NULL DEFAULT '0',
    
    transaction_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES customers.users(id) ON DELETE CASCADE,
    UNIQUE (user_id, timestamp)
);

CREATE INDEX idx_portfolio_snapshots_user_timestamp 
ON investments.portfolio_snapshots(user_id, timestamp DESC);
