package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/garcios/portfolio-insights/pkg/resourcenames"
	"github.com/garcios/portfolio-insights/services/user-service/internal/domain"
	pb "github.com/garcios/portfolio-insights/services/user-service/user"
)

// MockUserUsecase is a manual mock for domain.UserUsecase
type MockUserUsecase struct {
	GetUserFunc    func(id string) (*domain.User, error)
	CreateUserFunc func(email, name, password string) (*domain.User, error)
	VerifyUserFunc func(email, password string) (*domain.User, error)
}

func (m *MockUserUsecase) GetUser(id string) (*domain.User, error) {
	return m.GetUserFunc(id)
}

func (m *MockUserUsecase) CreateUser(email, username, password string) (*domain.User, error) {
	return m.CreateUserFunc(email, username, password)
}

func (m *MockUserUsecase) VerifyUser(email, password string) (*domain.User, error) {
	return m.VerifyUserFunc(email, password)
}

func TestUserHandler_GetUser(t *testing.T) {
	// Use valid UUIDs for testing
	existingUserID := "550e8400-e29b-41d4-a716-446655440000"
	nonExistentUserID := "550e8400-e29b-41d4-a716-446655440001"

	mockUC := &MockUserUsecase{
		GetUserFunc: func(id string) (*domain.User, error) {
			if id == existingUserID {
				return &domain.User{
					ID:       existingUserID,
					Username: "Test User",
					Email:    "test@example.com",
				}, nil
			}
			return nil, errors.New("user not found")
		},
	}

	h := NewUserHandler(mockUC)

	tests := []struct {
		name       string
		req        *pb.GetUserRequest
		wantUserID string
		wantErr    bool
	}{
		{
			name:       "Success",
			req:        &pb.GetUserRequest{Name: resourcenames.UserName(existingUserID)},
			wantUserID: existingUserID,
			wantErr:    false,
		},
		{
			name:       "NotFound",
			req:        &pb.GetUserRequest{Name: resourcenames.UserName(nonExistentUserID)},
			wantUserID: "",
			wantErr:    true,
		},
		{
			name:       "InvalidResourceName",
			req:        &pb.GetUserRequest{Name: "invalid-name"},
			wantUserID: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := h.GetUser(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserHandler.GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if resp.UserId != tt.wantUserID {
					t.Errorf("UserHandler.GetUser() UserId = %v, want %v", resp.UserId, tt.wantUserID)
				}
				if resp.Name != resourcenames.UserName(tt.wantUserID) {
					t.Errorf("UserHandler.GetUser() Name = %v, want %v", resp.Name, resourcenames.UserName(tt.wantUserID))
				}
			}
		})
	}
}

func TestUserHandler_CreateUser(t *testing.T) {
	generatedUserID := "550e8400-e29b-41d4-a716-446655440002"

	mockUC := &MockUserUsecase{
		CreateUserFunc: func(email, username, password string) (*domain.User, error) {
			if email == "error@example.com" {
				return nil, errors.New("creation failed")
			}
			return &domain.User{
				ID:       generatedUserID,
				Email:    email,
				Username: username,
			}, nil
		},
	}

	h := NewUserHandler(mockUC)

	tests := []struct {
		name       string
		req        *pb.CreateUserRequest
		wantUserID string
		wantErr    bool
	}{
		{
			name: "Success",
			req: &pb.CreateUserRequest{
				User: &pb.User{
					Email:    "test@example.com",
					Username: "Test User",
					Password: "password",
				},
			},
			wantUserID: generatedUserID,
			wantErr:    false,
		},
		{
			name: "Error",
			req: &pb.CreateUserRequest{
				User: &pb.User{
					Email:    "error@example.com",
					Username: "Error User",
					Password: "password",
				},
			},
			wantUserID: "",
			wantErr:    true,
		},
		{
			name:       "NilUser",
			req:        &pb.CreateUserRequest{User: nil},
			wantUserID: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := h.CreateUser(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserHandler.CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if resp.UserId != tt.wantUserID {
					t.Errorf("UserHandler.CreateUser() UserId = %v, want %v", resp.UserId, tt.wantUserID)
				}
				if resp.Name != resourcenames.UserName(tt.wantUserID) {
					t.Errorf("UserHandler.CreateUser() Name = %v, want %v", resp.Name, resourcenames.UserName(tt.wantUserID))
				}
			}
		})
	}
}

func TestUserHandler_VerifyUser(t *testing.T) {
	verifiedUserID := "550e8400-e29b-41d4-a716-446655440003"

	mockUC := &MockUserUsecase{
		VerifyUserFunc: func(email, password string) (*domain.User, error) {
			if email == "valid@example.com" && password == "password" {
				return &domain.User{
					ID:       verifiedUserID,
					Email:    email,
					Username: "Valid User",
				}, nil
			}
			return nil, errors.New("invalid credentials")
		},
	}

	h := NewUserHandler(mockUC)

	tests := []struct {
		name       string
		req        *pb.VerifyUserRequest
		wantValid  bool
		wantUserID string
	}{
		{
			name: "Valid",
			req: &pb.VerifyUserRequest{
				Email:    "valid@example.com",
				Password: "password",
			},
			wantValid:  true,
			wantUserID: verifiedUserID,
		},
		{
			name: "Invalid",
			req: &pb.VerifyUserRequest{
				Email:    "invalid@example.com",
				Password: "password",
			},
			wantValid:  false,
			wantUserID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := h.VerifyUser(context.Background(), tt.req)
			if err != nil {
				t.Errorf("UserHandler.VerifyUser() error = %v", err)
				return
			}
			if resp.Valid != tt.wantValid {
				t.Errorf("UserHandler.VerifyUser() Valid = %v, want %v", resp.Valid, tt.wantValid)
			}
			if tt.wantValid {
				if resp.User == nil {
					t.Error("UserHandler.VerifyUser() User is nil for valid response")
					return
				}
				if resp.User.UserId != tt.wantUserID {
					t.Errorf("UserHandler.VerifyUser() User.UserId = %v, want %v", resp.User.UserId, tt.wantUserID)
				}
				if resp.User.Name != resourcenames.UserName(tt.wantUserID) {
					t.Errorf("UserHandler.VerifyUser() User.Name = %v, want %v", resp.User.Name, resourcenames.UserName(tt.wantUserID))
				}
			}
		})
	}
}
