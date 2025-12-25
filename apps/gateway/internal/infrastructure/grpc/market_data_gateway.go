package grpc

import (
	"context"
	"fmt"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
	marketdatapb "github.com/garcios/portfolio-insights/services/marketdata-service/marketdata"
)

// MarketDataGRPCGateway implements the MarketDataGateway interface using gRPC
type MarketDataGRPCGateway struct {
	client marketdatapb.MarketDataServiceClient
}

// NewMarketDataGRPCGateway creates a new MarketDataGRPCGateway
func NewMarketDataGRPCGateway(client marketdatapb.MarketDataServiceClient) gateway.MarketDataGateway {
	return &MarketDataGRPCGateway{
		client: client,
	}
}

// GetExchangeRate retrieves the exchange rate between two currencies
func (g *MarketDataGRPCGateway) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (float64, error) {
	// If currencies are the same, rate is 1.0
	if fromCurrency == toCurrency {
		return 1.0, nil
	}

	req := &marketdatapb.GetLatestCurrencyRateRequest{
		BaseCurrency:   fromCurrency,
		TargetCurrency: toCurrency,
	}

	resp, err := g.client.GetLatestCurrencyRate(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	return resp.Rate, nil
}
