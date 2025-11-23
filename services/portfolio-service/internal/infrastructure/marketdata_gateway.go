package infrastructure

import (
	"context"
	"fmt"
	"os"
	"time"

	pb "github.com/garcios/portfolio-insights/services/marketdata-service/proto/marketdata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MarketDataGateway struct {
	client pb.MarketDataServiceClient
	conn   *grpc.ClientConn
	cache  *PriceCache
}

func NewMarketDataGateway(cache *PriceCache) (*MarketDataGateway, error) {
	marketDataAddr := os.Getenv("MARKETDATA_SERVICE_ADDR")
	if marketDataAddr == "" {
		marketDataAddr = "localhost:50054"
	}

	conn, err := grpc.NewClient(
		marketDataAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to marketdata service: %w", err)
	}

	client := pb.NewMarketDataServiceClient(conn)

	return &MarketDataGateway{
		client: client,
		conn:   conn,
		cache:  cache,
	}, nil
}

func (g *MarketDataGateway) Close() error {
	if g.conn != nil {
		return g.conn.Close()
	}
	return nil
}

// GetCurrentPrice fetches the current price for a single symbol
// Uses cache-aside pattern: check cache first, then fetch from service
func (g *MarketDataGateway) GetCurrentPrice(ctx context.Context, symbol string) (float64, error) {
	// Try cache first
	if g.cache != nil {
		cached, err := g.cache.Get(ctx, symbol)
		if err == nil && cached != nil {
			return cached.Price, nil
		}
	}

	// Cache miss - fetch from service
	req := &pb.GetLatestPriceRequest{
		Symbol: symbol,
	}

	resp, err := g.client.GetLatestPrice(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest price for %s: %w", symbol, err)
	}

	if resp.Price == nil {
		return 0, fmt.Errorf("no price data available for %s", symbol)
	}

	price := resp.Price.Price
	timestamp := resp.Price.Timestamp.AsTime()

	// Cache the result
	if g.cache != nil {
		_ = g.cache.Set(ctx, symbol, price, timestamp) // Ignore cache errors
	}

	return price, nil
}

// GetCurrentPrices fetches current prices for multiple symbols
// Uses cache-aside pattern with batch operations
func (g *MarketDataGateway) GetCurrentPrices(ctx context.Context, symbols []string) (map[string]float64, error) {
	if len(symbols) == 0 {
		return make(map[string]float64), nil
	}

	prices := make(map[string]float64)
	var uncachedSymbols []string

	// Check cache for all symbols
	if g.cache != nil {
		cachedPrices, err := g.cache.GetMultiple(ctx, symbols)
		if err == nil {
			for symbol, cached := range cachedPrices {
				prices[symbol] = cached.Price
			}
		}

		// Determine which symbols need to be fetched
		for _, symbol := range symbols {
			if _, found := prices[symbol]; !found {
				uncachedSymbols = append(uncachedSymbols, symbol)
			}
		}
	} else {
		uncachedSymbols = symbols
	}

	// Fetch uncached symbols from service
	if len(uncachedSymbols) > 0 {
		req := &pb.GetLatestPricesRequest{
			Symbols: uncachedSymbols,
		}

		resp, err := g.client.GetLatestPrices(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest prices: %w", err)
		}

		// Process response and cache results
		fetchedPrices := make(map[string]float64)
		timestamp := time.Now()

		for symbol, assetPrice := range resp.Prices {
			if assetPrice != nil {
				prices[symbol] = assetPrice.Price
				fetchedPrices[symbol] = assetPrice.Price
				if !assetPrice.Timestamp.AsTime().IsZero() {
					timestamp = assetPrice.Timestamp.AsTime()
				}
			}
		}

		// Cache the fetched prices
		if g.cache != nil && len(fetchedPrices) > 0 {
			_ = g.cache.SetMultiple(ctx, fetchedPrices, timestamp) // Ignore cache errors
		}
	}

	return prices, nil
}
