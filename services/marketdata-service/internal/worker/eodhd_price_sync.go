package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/client"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/config"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/domain"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/metrics"
)

// PriceSyncConfig holds configuration for the price sync worker
type PriceSyncConfig struct {
	SyncInterval   time.Duration
	StaleDuration  time.Duration
	BatchSize      int
	MaxConcurrency int
	HistoricalDays int
	RateLimit      float64
}

// DefaultPriceSyncConfig returns default configuration
func DefaultPriceSyncConfig() PriceSyncConfig {
	return PriceSyncConfig{
		SyncInterval:   1 * time.Hour,
		StaleDuration:  24 * time.Hour,
		BatchSize:      100,
		MaxConcurrency: 5,
		HistoricalDays: 30,
		RateLimit:      1.0, // 1 request per second
	}
}

// EODHDPriceSyncWorker handles periodic synchronization of prices from EODHD API
type EODHDPriceSyncWorker struct {
	repo   domain.MarketDataRepository
	client client.EODHDClient
	config PriceSyncConfig
}

// NewEODHDPriceSyncWorker creates a new price sync worker
func NewEODHDPriceSyncWorker(repo domain.MarketDataRepository, cfg config.Config) (*EODHDPriceSyncWorker, error) {
	if cfg.EodhdApiToken == "" {
		return nil, fmt.Errorf("EODHD_API_TOKEN configuration is required")
	}

	workerConfig := PriceSyncConfig{
		SyncInterval:   cfg.PriceSyncInterval,
		StaleDuration:  cfg.PriceStaleDuration,
		BatchSize:      cfg.PriceSyncBatchSize,
		MaxConcurrency: cfg.PriceSyncMaxConcurrency,
		HistoricalDays: cfg.PriceSyncHistoricalDays,
		RateLimit:      cfg.EodhdRateLimit,
	}

	// Apply defaults if zero values (though Viper should handle defaults in Config, explicit safety is good)
	defaults := DefaultPriceSyncConfig()
	if workerConfig.SyncInterval == 0 {
		workerConfig.SyncInterval = defaults.SyncInterval
	}
	if workerConfig.StaleDuration == 0 {
		workerConfig.StaleDuration = defaults.StaleDuration
	}
	if workerConfig.BatchSize == 0 {
		workerConfig.BatchSize = defaults.BatchSize
	}
	if workerConfig.MaxConcurrency == 0 {
		workerConfig.MaxConcurrency = defaults.MaxConcurrency
	}
	if workerConfig.HistoricalDays == 0 {
		workerConfig.HistoricalDays = defaults.HistoricalDays
	}
	if workerConfig.RateLimit == 0 {
		workerConfig.RateLimit = defaults.RateLimit
	}

	eodhd := client.NewEODHDClient(cfg.EodhdApiBaseUrl, cfg.EodhdApiToken, workerConfig.RateLimit)

	return &EODHDPriceSyncWorker{
		repo:   repo,
		client: eodhd,
		config: workerConfig,
	}, nil
}

// Start begins the periodic price synchronization
func (w *EODHDPriceSyncWorker) Start(ctx context.Context) {
	go func() {
		log.Println("EODHD Price Sync Worker: Starting...")

		// Run immediately on startup
		if err := w.syncPrices(ctx); err != nil {
			log.Printf("EODHD Price Sync Worker: Initial sync failed: %v", err)
		}

		ticker := time.NewTicker(w.config.SyncInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := w.syncPrices(ctx); err != nil {
					log.Printf("EODHD Price Sync Worker: Sync failed: %v", err)
				}
			case <-ctx.Done():
				log.Println("EODHD Price Sync Worker: Shutting down...")
				return
			}
		}
	}()
}

// syncPrices performs the main synchronization logic
func (w *EODHDPriceSyncWorker) syncPrices(ctx context.Context) error {
	start := time.Now()
	status := "success"
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.RecordIngestionJob("eodhd_price_sync", status, duration)
	}()

	log.Println("EODHD Price Sync Worker: Starting price synchronization...")

	// Get assets requiring price updates
	assets, err := w.repo.GetAssetsRequiringPriceUpdate(w.config.StaleDuration)
	if err != nil {
		status = "failure"
		return fmt.Errorf("failed to get assets requiring update: %w", err)
	}

	if len(assets) == 0 {
		log.Println("EODHD Price Sync Worker: No assets require price updates")
		return nil
	}

	log.Printf("EODHD Price Sync Worker: Found %d assets requiring updates", len(assets))

	// Process assets in batches with concurrency control
	totalPricesSynced := 0
	totalErrors := 0

	for i := 0; i < len(assets); i += w.config.BatchSize {
		end := i + w.config.BatchSize
		if end > len(assets) {
			end = len(assets)
		}
		batch := assets[i:end]

		pricesSynced, errors := w.processBatch(ctx, batch)
		totalPricesSynced += pricesSynced
		totalErrors += errors
	}

	log.Printf("EODHD Price Sync Worker: Completed. Synced %d prices with %d errors", totalPricesSynced, totalErrors)

	if totalErrors > 0 {
		status = "partial_failure"
	}

	return nil
}

// processBatch processes a batch of assets with concurrency control
func (w *EODHDPriceSyncWorker) processBatch(ctx context.Context, assets []*domain.Asset) (int, int) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, w.config.MaxConcurrency)

	results := make(chan priceSyncResult, len(assets))

	for _, asset := range assets {
		wg.Add(1)
		go func(a *domain.Asset) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := w.syncAsset(ctx, a)
			results <- result
		}(asset)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	totalPricesSynced := 0
	totalErrors := 0

	for result := range results {
		totalPricesSynced += result.PricesSynced
		if result.Error != nil {
			totalErrors++
		}
	}

	return totalPricesSynced, totalErrors
}

// priceSyncResult holds the result of syncing a single asset
type priceSyncResult struct {
	Symbol       string
	PricesSynced int
	Error        error
}

// syncAsset synchronizes prices for a single asset
func (w *EODHDPriceSyncWorker) syncAsset(ctx context.Context, asset *domain.Asset) priceSyncResult {
	result := priceSyncResult{
		Symbol: asset.Symbol,
	}

	// Determine date range to fetch
	fromDate, toDate := w.determineDateRange(asset)

	log.Printf("EODHD Price Sync Worker: Syncing %s from %s to %s",
		asset.Symbol, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

	// map exchange to EODHD exchange
	exchange := asset.Exchange
	switch exchange {
	case "ASX":
		exchange = "AU"
	case "NASDAQ":
		exchange = "US"
	case "NYSE":
		exchange = "US"
	default:
		exchange = "US"
	}

	// Fetch historical prices from EODHD
	historicalPrices, err := w.client.GetHistoricalPrices(ctx, asset.Symbol, exchange, fromDate, toDate)
	if err != nil {
		result.Error = fmt.Errorf("failed to fetch prices for %s: %w", asset.Symbol, err)
		log.Printf("EODHD Price Sync Worker: %v", result.Error)
		metrics.RecordIngestionJob("eodhd_api_call", "failure", 0)
		return result
	}

	metrics.RecordIngestionJob("eodhd_api_call", "success", 0)

	if len(historicalPrices) == 0 {
		log.Printf("EODHD Price Sync Worker: No prices returned for %s", asset.Symbol)
		return result
	}

	// Convert to domain.AssetPrice
	prices := make([]*domain.AssetPrice, 0, len(historicalPrices))
	for _, hp := range historicalPrices {
		timestamp, err := time.Parse("2006-01-02", hp.Date)
		if err != nil {
			log.Printf("EODHD Price Sync Worker: Failed to parse date %s for %s: %v", hp.Date, asset.Symbol, err)
			continue
		}

		prices = append(prices, &domain.AssetPrice{
			AssetID:   asset.ID,
			Price:     hp.AdjustedClose, // Use adjusted close for accuracy
			Timestamp: timestamp,
		})
	}

	// Insert prices into database
	if len(prices) > 0 {
		if err := w.repo.InsertPrices(prices); err != nil {
			result.Error = fmt.Errorf("failed to insert prices for %s: %w", asset.Symbol, err)
			log.Printf("EODHD Price Sync Worker: %v", result.Error)
			return result
		}

		result.PricesSynced = len(prices)
		metrics.RecordPricesIngested(len(prices))
		log.Printf("EODHD Price Sync Worker: Successfully synced %d prices for %s", len(prices), asset.Symbol)
	}

	return result
}

// determineDateRange determines the date range to fetch for an asset
func (w *EODHDPriceSyncWorker) determineDateRange(asset *domain.Asset) (time.Time, time.Time) {
	toDate := time.Now()

	// Get latest price timestamp
	latestTimestamp, err := w.repo.GetLatestPriceTimestamp(asset.ID)
	if err != nil || latestTimestamp == nil {
		// No existing prices - fetch historical data
		fromDate := toDate.AddDate(0, 0, -w.config.HistoricalDays)
		return fromDate, toDate
	}

	// Fetch from day after latest price to today
	fromDate := latestTimestamp.AddDate(0, 0, 1)

	// If fromDate is in the future, no need to fetch
	if fromDate.After(toDate) {
		return toDate, toDate
	}

	return fromDate, toDate
}

// SyncSpecificAssets synchronizes prices for specific assets (used by HTTP endpoint)
func (w *EODHDPriceSyncWorker) SyncSpecificAssets(ctx context.Context, symbols []string, exchange string, fromDate, toDate time.Time) (*SyncJobResult, error) {
	result := &SyncJobResult{
		StartedAt: time.Now(),
		Status:    "running",
		Details:   make([]AssetSyncDetail, 0, len(symbols)),
	}

	for _, symbol := range symbols {
		// Get asset from database
		asset, err := w.repo.GetAssetBySymbol(symbol)
		if err != nil {
			detail := AssetSyncDetail{
				Symbol: symbol,
				Status: "error",
				Error:  fmt.Sprintf("Asset not found: %v", err),
			}
			result.Details = append(result.Details, detail)
			result.Errors++
			continue
		}

		// Fetch prices
		historicalPrices, err := w.client.GetHistoricalPrices(ctx, asset.Symbol, asset.Exchange, fromDate, toDate)
		if err != nil {
			detail := AssetSyncDetail{
				Symbol: symbol,
				Status: "error",
				Error:  fmt.Sprintf("Failed to fetch prices: %v", err),
			}
			result.Details = append(result.Details, detail)
			result.Errors++
			continue
		}

		// Convert and insert
		prices := make([]*domain.AssetPrice, 0, len(historicalPrices))
		for _, hp := range historicalPrices {
			timestamp, err := time.Parse("2006-01-02", hp.Date)
			if err != nil {
				continue
			}

			prices = append(prices, &domain.AssetPrice{
				AssetID:   asset.ID,
				Price:     hp.AdjustedClose,
				Timestamp: timestamp,
			})
		}

		if len(prices) > 0 {
			if err := w.repo.InsertPrices(prices); err != nil {
				detail := AssetSyncDetail{
					Symbol: symbol,
					Status: "error",
					Error:  fmt.Sprintf("Failed to insert prices: %v", err),
				}
				result.Details = append(result.Details, detail)
				result.Errors++
				continue
			}
		}

		detail := AssetSyncDetail{
			Symbol:      symbol,
			PricesAdded: len(prices),
			Status:      "success",
		}
		result.Details = append(result.Details, detail)
		result.AssetsProcessed++
		result.PricesSynced += len(prices)
	}

	completedAt := time.Now()
	result.CompletedAt = &completedAt
	result.Status = "completed"

	return result, nil
}

// SyncJobResult holds the result of a sync job
type SyncJobResult struct {
	ID              string
	Status          string
	StartedAt       time.Time
	CompletedAt     *time.Time
	AssetsProcessed int
	PricesSynced    int
	Errors          int
	Details         []AssetSyncDetail
}

// AssetSyncDetail holds details for a single asset sync
type AssetSyncDetail struct {
	Symbol      string
	PricesAdded int
	Status      string
	Error       string
}
