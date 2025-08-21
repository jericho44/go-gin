package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

// AuthIntegrationTestSuite tests authentication and authorization
type AuthIntegrationTestSuite struct {
	IntegrationTestSuite
}

// TestAuthIntegrationSuite runs the authentication integration test suite
func TestAuthIntegrationSuite(t *testing.T) {
	suite.Run(t, new(AuthIntegrationTestSuite))
}

// TestProtectedEndpointsRequireAuthentication tests that protected endpoints require authentication
func (suite *AuthIntegrationTestSuite) TestProtectedEndpointsRequireAuthentication() {
	client := suite.GetHTTPClient()

	protectedEndpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/users"},
		{"POST", "/api/v1/users"},
		{"GET", "/api/v1/users/1"},
		{"PUT", "/api/v1/users/1"},
		{"DELETE", "/api/v1/users/1"},
		{"GET", "/api/v1/users/email/test@example.com"},
	}

	for _, endpoint := range protectedEndpoints {
		suite.Run(fmt.Sprintf("%s %s", endpoint.method, endpoint.path), func() {
			var resp *http.Response
			var err error

			switch endpoint.method {
			case "GET":
				resp, err = client.GET(endpoint.path)
			case "POST":
				resp, err = client.POST(endpoint.path, ValidUserRequests.User1)
			case "PUT":
				resp, err = client.PUT(endpoint.path, ValidUpdateRequests.UpdateUser1)
			case "DELETE":
				resp, err = client.DELETE(endpoint.path)
			}

			suite.Require().NoError(err, "Failed to make request to %s %s", endpoint.method, endpoint.path)

			// Should return 401 Unauthorized without authentication
			suite.Equal(http.StatusUnauthorized, resp.StatusCode,
				"Expected %s %s to require authentication", endpoint.method, endpoint.path)

			resp.Body.Close()
		})
	}
}

// TestInvalidJWTToken tests authentication with invalid JWT tokens
func (suite *AuthIntegrationTestSuite) TestInvalidJWTToken() {
	client := suite.GetHTTPClient()

	invalidTokens := []struct {
		name  string
		token string
	}{
		{"Empty token", ""},
		{"Invalid format", "invalid-token"},
		{"Malformed JWT", "Bearer invalid.jwt.token"},
		{"Expired token", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE1MTYyMzkwMjJ9.invalid"},
		{"Wrong secret", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJleHAiOjk5OTk5OTk5OTksImlhdCI6MTUxNjIzOTAyMn0.wrong_signature"},
	}

	for _, tc := range invalidTokens {
		suite.Run(tc.name, func() {
			headers := map[string]string{}
			if tc.token != "" {
				if tc.token == "Bearer invalid.jwt.token" || tc.token == "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE1MTYyMzkwMjJ9.invalid" || tc.token == "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJleHAiOjk5OTk5OTk5OTksImlhdCI6MTUxNjIzOTAyMn0.wrong_signature" {
					headers["Authorization"] = tc.token
				} else {
					headers["Authorization"] = "Bearer " + tc.token
				}
			}

			resp, err := client.GET("/api/v1/users", headers)
			suite.Require().NoError(err, "Failed to make request with invalid token")

			suite.Equal(http.StatusUnauthorized, resp.StatusCode,
				"Expected unauthorized status for %s", tc.name)

			resp.Body.Close()
		})
	}
}

// TestValidJWTToken tests authentication with valid JWT token
func (suite *AuthIntegrationTestSuite) TestValidJWTToken() {
	client, token, err := suite.GetAuthenticatedHTTPClient(1, "test@example.com")
	suite.Require().NoError(err, "Failed to get authenticated client")

	// Test that valid token allows access to protected endpoint
	resp, err := client.GET("/api/v1/users", AuthHeaders(token))
	suite.Require().NoError(err, "Failed to make authenticated request")

	// Should succeed with valid authentication
	suite.Equal(http.StatusOK, resp.StatusCode, "Expected success with valid token")
	resp.Body.Close()
}

// TestJWTTokenWithoutBearerPrefix tests authentication without Bearer prefix
func (suite *AuthIntegrationTestSuite) TestJWTTokenWithoutBearerPrefix() {
	client := suite.GetHTTPClient()

	token, err := GenerateJWTToken("test-secret-key", 1, "test@example.com")
	suite.Require().NoError(err, "Failed to generate JWT token")

	// Test with token but without Bearer prefix
	headers := map[string]string{
		"Authorization": token, // Missing "Bearer " prefix
	}

	resp, err := client.GET("/api/v1/users", headers)
	suite.Require().NoError(err, "Failed to make request")

	suite.Equal(http.StatusUnauthorized, resp.StatusCode,
		"Expected unauthorized status without Bearer prefix")

	resp.Body.Close()
}

// TestMissingAuthorizationHeader tests requests without Authorization header
func (suite *AuthIntegrationTestSuite) TestMissingAuthorizationHeader() {
	client := suite.GetHTTPClient()

	resp, err := client.GET("/api/v1/users")
	suite.Require().NoError(err, "Failed to make request")

	suite.Equal(http.StatusUnauthorized, resp.StatusCode,
		"Expected unauthorized status without Authorization header")

	resp.Body.Close()
}

// TestCORSHeaders tests that CORS headers are properly set
func (suite *AuthIntegrationTestSuite) TestCORSHeaders() {
	client := suite.GetHTTPClient()

	// Test OPTIONS request (preflight)
	req, err := http.NewRequest("OPTIONS", suite.GetBaseURL()+"/api/v1/users", nil)
	suite.Require().NoError(err, "Failed to create OPTIONS request")

	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")

	resp, err := client.client.Do(req)
	suite.Require().NoError(err, "Failed to make OPTIONS request")

	// Check CORS headers
	suite.NotEmpty(resp.Header.Get("Access-Control-Allow-Origin"), "Expected CORS Allow-Origin header")
	suite.NotEmpty(resp.Header.Get("Access-Control-Allow-Methods"), "Expected CORS Allow-Methods header")
	suite.NotEmpty(resp.Header.Get("Access-Control-Allow-Headers"), "Expected CORS Allow-Headers header")

	resp.Body.Close()
}

// TestRequestIDMiddleware tests that request ID middleware adds request ID header
func (suite *AuthIntegrationTestSuite) TestRequestIDMiddleware() {
	client := suite.GetHTTPClient()

	// Test health endpoint (doesn't require auth)
	resp, err := client.GET("/health")
	suite.Require().NoError(err, "Failed to make request")

	// Check that request ID header is present
	requestID := resp.Header.Get("X-Request-ID")
	suite.NotEmpty(requestID, "Expected X-Request-ID header to be present")
	suite.Len(requestID, 36, "Expected request ID to be UUID format") // UUID length

	resp.Body.Close()
}

// TestLoggingMiddleware tests that logging middleware is working
func (suite *AuthIntegrationTestSuite) TestLoggingMiddleware() {
	client := suite.GetHTTPClient()

	// Make a request to trigger logging
	resp, err := client.GET("/health")
	suite.Require().NoError(err, "Failed to make request")

	// We can't easily test log output in integration tests,
	// but we can verify the request was processed successfully
	suite.Equal(http.StatusOK, resp.StatusCode, "Expected successful response")

	// Verify that request ID is present (logged by logging middleware)
	requestID := resp.Header.Get("X-Request-ID")
	suite.NotEmpty(requestID, "Expected request ID to be present")

	resp.Body.Close()
}
