package infrastructure

import (
	"context"
	"fmt"
	"time"

	pb "github.com/garcios/portfolio-insights/services/marketdata-service/marketdata"
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
	history    *HistoricalCache
}

// NewMarketDataGateway creates a new market data gateway.
func NewMarketDataGateway(cache *PriceCache, assetCache *AssetCache, history *HistoricalCache, cfg config.Config) (*MarketDataGateway, error) {
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
		history:    history,
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
		Name: fmt.Sprintf("assets/%s", symbol),
	}

	serviceStart := time.Now()
	resp, err := g.client.GetLatestPrice(ctx, req)
	duration := time.Since(serviceStart).Seconds()

	if err != nil {
		metrics.RecordMarketDataRequest("get_latest_price", "error", duration)
		return 0, fmt.Errorf("failed to get latest price for %s: %w", symbol, err)
	}
	metrics.RecordMarketDataRequest("get_latest_price", "success", duration)

	if resp == nil {
		return 0, fmt.Errorf("no price data available for %s", symbol)
	}

	price := resp.Price
	timestamp := resp.Timestamp.AsTime()
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
		// Convert symbols to resource names
		var resourceNames []string
		for _, sym := range uncachedSymbols {
			resourceNames = append(resourceNames, fmt.Sprintf("assets/%s", sym))
		}

		req := &pb.GetLatestPricesRequest{
			Names: resourceNames,
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
		// Map resource names back to symbols
		resourceToSymbol := make(map[string]string)
		for _, sym := range uncachedSymbols {
			resourceToSymbol[fmt.Sprintf("assets/%s", sym)] = sym
		}

		pricesToCache := make(map[string]CachedPrice)
		count := 0

		for resourceName, assetPrice := range resp.Prices {
			if assetPrice != nil {
				symbol := resourceToSymbol[resourceName]
				if symbol == "" {
					// Fallback: the server might return the symbol directly as the key
					// instead of the resource name.
					symbol = resourceName
				}

				assetTimestamp := time.Now()
				if !assetPrice.Timestamp.AsTime().IsZero() {
					assetTimestamp = assetPrice.Timestamp.AsTime()
				}

				prices[symbol] = usecase.PriceData{
					Price:     assetPrice.Price,
					Timestamp: assetTimestamp,
				}

				pricesToCache[symbol] = CachedPrice{
					Symbol:    symbol,
					Price:     assetPrice.Price,
					Timestamp: assetTimestamp,
					CachedAt:  time.Now(),
				}
				count++
			}
		}
		metrics.RecordPriceFetch("service", count)

		// Cache the fetched prices
		if g.cache != nil && len(pricesToCache) > 0 {
			cacheStart := time.Now()
			err := g.cache.SetMultiple(ctx, pricesToCache)
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

	if resp == nil {
		return 0, fmt.Errorf("no currency rate available for %s/%s", baseCurrency, targetCurrency)
	}

	return resp.Rate, nil
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
		Name: fmt.Sprintf("assets/%s", symbol),
	}

	serviceStart := time.Now()
	resp, err := g.client.GetAsset(ctx, req)
	duration := time.Since(serviceStart).Seconds()

	if err != nil {
		metrics.RecordMarketDataRequest("get_asset", "error", duration)
		return nil, fmt.Errorf("failed to get asset for %s: %w", symbol, err)
	}
	metrics.RecordMarketDataRequest("get_asset", "success", duration)

	if resp == nil {
		return nil, fmt.Errorf("no asset data available for %s", symbol)
	}

	// Cache the result
	if g.assetCache != nil {
		cacheStart := time.Now()
		err := g.assetCache.Set(ctx, resp)
		metrics.RecordCacheOperation("set", "asset", err == nil, time.Since(cacheStart).Seconds())
	}

	metrics.RecordMarketDataRequest("total_asset_operation", "success", time.Since(start).Seconds())
	return resp, nil
}

// GetAssetName fetches just the name of the asset
func (g *MarketDataGateway) GetAssetName(ctx context.Context, symbol string) (string, error) {
	asset, err := g.GetAsset(ctx, symbol)
	if err != nil {
		return "", err
	}
	return asset.Name, nil
}

// GetPriceOnDate fetches the price of an asset on a specific date.
func (g *MarketDataGateway) GetPriceOnDate(ctx context.Context, symbol string, date time.Time) (float64, error) {
	// 1. Check Historical Cache
	if g.history != nil {
		if val, found, _ := g.history.GetPrice(ctx, symbol, date); found {
			return val, nil
		}
	}

	// Normalize date to UTC
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-1 * time.Nanosecond)

	// Attempt to find a price for the date, looking back up to 7 days if weekends/holidays
	const maxDaysToTry = 7
	currentDate := date.In(time.UTC)
	for i := 0; i < maxDaysToTry; i++ {
		// Try to verify if we have coverage for this date in recent fetch?
		// For now simple lookback loop
		req := &pb.GetHistoricalPricesRequest{
			Name:      fmt.Sprintf("assets/%s", symbol),
			StartTime: timestamppb.New(startOfDay),
			EndTime:   timestamppb.New(endOfDay),
			Interval:  "1d",
		}

		// ... RPC call ...
		resp, err := g.client.GetHistoricalPrices(ctx, req)
		if err == nil && len(resp.Prices) > 0 {
			price := resp.Prices[0].Price

			// Cache the result for the ORIGINAL requested date (and maybe the actual date found)
			if g.history != nil {
				// We cache it for the requested 'date' so we don't loop again.
				// NOTE: If we found data for T-2, we map T -> Price(T-2).
				// This approximates "nearest previous".
				// To be precise we might want to store specific date, but for "PriceOnDate" semantics
				// (value of holding ON that date) carrying forward last close is standard.
				_ = g.history.SetPrice(ctx, symbol, date, price)

				// Also cache for the actual date found if different?
				actualDate := resp.Prices[0].Timestamp.AsTime()
				if !actualDate.Equal(date) {
					_ = g.history.SetPrice(ctx, symbol, actualDate, price)
				}
			}
			return price, nil
		}

		// Move back one day
		currentDate = currentDate.AddDate(0, 0, -1)
		startOfDay = time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, time.UTC)
		endOfDay = startOfDay.Add(24 * time.Hour).Add(-1 * time.Nanosecond)
	}
	return 0, fmt.Errorf("no price found for %s after checking %d days, starting from %s", symbol, maxDaysToTry, date.Format("2006-01-02"))
}

// GetCurrencyRateOnDate fetches the conversion rate between two currencies on a specific date.
func (g *MarketDataGateway) GetCurrencyRateOnDate(ctx context.Context, baseCurrency, targetCurrency string, date time.Time) (float64, error) {
	if baseCurrency == targetCurrency {
		return 1.0, nil
	}

	// 1. Check Historical Cache
	if g.history != nil {
		if val, found, _ := g.history.GetCurrencyRate(ctx, baseCurrency, targetCurrency, date); found {
			return val, nil
		}
	}

	// Normalize date to UTC
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-1 * time.Nanosecond)

	const maxDaysToTry = 7
	currentDate := date.In(time.UTC)
	for i := 0; i < maxDaysToTry; i++ {
		req := &pb.GetHistoricalCurrencyRatesRequest{
			BaseCurrency:   baseCurrency,
			TargetCurrency: targetCurrency,
			StartTime:      timestamppb.New(startOfDay),
			EndTime:        timestamppb.New(endOfDay),
		}

		serviceStart := time.Now()
		resp, err := g.client.GetHistoricalCurrencyRates(ctx, req)
		duration := time.Since(serviceStart).Seconds()

		if err != nil {
			metrics.RecordMarketDataRequest("get_historical_currency_rates", "error", duration)
			// Continue loop
		} else {
			metrics.RecordMarketDataRequest("get_historical_currency_rates", "success", duration)
			if len(resp.Rates) > 0 {
				rate := resp.Rates[0].Rate
				// Cache result
				if g.history != nil {
					_ = g.history.SetCurrencyRate(ctx, baseCurrency, targetCurrency, date, rate)
				}
				return rate, nil
			}
		}

		// Move back one day
		currentDate = currentDate.AddDate(0, 0, -1)
		startOfDay = time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, time.UTC)
		endOfDay = startOfDay.Add(24 * time.Hour).Add(-1 * time.Nanosecond)
	}

	// If the loop completes, no rate was found within the maximum number of days checked (7 days total).
	return 0, fmt.Errorf("no currency rate found for %s/%s after checking %d days, starting from %s", baseCurrency, targetCurrency, maxDaysToTry, date.Format("2006-01-02"))
}

// GetHistoricalCurrencyRates fetches historical currency rates for a date range
func (g *MarketDataGateway) GetHistoricalCurrencyRates(ctx context.Context, baseCurrency, targetCurrency string, start, end time.Time) (map[time.Time]float64, error) {
	if baseCurrency == targetCurrency {
		return map[time.Time]float64{}, nil // No conversion needed
	}

	req := &pb.GetHistoricalCurrencyRatesRequest{
		BaseCurrency:   baseCurrency,
		TargetCurrency: targetCurrency,
		StartTime:      timestamppb.New(start),
		EndTime:        timestamppb.New(end),
	}

	serviceStart := time.Now()
	resp, err := g.client.GetHistoricalCurrencyRates(ctx, req)
	duration := time.Since(serviceStart).Seconds()

	if err != nil {
		metrics.RecordMarketDataRequest("get_historical_currency_rates", "error", duration)
		return nil, fmt.Errorf("failed to get historical currency rates: %w", err)
	}
	metrics.RecordMarketDataRequest("get_historical_currency_rates", "success", duration)

	rates := make(map[time.Time]float64)

	// Prepare for batch cache set
	// Using a goroutine for non-blocking cache write? Or just do it.
	// Given immutability and potential size, let's do it synchronously to ensure consistency or async for perf.
	// Async is better for latency of the main request.
	go func() {
		if g.history != nil && len(resp.Rates) > 0 {
			ctx := context.Background() // new context for async work
			for _, rate := range resp.Rates {
				rateDate := rate.RateDate.AsTime()
				_ = g.history.SetCurrencyRate(ctx, baseCurrency, targetCurrency, rateDate, rate.Rate)
			}
		}
	}()

	for _, rate := range resp.Rates {
		rateDate := rate.RateDate.AsTime()
		rates[rateDate] = rate.Rate
	}

	return rates, nil
}
