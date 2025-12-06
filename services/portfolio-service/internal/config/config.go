package config

import (
	"fmt"
	"log"
	"strings"

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

	RedisAddr     string `mapstructure:"redis_addr"`
	RedisPassword string `mapstructure:"redis_password"`
	RedisDB       int    `mapstructure:"redis_db"`

	NatsURL               string `mapstructure:"nats_url"`
	ExchangeRateTopic     string `mapstructure:"exchange_rate_topic"`
	MarketDataServiceAddr string `mapstructure:"marketdata_service_addr"`
	AssetCacheTTL         int    `mapstructure:"asset_cache_ttl_seconds"`
}

func LoadConfig() Config {
	// 1. Set Defaults
	viper.SetDefault("port", "50052")
	viper.SetDefault("metrics_port", "9098")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("nats_url", "nats://localhost:4222")
	viper.SetDefault("exchange_rate_topic", "marketdata.exchange_rates")
	viper.SetDefault("marketdata_service_addr", "localhost:50054")
	viper.SetDefault("redis_addr", "localhost:6379")

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

	viper.BindEnv("redis_addr", "REDIS_ADDR")
	viper.BindEnv("redis_password", "REDIS_PASSWORD")
	viper.BindEnv("redis_db", "REDIS_DB")

	viper.BindEnv("nats_url", "NATS_URL")
	viper.BindEnv("exchange_rate_topic", "EXCHANGE_RATE_TOPIC")
	viper.BindEnv("marketdata_service_addr", "MARKETDATA_SERVICE_ADDR")
	viper.BindEnv("asset_cache_ttl_seconds", "ASSET_CACHE_TTL_SECONDS")

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
