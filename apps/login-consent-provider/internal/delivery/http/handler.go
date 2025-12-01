package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/garcios/portfolio-insights/apps/login-consent-provider/internal/usecase"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	authUseCase *usecase.AuthUseCase
}

func NewHandler(authUseCase *usecase.AuthUseCase) *Handler {
	return &Handler{
		authUseCase: authUseCase,
	}
}

// Login GET - Display login form
func (h *Handler) LoginGet(c *gin.Context) {
	challenge := c.Query("login_challenge")
	if challenge == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Missing login_challenge parameter",
		})
		return
	}

	// Get login request from Hydra
	loginReq, err := h.authUseCase.GetLoginRequest(challenge)
	if err != nil {
		log.Printf("Error getting login request: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to get login request: %v", err),
		})
		return
	}

	// If skip is true, user is already authenticated
	if loginReq.Skip {
		redirectTo, err := h.authUseCase.AcceptLogin(challenge, loginReq.Subject, true)
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
func (h *Handler) LoginPost(c *gin.Context) {
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
	user, err := h.authUseCase.VerifyUser(c.Request.Context(), email, password)
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
	redirectTo, err := h.authUseCase.AcceptLogin(challenge, user.ID, remember)
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
func (h *Handler) ConsentGet(c *gin.Context) {
	challenge := c.Query("consent_challenge")
	if challenge == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Missing consent_challenge parameter",
		})
		return
	}

	// Get consent request from Hydra
	consentReq, err := h.authUseCase.GetConsentRequest(challenge)
	if err != nil {
		log.Printf("Error getting consent request: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to get consent request: %v", err),
		})
		return
	}

	// TODO: Remove this after investigating why this is always false.
	consentReq.Skip = true

	// If skip is true, user has already consented
	if consentReq.Skip {
		redirectTo, err := h.authUseCase.AcceptConsent(c.Request.Context(), challenge, consentReq.RequestedScope, consentReq.RequestedAudience, consentReq.Subject, true)
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
func (h *Handler) ConsentPost(c *gin.Context) {
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
	consentReq, err := h.authUseCase.GetConsentRequest(challenge)
	if err != nil {
		log.Printf("Error getting consent request: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to get consent request: %v", err),
		})
		return
	}

	// Check if user denied consent
	if submit == "deny" {
		redirectTo, err := h.authUseCase.RejectConsent(challenge, "User denied consent")
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
	redirectTo, err := h.authUseCase.AcceptConsent(c.Request.Context(), challenge, consentReq.RequestedScope, consentReq.RequestedAudience, consentReq.Subject, remember)
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
func (h *Handler) LogoutGet(c *gin.Context) {
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
func (h *Handler) LogoutPost(c *gin.Context) {
	challenge := c.PostForm("challenge")

	// Clear session
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	// Accept logout with Hydra
	redirectTo, err := h.authUseCase.AcceptLogout(challenge)
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
func (h *Handler) ErrorGet(c *gin.Context) {
	errorMsg := c.Query("error")
	errorDesc := c.Query("error_description")

	c.HTML(http.StatusOK, "error.html", gin.H{
		"error":       errorMsg,
		"description": errorDesc,
	})
}

// HealthCheck - Simple health check
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
