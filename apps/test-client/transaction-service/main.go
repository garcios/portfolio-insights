package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/garcios/portfolio-insights/services/transaction-service/transaction"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	serverAddr        = flag.String("addr", "localhost:50053", "The server address in the format of host:port")
	operation         = flag.String("op", "create", "Operation to perform: create, get, list, update, delete, oldest, test-errors")
	userID            = flag.String("user-id", "", "User ID for parent resource name")
	transactionID     = flag.String("transaction-id", "", "Transaction ID for get/update/delete operations")
	txnType           = flag.String("type", "", "Transaction type: BUY, SELL, DEP, WIT, INT, DIV")
	symbol            = flag.String("symbol", "", "Asset symbol (for equity transactions)")
	quantity          = flag.Float64("quantity", 0, "Quantity (for BUY/SELL)")
	price             = flag.Float64("price", 0, "Price per share (for BUY/SELL)")
	amount            = flag.Float64("amount", 0, "Cash amount (for DEP/WIT/INT/DIV)")
	executedAt        = flag.String("executed-at", "", "Execution timestamp (RFC3339 format, e.g., 2024-01-15T10:30:00Z)")
	brokerage         = flag.Float64("brokerage", 0, "Brokerage fee")
	notes             = flag.String("notes", "", "Transaction notes")
	priceCurrency     = flag.String("price-currency", "USD", "Price currency")
	brokerageCurrency = flag.String("brokerage-currency", "USD", "Brokerage currency")
	filterSymbol      = flag.String("filter-symbol", "", "Filter by symbol (for list)")
	filterType        = flag.String("filter-type", "", "Filter by type (for list)")
	pageSize          = flag.Int("page-size", 50, "Page size for list operation")
	pageToken         = flag.String("page-token", "", "Page token for pagination")
	updateFields      = flag.String("update-fields", "", "Comma-separated field paths for update")
	verbose           = flag.Bool("verbose", false, "Enable verbose error output")
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

	client := pb.NewTransactionServiceClient(conn)

	// Execute the requested operation
	switch *operation {
	case "create":
		createTransaction(client)
	case "get":
		getTransaction(client)
	case "list":
		listTransactions(client)
	case "update":
		updateTransaction(client)
	case "delete":
		deleteTransaction(client)
	case "oldest":
		getOldestTransaction(client)
	case "test-errors":
		testErrors(client)
	default:
		log.Fatalf("Unknown operation: %s. Valid operations: create, get, list, update, delete, oldest, test-errors", *operation)
	}
}

// createTransaction creates a new transaction
func createTransaction(client pb.TransactionServiceClient) {
	if *userID == "" {
		log.Fatal("user-id is required for create operation")
	}
	if *txnType == "" {
		log.Fatal("type is required for create operation")
	}
	if *executedAt == "" {
		log.Fatal("executed-at is required for create operation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Parse executed_at timestamp
	execTime, err := time.Parse(time.RFC3339, *executedAt)
	if err != nil {
		log.Fatalf("Invalid executed-at format (use RFC3339, e.g., 2024-01-15T10:30:00Z): %v", err)
	}

	parent := fmt.Sprintf("users/%s", *userID)
	txn := &pb.Transaction{
		Type:              *txnType,
		ExecutedAt:        timestamppb.New(execTime),
		Brokerage:         *brokerage,
		Notes:             *notes,
		PriceCurrency:     *priceCurrency,
		BrokerageCurrency: *brokerageCurrency,
	}

	// Set equity-specific fields if provided
	if *symbol != "" {
		txn.Symbol = symbol
	}
	if *quantity > 0 {
		txn.Quantity = quantity
	}
	if *price > 0 {
		txn.PricePerShare = price
	}

	// Set cash-specific field if provided
	if *amount != 0 {
		txn.Amount = amount
	}

	req := &pb.CreateTransactionRequest{
		Parent:      parent,
		Transaction: txn,
	}

	// Use client-specified transaction ID if provided
	if *transactionID != "" {
		req.TransactionId = *transactionID
	}

	log.Printf("Creating %s transaction for user %s", *txnType, *userID)
	result, err := client.CreateTransaction(ctx, req)
	if err != nil {
		handleError("CreateTransaction", err)
		return
	}

	log.Println("✓ Transaction created successfully!")
	printTransaction(result)
}

// getTransaction retrieves a transaction by its resource name
func getTransaction(client pb.TransactionServiceClient) {
	if *userID == "" || *transactionID == "" {
		log.Fatal("user-id and transaction-id are required for get operation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resourceName := fmt.Sprintf("users/%s/transactions/%s", *userID, *transactionID)
	req := &pb.GetTransactionRequest{
		Name: resourceName,
	}

	log.Printf("Getting transaction: %s", resourceName)
	txn, err := client.GetTransaction(ctx, req)
	if err != nil {
		handleError("GetTransaction", err)
		return
	}

	printTransaction(txn)
}

// listTransactions lists transactions for a user
func listTransactions(client pb.TransactionServiceClient) {
	if *userID == "" {
		log.Fatal("user-id is required for list operation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	parent := fmt.Sprintf("users/%s", *userID)
	req := &pb.ListTransactionsRequest{
		Parent:    parent,
		PageSize:  int32(*pageSize),
		PageToken: *pageToken,
	}

	// Add filters if specified
	if *filterSymbol != "" || *filterType != "" {
		req.Filter = &pb.TransactionFilter{
			Symbol: *filterSymbol,
			Type:   *filterType,
		}
	}

	log.Printf("Listing transactions for user %s", *userID)
	resp, err := client.ListTransactions(ctx, req)
	if err != nil {
		handleError("ListTransactions", err)
		return
	}

	fmt.Printf("\n=== Found %d transactions ===\n", len(resp.Transactions))
	for i, txn := range resp.Transactions {
		fmt.Printf("\n--- Transaction %d ---\n", i+1)
		printTransaction(txn)
	}

	if resp.NextPageToken != "" {
		fmt.Printf("\nNext page token: %s\n", resp.NextPageToken)
		fmt.Println("Use -page-token flag to retrieve next page")
	}
}

// updateTransaction updates an existing transaction
func updateTransaction(client pb.TransactionServiceClient) {
	if *userID == "" || *transactionID == "" {
		log.Fatal("user-id and transaction-id are required for update operation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resourceName := fmt.Sprintf("users/%s/transactions/%s", *userID, *transactionID)
	txn := &pb.Transaction{
		Name: resourceName,
	}

	// Build update based on provided flags
	if *txnType != "" {
		txn.Type = *txnType
	}
	if *symbol != "" {
		txn.Symbol = symbol
	}
	if *quantity > 0 {
		txn.Quantity = quantity
	}
	if *price > 0 {
		txn.PricePerShare = price
	}
	if *amount != 0 {
		txn.Amount = amount
	}
	if *executedAt != "" {
		execTime, err := time.Parse(time.RFC3339, *executedAt)
		if err != nil {
			log.Fatalf("Invalid executed-at format (use RFC3339): %v", err)
		}
		txn.ExecutedAt = timestamppb.New(execTime)
	}
	if *brokerage != 0 {
		txn.Brokerage = *brokerage
	}
	if *notes != "" {
		txn.Notes = *notes
	}
	if *priceCurrency != "USD" {
		txn.PriceCurrency = *priceCurrency
	}
	if *brokerageCurrency != "USD" {
		txn.BrokerageCurrency = *brokerageCurrency
	}

	req := &pb.UpdateTransactionRequest{
		Transaction: txn,
	}

	// Add field mask if specified
	if *updateFields != "" {
		paths := strings.Split(*updateFields, ",")
		for i := range paths {
			paths[i] = strings.TrimSpace(paths[i])
		}
		req.UpdateMask = &fieldmaskpb.FieldMask{Paths: paths}
	}

	log.Printf("Updating transaction: %s", resourceName)
	result, err := client.UpdateTransaction(ctx, req)
	if err != nil {
		handleError("UpdateTransaction", err)
		return
	}

	log.Println("✓ Transaction updated successfully!")
	printTransaction(result)
}

// deleteTransaction deletes a transaction
func deleteTransaction(client pb.TransactionServiceClient) {
	if *userID == "" || *transactionID == "" {
		log.Fatal("user-id and transaction-id are required for delete operation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resourceName := fmt.Sprintf("users/%s/transactions/%s", *userID, *transactionID)
	req := &pb.DeleteTransactionRequest{
		Name: resourceName,
	}

	log.Printf("Deleting transaction: %s", resourceName)
	_, err := client.DeleteTransaction(ctx, req)
	if err != nil {
		handleError("DeleteTransaction", err)
		return
	}

	log.Println("✓ Transaction deleted successfully!")
}

// getOldestTransaction retrieves the oldest transaction for a user
func getOldestTransaction(client pb.TransactionServiceClient) {
	if *userID == "" {
		log.Fatal("user-id is required for oldest operation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	parent := fmt.Sprintf("users/%s", *userID)
	req := &pb.GetOldestTransactionForUserRequest{
		Parent: parent,
	}

	log.Printf("Getting oldest transaction for user %s", *userID)
	txn, err := client.GetOldestTransactionForUser(ctx, req)
	if err != nil {
		handleError("GetOldestTransactionForUser", err)
		return
	}

	log.Println("✓ Found oldest transaction:")
	printTransaction(txn)
}

// printTransaction prints transaction details in a formatted way
func printTransaction(txn *pb.Transaction) {
	fmt.Println("\n=== Transaction Details ===")
	fmt.Printf("Resource Name:     %s\n", txn.Name)
	fmt.Printf("Transaction ID:    %s\n", txn.TransactionId)
	fmt.Printf("User ID:           %s\n", txn.UserId)
	fmt.Printf("Type:              %s\n", txn.Type)

	if txn.Symbol != nil {
		fmt.Printf("Symbol:            %s\n", *txn.Symbol)
	}
	if txn.Quantity != nil {
		fmt.Printf("Quantity:          %.4f\n", *txn.Quantity)
	}
	if txn.PricePerShare != nil {
		fmt.Printf("Price per Share:   %.2f %s\n", *txn.PricePerShare, txn.PriceCurrency)
	}
	if txn.Amount != nil {
		fmt.Printf("Amount:            %.2f %s\n", *txn.Amount, txn.PriceCurrency)
	}

	fmt.Printf("Executed At:       %s\n", txn.ExecutedAt.AsTime().Format(time.RFC3339))
	fmt.Printf("Brokerage:         %.2f %s\n", txn.Brokerage, txn.BrokerageCurrency)

	if txn.Notes != "" {
		fmt.Printf("Notes:             %s\n", txn.Notes)
	}

	fmt.Printf("Created At:        %s\n", txn.CreatedAt.AsTime().Format(time.RFC3339))
	fmt.Printf("Updated At:        %s\n", txn.UpdatedAt.AsTime().Format(time.RFC3339))
	fmt.Println("===========================")
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
func testErrors(client pb.TransactionServiceClient) {
	fmt.Println("\n🧪 Running Error Test Suite")
	fmt.Println("═══════════════════════════════════════════")

	testsPassed := 0
	testsFailed := 0

	execTime := timestamppb.New(time.Now())

	// Test 1: CreateTransaction with empty parent
	runTest("CreateTransaction with empty parent", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.CreateTransaction(ctx, &pb.CreateTransactionRequest{
			Parent: "",
			Transaction: &pb.Transaction{
				Type:              "BUY",
				ExecutedAt:        execTime,
				PriceCurrency:     "USD",
				BrokerageCurrency: "USD",
			},
		})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 2: CreateTransaction with invalid parent format
	runTest("CreateTransaction with invalid parent format", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.CreateTransaction(ctx, &pb.CreateTransactionRequest{
			Parent: "invalid-format",
			Transaction: &pb.Transaction{
				Type:              "BUY",
				ExecutedAt:        execTime,
				PriceCurrency:     "USD",
				BrokerageCurrency: "USD",
			},
		})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 3: CreateTransaction with missing type
	runTest("CreateTransaction with missing type", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.CreateTransaction(ctx, &pb.CreateTransactionRequest{
			Parent: "users/test-user",
			Transaction: &pb.Transaction{
				Type:              "",
				ExecutedAt:        execTime,
				PriceCurrency:     "USD",
				BrokerageCurrency: "USD",
			},
		})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 4: CreateTransaction with missing executed_at
	runTest("CreateTransaction with missing executed_at", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.CreateTransaction(ctx, &pb.CreateTransactionRequest{
			Parent: "users/test-user",
			Transaction: &pb.Transaction{
				Type:              "BUY",
				ExecutedAt:        nil,
				PriceCurrency:     "USD",
				BrokerageCurrency: "USD",
			},
		})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 5: CreateTransaction with nil transaction
	runTest("CreateTransaction with nil transaction", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.CreateTransaction(ctx, &pb.CreateTransactionRequest{
			Parent:      "users/test-user",
			Transaction: nil,
		})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 6: GetTransaction with empty name
	runTest("GetTransaction with empty name", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetTransaction(ctx, &pb.GetTransactionRequest{Name: ""})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 7: GetTransaction with invalid name format
	runTest("GetTransaction with invalid format (missing prefix)", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetTransaction(ctx, &pb.GetTransactionRequest{Name: "invalid-format"})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 8: GetTransaction with non-existent transaction
	runTest("GetTransaction with non-existent transaction", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetTransaction(ctx, &pb.GetTransactionRequest{
			Name: "users/test-user/transactions/00000000-0000-0000-0000-000000000000",
		})
		return err
	}, codes.NotFound, &testsPassed, &testsFailed)

	// Test 9: ListTransactions with empty parent
	runTest("ListTransactions with empty parent", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.ListTransactions(ctx, &pb.ListTransactionsRequest{Parent: ""})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 10: ListTransactions with invalid parent format
	runTest("ListTransactions with invalid parent format", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.ListTransactions(ctx, &pb.ListTransactionsRequest{Parent: "invalid"})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 11: UpdateTransaction with nil transaction
	runTest("UpdateTransaction with nil transaction", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.UpdateTransaction(ctx, &pb.UpdateTransactionRequest{
			Transaction: nil,
		})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 12: UpdateTransaction with invalid resource name
	runTest("UpdateTransaction with invalid resource name", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.UpdateTransaction(ctx, &pb.UpdateTransactionRequest{
			Transaction: &pb.Transaction{
				Name: "invalid-format",
			},
		})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 13: DeleteTransaction with empty name
	runTest("DeleteTransaction with empty name", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.DeleteTransaction(ctx, &pb.DeleteTransactionRequest{Name: ""})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 14: DeleteTransaction with non-existent transaction
	runTest("DeleteTransaction with non-existent transaction", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.DeleteTransaction(ctx, &pb.DeleteTransactionRequest{
			Name: "users/test-user/transactions/00000000-0000-0000-0000-000000000000",
		})
		return err
	}, codes.NotFound, &testsPassed, &testsFailed)

	// Test 15: GetOldestTransactionForUser with empty parent
	runTest("GetOldestTransactionForUser with empty parent", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetOldestTransactionForUser(ctx, &pb.GetOldestTransactionForUserRequest{
			Parent: "",
		})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 16: GetOldestTransactionForUser with invalid parent
	runTest("GetOldestTransactionForUser with invalid parent", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetOldestTransactionForUser(ctx, &pb.GetOldestTransactionForUserRequest{
			Parent: "invalid",
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
