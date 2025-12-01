package domain

// User model
type User struct {
	ID           string
	Email        string
	PasswordHash string
	Username     string
}

// Hydra API structures
type LoginRequest struct {
	Challenge string `json:"challenge"`
	Skip      bool   `json:"skip"`
	Subject   string `json:"subject"`
}

type AcceptLoginRequest struct {
	Subject     string `json:"subject"`
	Remember    bool   `json:"remember"`
	RememberFor int    `json:"remember_for"`
}

type AcceptLoginResponse struct {
	RedirectTo string `json:"redirect_to"`
}

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

type AcceptConsentRequest struct {
	GrantScope               []string             `json:"grant_scope"`
	GrantAccessTokenAudience []string             `json:"grant_access_token_audience"`
	Remember                 bool                 `json:"remember"`
	RememberFor              int                  `json:"remember_for"`
	Session                  AcceptConsentSession `json:"session"`
}

type AcceptConsentSession struct {
	IDToken     map[string]interface{} `json:"id_token"`
	AccessToken map[string]interface{} `json:"access_token"`
}

type AcceptConsentResponse struct {
	RedirectTo string `json:"redirect_to"`
}
