package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gin-golang-app/internal/models"
	"gin-golang-app/internal/repository"
)

// UserService defines the interface for user business logic operations
type UserService interface {
	// CreateUser creates a new user with validation
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	// UpdateUser updates an existing user with validation
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
	// DeleteUser removes a user by ID
	DeleteUser(ctx context.Context, id uint) error
	// ListUsers retrieves a paginated list of users
	ListUsers(ctx context.Context, page, limit int) ([]*models.User, int64, error)
	// UserExists checks if a user exists by ID
	UserExists(ctx context.Context, id uint) (bool, error)
}

// userService implements the UserService interface
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// Business rule constants
const (
	MinNameLength = 2
	MaxNameLength = 100
	MinPageSize   = 1
	MaxPageSize   = 100
	DefaultPage   = 1
	DefaultLimit  = 10
)

// Custom error types for business logic
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrInvalidInput      = errors.New("invalid input data")
	ErrInvalidEmail      = errors.New("invalid email format")
	ErrInvalidName       = errors.New("invalid name")
	ErrInvalidPagination = errors.New("invalid pagination parameters")
)

// validateUser performs comprehensive validation on user data
func (s *userService) validateUser(user *models.User) error {
	if user == nil {
		return fmt.Errorf("%w: user cannot be nil", ErrInvalidInput)
	}

	// Validate name
	if err := s.validateName(user.Name); err != nil {
		return err
	}

	// Validate email
	if err := s.validateEmail(user.Email); err != nil {
		return err
	}

	return nil
}

// validateName validates user name according to business rules
func (s *userService) validateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidName)
	}
	if len(name) < MinNameLength {
		return fmt.Errorf("%w: name must be at least %d characters long", ErrInvalidName, MinNameLength)
	}
	if len(name) > MaxNameLength {
		return fmt.Errorf("%w: name must not exceed %d characters", ErrInvalidName, MaxNameLength)
	}
	return nil
}

// validateEmail validates email format using regex
func (s *userService) validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidEmail)
	}

	// Basic email regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("%w: email format is invalid", ErrInvalidEmail)
	}

	return nil
}

// validatePagination validates pagination parameters
func (s *userService) validatePagination(page, limit int) (int, int, error) {
	// Set defaults if values are zero or negative
	if page <= 0 {
		page = DefaultPage
	}
	if limit <= 0 {
		limit = DefaultLimit
	}

	// Enforce maximum page size
	if limit > MaxPageSize {
		limit = MaxPageSize
	}

	return page, limit, nil
}

// CreateUser creates a new user with validation and business rule enforcement
func (s *userService) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	// Validate input
	if err := s.validateUser(user); err != nil {
		return nil, err
	}

	// Normalize data
	user.Name = strings.TrimSpace(user.Name)
	user.Email = strings.TrimSpace(strings.ToLower(user.Email))

	// Check if user with email already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("%w: %s", ErrUserAlreadyExists, user.Email)
	}

	// Create user
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID with business logic
func (s *userService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	if id == 0 {
		return nil, fmt.Errorf("%w: user ID must be greater than 0", ErrInvalidInput)
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		// Check if it's a not found error and wrap it appropriately
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("%w: user with ID %d", ErrUserNotFound, id)
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email with business logic
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if err := s.validateEmail(email); err != nil {
		return nil, err
	}

	// Normalize email
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Check if it's a not found error and wrap it appropriately
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("%w: user with email %s", ErrUserNotFound, email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// UpdateUser updates an existing user with validation and business rule enforcement
func (s *userService) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	// Validate input
	if err := s.validateUser(user); err != nil {
		return nil, err
	}

	if user.ID == 0 {
		return nil, fmt.Errorf("%w: user ID is required for update", ErrInvalidInput)
	}

	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("%w: user with ID %d", ErrUserNotFound, user.ID)
	}

	// Normalize data
	user.Name = strings.TrimSpace(user.Name)
	user.Email = strings.TrimSpace(strings.ToLower(user.Email))

	// Check if email is already taken by another user
	existingUser, err := s.userRepo.GetByEmail(ctx, user.Email)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if existingUser != nil && existingUser.ID != user.ID {
		return nil, fmt.Errorf("%w: %s", ErrUserAlreadyExists, user.Email)
	}

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser removes a user by ID with business logic validation
func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	if id == 0 {
		return fmt.Errorf("%w: user ID must be greater than 0", ErrInvalidInput)
	}

	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("%w: user with ID %d", ErrUserNotFound, id)
	}

	// Delete user
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers retrieves a paginated list of users with validation
func (s *userService) ListUsers(ctx context.Context, page, limit int) ([]*models.User, int64, error) {
	// Validate and normalize pagination parameters
	page, limit, err := s.validatePagination(page, limit)
	if err != nil {
		return nil, 0, err
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get users
	users, err := s.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// Get total count
	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	return users, total, nil
}

// UserExists checks if a user exists by ID
func (s *userService) UserExists(ctx context.Context, id uint) (bool, error) {
	if id == 0 {
		return false, fmt.Errorf("%w: user ID must be greater than 0", ErrInvalidInput)
	}

	exists, err := s.userRepo.Exists(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return exists, nil
}
