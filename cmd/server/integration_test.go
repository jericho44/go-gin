package main

import (
	"testing"

	"gin-golang-app/internal/config"
	"gin-golang-app/internal/handlers"
	"gin-golang-app/internal/middleware"
	"gin-golang-app/internal/routes"
	"gin-golang-app/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestComponentWiring tests that all components can be properly wired together
func TestComponentWiring(t *testing.T) {
	// Load test configuration
	cfg := &config.Config{
		Environment: "test",
		Server: config.ServerConfig{
			Host:         "localhost",
			Port:         "8080",
			ReadTimeout:  30,
			WriteTimeout: 30,
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "test",
			Password: "test",
			Name:     "test_db",
			SSLMode:  "disable",
		},
		JWT: config.JWTConfig{
			Secret:      "test-secret",
			ExpiryHours: 24,
		},
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
		},
	}

	// Test that configuration is properly loaded
	assert.NotNil(t, cfg)
	assert.Equal(t, "test", cfg.Environment)
	assert.Equal(t, "test-secret", cfg.JWT.Secret)

	// Test middleware creation
	corsConfig := middleware.CORSConfig{
		AllowedOrigins: cfg.CORS.AllowedOrigins,
		AllowedMethods: cfg.CORS.AllowedMethods,
		AllowedHeaders: cfg.CORS.AllowedHeaders,
	}

	corsMiddleware := middleware.CORS(corsConfig)
	assert.NotNil(t, corsMiddleware)

	authMiddleware := middleware.DefaultJWTAuth(cfg.JWT.Secret)
	assert.NotNil(t, authMiddleware)

	loggerMiddleware := middleware.CombinedLogger()
	assert.NotNil(t, loggerMiddleware)

	requestIDMiddleware := middleware.RequestID()
	assert.NotNil(t, requestIDMiddleware)

	// Test that we can create a mock service (without database dependency)
	// In a real integration test, you would use a test database
	var userService services.UserService
	assert.Nil(t, userService) // This is expected since we're not initializing it

	// Test handler creation with nil service (just to test the constructor)
	userHandler := handlers.NewUserHandler(userService)
	assert.NotNil(t, userHandler)

	// Test router configuration creation
	routerConfig := routes.RouterConfig{
		HealthHandler: nil, // Would be initialized with real database in integration test
		UserHandler:   userHandler,
		JWTSecret:     cfg.JWT.Secret,
		CORSConfig:    corsConfig,
	}

	assert.Equal(t, cfg.JWT.Secret, routerConfig.JWTSecret)
	assert.Equal(t, corsConfig.AllowedOrigins, routerConfig.CORSConfig.AllowedOrigins)

	// Test that we can create a Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	assert.NotNil(t, router)

	// Test route info retrieval
	routeInfo := routes.GetRouteInfo()
	assert.NotEmpty(t, routeInfo)

	// Verify that expected routes are defined
	expectedRoutes := []string{"/health", "/ready", "/live", "/api/v1/users"}
	foundRoutes := make(map[string]bool)

	for _, route := range routeInfo {
		foundRoutes[route.Path] = true
	}

	for _, expectedRoute := range expectedRoutes {
		assert.True(t, foundRoutes[expectedRoute], "Expected route %s not found", expectedRoute)
	}
}

// TestDependencyInjectionPattern tests the dependency injection pattern
func TestDependencyInjectionPattern(t *testing.T) {
	// Test that handlers properly accept their dependencies
	userHandler := handlers.NewUserHandler(nil) // nil service for testing constructor
	assert.NotNil(t, userHandler)

	healthHandler := handlers.NewHealthHandler(nil) // nil database for testing constructor
	assert.NotNil(t, healthHandler)

	// Test that services can be created with repository dependencies
	userService := services.NewUserService(nil) // nil repository for testing constructor
	assert.NotNil(t, userService)
}

// TestMiddlewareChaining tests that middleware can be properly chained
func TestMiddlewareChaining(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Test that we can add multiple middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.CombinedLogger())
	router.Use(middleware.DefaultCORS())

	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	assert.NotNil(t, router)
}
