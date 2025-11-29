package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
)

// MockUserGateway is a manual mock for UserGateway
type MockUserGateway struct {
	GetUserFunc    func(ctx context.Context, id string) (*entity.User, error)
	CreateUserFunc func(ctx context.Context, email, username, password string) (*entity.User, error)
}

func (m *MockUserGateway) GetUser(ctx context.Context, id string) (*entity.User, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserGateway) CreateUser(ctx context.Context, email, username, password string) (*entity.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, email, username, password)
	}
	return nil, nil
}

func TestUserUseCase_GetUser(t *testing.T) {
	mockGateway := &MockUserGateway{
		GetUserFunc: func(ctx context.Context, id string) (*entity.User, error) {
			if id == "user-1" {
				return entity.NewUser("user-1", "testuser", "test@example.com"), nil
			}
			return nil, errors.New("user not found")
		},
	}

	uc := NewUserUseCase(mockGateway)

	t.Run("success", func(t *testing.T) {
		user, err := uc.GetUser(context.Background(), "user-1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.ID != "user-1" {
			t.Errorf("expected ID user-1, got %s", user.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := uc.GetUser(context.Background(), "unknown")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestUserUseCase_CreateUser(t *testing.T) {
	mockGateway := &MockUserGateway{
		CreateUserFunc: func(ctx context.Context, email, username, password string) (*entity.User, error) {
			return entity.NewUser("new-id", username, email), nil
		},
	}

	uc := NewUserUseCase(mockGateway)

	user, err := uc.CreateUser(context.Background(), "test@example.com", "testuser", "password")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
}
