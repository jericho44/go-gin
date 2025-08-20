package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gin-golang-app/internal/handlers"
	"gin-golang-app/internal/middleware"
	"gin-golang-app/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of UserService for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) ListUsers(ctx context.Context, page, limit int) ([]*models.User, int64, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UserExists(ctx context.Context, id uint) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func TestSetupRoutes(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create mock services
	mockUserService := &MockUserService{}

	// Create handlers with mock services
	userHandler := handlers.NewUserHandler(mockUserService)
	healthHandler := handlers.NewHealthHandler(nil) // Health handler doesn't need database for basic tests

	// Create router configuration
	config := RouterConfig{
		HealthHandler: healthHandler,
		UserHandler:   userHandler,
		JWTSecret:     "test-secret",
		CORSConfig: middleware.CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
		},
	}

	// Create router and setup routes
	router := gin.New()
	SetupRoutes(router, config)

	// Test health routes
	t.Run("Health Routes", func(t *testing.T) {
		// Test health check endpoint
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Health check should return 200 or 503 (depending on database connection)
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusServiceUnavailable)

		// Test readiness check endpoint
		req, _ = http.NewRequest("GET", "/ready", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Readiness check should return 200 or 503
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusServiceUnavailable)

		// Test liveness check endpoint
		req, _ = http.NewRequest("GET", "/live", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Liveness check should always return 200
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test protected routes without authentication
	t.Run("Protected Routes Without Auth", func(t *testing.T) {
		// Test user routes without authentication - should return 401
		req, _ := http.NewRequest("GET", "/api/v1/users/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test CORS headers
	t.Run("CORS Headers", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "/api/v1/users", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should handle preflight request
		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestSetupDevelopmentRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock services
	mockUserService := &MockUserService{}
	userHandler := handlers.NewUserHandler(mockUserService)
	healthHandler := handlers.NewHealthHandler(nil)

	config := RouterConfig{
		HealthHandler: healthHandler,
		UserHandler:   userHandler,
		JWTSecret:     "test-secret",
		CORSConfig:    middleware.CORSConfig{}, // Will be overridden by development setup
	}

	router := gin.New()
	SetupDevelopmentRoutes(router, config)

	// Test that development CORS allows all origins
	req, _ := http.NewRequest("OPTIONS", "/api/v1/users", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "http://any-origin.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestSetupProductionRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock services
	mockUserService := &MockUserService{}
	userHandler := handlers.NewUserHandler(mockUserService)
	healthHandler := handlers.NewHealthHandler(nil)

	config := RouterConfig{
		HealthHandler: healthHandler,
		UserHandler:   userHandler,
		JWTSecret:     "test-secret",
		CORSConfig:    middleware.CORSConfig{}, // Will be overridden by production setup
	}

	router := gin.New()
	SetupProductionRoutes(router, config)

	// Test that production CORS only allows specific origins
	req, _ := http.NewRequest("OPTIONS", "/api/v1/users", nil)
	req.Header.Set("Origin", "https://yourdomain.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "https://yourdomain.com", w.Header().Get("Access-Control-Allow-Origin"))

	// Test that unauthorized origin is not allowed
	req, _ = http.NewRequest("OPTIONS", "/api/v1/users", nil)
	req.Header.Set("Origin", "https://malicious-site.com")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestGetRouteInfo(t *testing.T) {
	routeInfo := GetRouteInfo()

	// Verify that we have the expected number of routes
	assert.Len(t, routeInfo, 9) // 3 health routes + 6 user routes

	// Verify health routes are not protected
	healthRoutes := 0
	for _, route := range routeInfo {
		if route.Path == "/health" || route.Path == "/ready" || route.Path == "/live" {
			assert.False(t, route.Protected, "Health routes should not be protected")
			healthRoutes++
		}
	}
	assert.Equal(t, 3, healthRoutes, "Should have 3 health routes")

	// Verify user routes are protected
	userRoutes := 0
	for _, route := range routeInfo {
		if route.Path == "/api/v1/users" ||
			route.Path == "/api/v1/users/:id" ||
			route.Path == "/api/v1/users/email/:email" {
			assert.True(t, route.Protected, "User routes should be protected")
			userRoutes++
		}
	}
	assert.Equal(t, 6, userRoutes, "Should have 6 user routes")
}

func TestSetupRoutesWithCustomMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a custom middleware for testing
	customMiddlewareCalled := false
	customMiddleware := func(c *gin.Context) {
		customMiddlewareCalled = true
		c.Next()
	}

	// Create mock services
	mockUserService := &MockUserService{}
	userHandler := handlers.NewUserHandler(mockUserService)
	healthHandler := handlers.NewHealthHandler(nil)

	config := RouterConfig{
		HealthHandler: healthHandler,
		UserHandler:   userHandler,
		JWTSecret:     "test-secret",
		CORSConfig: middleware.CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
		},
	}

	router := gin.New()
	SetupRoutesWithCustomMiddleware(router, config, customMiddleware)

	// Test that custom middleware is called
	req, _ := http.NewRequest("GET", "/live", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, customMiddlewareCalled, "Custom middleware should be called")
	assert.Equal(t, http.StatusOK, w.Code)
}
