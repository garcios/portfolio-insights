package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

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
	GrantScope               []string `json:"grant_scope"`
	GrantAccessTokenAudience []string `json:"grant_access_token_audience"`
	Remember                 bool     `json:"remember"`
	RememberFor              int      `json:"remember_for"`
	Session                  struct {
		IDToken     map[string]interface{} `json:"id_token"`
		AccessToken map[string]interface{} `json:"access_token"`
	} `json:"session"`
}

type AcceptConsentResponse struct {
	RedirectTo string `json:"redirect_to"`
}

// User model
type User struct {
	ID           string
	Email        string
	PasswordHash string
	Username     string
}

// Login GET - Display login form
func (app *App) loginGet(c *gin.Context) {
	challenge := c.Query("login_challenge")
	if challenge == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Missing login_challenge parameter",
		})
		return
	}

	// Get login request from Hydra
	loginReq, err := app.getLoginRequest(challenge)
	if err != nil {
		log.Printf("Error getting login request: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to get login request: %v", err),
		})
		return
	}

	// If skip is true, user is already authenticated
	if loginReq.Skip {
		redirectTo, err := app.acceptLogin(challenge, loginReq.Subject, true)
		if err != nil {
			log.Printf("Error accepting login: %v", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"error": fmt.Sprintf("Failed to accept login: %v", err),
			})
			return
		}
		c.Redirect(http.StatusFound, redirectTo)
		return
	}

	// Show login form
	c.HTML(http.StatusOK, "login.html", gin.H{
		"challenge": challenge,
	})
}

// Login POST - Process login form
func (app *App) loginPost(c *gin.Context) {
	challenge := c.PostForm("challenge")
	email := c.PostForm("email")
	password := c.PostForm("password")
	remember := c.PostForm("remember") == "on"

	if challenge == "" || email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"challenge": challenge,
			"error":     "Email and password are required",
		})
		return
	}

	// Authenticate user
	user, err := app.authenticateUser(c.Request.Context(), email, password)
	if err != nil {
		log.Printf("Authentication failed for %s: %v", email, err)
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"challenge": challenge,
			"error":     "Invalid email or password",
		})
		return
	}

	// Store user in session
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("email", user.Email)
	session.Save()

	// Accept login with Hydra
	redirectTo, err := app.acceptLogin(challenge, user.ID, remember)
	if err != nil {
		log.Printf("Error accepting login: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to accept login: %v", err),
		})
		return
	}

	c.Redirect(http.StatusFound, redirectTo)
}

// Consent GET - Display consent form
func (app *App) consentGet(c *gin.Context) {
	challenge := c.Query("consent_challenge")
	if challenge == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Missing consent_challenge parameter",
		})
		return
	}

	// Get consent request from Hydra
	consentReq, err := app.getConsentRequest(challenge)
	if err != nil {
		log.Printf("Error getting consent request: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to get consent request: %v", err),
		})
		return
	}

	// If skip is true, user has already consented
	if consentReq.Skip {
		redirectTo, err := app.acceptConsent(challenge, consentReq.RequestedScope, consentReq.RequestedAudience, consentReq.Subject, true)
		if err != nil {
			log.Printf("Error accepting consent: %v", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"error": fmt.Sprintf("Failed to accept consent: %v", err),
			})
			return
		}
		c.Redirect(http.StatusFound, redirectTo)
		return
	}

	// Show consent form
	c.HTML(http.StatusOK, "consent.html", gin.H{
		"challenge":       challenge,
		"client_name":     consentReq.Client.ClientName,
		"requested_scope": consentReq.RequestedScope,
		"subject":         consentReq.Subject,
	})
}

// Consent POST - Process consent form
func (app *App) consentPost(c *gin.Context) {
	challenge := c.PostForm("challenge")
	submit := c.PostForm("submit")
	remember := c.PostForm("remember") == "on"

	if challenge == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Missing challenge parameter",
		})
		return
	}

	// Get consent request
	consentReq, err := app.getConsentRequest(challenge)
	if err != nil {
		log.Printf("Error getting consent request: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to get consent request: %v", err),
		})
		return
	}

	// Check if user denied consent
	if submit == "deny" {
		redirectTo, err := app.rejectConsent(challenge, "User denied consent")
		if err != nil {
			log.Printf("Error rejecting consent: %v", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"error": fmt.Sprintf("Failed to reject consent: %v", err),
			})
			return
		}
		c.Redirect(http.StatusFound, redirectTo)
		return
	}

	// Accept consent
	redirectTo, err := app.acceptConsent(challenge, consentReq.RequestedScope, consentReq.RequestedAudience, consentReq.Subject, remember)
	if err != nil {
		log.Printf("Error accepting consent: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to accept consent: %v", err),
		})
		return
	}

	c.Redirect(http.StatusFound, redirectTo)
}

// Logout GET - Display logout confirmation
func (app *App) logoutGet(c *gin.Context) {
	challenge := c.Query("logout_challenge")
	if challenge == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Missing logout_challenge parameter",
		})
		return
	}

	c.HTML(http.StatusOK, "logout.html", gin.H{
		"challenge": challenge,
	})
}

// Logout POST - Process logout
func (app *App) logoutPost(c *gin.Context) {
	challenge := c.PostForm("challenge")

	// Clear session
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	// Accept logout with Hydra
	redirectTo, err := app.acceptLogout(challenge)
	if err != nil {
		log.Printf("Error accepting logout: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to accept logout: %v", err),
		})
		return
	}

	c.Redirect(http.StatusFound, redirectTo)
}

// Error GET - Display error page
func (app *App) errorGet(c *gin.Context) {
	errorMsg := c.Query("error")
	errorDesc := c.Query("error_description")

	c.HTML(http.StatusOK, "error.html", gin.H{
		"error":       errorMsg,
		"description": errorDesc,
	})
}

// Helper functions for Hydra API calls

func (app *App) getLoginRequest(challenge string) (*LoginRequest, error) {
	url := fmt.Sprintf("%s/admin/oauth2/auth/requests/login?login_challenge=%s", app.hydraAdmin, challenge)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("hydra returned status %d: %s", resp.StatusCode, string(body))
	}

	var loginReq LoginRequest
	if err := json.NewDecoder(resp.Body).Decode(&loginReq); err != nil {
		return nil, err
	}

	return &loginReq, nil
}

func (app *App) acceptLogin(challenge, subject string, remember bool) (string, error) {
	acceptReq := AcceptLoginRequest{
		Subject:     subject,
		Remember:    remember,
		RememberFor: 3600, // 1 hour
	}

	body, err := json.Marshal(acceptReq)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/admin/oauth2/auth/requests/login/accept?login_challenge=%s", app.hydraAdmin, challenge)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("hydra returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var acceptResp AcceptLoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&acceptResp); err != nil {
		return "", err
	}

	return acceptResp.RedirectTo, nil
}

func (app *App) getConsentRequest(challenge string) (*ConsentRequest, error) {
	url := fmt.Sprintf("%s/admin/oauth2/auth/requests/consent?consent_challenge=%s", app.hydraAdmin, challenge)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("hydra returned status %d: %s", resp.StatusCode, string(body))
	}

	var consentReq ConsentRequest
	if err := json.NewDecoder(resp.Body).Decode(&consentReq); err != nil {
		return nil, err
	}

	return &consentReq, nil
}

func (app *App) acceptConsent(challenge string, grantScope, grantAudience []string, subject string, remember bool) (string, error) {
	// Get user info for token claims
	user, err := app.getUserByID(context.Background(), subject)
	if err != nil {
		return "", err
	}

	acceptReq := AcceptConsentRequest{
		GrantScope:               grantScope,
		GrantAccessTokenAudience: grantAudience,
		Remember:                 remember,
		RememberFor:              3600, // 1 hour
	}

	// Add user info to ID token
	acceptReq.Session.IDToken = map[string]interface{}{
		"email":    user.Email,
		"username": user.Username,
	}

	// Add user info to access token
	acceptReq.Session.AccessToken = map[string]interface{}{
		"email":    user.Email,
		"username": user.Username,
	}

	body, err := json.Marshal(acceptReq)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/admin/oauth2/auth/requests/consent/accept?consent_challenge=%s", app.hydraAdmin, challenge)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("hydra returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var acceptResp AcceptConsentResponse
	if err := json.NewDecoder(resp.Body).Decode(&acceptResp); err != nil {
		return "", err
	}

	return acceptResp.RedirectTo, nil
}

func (app *App) rejectConsent(challenge, reason string) (string, error) {
	rejectReq := map[string]interface{}{
		"error":             "access_denied",
		"error_description": reason,
	}

	body, err := json.Marshal(rejectReq)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/admin/oauth2/auth/requests/consent/reject?consent_challenge=%s", app.hydraAdmin, challenge)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("hydra returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var rejectResp AcceptConsentResponse
	if err := json.NewDecoder(resp.Body).Decode(&rejectResp); err != nil {
		return "", err
	}

	return rejectResp.RedirectTo, nil
}

func (app *App) acceptLogout(challenge string) (string, error) {
	url := fmt.Sprintf("%s/admin/oauth2/auth/requests/logout/accept?logout_challenge=%s", app.hydraAdmin, challenge)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

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

// Database functions

func (app *App) authenticateUser(ctx context.Context, email, password string) (*User, error) {
	var user User
	query := `SELECT id, email, password_hash, username FROM customers.users WHERE email = $1`
	err := app.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return &user, nil
}

func (app *App) getUserByID(ctx context.Context, userID string) (*User, error) {
	var user User
	query := `SELECT id, email, username FROM customers.users WHERE id = $1`
	err := app.db.QueryRow(ctx, query, userID).Scan(&user.ID, &user.Email, &user.Username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}
