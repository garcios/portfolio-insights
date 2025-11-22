-- Create marketdata schema for market data-related tables
CREATE SCHEMA IF NOT EXISTS marketdata;

-- Create assets table in marketdata schema
CREATE TABLE IF NOT EXISTS marketdata.assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- e.g., 'STOCK', 'CRYPTO'
    exchange VARCHAR(50),
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create asset_prices table in marketdata schema
CREATE TABLE IF NOT EXISTS marketdata.asset_prices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID NOT NULL REFERENCES marketdata.assets(id) ON DELETE CASCADE,
    price DECIMAL(20, 8) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_assets_symbol ON marketdata.assets(symbol);
CREATE INDEX idx_asset_prices_asset_id_timestamp ON marketdata.asset_prices(asset_id, timestamp DESC);
