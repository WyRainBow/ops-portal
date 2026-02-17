package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/WyRainBow/ops-portal/internal/ai/errors"
)

// Validator validates configuration.
type Validator struct {
	errors []string
}

// NewValidator creates a new validator.
func NewValidator() *Validator {
	return &Validator{
		errors: make([]string, 0),
	}
}

// Required checks that a required environment variable is set and non-empty.
func (v *Validator) Required(name string) *Validator {
	value := os.Getenv(name)
	if strings.TrimSpace(value) == "" {
		v.errors = append(v.errors, fmt.Sprintf("%s: required but not set", name))
	}
	return v
}

// Optional checks an optional variable with a default value.
func (v *Validator) Optional(name, defaultValue string) string {
	value := os.Getenv(name)
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}

// Int validates an integer environment variable.
func (v *Validator) Int(name string, min, max int64) int64 {
	valueStr := os.Getenv(name)
	if strings.TrimSpace(valueStr) == "" {
		v.errors = append(v.errors, fmt.Sprintf("%s: required but not set", name))
		return 0
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		v.errors = append(v.errors, fmt.Sprintf("%s: must be a valid integer", name))
		return 0
	}

	if value < min || value > max {
		v.errors = append(v.errors, fmt.Sprintf("%s: must be between %d and %d", name, min, max))
	}

	return value
}

// Port validates a port number.
func (v *Validator) Port(name string, defaultValue int) int {
	valueStr := os.Getenv(name)
	if strings.TrimSpace(valueStr) == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		v.errors = append(v.errors, fmt.Sprintf("%s: must be a valid port number", name))
		return defaultValue
	}

	if value < 1 || value > 65535 {
		v.errors = append(v.errors, fmt.Sprintf("%s: must be between 1 and 65535", name))
		return defaultValue
	}

	return value
}

// URL validates a URL environment variable.
func (v *Validator) URL(name string) string {
	value := os.Getenv(name)
	if strings.TrimSpace(value) == "" {
		v.errors = append(v.errors, fmt.Sprintf("%s: required but not set", name))
		return ""
	}

	// Basic URL validation
	if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		v.errors = append(v.errors, fmt.Sprintf("%s: must be a valid URL (http:// or https://)", name))
	}

	return value
}

// OneOf checks that the variable value is one of the allowed values.
func (v *Validator) OneOf(name string, allowed []string, defaultValue string) string {
	value := os.Getenv(name)
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}

	for _, allowedValue := range allowed {
		if value == allowedValue {
			return value
		}
	}

	v.errors = append(v.errors, fmt.Sprintf("%s: must be one of %v", name, allowed))
	return defaultValue
}

// Error returns an error if validation failed.
func (v *Validator) Error() error {
	if len(v.errors) == 0 {
		return nil
	}

	return fmt.Errorf("configuration validation failed:\n  - %s",
		strings.Join(v.errors, "\n  - "))
}

// Validate validates all configuration and returns an error if invalid.
func Validate(ctx context.Context) error {
	v := NewValidator()

	// Server configuration
	serverPort := v.Port("OPS_PORTAL_PORT", 6872)
	_ = serverPort // Use the validated port

	// Database configuration
	v.Required("DATABASE_HOST")
	v.Port("DATABASE_PORT", 5432)
	v.Required("DATABASE_USER")
	v.Required("DATABASE_PASSWORD")
	v.Required("DATABASE_NAME")

	// JWT configuration
	jwtSecret := v.Optional("OPS_PORTAL_JWT_SECRET", "")
	if jwtSecret == "change-me-in-production" || jwtSecret == "" {
		v.errors = append(v.errors, "OPS_PORTAL_JWT_SECRET: must be set to a secure value in production")
	}

	// Observability configuration
	v.Optional("OBS_LOKI_URL", "http://127.0.0.1:3100")
	v.Optional("OBS_PROM_URL", "http://127.0.0.1:9090")
	v.Optional("OBS_GRAFANA_URL", "http://127.0.0.1:3000")

	// LLM configuration
	llmProvider := v.OneOf("LLM_PROVIDER", []string{"doubao", "openai", "claude"}, "doubao")
	_ = llmProvider

	if llmProvider == "doubao" {
		v.Required("DOUBAO_API_KEY")
		v.Required("DOUBAO_BASE_URL")
	}

	// Milvus configuration (for RAG)
	v.Optional("MILVUS_HOST", "127.0.0.1")
	v.Port("MILVUS_PORT", 19530)

	// Redis configuration (optional, for caching)
	v.Optional("REDIS_ADDR", "")
	v.Port("REDIS_PORT", 6379)

	// Alertmanager configuration (optional)
	v.Optional("ALERTMANAGER_URL", "")

	// Feishu configuration (optional, for notifications)
	v.Optional("FEISHU_APP_ID", "")
	v.Optional("FEISHU_APP_SECRET", "")

	// Log configuration
	logLevel := v.OneOf("LOG_LEVEL", []string{"debug", "info", "warn", "error"}, "info")
	_ = logLevel

	return v.Error()
}

// ValidateCritical validates only critical configuration for startup.
// Use this for minimal validation to allow the service to start.
func ValidateCritical(ctx context.Context) error {
	v := NewValidator()

	// Only validate what's absolutely required for basic operation
	v.Required("DATABASE_HOST")
	v.Required("DATABASE_USER")
	v.Required("DATABASE_NAME")

	// JWT secret
	jwtSecret := v.Optional("OPS_PORTAL_JWT_SECRET", "")
	if jwtSecret == "" {
		v.errors = append(v.errors, "OPS_PORTAL_JWT_SECRET: must be set")
	}

	if err := v.Error(); err != nil {
		return err
	}

	errors.Info("config", "critical configuration validated successfully")
	return nil
}

// Config holds validated configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	LLM      LLMConfig
	Milvus   MilvusConfig
	Cache    CacheConfig
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	Port int
	Host string
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	Secret     string
	ExpireTime int // hours
}

// LLMConfig holds LLM configuration.
type LLMConfig struct {
	Provider string
	APIKey   string
	BaseURL  string
	Model    string
}

// MilvusConfig holds Milvus configuration.
type MilvusConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

// CacheConfig holds cache configuration.
type CacheConfig struct {
	Type  string // "memory" or "redis"
	Redis RedisConfig
	TTL   int // seconds
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// Load loads and validates configuration.
func Load(ctx context.Context) (*Config, error) {
	// Validate configuration
	if err := ValidateCritical(ctx); err != nil {
		return nil, err
	}

	v := NewValidator()
	config := &Config{
		Server: ServerConfig{
			Port: int(v.Port("OPS_PORTAL_PORT", 6872)),
			Host: v.Optional("OPS_PORTAL_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     os.Getenv("DATABASE_HOST"),
			Port:     int(v.Port("DATABASE_PORT", 5432)),
			User:     os.Getenv("DATABASE_USER"),
			Password: os.Getenv("DATABASE_PASSWORD"),
			Name:     os.Getenv("DATABASE_NAME"),
			SSLMode:  v.Optional("DATABASE_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:     os.Getenv("OPS_PORTAL_JWT_SECRET"),
			ExpireTime: int(v.Int("JWT_EXPIRE_HOURS", 1, 168)), // 1-168 hours
		},
		LLM: LLMConfig{
			Provider: v.OneOf("LLM_PROVIDER", []string{"doubao", "openai", "claude"}, "doubao"),
			APIKey:   os.Getenv("LLM_API_KEY"),
			BaseURL:  os.Getenv("LLM_BASE_URL"),
			Model:    v.Optional("LLM_MODEL", ""),
		},
		Milvus: MilvusConfig{
			Host:     v.Optional("MILVUS_HOST", "127.0.0.1"),
			Port:     int(v.Port("MILVUS_PORT", 19530)),
			Username: v.Optional("MILVUS_USERNAME", ""),
			Password: v.Optional("MILVUS_PASSWORD", ""),
		},
		Cache: CacheConfig{
			Type: v.OneOf("CACHE_TYPE", []string{"memory", "redis"}, "memory"),
			TTL:  int(v.Int("CACHE_TTL_SECONDS", 1, 3600)),
		},
	}

	if err := v.Error(); err != nil {
		return nil, err
	}

	// Log configuration summary (without secrets)
	errors.Info("config", fmt.Sprintf("loaded: server=%s:%d, db=%s:%d/%s, llm=%s",
		config.Server.Host, config.Server.Port,
		config.Database.Host, config.Database.Port, config.Database.Name,
		config.LLM.Provider))

	return config, nil
}
