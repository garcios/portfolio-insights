// Package domain holds the business entities and interface definitions for the user domain.
package domain

import "time"

// User represents a user in the system.
type User struct {
	ID        string
	Email     string
	Username  string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserRepository defines the interface for user data persistence.
type UserRepository interface {
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(id string) error
	Count() (int, error)
}

// UserUsecase defines the interface for user business logic.
type UserUsecase interface {
	GetUser(id string) (*User, error)
	CreateUser(email, username, password string) (*User, error)
	VerifyUser(email, password string) (*User, error)
}
