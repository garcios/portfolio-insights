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
	apiKey := flag.String("token", "", "Tiingo API Token")
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
		token = os.Getenv("TIINGO_API_TOKEN")
	}
	if token == "" {
		fmt.Println("Error: API token is required. Use -token flag or TIINGO_API_TOKEN env var.")
		os.Exit(1)
	}

	tickers, err := readTickers(*tickersFile)
	if err != nil {
		panic(err)
	}

	var allPrices []Price

	for i, tk := range tickers {
		if i > 0 {
			// Tiingo rate limits are generous (500/hour, 50/min?), but let's be safe.
			// Sleep 1s between requests.
			time.Sleep(1 * time.Second)
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

	if err := writeCSV(*outFile, unique); err != nil {
		panic(err)
	}
	fmt.Println("Saved", *outFile)
}

func fetchPrices(ticker, token string, fromDate, toDate time.Time) ([]Price, error) {
	// Tiingo API URL
	// https://api.tiingo.com/tiingo/daily/<ticker>/prices?startDate=...&endDate=...&format=csv&token=...

	startDateStr := fromDate.Format("2006-01-02")
	endDateStr := toDate.Format("2006-01-02")
	url := fmt.Sprintf("https://api.tiingo.com/tiingo/daily/%s/prices?startDate=%s&endDate=%s&format=csv&token=%s", ticker, startDateStr, endDateStr, token)

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

	// Header: date,close,high,low,open,volume,adjClose,adjHigh,adjLow,adjOpen,adjVolume,divCash,splitFactor
	// We need to map these to our Price struct.
	// Let's find indices dynamically to be safe, or assume order.
	// Tiingo CSV documentation says: date, close, high, low, open, volume, adjClose, adjHigh, adjLow, adjOpen, adjVolume, divCash, splitFactor

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
		prices = append(prices, Price{
			Ticker:   ticker,
			Date:     getVal(row, "date"),
			Open:     getVal(row, "open"),
			High:     getVal(row, "high"),
			Low:      getVal(row, "low"),
			Close:    getVal(row, "close"),
			AdjClose: getVal(row, "adjClose"),
			Volume:   getVal(row, "volume"),
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
