// Package config provides configuration loading for the login-consent-provider.
package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	Port            string `mapstructure:"port"`
	HydraAdminURL   string `mapstructure:"hydra_admin_url"`
	UserServiceAddr string `mapstructure:"user_service_addr"`
	SessionSecret   string `mapstructure:"session_secret"`
	LogLevel        string `mapstructure:"log_level"`
}

// LoadConfig loads the configuration from file and environment variables.
func LoadConfig() Config {
	// 1. Set Defaults
	viper.SetDefault("port", "3002")
	viper.SetDefault("hydra_admin_url", "http://localhost:4445")
	viper.SetDefault("user_service_addr", "localhost:50051")
	viper.SetDefault("session_secret", "change-this-secret")
	viper.SetDefault("log_level", "info")

	// 2. Load Config File
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/app/")

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Warning: config file not found, using defaults and ENV vars.")
		} else {
			log.Fatalf("Fatal error reading config file: %s", err)
		}
	}

	// 3. Bind to Environment Variables
	bindKeys := map[string]string{
		"port":              "PORT",
		"hydra_admin_url":   "HYDRA_ADMIN_URL",
		"user_service_addr": "USER_SERVICE_ADDR",
		"session_secret":    "SESSION_SECRET",
		"log_level":         "LOG_LEVEL",
	}

	for key, env := range bindKeys {
		if err := viper.BindEnv(key, env); err != nil {
			log.Printf("Could not bind env var %s to key %s: %s", env, key, err)
		}
	}

	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// 4. Unmarshal
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to unmarshal config into struct: %s", err)
	}

	return config
}
