// Package infrastructure provides external service implementations and configuration.
package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/garcios/portfolio-insights/services/marketdata-service/marketdata"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/config"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/metrics"
	"github.com/redis/go-redis/v9"
)

// AssetCache handles caching of asset metadata
type AssetCache struct {
	client *redis.Client
	ttl    time.Duration
}

// CachedAsset represents a cached asset entry
type CachedAsset struct {
	Symbol   string    `json:"symbol"`
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	Exchange string    `json:"exchange"`
	Currency string    `json:"currency"`
	CachedAt time.Time `json:"cached_at"`
}

// NewAssetCache creates a new asset cache.
func NewAssetCache(client *redis.Client, cfg config.Config) *AssetCache {
	// Default TTL: 24 hours (assets don't change frequently)
	ttl := 24 * time.Hour

	// Allow override via environment
	if cfg.AssetCacheTTL > 0 {
		ttl = time.Duration(cfg.AssetCacheTTL) * time.Second
	}

	return &AssetCache{
		client: client,
		ttl:    ttl,
	}
}

// Get retrieves a cached asset for a symbol
func (ac *AssetCache) Get(ctx context.Context, symbol string) (*CachedAsset, error) {
	start := time.Now()
	key := ac.assetKey(symbol)

	data, err := ac.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		// Cache miss
		metrics.RecordCacheOperation("get", "asset", false, time.Since(start).Seconds())
		return nil, nil
	}
	if err != nil {
		metrics.RecordCacheOperation("get", "asset", false, time.Since(start).Seconds())
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var cached CachedAsset
	if err := json.Unmarshal(data, &cached); err != nil {
		metrics.RecordCacheOperation("get", "asset", false, time.Since(start).Seconds())
		return nil, fmt.Errorf("failed to unmarshal cached asset: %w", err)
	}

	// Cache hit
	metrics.RecordCacheOperation("get", "asset", true, time.Since(start).Seconds())
	return &cached, nil
}

// Set stores an asset in the cache
func (ac *AssetCache) Set(ctx context.Context, asset *pb.Asset) error {
	start := time.Now()
	key := ac.assetKey(asset.Symbol)

	cached := &CachedAsset{
		Symbol:   asset.Symbol,
		Name:     asset.Name,
		Type:     asset.Type,
		Exchange: asset.Exchange,
		Currency: asset.Currency,
		CachedAt: time.Now(),
	}

	data, err := json.Marshal(cached)
	if err != nil {
		metrics.RecordCacheOperation("set", "asset", false, time.Since(start).Seconds())
		return fmt.Errorf("failed to marshal asset: %w", err)
	}

	if err := ac.client.Set(ctx, key, data, ac.ttl).Err(); err != nil {
		metrics.RecordCacheOperation("set", "asset", false, time.Since(start).Seconds())
		return fmt.Errorf("failed to set cache: %w", err)
	}

	metrics.RecordCacheOperation("set", "asset", true, time.Since(start).Seconds())
	return nil
}

// Delete removes an asset from the cache
func (ac *AssetCache) Delete(ctx context.Context, symbol string) error {
	start := time.Now()
	key := ac.assetKey(symbol)

	if err := ac.client.Del(ctx, key).Err(); err != nil {
		metrics.RecordCacheOperation("delete", "asset", false, time.Since(start).Seconds())
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	metrics.RecordCacheOperation("delete", "asset", true, time.Since(start).Seconds())
	return nil
}

// assetKey generates the Redis key for an asset
func (ac *AssetCache) assetKey(symbol string) string {
	return fmt.Sprintf("asset:%s", symbol)
}
