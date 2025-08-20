package middleware

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"

	"gin-golang-app/pkg/utils"
)

// AppError represents a custom application error with additional context
type AppError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Details any    `json:"details,omitempty"`
	Err     error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// Error type constants
const (
	ErrorTypeValidation     = "validation_error"
	ErrorTypeDatabase       = "database_error"
	ErrorTypeAuthentication = "authentication_error"
	ErrorTypeBusiness       = "business_error"
	ErrorTypeInternal       = "internal_error"
	ErrorTypeNotFound       = "not_found_error"
	ErrorTypeConflict       = "conflict_error"
	ErrorTypeUnauthorized   = "unauthorized_error"
	ErrorTypeForbidden      = "forbidden_error"
	ErrorTypeRateLimit      = "rate_limit_error"
)

// NewAppError creates a new application error
func NewAppError(errorType, message string, code int, err error) *AppError {
	return &AppError{
		Type:    errorType,
		Message: message,
		Code:    code,
		Err:     err,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string, details any) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: message,
		Code:    http.StatusBadRequest,
		Details: details,
	}
}

// NewDatabaseError creates a database error
func NewDatabaseError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeDatabase,
		Message: message,
		Code:    http.StatusInternalServerError,
		Err:     err,
	}
}

// NewBusinessError creates a business logic error
func NewBusinessError(message string, code int) *AppError {
	return &AppError{
		Type:    ErrorTypeBusiness,
		Message: message,
		Code:    code,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: message,
		Code:    http.StatusNotFound,
	}
}

// ErrorHandler is the centralized error handling middleware
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) == 0 {
			return
		}

		// Get the last error (most recent)
		err := c.Errors.Last()

		// Get request ID for logging
		requestID, _ := c.Get("request_id")

		// Log the error with context
		logError(err.Err, requestID, c)

		// Handle the error based on its type
		handleError(c, err.Err)
	}
}

// logError logs the error with appropriate context and level
func logError(err error, requestID any, c *gin.Context) {
	fields := logrus.Fields{
		"error":      err.Error(),
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"client_ip":  c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
	}

	if requestID != nil {
		fields["request_id"] = requestID
	}

	// Add user context if available
	if userID, exists := c.Get("user_id"); exists {
		fields["user_id"] = userID
	}

	// Determine log level based on error type
	var appErr *AppError
	if errors.As(err, &appErr) {
		fields["error_type"] = appErr.Type
		fields["error_code"] = appErr.Code

		switch appErr.Type {
		case ErrorTypeValidation, ErrorTypeNotFound, ErrorTypeUnauthorized, ErrorTypeForbidden:
			logrus.WithFields(fields).Warn("Client error occurred")
		case ErrorTypeDatabase, ErrorTypeInternal:
			logrus.WithFields(fields).Error("Server error occurred")
		default:
			logrus.WithFields(fields).Info("Application error occurred")
		}
	} else {
		// Unknown error type, log as error
		logrus.WithFields(fields).Error("Unhandled error occurred")
	}
}

// handleError processes the error and sends appropriate response
func handleError(c *gin.Context, err error) {
	// Check if response was already written
	if c.Writer.Written() {
		return
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		// Handle custom application errors
		handleAppError(c, appErr)
		return
	}

	// Handle specific error types
	switch {
	case errors.Is(err, sql.ErrNoRows):
		utils.NotFoundResponse(c, "Resource not found")
	case isValidationError(err):
		handleValidationError(c, err)
	case isDatabaseError(err):
		handleDatabaseError(c, err)
	default:
		// Unknown error - return generic internal server error
		utils.InternalServerErrorResponse(c, "An unexpected error occurred", "Internal server error")
	}
}

// handleAppError handles custom application errors
func handleAppError(c *gin.Context, appErr *AppError) {
	switch appErr.Type {
	case ErrorTypeValidation:
		if appErr.Details != nil {
			utils.ErrorResponseWithDetails(c, appErr.Code, appErr.Message, appErr.Error(), appErr.Details)
		} else {
			utils.BadRequestResponse(c, appErr.Message, appErr.Error())
		}
	case ErrorTypeNotFound:
		utils.NotFoundResponse(c, appErr.Message)
	case ErrorTypeUnauthorized:
		utils.UnauthorizedResponse(c, appErr.Message)
	case ErrorTypeForbidden:
		utils.ForbiddenResponse(c, appErr.Message)
	case ErrorTypeConflict:
		utils.ConflictResponse(c, appErr.Message, appErr.Error())
	case ErrorTypeRateLimit:
		utils.TooManyRequestsResponse(c, appErr.Message)
	case ErrorTypeDatabase, ErrorTypeInternal:
		utils.InternalServerErrorResponse(c, appErr.Message, "Internal server error")
	case ErrorTypeBusiness:
		utils.ErrorResponseWithCode(c, appErr.Code, appErr.Message, appErr.Error())
	default:
		utils.InternalServerErrorResponse(c, appErr.Message, appErr.Error())
	}
}

// handleValidationError handles validation errors from gin binding
func handleValidationError(c *gin.Context, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		var responseErrors []utils.ResponseValidationError
		for _, fieldErr := range validationErrors {
			responseErrors = append(responseErrors, utils.ResponseValidationError{
				Field:   strings.ToLower(fieldErr.Field()),
				Message: getValidationErrorMessage(fieldErr),
				Value:   fieldErr.Value(),
			})
		}
		utils.ValidationErrorResponse(c, responseErrors)
		return
	}

	// Generic validation error
	utils.BadRequestResponse(c, "Validation failed", err.Error())
}

// handleDatabaseError handles database-related errors
func handleDatabaseError(c *gin.Context, err error) {
	// Check for specific database errors
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique"):
		utils.ConflictResponse(c, "Resource already exists", "Duplicate entry")
	case strings.Contains(errStr, "foreign key"):
		utils.BadRequestResponse(c, "Invalid reference", "Referenced resource does not exist")
	case strings.Contains(errStr, "connection"):
		utils.ServiceUnavailableResponse(c, "Database temporarily unavailable")
	default:
		utils.InternalServerErrorResponse(c, "Database operation failed", "Internal server error")
	}
}

// isValidationError checks if the error is a validation error
func isValidationError(err error) bool {
	var validationErrors validator.ValidationErrors
	return errors.As(err, &validationErrors)
}

// isDatabaseError checks if the error is a database-related error
func isDatabaseError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "sql") ||
		strings.Contains(errStr, "database") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "duplicate") ||
		strings.Contains(errStr, "foreign key") ||
		strings.Contains(errStr, "constraint")
}

// getValidationErrorMessage returns a user-friendly validation error message
func getValidationErrorMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Value is too short (minimum " + fieldErr.Param() + " characters)"
	case "max":
		return "Value is too long (maximum " + fieldErr.Param() + " characters)"
	case "len":
		return "Must be exactly " + fieldErr.Param() + " characters long"
	case "numeric":
		return "Must be a number"
	case "alpha":
		return "Must contain only letters"
	case "alphanum":
		return "Must contain only letters and numbers"
	case "url":
		return "Must be a valid URL"
	case "uuid":
		return "Must be a valid UUID"
	case "gte":
		return "Must be greater than or equal to " + fieldErr.Param()
	case "lte":
		return "Must be less than or equal to " + fieldErr.Param()
	case "gt":
		return "Must be greater than " + fieldErr.Param()
	case "lt":
		return "Must be less than " + fieldErr.Param()
	case "oneof":
		return "Must be one of: " + fieldErr.Param()
	default:
		return "Invalid value"
	}
}

// RecoveryHandler handles panics and converts them to errors
func RecoveryHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		// Log the panic
		requestID, _ := c.Get("request_id")
		logrus.WithFields(logrus.Fields{
			"panic":      recovered,
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"client_ip":  c.ClientIP(),
		}).Error("Panic recovered")

		// Convert panic to error and let error handler deal with it
		err := NewAppError(ErrorTypeInternal, "Internal server error", http.StatusInternalServerError,
			errors.New("panic recovered"))
		c.Error(err)

		// Abort the request
		c.Abort()
	})
}

// AbortWithError is a helper function to abort request with custom error
func AbortWithError(c *gin.Context, appErr *AppError) {
	c.Error(appErr)
	c.Abort()
}

// AbortWithValidationError is a helper function to abort with validation error
func AbortWithValidationError(c *gin.Context, message string, details any) {
	err := NewValidationError(message, details)
	AbortWithError(c, err)
}

// AbortWithNotFoundError is a helper function to abort with not found error
func AbortWithNotFoundError(c *gin.Context, message string) {
	err := NewNotFoundError(message)
	AbortWithError(c, err)
}

// AbortWithBusinessError is a helper function to abort with business error
func AbortWithBusinessError(c *gin.Context, message string, code int) {
	err := NewBusinessError(message, code)
	AbortWithError(c, err)
}
