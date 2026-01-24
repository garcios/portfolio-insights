// Package usecase implements the business logic for the marketdata service.
package usecase

import (
	"strconv"
	"time"

	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/domain"
)

type marketDataUsecase struct {
	repo domain.MarketDataRepository
}

// NewMarketDataUsecase creates a new market data usecase.
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

func (uc *marketDataUsecase) GetLatestPrices(symbols []string) (map[string]*domain.AssetPrice, error) {
	return uc.repo.GetLatestPrices(symbols)
}

func (uc *marketDataUsecase) GetHistoricalPrices(symbol string, start, end time.Time) ([]*domain.AssetPrice, error) {
	return uc.repo.GetHistoricalPrices(symbol, start, end)
}

func (uc *marketDataUsecase) GetLatestCurrencyRate(baseCurrency, targetCurrency string) (*domain.CurrencyRate, error) {
	rate, err := uc.repo.GetLatestCurrencyRate(baseCurrency, targetCurrency)
	if err == nil {
		return rate, nil
	}

	// If not found, try inverse
	inverseRate, err := uc.repo.GetLatestCurrencyRate(targetCurrency, baseCurrency)
	if err != nil {
		return nil, err
	}

	return &domain.CurrencyRate{
		ID:             inverseRate.ID,
		BaseCurrency:   baseCurrency,
		TargetCurrency: targetCurrency,
		Rate:           1 / inverseRate.Rate,
		RateDate:       inverseRate.RateDate,
		CreatedAt:      inverseRate.CreatedAt,
	}, nil
}

func (uc *marketDataUsecase) GetHistoricalCurrencyRates(baseCurrency, targetCurrency string, start, end time.Time) ([]*domain.CurrencyRate, error) {
	rates, err := uc.repo.GetHistoricalCurrencyRates(baseCurrency, targetCurrency, start, end)
	if err == nil && len(rates) > 0 {
		return rates, nil
	}

	// If not found, try inverse
	inverseRates, err := uc.repo.GetHistoricalCurrencyRates(targetCurrency, baseCurrency, start, end)
	if err != nil {
		return nil, err
	}

	var computedRates []*domain.CurrencyRate
	for _, r := range inverseRates {
		computedRates = append(computedRates, &domain.CurrencyRate{
			ID:             r.ID,
			BaseCurrency:   baseCurrency,
			TargetCurrency: targetCurrency,
			Rate:           1 / r.Rate,
			RateDate:       r.RateDate,
			CreatedAt:      r.CreatedAt,
		})
	}

	return computedRates, nil
}
