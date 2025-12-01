package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/garcios/portfolio-insights/apps/login-consent-provider/internal/config"
	delivery "github.com/garcios/portfolio-insights/apps/login-consent-provider/internal/delivery/http"
	"github.com/garcios/portfolio-insights/apps/login-consent-provider/internal/infrastructure/grpc"
	"github.com/garcios/portfolio-insights/apps/login-consent-provider/internal/infrastructure/hydra"
	"github.com/garcios/portfolio-insights/apps/login-consent-provider/internal/usecase"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to user service
	userClient, err := grpc.NewUserServiceClient(cfg.UserServiceAddr)
	if err != nil {
		log.Fatalf("Unable to connect to user service: %v", err)
	}
	defer userClient.Close()

	// Initialize Hydra client
	httpClient := &http.Client{Timeout: 10 * time.Second}
	hydraClient := hydra.NewHydraClient(cfg.HydraAdminURL, httpClient)

	// Initialize UseCase
	authUseCase := usecase.NewAuthUseCase(userClient, hydraClient)

	// Initialize Handler
	handler := delivery.NewHandler(authUseCase)

	// Initialize Session Store
	sessionStore := cookie.NewStore([]byte(cfg.SessionSecret))

	// Setup Router
	router := delivery.NewRouter(handler, sessionStore)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Login/Consent Provider starting on %s", addr)
	log.Printf("Hydra Admin URL: %s", cfg.HydraAdminURL)

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
