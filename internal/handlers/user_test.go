package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gin-golang-app/internal/models"
	"gin-golang-app/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) ListUsers(ctx context.Context, page, limit int) ([]*models.User, int64, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserService) UserExists(ctx context.Context, id uint) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestUserHandler_CreateUser(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful user creation",
			requestBody: CreateUserRequest{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			mockSetup: func(m *MockUserService) {
				expectedUser := &models.User{
					Name:  "John Doe",
					Email: "john@example.com",
				}
				returnUser := &models.User{
					ID:        1,
					Name:      "John Doe",
					Email:     "john@example.com",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				m.On("CreateUser", mock.Anything, expectedUser).Return(returnUser, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"name":  "",
				"email": "invalid-email",
			},
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "user already exists",
			requestBody: CreateUserRequest{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			mockSetup: func(m *MockUserService) {
				m.On("CreateUser", mock.Anything, mock.Anything).Return(nil, services.ErrUserAlreadyExists)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			requestBody: CreateUserRequest{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			mockSetup: func(m *MockUserService) {
				m.On("CreateUser", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockUserService)
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService)
			router := setupTestRouter()
			router.POST("/users", handler.CreateUser)

			// Prepare request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockUserService)
		expectedStatus int
	}{
		{
			name:   "successful user retrieval",
			userID: "1",
			mockSetup: func(m *MockUserService) {
				user := &models.User{
					ID:    1,
					Name:  "John Doe",
					Email: "john@example.com",
				}
				m.On("GetUserByID", mock.Anything, uint(1)).Return(user, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid",
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "user not found",
			userID: "999",
			mockSetup: func(m *MockUserService) {
				m.On("GetUserByID", mock.Anything, uint(999)).Return(nil, services.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockUserService)
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService)
			router := setupTestRouter()
			router.GET("/users/:id", handler.GetUser)

			// Prepare request
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID, nil)
			w := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		mockSetup      func(*MockUserService)
		expectedStatus int
	}{
		{
			name:   "successful user update",
			userID: "1",
			requestBody: UpdateUserRequest{
				Name:  "John Updated",
				Email: "john.updated@example.com",
			},
			mockSetup: func(m *MockUserService) {
				expectedUser := &models.User{
					ID:    1,
					Name:  "John Updated",
					Email: "john.updated@example.com",
				}
				m.On("UpdateUser", mock.Anything, expectedUser).Return(expectedUser, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid",
			requestBody:    UpdateUserRequest{Name: "John", Email: "john@example.com"},
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "user not found",
			userID: "999",
			requestBody: UpdateUserRequest{
				Name:  "John",
				Email: "john@example.com",
			},
			mockSetup: func(m *MockUserService) {
				m.On("UpdateUser", mock.Anything, mock.Anything).Return(nil, services.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockUserService)
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService)
			router := setupTestRouter()
			router.PUT("/users/:id", handler.UpdateUser)

			// Prepare request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.userID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockUserService)
		expectedStatus int
	}{
		{
			name:   "successful user deletion",
			userID: "1",
			mockSetup: func(m *MockUserService) {
				m.On("DeleteUser", mock.Anything, uint(1)).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid",
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "user not found",
			userID: "999",
			mockSetup: func(m *MockUserService) {
				m.On("DeleteUser", mock.Anything, uint(999)).Return(services.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockUserService)
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService)
			router := setupTestRouter()
			router.DELETE("/users/:id", handler.DeleteUser)

			// Prepare request
			req := httptest.NewRequest(http.MethodDelete, "/users/"+tt.userID, nil)
			w := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockUserService)
		expectedStatus int
	}{
		{
			name:        "successful user listing with default pagination",
			queryParams: "",
			mockSetup: func(m *MockUserService) {
				users := []*models.User{
					{ID: 1, Name: "John", Email: "john@example.com"},
					{ID: 2, Name: "Jane", Email: "jane@example.com"},
				}
				m.On("ListUsers", mock.Anything, 1, 10).Return(users, int64(2), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful user listing with custom pagination",
			queryParams: "?page=2&limit=5",
			mockSetup: func(m *MockUserService) {
				users := []*models.User{
					{ID: 6, Name: "User 6", Email: "user6@example.com"},
				}
				m.On("ListUsers", mock.Anything, 2, 5).Return(users, int64(11), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "service error",
			queryParams: "",
			mockSetup: func(m *MockUserService) {
				m.On("ListUsers", mock.Anything, 1, 10).Return(nil, int64(0), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockUserService)
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService)
			router := setupTestRouter()
			router.GET("/users", handler.ListUsers)

			// Prepare request
			req := httptest.NewRequest(http.MethodGet, "/users"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUserByEmail(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		mockSetup      func(*MockUserService)
		expectedStatus int
	}{
		{
			name:  "successful user retrieval by email",
			email: "john@example.com",
			mockSetup: func(m *MockUserService) {
				user := &models.User{
					ID:    1,
					Name:  "John Doe",
					Email: "john@example.com",
				}
				m.On("GetUserByEmail", mock.Anything, "john@example.com").Return(user, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "invalid email format",
			email: "invalid-email",
			mockSetup: func(m *MockUserService) {
				m.On("GetUserByEmail", mock.Anything, "invalid-email").Return(nil, services.ErrInvalidEmail)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "user not found",
			email: "notfound@example.com",
			mockSetup: func(m *MockUserService) {
				m.On("GetUserByEmail", mock.Anything, "notfound@example.com").Return(nil, services.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockUserService)
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService)
			router := setupTestRouter()
			router.GET("/users/email/:email", handler.GetUserByEmail)

			// Prepare request
			url := "/users/email/" + tt.email
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_handleServiceError(t *testing.T) {
	tests := []struct {
		name           string
		error          error
		expectedStatus int
	}{
		{
			name:           "user not found error",
			error:          services.ErrUserNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "user already exists error",
			error:          services.ErrUserAlreadyExists,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid input error",
			error:          services.ErrInvalidInput,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid email error",
			error:          services.ErrInvalidEmail,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid name error",
			error:          services.ErrInvalidName,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "generic error",
			error:          errors.New("some generic error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockUserService)
			handler := NewUserHandler(mockService)

			router := setupTestRouter()
			router.GET("/test", func(c *gin.Context) {
				handler.handleServiceError(c, tt.error)
			})

			// Prepare request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
