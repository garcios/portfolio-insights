// Package domain defines the domain models and interfaces.
package domain

// User model
type User struct {
	ID           string
	Email        string
	PasswordHash string
	Username     string
}

// LoginRequest represents a login request from Hydra.
type LoginRequest struct {
	Challenge string `json:"challenge"`
	Skip      bool   `json:"skip"`
	Subject   string `json:"subject"`
}

// AcceptLoginRequest represents a request to accept a login challenge.
type AcceptLoginRequest struct {
	Subject     string `json:"subject"`
	Remember    bool   `json:"remember"`
	RememberFor int    `json:"remember_for"`
}

// AcceptLoginResponse represents the response after accepting a login challenge.
type AcceptLoginResponse struct {
	RedirectTo string `json:"redirect_to"`
}

// ConsentRequest represents a consent request from Hydra.
type ConsentRequest struct {
	Challenge         string   `json:"challenge"`
	Skip              bool     `json:"skip"`
	Subject           string   `json:"subject"`
	RequestedScope    []string `json:"requested_scope"`
	RequestedAudience []string `json:"requested_audience"`
	Client            struct {
		ClientID   string `json:"client_id"`
		ClientName string `json:"client_name"`
	} `json:"client"`
}

// AcceptConsentRequest represents a request to accept a consent challenge.
type AcceptConsentRequest struct {
	GrantScope               []string             `json:"grant_scope"`
	GrantAccessTokenAudience []string             `json:"grant_access_token_audience"`
	Remember                 bool                 `json:"remember"`
	RememberFor              int                  `json:"remember_for"`
	Session                  AcceptConsentSession `json:"session"`
}

// AcceptConsentSession represents the session data associated with an accepted consent.
type AcceptConsentSession struct {
	IDToken     map[string]interface{} `json:"id_token"`
	AccessToken map[string]interface{} `json:"access_token"`
}

// AcceptConsentResponse represents the response after accepting a consent challenge.
type AcceptConsentResponse struct {
	RedirectTo string `json:"redirect_to"`
}
