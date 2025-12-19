// Package domain holds the business entities and interface definitions for the user domain.
package domain

import "time"

// User represents a user in the system.
type User struct {
	ID          string
	Email       string
	Username    string
	Password    string
	FirstName   string
	LastName    string
	Role        string
	Preferences map[string]interface{}
	LastLoginAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
	VerifyUser(email, password string) (*User, error)
	CreateUser(user *User, password string) (*User, error)
	UpdateUser(user *User) (*User, error)
}
