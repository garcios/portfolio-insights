package config

import "os"

type Config struct {
	Port            string
	HydraAdminURL   string
	UserServiceAddr string
	SessionSecret   string
	LogLevel        string
}

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
