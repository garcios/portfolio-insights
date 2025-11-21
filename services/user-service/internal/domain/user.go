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
	Create(user *User) error
}
