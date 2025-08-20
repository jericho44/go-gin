package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse represents the standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ErrorResponse represents error-specific response structure
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
	Code    int    `json:"code"`
}

// PaginatedResponse represents paginated data response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// SuccessResponse sends a successful response with data
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.JSON(statusCode, response)
}

// ErrorResponseWithCode sends an error response with custom status code
func ErrorResponseWithCode(c *gin.Context, statusCode int, message string, err string) {
	response := ErrorResponse{
		Success: false,
		Message: message,
		Error:   err,
		Code:    statusCode,
	}
	c.JSON(statusCode, response)
}

// BadRequestResponse sends a 400 Bad Request response
func BadRequestResponse(c *gin.Context, message string, err string) {
	ErrorResponseWithCode(c, http.StatusBadRequest, message, err)
}

// UnauthorizedResponse sends a 401 Unauthorized response
func UnauthorizedResponse(c *gin.Context, message string) {
	ErrorResponseWithCode(c, http.StatusUnauthorized, message, "Unauthorized access")
}

// ForbiddenResponse sends a 403 Forbidden response
func ForbiddenResponse(c *gin.Context, message string) {
	ErrorResponseWithCode(c, http.StatusForbidden, message, "Access forbidden")
}

// NotFoundResponse sends a 404 Not Found response
func NotFoundResponse(c *gin.Context, message string) {
	ErrorResponseWithCode(c, http.StatusNotFound, message, "Resource not found")
}

// InternalServerErrorResponse sends a 500 Internal Server Error response
func InternalServerErrorResponse(c *gin.Context, message string, err string) {
	ErrorResponseWithCode(c, http.StatusInternalServerError, message, err)
}

// CreatedResponse sends a 201 Created response
func CreatedResponse(c *gin.Context, message string, data interface{}) {
	SuccessResponse(c, http.StatusCreated, message, data)
}

// OKResponse sends a 200 OK response
func OKResponse(c *gin.Context, message string, data interface{}) {
	SuccessResponse(c, http.StatusOK, message, data)
}

// NoContentResponse sends a 204 No Content response
func NoContentResponse(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

// PaginatedSuccessResponse sends a successful paginated response
func PaginatedSuccessResponse(c *gin.Context, message string, data interface{}, pagination Pagination) {
	response := PaginatedResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	}
	c.JSON(http.StatusOK, response)
}

// ValidationErrorResponse sends a validation error response
func ValidationErrorResponse(c *gin.Context, validationErrors []string) {
	response := APIResponse{
		Success: false,
		Message: "Validation failed",
		Error:   "Invalid input data",
		Data:    validationErrors,
	}
	c.JSON(http.StatusBadRequest, response)
}
