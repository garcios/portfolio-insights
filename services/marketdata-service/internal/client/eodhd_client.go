// Package client implements the EODHD API client.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)

// EODHDClient defines the interface for interacting with EODHD API
type EODHDClient interface {
	GetRealTimePrice(ctx context.Context, ticker, exchange string) (*RealTimePrice, error)
	GetHistoricalPrices(ctx context.Context, ticker, exchange string, from, to time.Time) ([]*HistoricalPrice, error)
}

// RealTimePrice represents a real-time price from EODHD API
type RealTimePrice struct {
	Code          string  `json:"code"`
	Timestamp     int64   `json:"timestamp"`
	GMTOffset     int     `json:"gmtoffset"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	Volume        int64   `json:"volume"`
	PreviousClose float64 `json:"previousClose"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_p"`
}

// HistoricalPrice represents a historical price from EODHD API
type HistoricalPrice struct {
	Date          string  `json:"date"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	AdjustedClose float64 `json:"adjusted_close"`
	Volume        int64   `json:"volume"`
}

// eodhd implements the EODHDClient interface
type eodhd struct {
	baseURL     string
	apiToken    string
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	retryConfig RetryConfig
}

// RetryConfig defines retry behavior for API calls
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
}

// DefaultRetryConfig returns sensible defaults for retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
	}
}

// NewEODHDClient creates a new EODHD API client
func NewEODHDClient(baseURL, apiToken string, requestsPerSecond float64) EODHDClient {
	return &eodhd{
		baseURL:  baseURL,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		rateLimiter: rate.NewLimiter(rate.Limit(requestsPerSecond), 1),
		retryConfig: DefaultRetryConfig(),
	}
}

// GetRealTimePrice fetches the latest real-time price for a ticker
func (c *eodhd) GetRealTimePrice(ctx context.Context, ticker, exchange string) (*RealTimePrice, error) {
	symbol := fmt.Sprintf("%s.%s", ticker, exchange)
	endpoint := fmt.Sprintf("%s/real-time/%s", c.baseURL, symbol)

	params := url.Values{}
	params.Set("api_token", c.apiToken)
	params.Set("fmt", "json")

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	var price RealTimePrice
	if err := c.doRequest(ctx, fullURL, &price); err != nil {
		return nil, fmt.Errorf("failed to get real-time price for %s: %w", symbol, err)
	}

	return &price, nil
}

// GetHistoricalPrices fetches historical EOD prices for a ticker within a date range
func (c *eodhd) GetHistoricalPrices(ctx context.Context, ticker, exchange string, from, to time.Time) ([]*HistoricalPrice, error) {
	symbol := fmt.Sprintf("%s.%s", ticker, exchange)
	endpoint := fmt.Sprintf("%s/eod/%s", c.baseURL, symbol)

	params := url.Values{}
	params.Set("api_token", c.apiToken)
	params.Set("fmt", "json")
	params.Set("from", from.Format("2006-01-02"))
	params.Set("to", to.Format("2006-01-02"))
	params.Set("period", "d")

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	var prices []*HistoricalPrice
	if err := c.doRequest(ctx, fullURL, &prices); err != nil {
		return nil, fmt.Errorf("failed to get historical prices for %s: %w", symbol, err)
	}

	return prices, nil
}

// doRequest performs an HTTP request with rate limiting, retries, and error handling
func (c *eodhd) doRequest(ctx context.Context, url string, result interface{}) error {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter error: %w", err)
	}

	// Perform request with retries
	backoff := c.retryConfig.InitialBackoff
	var lastErr error

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}

			// Increase backoff for next attempt
			backoff = time.Duration(float64(backoff) * c.retryConfig.Multiplier)
			if backoff > c.retryConfig.MaxBackoff {
				backoff = c.retryConfig.MaxBackoff
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("User-Agent", "portfolio-insights/1.0")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		// Handle response
		body, err := io.ReadAll(resp.Body)
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Ignore close error on read
			_ = closeErr
		}

		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		// Check status code
		switch resp.StatusCode {
		case http.StatusOK:
			// Success - parse response
			if err := json.Unmarshal(body, result); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			return nil

		case http.StatusTooManyRequests:
			// Rate limited - extract retry-after if available
			retryAfter := parseRetryAfter(resp.Header)
			if retryAfter > 0 {
				backoff = retryAfter
			}
			lastErr = fmt.Errorf("rate limited (429)")
			continue

		case http.StatusBadRequest:
			// Bad request - don't retry
			return fmt.Errorf("bad request (400): %s", string(body))

		case http.StatusUnauthorized:
			// Unauthorized - don't retry
			return fmt.Errorf("unauthorized (401): invalid API token")

		case http.StatusNotFound:
			// Not found - don't retry
			return fmt.Errorf("not found (404): %s", string(body))

		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			// Server errors - retry
			lastErr = fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
			continue

		default:
			// Unknown error - don't retry
			return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// parseRetryAfter extracts the Retry-After header value
func parseRetryAfter(headers http.Header) time.Duration {
	retryAfter := headers.Get("Retry-After")
	if retryAfter == "" {
		return 0
	}

	// Try parsing as seconds
	var seconds int
	if _, err := fmt.Sscanf(retryAfter, "%d", &seconds); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as HTTP date
	if t, err := http.ParseTime(retryAfter); err == nil {
		duration := time.Until(t)
		if duration > 0 {
			return duration
		}
	}

	return 0
}
