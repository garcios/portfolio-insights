package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

type App struct {
	userClient   UserServiceClientInterface
	hydraAdmin   string
	sessionStore cookie.Store
	httpClient   *http.Client
}

func main() {
	// Load configuration
	port := getEnv("PORT", "3001")
	hydraAdminURL := getEnv("HYDRA_ADMIN_URL", "http://localhost:4445")
	userServiceAddr := getEnv("USER_SERVICE_ADDR", "localhost:50051")
	sessionSecret := getEnv("SESSION_SECRET", "change-this-secret")
	logLevel := getEnv("LOG_LEVEL", "info")

	// Set Gin mode
	if logLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to user service
	userClient, err := NewUserServiceClient(userServiceAddr)
	if err != nil {
		log.Fatalf("Unable to connect to user service: %v", err)
	}
	defer userClient.Close()

	// Initialize app
	app := &App{
		userClient:   userClient,
		hydraAdmin:   hydraAdminURL,
		sessionStore: cookie.NewStore([]byte(sessionSecret)),
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}

	// Setup router
	router := gin.Default()

	// Session middleware
	router.Use(sessions.Sessions("auth_session", app.sessionStore))

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Routes
	router.GET("/health", app.healthCheck)
	router.GET("/login", app.loginGet)
	router.POST("/login", app.loginPost)
	router.GET("/consent", app.consentGet)
	router.POST("/consent", app.consentPost)
	router.GET("/logout", app.logoutGet)
	router.POST("/logout", app.logoutPost)
	router.GET("/error", app.errorGet)

	// Static files for UI
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Login/Consent Provider starting on %s", addr)
	log.Printf("Hydra Admin URL: %s", hydraAdminURL)

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func (app *App) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
