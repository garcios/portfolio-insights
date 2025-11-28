package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/domain"
)

// MockRepository
type MockMarketDataRepository struct {
	assets map[string]*domain.Asset
	prices map[string][]*domain.AssetPrice
}

func NewMockRepo() *MockMarketDataRepository {
	return &MockMarketDataRepository{
		assets: make(map[string]*domain.Asset),
		prices: make(map[string][]*domain.AssetPrice),
	}
}

func (m *MockMarketDataRepository) GetAssetBySymbol(symbol string) (*domain.Asset, error) {
	if asset, ok := m.assets[symbol]; ok {
		return asset, nil
	}
	return nil, errors.New("not found")
}

func (m *MockMarketDataRepository) ListAssets(limit, offset int) ([]*domain.Asset, error) {
	var result []*domain.Asset
	// Simple mock implementation for pagination
	// Convert map to slice (order not guaranteed without sorting, but sufficient for basic check)
	allAssets := make([]*domain.Asset, 0, len(m.assets))
	for _, a := range m.assets {
		allAssets = append(allAssets, a)
	}

	start := offset
	end := offset + limit
	if start >= len(allAssets) {
		return result, nil
	}
	if end > len(allAssets) {
		end = len(allAssets)
	}

	return allAssets[start:end], nil
}

func (m *MockMarketDataRepository) UpsertAssets(assets []*domain.Asset) error {
	for _, a := range assets {
		m.assets[a.Symbol] = a
	}
	return nil
}

func (m *MockMarketDataRepository) GetAllAssetIDs() (map[string]string, error) {
	return nil, nil
}

func (m *MockMarketDataRepository) GetLatestPrice(symbol string) (*domain.AssetPrice, error) {
	if prices, ok := m.prices[symbol]; ok && len(prices) > 0 {
		return prices[len(prices)-1], nil // Return last added
	}
	return nil, errors.New("not found")
}

func (m *MockMarketDataRepository) GetLatestPrices(symbols []string) (map[string]*domain.AssetPrice, error) {
	result := make(map[string]*domain.AssetPrice)
	for _, symbol := range symbols {
		if prices, ok := m.prices[symbol]; ok && len(prices) > 0 {
			result[symbol] = prices[len(prices)-1]
		}
	}
	return result, nil
}

func (m *MockMarketDataRepository) GetHistoricalPrices(symbol string, start, end time.Time) ([]*domain.AssetPrice, error) {
	if prices, ok := m.prices[symbol]; ok {
		return prices, nil
	}
	return nil, nil
}

func (m *MockMarketDataRepository) InsertPrices(prices []*domain.AssetPrice) error {
	return nil
}

func (m *MockMarketDataRepository) CountAssets() (int, error) {
	return len(m.assets), nil
}

func (m *MockMarketDataRepository) CountPrices() (int, error) {
	count := 0
	for _, p := range m.prices {
		count += len(p)
	}
	return count, nil
}

func (m *MockMarketDataRepository) InsertCurrencyRates(rates []*domain.CurrencyRate) error {
	return nil
}

func (m *MockMarketDataRepository) GetLatestCurrencyRate(baseCurrency, targetCurrency string) (*domain.CurrencyRate, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMarketDataRepository) GetHistoricalCurrencyRates(baseCurrency, targetCurrency string, start, end time.Time) ([]*domain.CurrencyRate, error) {
	return nil, nil
}

func TestGetAsset(t *testing.T) {
	repo := NewMockRepo()
	repo.assets["AAPL"] = &domain.Asset{Symbol: "AAPL", Name: "Apple"}

	uc := NewMarketDataUsecase(repo)

	t.Run("Success", func(t *testing.T) {
		asset, err := uc.GetAsset("AAPL")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if asset.Name != "Apple" {
			t.Errorf("expected Apple, got %s", asset.Name)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := uc.GetAsset("MSFT")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestGetLatestPrice(t *testing.T) {
	repo := NewMockRepo()
	repo.prices["AAPL"] = []*domain.AssetPrice{
		{Price: 150.0, Timestamp: time.Now()},
	}

	uc := NewMarketDataUsecase(repo)

	t.Run("Success", func(t *testing.T) {
		price, err := uc.GetLatestPrice("AAPL")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if price.Price != 150.0 {
			t.Errorf("expected 150.0, got %f", price.Price)
		}
	})
}
