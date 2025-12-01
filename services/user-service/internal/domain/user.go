package domain

import "time"

type User struct {
	ID        string
	Email     string
	Username  string
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
	Count() (int, error)
}

type UserUsecase interface {
	GetUser(id string) (*User, error)
	CreateUser(email, username, password string) (*User, error)
	VerifyUser(email, password string) (*User, error)
}
