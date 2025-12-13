package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/garcios/portfolio-insights/services/user-service/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var (
	serverAddr = flag.String("addr", "localhost:50051", "The server address in the format of host:port")
	operation  = flag.String("op", "get", "Operation to perform: get, create, verify, test-errors")
	userID     = flag.String("user-id", "", "User ID for get operation (e.g., '123')")
	email      = flag.String("email", "", "Email for create/verify operations")
	username   = flag.String("username", "", "Username for create operation")
	password   = flag.String("password", "", "Password for create/verify operations")
	verbose    = flag.Bool("verbose", false, "Enable verbose error output")
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

	client := pb.NewUserServiceClient(conn)

	// Execute the requested operation
	switch *operation {
	case "get":
		getUser(client)
	case "create":
		createUser(client)
	case "verify":
		verifyUser(client)
	case "test-errors":
		testErrors(client)
	default:
		log.Fatalf("Unknown operation: %s. Valid operations: get, create, verify, test-errors", *operation)
	}
}

// getUser retrieves a user by their resource name
func getUser(client pb.UserServiceClient) {
	if *userID == "" {
		log.Fatal("user-id is required for get operation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resourceName := fmt.Sprintf("users/%s", *userID)
	req := &pb.GetUserRequest{
		Name: resourceName,
	}

	log.Printf("Getting user: %s", resourceName)
	user, err := client.GetUser(ctx, req)
	if err != nil {
		handleError("GetUser", err)
		return
	}

	printUser(user)
}

// createUser creates a new user
func createUser(client pb.UserServiceClient) {
	if *email == "" || *username == "" || *password == "" {
		log.Fatal("email, username, and password are required for create operation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.CreateUserRequest{
		User: &pb.User{
			Email:    *email,
			Username: *username,
			Password: *password,
		},
	}

	// If user-id is provided, use it
	if *userID != "" {
		req.UserId = *userID
	}

	log.Printf("Creating user: email=%s, username=%s", *email, *username)
	user, err := client.CreateUser(ctx, req)
	if err != nil {
		handleError("CreateUser", err)
		return
	}

	log.Println("✓ User created successfully!")
	printUser(user)
}

// verifyUser verifies user credentials
func verifyUser(client pb.UserServiceClient) {
	if *email == "" || *password == "" {
		log.Fatal("email and password are required for verify operation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.VerifyUserRequest{
		Email:    *email,
		Password: *password,
	}

	log.Printf("Verifying user: email=%s", *email)
	resp, err := client.VerifyUser(ctx, req)
	if err != nil {
		handleError("VerifyUser", err)
		return
	}

	if resp.Valid {
		log.Println("✓ Credentials are VALID")
		if resp.User != nil {
			printUser(resp.User)
		}
	} else {
		log.Println("✗ Credentials are INVALID")
	}
}

// printUser prints user details in a formatted way
func printUser(user *pb.User) {
	fmt.Println("\n=== User Details ===")
	fmt.Printf("Resource Name: %s\n", user.Name)
	fmt.Printf("User ID:       %s\n", user.UserId)
	fmt.Printf("Email:         %s\n", user.Email)
	fmt.Printf("Username:      %s\n", user.Username)
	fmt.Println("===================")
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
func testErrors(client pb.UserServiceClient) {
	fmt.Println("\n🧪 Running Error Test Suite")
	fmt.Println("═══════════════════════════════════════════")

	testsPassed := 0
	testsFailed := 0

	// Test 1: GetUser with empty resource name
	runTest("GetUser with empty resource name", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetUser(ctx, &pb.GetUserRequest{Name: ""})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 2: GetUser with invalid resource name format
	runTest("GetUser with invalid format (missing prefix)", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetUser(ctx, &pb.GetUserRequest{Name: "123"})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 3: GetUser with malformed UUID
	runTest("GetUser with invalid UUID format", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetUser(ctx, &pb.GetUserRequest{Name: "users/invalid-uuid"})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 4: GetUser with non-existent user
	runTest("GetUser with non-existent UUID", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.GetUser(ctx, &pb.GetUserRequest{Name: "users/00000000-0000-0000-0000-000000000000"})
		return err
	}, codes.NotFound, &testsPassed, &testsFailed)

	// Test 5: CreateUser with missing email
	runTest("CreateUser with missing email", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.CreateUser(ctx, &pb.CreateUserRequest{
			User: &pb.User{
				Email:    "",
				Username: "testuser",
				Password: "password123",
			},
		})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 6: CreateUser with missing username
	runTest("CreateUser with missing username", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.CreateUser(ctx, &pb.CreateUserRequest{
			User: &pb.User{
				Email:    "test@example.com",
				Username: "",
				Password: "password123",
			},
		})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 7: CreateUser with missing password
	runTest("CreateUser with missing password", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.CreateUser(ctx, &pb.CreateUserRequest{
			User: &pb.User{
				Email:    "test@example.com",
				Username: "testuser",
				Password: "",
			},
		})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 8: CreateUser with invalid email format
	runTest("CreateUser with invalid email format", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.CreateUser(ctx, &pb.CreateUserRequest{
			User: &pb.User{
				Email:    "not-an-email",
				Username: "testuser",
				Password: "password123",
			},
		})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 9: VerifyUser with empty email
	runTest("VerifyUser with empty email", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.VerifyUser(ctx, &pb.VerifyUserRequest{
			Email:    "",
			Password: "password123",
		})
		return err
	}, codes.InvalidArgument, &testsPassed, &testsFailed)

	// Test 10: VerifyUser with empty password
	runTest("VerifyUser with empty password", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.VerifyUser(ctx, &pb.VerifyUserRequest{
			Email:    "test@example.com",
			Password: "",
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
