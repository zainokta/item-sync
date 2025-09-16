package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/labstack/echo/v4/middleware"
	"github.com/zainokta/item-sync/pkg/logger"
)

type Config struct {
	Environment string `env:"ENV" envDefault:"development"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`

	Server    ServerConfig    `envPrefix:"SERVER_"`
	CORS      CORSConfig      `envPrefix:"CORS_"`
	Database  DatabaseConfig  `envPrefix:"DATABASE_"`
	Redis     RedisConfig     `envPrefix:"REDIS_"`
	API       APIConfig       `envPrefix:"API_"`
	Cache     CacheConfig     `envPrefix:"CACHE_"`
	Worker    WorkerConfig    `envPrefix:"WORKER_"`
	Retry     RetryConfig     `envPrefix:"RETRY_"`
	Migration MigrationConfig `envPrefix:"MIGRATION_"`
}

type ServerConfig struct {
	Host            string        `env:"HOST" envDefault:"0.0.0.0"`
	Port            int           `env:"PORT" envDefault:"8080"`
	GracefulTimeout time.Duration `env:"GRACEFUL_TIMEOUT" envDefault:"10s"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" envDefault:"30s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" envDefault:"120s"`
	MaxRequestSize  int64         `env:"MAX_REQUEST_SIZE"`
}

type CORSConfig struct {
	AllowOrigins     string `env:"ALLOW_ORIGINS" envDefault:"*"`
	AllowHeaders     string `env:"ALLOW_HEADERS" envDefault:"Origin,Content-Type,Accept,Authorization,X-Requested-With"`
	AllowMethods     string `env:"ALLOW_METHODS" envDefault:"GET,POST,PUT,DELETE,OPTIONS,HEAD,PATCH"`
	AllowCredentials bool   `env:"ALLOW_CREDENTIALS" envDefault:"true"`
	MaxAge           int    `env:"MAX_AGE" envDefault:"86400"` // 24 hours in seconds
	ExposeHeaders    string `env:"EXPOSE_HEADERS" envDefault:""`
}

type DatabaseConfig struct {
	Host            string        `env:"HOST" envDefault:"localhost"`
	Port            int           `env:"PORT" envDefault:"3306"`
	User            string        `env:"USER" envDefault:"appuser"`
	Password        string        `env:"PASSWORD"`
	Database        string        `env:"DATABASE" envDefault:"item_sync"`
	MaxOpenConns    int           `env:"MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns    int           `env:"MAX_IDLE_CONNS" envDefault:"25"`
	ConnMaxLifetime time.Duration `env:"CONN_MAX_LIFETIME" envDefault:"5m"`
}

type RedisConfig struct {
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     int    `env:"PORT" envDefault:"6379"`
	Password string `env:"PASSWORD"`
	DB       int    `env:"DB" envDefault:"0"`
}

type APIConfig struct {
	APIType    string        `env:"API_TYPE" envDefault:"pokemon"`
	Timeout    time.Duration `env:"TIMEOUT" envDefault:"30s"`
	MaxRetries int           `env:"MAX_RETRIES" envDefault:"3"`
	RetryDelay time.Duration `env:"RETRY_DELAY" envDefault:"1s"`
	RateLimit  int           `env:"RATE_LIMIT" envDefault:"100"` // requests per minute

	// HTTP Transport Configuration
	MaxIdleConns        int           `env:"MAX_IDLE_CONNS" envDefault:"10"`
	IdleConnTimeout     time.Duration `env:"IDLE_CONN_TIMEOUT" envDefault:"30s"`
	DisableCompression  bool          `env:"DISABLE_COMPRESSION" envDefault:"false"`
	MaxIdleConnsPerHost int           `env:"MAX_IDLE_CONNS_PER_HOST" envDefault:"10"`

	// OpenWeather API Key (when using openweather API type)
	OpenWeatherAPIKey string `env:"OPENWEATHER_API_KEY"`
}

type CacheConfig struct {
	DefaultTTL     time.Duration `env:"DEFAULT_TTL" envDefault:"5m"`
	ItemsCacheTTL  time.Duration `env:"ITEMS_CACHE_TTL" envDefault:"10m"`
	StatusCacheTTL time.Duration `env:"STATUS_CACHE_TTL" envDefault:"5m"`
}

type WorkerConfig struct {
	Enabled      bool          `env:"ENABLED" envDefault:"true"`
	SyncInterval time.Duration `env:"SYNC_INTERVAL" envDefault:"15m"`
	JobTimeout   time.Duration `env:"JOB_TIMEOUT" envDefault:"10m"`
	MaxWorkers   int           `env:"MAX_WORKERS" envDefault:"5"`
}

type RetryConfig struct {
	MaxRetries       int           `env:"MAX_RETRIES" envDefault:"5"`
	InitialDelay     time.Duration `env:"INITIAL_DELAY" envDefault:"1s"`
	MaxDelay         time.Duration `env:"MAX_DELAY" envDefault:"30s"`
	BackoffFactor    float64       `env:"BACKOFF_FACTOR" envDefault:"2.0"`
	CircuitThreshold int           `env:"CIRCUIT_THRESHOLD" envDefault:"5"`
	CircuitTimeout   time.Duration `env:"CIRCUIT_TIMEOUT" envDefault:"60s"`
}

type MigrationConfig struct {
	Enabled        bool   `env:"ENABLED" envDefault:"true"`
	MigrationsPath string `env:"MIGRATIONS_PATH" envDefault:"./migrations"`
	FailOnError    bool   `env:"FAIL_ON_ERROR" envDefault:"true"`
}

func LoadConfig() (*Config, error) {
	environment := os.Getenv("ENV")
	if environment == "" {
		environment = "development"
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return cfg, nil
}

func (c *Config) CovertLogLevel(logLevel string) logger.LogLevel {
	switch logLevel {
	case "info":
		return logger.LevelInfo
	case "debug":
		return logger.LevelDebug
	case "error":
		return logger.LevelError
	case "warn":
		return logger.LevelWarn
	}

	return logger.LevelInfo
}

func (c CORSConfig) ToEchoCORSConfig() middleware.CORSConfig {
	return middleware.CORSConfig{
		AllowOrigins:     parseStringSlice(c.AllowOrigins),
		AllowHeaders:     parseStringSlice(c.AllowHeaders),
		AllowMethods:     parseStringSlice(c.AllowMethods),
		AllowCredentials: c.AllowCredentials,
		MaxAge:           c.MaxAge,
		ExposeHeaders:    parseStringSlice(c.ExposeHeaders),
	}
}

func parseStringSlice(s string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
