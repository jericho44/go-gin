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
	"gin-golang-app/internal/database"
	"gin-golang-app/internal/server"

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

	// Initialize database connection
	db, err := initializeDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Create server instance
	srv, err := server.NewServer(cfg, db)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      srv.Router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s:%s in %s mode", cfg.Server.Host, cfg.Server.Port, cfg.Environment)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return
	}

	log.Println("Server exited")
}

// initializeDatabase creates and configures the database connection
func initializeDatabase(cfg *config.Config) (*database.Database, error) {
	// Convert config to database config
	dbConfig := &database.DatabaseConfig{
		Driver:          "postgres", // Default to postgres, could be made configurable
		Host:            cfg.Database.Host,
		Port:            parsePort(cfg.Database.Port),
		Username:        cfg.Database.User,
		Password:        cfg.Database.Password,
		DatabaseName:    cfg.Database.Name,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	// Create database connection
	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	return db, nil
}

// parsePort converts string port to int with default fallback
func parsePort(portStr string) int {
	if portStr == "" {
		return 5432 // Default PostgreSQL port
	}

	// Simple conversion - in production you might want more robust parsing
	switch portStr {
	case "5432":
		return 5432
	case "3306":
		return 3306
	default:
		return 5432
	}
}
