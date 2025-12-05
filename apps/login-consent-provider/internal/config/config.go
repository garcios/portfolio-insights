// Package config functions to load configuration.
package config

import "os"

// Config holds the application configuration.
type Config struct {
	Port            string
	HydraAdminURL   string
	UserServiceAddr string
	SessionSecret   string
	LogLevel        string
}

// Load loads the configuration from environment variables.
func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "3001"),
		HydraAdminURL:   getEnv("HYDRA_ADMIN_URL", "http://localhost:4445"),
		UserServiceAddr: getEnv("USER_SERVICE_ADDR", "localhost:50051"),
		SessionSecret:   getEnv("SESSION_SECRET", "change-this-secret"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
