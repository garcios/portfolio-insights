// Package hydra implements a client for the Hydra administrative API.
package hydra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/garcios/portfolio-insights/apps/login-consent-provider/internal/domain"
)

// HydraClient is a client for the Hydra API.
type HydraClient struct {
	adminURL   string
	httpClient *http.Client
}

// NewHydraClient creates a new HydraClient.
func NewHydraClient(adminURL string, httpClient *http.Client) *HydraClient {
	return &HydraClient{
		adminURL:   adminURL,
		httpClient: httpClient,
	}
}

// GetLoginRequest retrieves information about a login request.
func (h *HydraClient) GetLoginRequest(challenge string) (*domain.LoginRequest, error) {
	url := fmt.Sprintf("%s/admin/oauth2/auth/requests/login?login_challenge=%s", h.adminURL, challenge)
	resp, err := h.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("hydra returned status %d: %s", resp.StatusCode, string(body))
	}

	var loginReq domain.LoginRequest
	if err := json.NewDecoder(resp.Body).Decode(&loginReq); err != nil {
		return nil, err
	}

	return &loginReq, nil
}

// AcceptLogin accepts a login request.
func (h *HydraClient) AcceptLogin(challenge, subject string, remember bool) (string, error) {
	acceptReq := domain.AcceptLoginRequest{
		Subject:     subject,
		Remember:    remember,
		RememberFor: 3600, // 1 hour
	}

	body, err := json.Marshal(acceptReq)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/admin/oauth2/auth/requests/login/accept?login_challenge=%s", h.adminURL, challenge)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("hydra returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var acceptResp domain.AcceptLoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&acceptResp); err != nil {
		return "", err
	}

	return acceptResp.RedirectTo, nil
}

// GetConsentRequest retrieves information about a consent request.
func (h *HydraClient) GetConsentRequest(challenge string) (*domain.ConsentRequest, error) {
	url := fmt.Sprintf("%s/admin/oauth2/auth/requests/consent?consent_challenge=%s", h.adminURL, challenge)
	resp, err := h.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("hydra returned status %d: %s", resp.StatusCode, string(body))
	}

	var consentReq domain.ConsentRequest
	if err := json.NewDecoder(resp.Body).Decode(&consentReq); err != nil {
		return nil, err
	}

	return &consentReq, nil
}

// AcceptConsent accepts a consent request.
func (h *HydraClient) AcceptConsent(challenge string, grantScope, grantAudience []string, user *domain.User, remember bool) (string, error) {
	acceptReq := domain.AcceptConsentRequest{
		GrantScope:               grantScope,
		GrantAccessTokenAudience: grantAudience,
		Remember:                 remember,
		RememberFor:              3600, // 1 hour
	}

	// Add user info to ID token
	acceptReq.Session.IDToken = map[string]interface{}{
		"email":      user.Email,
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"role":       user.Role,
	}

	// Add user info to access token
	acceptReq.Session.AccessToken = map[string]interface{}{
		"email":      user.Email,
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"role":       user.Role,
	}

	body, err := json.Marshal(acceptReq)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/admin/oauth2/auth/requests/consent/accept?consent_challenge=%s", h.adminURL, challenge)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("hydra returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var acceptResp domain.AcceptConsentResponse
	if err := json.NewDecoder(resp.Body).Decode(&acceptResp); err != nil {
		return "", err
	}

	return acceptResp.RedirectTo, nil
}

// RejectConsent rejects a consent request.
func (h *HydraClient) RejectConsent(challenge, reason string) (string, error) {
	rejectReq := map[string]interface{}{
		"error":             "access_denied",
		"error_description": reason,
	}

	body, err := json.Marshal(rejectReq)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/admin/oauth2/auth/requests/consent/reject?consent_challenge=%s", h.adminURL, challenge)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("hydra returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var rejectResp domain.AcceptConsentResponse
	if err := json.NewDecoder(resp.Body).Decode(&rejectResp); err != nil {
		return "", err
	}

	return rejectResp.RedirectTo, nil
}

// AcceptLogout accepts a logout request.
func (h *HydraClient) AcceptLogout(challenge string) (string, error) {
	url := fmt.Sprintf("%s/admin/oauth2/auth/requests/logout/accept?logout_challenge=%s", h.adminURL, challenge)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("hydra returned status %d: %s", resp.StatusCode, string(body))
	}

	var logoutResp struct {
		RedirectTo string `json:"redirect_to"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&logoutResp); err != nil {
		return "", err
	}

	return logoutResp.RedirectTo, nil
}
