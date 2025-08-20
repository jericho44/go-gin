package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"gin-golang-app/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) Exists(ctx context.Context, id uint) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// Implement base repository methods
func (m *MockUserRepository) GetDB() *sql.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*sql.DB)
}

func (m *MockUserRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sql.Tx), args.Error(1)
}

func TestNewUserService(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)

	assert.NotNil(t, service)
	assert.IsType(t, &userService{}, service)
}

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name          string
		user          *models.User
		setupMock     func(*MockUserRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name: "successful user creation",
			user: &models.User{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("ExistsByEmail", mock.Anything, "john@example.com").Return(false, nil)
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name: "user with existing email",
			user: &models.User{
				Name:  "John Doe",
				Email: "existing@example.com",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("ExistsByEmail", mock.Anything, "existing@example.com").Return(true, nil)
			},
			expectedError: "user with this email already exists",
		},
		{
			name: "invalid email format",
			user: &models.User{
				Name:  "John Doe",
				Email: "invalid-email",
			},
			setupMock:     func(m *MockUserRepository) {},
			expectedError: "invalid email format",
		},
		{
			name: "empty name",
			user: &models.User{
				Name:  "",
				Email: "john@example.com",
			},
			setupMock:     func(m *MockUserRepository) {},
			expectedError: "name is required",
		},
		{
			name: "name too short",
			user: &models.User{
				Name:  "J",
				Email: "john@example.com",
			},
			setupMock:     func(m *MockUserRepository) {},
			expectedError: "name must be at least 2 characters long",
		},
		{
			name:          "nil user",
			user:          nil,
			setupMock:     func(m *MockUserRepository) {},
			expectedError: "user cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			tt.setupMock(mockRepo)

			service := NewUserService(mockRepo)
			result, err := service.CreateUser(context.Background(), tt.user)

			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "john@example.com", result.Email) // Email should be normalized to lowercase
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	tests := []struct {
		name          string
		userID        uint
		setupMock     func(*MockUserRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name:   "successful user retrieval",
			userID: 1,
			setupMock: func(m *MockUserRepository) {
				user := &models.User{ID: 1, Name: "John Doe", Email: "john@example.com"}
				m.On("GetByID", mock.Anything, uint(1)).Return(user, nil)
			},
			expectSuccess: true,
		},
		{
			name:   "user not found",
			userID: 999,
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.New("user with id 999 not found"))
			},
			expectedError: "user not found",
		},
		{
			name:          "invalid user ID",
			userID:        0,
			setupMock:     func(m *MockUserRepository) {},
			expectedError: "user ID must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			tt.setupMock(mockRepo)

			service := NewUserService(mockRepo)
			result, err := service.GetUserByID(context.Background(), tt.userID)

			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.userID, result.ID)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	tests := []struct {
		name          string
		user          *models.User
		setupMock     func(*MockUserRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name: "successful user update",
			user: &models.User{
				ID:    1,
				Name:  "John Updated",
				Email: "john.updated@example.com",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("Exists", mock.Anything, uint(1)).Return(true, nil)
				m.On("GetByEmail", mock.Anything, "john.updated@example.com").Return(nil, errors.New("not found"))
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name: "user not found",
			user: &models.User{
				ID:    999,
				Name:  "John Updated",
				Email: "john.updated@example.com",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("Exists", mock.Anything, uint(999)).Return(false, nil)
			},
			expectedError: "user not found",
		},
		{
			name: "email already taken by another user",
			user: &models.User{
				ID:    1,
				Name:  "John Updated",
				Email: "taken@example.com",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("Exists", mock.Anything, uint(1)).Return(true, nil)
				existingUser := &models.User{ID: 2, Email: "taken@example.com"}
				m.On("GetByEmail", mock.Anything, "taken@example.com").Return(existingUser, nil)
			},
			expectedError: "user with this email already exists",
		},
		{
			name: "missing user ID",
			user: &models.User{
				Name:  "John Updated",
				Email: "john.updated@example.com",
			},
			setupMock:     func(m *MockUserRepository) {},
			expectedError: "user ID is required for update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			tt.setupMock(mockRepo)

			service := NewUserService(mockRepo)
			result, err := service.UpdateUser(context.Background(), tt.user)

			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	tests := []struct {
		name          string
		userID        uint
		setupMock     func(*MockUserRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name:   "successful user deletion",
			userID: 1,
			setupMock: func(m *MockUserRepository) {
				m.On("Exists", mock.Anything, uint(1)).Return(true, nil)
				m.On("Delete", mock.Anything, uint(1)).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name:   "user not found",
			userID: 999,
			setupMock: func(m *MockUserRepository) {
				m.On("Exists", mock.Anything, uint(999)).Return(false, nil)
			},
			expectedError: "user not found",
		},
		{
			name:          "invalid user ID",
			userID:        0,
			setupMock:     func(m *MockUserRepository) {},
			expectedError: "user ID must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			tt.setupMock(mockRepo)

			service := NewUserService(mockRepo)
			err := service.DeleteUser(context.Background(), tt.userID)

			if tt.expectSuccess {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_ListUsers(t *testing.T) {
	tests := []struct {
		name          string
		page          int
		limit         int
		setupMock     func(*MockUserRepository)
		expectedError string
		expectSuccess bool
		expectedPage  int
		expectedLimit int
	}{
		{
			name:  "successful user listing",
			page:  1,
			limit: 10,
			setupMock: func(m *MockUserRepository) {
				users := []*models.User{
					{ID: 1, Name: "User 1", Email: "user1@example.com"},
					{ID: 2, Name: "User 2", Email: "user2@example.com"},
				}
				m.On("List", mock.Anything, 10, 0).Return(users, nil)
				m.On("Count", mock.Anything).Return(int64(2), nil)
			},
			expectSuccess: true,
			expectedPage:  1,
			expectedLimit: 10,
		},
		{
			name:  "default pagination values",
			page:  0,
			limit: 0,
			setupMock: func(m *MockUserRepository) {
				users := []*models.User{}
				m.On("List", mock.Anything, DefaultLimit, 0).Return(users, nil)
				m.On("Count", mock.Anything).Return(int64(0), nil)
			},
			expectSuccess: true,
			expectedPage:  DefaultPage,
			expectedLimit: DefaultLimit,
		},
		{
			name:  "limit exceeds maximum",
			page:  1,
			limit: 200,
			setupMock: func(m *MockUserRepository) {
				users := []*models.User{}
				m.On("List", mock.Anything, MaxPageSize, 0).Return(users, nil)
				m.On("Count", mock.Anything).Return(int64(0), nil)
			},
			expectSuccess: true,
			expectedPage:  1,
			expectedLimit: MaxPageSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			tt.setupMock(mockRepo)

			service := NewUserService(mockRepo)
			users, total, err := service.ListUsers(context.Background(), tt.page, tt.limit)

			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, users)
				assert.GreaterOrEqual(t, total, int64(0))
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_ValidationMethods(t *testing.T) {
	service := &userService{}

	t.Run("validateName", func(t *testing.T) {
		tests := []struct {
			name          string
			input         string
			expectedError string
		}{
			{"valid name", "John Doe", ""},
			{"empty name", "", "name is required"},
			{"name too short", "J", "name must be at least 2 characters long"},
			{"name too long", strings.Repeat("a", 101), "name must not exceed 100 characters"},
			{"name with spaces", "  John Doe  ", ""}, // Should be trimmed
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := service.validateName(tt.input)
				if tt.expectedError == "" {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			})
		}
	})

	t.Run("validateEmail", func(t *testing.T) {
		tests := []struct {
			name          string
			input         string
			expectedError string
		}{
			{"valid email", "john@example.com", ""},
			{"empty email", "", "email is required"},
			{"invalid email format", "invalid-email", "email format is invalid"},
			{"email without domain", "john@", "email format is invalid"},
			{"email without @", "johnexample.com", "email format is invalid"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := service.validateEmail(tt.input)
				if tt.expectedError == "" {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			})
		}
	})
}
