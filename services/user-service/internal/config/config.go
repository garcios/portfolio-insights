// Package config provides configuration loading for the user service.
package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	Port        string `mapstructure:"port"`
	MetricsPort string `mapstructure:"metrics_port"`
	LogLevel    string `mapstructure:"log_level"`

	DBHost     string `mapstructure:"db_host"`
	DBPort     string `mapstructure:"db_port"`
	DBUser     string `mapstructure:"db_user"`
	DBPassword string `mapstructure:"db_password"`
	DBName     string `mapstructure:"db_name"`
	DBSSLMode  string `mapstructure:"db_sslmode"`
}

// LoadConfig loads the configuration from file and environment variables.
func LoadConfig() Config {
	// 1. Set Defaults
	viper.SetDefault("port", "50051")
	viper.SetDefault("metrics_port", "9096")
	viper.SetDefault("log_level", "info")

	// 2. Load Config File
	appConfigPath := os.Getenv("APP_CONFIG_PATH")
	log.Printf("***App Config Path: %s\n", appConfigPath)

	appEnv := os.Getenv("APP_ENV")
	log.Printf("***App Env: %s\n", appEnv)

	if appEnv != "" {
		viper.SetConfigName(strings.ToLower(appEnv))
	} else {
		viper.SetConfigName("config")
	}

	viper.SetConfigType("yaml")

	if appConfigPath != "" {
		viper.AddConfigPath(appConfigPath)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/app/")
	}

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
		"port":         "PORT",
		"metrics_port": "METRICS_PORT",
		"log_level":    "LOG_LEVEL",
		"db_host":      "DB_HOST",
		"db_port":      "DB_PORT",
		"db_user":      "DB_USER",
		"db_password":  "DB_PASSWORD",
		"db_name":      "DB_NAME",
		"db_sslmode":   "DB_SSLMODE",
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
