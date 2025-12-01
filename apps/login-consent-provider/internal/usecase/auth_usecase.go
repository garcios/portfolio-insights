package usecase

import (
	"context"

	"github.com/garcios/portfolio-insights/apps/login-consent-provider/internal/domain"
)

type AuthUseCase struct {
	userRepo  domain.UserRepository
	hydraRepo domain.HydraRepository
}

func NewAuthUseCase(userRepo domain.UserRepository, hydraRepo domain.HydraRepository) *AuthUseCase {
	return &AuthUseCase{
		userRepo:  userRepo,
		hydraRepo: hydraRepo,
	}
}

func (uc *AuthUseCase) GetLoginRequest(challenge string) (*domain.LoginRequest, error) {
	return uc.hydraRepo.GetLoginRequest(challenge)
}

func (uc *AuthUseCase) AcceptLogin(challenge, subject string, remember bool) (string, error) {
	return uc.hydraRepo.AcceptLogin(challenge, subject, remember)
}

func (uc *AuthUseCase) VerifyUser(ctx context.Context, email, password string) (*domain.User, error) {
	return uc.userRepo.VerifyUser(ctx, email, password)
}

func (uc *AuthUseCase) GetConsentRequest(challenge string) (*domain.ConsentRequest, error) {
	return uc.hydraRepo.GetConsentRequest(challenge)
}

func (uc *AuthUseCase) AcceptConsent(ctx context.Context, challenge string, grantScope, grantAudience []string, subject string, remember bool) (string, error) {
	// Get user info for token claims
	user, err := uc.userRepo.GetUser(ctx, subject)
	if err != nil {
		return "", err
	}

	return uc.hydraRepo.AcceptConsent(challenge, grantScope, grantAudience, user, remember)
}

func (uc *AuthUseCase) RejectConsent(challenge, reason string) (string, error) {
	return uc.hydraRepo.RejectConsent(challenge, reason)
}

func (uc *AuthUseCase) AcceptLogout(challenge string) (string, error) {
	return uc.hydraRepo.AcceptLogout(challenge)
}
