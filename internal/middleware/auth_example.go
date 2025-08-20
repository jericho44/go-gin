package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Example usage of the authentication middleware

// ExampleJWTAuthUsage demonstrates how to use the JWT authentication middleware
func ExampleJWTAuthUsage() {
	router := gin.Default()

	// Secret key for JWT signing (should be from environment variables in production)
	secretKey := "your-secret-key"

	// Method 1: Use default JWT auth middleware
	router.Use(DefaultJWTAuth(secretKey))

	// Method 2: Use custom JWT auth configuration
	customConfig := AuthConfig{
		SecretKey:     secretKey,
		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		SkipPaths:     []string{"/health", "/api/v1/auth/login", "/api/v1/auth/register"},
	}
	router.Use(JWTAuth(customConfig))

	// Method 3: Apply auth to specific route groups
	api := router.Group("/api/v1")
	api.Use(RequireAuth(secretKey))
	{
		api.GET("/users", func(c *gin.Context) {
			// Get user information from context
			userID, exists := GetUserID(c)
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
				return
			}

			username, _ := GetUsername(c)
			email, _ := GetUserEmail(c)

			c.JSON(http.StatusOK, gin.H{
				"user_id":  userID,
				"username": username,
				"email":    email,
			})
		})
	}

	// Method 4: Optional authentication for some routes
	public := router.Group("/api/v1/public")
	public.Use(OptionalAuth(secretKey))
	{
		public.GET("/posts", func(c *gin.Context) {
			// Check if user is authenticated (optional)
			if userID, exists := GetUserID(c); exists {
				// User is authenticated, show personalized content
				c.JSON(http.StatusOK, gin.H{
					"message": "Personalized posts",
					"user_id": userID,
				})
			} else {
				// User is not authenticated, show public content
				c.JSON(http.StatusOK, gin.H{
					"message": "Public posts",
				})
			}
		})
	}

	// Method 5: Conditional authentication
	router.Use(ConditionalAuth(
		func(c *gin.Context) bool {
			// Apply auth only to admin routes
			return c.Request.URL.Path == "/admin"
		},
		RequireAuth(secretKey),
	))

	// Method 6: Chain multiple middleware
	adminRoutes := router.Group("/admin")
	middlewareChain := ChainMiddleware(
		RequestID(),
		Logger(),
		RequireAuth(secretKey),
	)
	for _, mw := range middlewareChain {
		adminRoutes.Use(mw)
	}

	router.Run(":8080")
}

// ExampleGenerateToken demonstrates how to generate a JWT token
func ExampleGenerateToken(secretKey string, userID uint, username, email string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "gin-golang-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// ExampleLoginHandler demonstrates a login handler that generates JWT tokens
func ExampleLoginHandler(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This is a simplified example - in real applications, you would:
		// 1. Validate user credentials against database
		// 2. Hash and compare passwords
		// 3. Handle various error cases

		var loginRequest struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&loginRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request",
				"error":   err.Error(),
			})
			return
		}

		// Simulate user authentication (replace with real authentication logic)
		if loginRequest.Username == "admin" && loginRequest.Password == "password" {
			// Generate JWT token
			token, err := ExampleGenerateToken(secretKey, 1, "admin", "admin@example.com")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Failed to generate token",
					"error":   err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Login successful",
				"token":   token,
				"user": gin.H{
					"id":       1,
					"username": "admin",
					"email":    "admin@example.com",
				},
			})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid credentials",
				"error":   "Username or password is incorrect",
			})
		}
	}
}
