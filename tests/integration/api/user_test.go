package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

// UserIntegrationTestSuite tests user-related endpoints
type UserIntegrationTestSuite struct {
	IntegrationTestSuite
}

// TestUserIntegrationSuite runs the user integration test suite
func TestUserIntegrationSuite(t *testing.T) {
	suite.Run(t, new(UserIntegrationTestSuite))
}

// TestCreateUser tests user creation endpoint
func (suite *UserIntegrationTestSuite) TestCreateUser() {
	client, token, err := suite.GetAuthenticatedHTTPClient(1, "test@example.com")
	suite.Require().NoError(err, "Failed to get authenticated client")

	// Test successful user creation
	resp, err := client.POST("/api/v1/users", ValidUserRequests.User1, AuthHeaders(token))
	suite.Require().NoError(err, "Failed to make create user request")

	response := suite.AssertSuccessResponse(resp, http.StatusCreated, "User created successfully")

	// Verify response data
	userData, ok := response.Data.(map[string]interface{})
	suite.Require().True(ok, "Expected user data to be a map")

	suite.Equal(ValidUserRequests.User1.Name, userData["name"], "Expected correct name")
	suite.Equal(ValidUserRequests.User1.Email, userData["email"], "Expected correct email")
	suite.Contains(userData, "id", "Expected user ID")
	suite.Contains(userData, "created_at", "Expected created_at timestamp")
	suite.Contains(userData, "updated_at", "Expected updated_at timestamp")

	// Verify user was created in database
	userID := uint(userData["id"].(float64))
	suite.True(suite.UserExistsInDB(userID), "Expected user to exist in database")
}

// TestCreateUserValidation tests user creation validation
func (suite *UserIntegrationTestSuite) TestCreateUserValidation() {
	client, token, err := suite.GetAuthenticatedHTTPClient(1, "test@example.com")
	suite.Require().NoError(err, "Failed to get authenticated client")

	testCases := []struct {
		name           string
		request        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Empty name",
			request:        InvalidUserRequests.EmptyName,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name:           "Empty email",
			request:        InvalidUserRequests.EmptyEmail,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name:           "Invalid email",
			request:        InvalidUserRequests.InvalidEmail,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name:           "Missing name",
			request:        InvalidUserRequests.MissingName,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name:           "Missing email",
			request:        InvalidUserRequests.MissingEmail,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name:           "Empty request",
			request:        InvalidUserRequests.EmptyRequest,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := client.POST("/api/v1/users", tc.request, AuthHeaders(token))
			suite.Require().NoError(err, "Failed to make create user request")

			suite.AssertErrorResponse(resp, tc.expectedStatus, tc.expectedError)
		})
	}
}

// TestCreateUserInvalidJSON tests user creation with invalid JSON
func (suite *UserIntegrationTestSuite) TestCreateUserInvalidJSON() {
	client, token, err := suite.GetAuthenticatedHTTPClient(1, "test@example.com")
	suite.Require().NoError(err, "Failed to get authenticated client")

	resp, err := client.POSTRaw("/api/v1/users", InvalidUserRequests.InvalidJSON, AuthHeaders(token))
	suite.Require().NoError(err, "Failed to make create user request")

	suite.AssertErrorResponse(resp, http.StatusBadRequest, "Invalid request body")
}

// TestCreateUserDuplicateEmail tests user creation with duplicate email
func (suite *UserIntegrationTestSuite) TestCreateUserDuplicateEmail() {
	client, token, err := suite.GetAuthenticatedHTTPClient(1, "test@example.com")
	suite.Require().NoError(err, "Failed to get authenticated client")

	// Create first user
	suite.SeedUser(UserFixtures.ValidUser1)

	// Try to create user with same email
	duplicateRequest := CreateUserRequest{
		Name:  "Different Name",
		Email: UserFixtures.ValidUser1.Email,
	}

	resp, err := client.POST("/api/v1/users", duplicateRequest, AuthHeaders(token))
	suite.Require().NoError(err, "Failed to make create user request")

	suite.AssertErrorResponse(resp, http.StatusBadRequest, "User already exists")
}

// TestGetUser tests user retrieval endpoint
func (suite *UserIntegrationTestSuite) TestGetUser() {
	client, token, err := suite.GetAuthenticatedHTTPClient(1, "test@example.com")
	suite.Require().NoError(err, "Failed to get authenticated client")

	// Seed test user
	seededUser := suite.SeedUser(UserFixtures.ValidUser1)

	// Test successful user retrieval
	resp, err := client.GET(fmt.Sprintf("/api/v1/users/%d", seededUser.ID), AuthHeaders(token))
	suite.Require().NoError(err, "Failed to make get user request")

	response := suite.AssertSuccessResponse(resp, http.StatusOK, "User retrieved successfully")

	// Verify response data
	userData, ok := response.Data.(map[string]interface{})
	suite.Require().True(ok, "Expected user data to be a map")

	suite.Equal(float64(seededUser.ID), userData["id"], "Expected correct user ID")
	suite.Equal(seededUser.Name, userData["name"], "Expected correct name")
	suite.Equal(seededUser.Email, userData["email"], "Expected correct email")
}

// TestGetUserNotFound tests user retrieval with non-existent ID
func (suite *UserIntegrationTestSuite) TestGetUserNotFound() {
	client, token, err := suite.GetAuthenticatedHTTPClient(1, "test@example.com")
	suite.Require().NoError(err, "Failed to get authenticated client")

	resp, err := client.GET("/api/v1/users/999", AuthHeaders(token))
	suite.Require().NoError(err, "Failed to make get user request")

	suite.AssertErrorResponse(resp, http.StatusNotFound, "User not found")
}

// TestGetUserInvalidID tests user retrieval with invalid ID
func (suite *UserIntegrationTestSuite) TestGetUserInvalidID() {
	client, token, err := suite.GetAuthenticatedHTTPClient(1, "test@example.com")
	suite.Require().NoError(err, "Failed to get authenticated client")

	invalidIDs := []string{"abc", "0", "-1", "1.5"}

	for _, id := range invalidIDs {
		suite.Run(fmt.Sprintf("Invalid ID: %s", id), func() {
			resp, err := client.GET(fmt.Sprintf("/api/v1/users/%s", id), AuthHeaders(token))
			suite.Require().NoError(err, "Failed to make get user request")

			suite.AssertErrorResponse(resp, http.StatusBadRequest, "Invalid user ID")
		})
	}
}

// TestUpdateUser tests user update endpoint
func (suite *UserIntegrationTestSuite) TestUpdateUser() {
	client, token, err := suite.GetAuthenticatedHTTPClient(1, "test@example.com")
	suite.Require().NoError(err, "Failed to get authenticated client")

	// Seed test user
	seededUser := suite.SeedUser(UserFixtures.ValidUser1)

	// Test successful user update
	resp, err := client.PUT(fmt.Sprintf("/api/v1/users/%d", seededUser.ID), ValidUpdateRequests.UpdateUser1, AuthHeaders(token))
	suite.Require().NoError(err, "Failed to make update user request")

	response := suite.AssertSuccessResponse(resp, http.StatusOK, "User updated successfully")

	// Verify response data
	userData, ok := response.Data.(map[string]interface{})
	suite.Require().True(ok, "Expected user data to be a map")

	suite.Equal(float64(seededUser.ID), userData["id"], "Expected correct user ID")
	suite.Equal(ValidUpdateRequests.UpdateUser1.Name, userData["name"], "Expected updated name")
	suite.Equal(ValidUpdateRequests.UpdateUser1.Email, userData["email"], "Expected updated email")

	// Verify user was updated in database
	updatedUser, err := suite.GetUserFromDB(seededUser.ID)
	suite.Require().NoError(err, "Failed to get updated user from database")
	suite.Equal(ValidUpdateRequests.UpdateUser1.Name, updatedUser.Name, "Expected name to be updated in database")
	suite.Equal(ValidUpdateRequests.UpdateUser1.Email, updatedUser.Email, "Expected email to be updated in database")
}

// TestUpdateUserNotFound tests user update with non-existent ID
func (suite *UserIntegrationTestSuite) TestUpdateUserNotFound() {
	client, token, err := suite.GetAuthenticatedHTTPClient(1, "test@example.com")
	suite.Require().NoError(err, "Failed to get authenticated client")

	resp, err := client.PUT("/api/v1/users/999", ValidUpdateRequests.UpdateUser1, AuthHeaders(token))
	suite.Require().NoError(err, "Failed to make update user request")

	suite.AssertErrorResponse(resp, http.StatusNotFound, "User not found")
}

// TestUpdateUserValidation tests user update validation
func (suite *UserIntegrationTestSuite) TestUpdateUserValidation() {
	client, token, err := suite.GetAuthenticatedHTTPClient(1, "test@example.com")
	suite.Require().NoError(err, "Failed to get authenticated client")

	// Seed test user
	seededUser := suite.SeedUser(UserFixtures.ValidUser1)

	testCases := []struct {
		name           string
		request        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Empty name",
			request:        UpdateUserRequest{Name: "", Email: "valid@example.com"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name:           "Empty email",
			request:        UpdateUserRequest{Name: "Valid Name", Email: ""},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name:           "Invalid email",
			request:        UpdateUserRequest{Name: "Valid Name", Email: "invalid-email"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := client.PUT(fmt.Sprintf("/api/v1/users/%d", seededUser.ID), tc.request, AuthHeaders(token))
			suite.Require().NoError(err, "Failed to make update user request")

			suite.AssertErrorResponse(resp, tc.expectedStatus, tc.expectedError)
		})
	}
}

// TestDeleteUser tests user deletion endpoint
func (suite *UserIntegrationTestSuite) TestDeleteUser() {
	client, token, err := suite.GetAuthenticatedHTTPClient(1, "test@example.com")
	suite.Require().NoError(err, "Failed to get authenticated client")

	// Seed test user
	seededUser := suite.SeedUser(UserFixtures.ValidUser1)

	// Verify user exists before deletion
	suite.True(suite.UserExistsInDB(seededUser.ID), "Expected user to exist before deletion")

	// Test successful user deletion
	resp, err := client.DELETE(fmt.Sprintf("/api/v1/users/%d", seededUser.ID), AuthHeaders(token))
	suite.Require().NoError(err, "Failed to make delete user request")

	suite.Equal(http.StatusNoContent, resp.StatusCode, "Expected no content status")

	// Verify user was deleted from database
	suite.False(suite.UserExistsInDB(seededUser.ID), "Expected user to be deleted from database")
}

// TestDeleteUserNotFound tests user deletion with non-existent ID
func (suite *UserIntegrationTestSuite) TestDeleteUserNotFound() {
	client, token, err := suite.GetAuthenticatedHTTPClient(1, "test@example.com")
	suite.Require().NoError(err, "Failed to get authenticated client")

	resp, err := client.DELETE("/api/v1/users/999", AuthHeaders(token))
	suite.Require().NoError(err, "Failed to make delete user request")

	suite.AssertErrorResponse(resp, http.StatusNotFound, "User not found")
}

// TestListUsers tests user listing endpoint
func (suite *UserIntegrationTestSuite) TestListUsers() {
	client, token, err := suite.GetAuthenticatedHTTPClient(1, "test@example.com")
	suite.Require().NoError(err, "Failed to get authenticated client")

	// Seed test users
	users := suite.SeedUsers(UserFixtures.ValidUser1, UserFixtures.ValidUser2, UserFixtures.ValidUser3)

	// Test successful user listing
	resp, err := client.GET("/api/v1/users", AuthHeaders(token))
	suite.Require().NoError(err, "Failed to make list users request")

	response := suite.AssertPaginatedJSONResponse(resp, http.StatusOK)

	// Verify response data
	userList, ok := response.Data.([]interface{})
	suite.Require().True(ok, "Expected data to be a slice")
	suite.Len(userList, 3, "Expected 3 users in response")

	// Verify pagination
	suite.Equal(1, response.Pagination.Page, "Expected page 1")
	suite.Equal(10, response.Pagination.Limit, "Expected limit 10")
	suite.Equal(int64(3), response.Pagination.Total, "Expected total 3")
	suite.Equal(1, response.Pagination.TotalPages, "Expected 1 total page")

	// Verify user data
	for i, userInterface := range userList {
		userData, ok := userInterface.(map[string]interface{})
		suite.Require().True(ok, "Expected user data to be a map")

		suite.Equal(float64(users[i].ID), userData["id"], "Expected correct user ID")
		suite.Equal(users[i].Name, userData["name"], "Expected correct name")
		suite.Equal(users[i].Email, userData["email"], "Expected correct email")
	}
}

// TestListUsersWithPagination tests user listing with pagination parameters
func (suite *UserIntegrationTestSuite) TestListUsersWithPagination() {
	client, token, err := suite.GetAuthenticatedHTTPClient(1, "test@example.com")
	suite.Require().NoError(err, "Failed to get authenticated client")

	// Seed 5 test users
	users := make([]TestUser, 5)
	for i := 0; i < 5; i++ {
		users[i] = TestUser{
			Name:  fmt.Sprintf("User %d", i+1),
			Email: fmt.Sprintf("user%d@example.com", i+1),
		}
	}
	suite.SeedUsers(users...)

	// Test pagination with limit 2
	resp, err := client.GET("/api/v1/users?page=1&limit=2", AuthHeaders(token))
	suite.Require().NoError(err, "Failed to make list users request")

	response := suite.AssertPaginatedJSONResponse(resp, http.StatusOK)

	// Verify response data
	userList, ok := response.Data.([]interface{})
	suite.Require().True(ok, "Expected data to be a slice")
	suite.Len(userList, 2, "Expected 2 users in response")

	// Verify pagination
	suite.Equal(1, response.Pagination.Page, "Expected page 1")
	suite.Equal(2, response.Pagination.Limit, "Expected limit 2")
	suite.Equal(int64(5), response.Pagination.Total, "Expected total 5")
	suite.Equal(3, response.Pagination.TotalPages, "Expected 3 total pages")

	// Test second page
	resp, err = client.GET("/api/v1/users?page=2&limit=2", AuthHeaders(token))
	suite.Require().NoError(err, "Failed to make list users request")

	response = suite.AssertPaginatedJSONResponse(resp, http.StatusOK)

	// Verify response data
	userList, ok = response.Data.([]interface{})
	suite.Require().True(ok, "Expected data to be a slice")
	suite.Len(userList, 2, "Expected 2 users in response")

	// Verify pagination
	suite.Equal(2, response.Pagination.Page, "Expected page 2")
	suite.Equal(2, response.Pagination.Limit, "Expected limit 2")
	suite.Equal(int64(5), response.Pagination.Total, "Expected total 5")
	suite.Equal(3, response.Pagination.TotalPages, "Expected 3 total pages")
}

// TestGetUserByEmail tests user retrieval by email endpoint
func (suite *UserIntegrationTestSuite) TestGetUserByEmail() {
	client, token, err := suite.GetAuthenticatedHTTPClient(1, "test@example.com")
	suite.Require().NoError(err, "Failed to get authenticated client")

	// Seed test user
	seededUser := suite.SeedUser(UserFixtures.ValidUser1)

	// Test successful user retrieval by email
	resp, err := client.GET(fmt.Sprintf("/api/v1/users/email/%s", seededUser.Email), AuthHeaders(token))
	suite.Require().NoError(err, "Failed to make get user by email request")

	response := suite.AssertSuccessResponse(resp, http.StatusOK, "User retrieved successfully")

	// Verify response data
	userData, ok := response.Data.(map[string]interface{})
	suite.Require().True(ok, "Expected user data to be a map")

	suite.Equal(float64(seededUser.ID), userData["id"], "Expected correct user ID")
	suite.Equal(seededUser.Name, userData["name"], "Expected correct name")
	suite.Equal(seededUser.Email, userData["email"], "Expected correct email")
}

// TestGetUserByEmailNotFound tests user retrieval by email with non-existent email
func (suite *UserIntegrationTestSuite) TestGetUserByEmailNotFound() {
	client, token, err := suite.GetAuthenticatedHTTPClient(1, "test@example.com")
	suite.Require().NoError(err, "Failed to get authenticated client")

	resp, err := client.GET("/api/v1/users/email/nonexistent@example.com", AuthHeaders(token))
	suite.Require().NoError(err, "Failed to make get user by email request")

	suite.AssertErrorResponse(resp, http.StatusNotFound, "User not found")
}
