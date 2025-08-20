package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gin-golang-app/internal/config"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else if cfg.IsStaging() {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin router
	router := gin.New()

	// Add basic middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Setup routes
	setupRoutes(router)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s:%s in %s mode", cfg.Server.Host, cfg.Server.Port, cfg.Environment)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return
	}

	log.Println("Server exited")
}

// setupRoutes configures all the routes for the application
func setupRoutes(router *gin.Engine) {
	// Health check endpoint
	router.GET("/health", healthCheckHandler)

	// API version group
	v1 := router.Group("/api/v1")
	{
		// Health check endpoint under API group as well
		v1.GET("/health", healthCheckHandler)
	}
}

// healthCheckHandler handles health check requests
func healthCheckHandler(c *gin.Context) {
	response := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "gin-golang-app",
		"version":   "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}
