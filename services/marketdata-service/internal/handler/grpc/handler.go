// Package grpc implements gRPC handlers for the marketdata service.
package grpc

import (
	"context"
	"database/sql"

	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/domain"
	pb "github.com/garcios/portfolio-insights/services/marketdata-service/proto/marketdata"
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

// GetAsset retrieves an asset by symbol.
func (h *MarketDataHandler) GetAsset(ctx context.Context, req *pb.GetAssetRequest) (*pb.GetAssetResponse, error) {
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

	return &pb.GetAssetResponse{
		Asset: &pb.Asset{
			Id:       asset.ID,
			Symbol:   asset.Symbol,
			Name:     asset.Name,
			Type:     asset.Type,
			Exchange: asset.Exchange,
			Currency: asset.Currency,
		},
	}, nil
}

// ListAssets lists assets with pagination.
func (h *MarketDataHandler) ListAssets(ctx context.Context, req *pb.ListAssetsRequest) (*pb.ListAssetsResponse, error) {
	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	assets, nextPageToken, err := h.usecase.ListAssets(pageSize, req.GetPageToken())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list assets: %v", err)
	}

	var pbAssets []*pb.Asset
	for _, asset := range assets {
		pbAssets = append(pbAssets, &pb.Asset{
			Id:       asset.ID,
			Symbol:   asset.Symbol,
			Name:     asset.Name,
			Type:     asset.Type,
			Exchange: asset.Exchange,
			Currency: asset.Currency,
		})
	}

	return &pb.ListAssetsResponse{
		Assets:        pbAssets,
		NextPageToken: nextPageToken,
	}, nil
}

// GetLatestPrice retrieves the latest price for an asset.
func (h *MarketDataHandler) GetLatestPrice(ctx context.Context, req *pb.GetLatestPriceRequest) (*pb.GetLatestPriceResponse, error) {
	symbol := req.GetSymbol()
	if symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}

	price, err := h.usecase.GetLatestPrice(symbol)
	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "price for symbol %s not found", symbol)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query price: %v", err)
	}

	return &pb.GetLatestPriceResponse{
		Price: &pb.AssetPrice{
			AssetId:   price.AssetID,
			Price:     price.Price,
			Timestamp: timestamppb.New(price.Timestamp),
		},
	}, nil
}

// GetLatestPrices retrieves the latest prices for multiple assets.
func (h *MarketDataHandler) GetLatestPrices(ctx context.Context, req *pb.GetLatestPricesRequest) (*pb.GetLatestPricesResponse, error) {
	symbols := req.GetSymbols()
	if len(symbols) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one symbol is required")
	}

	prices, err := h.usecase.GetLatestPrices(symbols)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query prices: %v", err)
	}

	// Convert to proto map
	priceMap := make(map[string]*pb.AssetPrice)
	for symbol, price := range prices {
		priceMap[symbol] = &pb.AssetPrice{
			AssetId:   price.AssetID,
			Price:     price.Price,
			Timestamp: timestamppb.New(price.Timestamp),
		}
	}

	return &pb.GetLatestPricesResponse{
		Prices: priceMap,
	}, nil
}

// GetHistoricalPrices retrieves historical prices for an asset.
func (h *MarketDataHandler) GetHistoricalPrices(ctx context.Context, req *pb.GetHistoricalPricesRequest) (*pb.GetHistoricalPricesResponse, error) {
	symbol := req.GetSymbol()
	if symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}

	startTime := req.GetStartTime().AsTime()
	endTime := req.GetEndTime().AsTime()

	if startTime.IsZero() || endTime.IsZero() {
		return nil, status.Error(codes.InvalidArgument, "start_time and end_time are required")
	}

	if endTime.Before(startTime) {
		return nil, status.Error(codes.InvalidArgument, "end_time must be after start_time")
	}

	prices, err := h.usecase.GetHistoricalPrices(symbol, startTime, endTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query historical prices: %v", err)
	}

	var pbPrices []*pb.AssetPrice
	for _, price := range prices {
		pbPrices = append(pbPrices, &pb.AssetPrice{
			AssetId:   price.AssetID,
			Price:     price.Price,
			Timestamp: timestamppb.New(price.Timestamp),
		})
	}

	return &pb.GetHistoricalPricesResponse{Prices: pbPrices}, nil
}

// GetLatestCurrencyRate retrieves the latest currency exchange rate.
func (h *MarketDataHandler) GetLatestCurrencyRate(ctx context.Context, req *pb.GetLatestCurrencyRateRequest) (*pb.GetLatestCurrencyRateResponse, error) {
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

	return &pb.GetLatestCurrencyRateResponse{
		CurrencyRate: &pb.CurrencyRate{
			Id:             rate.ID,
			BaseCurrency:   rate.BaseCurrency,
			TargetCurrency: rate.TargetCurrency,
			Rate:           rate.Rate,
			RateDate:       timestamppb.New(rate.RateDate),
			CreatedAt:      timestamppb.New(rate.CreatedAt),
		},
	}, nil
}

// GetHistoricalCurrencyRates retrieves historical currency exchange rates.
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
		pbRates = append(pbRates, &pb.CurrencyRate{
			Id:             rate.ID,
			BaseCurrency:   rate.BaseCurrency,
			TargetCurrency: rate.TargetCurrency,
			Rate:           rate.Rate,
			RateDate:       timestamppb.New(rate.RateDate),
			CreatedAt:      timestamppb.New(rate.CreatedAt),
		})
	}

	return &pb.GetHistoricalCurrencyRatesResponse{Rates: pbRates}, nil
}
