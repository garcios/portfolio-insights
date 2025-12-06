package infrastructure

import (
	"context"
	"fmt"
	"time"

	pb "github.com/garcios/portfolio-insights/services/marketdata-service/proto/marketdata"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/config"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/metrics"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MarketDataGateway acts as a gateway to the marketdata service.
type MarketDataGateway struct {
	client     pb.MarketDataServiceClient
	conn       *grpc.ClientConn
	cache      *PriceCache
	assetCache *AssetCache
}

// NewMarketDataGateway creates a new market data gateway.
func NewMarketDataGateway(cache *PriceCache, assetCache *AssetCache, cfg config.Config) (*MarketDataGateway, error) {
	marketDataAddr := cfg.MarketDataServiceAddr
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

// Close closes the connection to the marketdata service.
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
func (g *MarketDataGateway) GetCurrentPrices(ctx context.Context, symbols []string) (map[string]usecase.PriceData, error) {
	start := time.Now()
	if len(symbols) == 0 {
		return make(map[string]usecase.PriceData), nil
	}

	prices := make(map[string]usecase.PriceData)
	var uncachedSymbols []string

	// Check cache for all symbols
	if g.cache != nil {
		cacheStart := time.Now()
		cachedPrices, err := g.cache.GetMultiple(ctx, symbols)
		metrics.RecordCacheOperation("get_multiple", "price", err == nil, time.Since(cacheStart).Seconds())

		if err == nil {
			for symbol, cached := range cachedPrices {
				prices[symbol] = usecase.PriceData{
					Price:     cached.Price,
					Timestamp: cached.Timestamp,
				}
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
				timestamp := time.Now()
				if !assetPrice.Timestamp.AsTime().IsZero() {
					timestamp = assetPrice.Timestamp.AsTime()
				}
				prices[symbol] = usecase.PriceData{
					Price:     assetPrice.Price,
					Timestamp: timestamp,
				}
				fetchedPrices[symbol] = assetPrice.Price
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

// GetPriceOnDate fetches the closing price of an asset on a specific date,
// or attempts to find the price on the previous 6 days if the specific date has no data.
func (g *MarketDataGateway) GetPriceOnDate(ctx context.Context, symbol string, date time.Time) (float64, error) {
	// The maximum number of days to check, including the original date.
	const maxDaysToTry = 7

	currentDate := date.In(time.UTC) // Start with the original date, converting to UTC for consistency

	for i := 0; i < maxDaysToTry; i++ {
		// Define start and end of the day for the current attempt date
		// Use UTC to avoid timezone issues, assuming market data is stored/queried in UTC or date-based
		startOfDay := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, time.UTC)
		endOfDay := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 23, 59, 59, 999999999, time.UTC)

		req := &pb.GetHistoricalPricesRequest{
			Symbol:    symbol,
			StartTime: timestamppb.New(startOfDay),
			EndTime:   timestamppb.New(endOfDay),
			Interval:  "1d",
		}

		resp, err := g.client.GetHistoricalPrices(ctx, req)
		if err != nil {
			// If the gRPC call fails, return the error immediately
			return 0, fmt.Errorf("failed to get historical prices for %s on %s: %w", symbol, currentDate.Format("2006-01-02"), err)
		}

		if len(resp.Prices) > 0 {
			// **Success:** Price found for the current date attempt
			return resp.Prices[0].Price, nil
		}

		// **Failure for this day:** Move to the previous day for the next iteration
		// Subtract 24 hours to get to the day before
		currentDate = currentDate.AddDate(0, 0, -1)
	}

	// If the loop completes, no price was found within the maximum number of days checked (7 days total).
	return 0, fmt.Errorf("no price found for %s after checking %d days, starting from %s", symbol, maxDaysToTry, date.Format("2006-01-02"))
}

// GetCurrencyRateOnDate fetches the currency exchange rate on a specific date,
// or attempts to find the rate on the previous 6 days if the specific date has no data.
func (g *MarketDataGateway) GetCurrencyRateOnDate(ctx context.Context, baseCurrency, targetCurrency string, date time.Time) (float64, error) {
	// If currencies are the same, rate is 1.0 (This logic remains the same)
	if baseCurrency == targetCurrency {
		return 1.0, nil
	}

	// The maximum number of days to check, including the original date.
	const maxDaysToTry = 7

	// Start with the original date, converting to UTC for consistency
	currentDate := date.In(time.UTC)

	for i := 0; i < maxDaysToTry; i++ {
		// Define start and end of the day for the current attempt date
		// Use UTC to avoid timezone issues
		startOfDay := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, time.UTC)
		endOfDay := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 23, 59, 59, 999999999, time.UTC)

		req := &pb.GetHistoricalCurrencyRatesRequest{
			BaseCurrency:   baseCurrency,
			TargetCurrency: targetCurrency,
			StartTime:      timestamppb.New(startOfDay),
			EndTime:        timestamppb.New(endOfDay),
		}

		resp, err := g.client.GetHistoricalCurrencyRates(ctx, req)
		if err != nil {
			// If the gRPC call fails, return the error immediately, but include the failing date
			return 0, fmt.Errorf("failed to get historical currency rates for %s/%s on %s: %w", baseCurrency, targetCurrency, currentDate.Format("2006-01-02"), err)
		}

		if len(resp.Rates) > 0 {
			// **Success:** Rate found for the current date attempt
			// Return the first rate found
			return resp.Rates[0].Rate, nil
		}

		// **Failure for this day:** Move to the previous day for the next iteration
		// Subtract 24 hours to get to the day before
		currentDate = currentDate.AddDate(0, 0, -1)
	}

	// If the loop completes, no rate was found within the maximum number of days checked (7 days total).
	return 0, fmt.Errorf("no currency rate found for %s/%s after checking %d days, starting from %s", baseCurrency, targetCurrency, maxDaysToTry, date.Format("2006-01-02"))
}
