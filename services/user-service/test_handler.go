// Package main provides a standalone test program to verify user handler logic.
package main

import (
	"context"
	"fmt"

	"github.com/garcios/portfolio-insights/pkg/resourcenames"
	"github.com/garcios/portfolio-insights/services/user-service/internal/domain"
	grpcHandler "github.com/garcios/portfolio-insights/services/user-service/internal/handler/grpc"
	pb "github.com/garcios/portfolio-insights/services/user-service/user"
)

// MockUserUsecase for testing
type MockUserUsecase struct{}

// GetUser retrieves a user by ID for testing purposes.
func (m *MockUserUsecase) GetUser(id string) (*domain.User, error) {
	return &domain.User{
		ID:       id,
		Email:    "test@example.com",
		Username: "testuser",
	}, nil
}

// CreateUser creates a new user for testing purposes.
func (m *MockUserUsecase) CreateUser(user *domain.User, password string) (*domain.User, error) {
	return &domain.User{
		ID:       "test-uuid-12345",
		Email:    user.Email,
		Username: user.Username,
	}, nil
}

// VerifyUser verifies user credentials for testing purposes.
func (m *MockUserUsecase) VerifyUser(email, password string) (*domain.User, error) {
	return &domain.User{
		ID:       "test-uuid-12345",
		Email:    email,
		Username: "testuser",
	}, nil
}

// UpdateUser updates a user for testing purposes.
func (m *MockUserUsecase) UpdateUser(user *domain.User) (*domain.User, error) {
	return user, nil
}

func main() {
	fmt.Println("=== Testing User Handler Locally ===")

	// Create handler with mock usecase
	mockUC := &MockUserUsecase{}
	handler := grpcHandler.NewUserHandler(mockUC)

	// Test 1: CreateUser
	fmt.Println("Test 1: CreateUser")
	createReq := &pb.CreateUserRequest{
		User: &pb.User{
			Email:    "test@example.com",
			Username: "testuser",
			Password: "password123",
		},
	}

	createResp, err := handler.CreateUser(context.Background(), createReq)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("Response:\n")
		fmt.Printf("  Name: %s\n", createResp.Name)
		fmt.Printf("  Email: %s\n", createResp.Email)
		fmt.Printf("  Username: %s\n", createResp.Username)
		fmt.Printf("  UserId: %s\n", createResp.UserId)

		// Verify resource name format
		expectedName := resourcenames.UserName("test-uuid-12345")
		if createResp.Name == expectedName {
			fmt.Printf("✓ Resource name correct: %s\n", createResp.Name)
		} else {
			fmt.Printf("✗ Resource name incorrect!\n")
			fmt.Printf("  Expected: %s\n", expectedName)
			fmt.Printf("  Got: %s\n", createResp.Name)
		}
	}
	fmt.Println()

	// Test 2: GetUser
	fmt.Println("Test 2: GetUser")
	getUserReq := &pb.GetUserRequest{
		Name: resourcenames.UserName("test-uuid-12345"),
	}

	getUserResp, err := handler.GetUser(context.Background(), getUserReq)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("Response:\n")
		fmt.Printf("  Name: %s\n", getUserResp.Name)
		fmt.Printf("  Email: %s\n", getUserResp.Email)
		fmt.Printf("  Username: %s\n", getUserResp.Username)
		fmt.Printf("  UserId: %s\n", getUserResp.UserId)
	}
	fmt.Println()

	// Test 3: GetUser with invalid resource name
	fmt.Println("Test 3: GetUser with invalid resource name")
	invalidReq := &pb.GetUserRequest{
		Name: "invalid-name",
	}

	_, err = handler.GetUser(context.Background(), invalidReq)
	if err != nil {
		fmt.Printf("✓ Correctly rejected invalid resource name: %v\n", err)
	} else {
		fmt.Printf("✗ Should have rejected invalid resource name\n")
	}
	fmt.Println()

	// Test 4: VerifyUser
	fmt.Println("Test 4: VerifyUser")
	verifyReq := &pb.VerifyUserRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	verifyResp, err := handler.VerifyUser(context.Background(), verifyReq)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("Response:\n")
		fmt.Printf("  Valid: %v\n", verifyResp.Valid)
		if verifyResp.User != nil {
			fmt.Printf("  User.Name: %s\n", verifyResp.User.Name)
			fmt.Printf("  User.Email: %s\n", verifyResp.User.Email)
			fmt.Printf("  User.UserId: %s\n", verifyResp.User.UserId)
		}
	}
	fmt.Println()

	fmt.Println("=== All tests completed ===")
}
