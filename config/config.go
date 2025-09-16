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

	Server   ServerConfig   `envPrefix:"SERVER_"`
	CORS     CORSConfig     `envPrefix:"CORS_"`
	Database DatabaseConfig `envPrefix:"DATABASE_"`
	Redis    RedisConfig    `envPrefix:"REDIS_"`
	API      APIConfig      `envPrefix:"API_"`
	Cache    CacheConfig    `envPrefix:"CACHE_"`
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
	User            string        `env:"USER" envDefault:"root"`
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

// ExternalAPIConfig represents configuration for a generic external API
type ExternalAPIConfig struct {
	BaseURL    string        `env:"BASE_URL" envDefault:""`
	Timeout    time.Duration `env:"TIMEOUT" envDefault:"30s"`
	MaxRetries int           `env:"MAX_RETRIES" envDefault:"3"`
	RetryDelay time.Duration `env:"RETRY_DELAY" envDefault:"1s"`
	Headers    map[string]string
	Enable     bool              `env:"ENABLE" envDefault:"false"`
	APIKey     string            `env:"API_KEY"`
	AuthType   string            `env:"AUTH_TYPE" envDefault:"none"` // none, bearer, basic, api_key
	ItemType   string            `env:"ITEM_TYPE" envDefault:"custom"`
	Endpoints  map[string]string // API endpoints for different operations
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

	// Generic External APIs
	ExternalAPIs map[string]ExternalAPIConfig
}

type CacheConfig struct {
	DefaultTTL     time.Duration `env:"DEFAULT_TTL" envDefault:"5m"`
	ItemsCacheTTL  time.Duration `env:"ITEMS_CACHE_TTL" envDefault:"10m"`
	StatusCacheTTL time.Duration `env:"STATUS_CACHE_TTL" envDefault:"5m"`
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

	// Initialize ExternalAPIs map if nil
	if cfg.API.ExternalAPIs == nil {
		cfg.API.ExternalAPIs = make(map[string]ExternalAPIConfig)
	}

	// Initialize FieldMapping and Endpoints maps for each API
	for name, apiConfig := range cfg.API.ExternalAPIs {
		if apiConfig.Endpoints == nil {
			apiConfig.Endpoints = make(map[string]string)
		}
		cfg.API.ExternalAPIs[name] = apiConfig
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

// GetExternalAPIConfig returns the configuration for a specific external API
func (c *APIConfig) GetExternalAPIConfig(name string) (ExternalAPIConfig, bool) {
	config, exists := c.ExternalAPIs[name]
	return config, exists
}

// AddExternalAPIConfig adds or updates an external API configuration
func (c *APIConfig) AddExternalAPIConfig(name string, config ExternalAPIConfig) {
	if c.ExternalAPIs == nil {
		c.ExternalAPIs = make(map[string]ExternalAPIConfig)
	}
	c.ExternalAPIs[name] = config
}
