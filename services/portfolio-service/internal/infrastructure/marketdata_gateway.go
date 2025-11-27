package infrastructure

import (
	"context"
	"fmt"
	"os"
	"time"

	pb "github.com/garcios/portfolio-insights/services/marketdata-service/proto/marketdata"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MarketDataGateway struct {
	client     pb.MarketDataServiceClient
	conn       *grpc.ClientConn
	cache      *PriceCache
	assetCache *AssetCache
}

func NewMarketDataGateway(cache *PriceCache, assetCache *AssetCache) (*MarketDataGateway, error) {
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
		client:     client,
		conn:       conn,
		cache:      cache,
		assetCache: assetCache,
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
	start := time.Now()

	// Try cache first
	if g.cache != nil {
		cacheStart := time.Now()
		cached, err := g.cache.Get(ctx, symbol)
		metrics.RecordCacheOperation("get", "price", cached != nil, time.Since(cacheStart).Seconds())

		if err == nil && cached != nil {
			metrics.RecordPriceFetch("cache", 1)
			return cached.Price, nil
		}
	}

	// Cache miss - fetch from service
	req := &pb.GetLatestPriceRequest{
		Symbol: symbol,
	}

	serviceStart := time.Now()
	resp, err := g.client.GetLatestPrice(ctx, req)
	duration := time.Since(serviceStart).Seconds()

	if err != nil {
		metrics.RecordMarketDataRequest("get_latest_price", "error", duration)
		return 0, fmt.Errorf("failed to get latest price for %s: %w", symbol, err)
	}
	metrics.RecordMarketDataRequest("get_latest_price", "success", duration)

	if resp.Price == nil {
		return 0, fmt.Errorf("no price data available for %s", symbol)
	}

	price := resp.Price.Price
	timestamp := resp.Price.Timestamp.AsTime()
	metrics.RecordPriceFetch("service", 1)

	// Cache the result
	if g.cache != nil {
		cacheStart := time.Now()
		err := g.cache.Set(ctx, symbol, price, timestamp)
		metrics.RecordCacheOperation("set", "price", err == nil, time.Since(cacheStart).Seconds())
	}

	metrics.RecordMarketDataRequest("total_operation", "success", time.Since(start).Seconds())
	return price, nil
}

// GetCurrentPrices fetches current prices for multiple symbols
// Uses cache-aside pattern with batch operations
func (g *MarketDataGateway) GetCurrentPrices(ctx context.Context, symbols []string) (map[string]float64, error) {
	start := time.Now()
	if len(symbols) == 0 {
		return make(map[string]float64), nil
	}

	prices := make(map[string]float64)
	var uncachedSymbols []string

	// Check cache for all symbols
	if g.cache != nil {
		cacheStart := time.Now()
		cachedPrices, err := g.cache.GetMultiple(ctx, symbols)
		metrics.RecordCacheOperation("get_multiple", "price", err == nil, time.Since(cacheStart).Seconds())

		if err == nil {
			for symbol, cached := range cachedPrices {
				prices[symbol] = cached.Price
			}
			metrics.RecordPriceFetch("cache", len(prices))
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

		serviceStart := time.Now()
		resp, err := g.client.GetLatestPrices(ctx, req)
		duration := time.Since(serviceStart).Seconds()

		if err != nil {
			metrics.RecordMarketDataRequest("get_latest_prices_batch", "error", duration)
			return nil, fmt.Errorf("failed to get latest prices: %w", err)
		}
		metrics.RecordMarketDataRequest("get_latest_prices_batch", "success", duration)

		// Process response and cache results
		fetchedPrices := make(map[string]float64)
		timestamp := time.Now()
		count := 0

		for symbol, assetPrice := range resp.Prices {
			if assetPrice != nil {
				prices[symbol] = assetPrice.Price
				fetchedPrices[symbol] = assetPrice.Price
				if !assetPrice.Timestamp.AsTime().IsZero() {
					timestamp = assetPrice.Timestamp.AsTime()
				}
				count++
			}
		}
		metrics.RecordPriceFetch("service", count)

		// Cache the fetched prices
		if g.cache != nil && len(fetchedPrices) > 0 {
			cacheStart := time.Now()
			err := g.cache.SetMultiple(ctx, fetchedPrices, timestamp)
			metrics.RecordCacheOperation("set_multiple", "price", err == nil, time.Since(cacheStart).Seconds())
		}
	}

	metrics.RecordMarketDataRequest("total_batch_operation", "success", time.Since(start).Seconds())
	return prices, nil
}

// GetCurrencyRate fetches the latest currency exchange rate
func (g *MarketDataGateway) GetCurrencyRate(ctx context.Context, baseCurrency, targetCurrency string) (float64, error) {
	// If currencies are the same, rate is 1.0
	if baseCurrency == targetCurrency {
		return 1.0, nil
	}

	start := time.Now()
	req := &pb.GetLatestCurrencyRateRequest{
		BaseCurrency:   baseCurrency,
		TargetCurrency: targetCurrency,
	}

	resp, err := g.client.GetLatestCurrencyRate(ctx, req)
	duration := time.Since(start).Seconds()

	if err != nil {
		metrics.RecordMarketDataRequest("get_currency_rate", "error", duration)
		return 0, fmt.Errorf("failed to get currency rate for %s/%s: %w", baseCurrency, targetCurrency, err)
	}
	metrics.RecordMarketDataRequest("get_currency_rate", "success", duration)

	if resp.CurrencyRate == nil {
		return 0, fmt.Errorf("no currency rate available for %s/%s", baseCurrency, targetCurrency)
	}

	return resp.CurrencyRate.Rate, nil
}

// GetAsset fetches asset details for a symbol
// Uses cache-aside pattern
func (g *MarketDataGateway) GetAsset(ctx context.Context, symbol string) (*pb.Asset, error) {
	start := time.Now()

	// Try cache first
	if g.assetCache != nil {
		cacheStart := time.Now()
		cached, err := g.assetCache.Get(ctx, symbol)
		metrics.RecordCacheOperation("get", "asset", cached != nil, time.Since(cacheStart).Seconds())

		if err == nil && cached != nil {
			return &pb.Asset{
				Symbol:   cached.Symbol,
				Name:     cached.Name,
				Type:     cached.Type,
				Exchange: cached.Exchange,
				Currency: cached.Currency,
			}, nil
		}
	}

	// Cache miss - fetch from service
	req := &pb.GetAssetRequest{
		Symbol: symbol,
	}

	serviceStart := time.Now()
	resp, err := g.client.GetAsset(ctx, req)
	duration := time.Since(serviceStart).Seconds()

	if err != nil {
		metrics.RecordMarketDataRequest("get_asset", "error", duration)
		return nil, fmt.Errorf("failed to get asset for %s: %w", symbol, err)
	}
	metrics.RecordMarketDataRequest("get_asset", "success", duration)

	if resp.Asset == nil {
		return nil, fmt.Errorf("no asset data available for %s", symbol)
	}

	// Cache the result
	if g.assetCache != nil {
		cacheStart := time.Now()
		err := g.assetCache.Set(ctx, resp.Asset)
		metrics.RecordCacheOperation("set", "asset", err == nil, time.Since(cacheStart).Seconds())
	}

	metrics.RecordMarketDataRequest("total_asset_operation", "success", time.Since(start).Seconds())
	return resp.Asset, nil
}

// GetAssetName fetches just the name of the asset
func (g *MarketDataGateway) GetAssetName(ctx context.Context, symbol string) (string, error) {
	asset, err := g.GetAsset(ctx, symbol)
	if err != nil {
		return "", err
	}
	return asset.Name, nil
}
