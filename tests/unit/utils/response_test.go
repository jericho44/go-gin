package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gin-golang-app/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSuccessResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	testData := map[string]string{"key": "value"}
	utils.SuccessResponse(c, http.StatusOK, "Success message", testData)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Success message", response.Message)
	assert.NotNil(t, response.Data)
}

func TestErrorResponseWithCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	utils.ErrorResponseWithCode(c, http.StatusBadRequest, "Error message", "Detailed error")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response utils.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Error message", response.Message)
	assert.Equal(t, "Detailed error", response.Error)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestBadRequestResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	utils.BadRequestResponse(c, "Bad request", "Invalid input")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response utils.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Bad request", response.Message)
	assert.Equal(t, "Invalid input", response.Error)
}

func TestUnauthorizedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	utils.UnauthorizedResponse(c, "Unauthorized")

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response utils.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Unauthorized", response.Message)
	assert.Equal(t, "Unauthorized access", response.Error)
}

func TestNotFoundResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	utils.NotFoundResponse(c, "Resource not found")

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response utils.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Resource not found", response.Message)
	assert.Equal(t, "Resource not found", response.Error)
}

func TestCreatedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	testData := map[string]int{"id": 123}
	utils.CreatedResponse(c, "Resource created", testData)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response utils.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Resource created", response.Message)
	assert.NotNil(t, response.Data)
}

func TestOKResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	testData := []string{"item1", "item2"}
	utils.OKResponse(c, "Data retrieved", testData)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Data retrieved", response.Message)
	assert.NotNil(t, response.Data)
}

func TestNoContentResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	utils.NoContentResponse(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	// Note: Gin may still write some content even for 204 responses
}

func TestPaginatedSuccessResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	testData := []string{"item1", "item2"}
	pagination := utils.Pagination{
		Page:       1,
		Limit:      10,
		Total:      2,
		TotalPages: 1,
	}

	utils.PaginatedSuccessResponse(c, "Data retrieved", testData, pagination)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Data retrieved", response.Message)
	assert.NotNil(t, response.Data)
	assert.Equal(t, 1, response.Pagination.Page)
	assert.Equal(t, 10, response.Pagination.Limit)
	assert.Equal(t, int64(2), response.Pagination.Total)
	assert.Equal(t, 1, response.Pagination.TotalPages)
}

func TestValidationErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	validationErrors := []utils.ResponseValidationError{
		{Field: "name", Message: "Name is required", Value: ""},
		{Field: "email", Message: "Email is invalid", Value: "invalid-email"},
	}
	utils.ValidationErrorResponse(c, validationErrors)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response utils.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Validation failed", response.Message)
	assert.Equal(t, "Invalid input data", response.Error)
	assert.NotNil(t, response.Details)
}

func TestValidationErrorResponseSimple(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	validationErrors := []string{"Name is required", "Email is invalid"}
	utils.ValidationErrorResponseSimple(c, validationErrors)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response utils.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Validation failed", response.Message)
	assert.Equal(t, "Invalid input data", response.Error)
	assert.NotNil(t, response.Data)
}

func TestCreatePagination(t *testing.T) {
	tests := []struct {
		name          string
		page          int
		limit         int
		total         int64
		expectedPage  int
		expectedLimit int
		expectedTotal int64
		expectedPages int
		expectedNext  bool
		expectedPrev  bool
	}{
		{
			name:          "First page with results",
			page:          1,
			limit:         10,
			total:         25,
			expectedPage:  1,
			expectedLimit: 10,
			expectedTotal: 25,
			expectedPages: 3,
			expectedNext:  true,
			expectedPrev:  false,
		},
		{
			name:          "Middle page",
			page:          2,
			limit:         10,
			total:         25,
			expectedPage:  2,
			expectedLimit: 10,
			expectedTotal: 25,
			expectedPages: 3,
			expectedNext:  true,
			expectedPrev:  true,
		},
		{
			name:          "Last page",
			page:          3,
			limit:         10,
			total:         25,
			expectedPage:  3,
			expectedLimit: 10,
			expectedTotal: 25,
			expectedPages: 3,
			expectedNext:  false,
			expectedPrev:  true,
		},
		{
			name:          "Invalid page defaults to 1",
			page:          0,
			limit:         10,
			total:         25,
			expectedPage:  1,
			expectedLimit: 10,
			expectedTotal: 25,
			expectedPages: 3,
			expectedNext:  true,
			expectedPrev:  false,
		},
		{
			name:          "Invalid limit defaults to 10",
			page:          1,
			limit:         0,
			total:         25,
			expectedPage:  1,
			expectedLimit: 10,
			expectedTotal: 25,
			expectedPages: 3,
			expectedNext:  true,
			expectedPrev:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pagination := utils.CreatePagination(tt.page, tt.limit, tt.total)

			assert.Equal(t, tt.expectedPage, pagination.Page)
			assert.Equal(t, tt.expectedLimit, pagination.Limit)
			assert.Equal(t, tt.expectedTotal, pagination.Total)
			assert.Equal(t, tt.expectedPages, pagination.TotalPages)
			assert.Equal(t, tt.expectedNext, pagination.HasNext)
			assert.Equal(t, tt.expectedPrev, pagination.HasPrevious)

			if tt.expectedNext {
				assert.NotNil(t, pagination.NextPage)
				assert.Equal(t, tt.expectedPage+1, *pagination.NextPage)
			} else {
				assert.Nil(t, pagination.NextPage)
			}

			if tt.expectedPrev {
				assert.NotNil(t, pagination.PreviousPage)
				assert.Equal(t, tt.expectedPage-1, *pagination.PreviousPage)
			} else {
				assert.Nil(t, pagination.PreviousPage)
			}
		})
	}
}

func TestGetPaginationParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		queryParams   map[string]string
		expectedPage  int
		expectedLimit int
	}{
		{
			name:          "Default values",
			queryParams:   map[string]string{},
			expectedPage:  1,
			expectedLimit: 10,
		},
		{
			name:          "Valid page and limit",
			queryParams:   map[string]string{"page": "2", "limit": "20"},
			expectedPage:  2,
			expectedLimit: 20,
		},
		{
			name:          "Invalid page defaults to 1",
			queryParams:   map[string]string{"page": "0", "limit": "20"},
			expectedPage:  1,
			expectedLimit: 20,
		},
		{
			name:          "Invalid limit defaults to 10",
			queryParams:   map[string]string{"page": "2", "limit": "0"},
			expectedPage:  2,
			expectedLimit: 10,
		},
		{
			name:          "Limit over 100 defaults to 10",
			queryParams:   map[string]string{"page": "1", "limit": "150"},
			expectedPage:  1,
			expectedLimit: 10,
		},
		{
			name:          "Non-numeric values default",
			queryParams:   map[string]string{"page": "abc", "limit": "xyz"},
			expectedPage:  1,
			expectedLimit: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set up query parameters
			req := httptest.NewRequest("GET", "/test", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()
			c.Request = req

			params := utils.GetPaginationParams(c)

			assert.Equal(t, tt.expectedPage, params.Page)
			assert.Equal(t, tt.expectedLimit, params.Limit)
		})
	}
}

func TestCalculateOffset(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		limit          int
		expectedOffset int
	}{
		{
			name:           "First page",
			page:           1,
			limit:          10,
			expectedOffset: 0,
		},
		{
			name:           "Second page",
			page:           2,
			limit:          10,
			expectedOffset: 10,
		},
		{
			name:           "Third page with different limit",
			page:           3,
			limit:          20,
			expectedOffset: 40,
		},
		{
			name:           "Invalid page defaults to 1",
			page:           0,
			limit:          10,
			expectedOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := utils.CalculateOffset(tt.page, tt.limit)
			assert.Equal(t, tt.expectedOffset, offset)
		})
	}
}

func TestErrorResponseWithDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	details := map[string]string{"field": "email", "reason": "invalid format"}
	utils.ErrorResponseWithDetails(c, http.StatusBadRequest, "Validation error", "Invalid input", details)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response utils.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Validation error", response.Message)
	assert.Equal(t, "Invalid input", response.Error)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotNil(t, response.Details)
}

func TestAcceptedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	testData := map[string]string{"status": "processing"}
	utils.AcceptedResponse(c, "Request accepted", testData)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var response utils.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Request accepted", response.Message)
	assert.NotNil(t, response.Data)
}

func TestConflictResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	utils.ConflictResponse(c, "Resource conflict", "Email already exists")

	assert.Equal(t, http.StatusConflict, w.Code)

	var response utils.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Resource conflict", response.Message)
	assert.Equal(t, "Email already exists", response.Error)
}

func TestUnprocessableEntityResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	utils.UnprocessableEntityResponse(c, "Cannot process", "Invalid data format")

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var response utils.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Cannot process", response.Message)
	assert.Equal(t, "Invalid data format", response.Error)
}

func TestTooManyRequestsResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	utils.TooManyRequestsResponse(c, "Rate limit exceeded")

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	var response utils.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Rate limit exceeded", response.Message)
	assert.Equal(t, "Rate limit exceeded", response.Error)
}

func TestServiceUnavailableResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	utils.ServiceUnavailableResponse(c, "Service down")

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response utils.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Service down", response.Message)
	assert.Equal(t, "Service temporarily unavailable", response.Error)
}

func TestCreateValidationError(t *testing.T) {
	validationError := utils.CreateValidationError("email", "Invalid email format", "invalid-email")

	assert.Equal(t, "email", validationError.Field)
	assert.Equal(t, "Invalid email format", validationError.Message)
	assert.Equal(t, "invalid-email", validationError.Value)
}

func TestFormatError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "Valid error",
			err:      assert.AnError,
			expected: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FormatError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatErrorWithContext(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		context  string
		expected string
	}{
		{
			name:     "Nil error",
			err:      nil,
			context:  "Database operation",
			expected: "Database operation",
		},
		{
			name:     "Valid error with context",
			err:      assert.AnError,
			context:  "Database operation",
			expected: "Database operation: " + assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FormatErrorWithContext(tt.err, tt.context)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertValidationErrors(t *testing.T) {
	validationErrors := utils.ValidationErrors{
		{Field: "name", Tag: "required", Value: "", Message: "Name is required"},
		{Field: "email", Tag: "email", Value: "invalid", Message: "Email is invalid"},
	}

	responseErrors := utils.ConvertValidationErrors(validationErrors)

	assert.Len(t, responseErrors, 2)
	assert.Equal(t, "name", responseErrors[0].Field)
	assert.Equal(t, "Name is required", responseErrors[0].Message)
	assert.Equal(t, "", responseErrors[0].Value)
	assert.Equal(t, "email", responseErrors[1].Field)
	assert.Equal(t, "Email is invalid", responseErrors[1].Message)
	assert.Equal(t, "invalid", responseErrors[1].Value)
}

func TestValidationErrorResponseFromValidator(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	validationErrors := utils.ValidationErrors{
		{Field: "name", Tag: "required", Value: "", Message: "Name is required"},
		{Field: "email", Tag: "email", Value: "invalid", Message: "Email is invalid"},
	}

	utils.ValidationErrorResponseFromValidator(c, validationErrors)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response utils.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Validation failed", response.Message)
	assert.Equal(t, "Invalid input data", response.Error)
	assert.NotNil(t, response.Details)
}
