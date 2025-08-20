package utils

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// APIResponse represents the standard API response structure
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ErrorResponse represents error-specific response structure
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details any    `json:"details,omitempty"`
}

// PaginatedResponse represents paginated data response
type PaginatedResponse struct {
	Success    bool       `json:"success"`
	Message    string     `json:"message"`
	Data       any        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Page         int   `json:"page"`
	Limit        int   `json:"limit"`
	Total        int64 `json:"total"`
	TotalPages   int   `json:"total_pages"`
	HasNext      bool  `json:"has_next"`
	HasPrevious  bool  `json:"has_previous"`
	NextPage     *int  `json:"next_page,omitempty"`
	PreviousPage *int  `json:"previous_page,omitempty"`
}

// ResponseValidationError represents a single validation error for API responses
type ResponseValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

// PaginationParams represents pagination query parameters
type PaginationParams struct {
	Page  int `form:"page" binding:"min=1"`
	Limit int `form:"limit" binding:"min=1,max=100"`
}

// SuccessResponse sends a successful response with data
func SuccessResponse(c *gin.Context, statusCode int, message string, data any) {
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

// ErrorResponseWithDetails sends an error response with additional details
func ErrorResponseWithDetails(c *gin.Context, statusCode int, message string, err string, details any) {
	response := ErrorResponse{
		Success: false,
		Message: message,
		Error:   err,
		Code:    statusCode,
		Details: details,
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
func CreatedResponse(c *gin.Context, message string, data any) {
	SuccessResponse(c, http.StatusCreated, message, data)
}

// OKResponse sends a 200 OK response
func OKResponse(c *gin.Context, message string, data any) {
	SuccessResponse(c, http.StatusOK, message, data)
}

// AcceptedResponse sends a 202 Accepted response
func AcceptedResponse(c *gin.Context, message string, data any) {
	SuccessResponse(c, http.StatusAccepted, message, data)
}

// NoContentResponse sends a 204 No Content response
func NoContentResponse(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

// PaginatedSuccessResponse sends a successful paginated response
func PaginatedSuccessResponse(c *gin.Context, message string, data any, pagination Pagination) {
	response := PaginatedResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	}
	c.JSON(http.StatusOK, response)
}

// CreatePagination creates pagination metadata from parameters
func CreatePagination(page, limit int, total int64) Pagination {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	hasNext := page < totalPages
	hasPrevious := page > 1

	pagination := Pagination{
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
	}

	if hasNext {
		nextPage := page + 1
		pagination.NextPage = &nextPage
	}

	if hasPrevious {
		previousPage := page - 1
		pagination.PreviousPage = &previousPage
	}

	return pagination
}

// GetPaginationParams extracts pagination parameters from query string
func GetPaginationParams(c *gin.Context) PaginationParams {
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	return PaginationParams{
		Page:  page,
		Limit: limit,
	}
}

// CalculateOffset calculates the database offset for pagination
func CalculateOffset(page, limit int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * limit
}

// ValidationErrorResponse sends a validation error response with structured errors
func ValidationErrorResponse(c *gin.Context, validationErrors []ResponseValidationError) {
	ErrorResponseWithDetails(c, http.StatusBadRequest, "Validation failed", "Invalid input data", validationErrors)
}

// ValidationErrorResponseSimple sends a validation error response with simple error messages
func ValidationErrorResponseSimple(c *gin.Context, validationErrors []string) {
	response := APIResponse{
		Success: false,
		Message: "Validation failed",
		Error:   "Invalid input data",
		Data:    validationErrors,
	}
	c.JSON(http.StatusBadRequest, response)
}

// ConflictResponse sends a 409 Conflict response
func ConflictResponse(c *gin.Context, message string, err string) {
	ErrorResponseWithCode(c, http.StatusConflict, message, err)
}

// UnprocessableEntityResponse sends a 422 Unprocessable Entity response
func UnprocessableEntityResponse(c *gin.Context, message string, err string) {
	ErrorResponseWithCode(c, http.StatusUnprocessableEntity, message, err)
}

// TooManyRequestsResponse sends a 429 Too Many Requests response
func TooManyRequestsResponse(c *gin.Context, message string) {
	ErrorResponseWithCode(c, http.StatusTooManyRequests, message, "Rate limit exceeded")
}

// ServiceUnavailableResponse sends a 503 Service Unavailable response
func ServiceUnavailableResponse(c *gin.Context, message string) {
	ErrorResponseWithCode(c, http.StatusServiceUnavailable, message, "Service temporarily unavailable")
}

// CreateValidationError creates a structured validation error
func CreateValidationError(field, message string, value any) ResponseValidationError {
	return ResponseValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// ConvertValidationErrors converts ValidationErrors from validator to ResponseValidationError for API responses
func ConvertValidationErrors(validationErrors ValidationErrors) []ResponseValidationError {
	var responseErrors []ResponseValidationError
	for _, err := range validationErrors {
		responseErrors = append(responseErrors, ResponseValidationError{
			Field:   err.Field,
			Message: err.Message,
			Value:   err.Value,
		})
	}
	return responseErrors
}

// ValidationErrorResponseFromValidator sends a validation error response from validator ValidationErrors
func ValidationErrorResponseFromValidator(c *gin.Context, validationErrors ValidationErrors) {
	responseErrors := ConvertValidationErrors(validationErrors)
	ValidationErrorResponse(c, responseErrors)
}

// FormatError formats an error into a consistent string format
func FormatError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// FormatErrorWithContext formats an error with additional context
func FormatErrorWithContext(err error, context string) string {
	if err == nil {
		return context
	}
	return fmt.Sprintf("%s: %s", context, err.Error())
}
