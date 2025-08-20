package routes

import (
	"gin-golang-app/internal/handlers"
	"gin-golang-app/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RouterConfig holds configuration for setting up routes
type RouterConfig struct {
	HealthHandler *handlers.HealthHandler
	UserHandler   *handlers.UserHandler
	JWTSecret     string
	CORSConfig    middleware.CORSConfig
}

// SetupRoutes configures all application routes with appropriate middleware
func SetupRoutes(router *gin.Engine, config RouterConfig) {
	// Apply global middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.CombinedLogger())
	router.Use(middleware.CORS(config.CORSConfig))

	// Recovery middleware to handle panics
	router.Use(gin.Recovery())

	// Health check routes (no authentication required)
	setupHealthRoutes(router, config.HealthHandler)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		setupPublicRoutes(v1, config)

		// Protected routes (authentication required)
		setupProtectedRoutes(v1, config)
	}
}

// setupHealthRoutes sets up health check endpoints
func setupHealthRoutes(router *gin.Engine, healthHandler *handlers.HealthHandler) {
	health := router.Group("/")
	{
		health.GET("/health", healthHandler.HealthCheck)
		health.GET("/ready", healthHandler.ReadinessCheck)
		health.GET("/live", healthHandler.LivenessCheck)
	}
}

// setupPublicRoutes sets up public API routes that don't require authentication
func setupPublicRoutes(v1 *gin.RouterGroup, config RouterConfig) {
	// Public user routes (if any)
	// For now, we'll keep user routes protected, but this is where you'd add
	// public endpoints like user registration, login, etc.

	// Example of how to add public routes:
	// public := v1.Group("/public")
	// {
	//     public.POST("/register", config.UserHandler.RegisterUser)
	//     public.POST("/login", config.UserHandler.LoginUser)
	// }
}

// setupProtectedRoutes sets up API routes that require authentication
func setupProtectedRoutes(v1 *gin.RouterGroup, config RouterConfig) {
	// Apply JWT authentication middleware to protected routes
	protected := v1.Group("/")
	protected.Use(middleware.DefaultJWTAuth(config.JWTSecret))
	{
		// User management routes
		setupUserRoutes(protected, config.UserHandler)
	}
}

// setupUserRoutes sets up user-related routes
func setupUserRoutes(group *gin.RouterGroup, userHandler *handlers.UserHandler) {
	users := group.Group("/users")
	{
		// CRUD operations for users
		users.POST("/", userHandler.CreateUser)      // POST /api/v1/users
		users.GET("/", userHandler.ListUsers)        // GET /api/v1/users
		users.GET("/:id", userHandler.GetUser)       // GET /api/v1/users/:id
		users.PUT("/:id", userHandler.UpdateUser)    // PUT /api/v1/users/:id
		users.DELETE("/:id", userHandler.DeleteUser) // DELETE /api/v1/users/:id

		// Additional user routes
		users.GET("/email/:email", userHandler.GetUserByEmail) // GET /api/v1/users/email/:email
	}
}

// SetupRoutesWithCustomMiddleware allows for custom middleware configuration
func SetupRoutesWithCustomMiddleware(router *gin.Engine, config RouterConfig, customMiddleware ...gin.HandlerFunc) {
	// Apply custom middleware first
	for _, middleware := range customMiddleware {
		router.Use(middleware)
	}

	// Then apply standard routes
	SetupRoutes(router, config)
}

// SetupDevelopmentRoutes sets up routes with development-friendly configuration
func SetupDevelopmentRoutes(router *gin.Engine, config RouterConfig) {
	// Development-specific CORS configuration
	devCORSConfig := middleware.CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Requested-With", "X-Request-ID"},
	}

	// Override CORS config for development
	config.CORSConfig = devCORSConfig

	// Set up routes with development configuration
	SetupRoutes(router, config)
}

// SetupProductionRoutes sets up routes with production-friendly configuration
func SetupProductionRoutes(router *gin.Engine, config RouterConfig) {
	// Production-specific CORS configuration
	prodCORSConfig := middleware.CORSConfig{
		AllowedOrigins: []string{
			"https://yourdomain.com",
			"https://www.yourdomain.com",
			"https://api.yourdomain.com",
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Requested-With"},
	}

	// Override CORS config for production
	config.CORSConfig = prodCORSConfig

	// Set up routes with production configuration
	SetupRoutes(router, config)
}

// RouteInfo represents information about a route
type RouteInfo struct {
	Method      string
	Path        string
	HandlerName string
	Protected   bool
}

// GetRouteInfo returns information about all configured routes
func GetRouteInfo() []RouteInfo {
	return []RouteInfo{
		// Health routes
		{Method: "GET", Path: "/health", HandlerName: "HealthCheck", Protected: false},
		{Method: "GET", Path: "/ready", HandlerName: "ReadinessCheck", Protected: false},
		{Method: "GET", Path: "/live", HandlerName: "LivenessCheck", Protected: false},

		// User routes (protected)
		{Method: "POST", Path: "/api/v1/users", HandlerName: "CreateUser", Protected: true},
		{Method: "GET", Path: "/api/v1/users", HandlerName: "ListUsers", Protected: true},
		{Method: "GET", Path: "/api/v1/users/:id", HandlerName: "GetUser", Protected: true},
		{Method: "PUT", Path: "/api/v1/users/:id", HandlerName: "UpdateUser", Protected: true},
		{Method: "DELETE", Path: "/api/v1/users/:id", HandlerName: "DeleteUser", Protected: true},
		{Method: "GET", Path: "/api/v1/users/email/:email", HandlerName: "GetUserByEmail", Protected: true},
	}
}
