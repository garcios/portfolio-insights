package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/metrics"
	"github.com/redis/go-redis/v9"
)

// HistoricalCache handles caching of immutable historical data
type HistoricalCache struct {
	client *redis.Client
	ttl    time.Duration // Should be very long for historical data
}

// NewHistoricalCache creates a new historical data cache
func NewHistoricalCache(client *redis.Client) *HistoricalCache {
	return &HistoricalCache{
		client: client,
		ttl:    24 * 365 * time.Hour, // 1 Year default (effectively permanent)
	}
}

// GetPrice retrieves a historical price from cache
func (c *HistoricalCache) GetPrice(ctx context.Context, symbol string, date time.Time) (float64, bool, error) {
	key := c.priceKey(symbol, date)
	val, err := c.client.Get(ctx, key).Float64()
	if err == redis.Nil {
		metrics.RecordCacheOperation("get", "historical_price", false, 0)
		return 0, false, nil
	}
	if err != nil {
		metrics.RecordCacheOperation("get", "historical_price", false, 0)
		return 0, false, err
	}
	metrics.RecordCacheOperation("get", "historical_price", true, 0)
	return val, true, nil
}

// SetPrice caches a historical price
func (c *HistoricalCache) SetPrice(ctx context.Context, symbol string, date time.Time, price float64) error {
	key := c.priceKey(symbol, date)
	err := c.client.Set(ctx, key, price, c.ttl).Err()
	metrics.RecordCacheOperation("set", "historical_price", err == nil, 0)
	return err
}

// GetCurrencyRate retrieves a historical currency rate from cache
func (c *HistoricalCache) GetCurrencyRate(ctx context.Context, base, target string, date time.Time) (float64, bool, error) {
	key := c.rateKey(base, target, date)
	val, err := c.client.Get(ctx, key).Float64()
	if err == redis.Nil {
		metrics.RecordCacheOperation("get", "historical_rate", false, 0)
		return 0, false, nil
	}
	if err != nil {
		metrics.RecordCacheOperation("get", "historical_rate", false, 0)
		return 0, false, err
	}
	metrics.RecordCacheOperation("get", "historical_rate", true, 0)
	return val, true, nil
}

// SetCurrencyRate caches a historical currency rate
func (c *HistoricalCache) SetCurrencyRate(ctx context.Context, base, target string, date time.Time, rate float64) error {
	key := c.rateKey(base, target, date)
	err := c.client.Set(ctx, key, rate, c.ttl).Err()
	metrics.RecordCacheOperation("set", "historical_rate", err == nil, 0)
	return err
}

// BatchGetCurrencyRates attempts to fetch multiple rates. Returns map of found items.
func (c *HistoricalCache) BatchGetCurrencyRates(ctx context.Context, base, target string, dates []time.Time) (map[time.Time]float64, error) {
	if len(dates) == 0 {
		return nil, nil
	}

	keys := make([]string, len(dates))
	for i, d := range dates {
		keys[i] = c.rateKey(base, target, d)
	}

	// MGET for efficiency
	start := time.Now()
	vals, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	metrics.RecordCacheOperation("mget", "historical_rate", true, time.Since(start).Seconds())

	result := make(map[time.Time]float64)
	for i, v := range vals {
		if v != nil {
			// Redis driver returns interface{}, need to cast
			if s, ok := v.(string); ok {
				var f float64
				// Using json unmarshal or simple ParseFloat?
				// MGet returns string representation.
				// But we stored as float/string.
				// Let's rely on standard parsing or just casting if it was set via Set
				// Redis Set stores string.
				if _, err := fmt.Sscanf(s, "%f", &f); err == nil {
					result[dates[i]] = f
				}
			}
		}
	}
	return result, nil
}

func (c *HistoricalCache) priceKey(symbol string, date time.Time) string {
	return fmt.Sprintf("hist:price:%s:%s", symbol, date.Format("2006-01-02"))
}

func (c *HistoricalCache) rateKey(base, target string, date time.Time) string {
	return fmt.Sprintf("hist:fx:%s:%s:%s", base, target, date.Format("2006-01-02"))
}
