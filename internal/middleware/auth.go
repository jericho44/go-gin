package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	SecretKey     string
	TokenLookup   string // "header:Authorization" or "query:token" or "cookie:jwt"
	TokenHeadName string // "Bearer"
	SkipPaths     []string
}

// Claims represents the JWT claims structure
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// JWTAuth creates a JWT authentication middleware
func JWTAuth(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for specified paths
		if shouldSkipAuth(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		// Extract token from request
		token, err := extractToken(c, config)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
				"path":  c.Request.URL.Path,
			}).Warn("Token extraction failed")

			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authentication required",
				"error":   "Missing or invalid token",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := validateToken(token, config.SecretKey)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
				"path":  c.Request.URL.Path,
			}).Warn("Token validation failed")

			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authentication failed",
				"error":   "Invalid token",
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("claims", claims)

		logrus.WithFields(logrus.Fields{
			"user_id":  claims.UserID,
			"username": claims.Username,
			"path":     c.Request.URL.Path,
		}).Info("User authenticated successfully")

		c.Next()
	}
}

// extractToken extracts JWT token from the request based on configuration
func extractToken(c *gin.Context, config AuthConfig) (string, error) {
	var token string
	var err error

	// Parse token lookup configuration
	parts := strings.Split(config.TokenLookup, ":")
	if len(parts) != 2 {
		return "", jwt.ErrTokenMalformed
	}

	switch parts[0] {
	case "header":
		token, err = extractTokenFromHeader(c, parts[1], config.TokenHeadName)
	case "query":
		token, err = extractTokenFromQuery(c, parts[1])
	case "cookie":
		token, err = extractTokenFromCookie(c, parts[1])
	default:
		return "", jwt.ErrTokenMalformed
	}

	return token, err
}

// extractTokenFromHeader extracts token from HTTP header
func extractTokenFromHeader(c *gin.Context, headerName, tokenHeadName string) (string, error) {
	authHeader := c.GetHeader(headerName)
	if authHeader == "" {
		return "", jwt.ErrTokenMalformed
	}

	// Check if token has the expected prefix (e.g., "Bearer ")
	if tokenHeadName != "" {
		prefix := tokenHeadName + " "
		if !strings.HasPrefix(authHeader, prefix) {
			return "", jwt.ErrTokenMalformed
		}
		return strings.TrimPrefix(authHeader, prefix), nil
	}

	return authHeader, nil
}

// extractTokenFromQuery extracts token from query parameter
func extractTokenFromQuery(c *gin.Context, paramName string) (string, error) {
	token := c.Query(paramName)
	if token == "" {
		return "", jwt.ErrTokenMalformed
	}
	return token, nil
}

// extractTokenFromCookie extracts token from cookie
func extractTokenFromCookie(c *gin.Context, cookieName string) (string, error) {
	token, err := c.Cookie(cookieName)
	if err != nil {
		return "", jwt.ErrTokenMalformed
	}
	return token, nil
}

// validateToken validates the JWT token and returns claims
func validateToken(tokenString, secretKey string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenMalformed
}

// shouldSkipAuth checks if the path should skip authentication
func shouldSkipAuth(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// DefaultJWTAuth creates a JWT middleware with default configuration
func DefaultJWTAuth(secretKey string) gin.HandlerFunc {
	return JWTAuth(AuthConfig{
		SecretKey:     secretKey,
		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		SkipPaths:     []string{"/health", "/api/v1/auth/login", "/api/v1/auth/register"},
	})
}

// RequireAuth is a middleware that requires authentication for specific routes
func RequireAuth(secretKey string) gin.HandlerFunc {
	return JWTAuth(AuthConfig{
		SecretKey:     secretKey,
		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		SkipPaths:     []string{}, // No paths are skipped
	})
}

// OptionalAuth is a middleware that allows optional authentication
func OptionalAuth(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from request
		config := AuthConfig{
			SecretKey:     secretKey,
			TokenLookup:   "header:Authorization",
			TokenHeadName: "Bearer",
		}

		token, err := extractToken(c, config)
		if err != nil {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Validate token if provided
		claims, err := validateToken(token, config.SecretKey)
		if err != nil {
			// Invalid token, continue without authentication
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
				"path":  c.Request.URL.Path,
			}).Warn("Optional auth: Invalid token provided")
			c.Next()
			return
		}

		// Set user information in context if token is valid
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("claims", claims)

		c.Next()
	}
}

// ChainMiddleware chains multiple middleware functions together
// Note: In Gin, middleware chaining is typically done by calling router.Use() multiple times
// This function is provided for cases where you need to combine middleware programmatically
func ChainMiddleware(middlewares ...gin.HandlerFunc) []gin.HandlerFunc {
	return middlewares
}

// ConditionalAuth applies authentication middleware based on a condition
func ConditionalAuth(condition func(*gin.Context) bool, authMiddleware gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if condition(c) {
			authMiddleware(c)
		} else {
			c.Next()
		}
	}
}

// GetUserID retrieves the user ID from the context
func GetUserID(c *gin.Context) (uint, bool) {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return id, true
		}
	}
	return 0, false
}

// GetUsername retrieves the username from the context
func GetUsername(c *gin.Context) (string, bool) {
	if username, exists := c.Get("username"); exists {
		if name, ok := username.(string); ok {
			return name, true
		}
	}
	return "", false
}

// GetUserEmail retrieves the user email from the context
func GetUserEmail(c *gin.Context) (string, bool) {
	if email, exists := c.Get("email"); exists {
		if userEmail, ok := email.(string); ok {
			return userEmail, true
		}
	}
	return "", false
}

// GetClaims retrieves the full claims from the context
func GetClaims(c *gin.Context) (*Claims, bool) {
	if claims, exists := c.Get("claims"); exists {
		if userClaims, ok := claims.(*Claims); ok {
			return userClaims, true
		}
	}
	return nil, false
}
