package infrastructure

import (
	"context"
	"fmt"
	"os"

	marketdatapb "github.com/garcios/portfolio-insights/services/marketdata-service/proto/marketdata"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type marketDataGateway struct {
	client marketdatapb.MarketDataServiceClient
}

// NewMarketDataGateway creates a new market data gateway.
func NewMarketDataGateway() (domain.MarketDataGateway, error) {
	host := os.Getenv("MARKETDATA_SERVICE_HOST")
	port := os.Getenv("MARKETDATA_SERVICE_PORT")

	if host == "" {
		host = "marketdata-service"
	}
	if port == "" {
		port = "50054"
	}

	target := fmt.Sprintf("%s:%s", host, port)
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
