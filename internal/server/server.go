package server

import (
	"gin-golang-app/internal/config"
	"gin-golang-app/internal/database"
	"gin-golang-app/internal/handlers"
	"gin-golang-app/internal/middleware"
	"gin-golang-app/internal/repository"
	"gin-golang-app/internal/routes"
	"gin-golang-app/internal/services"

	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server with all dependencies
type Server struct {
	Router *gin.Engine
	Config *config.Config
	DB     *database.Database
}

// NewServer creates a new server instance with all dependencies wired up
func NewServer(cfg *config.Config, db *database.Database) (*Server, error) {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	userService := services.NewUserService(userRepo)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(db)
	userHandler := handlers.NewUserHandler(userService)

	// Create Gin router
	router := gin.New()

	// Setup routes with dependency injection
	setupRoutes(router, cfg, healthHandler, userHandler)

	return &Server{
		Router: router,
		Config: cfg,
		DB:     db,
	}, nil
}

// setupRoutes configures all the routes for the application
func setupRoutes(router *gin.Engine, cfg *config.Config, healthHandler *handlers.HealthHandler, userHandler *handlers.UserHandler) {
	// Create CORS configuration from app config
	corsConfig := middleware.CORSConfig{
		AllowedOrigins: cfg.CORS.AllowedOrigins,
		AllowedMethods: cfg.CORS.AllowedMethods,
		AllowedHeaders: cfg.CORS.AllowedHeaders,
	}

	// Create router configuration
	routerConfig := routes.RouterConfig{
		HealthHandler: healthHandler,
		UserHandler:   userHandler,
		JWTSecret:     cfg.JWT.Secret,
		CORSConfig:    corsConfig,
	}

	// Setup routes based on environment
	switch cfg.Environment {
	case "production":
		routes.SetupProductionRoutes(router, routerConfig)
	case "staging":
		routes.SetupRoutes(router, routerConfig)
	default:
		routes.SetupDevelopmentRoutes(router, routerConfig)
	}
}
