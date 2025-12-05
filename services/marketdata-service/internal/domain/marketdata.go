package domain

import "time"

type Asset struct {
	ID        string
	Symbol    string
	Name      string
	Type      string
	Exchange  string
	Currency  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AssetPrice struct {
	ID        string
	AssetID   string
	Price     float64
	Timestamp time.Time
	CreatedAt time.Time
}

type CurrencyRate struct {
	ID             string
	BaseCurrency   string
	TargetCurrency string
	Rate           float64
	RateDate       time.Time
	CreatedAt      time.Time
}

type MarketDataRepository interface {
	// Asset operations
	GetAssetBySymbol(symbol string) (*Asset, error)
	ListAssets(limit, offset int) ([]*Asset, error)
	UpsertAssets(assets []*Asset) error         // For worker
	GetAllAssetIDs() (map[string]string, error) // For price worker
	CountAssets() (int, error)                  // For metrics

	// Price operations
	GetLatestPrice(symbol string) (*AssetPrice, error)
	GetLatestPrices(symbols []string) (map[string]*AssetPrice, error)
	GetHistoricalPrices(symbol string, start, end time.Time) ([]*AssetPrice, error)
	InsertPrices(prices []*AssetPrice) error // For worker
	CountPrices() (int, error)               // For metrics

	// Currency rate operations
	InsertCurrencyRates(rates []*CurrencyRate) error // For worker
	GetLatestCurrencyRate(baseCurrency, targetCurrency string) (*CurrencyRate, error)
	GetHistoricalCurrencyRates(baseCurrency, targetCurrency string, start, end time.Time) ([]*CurrencyRate, error)

	// EODHD price sync operations
	GetAssetsRequiringPriceUpdate(staleDuration time.Duration) ([]*Asset, error)    // For EODHD sync worker
	GetLatestPriceTimestamp(assetID string) (*time.Time, error)                     // For EODHD sync worker
	GetMissingPriceDates(assetID string, start, end time.Time) ([]time.Time, error) // For EODHD sync worker

	// EODHD currency sync operations
	GetTargetCurrencies() ([]string, error)
	GetLatestCurrencyRateTimestamp(baseCurrency, targetCurrency string) (*time.Time, error)
}

type MarketDataUsecase interface {
	GetAsset(symbol string) (*Asset, error)
	ListAssets(pageSize int, pageToken string) ([]*Asset, string, error)
	GetLatestPrice(symbol string) (*AssetPrice, error)
	GetLatestPrices(symbols []string) (map[string]*AssetPrice, error)
	GetHistoricalPrices(symbol string, start, end time.Time) ([]*AssetPrice, error)
	GetLatestCurrencyRate(baseCurrency, targetCurrency string) (*CurrencyRate, error)
	GetHistoricalCurrencyRates(baseCurrency, targetCurrency string, start, end time.Time) ([]*CurrencyRate, error)
}
