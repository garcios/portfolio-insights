package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/garcios/portfolio-insights/apps/login-consent-provider/internal/domain"
	"github.com/garcios/portfolio-insights/apps/login-consent-provider/internal/usecase"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// MockUserRepository
type MockUserRepository struct {
	VerifyUserFunc func(ctx context.Context, email, password string) (*domain.User, error)
	GetUserFunc    func(ctx context.Context, userID string) (*domain.User, error)
}

func (m *MockUserRepository) VerifyUser(ctx context.Context, email, password string) (*domain.User, error) {
	if m.VerifyUserFunc != nil {
		return m.VerifyUserFunc(ctx, email, password)
	}
	return nil, nil
}

func (m *MockUserRepository) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockUserRepository) Close() error {
	return nil
}

// MockHydraRepository
type MockHydraRepository struct {
	GetLoginRequestFunc   func(challenge string) (*domain.LoginRequest, error)
	AcceptLoginFunc       func(challenge, subject string, remember bool) (string, error)
	GetConsentRequestFunc func(challenge string) (*domain.ConsentRequest, error)
	AcceptConsentFunc     func(challenge string, grantScope, grantAudience []string, user *domain.User, remember bool) (string, error)
	RejectConsentFunc     func(challenge, reason string) (string, error)
	AcceptLogoutFunc      func(challenge string) (string, error)
}

func (m *MockHydraRepository) GetLoginRequest(challenge string) (*domain.LoginRequest, error) {
	if m.GetLoginRequestFunc != nil {
		return m.GetLoginRequestFunc(challenge)
	}
	return nil, nil
}

func (m *MockHydraRepository) AcceptLogin(challenge, subject string, remember bool) (string, error) {
	if m.AcceptLoginFunc != nil {
		return m.AcceptLoginFunc(challenge, subject, remember)
	}
	return "", nil
}

func (m *MockHydraRepository) GetConsentRequest(challenge string) (*domain.ConsentRequest, error) {
	if m.GetConsentRequestFunc != nil {
		return m.GetConsentRequestFunc(challenge)
	}
	return nil, nil
}

func (m *MockHydraRepository) AcceptConsent(challenge string, grantScope, grantAudience []string, user *domain.User, remember bool) (string, error) {
	if m.AcceptConsentFunc != nil {
		return m.AcceptConsentFunc(challenge, grantScope, grantAudience, user, remember)
	}
	return "", nil
}

func (m *MockHydraRepository) RejectConsent(challenge, reason string) (string, error) {
	if m.RejectConsentFunc != nil {
		return m.RejectConsentFunc(challenge, reason)
	}
	return "", nil
}

func (m *MockHydraRepository) AcceptLogout(challenge string) (string, error) {
	if m.AcceptLogoutFunc != nil {
		return m.AcceptLogoutFunc(challenge)
	}
	return "", nil
}

func setupTestApp(userRepo domain.UserRepository, hydraRepo domain.HydraRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)

	authUseCase := usecase.NewAuthUseCase(userRepo, hydraRepo)
	handler := NewHandler(authUseCase)
	sessionStore := cookie.NewStore([]byte("secret"))

	// We need to recreate the router setup here or export NewRouter.
	// Since NewRouter is in the same package, we can use it.
	// But NewRouter loads templates from "templates/*".
	// We need to make sure the test runs in a directory where "templates" exists,
	// or we need to mock the template loading.
	// Gin's LoadHTMLGlob is hard to mock without changing the code structure significantly (e.g. interface for router).
	// However, usually tests run in the package directory.
	// The templates are in `apps/login-consent-provider/templates`.
	// The test is in `apps/login-consent-provider/internal/delivery/http`.
	// So `templates/*` won't work. It needs to be `../../../templates/*`.

	// To fix this, we can pass the template path to NewRouter or handle it here.
	// Let's copy NewRouter logic but adjust template path.

	router := gin.Default()
	router.Use(sessions.Sessions("auth_session", sessionStore))

	// Adjust path for tests
	router.LoadHTMLGlob("../../../templates/*")

	router.GET("/health", handler.HealthCheck)
	router.GET("/login", handler.LoginGet)
	router.POST("/login", handler.LoginPost)
	// ... add other routes as needed for tests

	return router
}

func TestHealthCheck(t *testing.T) {
	router := setupTestApp(&MockUserRepository{}, &MockHydraRepository{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response["status"])
	}
}

func TestLoginGet_MissingChallenge(t *testing.T) {
	router := setupTestApp(&MockUserRepository{}, &MockHydraRepository{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/login", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestLoginGet_Success(t *testing.T) {
	mockHydra := &MockHydraRepository{
		GetLoginRequestFunc: func(challenge string) (*domain.LoginRequest, error) {
			return &domain.LoginRequest{
				Challenge: challenge,
				Skip:      false,
			}, nil
		},
	}

	router := setupTestApp(&MockUserRepository{}, mockHydra)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/login?login_challenge=challenge123", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestLoginPost_Success(t *testing.T) {
	mockUser := &MockUserRepository{
		VerifyUserFunc: func(ctx context.Context, email, password string) (*domain.User, error) {
			if email == "test@example.com" && password == "password123" {
				return &domain.User{
					ID:       "user123",
					Email:    "test@example.com",
					Username: "testuser",
				}, nil
			}
			return nil, fmt.Errorf("invalid credentials")
		},
	}

	mockHydra := &MockHydraRepository{
		AcceptLoginFunc: func(challenge, subject string, remember bool) (string, error) {
			return "http://callback", nil
		},
	}

	router := setupTestApp(mockUser, mockHydra)

	form := url.Values{}
	form.Add("challenge", "challenge123")
	form.Add("email", "test@example.com")
	form.Add("password", "password123")
	form.Add("remember", "on")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected status 302, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location != "http://callback" {
		t.Errorf("Expected redirect to http://callback, got %s", location)
	}
}

func TestLoginPost_InvalidCredentials(t *testing.T) {
	mockUser := &MockUserRepository{
		VerifyUserFunc: func(ctx context.Context, email, password string) (*domain.User, error) {
			return nil, fmt.Errorf("invalid credentials")
		},
	}

	router := setupTestApp(mockUser, &MockHydraRepository{})

	form := url.Values{}
	form.Add("challenge", "challenge123")
	form.Add("email", "test@example.com")
	form.Add("password", "wrongpassword")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}
