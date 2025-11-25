package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/domain"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/metrics"
)

type postgresMarketDataRepo struct {
	db *sql.DB
}

func NewPostgresMarketDataRepository(db *sql.DB) domain.MarketDataRepository {
	return &postgresMarketDataRepo{db: db}
}

func (r *postgresMarketDataRepo) GetAssetBySymbol(symbol string) (*domain.Asset, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("get_asset_by_symbol", "assets", time.Since(start).Seconds(), nil)
	}()

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
		metrics.RecordDatabaseQuery("get_asset_by_symbol", "assets", time.Since(start).Seconds(), err)
		return nil, err
	}
	return &asset, nil
}

func (r *postgresMarketDataRepo) ListAssets(limit, offset int) ([]*domain.Asset, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("list_assets", "assets", time.Since(start).Seconds(), nil)
	}()

	query := `
		SELECT id, symbol, name, type, exchange, currency, created_at, updated_at
		FROM marketdata.assets
		ORDER BY symbol
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		metrics.RecordDatabaseQuery("list_assets", "assets", time.Since(start).Seconds(), err)
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
	start := time.Now()
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
	metrics.RecordDatabaseQuery("upsert_assets", "assets", time.Since(start).Seconds(), err)
	return err
}

func (r *postgresMarketDataRepo) GetAllAssetIDs() (map[string]string, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("get_all_asset_ids", "assets", time.Since(start).Seconds(), nil)
	}()

	rows, err := r.db.Query("SELECT symbol, id FROM marketdata.assets")
	if err != nil {
		metrics.RecordDatabaseQuery("get_all_asset_ids", "assets", time.Since(start).Seconds(), err)
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

func (r *postgresMarketDataRepo) CountAssets() (int, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("count_assets", "assets", time.Since(start).Seconds(), nil)
	}()

	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM marketdata.assets").Scan(&count)
	if err != nil {
		metrics.RecordDatabaseQuery("count_assets", "assets", time.Since(start).Seconds(), err)
		return 0, err
	}
	return count, nil
}

func (r *postgresMarketDataRepo) GetLatestPrice(symbol string) (*domain.AssetPrice, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("get_latest_price", "asset_prices", time.Since(start).Seconds(), nil)
	}()

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
		metrics.RecordDatabaseQuery("get_latest_price", "asset_prices", time.Since(start).Seconds(), err)
		return nil, err
	}
	return &price, nil
}

func (r *postgresMarketDataRepo) GetLatestPrices(symbols []string) (map[string]*domain.AssetPrice, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("get_latest_prices_batch", "asset_prices", time.Since(start).Seconds(), nil)
	}()

	if len(symbols) == 0 {
		return make(map[string]*domain.AssetPrice), nil
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(symbols))
	args := make([]interface{}, len(symbols))
	for i, symbol := range symbols {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = symbol
	}

	query := fmt.Sprintf(`
		SELECT a.symbol, p.id, p.asset_id, p.price, p.timestamp, p.created_at
		FROM marketdata.asset_prices p
		JOIN marketdata.assets a ON p.asset_id = a.id
		WHERE a.symbol IN (%s)
		AND p.id IN (
			SELECT p2.id
			FROM marketdata.asset_prices p2
			JOIN marketdata.assets a2 ON p2.asset_id = a2.id
			WHERE a2.symbol = a.symbol
			ORDER BY p2.timestamp DESC
			LIMIT 1
		)
	`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		metrics.RecordDatabaseQuery("get_latest_prices_batch", "asset_prices", time.Since(start).Seconds(), err)
		return nil, err
	}
	defer rows.Close()

	prices := make(map[string]*domain.AssetPrice)
	for rows.Next() {
		var symbol string
		var price domain.AssetPrice
		if err := rows.Scan(&symbol, &price.ID, &price.AssetID, &price.Price, &price.Timestamp, &price.CreatedAt); err != nil {
			return nil, err
		}
		prices[symbol] = &price
	}

	return prices, nil
}

func (r *postgresMarketDataRepo) GetHistoricalPrices(symbol string, start, end time.Time) ([]*domain.AssetPrice, error) {
	startTime := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("get_historical_prices", "asset_prices", time.Since(startTime).Seconds(), nil)
	}()

	query := `
		SELECT p.id, p.asset_id, p.price, p.timestamp, p.created_at
		FROM marketdata.asset_prices p
		JOIN marketdata.assets a ON p.asset_id = a.id
		WHERE a.symbol = $1 AND p.timestamp >= $2 AND p.timestamp <= $3
		ORDER BY p.timestamp ASC
	`
	rows, err := r.db.Query(query, symbol, start, end)
	if err != nil {
		metrics.RecordDatabaseQuery("get_historical_prices", "asset_prices", time.Since(startTime).Seconds(), err)
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
	start := time.Now()
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
		ON CONFLICT (asset_id, timestamp) DO UPDATE SET
			price = EXCLUDED.price
	`, strings.Join(valueStrings, ","))

	_, err := r.db.Exec(stmt, valueArgs...)
	metrics.RecordDatabaseQuery("insert_prices", "asset_prices", time.Since(start).Seconds(), err)
	return err
}

func (r *postgresMarketDataRepo) CountPrices() (int, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("count_prices", "asset_prices", time.Since(start).Seconds(), nil)
	}()

	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM marketdata.asset_prices").Scan(&count)
	if err != nil {
		metrics.RecordDatabaseQuery("count_prices", "asset_prices", time.Since(start).Seconds(), err)
		return 0, err
	}
	return count, nil
}

func (r *postgresMarketDataRepo) InsertCurrencyRates(rates []*domain.CurrencyRate) error {
	start := time.Now()
	if len(rates) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(rates))
	valueArgs := make([]interface{}, 0, len(rates)*4)

	for i, rate := range rates {
		n := i * 4
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", n+1, n+2, n+3, n+4))
		valueArgs = append(valueArgs, rate.BaseCurrency, rate.TargetCurrency, rate.Rate, rate.RateDate)
	}

	stmt := fmt.Sprintf(`
		INSERT INTO marketdata.currency_rates (base_currency, target_currency, rate, rate_date)
		VALUES %s
		ON CONFLICT (base_currency, target_currency, rate_date) DO UPDATE SET
			rate = EXCLUDED.rate
	`, strings.Join(valueStrings, ","))

	_, err := r.db.Exec(stmt, valueArgs...)
	metrics.RecordDatabaseQuery("insert_currency_rates", "currency_rates", time.Since(start).Seconds(), err)
	return err
}

func (r *postgresMarketDataRepo) GetLatestCurrencyRate(baseCurrency, targetCurrency string) (*domain.CurrencyRate, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("get_latest_currency_rate", "currency_rates", time.Since(start).Seconds(), nil)
	}()

	query := `
		SELECT id, base_currency, target_currency, rate, rate_date, created_at
		FROM marketdata.currency_rates
		WHERE base_currency = $1 AND target_currency = $2
		ORDER BY rate_date DESC
		LIMIT 1
	`
	var rate domain.CurrencyRate
	err := r.db.QueryRow(query, baseCurrency, targetCurrency).Scan(
		&rate.ID, &rate.BaseCurrency, &rate.TargetCurrency, &rate.Rate, &rate.RateDate, &rate.CreatedAt,
	)
	if err != nil {
		metrics.RecordDatabaseQuery("get_latest_currency_rate", "currency_rates", time.Since(start).Seconds(), err)
		return nil, err
	}
	return &rate, nil
}
