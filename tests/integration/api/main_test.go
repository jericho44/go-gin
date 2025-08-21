package api

import (
	"log"
	"os"
	"testing"
)

// TestMain runs before all tests and handles setup/teardown
func TestMain(m *testing.M) {
	log.Println("Starting integration tests...")

	// Set test environment variables
	setupTestEnvironment()

	// Run tests
	code := m.Run()

	log.Println("Integration tests completed")

	// Exit with the same code as the tests
	os.Exit(code)
}

// setupTestEnvironment sets up environment variables for testing
func setupTestEnvironment() {
	testEnvVars := map[string]string{
		"APP_ENV":     "test",
		"DB_NAME":     "gin_app_test",
		"DB_HOST":     "localhost",
		"DB_PORT":     "5432",
		"DB_USER":     "postgres",
		"DB_PASSWORD": "",
		"JWT_SECRET":  "test-secret-key",
		"SERVER_HOST": "localhost",
		"SERVER_PORT": "0", // Use random port for testing
	}

	for key, value := range testEnvVars {
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}
