package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSuccessResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	testData := map[string]string{"key": "value"}
	SuccessResponse(c, http.StatusOK, "Success message", testData)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
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

	ErrorResponseWithCode(c, http.StatusBadRequest, "Error message", "Detailed error")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
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

	BadRequestResponse(c, "Bad request", "Invalid input")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
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

	UnauthorizedResponse(c, "Unauthorized")

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
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

	NotFoundResponse(c, "Resource not found")

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
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
	CreatedResponse(c, "Resource created", testData)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response APIResponse
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
	OKResponse(c, "Data retrieved", testData)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
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

	NoContentResponse(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestPaginatedSuccessResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	testData := []string{"item1", "item2"}
	pagination := Pagination{
		Page:       1,
		Limit:      10,
		Total:      2,
		TotalPages: 1,
	}

	PaginatedSuccessResponse(c, "Data retrieved", testData, pagination)

	assert.Equal(t, http.StatusOK, w.Code)

	var response PaginatedResponse
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

	validationErrors := []string{"Name is required", "Email is invalid"}
	ValidationErrorResponse(c, validationErrors)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Validation failed", response.Message)
	assert.Equal(t, "Invalid input data", response.Error)
	assert.NotNil(t, response.Data)
}
