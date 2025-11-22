package usecase

import (
	"strconv"
	"time"

	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/domain"
)

type marketDataUsecase struct {
	repo domain.MarketDataRepository
}

func NewMarketDataUsecase(repo domain.MarketDataRepository) domain.MarketDataUsecase {
	return &marketDataUsecase{repo: repo}
}

func (uc *marketDataUsecase) GetAsset(symbol string) (*domain.Asset, error) {
	return uc.repo.GetAssetBySymbol(symbol)
}

func (uc *marketDataUsecase) ListAssets(pageSize int, pageToken string) ([]*domain.Asset, string, error) {
	limit := pageSize
	offset := 0
	if pageToken != "" {
		var err error
		offset, err = strconv.Atoi(pageToken)
		if err != nil {
			return nil, "", err
		}
	}

	assets, err := uc.repo.ListAssets(limit, offset)
	if err != nil {
		return nil, "", err
	}

	nextPageToken := ""
	if len(assets) == limit {
		nextPageToken = strconv.Itoa(offset + limit)
	}

	return assets, nextPageToken, nil
}

func (uc *marketDataUsecase) GetLatestPrice(symbol string) (*domain.AssetPrice, error) {
	return uc.repo.GetLatestPrice(symbol)
}

func (uc *marketDataUsecase) GetHistoricalPrices(symbol string, start, end time.Time) ([]*domain.AssetPrice, error) {
	return uc.repo.GetHistoricalPrices(symbol, start, end)
}
