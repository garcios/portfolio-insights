package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/domain"
)

type postgresMarketDataRepo struct {
	db *sql.DB
}

func NewPostgresMarketDataRepository(db *sql.DB) domain.MarketDataRepository {
	return &postgresMarketDataRepo{db: db}
}

func (r *postgresMarketDataRepo) GetAssetBySymbol(symbol string) (*domain.Asset, error) {
	query := `
		SELECT id, symbol, name, type, exchange, currency, created_at, updated_at
		FROM marketdata.assets
		WHERE symbol = $1
	`
	var asset domain.Asset
	err := r.db.QueryRow(query, symbol).Scan(
		&asset.ID, &asset.Symbol, &asset.Name, &asset.Type, &asset.Exchange, &asset.Currency, &asset.CreatedAt, &asset.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *postgresMarketDataRepo) ListAssets(limit, offset int) ([]*domain.Asset, error) {
	query := `
		SELECT id, symbol, name, type, exchange, currency, created_at, updated_at
		FROM marketdata.assets
		ORDER BY symbol
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*domain.Asset
	for rows.Next() {
		var asset domain.Asset
		if err := rows.Scan(&asset.ID, &asset.Symbol, &asset.Name, &asset.Type, &asset.Exchange, &asset.Currency, &asset.CreatedAt, &asset.UpdatedAt); err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}
	return assets, nil
}

func (r *postgresMarketDataRepo) UpsertAssets(assets []*domain.Asset) error {
	if len(assets) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(assets))
	valueArgs := make([]interface{}, 0, len(assets)*5)

	for i, asset := range assets {
		n := i * 5
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", n+1, n+2, n+3, n+4, n+5))
		valueArgs = append(valueArgs, asset.Symbol, asset.Name, asset.Type, asset.Exchange, asset.Currency)
	}

	stmt := fmt.Sprintf(`
		INSERT INTO marketdata.assets (symbol, name, type, exchange, currency)
		VALUES %s
		ON CONFLICT (symbol) DO UPDATE SET
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			exchange = EXCLUDED.exchange,
			currency = EXCLUDED.currency,
			updated_at = NOW()
	`, strings.Join(valueStrings, ","))

	_, err := r.db.Exec(stmt, valueArgs...)
	return err
}

func (r *postgresMarketDataRepo) GetAllAssetIDs() (map[string]string, error) {
	rows, err := r.db.Query("SELECT symbol, id FROM marketdata.assets")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assetMap := make(map[string]string)
	for rows.Next() {
		var symbol, id string
		if err := rows.Scan(&symbol, &id); err != nil {
			return nil, err
		}
		assetMap[symbol] = id
	}
	return assetMap, nil
}

func (r *postgresMarketDataRepo) GetLatestPrice(symbol string) (*domain.AssetPrice, error) {
	query := `
		SELECT p.id, p.asset_id, p.price, p.timestamp, p.created_at
		FROM marketdata.asset_prices p
		JOIN marketdata.assets a ON p.asset_id = a.id
		WHERE a.symbol = $1
		ORDER BY p.timestamp DESC
		LIMIT 1
	`
	var price domain.AssetPrice
	err := r.db.QueryRow(query, symbol).Scan(
		&price.ID, &price.AssetID, &price.Price, &price.Timestamp, &price.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &price, nil
}

func (r *postgresMarketDataRepo) GetHistoricalPrices(symbol string, start, end time.Time) ([]*domain.AssetPrice, error) {
	query := `
		SELECT p.id, p.asset_id, p.price, p.timestamp, p.created_at
		FROM marketdata.asset_prices p
		JOIN marketdata.assets a ON p.asset_id = a.id
		WHERE a.symbol = $1 AND p.timestamp >= $2 AND p.timestamp <= $3
		ORDER BY p.timestamp ASC
	`
	rows, err := r.db.Query(query, symbol, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []*domain.AssetPrice
	for rows.Next() {
		var price domain.AssetPrice
		if err := rows.Scan(&price.ID, &price.AssetID, &price.Price, &price.Timestamp, &price.CreatedAt); err != nil {
			return nil, err
		}
		prices = append(prices, &price)
	}
	return prices, nil
}

func (r *postgresMarketDataRepo) InsertPrices(prices []*domain.AssetPrice) error {
	if len(prices) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(prices))
	valueArgs := make([]interface{}, 0, len(prices)*3)

	for i, price := range prices {
		n := i * 3
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", n+1, n+2, n+3))
		valueArgs = append(valueArgs, price.AssetID, price.Price, price.Timestamp)
	}

	stmt := fmt.Sprintf(`
		INSERT INTO marketdata.asset_prices (asset_id, price, timestamp)
		VALUES %s
	`, strings.Join(valueStrings, ","))

	_, err := r.db.Exec(stmt, valueArgs...)
	return err
}
