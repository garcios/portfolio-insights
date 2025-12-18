package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/garcios/portfolio-insights/pkg/database"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/config"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/metrics"
)

type cachingPortfolioUsecase struct {
	delegate PortfolioUsecase
	cache    database.RedisClient
	cfg      config.CachingConfig
}

// NewCachingPortfolioUsecase creates a new caching decorator for PortfolioUsecase
func NewCachingPortfolioUsecase(delegate PortfolioUsecase, cache database.RedisClient, cfg config.CachingConfig) PortfolioUsecase {
	return &cachingPortfolioUsecase{
		delegate: delegate,
		cache:    cache,
		cfg:      cfg,
	}
}

func (uc *cachingPortfolioUsecase) GetHoldings(ctx context.Context, userID string) ([]*domain.Holding, error) {
	// Not caching holdings for now as they might change frequently with market data
	// and are often an intermediate step.
	return uc.delegate.GetHoldings(ctx, userID)
}

func (uc *cachingPortfolioUsecase) GetPortfolioSummary(ctx context.Context, userID string, startDate, endDate *time.Time) (*domain.PortfolioSummary, error) {
	if !uc.cfg.Enabled {
		return uc.delegate.GetPortfolioSummary(ctx, userID, startDate, endDate)
	}

	key := fmt.Sprintf("portfolio:summary:%s", userID)
	if startDate != nil {
		key += fmt.Sprintf(":start:%s", startDate.Format("2006-01-02"))
	}
	if endDate != nil {
		key += fmt.Sprintf(":end:%s", endDate.Format("2006-01-02"))
	}

	// Try to get from cache
	start := time.Now()
	val, err := uc.cache.Get(ctx, key)
	if err == nil && val != "" {
		var summary domain.PortfolioSummary
		if err := json.Unmarshal([]byte(val), &summary); err == nil {
			metrics.RecordCacheOperation("get_summary", "redis", true, time.Since(start).Seconds())
			return &summary, nil
		}
		// If unmarshal fails, log error and continue to fetch from delegate
		fmt.Printf("failed to unmarshal cached summary: %v\n", err)
	}
	metrics.RecordCacheOperation("get_summary", "redis", false, time.Since(start).Seconds())

	// Fetch from delegate
	summary, err := uc.delegate.GetPortfolioSummary(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Cache the result
	marshaled, err := json.Marshal(summary)
	if err == nil {
		ttl := time.Duration(uc.cfg.SummaryTTLSeconds) * time.Second
		if err := uc.cache.Set(ctx, key, string(marshaled), ttl); err != nil {
			fmt.Printf("failed to cache summary: %v\n", err)
		}
	}

	return summary, nil
}

func (uc *cachingPortfolioUsecase) GetHistoricalPortfolioSummary(ctx context.Context, userID string, date time.Time) (*domain.PortfolioSummary, error) {
	if !uc.cfg.Enabled {
		return uc.delegate.GetHistoricalPortfolioSummary(ctx, userID, date)
	}

	// Key includes date. Format date as YYYY-MM-DD
	dateStr := date.Format("2006-01-02")
	key := fmt.Sprintf("portfolio:history:%s:%s", userID, dateStr)

	// Try to get from cache
	start := time.Now()
	val, err := uc.cache.Get(ctx, key)
	if err == nil && val != "" {
		var summary domain.PortfolioSummary
		if err := json.Unmarshal([]byte(val), &summary); err == nil {
			metrics.RecordCacheOperation("get_history", "redis", true, time.Since(start).Seconds())
			return &summary, nil
		}
		fmt.Printf("failed to unmarshal cached history: %v\n", err)
	}
	metrics.RecordCacheOperation("get_history", "redis", false, time.Since(start).Seconds())

	// Fetch from delegate
	summary, err := uc.delegate.GetHistoricalPortfolioSummary(ctx, userID, date)
	if err != nil {
		return nil, err
	}

	// Cache the result
	marshaled, err := json.Marshal(summary)
	if err == nil {
		ttl := time.Duration(uc.cfg.HistoryTTLSeconds) * time.Second
		if err := uc.cache.Set(ctx, key, string(marshaled), ttl); err != nil {
			fmt.Printf("failed to cache history: %v\n", err)
		}
	}

	return summary, nil
}

func (uc *cachingPortfolioUsecase) BackfillPortfolioHistory(ctx context.Context, userIDs []string, startDate, endDate time.Time, dryRun bool) BackfillResult {
	// Backfilling is a write/heavy operation, usually one-off. No read caching strategy needed here.
	// We pass through to delegate.
	return uc.delegate.BackfillPortfolioHistory(ctx, userIDs, startDate, endDate, dryRun)
}

func (uc *cachingPortfolioUsecase) RefreshSnapshot(ctx context.Context, userID string) error {
	// Pass through to delegate - this is a write operation
	return uc.delegate.RefreshSnapshot(ctx, userID)
}
