package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test default configuration
	config := Load()

	if config.Environment != "development" {
		t.Errorf("Expected environment to be 'development', got '%s'", config.Environment)
	}

	if config.Server.Host != "localhost" {
		t.Errorf("Expected server host to be 'localhost', got '%s'", config.Server.Host)
	}

	if config.Server.Port != "8080" {
		t.Errorf("Expected server port to be '8080', got '%s'", config.Server.Port)
	}

	if config.Database.Host != "localhost" {
		t.Errorf("Expected database host to be 'localhost', got '%s'", config.Database.Host)
	}
}

func TestLoadWithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("APP_ENV", "production")
	os.Setenv("SERVER_HOST", "0.0.0.0")
	os.Setenv("SERVER_PORT", "9000")
	os.Setenv("DB_HOST", "prod-db")

	defer func() {
		// Clean up
		os.Unsetenv("APP_ENV")
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("DB_HOST")
	}()

	config := Load()

	if config.Environment != "production" {
		t.Errorf("Expected environment to be 'production', got '%s'", config.Environment)
	}

	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Expected server host to be '0.0.0.0', got '%s'", config.Server.Host)
	}

	if config.Server.Port != "9000" {
		t.Errorf("Expected server port to be '9000', got '%s'", config.Server.Port)
	}

	if config.Database.Host != "prod-db" {
		t.Errorf("Expected database host to be 'prod-db', got '%s'", config.Database.Host)
	}
}

func TestEnvironmentMethods(t *testing.T) {
	// Test development environment
	os.Setenv("APP_ENV", "development")
	config := Load()

	if !config.IsDevelopment() {
		t.Error("Expected IsDevelopment() to return true")
	}

	if config.IsProduction() {
		t.Error("Expected IsProduction() to return false")
	}

	if config.IsStaging() {
		t.Error("Expected IsStaging() to return false")
	}

	// Test production environment
	os.Setenv("APP_ENV", "production")
	config = Load()

	if config.IsDevelopment() {
		t.Error("Expected IsDevelopment() to return false")
	}

	if !config.IsProduction() {
		t.Error("Expected IsProduction() to return true")
	}

	if config.IsStaging() {
		t.Error("Expected IsStaging() to return false")
	}

	// Clean up
	os.Unsetenv("APP_ENV")
}

func TestApplyEnvironmentDefaults(t *testing.T) {
	// Test production defaults
	os.Setenv("APP_ENV", "production")
	config := Load()

	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Expected production server host to be '0.0.0.0', got '%s'", config.Server.Host)
	}

	if config.Database.SSLMode != "require" {
		t.Errorf("Expected production SSL mode to be 'require', got '%s'", config.Database.SSLMode)
	}

	// Test staging defaults
	os.Setenv("APP_ENV", "staging")
	config = Load()

	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Expected staging server host to be '0.0.0.0', got '%s'", config.Server.Host)
	}

	if config.Database.SSLMode != "prefer" {
		t.Errorf("Expected staging SSL mode to be 'prefer', got '%s'", config.Database.SSLMode)
	}

	// Clean up
	os.Unsetenv("APP_ENV")
}
