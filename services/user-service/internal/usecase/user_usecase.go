package usecase

import (
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

func (uc *UserUsecase) CreateUser(email, name, password string) (*domain.User, error) {
	// Add validation logic here
	u := &domain.User{
		Email:    email,
		Name:     name,
		Password: password, // Hash this in real impl
	}
	err := uc.repo.Create(u)
	if err != nil {
		return nil, err
	}
	return u, nil
}
