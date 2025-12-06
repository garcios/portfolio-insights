package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port            string `mapstructure:"port"`
	HydraAdminURL   string `mapstructure:"hydra_admin_url"`
	UserServiceAddr string `mapstructure:"user_service_addr"`
	SessionSecret   string `mapstructure:"session_secret"`
	LogLevel        string `mapstructure:"log_level"`
}

func LoadConfig() Config {
	// 1. Set Defaults
	viper.SetDefault("port", "3001")
	viper.SetDefault("hydra_admin_url", "http://localhost:4445")
	viper.SetDefault("user_service_addr", "localhost:50051")
	viper.SetDefault("session_secret", "change-this-secret")
	viper.SetDefault("log_level", "info")

	// 2. Load Config File
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/app/")
	viper.AddConfigPath("/etc/portfolio-insights/login-consent-provider/")

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Warning: config file not found, using defaults and ENV vars.")
		} else {
			log.Fatalf("Fatal error reading config file: %s", err)
		}
	}

	// 3. Bind to Environment Variables
	viper.BindEnv("port", "PORT")
	viper.BindEnv("hydra_admin_url", "HYDRA_ADMIN_URL")
	viper.BindEnv("user_service_addr", "USER_SERVICE_ADDR")
	viper.BindEnv("session_secret", "SESSION_SECRET")
	viper.BindEnv("log_level", "LOG_LEVEL")

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
