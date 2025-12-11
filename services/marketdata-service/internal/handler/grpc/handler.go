// Package grpc implements gRPC handlers for the marketdata service.
package grpc

import (
	"context"
	"database/sql"

	"github.com/garcios/portfolio-insights/pkg/resourcenames"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/domain"
	pb "github.com/garcios/portfolio-insights/services/marketdata-service/marketdata"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MarketDataHandler implements the gRPC market data service.
type MarketDataHandler struct {
	pb.UnimplementedMarketDataServiceServer
	usecase domain.MarketDataUsecase
}

// NewMarketDataHandler creates a new market data handler.
func NewMarketDataHandler(usecase domain.MarketDataUsecase) *MarketDataHandler {
	return &MarketDataHandler{usecase: usecase}
}

// GetAsset retrieves an asset by its resource name.
// AIP-131 compliant: uses resource name instead of ID or symbol.
func (h *MarketDataHandler) GetAsset(ctx context.Context, req *pb.GetAssetRequest) (*pb.Asset, error) {
	// Parse resource name to extract asset ID
	assetID, err := resourcenames.ParseAssetName(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
	}

	// Get asset by ID (which is the symbol in lowercase)
	asset, err := h.usecase.GetAsset(assetID)
	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "asset not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query asset: %v", err)
	}

	return mapAssetToProto(asset), nil
}

// GetAssetBySymbol retrieves an asset by symbol.
// This is a custom method (AIP-136) for symbol-based lookup.
func (h *MarketDataHandler) GetAssetBySymbol(ctx context.Context, req *pb.GetAssetBySymbolRequest) (*pb.Asset, error) {
	symbol := req.GetSymbol()
	if symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}

	asset, err := h.usecase.GetAsset(symbol)
	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "asset with symbol %s not found", symbol)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query asset: %v", err)
	}

	return mapAssetToProto(asset), nil
}

// ListAssets lists assets with pagination.
// AIP-132 compliant: uses page_size and page_token.
func (h *MarketDataHandler) ListAssets(ctx context.Context, req *pb.ListAssetsRequest) (*pb.ListAssetsResponse, error) {
	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 50 // Default per AIP-132
	}
	if pageSize > 1000 {
		pageSize = 1000 // Max per AIP-132
	}

	assets, nextPageToken, err := h.usecase.ListAssets(pageSize, req.GetPageToken())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list assets: %v", err)
	}

	var pbAssets []*pb.Asset
	for _, asset := range assets {
		pbAssets = append(pbAssets, mapAssetToProto(asset))
	}

	return &pb.ListAssetsResponse{
		Assets:        pbAssets,
		NextPageToken: nextPageToken,
	}, nil
}

// GetLatestPrice retrieves the latest price for an asset.
// Custom method using resource name.
func (h *MarketDataHandler) GetLatestPrice(ctx context.Context, req *pb.GetLatestPriceRequest) (*pb.AssetPrice, error) {
	// Parse asset resource name
	assetID, err := resourcenames.ParseAssetName(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
	}

	price, err := h.usecase.GetLatestPrice(assetID)
	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "price not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query price: %v", err)
	}

	return &pb.AssetPrice{
		AssetName: resourcenames.AssetName(price.AssetID),
		Price:     price.Price,
		Timestamp: timestamppb.New(price.Timestamp),
	}, nil
}

// GetLatestPrices retrieves the latest prices for multiple assets.
// Custom batch method using resource names.
func (h *MarketDataHandler) GetLatestPrices(ctx context.Context, req *pb.GetLatestPricesRequest) (*pb.GetLatestPricesResponse, error) {
	if len(req.Names) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one asset name is required")
	}

	// Parse resource names to extract asset IDs (symbols)
	var symbols []string
	for _, name := range req.Names {
		assetID, err := resourcenames.ParseAssetName(name)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid resource name %s: %v", name, err)
		}
		symbols = append(symbols, assetID)
	}

	prices, err := h.usecase.GetLatestPrices(symbols)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query prices: %v", err)
	}

	// Convert to proto map
	priceMap := make(map[string]*pb.AssetPrice)
	for symbol, price := range prices {
		priceMap[symbol] = &pb.AssetPrice{
			AssetName: resourcenames.AssetName(price.AssetID),
			Price:     price.Price,
			Timestamp: timestamppb.New(price.Timestamp),
		}
	}

	return &pb.GetLatestPricesResponse{
		Prices: priceMap,
	}, nil
}

// GetHistoricalPrices retrieves historical prices for an asset.
// Custom method using resource name.
func (h *MarketDataHandler) GetHistoricalPrices(ctx context.Context, req *pb.GetHistoricalPricesRequest) (*pb.GetHistoricalPricesResponse, error) {
	// Parse asset resource name
	assetID, err := resourcenames.ParseAssetName(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
	}

	startTime := req.GetStartTime().AsTime()
	endTime := req.GetEndTime().AsTime()

	if startTime.IsZero() || endTime.IsZero() {
		return nil, status.Error(codes.InvalidArgument, "start_time and end_time are required")
	}

	if endTime.Before(startTime) {
		return nil, status.Error(codes.InvalidArgument, "end_time must be after start_time")
	}

	prices, err := h.usecase.GetHistoricalPrices(assetID, startTime, endTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query historical prices: %v", err)
	}

	var pbPrices []*pb.AssetPrice
	for _, price := range prices {
		pbPrices = append(pbPrices, &pb.AssetPrice{
			AssetName: resourcenames.AssetName(price.AssetID),
			Price:     price.Price,
			Timestamp: timestamppb.New(price.Timestamp),
		})
	}

	return &pb.GetHistoricalPricesResponse{Prices: pbPrices}, nil
}

// GetLatestCurrencyRate retrieves the latest currency exchange rate.
// Custom method for currency rates.
func (h *MarketDataHandler) GetLatestCurrencyRate(ctx context.Context, req *pb.GetLatestCurrencyRateRequest) (*pb.CurrencyRate, error) {
	baseCurrency := req.GetBaseCurrency()
	targetCurrency := req.GetTargetCurrency()

	if baseCurrency == "" {
		return nil, status.Error(codes.InvalidArgument, "base_currency is required")
	}
	if targetCurrency == "" {
		return nil, status.Error(codes.InvalidArgument, "target_currency is required")
	}

	rate, err := h.usecase.GetLatestCurrencyRate(baseCurrency, targetCurrency)
	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "currency rate for %s/%s not found", baseCurrency, targetCurrency)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query currency rate: %v", err)
	}

	return mapCurrencyRateToProto(rate), nil
}

// GetHistoricalCurrencyRates retrieves historical currency exchange rates.
// Custom method for historical currency rates.
func (h *MarketDataHandler) GetHistoricalCurrencyRates(ctx context.Context, req *pb.GetHistoricalCurrencyRatesRequest) (*pb.GetHistoricalCurrencyRatesResponse, error) {
	baseCurrency := req.GetBaseCurrency()
	targetCurrency := req.GetTargetCurrency()

	if baseCurrency == "" {
		return nil, status.Error(codes.InvalidArgument, "base_currency is required")
	}
	if targetCurrency == "" {
		return nil, status.Error(codes.InvalidArgument, "target_currency is required")
	}

	startTime := req.GetStartTime().AsTime()
	endTime := req.GetEndTime().AsTime()

	if startTime.IsZero() || endTime.IsZero() {
		return nil, status.Error(codes.InvalidArgument, "start_time and end_time are required")
	}

	if endTime.Before(startTime) {
		return nil, status.Error(codes.InvalidArgument, "end_time must be after start_time")
	}

	rates, err := h.usecase.GetHistoricalCurrencyRates(baseCurrency, targetCurrency, startTime, endTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query historical currency rates: %v", err)
	}

	var pbRates []*pb.CurrencyRate
	for _, rate := range rates {
		pbRates = append(pbRates, mapCurrencyRateToProto(rate))
	}

	return &pb.GetHistoricalCurrencyRatesResponse{Rates: pbRates}, nil
}

// Helper function to map domain Asset to proto Asset
func mapAssetToProto(asset *domain.Asset) *pb.Asset {
	return &pb.Asset{
		Name:        resourcenames.AssetName(asset.ID),
		AssetId:     asset.ID,
		Symbol:      asset.Symbol,
		DisplayName: asset.Name,
		Type:        asset.Type,
		Exchange:    asset.Exchange,
		Currency:    asset.Currency,
	}
}

// Helper function to map domain CurrencyRate to proto CurrencyRate
func mapCurrencyRateToProto(rate *domain.CurrencyRate) *pb.CurrencyRate {
	// Construct currency rate resource name
	currencyRateID := rate.BaseCurrency + "-" + rate.TargetCurrency
	return &pb.CurrencyRate{
		Name:           resourcenames.CurrencyRateName(currencyRateID),
		CurrencyRateId: rate.ID,
		BaseCurrency:   rate.BaseCurrency,
		TargetCurrency: rate.TargetCurrency,
		Rate:           rate.Rate,
		RateDate:       timestamppb.New(rate.RateDate),
		CreatedAt:      timestamppb.New(rate.CreatedAt),
	}
}
