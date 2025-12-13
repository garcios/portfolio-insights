package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/garcios/portfolio-insights/services/marketdata-service/marketdata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	serverAddr     = flag.String("addr", "localhost:50054", "The server address in the format of host:port")
	operation      = flag.String("op", "get-asset-by-symbol", "Operation to perform: get-asset, get-asset-by-symbol, list-assets, get-price, get-prices, get-historical, get-currency-rate, test-errors")
	assetName      = flag.String("asset-name", "", "Asset resource name (e.g., 'assets/AAPL')")
	symbol         = flag.String("symbol", "", "Asset symbol (e.g., 'AAPL')")
	assetNames     = flag.String("asset-names", "", "Comma-separated list of asset resource names for batch operations")
	pageSize       = flag.Int("page-size", 50, "Page size for list operations")
	pageToken      = flag.String("page-token", "", "Page token for pagination")
	startTime      = flag.String("start-time", "", "Start time for historical data (RFC3339 format)")
	endTime        = flag.String("end-time", "", "End time for historical data (RFC3339 format)")
	interval       = flag.String("interval", "1d", "Interval for historical data (e.g., '1d', '1h')")
	baseCurrency   = flag.String("base-currency", "", "Base currency code (e.g., 'USD')")
	targetCurrency = flag.String("target-currency", "", "Target currency code (e.g., 'EUR')")
	verbose        = flag.Bool("verbose", false, "Enable verbose error output")
)

func main() {
	flag.Parse()

	// Set up a connection to the server
	conn, err := grpc.NewClient(*serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewMarketDataServiceClient(conn)

	// Execute the requested operation
	switch *operation {
	case "get-asset":
		getAsset(client)
	case "get-asset-by-symbol":
		getAssetBySymbol(client)
	case "list-assets":
		listAssets(client)
	case "get-price":
		getLatestPrice(client)
	case "get-prices":
		getLatestPrices(client)
	case "get-historical":
		getHistoricalPrices(client)
	case "get-currency-rate":
		getLatestCurrencyRate(client)
	case "test-errors":
		testErrors(client)
	default:
		log.Fatalf("Unknown operation: %s. Valid operations: get-asset, get-asset-by-symbol, list-assets, get-price, get-prices, get-historical, get-currency-rate, test-errors", *operation)
	}
}

// getAsset retrieves an asset by its resource name
func getAsset(client pb.MarketDataServiceClient) {
	if *assetName == "" {
		log.Fatal("asset-name is required for get-asset operation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.GetAssetRequest{
		Name: *assetName,
	}

	log.Printf("Getting asset: %s", *assetName)
	asset, err := client.GetAsset(ctx, req)
	if err != nil {
		handleError("GetAsset", err)
		return
	}

	printAsset(asset)
}

// getAssetBySymbol retrieves an asset by its symbol
func getAssetBySymbol(client pb.MarketDataServiceClient) {
	if *symbol == "" {
		log.Fatal("symbol is required for get-asset-by-symbol operation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.GetAssetBySymbolRequest{
		Symbol: *symbol,
	}

	log.Printf("Getting asset by symbol: %s", *symbol)
	asset, err := client.GetAssetBySymbol(ctx, req)
	if err != nil {
		handleError("GetAssetBySymbol", err)
		return
	}

	printAsset(asset)
}

// listAssets retrieves a paginated list of assets
func listAssets(client pb.MarketDataServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.ListAssetsRequest{
		PageSize:  int32(*pageSize),
		PageToken: *pageToken,
	}

	log.Printf("Listing assets (page_size=%d)", *pageSize)
	resp, err := client.ListAssets(ctx, req)
	if err != nil {
		handleError("ListAssets", err)
		return
	}

	fmt.Printf("\n=== Assets (showing %d) ===\n", len(resp.Assets))
	for i, asset := range resp.Assets {
		fmt.Printf("\n[%d] %s\n", i+1, asset.Symbol)
		fmt.Printf("    Resource Name: %s\n", asset.Name)
		fmt.Printf("    Display Name:  %s\n", asset.DisplayName)
		fmt.Printf("    Type:          %s\n", asset.Type)
		fmt.Printf("    Exchange:      %s\n", asset.Exchange)
		fmt.Printf("    Currency:      %s\n", asset.Currency)
	}

	if resp.NextPageToken != "" {
		fmt.Printf("\nNext Page Token: %s\n", resp.NextPageToken)
		fmt.Printf("To get next page, use: -page-token \"%s\"\n", resp.NextPageToken)
	}
	fmt.Println("===================")
}

// getLatestPrice retrieves the latest price for an asset
func getLatestPrice(client pb.MarketDataServiceClient) {
	if *assetName == "" {
		log.Fatal("asset-name is required for get-price operation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.GetLatestPriceRequest{
		Name: *assetName,
	}

	log.Printf("Getting latest price for: %s", *assetName)
	price, err := client.GetLatestPrice(ctx, req)
	if err != nil {
		handleError("GetLatestPrice", err)
		return
	}

	printPrice(price)
}

// getLatestPrices retrieves the latest prices for multiple assets
func getLatestPrices(client pb.MarketDataServiceClient) {
	if *assetNames == "" {
		log.Fatal("asset-names is required for get-prices operation (comma-separated)")
	}

	names := strings.Split(*assetNames, ",")
	for i := range names {
		names[i] = strings.TrimSpace(names[i])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.GetLatestPricesRequest{
		Names: names,
	}

	log.Printf("Getting latest prices for %d assets", len(names))
	resp, err := client.GetLatestPrices(ctx, req)
	if err != nil {
		handleError("GetLatestPrices", err)
		return
	}

	fmt.Printf("\n=== Latest Prices (%d assets) ===\n", len(resp.Prices))
	for assetName, price := range resp.Prices {
		fmt.Printf("\n%s:\n", assetName)
		fmt.Printf("  Price:     $%.2f\n", price.Price)
		if price.Timestamp != nil {
			fmt.Printf("  Timestamp: %s\n", price.Timestamp.AsTime().Format(time.RFC3339))
		}
	}
	fmt.Println("===================")
}

// getHistoricalPrices retrieves historical prices for an asset
func getHistoricalPrices(client pb.MarketDataServiceClient) {
	if *assetName == "" {
		log.Fatal("asset-name is required for get-historical operation")
	}
	if *startTime == "" || *endTime == "" {
		log.Fatal("start-time and end-time are required for get-historical operation")
	}

	start, err := time.Parse(time.RFC3339, *startTime)
	if err != nil {
		log.Fatalf("Invalid start-time format (use RFC3339): %v", err)
	}

	end, err := time.Parse(time.RFC3339, *endTime)
	if err != nil {
		log.Fatalf("Invalid end-time format (use RFC3339): %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &pb.GetHistoricalPricesRequest{
		Name:      *assetName,
		StartTime: timestamppb.New(start),
		EndTime:   timestamppb.New(end),
		Interval:  *interval,
	}

	log.Printf("Getting historical prices for: %s", *assetName)
	log.Printf("Time range: %s to %s", *startTime, *endTime)
	resp, err := client.GetHistoricalPrices(ctx, req)
	if err != nil {
		handleError("GetHistoricalPrices", err)
		return
	}

	fmt.Printf("\n=== Historical Prices (%d data points) ===\n", len(resp.Prices))
	for i, price := range resp.Prices {
		timestamp := "N/A"
		if price.Timestamp != nil {
			timestamp = price.Timestamp.AsTime().Format(time.RFC3339)
		}
		fmt.Printf("[%d] %s: $%.2f\n", i+1, timestamp, price.Price)
	}
	fmt.Println("===================")
}

// getLatestCurrencyRate retrieves the latest currency exchange rate
func getLatestCurrencyRate(client pb.MarketDataServiceClient) {
	if *baseCurrency == "" || *targetCurrency == "" {
		log.Fatal("base-currency and target-currency are required for get-currency-rate operation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.GetLatestCurrencyRateRequest{
		BaseCurrency:   *baseCurrency,
		TargetCurrency: *targetCurrency,
	}

	log.Printf("Getting currency rate: %s -> %s", *baseCurrency, *targetCurrency)
	rate, err := client.GetLatestCurrencyRate(ctx, req)
	if err != nil {
		handleError("GetLatestCurrencyRate", err)
		return
	}

	printCurrencyRate(rate)
}

// printAsset prints asset details in a formatted way
func printAsset(asset *pb.Asset) {
	fmt.Println("\n=== Asset Details ===")
	fmt.Printf("Resource Name: %s\n", asset.Name)
	fmt.Printf("Asset ID:      %s\n", asset.AssetId)
	fmt.Printf("Symbol:        %s\n", asset.Symbol)
	fmt.Printf("Display Name:  %s\n", asset.DisplayName)
	fmt.Printf("Type:          %s\n", asset.Type)
	fmt.Printf("Exchange:      %s\n", asset.Exchange)
	fmt.Printf("Currency:      %s\n", asset.Currency)
	fmt.Println("=====================")
}

// printPrice prints price details in a formatted way
func printPrice(price *pb.AssetPrice) {
	fmt.Println("\n=== Asset Price ===")
	fmt.Printf("Asset:     %s\n", price.AssetName)
	fmt.Printf("Price:     $%.2f\n", price.Price)
	if price.Timestamp != nil {
		fmt.Printf("Timestamp: %s\n", price.Timestamp.AsTime().Format(time.RFC3339))
	}
	fmt.Println("===================")
}

// printCurrencyRate prints currency rate details in a formatted way
func printCurrencyRate(rate *pb.CurrencyRate) {
	fmt.Println("\n=== Currency Rate ===")
	fmt.Printf("Resource Name:   %s\n", rate.Name)
	fmt.Printf("Base Currency:   %s\n", rate.BaseCurrency)
	fmt.Printf("Target Currency: %s\n", rate.TargetCurrency)
	fmt.Printf("Rate:            %.6f\n", rate.Rate)
	if rate.RateDate != nil {
		fmt.Printf("Rate Date:       %s\n", rate.RateDate.AsTime().Format(time.RFC3339))
	}
	fmt.Println("=====================")
}

// handleError provides detailed error information from gRPC calls
func handleError(operation string, err error) {
	fmt.Fprintf(os.Stderr, "\n❌ %s failed\n", operation)
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Extract gRPC status
	st, ok := status.FromError(err)
	if ok {
		fmt.Fprintf(os.Stderr, "Status Code:  %s (%d)\n", st.Code(), st.Code())
		fmt.Fprintf(os.Stderr, "Message:      %s\n", st.Message())

		if *verbose && len(st.Details()) > 0 {
			fmt.Fprintf(os.Stderr, "\nDetails:\n")
			for i, detail := range st.Details() {
				fmt.Fprintf(os.Stderr, "  [%d] %v\n", i+1, detail)
			}
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error:        %v\n", err)
	}

	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	os.Exit(1)
}

// testErrors runs a comprehensive suite of error tests
func testErrors(client pb.MarketDataServiceClient) {
	fmt.Println("\n🧪 Running Error Test Suite")
	fmt.Println("═══════════════════════════════════════════")

	testsPassed := 0
	testsFailed := 0

	// Test 1: GetAsset with empty resource name
	runTest("GetAsset with empty resource name", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetAsset(ctx, &pb.GetAssetRequest{Name: ""})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 2: GetAsset with invalid resource name format
	runTest("GetAsset with invalid format (missing prefix)", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetAsset(ctx, &pb.GetAssetRequest{Name: "AAPL"})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 3: GetAsset with non-existent asset
	runTest("GetAsset with non-existent asset", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetAsset(ctx, &pb.GetAssetRequest{Name: "assets/NONEXISTENT"})
		return err
	}, codes.NotFound, &testsPassed, &testsFailed)

	// Test 4: GetAssetBySymbol with empty symbol
	runTest("GetAssetBySymbol with empty symbol", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetAssetBySymbol(ctx, &pb.GetAssetBySymbolRequest{Symbol: ""})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 5: GetAssetBySymbol with non-existent symbol
	runTest("GetAssetBySymbol with non-existent symbol", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetAssetBySymbol(ctx, &pb.GetAssetBySymbolRequest{Symbol: "NONEXISTENT123"})
		return err
	}, codes.NotFound, &testsPassed, &testsFailed)

	// Test 6: GetLatestPrice with empty resource name
	runTest("GetLatestPrice with empty resource name", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetLatestPrice(ctx, &pb.GetLatestPriceRequest{Name: ""})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 7: GetLatestPrice with invalid resource name format
	runTest("GetLatestPrice with invalid format", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetLatestPrice(ctx, &pb.GetLatestPriceRequest{Name: "AAPL"})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 8: GetLatestPrices with empty names list
	runTest("GetLatestPrices with empty names list", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetLatestPrices(ctx, &pb.GetLatestPricesRequest{Names: []string{}})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 9: GetHistoricalPrices with empty resource name
	runTest("GetHistoricalPrices with empty resource name", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		now := time.Now()
		_, err := client.GetHistoricalPrices(ctx, &pb.GetHistoricalPricesRequest{
			Name:      "",
			StartTime: timestamppb.New(now.Add(-24 * time.Hour)),
			EndTime:   timestamppb.New(now),
		})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 10: GetHistoricalPrices with end time before start time
	runTest("GetHistoricalPrices with invalid time range", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		now := time.Now()
		_, err := client.GetHistoricalPrices(ctx, &pb.GetHistoricalPricesRequest{
			Name:      "assets/AAPL",
			StartTime: timestamppb.New(now),
			EndTime:   timestamppb.New(now.Add(-24 * time.Hour)),
		})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 11: GetLatestCurrencyRate with empty base currency
	runTest("GetLatestCurrencyRate with empty base currency", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetLatestCurrencyRate(ctx, &pb.GetLatestCurrencyRateRequest{
			BaseCurrency:   "",
			TargetCurrency: "EUR",
		})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 12: GetLatestCurrencyRate with empty target currency
	runTest("GetLatestCurrencyRate with empty target currency", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetLatestCurrencyRate(ctx, &pb.GetLatestCurrencyRateRequest{
			BaseCurrency:   "USD",
			TargetCurrency: "",
		})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Print summary
	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Printf("Test Summary: %d passed, %d failed\n", testsPassed, testsFailed)
	fmt.Println("═══════════════════════════════════════════")

	if testsFailed > 0 {
		os.Exit(1)
	}
}

// runTest executes a single error test case
func runTest(name string, testFunc func() error, expectedCode codes.Code, passed *int, failed *int) {
	fmt.Printf("Testing: %s\n", name)
	err := testFunc()

	if err == nil {
		fmt.Printf("  ❌ FAIL: Expected error with code %s, but got no error\n\n", expectedCode)
		*failed++
		return
	}

	st, ok := status.FromError(err)
	if !ok {
		fmt.Printf("  ❌ FAIL: Error is not a gRPC status error: %v\n\n", err)
		*failed++
		return
	}

	if st.Code() == expectedCode {
		fmt.Printf("  ✓ PASS: Got expected error code %s\n", st.Code())
		fmt.Printf("  Message: %s\n\n", st.Message())
		*passed++
	} else {
		fmt.Printf("  ❌ FAIL: Expected code %s, got %s\n", expectedCode, st.Code())
		fmt.Printf("  Message: %s\n\n", st.Message())
		*failed++
	}
}
