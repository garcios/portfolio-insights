package infrastructure

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/garcios/portfolio-insights/services/marketdata-service/proto/marketdata"
)

// CacheWarmer handles pre-populating the asset cache
type CacheWarmer struct {
	marketDataGateway *MarketDataGateway
	assetCache        *AssetCache
	logger            *slog.Logger
}

// NewCacheWarmer creates a new cache warmer instance
func NewCacheWarmer(marketDataGateway *MarketDataGateway, assetCache *AssetCache, logger *slog.Logger) *CacheWarmer {
	return &CacheWarmer{
		marketDataGateway: marketDataGateway,
		assetCache:        assetCache,
		logger:            logger,
	}
}

// WarmCache fetches all assets from marketdata service and caches them
func (cw *CacheWarmer) WarmCache(ctx context.Context) error {
	if cw.assetCache == nil {
		cw.logger.Warn("Asset cache is disabled, skipping cache warming")
		return nil
	}

	if cw.marketDataGateway == nil {
		cw.logger.Warn("MarketData gateway is not available, skipping cache warming")
		return nil
	}

	cw.logger.Info("Starting asset cache warming...")
	start := time.Now()

	// Fetch all assets from marketdata service
	assets, err := cw.fetchAllAssets(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch assets: %w", err)
	}

	if len(assets) == 0 {
		cw.logger.Warn("No assets found to cache")
		return nil
	}

	// Cache each asset
	successCount := 0
	failCount := 0

	for _, asset := range assets {
		if err := cw.assetCache.Set(ctx, asset); err != nil {
			cw.logger.Warn("failed to cache asset",
				"symbol", asset.Symbol,
				"error", err,
			)
			failCount++
		} else {
			successCount++
		}
	}

	duration := time.Since(start)
	cw.logger.Info("Asset cache warming completed",
		"total_assets", len(assets),
		"cached", successCount,
		"failed", failCount,
		"duration_ms", duration.Milliseconds(),
	)

	return nil
}

// fetchAllAssets retrieves all assets from the marketdata service using pagination
func (cw *CacheWarmer) fetchAllAssets(ctx context.Context) ([]*pb.Asset, error) {
	var allAssets []*pb.Asset
	pageToken := ""
	pageSize := int32(100) // Fetch 100 assets per page

	for {
		req := &pb.ListAssetsRequest{
			PageSize:  pageSize,
			PageToken: pageToken,
		}

		resp, err := cw.marketDataGateway.client.ListAssets(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to list assets: %w", err)
		}

		allAssets = append(allAssets, resp.Assets...)

		// Check if there are more pages
		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	return allAssets, nil
}

// WarmCacheAsync runs cache warming in the background
func (cw *CacheWarmer) WarmCacheAsync() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := cw.WarmCache(ctx); err != nil {
			cw.logger.Error("cache warming failed", "error", err)
		}
	}()
}

// SchedulePeriodicWarming runs cache warming on a schedule
func (cw *CacheWarmer) SchedulePeriodicWarming(interval time.Duration) {
	go func() {
		// Initial warming on startup
		cw.WarmCacheAsync()

		// Periodic warming
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			cw.logger.Info("Starting scheduled cache warming")
			cw.WarmCacheAsync()
		}
	}()
}
