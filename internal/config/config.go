package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	JWT         JWTConfig
	CORS        CORSConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host         string
	Port         string
	ReadTimeout  int
	WriteTimeout int
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

// CORSConfig holds CORS-related configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	config := &Config{
		Environment: getEnv("APP_ENV", "development"),
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "localhost"),
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getEnvAsInt("SERVER_READ_TIMEOUT", 30),
			WriteTimeout: getEnvAsInt("SERVER_WRITE_TIMEOUT", 30),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "gin_app"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "your-secret-key"),
			ExpiryHours: getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
			AllowedMethods: getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders: getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization"}),
		},
	}

	// Apply environment-specific overrides
	applyEnvironmentDefaults(config)

	return config
}

// applyEnvironmentDefaults applies environment-specific configuration defaults
func applyEnvironmentDefaults(config *Config) {
	switch config.Environment {
	case "production":
		// Production-specific defaults
		if config.Server.Host == "localhost" {
			config.Server.Host = "0.0.0.0"
		}
		if config.Database.SSLMode == "disable" {
			config.Database.SSLMode = "require"
		}
	case "staging":
		// Staging-specific defaults
		if config.Server.Host == "localhost" {
			config.Server.Host = "0.0.0.0"
		}
		if config.Database.SSLMode == "disable" {
			config.Database.SSLMode = "prefer"
		}
	case "development":
		// Development-specific defaults (already set in Load())
		break
	default:
		log.Printf("Unknown environment: %s, using development defaults", config.Environment)
		config.Environment = "development"
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Invalid integer value for %s: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}

// getEnvAsSlice gets an environment variable as slice (comma-separated) or returns a default value
func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsStaging returns true if the environment is staging
func (c *Config) IsStaging() bool {
	return c.Environment == "staging"
}
