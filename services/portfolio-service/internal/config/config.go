// Package config provides configuration loading for the portfolio service.
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

	RedisAddr     string `mapstructure:"redis_addr"`
	RedisPassword string `mapstructure:"redis_password"`
	RedisDB       int    `mapstructure:"redis_db"`

	NatsURL                 string        `mapstructure:"nats_url"`
	ExchangeRateTopic       string        `mapstructure:"exchange_rate_topic"`
	TransactionCreatedTopic string        `mapstructure:"transaction_created_topic"`
	TransactionUpdatedTopic string        `mapstructure:"transaction_updated_topic"`
	TransactionDeletedTopic string        `mapstructure:"transaction_deleted_topic"`
	MarketDataServiceAddr   string        `mapstructure:"marketdata_service_addr"`
	TransactionServiceAddr  string        `mapstructure:"transaction_service_addr"`
	AssetCacheTTL           int           `mapstructure:"asset_cache_ttl_seconds"`
	CacheWarmingInterval    string        `mapstructure:"cache_warming_interval"`
	Caching                 CachingConfig `mapstructure:"caching"`
}

// CachingConfig holds configuration for portfolio caching
type CachingConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	SummaryTTLSeconds int  `mapstructure:"summary_ttl_seconds"`
	HistoryTTLSeconds int  `mapstructure:"history_ttl_seconds"`
}

// LoadConfig loads the configuration from file and environment variables.
func LoadConfig() Config {
	// 1. Set Defaults
	viper.SetDefault("port", "50052")
	viper.SetDefault("metrics_port", "9098")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("nats_url", "nats://localhost:4222")
	viper.SetDefault("exchange_rate_topic", "marketdata.exchange_rates")
	viper.SetDefault("transaction_created_topic", "transaction-service.transaction.created")
	viper.SetDefault("transaction_updated_topic", "transaction-service.transaction.updated")
	viper.SetDefault("transaction_deleted_topic", "transaction-service.transaction.deleted")
	viper.SetDefault("marketdata_service_addr", "localhost:50054")
	viper.SetDefault("transaction_service_addr", "localhost:50053")
	viper.SetDefault("redis_addr", "localhost:6379")
	viper.SetDefault("caching.enabled", true)
	viper.SetDefault("caching.summary_ttl_seconds", 300)
	viper.SetDefault("caching.history_ttl_seconds", 86400)
	viper.SetDefault("cache_warming_interval", "6h")

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
		"port":                        "PORT",
		"metrics_port":                "METRICS_PORT",
		"log_level":                   "LOG_LEVEL",
		"db_host":                     "DB_HOST",
		"db_port":                     "DB_PORT",
		"db_user":                     "DB_USER",
		"db_password":                 "DB_PASSWORD",
		"db_name":                     "DB_NAME",
		"db_sslmode":                  "DB_SSLMODE",
		"redis_addr":                  "REDIS_ADDR",
		"redis_password":              "REDIS_PASSWORD",
		"redis_db":                    "REDIS_DB",
		"nats_url":                    "NATS_URL",
		"exchange_rate_topic":         "EXCHANGE_RATE_TOPIC",
		"transaction_created_topic":   "TRANSACTION_CREATED_TOPIC",
		"transaction_updated_topic":   "TRANSACTION_UPDATED_TOPIC",
		"transaction_deleted_topic":   "TRANSACTION_DELETED_TOPIC",
		"marketdata_service_addr":     "MARKETDATA_SERVICE_ADDR",
		"transaction_service_addr":    "TRANSACTION_SERVICE_ADDR",
		"asset_cache_ttl_seconds":     "ASSET_CACHE_TTL_SECONDS",
		"caching.enabled":             "CACHING_ENABLED",
		"caching.summary_ttl_seconds": "CACHING_SUMMARY_TTL_SECONDS",
		"caching.history_ttl_seconds": "CACHING_HISTORY_TTL_SECONDS",
		"cache_warming_interval":      "CACHE_WARMING_INTERVAL",
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
