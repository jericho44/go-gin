package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"gin-golang-app/internal/config"
	"gin-golang-app/internal/database"
	"gin-golang-app/internal/models"
	"gin-golang-app/internal/server"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite provides a test suite for integration tests
type IntegrationTestSuite struct {
	suite.Suite
	server  *httptest.Server
	db      *database.Database
	config  *config.Config
	router  *gin.Engine
	cleanup func()
}

// SetupSuite runs once before all tests in the suite
func (suite *IntegrationTestSuite) SetupSuite() {
	// Set test environment
	os.Setenv("APP_ENV", "test")
	os.Setenv("DB_NAME", "gin_app_test")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "")
	os.Setenv("JWT_SECRET", "test-secret-key")

	// Load test configuration
	suite.config = config.Load()

	// Set up test database
	suite.setupTestDatabase()

	// Create test server
	suite.setupTestServer()

	log.Println("Integration test suite setup completed")
}

// TearDownSuite runs once after all tests in the suite
func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}

	if suite.cleanup != nil {
		suite.cleanup()
	}

	if suite.db != nil {
		suite.db.Close()
	}

	log.Println("Integration test suite teardown completed")
}

// SetupTest runs before each test
func (suite *IntegrationTestSuite) SetupTest() {
	// Clean up database before each test
	suite.cleanupDatabase()
}

// TearDownTest runs after each test
func (suite *IntegrationTestSuite) TearDownTest() {
	// Clean up database after each test
	suite.cleanupDatabase()
}

// setupTestDatabase creates and configures the test database
func (suite *IntegrationTestSuite) setupTestDatabase() {
	// Create database configuration for testing
	dbConfig := &database.DatabaseConfig{
		Driver:          "postgres",
		Host:            suite.config.Database.Host,
		Port:            5432, // Convert string to int
		Username:        suite.config.Database.User,
		Password:        suite.config.Database.Password,
		DatabaseName:    suite.config.Database.Name,
		SSLMode:         suite.config.Database.SSLMode,
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	// First, connect to postgres database to create test database
	suite.createTestDatabase(dbConfig)

	// Connect to the test database
	db, err := database.NewDatabase(dbConfig)
	suite.Require().NoError(err, "Failed to connect to test database")

	suite.db = db

	// Create tables
	suite.createTables()
}

// createTestDatabase creates the test database if it doesn't exist
func (suite *IntegrationTestSuite) createTestDatabase(dbConfig *database.DatabaseConfig) {
	// Connect to postgres database to create test database
	postgresConfig := *dbConfig
	postgresConfig.DatabaseName = "postgres"

	db, err := database.NewDatabase(&postgresConfig)
	suite.Require().NoError(err, "Failed to connect to postgres database")
	defer db.Close()

	// Check if test database exists
	var exists bool
	query := "SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)"
	err = db.DB.QueryRow(query, dbConfig.DatabaseName).Scan(&exists)
	suite.Require().NoError(err, "Failed to check if test database exists")

	// Create test database if it doesn't exist
	if !exists {
		createQuery := fmt.Sprintf("CREATE DATABASE %s", dbConfig.DatabaseName)
		_, err = db.DB.Exec(createQuery)
		suite.Require().NoError(err, "Failed to create test database")
		log.Printf("Created test database: %s", dbConfig.DatabaseName)
	}

	// Set up cleanup function to drop test database
	suite.cleanup = func() {
		dropQuery := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbConfig.DatabaseName)
		_, err := db.DB.Exec(dropQuery)
		if err != nil {
			log.Printf("Failed to drop test database: %v", err)
		} else {
			log.Printf("Dropped test database: %s", dbConfig.DatabaseName)
		}
	}
}

// createTables creates the necessary tables for testing
func (suite *IntegrationTestSuite) createTables() {
	// Drop existing trigger and table if they exist
	dropTrigger := `DROP TRIGGER IF EXISTS update_users_updated_at ON users`
	_, err := suite.db.DB.Exec(dropTrigger)
	suite.Require().NoError(err, "Failed to drop existing trigger")

	dropFunction := `DROP FUNCTION IF EXISTS update_updated_at_column()`
	_, err = suite.db.DB.Exec(dropFunction)
	suite.Require().NoError(err, "Failed to drop existing function")

	dropUsersTable := `DROP TABLE IF EXISTS users CASCADE`
	_, err = suite.db.DB.Exec(dropUsersTable)
	suite.Require().NoError(err, "Failed to drop existing users table")

	// Create users table
	createUsersTable := `
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err = suite.db.DB.Exec(createUsersTable)
	suite.Require().NoError(err, "Failed to create users table")

	// Create trigger to update updated_at column
	createTrigger := `
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		DROP TRIGGER IF EXISTS update_users_updated_at ON users;
		CREATE TRIGGER update_users_updated_at
			BEFORE UPDATE ON users
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();
	`

	_, err = suite.db.DB.Exec(createTrigger)
	suite.Require().NoError(err, "Failed to create update trigger")
}

// setupTestServer creates and configures the test server
func (suite *IntegrationTestSuite) setupTestServer() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create the server with test configuration
	srv, err := server.NewServer(suite.config, suite.db)
	suite.Require().NoError(err, "Failed to create test server")

	suite.router = srv.Router

	// Create test server
	suite.server = httptest.NewServer(suite.router)
}

// cleanupDatabase removes all data from tables
func (suite *IntegrationTestSuite) cleanupDatabase() {
	// Delete all users
	_, err := suite.db.DB.Exec("DELETE FROM users")
	suite.Require().NoError(err, "Failed to clean up users table")

	// Reset sequences
	_, err = suite.db.DB.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1")
	suite.Require().NoError(err, "Failed to reset users sequence")
}

// GetBaseURL returns the base URL for the test server
func (suite *IntegrationTestSuite) GetBaseURL() string {
	return suite.server.URL
}

// GetDB returns the test database connection
func (suite *IntegrationTestSuite) GetDB() *sql.DB {
	return suite.db.DB
}

// ExecuteInTransaction executes a function within a database transaction
func (suite *IntegrationTestSuite) ExecuteInTransaction(fn func(*sql.Tx) error) error {
	tx, err := suite.db.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// WaitForServer waits for the server to be ready
func (suite *IntegrationTestSuite) WaitForServer(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("server not ready within timeout")
		case <-ticker.C:
			if suite.server != nil {
				return nil
			}
		}
	}
}

// SeedUsers inserts test users into the database
func (suite *IntegrationTestSuite) SeedUsers(users ...TestUser) []TestUser {
	var seededUsers []TestUser

	for _, user := range users {
		query := `
			INSERT INTO users (name, email, created_at, updated_at)
			VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			RETURNING id, name, email, created_at, updated_at
		`

		var seededUser TestUser
		err := suite.GetDB().QueryRow(query, user.Name, user.Email).Scan(
			&seededUser.ID,
			&seededUser.Name,
			&seededUser.Email,
			&seededUser.CreatedAt,
			&seededUser.UpdatedAt,
		)
		suite.Require().NoError(err, "Failed to seed user")

		seededUsers = append(seededUsers, seededUser)
	}

	return seededUsers
}

// SeedUser inserts a single test user into the database
func (suite *IntegrationTestSuite) SeedUser(user TestUser) TestUser {
	users := suite.SeedUsers(user)
	return users[0]
}

// GetUserFromDB retrieves a user from the database by ID
func (suite *IntegrationTestSuite) GetUserFromDB(id uint) (*models.User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := suite.GetDB().QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserFromDBByEmail retrieves a user from the database by email
func (suite *IntegrationTestSuite) GetUserFromDBByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := suite.GetDB().QueryRow(query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// CountUsersInDB returns the total number of users in the database
func (suite *IntegrationTestSuite) CountUsersInDB() int64 {
	var count int64
	err := suite.GetDB().QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	suite.Require().NoError(err, "Failed to count users")
	return count
}

// UserExistsInDB checks if a user exists in the database by ID
func (suite *IntegrationTestSuite) UserExistsInDB(id uint) bool {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)"
	err := suite.GetDB().QueryRow(query, id).Scan(&exists)
	suite.Require().NoError(err, "Failed to check if user exists")
	return exists
}

// UserExistsInDBByEmail checks if a user exists in the database by email
func (suite *IntegrationTestSuite) UserExistsInDBByEmail(email string) bool {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"
	err := suite.GetDB().QueryRow(query, email).Scan(&exists)
	suite.Require().NoError(err, "Failed to check if user exists by email")
	return exists
}

// AssertJSONResponse is a helper to assert JSON response structure
func (suite *IntegrationTestSuite) AssertJSONResponse(resp *http.Response, expectedStatus int, expectedSuccess bool) APIResponse {
	suite.Equal(expectedStatus, resp.StatusCode, "Unexpected status code")
	suite.Equal("application/json; charset=utf-8", resp.Header.Get("Content-Type"), "Unexpected content type")

	var response APIResponse
	err := DecodeResponse(resp, &response)
	suite.Require().NoError(err, "Failed to decode response")

	suite.Equal(expectedSuccess, response.Success, "Unexpected success value")

	return response
}

// AssertPaginatedJSONResponse is a helper to assert paginated JSON response structure
func (suite *IntegrationTestSuite) AssertPaginatedJSONResponse(resp *http.Response, expectedStatus int) PaginatedResponse {
	suite.Equal(expectedStatus, resp.StatusCode, "Unexpected status code")
	suite.Equal("application/json; charset=utf-8", resp.Header.Get("Content-Type"), "Unexpected content type")

	var response PaginatedResponse
	err := DecodeResponse(resp, &response)
	suite.Require().NoError(err, "Failed to decode paginated response")

	suite.True(response.Success, "Expected success to be true")
	suite.NotNil(response.Data, "Expected data to be present")
	suite.NotZero(response.Pagination, "Expected pagination to be present")

	return response
}

// AssertErrorResponse is a helper to assert error response structure
func (suite *IntegrationTestSuite) AssertErrorResponse(resp *http.Response, expectedStatus int, expectedMessage string) APIResponse {
	response := suite.AssertJSONResponse(resp, expectedStatus, false)

	if expectedMessage != "" {
		suite.Contains(response.Message, expectedMessage, "Unexpected error message")
	}

	suite.NotEmpty(response.Error, "Expected error field to be present")

	return response
}

// AssertSuccessResponse is a helper to assert success response structure
func (suite *IntegrationTestSuite) AssertSuccessResponse(resp *http.Response, expectedStatus int, expectedMessage string) APIResponse {
	response := suite.AssertJSONResponse(resp, expectedStatus, true)

	if expectedMessage != "" {
		suite.Contains(response.Message, expectedMessage, "Unexpected success message")
	}

	suite.NotNil(response.Data, "Expected data field to be present")

	return response
}

// GetHTTPClient returns an HTTP client configured for the test server
func (suite *IntegrationTestSuite) GetHTTPClient() *HTTPClient {
	return NewHTTPClient(suite.GetBaseURL())
}

// GetAuthenticatedHTTPClient returns an HTTP client with authentication headers
func (suite *IntegrationTestSuite) GetAuthenticatedHTTPClient(userID uint, email string) (*HTTPClient, string, error) {
	client := suite.GetHTTPClient()

	token, err := GenerateJWTToken("test-secret-key", userID, email)
	if err != nil {
		return nil, "", err
	}

	return client, token, nil
}
