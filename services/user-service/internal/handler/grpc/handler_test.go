package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/garcios/portfolio-insights/services/user-service/internal/domain"
	pb "github.com/garcios/portfolio-insights/services/user-service/proto/user"
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
	mockUC := &MockUserUsecase{
		GetUserFunc: func(id string) (*domain.User, error) {
			if id == "existing-id" {
				return &domain.User{
					ID:       "existing-id",
					Username: "Test User",
					Email:    "test@example.com",
				}, nil
			}
			return nil, errors.New("user not found")
		},
	}

	h := NewUserHandler(mockUC)

	tests := []struct {
		name    string
		req     *pb.GetUserRequest
		wantErr bool
	}{
		{
			name:    "Success",
			req:     &pb.GetUserRequest{Id: "existing-id"},
			wantErr: false,
		},
		{
			name:    "NotFound",
			req:     &pb.GetUserRequest{Id: "non-existent-id"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := h.GetUser(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserHandler.GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && resp.Id != tt.req.Id {
				t.Errorf("UserHandler.GetUser() ID = %v, want %v", resp.Id, tt.req.Id)
			}
		})
	}
}

func TestUserHandler_CreateUser(t *testing.T) {
	mockUC := &MockUserUsecase{
		CreateUserFunc: func(email, username, password string) (*domain.User, error) {
			if email == "error@example.com" {
				return nil, errors.New("creation failed")
			}
			return &domain.User{
				ID:       "generated-id",
				Email:    email,
				Username: username,
			}, nil
		},
	}

	h := NewUserHandler(mockUC)

	tests := []struct {
		name    string
		req     *pb.CreateUserRequest
		wantID  string
		wantErr bool
	}{
		{
			name: "Success",
			req: &pb.CreateUserRequest{
				Email:    "test@example.com",
				Username: "Test User",
				Password: "password",
			},
			wantID:  "generated-id",
			wantErr: false,
		},
		{
			name: "Error",
			req: &pb.CreateUserRequest{
				Email:    "error@example.com",
				Username: "Error User",
				Password: "password",
			},
			wantID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := h.CreateUser(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserHandler.CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && resp.Id != tt.wantID {
				t.Errorf("UserHandler.CreateUser() ID = %v, want %v", resp.Id, tt.wantID)
			}
		})
	}
}

func TestUserHandler_VerifyUser(t *testing.T) {
	mockUC := &MockUserUsecase{
		VerifyUserFunc: func(email, password string) (*domain.User, error) {
			if email == "valid@example.com" && password == "password" {
				return &domain.User{
					ID:       "user-id",
					Email:    email,
					Username: "Valid User",
				}, nil
			}
			return nil, errors.New("invalid credentials")
		},
	}

	h := NewUserHandler(mockUC)

	tests := []struct {
		name      string
		req       *pb.VerifyUserRequest
		wantValid bool
	}{
		{
			name: "Valid",
			req: &pb.VerifyUserRequest{
				Email:    "valid@example.com",
				Password: "password",
			},
			wantValid: true,
		},
		{
			name: "Invalid",
			req: &pb.VerifyUserRequest{
				Email:    "invalid@example.com",
				Password: "password",
			},
			wantValid: false,
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
		})
	}
}
