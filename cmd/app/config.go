package app

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the complete application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Email    EmailConfig    `yaml:"email"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Port         int    `yaml:"port"`
	Host         string `yaml:"host"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	RateLimit    int    `yaml:"rate_limit"` // RateLimiter max tokens per minute
}

// DatabaseConfig contains connection settings for MySQL database
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	MaxConns int    `yaml:"max_conns"`
	MinConns int    `yaml:"min_conns"`
}

// RedisConfig contains connection settings and cache configuration for Redis
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	CacheTTL int    `yaml:"cache_ttl"` // Cache TTL in minutes
}

// JWTConfig contains JWT token configuration
type JWTConfig struct {
	Secret     string `yaml:"secret"`
	Expiration int    `yaml:"expiration"` // in hours
}

// EmailConfig contains email service configuration
type EmailConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// Load reads and parses the configuration from a YAML file
// Environment variables can override values from the config file
func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables if set
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		cfg.Database.Port = parseInt(port)
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Database.Password = password
	}
	if redis := os.Getenv("REDIS_HOST"); redis != "" {
		cfg.Redis.Host = redis
	}
	if redisPort := os.Getenv("REDIS_PORT"); redisPort != "" {
		cfg.Redis.Port = parseInt(redisPort)
	}
	if cacheTTL := os.Getenv("CACHE_TTL"); cacheTTL != "" {
		cfg.Redis.CacheTTL = parseInt(cacheTTL)
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cfg.JWT.Secret = jwtSecret
	}

	return &cfg, nil
}

// parseInt safely converts a string to an integer, returning 0 if conversion fails
func parseInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i) //nolint:errcheck
	return i
}
