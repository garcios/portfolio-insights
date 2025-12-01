package usecase

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/garcios/portfolio-insights/services/user-service/internal/domain"
)

type UserUsecase struct {
	repo domain.UserRepository
}

func NewUserUsecase(repo domain.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (uc *UserUsecase) GetUser(id string) (*domain.User, error) {
	return uc.repo.GetByID(id)
}

func (uc *UserUsecase) CreateUser(email, username, password string) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Add validation logic here
	u := &domain.User{
		Email:    email,
		Username: username,
		Password: string(hashedPassword),
	}
	err = uc.repo.Create(u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (uc *UserUsecase) VerifyUser(email, password string) (*domain.User, error) {
	user, err := uc.repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}
