package entity

import "time"

// User represents a user in the system
type User struct {
	ID          string
	Username    string
	Email       string
	FirstName   string
	LastName    string
	Role        string
	Preferences map[string]interface{}
	LastLoginAt *time.Time
}

// NewUser creates a new User entity
func NewUser(id, username, email string) *User {
	return &User{
		ID:       id,
		Username: username,
		Email:    email,
	}
}
