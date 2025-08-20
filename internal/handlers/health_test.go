package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gin-golang-app/internal/database"
	"gin-golang-app/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_HealthCheck_NilDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handler with nil database
	handler := NewHealthHandler(nil)

	// Create test router
	router := gin.New()
	router.GET("/health", handler.HealthCheck)

	// Create request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert status code
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	// Parse response
	var response utils.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Assert response
	assert.False(t, response.Success)
	assert.Equal(t, "System is unhealthy", response.Message)
}

func TestHealthHandler_HealthCheck_WithDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a database config for testing
	config := &database.DatabaseConfig{
		Driver: "postgres",
		Host:   "localhost",
		Port:   5432,
	}

	// Create a database instance (without actual connection for testing)
	db := &database.Database{
		Config: config,
		DB:     nil, // This will cause health check to fail, which is expected for testing
	}

	// Create handler
	handler := NewHealthHandler(db)

	// Create test router
	router := gin.New()
	router.GET("/health", handler.HealthCheck)

	// Create request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert status code (should be unhealthy due to no actual DB connection)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	// Parse response
	var response utils.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Assert response structure
	assert.False(t, response.Success)
	assert.Equal(t, "System is unhealthy", response.Message)
}

func TestHealthHandler_ReadinessCheck_NilDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handler with nil database
	handler := NewHealthHandler(nil)

	// Create test router
	router := gin.New()
	router.GET("/ready", handler.ReadinessCheck)

	// Create request
	req, _ := http.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert status code
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	// Parse response
	var response utils.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Assert response
	assert.False(t, response.Success)
	assert.Equal(t, "Service is not ready", response.Message)
}

func TestHealthHandler_LivenessCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handler (database can be nil for liveness check)
	handler := NewHealthHandler(nil)

	// Create test router
	router := gin.New()
	router.GET("/live", handler.LivenessCheck)

	// Create request
	req, _ := http.NewRequest("GET", "/live", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response utils.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Assert response
	assert.True(t, response.Success)
	assert.Equal(t, "Service is alive", response.Message)

	// Assert response data
	assert.NotNil(t, response.Data)
	data, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "alive", data["status"])
	assert.NotNil(t, data["timestamp"])
	assert.NotNil(t, data["uptime"])
}

func TestNewHealthHandler(t *testing.T) {
	config := &database.DatabaseConfig{Driver: "postgres"}
	mockDB := &database.Database{Config: config}
	handler := NewHealthHandler(mockDB)

	assert.NotNil(t, handler)
	assert.Equal(t, mockDB, handler.db)
}

func TestHealthHandler_checkDatabaseHealth_NilDatabase(t *testing.T) {
	handler := NewHealthHandler(nil)
	result := handler.checkDatabaseHealth()

	assert.Equal(t, "unhealthy", result.Status)
	assert.Equal(t, "Database connection not initialized", result.Error)
}

func TestHealthHandler_checkDatabaseHealth_WithDatabase(t *testing.T) {
	config := &database.DatabaseConfig{
		Driver: "postgres",
	}

	// Create database instance without actual connection
	db := &database.Database{
		Config: config,
		DB:     nil, // This will cause health check to fail
	}

	handler := NewHealthHandler(db)
	result := handler.checkDatabaseHealth()

	assert.Equal(t, "unhealthy", result.Status)
	assert.Equal(t, config.Driver, result.Driver)
	assert.NotEmpty(t, result.Error)
	assert.GreaterOrEqual(t, result.ResponseTimeMs, int64(0))
}
