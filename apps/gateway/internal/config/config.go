// Package config provides configuration loading for the gateway.
package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	Port                       string `mapstructure:"port"`
	PortfolioServiceAddr       string `mapstructure:"portfolio_service_addr"`
	UserServiceAddr            string `mapstructure:"user_service_addr"`
	TransactionServiceAddr     string `mapstructure:"transaction_service_addr"`
	TransactionServiceHTTPAddr string `mapstructure:"transaction_service_http_addr"`
	HydraPublicURL             string `mapstructure:"hydra_public_url"`
	JWKSURL                    string `mapstructure:"jwks_url"`
	JWTIssuer                  string `mapstructure:"jwt_issuer"`
	JWTAudience                string `mapstructure:"jwt_audience"`
	LogLevel                   string `mapstructure:"log_level"`
}

// LoadConfig loads the configuration from file and environment variables.
func LoadConfig() Config {
	// 1. Set Defaults
	viper.SetDefault("port", "8080")
	viper.SetDefault("portfolio_service_addr", "localhost:50052")
	viper.SetDefault("user_service_addr", "localhost:50051")
	viper.SetDefault("transaction_service_addr", "localhost:50053")
	viper.SetDefault("transaction_service_http_addr", "http://localhost:8081")
	viper.SetDefault("jwt_audience", "portfolio-insights-spa")
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
		"port":                          "PORT",
		"portfolio_service_addr":        "PORTFOLIO_SERVICE_ADDR",
		"user_service_addr":             "USER_SERVICE_ADDR",
		"transaction_service_addr":      "TRANSACTION_SERVICE_ADDR",
		"transaction_service_http_addr": "TRANSACTION_SERVICE_HTTP_ADDR",
		"hydra_public_url":              "HYDRA_PUBLIC_URL",
		"jwks_url":                      "JWKS_URL",
		"jwt_issuer":                    "JWT_ISSUER",
		"jwt_audience":                  "JWT_AUDIENCE",
		"log_level":                     "LOG_LEVEL",
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
