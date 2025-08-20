package middleware

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppError(t *testing.T) {
	err := errors.New("original error")
	appErr := NewAppError(ErrorTypeValidation, "validation failed", http.StatusBadRequest, err)

	assert.Equal(t, ErrorTypeValidation, appErr.Type)
	assert.Equal(t, "validation failed", appErr.Message)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	assert.Equal(t, err, appErr.Err)
	assert.Equal(t, "original error", appErr.Error())
}

func TestNewValidationError(t *testing.T) {
	details := map[string]string{"field": "error"}
	appErr := NewValidationError("validation failed", details)

	assert.Equal(t, ErrorTypeValidation, appErr.Type)
	assert.Equal(t, "validation failed", appErr.Message)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	assert.Equal(t, details, appErr.Details)
}

func TestNewDatabaseError(t *testing.T) {
	originalErr := errors.New("database connection failed")
	appErr := NewDatabaseError("database error", originalErr)

	assert.Equal(t, ErrorTypeDatabase, appErr.Type)
	assert.Equal(t, "database error", appErr.Message)
	assert.Equal(t, http.StatusInternalServerError, appErr.Code)
	assert.Equal(t, originalErr, appErr.Err)
}

func TestNewBusinessError(t *testing.T) {
	appErr := NewBusinessError("business rule violated", http.StatusConflict)

	assert.Equal(t, ErrorTypeBusiness, appErr.Type)
	assert.Equal(t, "business rule violated", appErr.Message)
	assert.Equal(t, http.StatusConflict, appErr.Code)
}

func TestNewNotFoundError(t *testing.T) {
	appErr := NewNotFoundError("user not found")

	assert.Equal(t, ErrorTypeNotFound, appErr.Type)
	assert.Equal(t, "user not found", appErr.Message)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
}

func TestErrorHandler_NoErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create request
	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req

	// Create middleware
	middleware := ErrorHandler()

	// Execute middleware with no errors
	middleware(c)

	// Should not write any response
	assert.Equal(t, 200, w.Code) // Default status
	assert.Empty(t, w.Body.String())
}

func TestErrorHandler_AppError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create request
	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req

	// Add request ID for logging
	c.Set("request_id", "test-request-id")

	// Add an app error
	appErr := NewValidationError("validation failed", nil)
	c.Error(appErr)

	// Create and execute middleware
	middleware := ErrorHandler()
	middleware(c)

	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "validation failed")
}

func TestErrorHandler_SQLNoRows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create request
	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req

	// Add SQL no rows error
	c.Error(sql.ErrNoRows)

	// Create and execute middleware
	middleware := ErrorHandler()
	middleware(c)

	// Should return 404 Not Found
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Resource not found")
}

func TestErrorHandler_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create request
	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req

	// Add database error
	dbErr := errors.New("duplicate key constraint violation")
	c.Error(dbErr)

	// Create and execute middleware
	middleware := ErrorHandler()
	middleware(c)

	// Should return 409 Conflict for duplicate key
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "Resource already exists")
}

func TestErrorHandler_UnknownError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create request
	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req

	// Add unknown error
	unknownErr := errors.New("some unknown error")
	c.Error(unknownErr)

	// Create and execute middleware
	middleware := ErrorHandler()
	middleware(c)

	// Should return 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "An unexpected error occurred")
}

func TestHandleAppError_ValidationWithDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	details := []map[string]string{{"field": "name", "error": "required"}}
	appErr := NewValidationError("validation failed", details)

	handleAppError(c, appErr)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "validation failed")
	assert.Contains(t, w.Body.String(), "details")
}

func TestHandleAppError_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	appErr := NewNotFoundError("user not found")

	handleAppError(c, appErr)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "user not found")
}

func TestHandleAppError_Business(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	appErr := NewBusinessError("insufficient funds", http.StatusPaymentRequired)

	handleAppError(c, appErr)

	assert.Equal(t, http.StatusPaymentRequired, w.Code)
	assert.Contains(t, w.Body.String(), "insufficient funds")
}

func TestIsDatabaseError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "SQL error",
			err:      errors.New("SQL syntax error"),
			expected: true,
		},
		{
			name:     "Database connection error",
			err:      errors.New("database connection failed"),
			expected: true,
		},
		{
			name:     "Duplicate key error",
			err:      errors.New("duplicate key constraint"),
			expected: true,
		},
		{
			name:     "Foreign key error",
			err:      errors.New("foreign key violation"),
			expected: true,
		},
		{
			name:     "Regular error",
			err:      errors.New("some regular error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDatabaseError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRecoveryHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create request
	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req

	// Add request ID
	c.Set("request_id", "test-request-id")

	// Create recovery middleware
	recovery := RecoveryHandler()

	// Test that recovery middleware is created successfully
	assert.NotNil(t, recovery)
}

func TestAbortWithError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	appErr := NewValidationError("test error", nil)

	AbortWithError(c, appErr)

	assert.True(t, c.IsAborted())
	assert.Len(t, c.Errors, 1)
	assert.Equal(t, appErr, c.Errors[0].Err)
}

func TestAbortWithValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	details := map[string]string{"field": "error"}

	AbortWithValidationError(c, "validation failed", details)

	assert.True(t, c.IsAborted())
	assert.Len(t, c.Errors, 1)

	var appErr *AppError
	require.True(t, errors.As(c.Errors[0].Err, &appErr))
	assert.Equal(t, ErrorTypeValidation, appErr.Type)
	assert.Equal(t, "validation failed", appErr.Message)
	assert.Equal(t, details, appErr.Details)
}

func TestAbortWithNotFoundError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	AbortWithNotFoundError(c, "user not found")

	assert.True(t, c.IsAborted())
	assert.Len(t, c.Errors, 1)

	var appErr *AppError
	require.True(t, errors.As(c.Errors[0].Err, &appErr))
	assert.Equal(t, ErrorTypeNotFound, appErr.Type)
	assert.Equal(t, "user not found", appErr.Message)
}

func TestAbortWithBusinessError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	AbortWithBusinessError(c, "business rule violated", http.StatusConflict)

	assert.True(t, c.IsAborted())
	assert.Len(t, c.Errors, 1)

	var appErr *AppError
	require.True(t, errors.As(c.Errors[0].Err, &appErr))
	assert.Equal(t, ErrorTypeBusiness, appErr.Type)
	assert.Equal(t, "business rule violated", appErr.Message)
	assert.Equal(t, http.StatusConflict, appErr.Code)
}
