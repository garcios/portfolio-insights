// Package infrastructure provides external service integrations and low-level implementations.
package infrastructure

import (
	"context"
	"fmt"

	marketdatapb "github.com/garcios/portfolio-insights/services/marketdata-service/proto/marketdata"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/config"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type marketDataGateway struct {
	client marketdatapb.MarketDataServiceClient
}

// NewMarketDataGateway creates a new market data gateway.
func NewMarketDataGateway(cfg config.Config) (domain.MarketDataGateway, error) {
	target := cfg.MarketDataServiceAddr
	if target == "" {
		target = "localhost:50054"
	}

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to marketdata service: %w", err)
	}

	client := marketdatapb.NewMarketDataServiceClient(conn)
	return &marketDataGateway{client: client}, nil
}

func (g *marketDataGateway) Exists(ctx context.Context, symbol string) (bool, error) {
	_, err := g.client.GetAsset(ctx, &marketdatapb.GetAssetRequest{Symbol: symbol})
	if err != nil {
		// TODO: Check specific gRPC error code for NotFound
		return false, nil
	}
	return true, nil
}
