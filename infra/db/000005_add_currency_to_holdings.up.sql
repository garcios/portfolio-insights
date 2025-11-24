-- Add currency column to holdings table
ALTER TABLE investments.holdings 
ADD COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'USD';

-- Create index on currency for potential filtering
CREATE INDEX idx_holdings_currency ON investments.holdings(currency);
