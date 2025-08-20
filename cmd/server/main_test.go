package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckHandler(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	router := gin.New()
	setupRoutes(router)

	// Test the /health endpoint
	t.Run("GET /health", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "healthy", response["status"])
		assert.Equal(t, "gin-golang-app", response["service"])
		assert.Equal(t, "1.0.0", response["version"])
		assert.NotEmpty(t, response["timestamp"])
	})

	// Test the /api/v1/health endpoint
	t.Run("GET /api/v1/health", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "healthy", response["status"])
		assert.Equal(t, "gin-golang-app", response["service"])
		assert.Equal(t, "1.0.0", response["version"])
		assert.NotEmpty(t, response["timestamp"])
	})
}
