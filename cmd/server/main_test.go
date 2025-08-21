package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gin-golang-app/internal/handlers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckHandler(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create mock handlers
	healthHandler := handlers.NewHealthHandler(nil) // nil database for testing

	// Create a new router and set up basic routes manually for testing
	router := gin.New()

	// Add basic health routes for testing
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/ready", healthHandler.ReadinessCheck)
	router.GET("/live", healthHandler.LivenessCheck)

	// Test the /health endpoint
	t.Run("GET /health", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Since we don't have a real database, the health check will return 503
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		// Just verify that we get a response - the exact structure may vary
		assert.NotEmpty(t, w.Body.String())

		// Try to parse as JSON to ensure it's valid JSON
		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Basic validation that it's a proper error response
		assert.NotNil(t, response)
	})

	// Test the /ready endpoint
	t.Run("GET /ready", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return service unavailable since we don't have a real database
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})

	// Test the /live endpoint
	t.Run("GET /live", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/live", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "Service is alive", response["message"])
	})
}
