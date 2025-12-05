package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/domain"
)

// MockRepository
type MockMarketDataRepository struct {
	assets        map[string]*domain.Asset
	prices        map[string][]*domain.AssetPrice
	currencyRates map[string]*domain.CurrencyRate
}

func NewMockRepo() *MockMarketDataRepository {
	return &MockMarketDataRepository{
		assets:        make(map[string]*domain.Asset),
		prices:        make(map[string][]*domain.AssetPrice),
		currencyRates: make(map[string]*domain.CurrencyRate),
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
		return prices[len(prices)-1], nil
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
	key := baseCurrency + "-" + targetCurrency
	if rate, ok := m.currencyRates[key]; ok {
		return rate, nil
	}
	return nil, errors.New("not found")
}

func (m *MockMarketDataRepository) GetHistoricalCurrencyRates(baseCurrency, targetCurrency string, start, end time.Time) ([]*domain.CurrencyRate, error) {
	return nil, nil
}

// EODHD price sync methods
func (m *MockMarketDataRepository) GetAssetsRequiringPriceUpdate(staleDuration time.Duration) ([]*domain.Asset, error) {
	return nil, nil
}

func (m *MockMarketDataRepository) GetLatestPriceTimestamp(assetID string) (*time.Time, error) {
	return nil, nil
}

func (m *MockMarketDataRepository) GetMissingPriceDates(assetID string, start, end time.Time) ([]time.Time, error) {
	return nil, nil
}

func TestGetAsset(t *testing.T) {
	repo := NewMockRepo()
	repo.assets["AAPL"] = &domain.Asset{Symbol: "AAPL", Name: "Apple Inc.", Type: "stock"}

	uc := NewMarketDataUsecase(repo)

	t.Run("Success", func(t *testing.T) {
		asset, err := uc.GetAsset("AAPL")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if asset.Name != "Apple Inc." {
			t.Errorf("expected 'Apple Inc.', got %s", asset.Name)
		}
		if asset.Type != "stock" {
			t.Errorf("expected 'stock', got %s", asset.Type)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := uc.GetAsset("MSFT")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestListAssets(t *testing.T) {
	repo := NewMockRepo()
	repo.assets["AAPL"] = &domain.Asset{Symbol: "AAPL", Name: "Apple Inc."}
	repo.assets["GOOGL"] = &domain.Asset{Symbol: "GOOGL", Name: "Alphabet Inc."}
	repo.assets["MSFT"] = &domain.Asset{Symbol: "MSFT", Name: "Microsoft Corp."}

	uc := NewMarketDataUsecase(repo)

	t.Run("FirstPage", func(t *testing.T) {
		assets, nextToken, err := uc.ListAssets(2, "")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(assets) != 2 {
			t.Errorf("expected 2 assets, got %d", len(assets))
		}
		if nextToken == "" {
			t.Error("expected non-empty next token")
		}
	})

	t.Run("AllAssets", func(t *testing.T) {
		assets, _, err := uc.ListAssets(10, "")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(assets) != 3 {
			t.Errorf("expected 3 assets, got %d", len(assets))
		}
	})
}

func TestGetLatestPrice(t *testing.T) {
	repo := NewMockRepo()
	now := time.Now()
	repo.prices["AAPL"] = []*domain.AssetPrice{
		{Price: 150.0, Timestamp: now.Add(-time.Hour)},
		{Price: 155.0, Timestamp: now},
	}

	uc := NewMarketDataUsecase(repo)

	t.Run("Success", func(t *testing.T) {
		price, err := uc.GetLatestPrice("AAPL")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if price.Price != 155.0 {
			t.Errorf("expected 155.0, got %f", price.Price)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := uc.GetLatestPrice("INVALID")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestGetLatestPrices(t *testing.T) {
	repo := NewMockRepo()
	now := time.Now()
	repo.prices["AAPL"] = []*domain.AssetPrice{{Price: 150.0, Timestamp: now}}
	repo.prices["GOOGL"] = []*domain.AssetPrice{{Price: 2500.0, Timestamp: now}}

	uc := NewMarketDataUsecase(repo)

	t.Run("Success", func(t *testing.T) {
		prices, err := uc.GetLatestPrices([]string{"AAPL", "GOOGL"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(prices) != 2 {
			t.Errorf("expected 2 prices, got %d", len(prices))
		}
		if prices["AAPL"].Price != 150.0 {
			t.Errorf("expected AAPL price 150.0, got %f", prices["AAPL"].Price)
		}
		if prices["GOOGL"].Price != 2500.0 {
			t.Errorf("expected GOOGL price 2500.0, got %f", prices["GOOGL"].Price)
		}
	})

	t.Run("PartialResults", func(t *testing.T) {
		prices, err := uc.GetLatestPrices([]string{"AAPL", "INVALID"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(prices) != 1 {
			t.Errorf("expected 1 price, got %d", len(prices))
		}
	})
}

func TestGetHistoricalPrices(t *testing.T) {
	repo := NewMockRepo()
	start := time.Now().AddDate(0, 0, -7)
	end := time.Now()

	repo.prices["AAPL"] = []*domain.AssetPrice{
		{Price: 145.0, Timestamp: start},
		{Price: 150.0, Timestamp: start.AddDate(0, 0, 3)},
		{Price: 155.0, Timestamp: end},
	}

	uc := NewMarketDataUsecase(repo)

	t.Run("Success", func(t *testing.T) {
		prices, err := uc.GetHistoricalPrices("AAPL", start, end)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(prices) != 3 {
			t.Errorf("expected 3 prices, got %d", len(prices))
		}
	})

	t.Run("EmptyResult", func(t *testing.T) {
		prices, err := uc.GetHistoricalPrices("INVALID", start, end)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(prices) != 0 {
			t.Errorf("expected 0 prices, got %d", len(prices))
		}
	})
}

func TestGetLatestCurrencyRate(t *testing.T) {
	repo := NewMockRepo()
	now := time.Now()
	repo.currencyRates["USD-EUR"] = &domain.CurrencyRate{
		BaseCurrency:   "USD",
		TargetCurrency: "EUR",
		Rate:           0.85,
		RateDate:       now,
	}

	uc := NewMarketDataUsecase(repo)

	t.Run("Success", func(t *testing.T) {
		rate, err := uc.GetLatestCurrencyRate("USD", "EUR")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rate.Rate != 0.85 {
			t.Errorf("expected rate 0.85, got %f", rate.Rate)
		}
		if rate.BaseCurrency != "USD" {
			t.Errorf("expected base currency 'USD', got '%s'", rate.BaseCurrency)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := uc.GetLatestCurrencyRate("USD", "GBP")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestGetHistoricalCurrencyRates(t *testing.T) {
	repo := NewMockRepo()
	uc := NewMarketDataUsecase(repo)

	start := time.Now().AddDate(0, 0, -7)
	end := time.Now()

	t.Run("Success", func(t *testing.T) {
		rates, err := uc.GetHistoricalCurrencyRates("USD", "EUR", start, end)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		// Mock returns nil, so we expect empty slice
		if rates != nil && len(rates) != 0 {
			t.Errorf("expected empty rates, got %d", len(rates))
		}
	})
}
