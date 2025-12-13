// Package domain holds the business entities and interface definitions for the user domain.
package domain

import "errors"

// Domain-level errors for the user service.
var (
	// ErrUserNotFound is returned when a user cannot be found.
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists is returned when attempting to create a user that already exists.
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrInvalidCredentials is returned when user credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
)
