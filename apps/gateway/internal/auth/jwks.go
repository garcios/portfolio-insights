package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

// JWKSFetcher fetches and caches JWKS from Hydra
type JWKSFetcher struct {
	jwksURL    string
	cache      jwk.Set
	cacheMutex sync.RWMutex
	cacheTime  time.Time
	cacheTTL   time.Duration
	httpClient *http.Client
}

// NewJWKSFetcher creates a new JWKS fetcher
func NewJWKSFetcher(jwksURL string, cacheTTL time.Duration) *JWKSFetcher {
	return &JWKSFetcher{
		jwksURL:  jwksURL,
		cacheTTL: cacheTTL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetKeySet returns the JWKS, using cache if available
func (f *JWKSFetcher) GetKeySet(ctx context.Context) (jwk.Set, error) {
	// Check cache first
	f.cacheMutex.RLock()
	if f.cache != nil && time.Since(f.cacheTime) < f.cacheTTL {
		cache := f.cache
		f.cacheMutex.RUnlock()
		return cache, nil
	}
	f.cacheMutex.RUnlock()

	// Fetch new JWKS
	keySet, err := f.fetchJWKS(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	// Update cache
	f.cacheMutex.Lock()
	f.cache = keySet
	f.cacheTime = time.Now()
	f.cacheMutex.Unlock()

	return keySet, nil
}

// fetchJWKS fetches JWKS from the Hydra endpoint
func (f *JWKSFetcher) fetchJWKS(ctx context.Context) (jwk.Set, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", f.jwksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	// Parse JWKS
	var jwksData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jwksData); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	// Convert to jwk.Set
	jwksBytes, err := json.Marshal(jwksData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JWKS: %w", err)
	}

	keySet, err := jwk.Parse(jwksBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWKS: %w", err)
	}

	return keySet, nil
}

// RefreshCache forces a refresh of the JWKS cache
func (f *JWKSFetcher) RefreshCache(ctx context.Context) error {
	keySet, err := f.fetchJWKS(ctx)
	if err != nil {
		return err
	}

	f.cacheMutex.Lock()
	f.cache = keySet
	f.cacheTime = time.Now()
	f.cacheMutex.Unlock()

	return nil
}
