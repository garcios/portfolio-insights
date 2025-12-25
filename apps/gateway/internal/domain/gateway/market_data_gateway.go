package gateway

import (
	"context"
)

// MarketDataGateway defines the interface for interacting with the market data service
type MarketDataGateway interface {
	// GetExchangeRate retrieves the exchange rate between two currencies
	GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (float64, error)
}
