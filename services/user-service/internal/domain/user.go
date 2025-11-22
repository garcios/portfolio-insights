package domain

import "time"

type User struct {
	ID        string
	Email     string
	Name      string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRepository interface {
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(id string) error
}

type UserUsecase interface {
	GetUser(id string) (*User, error)
	CreateUser(email, name, password string) (*User, error)
}
