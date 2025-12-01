package domain

import "context"

// UserRepository defines the interface for user service operations
type UserRepository interface {
	VerifyUser(ctx context.Context, email, password string) (*User, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	Close() error
}

// HydraRepository defines the interface for Hydra operations
type HydraRepository interface {
	GetLoginRequest(challenge string) (*LoginRequest, error)
	AcceptLogin(challenge, subject string, remember bool) (string, error)
	GetConsentRequest(challenge string) (*ConsentRequest, error)
	AcceptConsent(challenge string, grantScope, grantAudience []string, user *User, remember bool) (string, error)
	RejectConsent(challenge, reason string) (string, error)
	AcceptLogout(challenge string) (string, error)
}
