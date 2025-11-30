package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/services/user-service/internal/domain"
)

// MockUserRepository is a manual mock for domain.UserRepository
type MockUserRepository struct {
	GetByIDFunc    func(id string) (*domain.User, error)
	GetByEmailFunc func(email string) (*domain.User, error)
	CreateFunc     func(user *domain.User) error
	UpdateFunc     func(user *domain.User) error
	DeleteFunc     func(id string) error
	CountFunc      func() (int, error)
}

func (m *MockUserRepository) GetByID(id string) (*domain.User, error) {
	return m.GetByIDFunc(id)
}

func (m *MockUserRepository) GetByEmail(email string) (*domain.User, error) {
	return m.GetByEmailFunc(email)
}

func (m *MockUserRepository) Create(user *domain.User) error {
	return m.CreateFunc(user)
}

func (m *MockUserRepository) Update(user *domain.User) error {
	return m.UpdateFunc(user)
}

func (m *MockUserRepository) Delete(id string) error {
	return m.DeleteFunc(id)
}

func (m *MockUserRepository) Count() (int, error) {
	if m.CountFunc != nil {
		return m.CountFunc()
	}
	return 0, nil
}

func TestUserUsecase_GetUser(t *testing.T) {
	mockRepo := &MockUserRepository{
		GetByIDFunc: func(id string) (*domain.User, error) {
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

	uc := NewUserUsecase(mockRepo)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "Success",
			id:      "existing-id",
			wantErr: false,
		},
		{
			name:    "NotFound",
			id:      "non-existent-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := uc.GetUser(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserUsecase.GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.id {
				t.Errorf("UserUsecase.GetUser() = %v, want %v", got.ID, tt.id)
			}
		})
	}
}

func TestUserUsecase_CreateUser(t *testing.T) {
	mockRepo := &MockUserRepository{
		CreateFunc: func(user *domain.User) error {
			if user.Email == "error@example.com" {
				return errors.New("database error")
			}
			user.ID = "generated-id"
			user.CreatedAt = time.Now()
			return nil
		},
	}

	uc := NewUserUsecase(mockRepo)

	tests := []struct {
		name    string
		email   string
		wantID  string
		wantErr bool
	}{
		{
			name:    "Success",
			email:   "test@example.com",
			wantID:  "generated-id",
			wantErr: false,
		},
		{
			name:    "DatabaseError",
			email:   "error@example.com",
			wantID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := uc.CreateUser(tt.email, "Test User", "password")
			if (err != nil) != tt.wantErr {
				t.Errorf("UserUsecase.CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.wantID {
				t.Errorf("UserUsecase.CreateUser() ID = %v, want %v", got.ID, tt.wantID)
			}
		})
	}
}
