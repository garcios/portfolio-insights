// Package usecase implements the business logic for the application.
package usecase

import (
	"context"

	"github.com/garcios/portfolio-insights/apps/login-consent-provider/internal/domain"
)

// AuthUseCase handles authentication and consent logic.
type AuthUseCase struct {
	userRepo  domain.UserRepository
	hydraRepo domain.HydraRepository
}

// NewAuthUseCase creates a new AuthUseCase.
func NewAuthUseCase(userRepo domain.UserRepository, hydraRepo domain.HydraRepository) *AuthUseCase {
	return &AuthUseCase{
		userRepo:  userRepo,
		hydraRepo: hydraRepo,
	}
}

// GetLoginRequest retrieves a login request from Hydra.
func (uc *AuthUseCase) GetLoginRequest(challenge string) (*domain.LoginRequest, error) {
	return uc.hydraRepo.GetLoginRequest(challenge)
}

// AcceptLogin accepts a login request in Hydra.
func (uc *AuthUseCase) AcceptLogin(challenge, subject string, remember bool) (string, error) {
	return uc.hydraRepo.AcceptLogin(challenge, subject, remember)
}

// VerifyUser verifies a user's credentials.
func (uc *AuthUseCase) VerifyUser(ctx context.Context, email, password string) (*domain.User, error) {
	return uc.userRepo.VerifyUser(ctx, email, password)
}

// GetConsentRequest retrieves a consent request from Hydra.
func (uc *AuthUseCase) GetConsentRequest(challenge string) (*domain.ConsentRequest, error) {
	return uc.hydraRepo.GetConsentRequest(challenge)
}

// AcceptConsent accepts a consent request in Hydra.
func (uc *AuthUseCase) AcceptConsent(ctx context.Context, challenge string, grantScope, grantAudience []string, subject string, remember bool) (string, error) {
	// Get user info for token claims
	user, err := uc.userRepo.GetUser(ctx, subject)
	if err != nil {
		return "", err
	}

	return uc.hydraRepo.AcceptConsent(challenge, grantScope, grantAudience, user, remember)
}

// RejectConsent rejects a consent request in Hydra.
func (uc *AuthUseCase) RejectConsent(challenge, reason string) (string, error) {
	return uc.hydraRepo.RejectConsent(challenge, reason)
}

// AcceptLogout accepts a logout request in Hydra.
func (uc *AuthUseCase) AcceptLogout(challenge string) (string, error) {
	return uc.hydraRepo.AcceptLogout(challenge)
}
