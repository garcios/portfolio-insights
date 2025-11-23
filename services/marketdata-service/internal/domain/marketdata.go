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

type MarketDataRepository interface {
	// Asset operations
	GetAssetBySymbol(symbol string) (*Asset, error)
	ListAssets(limit, offset int) ([]*Asset, error)
	UpsertAssets(assets []*Asset) error         // For worker
	GetAllAssetIDs() (map[string]string, error) // For price worker

	// Price operations
	GetLatestPrice(symbol string) (*AssetPrice, error)
	GetLatestPrices(symbols []string) (map[string]*AssetPrice, error)
	GetHistoricalPrices(symbol string, start, end time.Time) ([]*AssetPrice, error)
	InsertPrices(prices []*AssetPrice) error // For worker
}

type MarketDataUsecase interface {
	GetAsset(symbol string) (*Asset, error)
	ListAssets(pageSize int, pageToken string) ([]*Asset, string, error)
	GetLatestPrice(symbol string) (*AssetPrice, error)
	GetLatestPrices(symbols []string) (map[string]*AssetPrice, error)
	GetHistoricalPrices(symbol string, start, end time.Time) ([]*AssetPrice, error)
}
