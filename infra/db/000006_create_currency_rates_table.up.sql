-- Create currency_rates table in marketdata schema
CREATE TABLE IF NOT EXISTS marketdata.currency_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_currency VARCHAR(3) NOT NULL,
    target_currency VARCHAR(3) NOT NULL,
    rate DECIMAL(20, 8) NOT NULL,
    rate_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_currency_rate UNIQUE (base_currency, target_currency, rate_date)
);

-- Create index for efficient lookups by currency pair and date
CREATE INDEX idx_currency_rates_lookup ON marketdata.currency_rates(base_currency, target_currency, rate_date DESC);

-- Create index for date-based queries
CREATE INDEX idx_currency_rates_date ON marketdata.currency_rates(rate_date DESC);
