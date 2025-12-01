package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// MockDB implements DBExecutor for testing
type MockDB struct {
	QueryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
}

func (m *MockDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(ctx, sql, args...)
	}
	return nil
}

// MockRow implements pgx.Row for testing
type MockRow struct {
	ScanFunc func(dest ...any) error
}

func (m *MockRow) Scan(dest ...any) error {
	if m.ScanFunc != nil {
		return m.ScanFunc(dest...)
	}
	return nil
}

func setupTestApp() (*App, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	app := &App{
		db:           &MockDB{},
		sessionStore: cookie.NewStore([]byte("secret")),
		httpClient:   &http.Client{},
	}

	router := gin.Default()
	router.Use(sessions.Sessions("auth_session", app.sessionStore))
	router.LoadHTMLGlob("templates/*")

	return app, router
}

func TestHealthCheck(t *testing.T) {
	app, router := setupTestApp()
	router.GET("/health", app.healthCheck)

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
	app, router := setupTestApp()
	router.GET("/login", app.loginGet)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/login", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestLoginGet_Success(t *testing.T) {
	// Mock Hydra Server
	hydraServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/admin/oauth2/auth/requests/login" {
			json.NewEncoder(w).Encode(LoginRequest{
				Challenge: "challenge123",
				Skip:      false,
				Subject:   "",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer hydraServer.Close()

	app, router := setupTestApp()
	app.hydraAdmin = hydraServer.URL
	router.GET("/login", app.loginGet)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/login?login_challenge=challenge123", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check if the login page was rendered (checking for some content)
	// Note: Since we are using LoadHTMLGlob, we need the templates to exist.
	// In the test environment, we might need to point to the correct path or mock the renderer.
	// However, gin's LoadHTMLGlob relies on actual files.
	// Assuming the test runs from apps/login-consent-provider, templates/* should be found.
}

func TestLoginPost_Success(t *testing.T) {
	// Mock Hydra Server
	hydraServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/admin/oauth2/auth/requests/login/accept" {
			json.NewEncoder(w).Encode(AcceptLoginResponse{
				RedirectTo: "http://callback",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer hydraServer.Close()

	app, router := setupTestApp()
	app.hydraAdmin = hydraServer.URL
	router.POST("/login", app.loginPost)

	// Mock DB to return a user
	mockDB := app.db.(*MockDB)
	mockDB.QueryRowFunc = func(ctx context.Context, sql string, args ...any) pgx.Row {
		return &MockRow{
			ScanFunc: func(dest ...any) error {
				// dest[0] = &user.ID
				// dest[1] = &user.Email
				// dest[2] = &user.PasswordHash
				// dest[3] = &user.Username

				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

				if len(dest) >= 4 {
					*dest[0].(*string) = "user123"
					*dest[1].(*string) = "test@example.com"
					*dest[2].(*string) = string(hashedPassword)
					*dest[3].(*string) = "testuser"
				}
				return nil
			},
		}
	}

	// Create form data
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
	app, router := setupTestApp()
	router.POST("/login", app.loginPost)

	// Mock DB to return a user
	mockDB := app.db.(*MockDB)
	mockDB.QueryRowFunc = func(ctx context.Context, sql string, args ...any) pgx.Row {
		return &MockRow{
			ScanFunc: func(dest ...any) error {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

				if len(dest) >= 4 {
					*dest[0].(*string) = "user123"
					*dest[1].(*string) = "test@example.com"
					*dest[2].(*string) = string(hashedPassword)
					*dest[3].(*string) = "testuser"
				}
				return nil
			},
		}
	}

	// Create form data with wrong password
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
