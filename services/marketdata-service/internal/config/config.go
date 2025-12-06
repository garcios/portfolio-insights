package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

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

	EodhdApiToken   string `mapstructure:"eodhd_api_token"`
	EodhdApiBaseUrl string `mapstructure:"eodhd_api_base_url"`

	CurrencySyncInterval       time.Duration `mapstructure:"currency_sync_interval"`
	CurrencyStaleDuration      time.Duration `mapstructure:"currency_stale_duration"`
	CurrencySyncBatchSize      int           `mapstructure:"currency_sync_batch_size"`
	CurrencySyncMaxConcurrency int           `mapstructure:"currency_sync_max_concurrency"`
	CurrencySyncHistoricalDays int           `mapstructure:"currency_sync_historical_days"`

	PriceSyncInterval       time.Duration `mapstructure:"price_sync_interval"`
	PriceStaleDuration      time.Duration `mapstructure:"price_stale_duration"`
	PriceSyncBatchSize      int           `mapstructure:"price_sync_batch_size"`
	PriceSyncMaxConcurrency int           `mapstructure:"price_sync_max_concurrency"`
	PriceSyncHistoricalDays int           `mapstructure:"price_sync_historical_days"`
	EodhdRateLimit          float64       `mapstructure:"eodhd_rate_limit"`

	MinioEndpoint   string `mapstructure:"minio_endpoint"`
	MinioAccessKey  string `mapstructure:"minio_access_key"`
	MinioSecretKey  string `mapstructure:"minio_secret_key"`
	MinioUseSSL     bool   `mapstructure:"minio_use_ssl"`
	MinioBucketName string `mapstructure:"minio_bucket_name"`
}

func LoadConfig() Config {
	// 1. Set Defaults
	viper.SetDefault("port", "50054")
	viper.SetDefault("metrics_port", "9099")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("eodhd_api_base_url", "https://eodhd.com/api")
	viper.SetDefault("currency_sync_interval", 1*time.Hour)
	viper.SetDefault("currency_stale_duration", 24*time.Hour)
	viper.SetDefault("currency_sync_batch_size", 10)
	viper.SetDefault("currency_sync_max_concurrency", 2)
	viper.SetDefault("currency_sync_historical_days", 30)

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
	viper.BindEnv("port", "PORT")
	viper.BindEnv("metrics_port", "METRICS_PORT")
	viper.BindEnv("log_level", "LOG_LEVEL")

	viper.BindEnv("db_host", "DB_HOST")
	viper.BindEnv("db_port", "DB_PORT")
	viper.BindEnv("db_user", "DB_USER")
	viper.BindEnv("db_password", "DB_PASSWORD")
	viper.BindEnv("db_name", "DB_NAME")
	viper.BindEnv("db_sslmode", "DB_SSLMODE")

	viper.BindEnv("eodhd_api_token", "EODHD_API_TOKEN")
	viper.BindEnv("eodhd_api_base_url", "EODHD_API_BASE_URL")

	viper.BindEnv("currency_sync_interval", "CURRENCY_SYNC_INTERVAL")
	viper.BindEnv("currency_stale_duration", "CURRENCY_STALE_DURATION")
	viper.BindEnv("currency_sync_batch_size", "CURRENCY_SYNC_BATCH_SIZE")
	viper.BindEnv("currency_sync_max_concurrency", "CURRENCY_SYNC_MAX_CONCURRENCY")
	viper.BindEnv("currency_sync_historical_days", "CURRENCY_SYNC_HISTORICAL_DAYS")

	viper.BindEnv("price_sync_interval", "PRICE_SYNC_INTERVAL")
	viper.BindEnv("price_stale_duration", "PRICE_STALE_DURATION")
	viper.BindEnv("price_sync_batch_size", "PRICE_SYNC_BATCH_SIZE")
	viper.BindEnv("price_sync_max_concurrency", "PRICE_SYNC_MAX_CONCURRENCY")
	viper.BindEnv("price_sync_historical_days", "PRICE_SYNC_HISTORICAL_DAYS")
	viper.BindEnv("eodhd_rate_limit", "EODHD_RATE_LIMIT")

	viper.BindEnv("minio_endpoint", "MINIO_ENDPOINT")
	viper.BindEnv("minio_access_key", "MINIO_ACCESS_KEY")
	viper.BindEnv("minio_secret_key", "MINIO_SECRET_KEY")
	viper.BindEnv("minio_use_ssl", "MINIO_USE_SSL")
	viper.BindEnv("minio_bucket_name", "MINIO_BUCKET_NAME")

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
