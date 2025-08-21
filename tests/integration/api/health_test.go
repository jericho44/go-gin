package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

// HealthIntegrationTestSuite tests health check endpoints
type HealthIntegrationTestSuite struct {
	IntegrationTestSuite
}

// TestHealthIntegrationSuite runs the health integration test suite
func TestHealthIntegrationSuite(t *testing.T) {
	suite.Run(t, new(HealthIntegrationTestSuite))
}

// TestHealthCheck tests the main health check endpoint
func (suite *HealthIntegrationTestSuite) TestHealthCheck() {
	client := suite.GetHTTPClient()

	// Test health check endpoint
	resp, err := client.GET("/health")
	suite.Require().NoError(err, "Failed to make health check request")

	// Assert response
	response := suite.AssertSuccessResponse(resp, http.StatusOK, "System is healthy")

	// Verify response data structure
	data, ok := response.Data.(map[string]interface{})
	suite.Require().True(ok, "Expected data to be a map")

	suite.Contains(data, "status", "Expected status field")
	suite.Contains(data, "timestamp", "Expected timestamp field")
	suite.Contains(data, "database", "Expected database field")

	suite.Equal("healthy", data["status"], "Expected status to be healthy")

	// Verify database status
	dbStatus, ok := data["database"].(map[string]interface{})
	suite.Require().True(ok, "Expected database to be a map")
	suite.Contains(dbStatus, "status", "Expected database status field")
}

// TestReadinessCheck tests the readiness check endpoint
func (suite *HealthIntegrationTestSuite) TestReadinessCheck() {
	client := suite.GetHTTPClient()

	// Test readiness check endpoint
	resp, err := client.GET("/ready")
	suite.Require().NoError(err, "Failed to make readiness check request")

	// Assert response
	response := suite.AssertSuccessResponse(resp, http.StatusOK, "Service is ready")

	// Verify response data structure
	data, ok := response.Data.(map[string]interface{})
	suite.Require().True(ok, "Expected data to be a map")

	suite.Contains(data, "status", "Expected status field")
	suite.Contains(data, "timestamp", "Expected timestamp field")
	suite.Contains(data, "services", "Expected services field")

	suite.Equal("ready", data["status"], "Expected status to be ready")
}

// TestLivenessCheck tests the liveness check endpoint
func (suite *HealthIntegrationTestSuite) TestLivenessCheck() {
	client := suite.GetHTTPClient()

	// Test liveness check endpoint
	resp, err := client.GET("/live")
	suite.Require().NoError(err, "Failed to make liveness check request")

	// Assert response
	response := suite.AssertSuccessResponse(resp, http.StatusOK, "Service is alive")

	// Verify response data structure
	data, ok := response.Data.(map[string]interface{})
	suite.Require().True(ok, "Expected data to be a map")

	suite.Contains(data, "status", "Expected status field")
	suite.Contains(data, "timestamp", "Expected timestamp field")

	suite.Equal("alive", data["status"], "Expected status to be alive")
}

// TestHealthEndpointsWithoutAuthentication tests that health endpoints don't require authentication
func (suite *HealthIntegrationTestSuite) TestHealthEndpointsWithoutAuthentication() {
	client := suite.GetHTTPClient()

	endpoints := []string{"/health", "/ready", "/live"}

	for _, endpoint := range endpoints {
		// Test without any authentication headers
		resp, err := client.GET(endpoint)
		suite.Require().NoError(err, "Failed to make request to %s", endpoint)

		// Should succeed without authentication
		suite.Equal(http.StatusOK, resp.StatusCode, "Expected %s to work without authentication", endpoint)
		resp.Body.Close()
	}
}

// TestHealthEndpointsResponseHeaders tests that health endpoints return correct headers
func (suite *HealthIntegrationTestSuite) TestHealthEndpointsResponseHeaders() {
	client := suite.GetHTTPClient()

	endpoints := []string{"/health", "/ready", "/live"}

	for _, endpoint := range endpoints {
		resp, err := client.GET(endpoint)
		suite.Require().NoError(err, "Failed to make request to %s", endpoint)

		// Check content type
		suite.Equal("application/json; charset=utf-8", resp.Header.Get("Content-Type"),
			"Expected correct content type for %s", endpoint)

		// Check that response has proper headers
		suite.NotEmpty(resp.Header.Get("Date"), "Expected Date header for %s", endpoint)

		resp.Body.Close()
	}
}
