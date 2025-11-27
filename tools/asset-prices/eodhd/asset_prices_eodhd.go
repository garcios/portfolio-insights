package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Price struct {
	Ticker   string
	Date     string
	Open     string
	High     string
	Low      string
	Close    string
	AdjClose string
	Volume   string
}

func main() {
	tickersFile := flag.String("tickers", "tickers.txt", "File with tickers, one per line")
	fromDateStr := flag.String("from", "20000101", "Start date, YYYYMMDD")
	toDateStr := flag.String("to", time.Now().Format("20060102"), "End date, YYYYMMDD")
	outFile := flag.String("output", "all_prices.csv", "Output CSV file")
	apiKey := flag.String("token", "", "EODHD API Token")
	flag.Parse()

	fromDate, err := time.Parse("20060102", *fromDateStr)
	if err != nil {
		panic(fmt.Errorf("invalid -from date: %w", err))
	}

	toDate, err := time.Parse("20060102", *toDateStr)
	if err != nil {
		panic(fmt.Errorf("invalid -to date: %w", err))
	}

	token := *apiKey
	if token == "" {
		token = os.Getenv("EODHD_API_TOKEN")
	}
	if token == "" {
		fmt.Println("Error: API token is required. Use -token flag or EODHD_API_TOKEN env var.")
		os.Exit(1)
	}

	tickers, err := readTickers(*tickersFile)
	if err != nil {
		panic(err)
	}

	var allPrices []Price
	isForex := false

	for i, tk := range tickers {
		if i > 0 {
			// Sleep to be safe with rate limits
			time.Sleep(1 * time.Second)
		}

		// Check if this is forex data
		if strings.Contains(tk, ".FOREX") {
			isForex = true
		}

		fmt.Printf("Fetching %s\n", tk)
		prices, err := fetchPrices(tk, token, fromDate, toDate)
		if err != nil {
			fmt.Printf("Error fetching %s: %v\n", tk, err)
			continue
		}
		allPrices = append(allPrices, prices...)
	}

	// Deduplicate (ticker + date)
	unique := dedupe(allPrices)

	// Sort by date, then ticker
	sort.Slice(unique, func(i, j int) bool {
		if unique[i].Date == unique[j].Date {
			return unique[i].Ticker < unique[j].Ticker
		}
		return unique[i].Date < unique[j].Date
	})

	// Write output using appropriate format
	if isForex {
		if err := writeCurrencyRatesCSV(*outFile, unique, "USD"); err != nil {
			panic(err)
		}
	} else {
		if err := writeCSV(*outFile, unique); err != nil {
			panic(err)
		}
	}
	fmt.Println("Saved", *outFile)
}

func fetchPrices(ticker, token string, fromDate, toDate time.Time) ([]Price, error) {
	// EODHD API URL
	// https://eodhd.com/api/eod/<ticker>?from=<YYYY-MM-DD>&to=<YYYY-MM-DD>&period=d&api_token=<token>&fmt=csv

	startDateStr := fromDate.Format("2006-01-02")
	endDateStr := toDate.Format("2006-01-02")

	// Build query parameters
	params := map[string]string{
		"from":      startDateStr,
		"to":        endDateStr,
		"api_token": token,
		"fmt":       "csv",
	}

	// FOREX endpoints use 'order' instead of 'period'
	if strings.Contains(ticker, ".FOREX") {
		params["order"] = "d"
	} else {
		params["period"] = "d"
	}

	// Build query string
	queryParams := ""
	for k, v := range params {
		if queryParams != "" {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("%s=%s", k, v)
	}

	// Note: EODHD tickers often need an exchange suffix (e.g. IVV.AU).
	// We assume the input tickers already have the suffix if needed, or the user is querying US stocks (default).
	url := fmt.Sprintf("https://eodhd.com/api/eod/%s?%s", ticker, queryParams)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	// Read CSV response
	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		// No data or just header
		return []Price{}, nil
	}

	// EODHD CSV Header: Date,Open,High,Low,Close,Adjusted_close,Volume
	headers := records[0]
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[h] = i
	}

	// Helper to get value safely
	getVal := func(row []string, col string) string {
		idx, ok := headerMap[col]
		if !ok || idx >= len(row) {
			return ""
		}
		return row[idx]
	}

	var prices []Price
	for _, row := range records[1:] {
		cleanTicker := strings.TrimSuffix(ticker, ".AU")
		cleanTicker = strings.TrimSuffix(cleanTicker, ".FOREX")
		prices = append(prices, Price{
			Ticker:   cleanTicker,
			Date:     getVal(row, "Date"),
			Open:     getVal(row, "Open"),
			High:     getVal(row, "High"),
			Low:      getVal(row, "Low"),
			Close:    getVal(row, "Close"),
			AdjClose: getVal(row, "Adjusted_close"),
			Volume:   getVal(row, "Volume"),
		})
	}

	return prices, nil
}

// readTickers reads tickers from file (one per line)
func readTickers(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Support newline splitting
	parts := strings.Split(string(data), "\n")
	out := []string{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out, nil
}

func dedupe(prices []Price) []Price {
	m := make(map[string]Price, len(prices))
	for _, p := range prices {
		key := p.Ticker + "__" + p.Date
		m[key] = p
	}
	out := make([]Price, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

func writeCSV(path string, prices []Price) error {
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// header
	w.Write([]string{"symbol", "price", "timestamp"})

	for _, p := range prices {
		w.Write([]string{
			p.Ticker,
			p.AdjClose,
			p.Date,
		})
	}
	return nil
}

func writeCurrencyRatesCSV(path string, prices []Price, baseCurrency string) error {
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// header: base_currency,target_currency,rate,rate_date
	w.Write([]string{"base_currency", "target_currency", "rate", "rate_date"})

	for _, p := range prices {
		// For forex data, the ticker is the base currency (e.g., "AUD")
		// The Close price represents the exchange rate
		w.Write([]string{
			baseCurrency, // base_currency (e.g., "USD")
			p.Ticker,     // target_currency (e.g., "AUD")
			p.Close,      // rate (exchange rate)
			p.Date,       // rate_date
		})
	}
	return nil
}
