download-prices:
	go run eodhd/asset_prices_eodhd.go \
    -tickers=tickers.txt \
    -from=20251126 \
	-to=20251126 \
    -output=prices.csv

download-currency-rates:
	go run eodhd/asset_prices_eodhd.go \
    -tickers=forex.txt \
    -from=20230101 \
	-to=20251127 \
    -output=currency_rates.csv
