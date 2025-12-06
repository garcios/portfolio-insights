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

// CurrencySyncConfig holds configuration for the currency sync worker
type CurrencySyncConfig struct {
	SyncInterval   time.Duration
	StaleDuration  time.Duration
	BatchSize      int
	MaxConcurrency int
	HistoricalDays int
	RateLimit      float64
}

// DefaultCurrencySyncConfig returns default configuration
func DefaultCurrencySyncConfig() CurrencySyncConfig {
	return CurrencySyncConfig{
		SyncInterval:   1 * time.Hour,
		StaleDuration:  24 * time.Hour,
		BatchSize:      10,
		MaxConcurrency: 2,
		HistoricalDays: 30,
		RateLimit:      1.0,
	}
}

// EODHDCurrencySyncWorker handles periodic synchronization of currency rates from EODHD API
type EODHDCurrencySyncWorker struct {
	repo   domain.MarketDataRepository
	client client.EODHDClient
	config CurrencySyncConfig
}

// NewEODHDCurrencySyncWorker creates a new currency sync worker
func NewEODHDCurrencySyncWorker(repo domain.MarketDataRepository, cfg config.Config) (*EODHDCurrencySyncWorker, error) {
	if cfg.EodhdApiToken == "" {
		return nil, fmt.Errorf("EODHD_API_TOKEN configuration is required")
	}

	workerConfig := CurrencySyncConfig{
		SyncInterval:   cfg.CurrencySyncInterval,
		StaleDuration:  cfg.CurrencyStaleDuration,
		BatchSize:      cfg.CurrencySyncBatchSize,
		MaxConcurrency: cfg.CurrencySyncMaxConcurrency,
		HistoricalDays: cfg.CurrencySyncHistoricalDays,
		RateLimit:      1.0,
	}

	// Reuse EODHD client
	eodhd := client.NewEODHDClient(cfg.EodhdApiBaseUrl, cfg.EodhdApiToken, workerConfig.RateLimit)

	return &EODHDCurrencySyncWorker{
		repo:   repo,
		client: eodhd,
		config: workerConfig,
	}, nil
}

// Start begins the periodic currency synchronization
func (w *EODHDCurrencySyncWorker) Start(ctx context.Context) {
	go func() {
		log.Println("EODHD Currency Sync Worker: Starting...")

		// Run immediately on startup
		if err := w.syncCurrencies(ctx); err != nil {
			log.Printf("EODHD Currency Sync Worker: Initial sync failed: %v", err)
		}

		ticker := time.NewTicker(w.config.SyncInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := w.syncCurrencies(ctx); err != nil {
					log.Printf("EODHD Currency Sync Worker: Sync failed: %v", err)
				}
			case <-ctx.Done():
				log.Println("EODHD Currency Sync Worker: Shutting down...")
				return
			}
		}
	}()
}

// TriggerSync manually triggers the currency synchronization
func (w *EODHDCurrencySyncWorker) TriggerSync(ctx context.Context) error {
	return w.syncCurrencies(ctx)
}

// syncCurrencies performs the main synchronization logic
func (w *EODHDCurrencySyncWorker) syncCurrencies(ctx context.Context) error {
	start := time.Now()
	status := "success"
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.RecordIngestionJob("eodhd_currency_sync", status, duration)
	}()

	log.Println("EODHD Currency Sync Worker: Starting currency synchronization...")

	// Get target currencies
	currencies, err := w.repo.GetTargetCurrencies()
	if err != nil {
		status = "failure"
		return fmt.Errorf("failed to get target currencies: %w", err)
	}

	if len(currencies) == 0 {
		log.Println("EODHD Currency Sync Worker: No currencies found to sync")
		return nil
	}

	log.Printf("EODHD Currency Sync Worker: Found %d currencies to sync", len(currencies))

	// Process in batches
	totalSynced := 0
	totalErrors := 0

	for i := 0; i < len(currencies); i += w.config.BatchSize {
		end := i + w.config.BatchSize
		if end > len(currencies) {
			end = len(currencies)
		}
		batch := currencies[i:end]

		synced, errors := w.processBatch(ctx, batch)
		totalSynced += synced
		totalErrors += errors
	}

	log.Printf("EODHD Currency Sync Worker: Completed. Synced %d rates with %d errors", totalSynced, totalErrors)

	if totalErrors > 0 {
		status = "partial_failure"
	}

	return nil
}

// processBatch processes a batch of currencies with concurrency control
func (w *EODHDCurrencySyncWorker) processBatch(ctx context.Context, currencies []string) (int, int) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, w.config.MaxConcurrency)
	results := make(chan syncResult, len(currencies))

	for _, currency := range currencies {
		wg.Add(1)
		go func(curr string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := w.syncCurrency(ctx, curr)
			results <- result
		}(currency)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	totalSynced := 0
	totalErrors := 0

	for result := range results {
		totalSynced += result.PricesSynced
		if result.Error != nil {
			totalErrors++
		}
	}

	return totalSynced, totalErrors
}

// syncResult holds the result of syncing a single item
type syncResult struct {
	Symbol       string
	PricesSynced int
	Error        error
}

// syncCurrency synchronizes rates for a single currency
func (w *EODHDCurrencySyncWorker) syncCurrency(ctx context.Context, targetCurrency string) syncResult {
	result := syncResult{
		Symbol: targetCurrency,
	}

	// Assume base currency is USD
	baseCurrency := "USD"

	// Determine date range
	fromDate, toDate := w.determineDateRange(baseCurrency, targetCurrency)

	log.Printf("EODHD Currency Sync Worker: Syncing %s from %s to %s",
		targetCurrency, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

	// Fetch historical prices (rates)
	// EODHD format: {CURRENCY}.FOREX
	ticker := targetCurrency
	exchange := "FOREX"

	historicalPrices, err := w.client.GetHistoricalPrices(ctx, ticker, exchange, fromDate, toDate)
	if err != nil {
		result.Error = fmt.Errorf("failed to fetch rates for %s: %w", targetCurrency, err)
		log.Printf("EODHD Currency Sync Worker: %v", result.Error)
		metrics.RecordIngestionJob("eodhd_currency_api_call", "failure", 0)
		return result
	}

	metrics.RecordIngestionJob("eodhd_currency_api_call", "success", 0)

	if len(historicalPrices) == 0 {
		log.Printf("EODHD Currency Sync Worker: No rates returned for %s", targetCurrency)
		return result
	}

	// Convert to domain.CurrencyRate
	rates := make([]*domain.CurrencyRate, 0, len(historicalPrices))
	for _, hp := range historicalPrices {
		timestamp, err := time.Parse("2006-01-02", hp.Date)
		if err != nil {
			log.Printf("EODHD Currency Sync Worker: Failed to parse date %s for %s: %v", hp.Date, targetCurrency, err)
			continue
		}

		rates = append(rates, &domain.CurrencyRate{
			BaseCurrency:   baseCurrency,
			TargetCurrency: targetCurrency,
			Rate:           hp.Close, // Using Close as the rate
			RateDate:       timestamp,
		})
	}

	// Insert rates into database
	if len(rates) > 0 {
		if err := w.repo.InsertCurrencyRates(rates); err != nil {
			result.Error = fmt.Errorf("failed to insert rates for %s: %w", targetCurrency, err)
			log.Printf("EODHD Currency Sync Worker: %v", result.Error)
			return result
		}

		result.PricesSynced = len(rates)
		metrics.RecordCurrenciesIngested(len(rates))
		log.Printf("EODHD Currency Sync Worker: Successfully synced %d rates for %s", len(rates), targetCurrency)
	}

	return result
}

// determineDateRange determines the date range to fetch for a currency pair
func (w *EODHDCurrencySyncWorker) determineDateRange(base, target string) (time.Time, time.Time) {
	toDate := time.Now()

	// Get latest rate timestamp
	latestTimestamp, err := w.repo.GetLatestCurrencyRateTimestamp(base, target)
	if err != nil || latestTimestamp == nil {
		// No existing rates - fetch historical data
		fromDate := toDate.AddDate(0, 0, -w.config.HistoricalDays)
		return fromDate, toDate
	}

	// Fetch from day after latest rate to today
	fromDate := latestTimestamp.AddDate(0, 0, 1)

	// If fromDate is in the future, no need to fetch
	if fromDate.After(toDate) {
		return toDate, toDate
	}

	return fromDate, toDate
}

// SyncSpecificCurrencies synchronizes rates for specific currencies (used by HTTP endpoint)
func (w *EODHDCurrencySyncWorker) SyncSpecificCurrencies(ctx context.Context, currencies []string, fromDate, toDate time.Time) (*SyncJobResult, error) {
	result := &SyncJobResult{
		StartedAt: time.Now(),
		Status:    "running",
		Details:   make([]AssetSyncDetail, 0, len(currencies)),
	}

	for _, currency := range currencies {
		// Assume base USD
		baseCurrency := "USD"

		// Fetch rates
		historicalPrices, err := w.client.GetHistoricalPrices(ctx, currency, "FOREX", fromDate, toDate)
		if err != nil {
			detail := AssetSyncDetail{
				Symbol: currency,
				Status: "error",
				Error:  fmt.Sprintf("Failed to fetch rates: %v", err),
			}
			result.Details = append(result.Details, detail)
			result.Errors++
			continue
		}

		// Convert and insert
		rates := make([]*domain.CurrencyRate, 0, len(historicalPrices))
		for _, hp := range historicalPrices {
			timestamp, err := time.Parse("2006-01-02", hp.Date)
			if err != nil {
				continue
			}

			rates = append(rates, &domain.CurrencyRate{
				BaseCurrency:   baseCurrency,
				TargetCurrency: currency,
				Rate:           hp.Close,
				RateDate:       timestamp,
			})
		}

		if len(rates) > 0 {
			if err := w.repo.InsertCurrencyRates(rates); err != nil {
				detail := AssetSyncDetail{
					Symbol: currency,
					Status: "error",
					Error:  fmt.Sprintf("Failed to insert rates: %v", err),
				}
				result.Details = append(result.Details, detail)
				result.Errors++
				continue
			}
		}

		detail := AssetSyncDetail{
			Symbol:      currency,
			PricesAdded: len(rates),
			Status:      "success",
		}
		result.Details = append(result.Details, detail)
		result.AssetsProcessed++
		result.PricesSynced += len(rates)
	}

	completedAt := time.Now()
	result.CompletedAt = &completedAt
	result.Status = "completed"

	return result, nil
}
