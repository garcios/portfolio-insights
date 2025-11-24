package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/metrics"
	"github.com/redis/go-redis/v9"
)

// PriceCache handles caching of asset prices
type PriceCache struct {
	client *redis.Client
	ttl    time.Duration
}

// CachedPrice represents a cached price entry
type CachedPrice struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
	CachedAt  time.Time `json:"cached_at"`
}

func NewPriceCache(client *redis.Client) *PriceCache {
	// Default TTL: 60 seconds
	ttl := 60 * time.Second

	// Allow override via environment
	if ttlStr := os.Getenv("PRICE_CACHE_TTL_SECONDS"); ttlStr != "" {
		if seconds, err := time.ParseDuration(ttlStr + "s"); err == nil {
			ttl = seconds
		}
	}

	return &PriceCache{
		client: client,
		ttl:    ttl,
	}
}

// Get retrieves a cached price for a symbol
func (pc *PriceCache) Get(ctx context.Context, symbol string) (*CachedPrice, error) {
	start := time.Now()
	key := pc.priceKey(symbol)

	data, err := pc.client.Get(ctx, key).Bytes()
	duration := time.Since(start).Seconds()

	if err == redis.Nil {
		metrics.RecordCacheOperation("get", "redis", false, duration)
		return nil, nil // Cache miss
	}
	if err != nil {
		metrics.RecordCacheOperation("get", "redis", false, duration)
		return nil, fmt.Errorf("failed to get cached price: %w", err)
	}

	var cached CachedPrice
	if err := json.Unmarshal(data, &cached); err != nil {
		metrics.RecordCacheOperation("get", "redis", false, duration) // Treat unmarshal error as miss/fail
		return nil, fmt.Errorf("failed to unmarshal cached price: %w", err)
	}

	metrics.RecordCacheOperation("get", "redis", true, duration)
	return &cached, nil
}

// GetMultiple retrieves cached prices for multiple symbols
func (pc *PriceCache) GetMultiple(ctx context.Context, symbols []string) (map[string]*CachedPrice, error) {
	start := time.Now()
	if len(symbols) == 0 {
		return make(map[string]*CachedPrice), nil
	}

	// Build keys
	keys := make([]string, len(symbols))
	for i, symbol := range symbols {
		keys[i] = pc.priceKey(symbol)
	}

	// Use pipeline for efficiency
	pipe := pc.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		// Ignore redis.Nil as it just means some keys don't exist
	}

	// Parse results
	result := make(map[string]*CachedPrice)
	hits := 0
	misses := 0

	for i, cmd := range cmds {
		data, err := cmd.Bytes()
		if err == redis.Nil {
			misses++
			continue // Cache miss for this symbol
		}
		if err != nil {
			misses++
			continue // Skip errors for individual keys
		}

		var cached CachedPrice
		if err := json.Unmarshal(data, &cached); err != nil {
			misses++
			continue // Skip unmarshaling errors
		}

		result[symbols[i]] = &cached
		hits++
	}

	duration := time.Since(start).Seconds()
	metrics.CacheOperationDuration.WithLabelValues("get_multiple", "redis").Observe(duration)
	metrics.CacheHitsTotal.WithLabelValues("redis").Add(float64(hits))
	metrics.CacheMissesTotal.WithLabelValues("redis").Add(float64(misses))

	return result, nil
}

// Set caches a price for a symbol
func (pc *PriceCache) Set(ctx context.Context, symbol string, price float64, timestamp time.Time) error {
	start := time.Now()
	key := pc.priceKey(symbol)

	cached := CachedPrice{
		Symbol:    symbol,
		Price:     price,
		Timestamp: timestamp,
		CachedAt:  time.Now(),
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return fmt.Errorf("failed to marshal price: %w", err)
	}

	if err := pc.client.Set(ctx, key, data, pc.ttl).Err(); err != nil {
		return fmt.Errorf("failed to cache price: %w", err)
	}

	metrics.CacheOperationDuration.WithLabelValues("set", "redis").Observe(time.Since(start).Seconds())
	return nil
}

// SetMultiple caches multiple prices
func (pc *PriceCache) SetMultiple(ctx context.Context, prices map[string]float64, timestamp time.Time) error {
	start := time.Now()
	if len(prices) == 0 {
		return nil
	}

	pipe := pc.client.Pipeline()

	for symbol, price := range prices {
		key := pc.priceKey(symbol)
		cached := CachedPrice{
			Symbol:    symbol,
			Price:     price,
			Timestamp: timestamp,
			CachedAt:  time.Now(),
		}

		data, err := json.Marshal(cached)
		if err != nil {
			continue // Skip errors for individual prices
		}

		pipe.Set(ctx, key, data, pc.ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to cache prices: %w", err)
	}

	metrics.CacheOperationDuration.WithLabelValues("set_multiple", "redis").Observe(time.Since(start).Seconds())
	return nil
}

// Delete removes a cached price
func (pc *PriceCache) Delete(ctx context.Context, symbol string) error {
	key := pc.priceKey(symbol)
	return pc.client.Del(ctx, key).Err()
}

// Clear removes all cached prices
func (pc *PriceCache) Clear(ctx context.Context) error {
	pattern := pc.priceKey("*")

	iter := pc.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := pc.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}

// priceKey generates a Redis key for a symbol
func (pc *PriceCache) priceKey(symbol string) string {
	return fmt.Sprintf("price:%s", symbol)
}

// GetStats returns cache statistics
func (pc *PriceCache) GetStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := pc.client.Info(ctx, "stats").Result()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"ttl_seconds": pc.ttl.Seconds(),
		"redis_info":  info,
	}, nil
}
