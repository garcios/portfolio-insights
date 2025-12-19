// Package usecase implements the business logic for the user domain.
package usecase

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/garcios/portfolio-insights/services/user-service/internal/domain"
)

// UserUsecase implements the UserUsecase interface.
type UserUsecase struct {
	repo domain.UserRepository
}

// NewUserUsecase creates a new user usecase.
func NewUserUsecase(repo domain.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

// GetUser retrieves a user by ID.
func (uc *UserUsecase) GetUser(id string) (*domain.User, error) {
	return uc.repo.GetByID(id)
}

// CreateUser creates a new user.
func (uc *UserUsecase) CreateUser(user *domain.User, password string) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Add validation logic here
	user.Password = string(hashedPassword)

	err = uc.repo.Create(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// VerifyUser verifies a user's credentials.
func (uc *UserUsecase) VerifyUser(email, password string) (*domain.User, error) {
	user, err := uc.repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return user, nil
}

// UpdateUser updates specific fields of a user.
func (uc *UserUsecase) UpdateUser(user *domain.User) (*domain.User, error) {
	// Fetch existing user to ensure they exist
	existingUser, err := uc.repo.GetByID(user.ID)
	if err != nil {
		return nil, err
	}

	// Safety check: ensure ID matches
	user.ID = existingUser.ID

	err = uc.repo.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
