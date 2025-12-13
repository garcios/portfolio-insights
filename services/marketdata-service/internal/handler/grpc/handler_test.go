package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/pkg/resourcenames"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/domain"
	pb "github.com/garcios/portfolio-insights/services/marketdata-service/marketdata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MockMarketDataUsecase struct {
	assets map[string]*domain.Asset
}

func (m *MockMarketDataUsecase) GetAsset(assetID string) (*domain.Asset, error) {
	if asset, ok := m.assets[assetID]; ok {
		return asset, nil
	}
	return nil, errors.New("not found")
}

func (m *MockMarketDataUsecase) ListAssets(pageSize int, pageToken string) ([]*domain.Asset, string, error) {
	var assets []*domain.Asset
	for _, asset := range m.assets {
		assets = append(assets, asset)
	}
	return assets, "", nil
}

func (m *MockMarketDataUsecase) GetLatestPrice(assetID string) (*domain.AssetPrice, error) {
	if assetID == "aapl" {
		return &domain.AssetPrice{AssetID: assetID, Price: 150.0, Timestamp: time.Now()}, nil
	}
	return nil, errors.New("not found")
}

func (m *MockMarketDataUsecase) GetLatestPrices(symbols []string) (map[string]*domain.AssetPrice, error) {
	result := make(map[string]*domain.AssetPrice)
	for _, symbol := range symbols {
		if symbol == "AAPL" || symbol == "GOOGL" {
			result[symbol] = &domain.AssetPrice{AssetID: symbol, Price: 150.0, Timestamp: time.Now()}
		}
	}
	return result, nil
}

func (m *MockMarketDataUsecase) GetHistoricalPrices(assetID string, start, end time.Time) ([]*domain.AssetPrice, error) {
	if assetID == "aapl" {
		return []*domain.AssetPrice{
			{AssetID: assetID, Price: 145.0, Timestamp: start},
			{AssetID: assetID, Price: 150.0, Timestamp: end},
		}, nil
	}
	return nil, nil
}

func (m *MockMarketDataUsecase) GetLatestCurrencyRate(baseCurrency, targetCurrency string) (*domain.CurrencyRate, error) {
	if baseCurrency == "USD" && targetCurrency == "EUR" {
		return &domain.CurrencyRate{
			ID:             "1",
			BaseCurrency:   "USD",
			TargetCurrency: "EUR",
			Rate:           0.85,
			RateDate:       time.Now(),
			CreatedAt:      time.Now(),
		}, nil
	}
	return nil, errors.New("not found")
}

func (m *MockMarketDataUsecase) GetHistoricalCurrencyRates(baseCurrency, targetCurrency string, start, end time.Time) ([]*domain.CurrencyRate, error) {
	return nil, nil
}

func TestGetAsset_Success(t *testing.T) {
	mockUC := &MockMarketDataUsecase{
		assets: map[string]*domain.Asset{
			"aapl": {ID: "aapl", Symbol: "AAPL", Name: "Apple Inc.", Type: "stock"},
		},
	}
	handler := NewMarketDataHandler(mockUC)

	req := &pb.GetAssetRequest{
		Name: resourcenames.AssetName("aapl"),
	}
	resp, err := handler.GetAsset(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.AssetId != "aapl" {
		t.Errorf("expected asset_id aapl, got %s", resp.AssetId)
	}

	if resp.DisplayName != "Apple Inc." {
		t.Errorf("expected display_name 'Apple Inc.', got %s", resp.DisplayName)
	}

	expectedName := resourcenames.AssetName("aapl")
	if resp.Name != expectedName {
		t.Errorf("expected name %s, got %s", expectedName, resp.Name)
	}
}

func TestGetAsset_InvalidResourceName(t *testing.T) {
	mockUC := &MockMarketDataUsecase{assets: make(map[string]*domain.Asset)}
	handler := NewMarketDataHandler(mockUC)

	req := &pb.GetAssetRequest{
		Name: "invalid-name",
	}
	_, err := handler.GetAsset(context.Background(), req)

	if err == nil {
		t.Fatal("expected error for invalid resource name, got nil")
	}
}

func TestGetAssetBySymbol_Success(t *testing.T) {
	mockUC := &MockMarketDataUsecase{
		assets: map[string]*domain.Asset{
			"aapl": {ID: "aapl", Symbol: "AAPL", Name: "Apple Inc."},
		},
	}
	handler := NewMarketDataHandler(mockUC)

	req := &pb.GetAssetBySymbolRequest{
		Symbol: "aapl",
	}
	resp, err := handler.GetAssetBySymbol(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", resp.Symbol)
	}
}

func TestGetAssetBySymbol_EmptySymbol(t *testing.T) {
	mockUC := &MockMarketDataUsecase{assets: make(map[string]*domain.Asset)}
	handler := NewMarketDataHandler(mockUC)

	req := &pb.GetAssetBySymbolRequest{
		Symbol: "",
	}
	_, err := handler.GetAssetBySymbol(context.Background(), req)

	if err == nil {
		t.Fatal("expected error for empty symbol, got nil")
	}
}

func TestListAssets_Success(t *testing.T) {
	mockUC := &MockMarketDataUsecase{
		assets: map[string]*domain.Asset{
			"aapl":  {ID: "aapl", Symbol: "AAPL", Name: "Apple Inc."},
			"googl": {ID: "googl", Symbol: "GOOGL", Name: "Alphabet Inc."},
		},
	}
	handler := NewMarketDataHandler(mockUC)

	req := &pb.ListAssetsRequest{
		PageSize: 10,
	}
	resp, err := handler.ListAssets(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.Assets) != 2 {
		t.Errorf("expected 2 assets, got %d", len(resp.Assets))
	}

	// Verify all assets have resource names
	for _, asset := range resp.Assets {
		if asset.Name == "" {
			t.Error("expected asset to have resource name")
		}
		if asset.AssetId == "" {
			t.Error("expected asset to have asset_id")
		}
	}
}

func TestGetLatestPrice_Success(t *testing.T) {
	mockUC := &MockMarketDataUsecase{
		assets: map[string]*domain.Asset{
			"aapl": {ID: "aapl", Symbol: "AAPL"},
		},
	}
	handler := NewMarketDataHandler(mockUC)

	req := &pb.GetLatestPriceRequest{
		Name: resourcenames.AssetName("aapl"),
	}
	resp, err := handler.GetLatestPrice(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Price != 150.0 {
		t.Errorf("expected price 150.0, got %f", resp.Price)
	}
}

func TestGetLatestPrice_InvalidResourceName(t *testing.T) {
	mockUC := &MockMarketDataUsecase{assets: make(map[string]*domain.Asset)}
	handler := NewMarketDataHandler(mockUC)

	req := &pb.GetLatestPriceRequest{
		Name: "invalid-name",
	}
	_, err := handler.GetLatestPrice(context.Background(), req)

	if err == nil {
		t.Fatal("expected error for invalid resource name, got nil")
	}
}

func TestGetHistoricalPrices_Success(t *testing.T) {
	mockUC := &MockMarketDataUsecase{
		assets: map[string]*domain.Asset{
			"aapl": {ID: "aapl", Symbol: "AAPL"},
		},
	}
	handler := NewMarketDataHandler(mockUC)

	now := time.Now()
	req := &pb.GetHistoricalPricesRequest{
		Name:      resourcenames.AssetName("aapl"),
		StartTime: timestamppb.New(now.Add(-24 * time.Hour)),
		EndTime:   timestamppb.New(now),
	}
	resp, err := handler.GetHistoricalPrices(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.Prices) != 2 {
		t.Errorf("expected 2 prices, got %d", len(resp.Prices))
	}
}

func TestGetLatestCurrencyRate_Success(t *testing.T) {
	mockUC := &MockMarketDataUsecase{}
	handler := NewMarketDataHandler(mockUC)

	req := &pb.GetLatestCurrencyRateRequest{
		BaseCurrency:   "USD",
		TargetCurrency: "EUR",
	}
	resp, err := handler.GetLatestCurrencyRate(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Rate != 0.85 {
		t.Errorf("expected rate 0.85, got %f", resp.Rate)
	}

	// Verify resource name is set
	if resp.Name == "" {
		t.Error("expected currency rate to have resource name")
	}
}
