-- Add unique constraint to ensure one price per asset per timestamp
ALTER TABLE marketdata.asset_prices
ADD CONSTRAINT uq_asset_prices_asset_timestamp UNIQUE (asset_id, "timestamp");
