package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port        string `mapstructure:"port"`
	HTTPPort    string `mapstructure:"http_port"`
	MetricsPort string `mapstructure:"metrics_port"`
	LogLevel    string `mapstructure:"log_level"`

	DBHost     string `mapstructure:"db_host"`
	DBPort     string `mapstructure:"db_port"`
	DBUser     string `mapstructure:"db_user"`
	DBPassword string `mapstructure:"db_password"`
	DBName     string `mapstructure:"db_name"`
	DBSSLMode  string `mapstructure:"db_sslmode"`

	UserServiceAddr       string `mapstructure:"user_service_addr"`
	MarketDataServiceAddr string `mapstructure:"marketdata_service_addr"`

	NatsURL          string `mapstructure:"nats_url"`
	TransactionTopic string `mapstructure:"transaction_topic"`
}

func LoadConfig() Config {
	// 1. Set Defaults
	viper.SetDefault("port", "50053")
	viper.SetDefault("http_port", "8081")
	viper.SetDefault("metrics_port", "9097")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("nats_url", "nats://localhost:4222")
	viper.SetDefault("transaction_topic", "transaction-service.transaction.created")
	viper.SetDefault("user_service_addr", "localhost:50051")
	viper.SetDefault("marketdata_service_addr", "localhost:50054")

	// 2. Load Config File
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/app/")
	viper.AddConfigPath("/etc/portfolio-insights/transaction-service/")

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
	viper.BindEnv("http_port", "HTTP_PORT")
	viper.BindEnv("metrics_port", "METRICS_PORT")
	viper.BindEnv("log_level", "LOG_LEVEL")

	viper.BindEnv("db_host", "DB_HOST")
	viper.BindEnv("db_port", "DB_PORT")
	viper.BindEnv("db_user", "DB_USER")
	viper.BindEnv("db_password", "DB_PASSWORD")
	viper.BindEnv("db_name", "DB_NAME")
	viper.BindEnv("db_sslmode", "DB_SSLMODE")

	viper.BindEnv("user_service_addr", "USER_SERVICE_ADDR")
	viper.BindEnv("marketdata_service_addr", "MARKETDATA_SERVICE_ADDR")
	viper.BindEnv("nats_url", "NATS_URL")
	viper.BindEnv("transaction_topic", "TRANSACTION_TOPIC")

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
