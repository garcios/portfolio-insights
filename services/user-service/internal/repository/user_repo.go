package repository

import (
	"database/sql"
	"errors"

	"github.com/garcios/portfolio-insights/services/user-service/internal/domain"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(id string) (*domain.User, error) {
	// Mock implementation
	if id == "1" {
		return &domain.User{ID: "1", Name: "John Doe", Email: "john@example.com"}, nil
	}
	return nil, errors.New("user not found")
}

func (r *userRepository) Create(user *domain.User) error {
	// Mock implementation
	return nil
}
