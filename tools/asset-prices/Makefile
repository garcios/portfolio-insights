download-prices:
	go run eodhd/asset_prices_eodhd.go \
    -tickers=tickers.txt \
    -from=20150101 \
	-to=20251216 \
    -output=all_prices.csv

download-currency-rates:
	go run eodhd/asset_prices_eodhd.go \
    -tickers=forex.txt \
    -from=20150101 \
	-to=20251216 \
    -output=currency_rates.csv
