-- Drop currency_rates table and its indexes
DROP INDEX IF EXISTS marketdata.idx_currency_rates_date;
DROP INDEX IF EXISTS marketdata.idx_currency_rates_lookup;
DROP TABLE IF EXISTS marketdata.currency_rates;
