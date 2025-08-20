package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestJWTAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secretKey := "test-secret-key"
	config := AuthConfig{
		SecretKey:     secretKey,
		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		SkipPaths:     []string{"/health"},
	}

	tests := []struct {
		name           string
		path           string
		token          string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid token",
			path:           "/protected",
			token:          generateTestToken(secretKey, 1, "testuser", "test@example.com"),
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "Missing token",
			path:           "/protected",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Authentication required",
		},
		{
			name:           "Invalid token",
			path:           "/protected",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Authentication failed",
		},
		{
			name:           "Skip path",
			path:           "/health",
			token:          "",
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(JWTAuth(config))
			router.GET("/*path", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

func TestExtractTokenFromHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		headerValue   string
		tokenHeadName string
		expectedToken string
		expectError   bool
	}{
		{
			name:          "Valid Bearer token",
			headerValue:   "Bearer valid-token",
			tokenHeadName: "Bearer",
			expectedToken: "valid-token",
			expectError:   false,
		},
		{
			name:          "Missing Bearer prefix",
			headerValue:   "valid-token",
			tokenHeadName: "Bearer",
			expectedToken: "",
			expectError:   true,
		},
		{
			name:          "Empty header",
			headerValue:   "",
			tokenHeadName: "Bearer",
			expectedToken: "",
			expectError:   true,
		},
		{
			name:          "No token head name",
			headerValue:   "valid-token",
			tokenHeadName: "",
			expectedToken: "valid-token",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Request.Header.Set("Authorization", tt.headerValue)

			token, err := extractTokenFromHeader(c, "Authorization", tt.tokenHeadName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}

func TestExtractTokenFromQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		queryParam    string
		expectedToken string
		expectError   bool
	}{
		{
			name:          "Valid query token",
			queryParam:    "valid-token",
			expectedToken: "valid-token",
			expectError:   false,
		},
		{
			name:          "Empty query token",
			queryParam:    "",
			expectedToken: "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			url := "/"
			if tt.queryParam != "" {
				url += "?token=" + tt.queryParam
			}
			c.Request = httptest.NewRequest("GET", url, nil)

			token, err := extractTokenFromQuery(c, "token")

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	secretKey := "test-secret-key"

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "Valid token",
			token:       generateTestToken(secretKey, 1, "testuser", "test@example.com"),
			expectError: false,
		},
		{
			name:        "Invalid token",
			token:       "invalid-token",
			expectError: true,
		},
		{
			name:        "Expired token",
			token:       generateExpiredTestToken(secretKey, 1, "testuser", "test@example.com"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := validateToken(tt.token, secretKey)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, uint(1), claims.UserID)
				assert.Equal(t, "testuser", claims.Username)
				assert.Equal(t, "test@example.com", claims.Email)
			}
		})
	}
}

func TestShouldSkipAuth(t *testing.T) {
	skipPaths := []string{"/health", "/api/v1/auth"}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Skip health path",
			path:     "/health",
			expected: true,
		},
		{
			name:     "Skip auth path",
			path:     "/api/v1/auth/login",
			expected: true,
		},
		{
			name:     "Don't skip protected path",
			path:     "/api/v1/users",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkipAuth(tt.path, skipPaths)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestChainMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var executionOrder []string

	middleware1 := func(c *gin.Context) {
		executionOrder = append(executionOrder, "middleware1")
		c.Next()
	}

	middleware2 := func(c *gin.Context) {
		executionOrder = append(executionOrder, "middleware2")
		c.Next()
	}

	middleware3 := func(c *gin.Context) {
		executionOrder = append(executionOrder, "middleware3")
		c.Next()
	}

	router := gin.New()
	// Use the chained middleware
	chainedMiddleware := ChainMiddleware(middleware1, middleware2, middleware3)
	for _, mw := range chainedMiddleware {
		router.Use(mw)
	}
	router.GET("/test", func(c *gin.Context) {
		executionOrder = append(executionOrder, "handler")
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []string{"middleware1", "middleware2", "middleware3", "handler"}, executionOrder)
}

func TestGetUserHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Test when no user data is set
	userID, exists := GetUserID(c)
	assert.False(t, exists)
	assert.Equal(t, uint(0), userID)

	username, exists := GetUsername(c)
	assert.False(t, exists)
	assert.Equal(t, "", username)

	email, exists := GetUserEmail(c)
	assert.False(t, exists)
	assert.Equal(t, "", email)

	claims, exists := GetClaims(c)
	assert.False(t, exists)
	assert.Nil(t, claims)

	// Set user data in context
	c.Set("user_id", uint(123))
	c.Set("username", "testuser")
	c.Set("email", "test@example.com")
	testClaims := &Claims{
		UserID:   123,
		Username: "testuser",
		Email:    "test@example.com",
	}
	c.Set("claims", testClaims)

	// Test when user data is set
	userID, exists = GetUserID(c)
	assert.True(t, exists)
	assert.Equal(t, uint(123), userID)

	username, exists = GetUsername(c)
	assert.True(t, exists)
	assert.Equal(t, "testuser", username)

	email, exists = GetUserEmail(c)
	assert.True(t, exists)
	assert.Equal(t, "test@example.com", email)

	claims, exists = GetClaims(c)
	assert.True(t, exists)
	assert.Equal(t, testClaims, claims)
}

// Helper function to generate a test JWT token
func generateTestToken(secretKey string, userID uint, username, email string) string {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secretKey))
	return tokenString
}

// Helper function to generate an expired test JWT token
func generateExpiredTestToken(secretKey string, userID uint, username, email string) string {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secretKey))
	return tokenString
}
