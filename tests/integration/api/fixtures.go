package api

import (
	"time"
)

// TestUser represents a test user fixture
type TestUser struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserFixtures provides predefined user data for testing
var UserFixtures = struct {
	ValidUser1   TestUser
	ValidUser2   TestUser
	ValidUser3   TestUser
	UpdatedUser1 TestUser
}{
	ValidUser1: TestUser{
		Name:  "John Doe",
		Email: "john.doe@example.com",
	},
	ValidUser2: TestUser{
		Name:  "Jane Smith",
		Email: "jane.smith@example.com",
	},
	ValidUser3: TestUser{
		Name:  "Bob Johnson",
		Email: "bob.johnson@example.com",
	},
	UpdatedUser1: TestUser{
		Name:  "John Updated",
		Email: "john.updated@example.com",
	},
}

// CreateUserRequest represents the JSON structure for creating a user
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UpdateUserRequest represents the JSON structure for updating a user
type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// APIResponse represents the standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// InvalidUserRequests provides invalid user data for testing validation
var InvalidUserRequests = struct {
	EmptyName    CreateUserRequest
	EmptyEmail   CreateUserRequest
	InvalidEmail CreateUserRequest
	TooShortName CreateUserRequest
	TooLongName  CreateUserRequest
	MissingName  map[string]interface{}
	MissingEmail map[string]interface{}
	InvalidJSON  string
	EmptyRequest map[string]interface{}
}{
	EmptyName: CreateUserRequest{
		Name:  "",
		Email: "test@example.com",
	},
	EmptyEmail: CreateUserRequest{
		Name:  "Test User",
		Email: "",
	},
	InvalidEmail: CreateUserRequest{
		Name:  "Test User",
		Email: "invalid-email",
	},
	TooShortName: CreateUserRequest{
		Name:  "A",
		Email: "test@example.com",
	},
	TooLongName: CreateUserRequest{
		Name:  "This is a very long name that exceeds the maximum allowed length for a user name field in our system",
		Email: "test@example.com",
	},
	MissingName: map[string]interface{}{
		"email": "test@example.com",
	},
	MissingEmail: map[string]interface{}{
		"name": "Test User",
	},
	InvalidJSON:  `{"name": "Test User", "email": "test@example.com"`,
	EmptyRequest: map[string]interface{}{},
}

// ValidUserRequests provides valid user data for testing
var ValidUserRequests = struct {
	User1 CreateUserRequest
	User2 CreateUserRequest
	User3 CreateUserRequest
}{
	User1: CreateUserRequest{
		Name:  UserFixtures.ValidUser1.Name,
		Email: UserFixtures.ValidUser1.Email,
	},
	User2: CreateUserRequest{
		Name:  UserFixtures.ValidUser2.Name,
		Email: UserFixtures.ValidUser2.Email,
	},
	User3: CreateUserRequest{
		Name:  UserFixtures.ValidUser3.Name,
		Email: UserFixtures.ValidUser3.Email,
	},
}

// ValidUpdateRequests provides valid update data for testing
var ValidUpdateRequests = struct {
	UpdateUser1 UpdateUserRequest
}{
	UpdateUser1: UpdateUserRequest{
		Name:  UserFixtures.UpdatedUser1.Name,
		Email: UserFixtures.UpdatedUser1.Email,
	},
}
