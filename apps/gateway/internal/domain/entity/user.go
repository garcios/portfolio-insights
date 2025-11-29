package entity

// User represents a user in the system
type User struct {
	ID       string
	Username string
	Email    string
}

// NewUser creates a new User entity
func NewUser(id, username, email string) *User {
	return &User{
		ID:       id,
		Username: username,
		Email:    email,
	}
}
