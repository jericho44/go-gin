package middleware

import (
	"database/sql"
	"errors"
	"net/http"

	"gin-golang-app/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ExampleErrorHandlerUsage demonstrates how to use the error handling middleware
func ExampleErrorHandlerUsage() {
	r := gin.New()

	// Add the request ID middleware first (for logging context)
	r.Use(RequestID())

	// Add the error handling middleware
	r.Use(ErrorHandler())

	// Add the recovery middleware to handle panics
	r.Use(RecoveryHandler())

	// Example route that demonstrates different error scenarios
	r.GET("/users/:id", func(c *gin.Context) {
		userID := c.Param("id")

		// Example 1: Validation error
		if userID == "" {
			AbortWithValidationError(c, "User ID is required", nil)
			return
		}

		// Example 2: Not found error
		if userID == "404" {
			AbortWithNotFoundError(c, "User not found")
			return
		}

		// Example 3: Business logic error
		if userID == "inactive" {
			AbortWithBusinessError(c, "User account is inactive", http.StatusForbidden)
			return
		}

		// Example 4: Database error (simulated)
		if userID == "db-error" {
			c.Error(sql.ErrNoRows)
			return
		}

		// Example 5: Custom app error
		if userID == "custom" {
			customErr := NewAppError(ErrorTypeInternal, "Custom error occurred", http.StatusInternalServerError, nil)
			c.Error(customErr)
			return
		}

		// Example 6: Panic (will be caught by RecoveryHandler)
		if userID == "panic" {
			panic("Something went wrong!")
		}

		// Success case
		utils.OKResponse(c, "User retrieved successfully", map[string]interface{}{
			"id":   userID,
			"name": "John Doe",
		})
	})

	// Example route that demonstrates validation error with details
	r.POST("/users", func(c *gin.Context) {
		var user struct {
			Name  string `json:"name" binding:"required,min=2"`
			Email string `json:"email" binding:"required,email"`
		}

		// This will trigger validation errors if the request body is invalid
		if err := c.ShouldBindJSON(&user); err != nil {
			// The error handling middleware will automatically handle validation errors
			c.Error(err)
			return
		}

		utils.CreatedResponse(c, "User created successfully", user)
	})

	// Example route that demonstrates database error handling
	r.GET("/users", func(c *gin.Context) {
		// Simulate different database errors
		errorType := c.Query("error")

		switch errorType {
		case "duplicate":
			c.Error(NewDatabaseError("User already exists", sql.ErrConnDone))
		case "connection":
			c.Error(NewDatabaseError("Database connection failed", sql.ErrConnDone))
		case "constraint":
			c.Error(NewDatabaseError("Foreign key constraint violation", sql.ErrTxDone))
		default:
			utils.OKResponse(c, "Users retrieved successfully", []interface{}{})
		}
	})

	r.Run(":8080")
}

// ExampleServiceWithErrorHandling demonstrates how to use error handling in service layer
func ExampleServiceWithErrorHandling() {
	// Example service function that returns different types of errors
	getUserByID := func(id string) (interface{}, error) {
		if id == "" {
			return nil, NewValidationError("User ID is required", nil)
		}

		if id == "404" {
			return nil, NewNotFoundError("User not found")
		}

		if id == "inactive" {
			return nil, NewBusinessError("User account is inactive", http.StatusForbidden)
		}

		// Simulate database error
		if id == "db-error" {
			return nil, NewDatabaseError("Failed to query user", sql.ErrNoRows)
		}

		return map[string]interface{}{
			"id":   id,
			"name": "John Doe",
		}, nil
	}

	// Example handler that uses the service
	handler := func(c *gin.Context) {
		userID := c.Param("id")

		user, err := getUserByID(userID)
		if err != nil {
			// Simply add the error to the context
			// The error handling middleware will take care of the rest
			c.Error(err)
			return
		}

		utils.OKResponse(c, "User retrieved successfully", user)
	}

	r := gin.New()
	r.Use(RequestID())
	r.Use(ErrorHandler())
	r.GET("/users/:id", handler)

	r.Run(":8080")
}

// ExampleCustomErrorTypes demonstrates how to create and use custom error types
func ExampleCustomErrorTypes() {
	// Custom error types for specific business domains
	const (
		ErrorTypePayment   = "payment_error"
		ErrorTypeInventory = "inventory_error"
		ErrorTypeShipping  = "shipping_error"
	)

	// Helper functions for creating domain-specific errors
	NewPaymentError := func(message string, code int) *AppError {
		return &AppError{
			Type:    ErrorTypePayment,
			Message: message,
			Code:    code,
		}
	}

	NewInventoryError := func(message string) *AppError {
		return &AppError{
			Type:    ErrorTypeInventory,
			Message: message,
			Code:    http.StatusConflict,
		}
	}

	// Example usage in handlers
	r := gin.New()
	r.Use(RequestID())
	r.Use(ErrorHandler())

	r.POST("/orders", func(c *gin.Context) {
		// Simulate different business errors
		errorType := c.Query("error")

		switch errorType {
		case "payment":
			c.Error(NewPaymentError("Payment failed", http.StatusPaymentRequired))
		case "inventory":
			c.Error(NewInventoryError("Insufficient inventory"))
		case "shipping":
			c.Error(NewAppError(ErrorTypeShipping, "Shipping not available", http.StatusServiceUnavailable, nil))
		default:
			utils.CreatedResponse(c, "Order created successfully", map[string]interface{}{
				"id": "order-123",
			})
		}
	})

	r.Run(":8080")
}

// ExampleErrorLoggingAndMonitoring demonstrates error logging and monitoring hooks
func ExampleErrorLoggingAndMonitoring() {
	r := gin.New()

	// Add request ID for correlation
	r.Use(RequestID())

	// Custom error handling middleware with monitoring hooks
	r.Use(func(c *gin.Context) {
		c.Next()

		// Check for errors after request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Custom monitoring/alerting logic
			var appErr *AppError
			if errors.As(err, &appErr) {
				// Send metrics to monitoring system
				switch appErr.Type {
				case ErrorTypeDatabase:
					// Alert on database errors
					// metrics.IncrementCounter("database_errors")
				case ErrorTypeInternal:
					// Alert on internal errors
					// metrics.IncrementCounter("internal_errors")
				}
			}
		}
	})

	// Add the standard error handler
	r.Use(ErrorHandler())

	r.GET("/test", func(c *gin.Context) {
		c.Error(NewDatabaseError("Database connection failed", sql.ErrConnDone))
	})

	r.Run(":8080")
}
