package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/domain"
	pb "github.com/garcios/portfolio-insights/services/marketdata-service/proto/marketdata"
)

type MockMarketDataUsecase struct {
	assets map[string]*domain.Asset
}

func (m *MockMarketDataUsecase) GetAsset(symbol string) (*domain.Asset, error) {
	if asset, ok := m.assets[symbol]; ok {
		return asset, nil
	}
	return nil, errors.New("not found")
}

func (m *MockMarketDataUsecase) ListAssets(pageSize int, pageToken string) ([]*domain.Asset, string, error) {
	return nil, "", nil
}

func (m *MockMarketDataUsecase) GetLatestPrice(symbol string) (*domain.AssetPrice, error) {
	if symbol == "AAPL" {
		return &domain.AssetPrice{Price: 150.0, Timestamp: time.Now()}, nil
	}
	return nil, errors.New("not found")
}

func (m *MockMarketDataUsecase) GetLatestPrices(symbols []string) (map[string]*domain.AssetPrice, error) {
	result := make(map[string]*domain.AssetPrice)
	for _, symbol := range symbols {
		if symbol == "AAPL" || symbol == "GOOGL" {
			result[symbol] = &domain.AssetPrice{Price: 150.0, Timestamp: time.Now()}
		}
	}
	return result, nil
}

func (m *MockMarketDataUsecase) GetHistoricalPrices(symbol string, start, end time.Time) ([]*domain.AssetPrice, error) {
	return nil, nil
}

func (m *MockMarketDataUsecase) GetLatestCurrencyRate(baseCurrency, targetCurrency string) (*domain.CurrencyRate, error) {
	return nil, errors.New("not implemented")
}

func TestGetAssetHandler(t *testing.T) {
	mockUC := &MockMarketDataUsecase{
		assets: map[string]*domain.Asset{
			"AAPL": {Symbol: "AAPL", Name: "Apple"},
		},
	}
	handler := NewMarketDataHandler(mockUC)

	t.Run("Success", func(t *testing.T) {
		req := &pb.GetAssetRequest{Symbol: "AAPL"}
		resp, err := handler.GetAsset(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.Asset.Name != "Apple" {
			t.Errorf("expected Apple, got %s", resp.Asset.Name)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		req := &pb.GetAssetRequest{Symbol: "MSFT"}
		_, err := handler.GetAsset(context.Background(), req)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
