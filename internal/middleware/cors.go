package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// CORS creates a gin middleware for handling Cross-Origin Resource Sharing (CORS)
func CORS(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		if isOriginAllowed(origin, config.AllowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		// Set allowed methods
		if len(config.AllowedMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		}

		// Set allowed headers
		if len(config.AllowedHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		}

		// Set additional CORS headers
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isOriginAllowed checks if the given origin is allowed based on the configuration
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return false
	}

	for _, allowedOrigin := range allowedOrigins {
		// Allow all origins if "*" is specified
		if allowedOrigin == "*" {
			return true
		}

		// Exact match
		if allowedOrigin == origin {
			return true
		}

		// Wildcard subdomain matching (e.g., "*.example.com")
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := strings.TrimPrefix(allowedOrigin, "*.")
			// Extract the domain from the origin (remove protocol)
			originDomain := origin
			if strings.Contains(origin, "://") {
				parts := strings.Split(origin, "://")
				if len(parts) > 1 {
					originDomain = parts[1]
				}
			}
			if strings.HasSuffix(originDomain, "."+domain) || originDomain == domain {
				return true
			}
		}
	}

	return false
}

// DefaultCORS creates a CORS middleware with default configuration
func DefaultCORS() gin.HandlerFunc {
	return CORS(CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Requested-With"},
	})
}
