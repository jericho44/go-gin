package middleware

import (
	"gin-golang-app/internal/config"

	"github.com/gin-gonic/gin"
)

// NewCORSFromConfig creates a CORS middleware using the application configuration
func NewCORSFromConfig(cfg *config.Config) gin.HandlerFunc {
	return CORS(CORSConfig{
		AllowedOrigins: cfg.CORS.AllowedOrigins,
		AllowedMethods: cfg.CORS.AllowedMethods,
		AllowedHeaders: cfg.CORS.AllowedHeaders,
	})
}

// Example usage:
//
// func main() {
//     cfg := config.Load()
//     router := gin.Default()
//
//     // Use CORS middleware with configuration
//     router.Use(middleware.NewCORSFromConfig(cfg))
//
//     // Or use default CORS
//     // router.Use(middleware.DefaultCORS())
//
//     router.Run(":" + cfg.Server.Port)
// }
